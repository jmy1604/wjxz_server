package main

import (
	"libs/log"
	"libs/utils"
	"math"
	"public_message/gen_go/client_message"
	"sync"
	"time"
	"youma/rpc_common"
)

const PERSONAL_SPACE_MAX_PICTURE_NUM int32 = 6
const LEAVE_MSG_DEFAULT_LENGTH int32 = 100
const COMMENT_DEFAULT_LENGTH int32 = 100
const PERSONAL_SPACE_GET_LEAVE_MSG_NUM int32 = 10
const PERSONAL_SPACE_GET_COMMNET_NUM int32 = 10

// 留言评论
type PSLeaveMessageComment struct {
	id             int32
	content        []byte
	send_player_id int32
	send_time      int32
}

func (this *PSLeaveMessageComment) ToRedisData() *RedisPSComment {
	return &RedisPSComment{
		CommentId:    this.id,
		Content:      this.content,
		SendPlayerId: this.send_player_id,
		SendTime:     this.send_time,
	}
}

// 留言
type PSLeaveMessage struct {
	id               int32
	content          []byte
	send_player_id   int32
	comments         []*PSLeaveMessageComment
	curr_comment_id  int32
	send_time        int32
	comment_ids_load []int32
}

func (this *PSLeaveMessage) Init(send_player_id, msg_id int32, content []byte) {
	this.id = msg_id
	this.content = content
	this.send_player_id = send_player_id
	this.send_time = int32(time.Now().Unix())
}

func (this *PSLeaveMessage) ToRedisData() *RedisPSLeaveMessage {
	return &RedisPSLeaveMessage{
		MsgId:         this.id,
		Content:       this.content,
		SendPlayerId:  this.send_player_id,
		SendTime:      this.send_time,
		CommentIds:    this.comment_ids_load,
		CurrCommentId: this.curr_comment_id,
	}
}

func (this *PSLeaveMessage) get_comment(comment_id int32) *PSLeaveMessageComment {
	if this.comments == nil {
		return nil
	}

	comments_len := int32(len(this.comments))
	for i := int32(0); i < comments_len; i++ {
		if this.comments[i].id == comment_id {
			return this.comments[i]
		}
	}
	return nil
}

func (this *PSLeaveMessage) GetSomeComments(start_index, comments_num int32) (comments []*rpc_common.H2R_PSLeaveMessageCommentData) {
	if this.comments == nil {
		return
	}

	if comments_num > PERSONAL_SPACE_GET_COMMNET_NUM {
		comments_num = PERSONAL_SPACE_GET_COMMNET_NUM
	}

	all_comments_num := int32(len(this.comments))
	if comments_num > all_comments_num {
		comments_num = all_comments_num
	}

	for i := start_index; i < start_index+comments_num; i++ {
		idx := all_comments_num - i - 1
		if idx < 0 || idx >= all_comments_num {
			break
		}
		c := this.comments[idx]
		if c == nil {
			log.Warn("leave message comment index[%v] is null", idx)
			continue
		}
		d := &rpc_common.H2R_PSLeaveMessageCommentData{
			Id:           c.id,
			Content:      c.content,
			SendPlayerId: c.send_player_id,
			SendTime:     c.send_time,
		}
		comments = append(comments, d)
	}
	return
}

func (this *PSLeaveMessage) AddNewComment(send_player_id int32, content []byte) int32 {
	new_comment := &PSLeaveMessageComment{}
	new_comment.content = content
	new_comment.send_player_id = send_player_id
	new_comment.send_time = int32(time.Now().Unix())
	this.curr_comment_id += 1
	new_comment.id = this.curr_comment_id
	this.comment_ids_load = append(this.comment_ids_load, this.curr_comment_id)
	this.comments = append(this.comments, new_comment)
	return this.curr_comment_id
}

func (this *PSLeaveMessage) DeleteComment(comment_id int32, send_player_id int32) bool {
	if this.comments == nil || len(this.comments) == 0 {
		return false
	}

	idx := int32(-1)
	ll := int32(len(this.comments))
	for i := int32(0); i < ll; i++ {
		if this.comments[i].id == comment_id {
			if this.comments[i].send_player_id != send_player_id {
				return false
			}
			idx = i
			break
		}
	}

	if idx < 0 {
		return false
	}

	for i := idx; i < ll-1; i++ {
		this.comments[i] = this.comments[i+1]
		this.comment_ids_load[i] = this.comment_ids_load[i+1]
	}

	this.comments = this.comments[:ll-1]
	this.comment_ids_load = this.comment_ids_load[:ll-1]

	return true
}

// 照片留言管理器
type PSPicLeaveMessageMgr struct {
	messages map[int32][]map[int32]*PSLeaveMessage
	locker   *sync.RWMutex
}

var ps_pic_leave_messages_mgr PSPicLeaveMessageMgr

func (this *PSPicLeaveMessageMgr) Init() {
	this.messages = make(map[int32][]map[int32]*PSLeaveMessage)
	this.locker = &sync.RWMutex{}
}

// 从redis数据载入
func (this *PSPicLeaveMessageMgr) LoadRedisLeaveMessage(player_id, pic_id int32, data *RedisPSLeaveMessage) bool {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("load redis leave msg[%v] for player_id[%v] with pic_id[%v] invalid", data.MsgId, player_id, pic_id)
		return false
	}

	pics_msgs := this.messages[player_id]
	if pics_msgs == nil {
		pics_msgs = make([]map[int32]*PSLeaveMessage, PERSONAL_SPACE_MAX_PICTURE_NUM)
	}

	pic_msgs := pics_msgs[pic_id-1]
	if pic_msgs == nil {
		pic_msgs = make(map[int32]*PSLeaveMessage)
		pics_msgs[pic_id-1] = pic_msgs
	}

	if pic_msgs[data.MsgId] != nil {
		log.Error("load redis leave msg[%v] for player_id[%v] with pic_id[%v] already exists", data.MsgId, player_id, pic_id)
		return false
	}

	msg := &PSLeaveMessage{}
	msg.id = data.MsgId
	msg.content = data.Content
	msg.send_player_id = data.SendPlayerId
	msg.send_time = data.SendTime
	msg.comment_ids_load = data.CommentIds
	msg.curr_comment_id = data.CurrCommentId
	msg.comments = data.load_comments(player_id, pic_id, data.MsgId)
	pic_msgs[data.MsgId] = msg

	return true
}

func get_pic_leave_message_id(player_id int32, pic_id int32, msg_id int32) int64 {
	low_id := int32((pic_id << 28) | msg_id)
	return utils.Int64From2Int32(player_id, low_id)
}

func (this *PSPicLeaveMessageMgr) AddNewLeaveMsg(player_id, pic_id, msg_id, send_player_id int32, content []byte) bool {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("Player[%v] add new leave msg with pic_id[%v] msg_id[%v] msg_content[%v] to player[%v] personal space invalid", player_id, pic_id, msg_id, content, send_player_id)
		return false
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	pics_msgs := this.messages[player_id]
	if pics_msgs == nil {
		pics_msgs = make([]map[int32]*PSLeaveMessage, PERSONAL_SPACE_MAX_PICTURE_NUM)
	}

	pic_msgs := pics_msgs[pic_id-1]
	if pic_msgs == nil {
		pic_msgs = make(map[int32]*PSLeaveMessage)
		pics_msgs[pic_id-1] = pic_msgs
	}

	if pic_msgs[msg_id] != nil {
		log.Error("Player[%v] add new leave msg with pic_id[%v] msg_id[%v] msg_content[%v] to player[%v] personal space failed, already exists", player_id, pic_id, msg_id, content, send_player_id)
		return false
	}

	new_msg := &PSLeaveMessage{}
	new_msg.Init(msg_id, send_player_id, content)
	pic_msgs[msg_id] = new_msg

	return true
}

func (this *PSPicLeaveMessageMgr) DeleteLeaveMsg(send_player_id, player_id, pic_id, msg_id int32) bool {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("Player[%v] delete leave msg with pic_id[%v] msg_id[%v] in player[%v] personal space invalid", send_player_id, pic_id, msg_id, player_id)
		return false
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	pics_msgs := this.messages[player_id]
	if pics_msgs == nil {
		log.Error("Player[%v] no pic leave messages data", player_id)
		return false
	}

	pic_msgs := pics_msgs[pic_id-1]
	if pic_msgs == nil {
		log.Error("Player[%v] no pic_id[%v] in pic leave messages data", player_id, pic_id)
		return false
	}

	msg := pic_msgs[msg_id]
	if msg == nil {
		log.Error("Player[%v] no pic_id[%v] msg_id[%v] in pic leave messages data", player_id, pic_id, msg_id)
		return false
	}

	if msg.send_player_id != send_player_id {
		log.Error("Player[%v] cant delete player[%v] leave msg[%v] in player[%v] personal space pic_id[%v]", send_player_id, msg.send_player_id, msg_id, player_id, pic_id)
		return false
	}

	delete(pic_msgs, msg_id)

	return true
}

func (this *PSPicLeaveMessageMgr) get_pic_msgs(player_id, pic_id int32) map[int32]*PSLeaveMessage {
	pics_msgs := this.messages[player_id]
	if pics_msgs == nil {
		log.Error("Player[%v] no pic leave messages data", player_id)
		return nil
	}

	pic_msgs := pics_msgs[pic_id-1]
	if pic_msgs == nil {
		log.Error("Player[%v] no pic_id[%v] in pic leave messages data", player_id, pic_id)
		return nil
	}

	return pic_msgs
}

func (this *PSPicLeaveMessageMgr) GetLeaveMsgs(player_id, pic_id int32, leave_msg_ids []int32) (leave_msgs []*rpc_common.H2R_PSLeaveMessageData) {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("pull leave msg with pic_id[%v] in player[%v] personal space invalid", pic_id, player_id)
		return nil
	}

	this.locker.RLock()
	defer this.locker.RUnlock()

	if leave_msg_ids == nil || len(leave_msg_ids) == 0 {
		return nil
	}

	pic_msgs := this.get_pic_msgs(player_id, pic_id)
	if pic_msgs == nil {
		return nil
	}

	for i := 0; i < len(leave_msg_ids); i++ {
		msg := pic_msgs[leave_msg_ids[i]]
		if msg == nil {
			continue
		}
		// 获取评论
		comments := msg.GetSomeComments(0, PERSONAL_SPACE_GET_COMMNET_NUM)
		d := &rpc_common.H2R_PSLeaveMessageData{
			Id:           msg.id,
			Content:      msg.content,
			SendPlayerId: msg.send_player_id,
			SendTime:     msg.send_time,
			Comments:     comments,
		}

		leave_msgs = append(leave_msgs, d)
	}

	return
}

func (this *PSPicLeaveMessageMgr) get_pic_msg(player_id, pic_id, msg_id int32) *PSLeaveMessage {
	pic_msgs := this.get_pic_msgs(player_id, pic_id)
	if pic_msgs == nil {
		log.Error("get player[%v] personal space pic_id[%v] data failed", player_id, pic_id)
		return nil
	}

	msg := pic_msgs[msg_id]
	if msg == nil {
		log.Error("get player[%v] personal space pic_id[%v] msg_id[%v] data failed", player_id, pic_id, msg_id)
		return nil
	}

	return msg
}

func (this *PSPicLeaveMessageMgr) AddNewLeaveMsgComment(player_id, pic_id, msg_id int32, comment []byte, send_player_id int32) int32 {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("add leave msg_id[%v] comment with pic_id[%v] in player[%v] personal space invalid", msg_id, pic_id, player_id)
		return -1
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	msg := this.get_pic_msg(player_id, pic_id, msg_id)
	if msg == nil {
		return -1
	}
	return msg.AddNewComment(send_player_id, comment)
}

func (this *PSPicLeaveMessageMgr) DeleteLeaveMsgComment(player_id, pic_id, msg_id, comment_id int32, send_player_id int32) int32 {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("delete leave msg_id[%v] comment_id[%v] with pic_id[%v] in player[%v] personal space invalid", msg_id, comment_id, pic_id, player_id)
		return -1
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	msg := this.get_pic_msg(player_id, pic_id, msg_id)
	if msg == nil {
		return int32(msg_client_message.E_ERR_PERSONAL_SPACE_NOT_FOUND_THE_LEAVE_MSG)
	}
	if !msg.DeleteComment(comment_id, send_player_id) {
		return int32(msg_client_message.E_ERR_PERSONAL_SPACE_NO_COMMENT_WITH_LEAVE_MSG)
	}
	return 1
}

func (this *PSPicLeaveMessageMgr) GetLeaveMsgComments(player_id, pic_id, msg_id, start_index, comment_num int32) (comments []*rpc_common.H2R_PSLeaveMessageCommentData, err int32) {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("get leave msg_id[%v] comment range[%v,%v] with pic_id[%v] in player[%v] personal space invalid", msg_id, start_index, comment_num, pic_id, player_id)
		err = int32(msg_client_message.E_ERR_PERSONAL_SPACE_PIC_ID_INVALID)
		return
	}

	this.locker.RLock()
	defer this.locker.RUnlock()

	msg := this.get_pic_msg(player_id, pic_id, msg_id)
	if msg == nil {
		log.Error("player[%v] get leave msg[%v] comments with pic_id[%v] not found the msg", player_id, msg_id, pic_id)
		err = int32(msg_client_message.E_ERR_PERSONAL_SPACE_NOT_FOUND_THE_LEAVE_MSG)
		return
	}

	comments = msg.GetSomeComments(start_index, comment_num)
	return
}

func (this *PSPicLeaveMessageMgr) ToLeaveMsgRedisData(player_id, pic_id, msg_id int32) *RedisPSLeaveMessage {
	msg := this.get_pic_msg(player_id, pic_id, msg_id)
	if msg == nil {
		return nil
	}
	return msg.ToRedisData()
}

func (this *PSPicLeaveMessageMgr) ToCommentRedisData(player_id, pic_id, msg_id, comment_id int32) *RedisPSComment {
	msg := this.get_pic_msg(player_id, pic_id, msg_id)
	if msg == nil {
		return nil
	}
	comment := msg.get_comment(comment_id)
	if comment == nil {
		return nil
	}
	return comment.ToRedisData()
}

// 赞管理器
type PersonalSpacePicZanMgr struct {
	zans   map[int32][]map[int32]int32
	locker *sync.RWMutex
}

var ps_pic_zan_mgr PersonalSpacePicZanMgr

func (this *PersonalSpacePicZanMgr) Init() {
	this.zans = make(map[int32][]map[int32]int32)
	this.locker = &sync.RWMutex{}
}

func (this *PersonalSpacePicZanMgr) IsZaned(player_id int32, pic_id int32, send_player_id int32) bool {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("Player[%v] check zaned pic_id[%v] to player[%v] personal space invalid", send_player_id, pic_id, player_id)
		return false
	}

	this.locker.RLock()
	defer this.locker.RUnlock()

	pics_zaned := this.zans[player_id]
	if pics_zaned == nil {
		return false
	}

	pic_zaned := pics_zaned[pic_id-1]
	if pic_zaned == nil {
		return false
	}

	if _, o := pic_zaned[send_player_id]; !o {
		return false
	}

	return true
}

func (this *PersonalSpacePicZanMgr) CheckAndZan(player_id int32, pic_id int32, send_player_id int32) int32 {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("Player[%v] check zaned pic_id[%v] to player[%v] personal space invalid", send_player_id, pic_id, player_id)
		return int32(msg_client_message.E_ERR_PERSONAL_SPACE_PIC_ID_INVALID)
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	pics_zaned := this.zans[player_id]
	if pics_zaned == nil {
		pics_zaned = make([]map[int32]int32, PERSONAL_SPACE_MAX_PICTURE_NUM)
		this.zans[player_id] = pics_zaned
	}

	pic_zaned := pics_zaned[pic_id]
	if pic_zaned == nil {
		pic_zaned = make(map[int32]int32)
		pics_zaned[pic_id] = pic_zaned
	}

	_, o := pic_zaned[send_player_id]
	if o {
		log.Error("Player[%v] already zaned player[%v] personal space pic[%v]", send_player_id, player_id, pic_id)
		return int32(msg_client_message.E_ERR_PERSONAL_SPACE_PIC_ALREADY_ZANED)
	}

	pic_zaned[send_player_id] = send_player_id
	return 1
}

// 照片数据
type PSPictureData struct {
	id                int32
	url               string
	upload_time       int32
	zaned             int32
	leave_msg_ids     []int32 // 留言按时间顺序降序排列
	curr_leave_msg_id int32
	leave_msg_ds_type int32 // 0 表示数组
}

func (this *PSPictureData) Init(id int32, url string) {
	this.id = id
	this.url = url
	this.upload_time = int32(time.Now().Unix())
}

func (this *PSPictureData) ToRedisData() *RedisPSPicture {
	return &RedisPSPicture{
		PicId:          this.id,
		Url:            this.url,
		UploadTime:     this.upload_time,
		Zaned:          this.zaned,
		LeaveMsgIds:    this.leave_msg_ids,
		CurrLeaveMsgId: this.curr_leave_msg_id,
		LeaveMsgDsType: this.leave_msg_ds_type,
	}
}

// 照片管理器
type PersonalSpacePictureMgr struct {
	pictures map[int32][]*PSPictureData
	locker   *sync.RWMutex
}

var ps_pic_mgr PersonalSpacePictureMgr

func (this *PersonalSpacePictureMgr) Init() {
	this.pictures = make(map[int32][]*PSPictureData)
	this.locker = &sync.RWMutex{}
}

// 载入redis数据
func (this *PersonalSpacePictureMgr) LoadPicture(player_id int32, data *RedisPSPicture) bool {
	if data.PicId <= 0 || data.PicId > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("Player[%v] add pic_id[%v] to personal space invalid", player_id, data.PicId)
		return false
	}

	pics := this.pictures[player_id]
	if pics == nil {
		pics = make([]*PSPictureData, PERSONAL_SPACE_MAX_PICTURE_NUM)
	}

	if pics[data.PicId-1] != nil {
		log.Error("Player[%v] personal space pic_id[%v] already exists", player_id, data.PicId)
		return false
	}

	new_pic := &PSPictureData{}
	new_pic.id = data.PicId
	new_pic.url = data.Url
	new_pic.upload_time = data.UploadTime
	new_pic.zaned = data.Zaned
	new_pic.leave_msg_ids = data.LeaveMsgIds
	new_pic.leave_msg_ds_type = data.LeaveMsgDsType
	pics[data.PicId-1] = new_pic

	return true
}

func (this *PersonalSpacePictureMgr) GetPicBaseData(player_id int32, pic_id int32) (has bool, url string, upload_time int32, zaned int32, msg_num int32) {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("Player[%v] get pic base data with pic_id[%v] to personal space invalid", player_id, pic_id)
		return
	}

	this.locker.RLock()
	defer this.locker.RUnlock()

	pics := this.pictures[player_id]
	if pics == nil || len(pics) == 0 {
		return
	}

	var pic *PSPictureData
	for i := 0; i < len(pics); i++ {
		if pics[i] != nil && pics[i].id == pic_id {
			pic = pics[i]
			break
		}
	}

	if pic == nil {
		return
	}

	url = pic.url
	upload_time = pic.upload_time
	zaned = pic.zaned
	if pic.leave_msg_ids != nil {
		msg_num = int32(len(pic.leave_msg_ids))
	}
	has = true

	return
}

func (this *PersonalSpacePictureMgr) AddPicture(player_id int32, pic_id int32, pic_url string) bool {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("Player[%v] add pic_id[%v] to personal space invalid", player_id, pic_id)
		return false
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	pics := this.pictures[player_id]
	if pics == nil {
		pics = make([]*PSPictureData, PERSONAL_SPACE_MAX_PICTURE_NUM)
	}

	if pics[pic_id-1] != nil {
		log.Error("Player[%v] personal space pic_id[%v] already exists", player_id, pic_id)
		return false
	}

	new_pic := &PSPictureData{}
	new_pic.Init(pic_id, pic_url)
	new_pic.url = pic_url
	new_pic.upload_time = int32(time.Now().Unix())
	pics[pic_id-1] = new_pic

	return true
}

func (this *PersonalSpacePictureMgr) DeletePicture(player_id int32, pic_id int32) bool {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("Player[%v] delete pic_id[%v] to personal space invalid", player_id, pic_id)
		return false
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	pics := this.pictures[player_id]
	if pics == nil {
		log.Error("Player[%v] delete pic_id[%v] not found player pics data", player_id, pic_id)
		return false
	}

	if pics[pic_id-1] == nil {
		log.Error("Player[%v] delete pic_id[%v] not found pic data", player_id, pic_id)
		return false
	}

	pics[pic_id-1] = nil

	return true
}

func (this *PersonalSpacePictureMgr) get_pic(player_id, pic_id int32) *PSPictureData {
	pics := this.pictures[player_id]
	if pics == nil {
		log.Error("Player[%v] with pic_id[%v] not found player pics data", player_id, pic_id)
		return nil
	}

	pic := pics[pic_id-1]
	if pic == nil {
		log.Error("Player[%v] with pic_id[%v] not found pic data", player_id, pic_id)
		return nil
	}

	return pic
}

func (this *PersonalSpacePictureMgr) ToPicRedisData(player_id, pic_id int32) *RedisPSPicture {
	this.locker.RLock()
	defer this.locker.RUnlock()

	pic := this.get_pic(player_id, pic_id)
	if pic == nil {
		return nil
	}

	return pic.ToRedisData()
}

func (this *PersonalSpacePictureMgr) AddZan(player_id int32, pic_id int32, add_val int32) int32 {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("Player[%v] add zan[%v] with pic_id[%v] to personal space invalid", player_id, add_val, pic_id)
		return -1
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	pic := this.get_pic(player_id, pic_id)
	if pic == nil {
		return -1
	}

	if math.MaxInt32-pic.zaned < add_val {
		pic.zaned = math.MaxInt32
	} else {
		pic.zaned += add_val
	}

	return pic.zaned
}

func (this *PersonalSpacePictureMgr) GetZan(player_id int32, pic_id int32) int32 {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("Player[%v] get zan with pic_id[%v] to personal space invalid", player_id, pic_id)
		return -1
	}

	this.locker.RLock()
	defer this.locker.RUnlock()

	pic := this.get_pic(player_id, pic_id)
	if pic == nil {
		return -1
	}

	return pic.zaned
}

func (this *PersonalSpacePictureMgr) GetNewMsgId(player_id int32, pic_id int32) int32 {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("Player[%v] get new msg id with pic_id[%v] to personal space invalid", player_id, pic_id)
		return -1
	}

	this.locker.RLock()
	defer this.locker.RUnlock()

	pic := this.get_pic(player_id, pic_id)
	if pic == nil {
		return -1
	}

	if pic.curr_leave_msg_id >= math.MaxInt32 {
		pic.curr_leave_msg_id = 1
	} else {
		pic.curr_leave_msg_id += 1
	}
	pic.leave_msg_ids = append(pic.leave_msg_ids, pic.curr_leave_msg_id)
	return pic.curr_leave_msg_id
}

func (this *PersonalSpacePictureMgr) GetLeaveMsgIds(player_id, pic_id, start_index, msg_num int32) ([]int32, bool) {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("Player[%v] get leave msg ids with pic_id[%v] range[%v,%v] to personal space invalid", player_id, pic_id, start_index, msg_num)
		return nil, false
	}

	this.locker.RLock()
	defer this.locker.RUnlock()

	pic := this.get_pic(player_id, pic_id)
	if pic == nil {
		return nil, false
	}

	msg_ids_len := int32(len(pic.leave_msg_ids))
	if pic.leave_msg_ids == nil || msg_num == 0 || start_index >= msg_ids_len {
		return make([]int32, 0), false
	}

	var is_more bool
	if start_index+msg_num > msg_ids_len {
		msg_num = msg_ids_len - start_index
	} else if start_index+msg_num < msg_ids_len {
		is_more = true
	}

	var msg_ids []int32
	for i := start_index; i < start_index+msg_num; i++ {
		idx := msg_ids_len - start_index - 1
		if idx < 0 || idx >= msg_ids_len {
			break
		}
		msg_ids = append(msg_ids, pic.leave_msg_ids[idx])
	}

	return msg_ids, is_more
}

func (this *PersonalSpacePictureMgr) DeleteLeaveMsgId(player_id, pic_id, msg_id int32) bool {
	if pic_id <= 0 || pic_id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		log.Error("Player[%v] delete leave msg[%v] with pic_id[%v] to personal space invalid", player_id, msg_id, pic_id)
		return false
	}

	this.locker.Lock()
	defer this.locker.Unlock()

	pic := this.get_pic(player_id, pic_id)
	if pic == nil {
		return false
	}

	if pic.leave_msg_ids == nil {
		return false
	}

	msg_ids_len := int32(len(pic.leave_msg_ids))
	idx := int32(-1)
	for i := int32(0); i < msg_ids_len; i++ {
		if pic.leave_msg_ids[i] == msg_id {
			idx = i
			break
		}
	}

	if idx < 0 {
		log.Error("player[%v] delete leave msg[%v] with pic_id[%v] not found", player_id, msg_id, pic_id)
		return false
	}

	for i := idx; i < msg_ids_len-1; i++ {
		pic.leave_msg_ids[i] = pic.leave_msg_ids[i+1]
	}

	pic.leave_msg_ids = pic.leave_msg_ids[:msg_ids_len-1]

	return true
}

// 空间留言管理器
type PSLeaveMessageMgr struct {
	player2msgs map[int32]map[int32]*PSLeaveMessage
	locker      *sync.RWMutex
}

var ps_leave_messages_mgr PSLeaveMessageMgr

func (this *PSLeaveMessageMgr) Init() {
	this.player2msgs = make(map[int32]map[int32]*PSLeaveMessage)
	this.locker = &sync.RWMutex{}
}

// 从redis数据载入
func (this *PSLeaveMessageMgr) LoadLeaveMessage(player_id int32, data *RedisPSLeaveMessage) bool {
	player_msgs := this.player2msgs[player_id]
	if player_msgs == nil {
		player_msgs = make(map[int32]*PSLeaveMessage)
		this.player2msgs[player_id] = player_msgs
	}

	if player_msgs[data.MsgId] != nil {
		log.Warn("Player[%v] already has personal space leave message[%v]", player_id, data.MsgId)
		return false
	}

	msg := &PSLeaveMessage{}
	msg.id = data.MsgId
	msg.content = data.Content
	msg.send_player_id = data.SendPlayerId
	msg.send_time = data.SendTime
	msg.comment_ids_load = data.CommentIds
	msg.curr_comment_id = data.CurrCommentId
	msg.comments = data.load_comments(player_id, 0, data.MsgId)
	player_msgs[data.MsgId] = msg

	return true
}

func (this *PSLeaveMessageMgr) AddNew(player_id, send_player_id, msg_id int32, content []byte) bool {
	this.locker.Lock()
	defer this.locker.Unlock()

	player_msgs := this.player2msgs[player_id]
	if player_msgs == nil {
		player_msgs = make(map[int32]*PSLeaveMessage)
		this.player2msgs[player_id] = player_msgs
	}

	if player_msgs[msg_id] != nil {
		log.Warn("Player[%v] already has personal space leave message[%v], overwirte it", player_id, msg_id)
		//return false
	}

	new_msg := &PSLeaveMessage{}
	new_msg.Init(send_player_id, msg_id, content)
	player_msgs[msg_id] = new_msg
	return true
}

func (this *PSLeaveMessageMgr) Delete(player_id, msg_id, send_player_id int32) bool {
	this.locker.Lock()
	defer this.locker.Unlock()

	player_msgs := this.player2msgs[player_id]
	if player_msgs == nil {
		log.Warn("Player[%v] not found in personal space leave message manager for delete msg[%v]", player_id, msg_id)
		return false
	}
	msg := player_msgs[msg_id]
	if msg == nil {
		log.Warn("Player[%v] no message[%v] in personal space leave message manager", player_id, msg_id)
		return false
	}

	if msg.send_player_id != send_player_id {
		log.Error("Player[%v] personal space leave message[%v] was not player[%v] sent", player_id, msg_id, send_player_id)
		return false
	}
	delete(player_msgs, msg_id)
	return true
}

func (this *PSLeaveMessageMgr) GetSome(player_id int32, leave_msg_ids []int32) (leave_msgs []*rpc_common.H2R_PSLeaveMessageData) {
	this.locker.RLock()
	defer this.locker.RUnlock()

	player_msgs := this.player2msgs[player_id]
	if player_msgs == nil {
		return
	}

	for i := 0; i < len(leave_msg_ids); i++ {
		msg := player_msgs[leave_msg_ids[i]]
		if msg == nil {
			continue
		}
		// 获取评论
		comments := msg.GetSomeComments(0, PERSONAL_SPACE_GET_COMMNET_NUM)
		d := &rpc_common.H2R_PSLeaveMessageData{
			Id:           msg.id,
			Content:      msg.content,
			SendPlayerId: msg.send_player_id,
			SendTime:     msg.send_time,
			Comments:     comments,
		}

		leave_msgs = append(leave_msgs, d)
	}

	return
}

func (this *PSLeaveMessageMgr) get_leave_msg(player_id, msg_id int32) (*PSLeaveMessage, int32) {
	player_msgs := this.player2msgs[player_id]
	if player_msgs == nil {
		return nil, int32(msg_client_message.E_ERR_PERSONAL_SPACE_NOT_GET_YET)
	}

	msg := player_msgs[msg_id]
	if msg == nil {
		log.Error("player[%v] no leave msg_id[%v]", player_id, msg_id)
		return nil, int32(msg_client_message.E_ERR_PERSONAL_SPACE_NOT_FOUND_THE_LEAVE_MSG)
	}
	return msg, 0
}

func (this *PSLeaveMessageMgr) GetLeaveMsgComments(player_id, msg_id, start_index, comment_num int32) (comments []*rpc_common.H2R_PSLeaveMessageCommentData, err int32) {
	this.locker.RLock()
	defer this.locker.RUnlock()

	var msg *PSLeaveMessage
	msg, err = this.get_leave_msg(player_id, msg_id)
	if err < 0 {
		log.Error("player[%v] no leave msg_id[%v]", player_id, msg_id)
		return
	}

	comments = msg.GetSomeComments(start_index, comment_num)
	err = 1
	return
}

func (this *PSLeaveMessageMgr) AddLeaveMsgComment(player_id, msg_id, send_player_id int32, comment []byte) int32 {
	this.locker.Lock()
	defer this.locker.Unlock()

	leave_msg, err := this.get_leave_msg(player_id, msg_id)
	if err < 0 {
		return err
	}

	comment_id := leave_msg.AddNewComment(send_player_id, comment)

	return comment_id
}

func (this *PSLeaveMessageMgr) DeleteLeaveMsgComment(player_id, msg_id, comment_id, send_player_id int32) int32 {
	this.locker.Lock()
	defer this.locker.Unlock()

	msg, err := this.get_leave_msg(player_id, msg_id)
	if err < 0 {
		return err
	}

	if !msg.DeleteComment(comment_id, send_player_id) {
		log.Warn("Player[%v] not found comment[%v] in personal space leave message[%v]", player_id, comment_id, msg_id)
		return int32(msg_client_message.E_ERR_PERSONAL_SPACE_NO_COMMENT_WITH_LEAVE_MSG)
	}

	log.Debug("Player[%v] deleted leave msg", player_id)

	return 1
}

func (this *PSLeaveMessageMgr) ToLeaveMsgRedisData(player_id int32, msg_id int32) *RedisPSLeaveMessage {
	this.locker.RLock()
	defer this.locker.RUnlock()

	msg, err := this.get_leave_msg(player_id, msg_id)
	if err < 0 || msg == nil {
		return nil
	}

	return msg.ToRedisData()
}

func (this *PSLeaveMessageMgr) ToCommentRedisData(player_id, msg_id, comment_id int32) *RedisPSComment {
	this.locker.RLock()
	defer this.locker.RUnlock()

	msg, err := this.get_leave_msg(player_id, msg_id)
	if err < 0 {
		return nil
	}

	comment := msg.get_comment(comment_id)
	if comment == nil {
		return nil
	}

	return comment.ToRedisData()
}

// 个人空间
type PersonalSpace struct {
	signature         string  // 签名
	picture_ids       []int32 // 照片
	leave_msg_ids     []int32 // 留言
	curr_leave_msg_id int32   // 当前留言ID
	leave_msg_ds_type int32
	locker            *sync.RWMutex
}

func (this *PersonalSpace) Init() {
	this.picture_ids = make([]int32, PERSONAL_SPACE_MAX_PICTURE_NUM)
	this.locker = &sync.RWMutex{}
}

func (this *PersonalSpace) SetSignature(signature string) {
	this.locker.Lock()
	defer this.locker.Unlock()
	this.signature = signature
}

func (this *PersonalSpace) GetSignature() string {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.signature
}

func (this *PersonalSpace) GetNewPictureId() int32 {
	this.locker.Lock()
	defer this.locker.Unlock()
	for i := 0; i < len(this.picture_ids); i++ {
		if this.picture_ids[i] == 0 {
			this.picture_ids[i] = int32(i + 1)
			return int32(i + 1)
		}
	}
	return 0
}

func (this *PersonalSpace) HasPicture(id int32) bool {
	this.locker.RLock()
	defer this.locker.RUnlock()
	if id <= 0 || id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		return false
	}
	if this.picture_ids[id-1] == 0 {
		return false
	}
	return true
}

func (this *PersonalSpace) DeletePicture(id int32) bool {
	this.locker.Lock()
	defer this.locker.Unlock()
	if id <= 0 || id > PERSONAL_SPACE_MAX_PICTURE_NUM {
		return false
	}
	if this.picture_ids[id-1] == 0 {
		return false
	}
	this.picture_ids[id-1] = 0
	return true
}

func (this *PersonalSpace) GetLeaveMsgIds(start_index int32, msg_num int32) (leave_msg_ids []int32, is_more bool) {
	this.locker.RLock()
	defer this.locker.RUnlock()

	if this.leave_msg_ids == nil {
		return
	}

	if msg_num > PERSONAL_SPACE_GET_LEAVE_MSG_NUM {
		msg_num = PERSONAL_SPACE_GET_LEAVE_MSG_NUM
	}

	all_msgs := int32(len(this.leave_msg_ids))
	if msg_num > all_msgs {
		msg_num = all_msgs
	}

	if start_index+msg_num > all_msgs {
		msg_num = start_index + msg_num - all_msgs
	} else if start_index+msg_num < all_msgs {
		is_more = true
	}

	for i := int32(start_index); i < start_index+msg_num; i++ {
		idx := all_msgs - i - 1
		if idx < 0 || int(idx) >= len(this.leave_msg_ids) {
			break
		}
		m := this.leave_msg_ids[idx]
		leave_msg_ids = append(leave_msg_ids, m)
	}

	return
}

// 获取一个新的留言ID
func (this *PersonalSpace) GetNewLeaveMsgId() int32 {
	this.locker.Lock()
	defer this.locker.Unlock()

	if this.curr_leave_msg_id >= math.MaxInt32 {
		this.curr_leave_msg_id = 1
	} else {
		this.curr_leave_msg_id += 1
	}
	this.leave_msg_ids = append(this.leave_msg_ids, this.curr_leave_msg_id)
	return this.curr_leave_msg_id
}

// 仅仅删除留言ID
func (this *PersonalSpace) DeleteLeaveMsg(id int32) bool {
	this.locker.Lock()
	defer this.locker.Unlock()

	index := int32(-1)
	all_msgs := int32(len(this.leave_msg_ids))
	for i := int32(0); i < all_msgs; i++ {
		if this.leave_msg_ids[all_msgs-i-1] == id {
			index = all_msgs - i - 1
			break
		}
	}
	if index < 0 {
		return false
	}
	for i := index; i < all_msgs-1; i++ {
		this.leave_msg_ids[i] = this.leave_msg_ids[i+1]
	}
	this.leave_msg_ids = this.leave_msg_ids[:all_msgs-1]
	return true
}

// 个人空间管理器
type PersonalSpaceMgr struct {
	personal_space_map map[int32]*PersonalSpace
	locker             *sync.RWMutex
}

var ps_mgr PersonalSpaceMgr

func (this *PersonalSpaceMgr) Init() {
	this.personal_space_map = make(map[int32]*PersonalSpace)
	this.locker = &sync.RWMutex{}
}

func (this *PersonalSpaceMgr) CreateSpace(player_id int32) *PersonalSpace {
	this.locker.Lock()
	defer this.locker.Unlock()
	d := this.personal_space_map[player_id]
	if d != nil {
		log.Warn("Player[%v] Personal Space exists", player_id)
		return d
	}
	d = &PersonalSpace{}
	d.Init()
	this.personal_space_map[player_id] = d
	return d
}

func (this *PersonalSpaceMgr) GetSpace(player_id int32) *PersonalSpace {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.personal_space_map[player_id]
}

func (this *PersonalSpaceMgr) LoadSpace(data *RedisPSBaseData) bool {
	if this.personal_space_map[data.PlayerId] != nil {
		log.Warn("Already load player[%v] redis personal space")
		return false
	}
	d := &PersonalSpace{
		signature:         data.Signature,
		picture_ids:       data.PictureIds,
		leave_msg_ids:     data.LeaveMsgIds,
		curr_leave_msg_id: data.CurrLeaveMsgId,
	}
	d.Init()
	this.personal_space_map[data.PlayerId] = d
	return true
}

func (this *PersonalSpaceMgr) ToRedisData(player_id int32) *RedisPSBaseData {
	this.locker.RLock()
	defer this.locker.RUnlock()
	ps := this.personal_space_map[player_id]
	if ps == nil {
		return nil
	}
	return &RedisPSBaseData{
		PlayerId:       player_id,
		Signature:      ps.signature,
		PictureIds:     ps.picture_ids,
		LeaveMsgIds:    ps.leave_msg_ids,
		CurrLeaveMsgId: ps.curr_leave_msg_id,
		LeaveMsgDsType: ps.leave_msg_ds_type,
	}
}
