package main

import (
	"libs/log"
	_ "libs/utils"
	_ "main/rpc_common"
	_ "main/table_config"
	"math/rand"
	_ "net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	"sync"
	"sync/atomic"
	"time"

	_ "3p/code.google.com.protobuf/proto"
	_ "github.com/golang/protobuf/proto"
)

//const FRIEND_UNREAD_MESSAGE_MAX_NUM int = 200
//const FRIEND_MESSAGE_MAX_LENGTH int = 200

const MAX_FRIEND_RECOMMEND_PLAYER_NUM int32 = 10000

type FriendRecommendMgr struct {
	player_ids    map[int32]int32
	players_array []int32
	locker        *sync.RWMutex
	add_chan      chan int32
	to_end        int32
}

var friend_recommend_mgr FriendRecommendMgr

func (this *FriendRecommendMgr) Init() {
	this.player_ids = make(map[int32]int32)
	this.players_array = make([]int32, MAX_FRIEND_RECOMMEND_PLAYER_NUM)
	this.locker = &sync.RWMutex{}
	this.add_chan = make(chan int32, 10000)
	this.to_end = 0
}

func (this *FriendRecommendMgr) AddPlayer(player_id int32) {
	this.add_chan <- player_id
	log.Debug("Friend Recommend Manager to add player[%v]", player_id)
}

func (this *FriendRecommendMgr) CheckAndAddPlayer(player_id int32) bool {
	p := player_mgr.GetPlayerById(player_id)
	if p == nil {
		return false
	}

	if _, o := this.player_ids[player_id]; o {
		log.Warn("Player[%v] already added Friend Recommend mgr", player_id)
		return false
	}

	var add_pos int32
	num := int32(len(this.player_ids))
	if num >= MAX_FRIEND_RECOMMEND_PLAYER_NUM {
		add_pos = rand.Int31n(num)
		// 删掉一个随机位置的
		delete(this.player_ids, this.players_array[add_pos])
		this.players_array[add_pos] = 0
	} else {
		add_pos = num
	}

	now_time := int32(time.Now().Unix())
	if now_time-p.db.Info.GetLastLogout() > 24*3600*2 && p.is_logout {
		return false
	}

	if p.db.Friends.NumAll() >= global_config.FriendMaxNum {
		return false
	}

	this.player_ids[player_id] = add_pos
	this.players_array[add_pos] = player_id

	log.Debug("Friend Recommend Manager add player[%v], total count[%v], player_ids: %v, players_array: %v", player_id, len(this.player_ids), this.player_ids, this.players_array[:len(this.player_ids)])

	return true
}

func (this *FriendRecommendMgr) Run() {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	last_check_remove_time := int32(time.Now().Unix())
	for {
		if atomic.LoadInt32(&this.to_end) > 0 {
			break
		}
		// 处理操作队列
		is_break := false
		for !is_break {
			select {
			case player_id, ok := <-this.add_chan:
				{
					if !ok {
						log.Error("conn timer wheel op chan receive invalid !!!!!")
						return
					}
					this.CheckAndAddPlayer(player_id)
				}
			default:
				{
					is_break = true
				}
			}
		}

		now_time := int32(time.Now().Unix())
		if now_time-last_check_remove_time >= 60*10 {
			this.locker.Lock()
			player_num := len(this.player_ids)
			for i := 0; i < player_num; i++ {
				p := player_mgr.GetPlayerById(this.players_array[i])
				if p == nil {
					continue
				}
				if (now_time-p.db.Info.GetLastLogout() >= 2*24*3600 && p.is_logout) || p.db.Friends.NumAll() >= global_config.FriendMaxNum {
					delete(this.player_ids, this.players_array[i])
					this.players_array[i] = this.players_array[player_num-1]
					player_num -= 1
				}
			}
			this.locker.Unlock()
			last_check_remove_time = now_time
		}

		time.Sleep(time.Second * 1)
	}
}

func (this *FriendRecommendMgr) Random(player_id int32) (ids []int32) {
	this.locker.RLock()
	defer this.locker.RUnlock()

	cnt := int32(len(this.player_ids))
	if cnt == 0 {
		return
	}

	if cnt > global_config.FriendRecommendNum {
		cnt = global_config.FriendRecommendNum
	}
	rand.Seed(time.Now().Unix())
	for i := int32(0); i < cnt; i++ {
		r := rand.Int31n(int32(len(this.player_ids)))
		sr := r
		for {
			has := false
			if this.players_array[sr] == player_id {
				has = true
			} else {
				if ids != nil {
					for n := 0; n < len(ids); n++ {
						if ids[n] == this.players_array[sr] {
							has = true
							break
						}
					}
				}
			}
			if !has {
				break
			}
			sr = (sr + 1) % int32(len(this.player_ids))
			if sr == r {
				log.Info("Friend Recommend Mgr player count[%v] not enough to random a player to recommend", len(this.player_ids))
				return
			}
		}
		pid := this.players_array[sr]
		if pid <= 0 {
			break
		}
		ids = append(ids, pid)
	}
	return ids
}

// ----------------------------------------------------------------------------

func _format_friend_info(p *Player, now_time int32) (friend_info *msg_client_message.FriendInfo) {
	offline_seconds := int32(0)
	if p.is_logout {
		offline_seconds = now_time - p.db.Info.GetLastLogout()
	}
	friend_info = &msg_client_message.FriendInfo{
		Id:             p.Id,
		Name:           p.db.GetName(),
		Level:          p.db.Info.GetLvl(),
		IsOnline:       !p.is_logout,
		OfflineSeconds: offline_seconds,
	}
	return
}

func _format_friends_info(friend_ids []int32) (friends_info []*msg_client_message.FriendInfo) {
	if friend_ids == nil || len(friend_ids) == 0 {
		friends_info = make([]*msg_client_message.FriendInfo, 0)
	} else {
		now_time := int32(time.Now().Unix())
		for i := 0; i < len(friend_ids); i++ {
			p := player_mgr.GetPlayerById(friend_ids[i])
			if p == nil {
				continue
			}
			player := _format_friend_info(p, now_time)
			friends_info = append(friends_info, player)
		}
	}
	return
}

func _format_ask_friends_info(friend_ids []int32) (ask_friends_info []*msg_client_message.AskFriendInfo) {
	if friend_ids == nil || len(friend_ids) == 0 {
		ask_friends_info = make([]*msg_client_message.AskFriendInfo, 0)
	} else {
		for i := 0; i < len(friend_ids); i++ {
			p := player_mgr.GetPlayerById(friend_ids[i])
			if p == nil {
				continue
			}

			player := &msg_client_message.AskFriendInfo{
				Id:    friend_ids[i],
				Name:  p.db.GetName(),
				Level: p.db.Info.GetLvl(),
			}
			ask_friends_info = append(ask_friends_info, player)
		}
	}
	return
}

// 好友推荐列表
func (this *Player) send_recommend_friends() int32 {
	player_ids := friend_recommend_mgr.Random(this.Id)
	players := _format_friends_info(player_ids)
	response := &msg_client_message.S2CFriendRecommendResponse{
		Players: players,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_FRIEND_RECOMMEND_RESPONSE), response)
	log.Debug("Player[%v] recommend friends %v", this.Id, response)
	return 1
}

// 好友列表
func (this *Player) send_friend_list() int32 {
	friend_ids := this.db.Friends.GetAllIndex()
	friends := _format_friends_info(friend_ids)
	response := &msg_client_message.S2CFriendListResponse{
		Friends: friends,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_FRIEND_LIST_RESPONSE), response)
	log.Debug("Player[%v] friend list: %v", this.Id, response)
	return 1
}

// 检测是否好友增加
func (this *Player) check_and_send_friend_add() int32 {
	if this.friend_add == nil || len(this.friend_add) == 0 {
		return 0
	}
	friends := _format_friends_info(this.friend_add)
	this.friend_add = nil
	response := &msg_client_message.S2CFriendListAddNotify{
		FriendsAdd: friends,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_FRIEND_LIST_ADD_NOTIFY), response)
	log.Debug("Player[%v] friend add: %v", this.Id, response)
	return 1
}

// 申请好友
func (this *Player) friend_ask(player_id int32) int32 {
	p := player_mgr.GetPlayerById(player_id)
	if p == nil {
		log.Error("Player[%v] not found", player_id)
		return int32(msg_client_message.E_ERR_PLAYER_NOT_EXIST)
	}

	if this.db.Friends.HasIndex(player_id) {
		log.Error("Player[%v] already add player[%v] to friend", this.Id, player_id)
		return int32(msg_client_message.E_ERR_PLAYER_FRIEND_ALREADY_ADD)
	}

	if p.db.FriendAsks.HasIndex(this.Id) {
		log.Error("Player[%v] already asked player[%v] to friend", this.Id, player_id)
		return int32(msg_client_message.E_ERR_PLAYER_FRIEND_ALREADY_ASKED)
	}

	p.db.FriendAsks.Add(&dbPlayerFriendAskData{
		PlayerId: this.Id,
	})

	response := &msg_client_message.S2CFriendAskResponse{
		PlayerId: player_id,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_FRIEND_ASK_RESPONSE), response)

	this.friend_ask_add = append(this.friend_ask_add, player_id)

	log.Debug("Player[%v] asked player[%v] to friend", this.Id, player_id)

	return 1
}

// 检测好友申请增加
func (this *Player) check_and_send_friend_ask_add() int32 {
	if this.friend_ask_add == nil || len(this.friend_ask_add) == 0 {
		return 0
	}
	players := _format_ask_friends_info(this.friend_ask_add)
	this.friend_ask_add = nil

	response := &msg_client_message.S2CFriendAskPlayerListAddNotify{
		PlayersAdd: players,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_FRIEND_ASK_PLAYER_LIST_ADD_NOTIFY), response)
	log.Debug("Player[%v] checked friend ask add %v", this.Id, response)
	return 1
}

// 好友申请列表
func (this *Player) send_friend_ask_list() int32 {
	friend_ask_ids := this.db.FriendAsks.GetAllIndex()
	players := _format_ask_friends_info(friend_ask_ids)
	response := &msg_client_message.S2CFriendAskPlayerListResponse{
		Players: players,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_FRIEND_ASK_PLAYER_LIST_RESPONSE), response)
	log.Debug("Player[%v] friend ask list %v", this.Id, response)
	return 1
}

// 同意加为好友
func (this *Player) agree_friend_ask(player_ids []int32) int32 {
	for i := 0; i < len(player_ids); i++ {
		p := player_mgr.GetPlayerById(player_ids[i])
		if p == nil {
			log.Error("Player[%v] not found on agree friend ask", player_ids[i])
			return int32(msg_client_message.E_ERR_PLAYER_NOT_EXIST)
		}
		if !this.db.FriendAsks.HasIndex(player_ids[i]) {
			log.Error("Player[%v] friend ask list not player[%v]", this.Id, player_ids[i])
			return int32(msg_client_message.E_ERR_PLAYER_FRIEND_PLAYER_NO_IN_ASK_LIST)
		}
	}

	for i := 0; i < len(player_ids); i++ {
		this.db.Friends.Add(&dbPlayerFriendData{
			PlayerId: player_ids[i],
		})
		this.db.FriendAsks.Remove(player_ids[i])
	}

	response := &msg_client_message.S2CFriendAgreeResponse{
		Friends: _format_friends_info(player_ids),
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_FRIEND_AGREE_RESPONSE), response)

	this.friend_add = append(this.friend_add, player_ids...)

	log.Debug("Player[%v] agreed players[%v] friend ask", this.Id, player_ids)
	return 1
}

// 拒绝申请
func (this *Player) refuse_friend_ask(player_id int32) int32 {
	if !this.db.FriendAsks.HasIndex(player_id) {
		log.Error("Player[%v] ask list no player[%v]", this.Id, player_id)
		return int32(msg_client_message.E_ERR_PLAYER_FRIEND_PLAYER_NO_IN_ASK_LIST)
	}

	this.db.FriendAsks.Remove(player_id)
	response := &msg_client_message.S2CFriendRefuseResponse{
		PlayerId: player_id,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_FRIEND_REFUSE_RESPONSE), response)

	log.Debug("Player[%v] refuse player[%v] friend ask", this.Id, player_id)

	return 1
}

func (this *Player) remove_friend(friend_ids []int32) int32 {
	for i := 0; i < len(friend_ids); i++ {
		if !this.db.Friends.HasIndex(friend_ids[i]) {
			log.Error("Player[%v] no friend[%v]", this.Id, friend_ids[i])
			return int32(msg_client_message.E_ERR_PLAYER_FRIEND_NOT_FOUND)
		}
	}

	for i := 0; i < len(friend_ids); i++ {
		this.db.Friends.Remove(friend_ids[i])
	}

	response := &msg_client_message.S2CFriendRemoveResponse{
		PlayerIds: friend_ids,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_FRIEND_REMOVE_RESPONSE), response)

	log.Debug("Player[%v] removed friends: %v", this.Id, friend_ids)

	return 1
}

func (this *Player) refresh_friend_give_points(friend_id int32) bool {
	return true
}

func (this *Player) check_friends_give_points_refresh() (remain_seconds int32) {
	friends := this.db.Friends.GetAllIndex()
	if friends == nil || len(friends) <= 0 {
		return
	}

	//rt := &global_config.FriendGivePointsRefreshTime
	//remain_seconds = utils.GetRemainSeconds4NextRefresh(rt.Hour, rt.Minute, rt.Second, this.db.FriendRelative.GetLastRefreshTime())

	//if remain_seconds <= 0 {
	/*for i := 0; i < len(friends); i++ {
		friend := player_mgr.GetPlayerById(friends[i])
		if friend != nil {
			friend.refresh_friend_give_points(this.Id)
		} else {
			result := this.rpc_call_refresh_give_friend_point(friends[i])
			if result == nil {
				log.Error("Player[%v] to refresh friend[%v] give points error", this.Id, friends[i])
			}
		}
	}*/
	//this.db.FriendRelative.SetLastRefreshTime(int32(time.Now().Unix()))
	//this.db.FriendRelative.SetGiveNumToday(0)
	//}

	return
}

func (this *Player) get_friend_list(get_foster bool) int32 {
	//remain_seconds := this.check_friends_give_points_refresh()

	response := &msg_client_message.S2CFriendListResponse{}
	this.Send(1, response)
	return 1
}

func (this *Player) store_friend_points(friend_id int32) (err int32, last_save int32, remain_seconds int32) {

	return
}

func (this *Player) give_friend_points(friend_list []int32) int32 {
	this.check_friends_give_points_refresh()

	return 1
}

func (this *Player) get_friend_points(friend_list []int32) int32 {
	this.check_friends_give_points_refresh()

	return 1
}

func (this *Player) friend_chat_add(friend_id int32, message []byte) int32 {
	// 未读消息数量
	/*is_full, next_id := this.db.FriendChatUnreadIds.CheckUnreadNumFull(friend_id)
	if is_full {
		log.Debug("Player[%v] chat message from friend[%v] is full", this.Id, friend_id)
		return int32(msg_client_message.E_ERR_FRIEND_MESSAGE_NUM_MAX)
	}

	new_long_id := utils.Int64From2Int32(friend_id, next_id)
	message_data := &dbPlayerFriendChatUnreadMessageData{
		PlayerMessageId: new_long_id,
		Message:         message,
		SendTime:        int32(time.Now().Unix()),
		IsRead:          int32(0),
	}

	if !this.db.FriendChatUnreadMessages.Add(message_data) {
		log.Error("Player[%v] add friend[%v] chat message failed", this.Id, friend_id)
		return -1
	}

	res := this.db.FriendChatUnreadIds.AddNewMessageId(friend_id, next_id)
	if res < 0 {
		// 增加新ID失败则删除刚加入的消息
		this.db.FriendChatUnreadMessages.Remove(new_long_id)
		log.Error("Player[%v] add new message id[%v,%v] from friend[%v] failed", this.Id, next_id, new_long_id)
		return res
	}

	log.Debug("Player[%v] add friend[%v] chat message[id:%v, long_id:%v, content:%v]", this.Id, friend_id, next_id, new_long_id, message)*/

	return 1
}

func (this *Player) friend_chat(friend_id int32, message []byte) int32 {
	/*if !this.db.Friends.HasIndex(friend_id) {
		log.Error("Player[%v] no friend[%v], chat failed", this.Id, friend_id)
		return int32(msg_client_message.E_ERR_FRIEND_NO_THE_FRIEND)
	}

	if len(message) > FRIEND_MESSAGE_MAX_LENGTH {
		log.Error("Player[%v] from friend[%v] chat content is too long[%v]", this.Id, friend_id, len(message))
		return int32(msg_client_message.E_ERR_FRIEND_MESSAGE_TOO_LONG)
	}

	friend := player_mgr.GetPlayerById(friend_id)
	if friend != nil {
		res := friend.friend_chat_add(this.Id, message)
		if res < 0 {
			return res
		}
	} else {
		result := this.rpc_friend_chat(friend_id, message)
		if result == nil {
			log.Error("Player[%v] chat message[%v] to friend[%v] failed", this.Id, message, friend_id)
			return int32(msg_client_message.E_ERR_FRIEND_CHAT_FAILED)
		}
		if result.Error < 0 {
			log.Error("Player[%v] chat message[%v] to friend[%v] error[%v]", this.Id, message, friend_id, result.Error)
			return result.Error
		}
	}

	response := &msg_client_message.S2CFriendChatResult{}
	response.PlayerId = proto.Int32(friend_id)
	response.Content = message
	this.Send(response)*/

	return 1
}

func (this *Player) friend_get_unread_message_num(friend_ids []int32) int32 {
	/*data := make([]*msg_client_message.FriendUnreadMessageNumData, len(friend_ids))
	for i := 0; i < len(friend_ids); i++ {
		message_num := int32(0)
		if !this.db.Friends.HasIndex(friend_ids[i]) {
			message_num = int32(msg_client_message.E_ERR_FRIEND_NO_THE_FRIEND)
			log.Error("Player[%v] no friend[%v], get unread message num failed", this.Id, friend_ids[i])
		} else {
			message_num = this.db.FriendChatUnreadIds.GetUnreadMessageNum(friend_ids[i])
		}
		data[i] = &msg_client_message.FriendUnreadMessageNumData{
			FriendId:   proto.Int32(friend_ids[i]),
			MessageNum: proto.Int32(message_num),
		}
	}

	response := &msg_client_message.S2CFriendGetUnreadMessageNumResult{}
	response.Data = data
	this.Send(response)*/
	return 1
}

func (this *Player) friend_pull_unread_message(friend_id int32) int32 {
	/*if !this.db.Friends.HasIndex(friend_id) {
		log.Error("Player[%v] no friend[%v], pull unread message failed", this.Id, friend_id)
		return int32(msg_client_message.E_ERR_FRIEND_NO_THE_FRIEND)
	}

	c := 0
	var data []*msg_client_message.FriendChatData
	all_unread_ids, o := this.db.FriendChatUnreadIds.GetMessageIds(friend_id)
	if !o || all_unread_ids == nil || len(all_unread_ids) == 0 {
		data = make([]*msg_client_message.FriendChatData, 0)
	} else {
		data = make([]*msg_client_message.FriendChatData, len(all_unread_ids))
		for i := 0; i < len(all_unread_ids); i++ {
			long_id := utils.Int64From2Int32(friend_id, all_unread_ids[i])
			content, o := this.db.FriendChatUnreadMessages.GetMessage(long_id)
			if !o {
				log.Warn("Player[%v] no unread message[%v] from friend[%v]", this.Id, all_unread_ids[i], friend_id)
				continue
			}
			send_time, _ := this.db.FriendChatUnreadMessages.GetSendTime(long_id)
			data[c] = &msg_client_message.FriendChatData{
				Content:  content,
				SendTime: proto.Int32(send_time),
			}
			c += 1
		}
	}

	response := &msg_client_message.S2CFriendPullUnreadMessageResult{}
	response.Data = data[:c]
	response.FriendId = proto.Int32(friend_id)
	this.Send(response)

	log.Debug("Player[%v] pull unread message[%v] from friend[%v]", this.Id, response.Data, friend_id)*/

	return 1
}

func (this *Player) friend_confirm_unread_message(friend_id int32, message_num int32) int32 {
	/*if !this.db.Friends.HasIndex(friend_id) {
		log.Error("Player[%v] no friend[%v], confirm unread message failed", this.Id, friend_id)
		return int32(msg_client_message.E_ERR_FRIEND_NO_THE_FRIEND)
	}

	res, remove_ids := this.db.FriendChatUnreadIds.ConfirmUnreadIds(friend_id, message_num)
	if res < 0 {
		return res
	}

	this.db.FriendChatUnreadMessages.RemoveMessages(friend_id, remove_ids)

	response := &msg_client_message.S2CFriendConfirmUnreadMessageResult{}
	response.FriendId = proto.Int32(friend_id)
	response.MessageNum = proto.Int32(message_num)
	this.Send(response)*/

	log.Debug("Player[%v] confirm friend[%v] unread message num[%v]", this.Id, friend_id, message_num)

	return 1
}

// ------------------------------------------------------
