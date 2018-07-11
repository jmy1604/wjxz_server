package main

import (
	"libs/log"
	"libs/utils"
	"main/rpc_common"
	_ "main/table_config"
	"math/rand"
	_ "net/http"
	"public_message/gen_go/client_message"
	"sync"
	"sync/atomic"
	"time"

	_ "3p/code.google.com.protobuf/proto"
	_ "github.com/golang/protobuf/proto"
)

const FRIEND_UNREAD_MESSAGE_MAX_NUM int = 200
const FRIEND_MESSAGE_MAX_LENGTH int = 200

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
	this.add_chan = make(chan int32, 100)
	this.to_end = 0
}

func (this *FriendRecommendMgr) CheckAndAddPlayer(player_id int32) bool {
	p := player_mgr.GetPlayerById(player_id)
	if p == nil {
		return false
	}

	if _, o := this.player_ids[player_id]; o {
		log.Warn("Player[%v] already added Friend Recommend mgr")
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
	if now_time-p.db.Info.GetLastLogout() > 24*3600*2 {
		return false
	}

	if p.db.Friends.NumAll() >= global_config.FriendMaxNum {
		return false
	}

	this.player_ids[player_id] = add_pos
	this.players_array[add_pos] = player_id

	log.Debug("Friend Recommend Manager add player[%v]", player_id)

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
			player_num := len(this.player_ids)
			for i := 0; i < player_num; i++ {
				p := player_mgr.GetPlayerById(this.players_array[i])
				if p == nil {
					continue
				}
				if now_time-p.db.Info.GetLastLogout() >= 2*24*3600 {
					delete(this.player_ids, this.players_array[i])
					this.players_array[i] = this.players_array[player_num-1]
					player_num -= 1
				}
			}
			last_check_remove_time = now_time
		}

		time.Sleep(time.Second * 1)
	}
}

// ----------------------------------------------------------------------------

func (this *dbPlayerFriendColumn) FillAllListMsg(msg *msg_client_message.S2CFriendListResponse) {
	//var tmp_info *msg_client_message.FriendInfo
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.FillAllListMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()
	msg.Friends = make([]*msg_client_message.FriendInfo, 0, len(this.m_data))
	for _, val := range this.m_data {
		if nil == val {
			continue
		}

		/*tmp_info = &msg_client_message.FriendInfo{}
		tmp_info.PlayerId = val.FriendId
		tmp_info.Name = val.FriendName
		tmp_info.Level = val.Level
		tmp_info.VipLevel = val.VipLevel
		tmp_info.LastLogin = val.LastLogin
		tmp_info.Head = val.Head
		tmp_info.IsOnline = true
		log.Info("附加值到好友列表 %v", tmp_info)
		msg.FriendList = append(msg.FriendList, tmp_info)*/
	}

	return
}

func (this *dbPlayerFriendColumn) GetAviFriendId() int32 {
	return 0
}

func (this dbPlayerFriendColumn) TryAddFriend(new_friend *dbPlayerFriendData) {
	if nil == new_friend {
		log.Error("dbPlayerFriendColumn TryAddFriend ")
		return
	}

	this.m_row.m_lock.UnSafeLock("dbPlayerFriendColumn.TryAddFriend")
	defer this.m_row.m_lock.UnSafeUnlock()

	if nil == this.m_data[new_friend.PlayerId] {
		this.m_data[new_friend.PlayerId] = new_friend
		this.m_changed = true
	}

	return
}

func (this *dbPlayerFriendReqColumn) FillAllListMsg(msg *msg_client_message.S2CFriendListResponse) {

	/*var tmp_info *msg_client_message.FriendReq
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendReqColumn.FillAllListMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()

	msg.Reqs = make([]*msg_client_message.FriendReq, 0, len(this.m_data))
	for _, val := range this.m_data {
		if nil == val {
			continue
		}

		tmp_info = &msg_client_message.FriendReq{}
		tmp_info.PlayerId = val.PlayerId
		tmp_info.Name = val.PlayerName
		msg.Reqs = append(msg.Reqs, tmp_info)
	}*/

	return
}

func send_search_player_msg(p *Player, players_info []*rpc_common.H2R_SearchPlayerInfo) {
	var results []*msg_client_message.FriendInfo
	if players_info == nil || len(players_info) == 0 {
		results = make([]*msg_client_message.FriendInfo, 0)
	} else {
		results = make([]*msg_client_message.FriendInfo, len(players_info))
		for i := 0; i < len(players_info); i++ {
			r := &msg_client_message.FriendInfo{
			/*PlayerId:  players_info[i].Id,
			Name:      players_info[i].Nick,
			Head:      players_info[i].Head,
			Level:     players_info[i].Level,
			VipLevel:  players_info[i].VipLevel,
			LastLogin: players_info[i].LastLogin,*/
			}
			results[i] = r
		}

	}
	/*response := msg_client_message.S2CFriendSearchResult{}
	response.Result = results
	p.Send(&response)*/
}

func (this *dbPlayerFriendReqColumn) CheckAndAdd(player_id int32, player_name string) int32 {
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendReqColumn.CheckAndAdd")
	defer this.m_row.m_lock.UnSafeUnlock()

	d := this.m_data[player_id]
	if d != nil {
		log.Warn("!!! Player[%v,%v] already in request list of player[%v]", player_id, player_name, this.m_row.GetPlayerId())
		return int32(msg_client_message.E_ERR_FRIEND_THE_PLAYER_REQUESTED)
	}

	d = &dbPlayerFriendReqData{}
	d.PlayerId = player_id
	this.m_data[player_id] = d
	this.m_changed = true
	return 1
}

func (this *dbPlayerFriendReqColumn) AgreeFriend(friend_id int32) bool {
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendReqColumn.AgreeFriend")
	defer this.m_row.m_lock.UnSafeUnlock()

	d := this.m_data[friend_id]
	if d != nil {

	}
	return true
}

func (this *dbPlayerFriendColumn) GetAllIds() (ret_ids []int32) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.GetAllIds")
	defer this.m_row.m_lock.UnSafeRUnlock()
	tmp_len := len(this.m_data)
	if tmp_len <= 0 {
		return nil
	}

	ret_ids = make([]int32, 0, len(this.m_data))
	for _, v := range this.m_data {
		ret_ids = append(ret_ids, v.PlayerId)
	}
	return
}

func (this *dbPlayerFriendColumn) GetFriendInfoMsg(friend_id int32) *msg_client_message.FriendInfo {
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.GetMsgFriendInfo")
	defer this.m_row.m_lock.UnSafeRUnlock()

	d := this.m_data[friend_id]
	if d == nil {
		return nil
	}

	return &msg_client_message.FriendInfo{
	/*PlayerId:  d.FriendId,
	Name:      d.FriendName,
	Head:      d.Head,
	Level:     d.Level,
	VipLevel:  d.VipLevel,
	LastLogin: d.LastLogin,*/
	}
}

func (this *dbPlayerFriendChatUnreadIdColumn) CheckUnreadNumFull(friend_id int32) (full bool, next_id int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadIdColumn.CheckUnreadNumFull")
	defer this.m_row.m_lock.UnSafeUnlock()

	d := this.m_data[friend_id]
	if d == nil {
		next_id = 1
		this.m_changed = true
	} else if len(d.MessageIds) < FRIEND_UNREAD_MESSAGE_MAX_NUM {
		if d.CurrMessageId >= int32(FRIEND_UNREAD_MESSAGE_MAX_NUM) {
			d.CurrMessageId = 1
		} else {
			d.CurrMessageId += 1
		}
		next_id = d.CurrMessageId
		this.m_changed = true
	} else {
		full = true
	}

	return
}

func (this *dbPlayerFriendChatUnreadIdColumn) AddNewMessageId(friend_id, message_id int32) int32 {
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadIdColumn.AddNewMessageId")
	defer this.m_row.m_lock.UnSafeUnlock()

	d := this.m_data[friend_id]
	if d == nil {
		this.m_data[friend_id] = &dbPlayerFriendChatUnreadIdData{
			FriendId:   friend_id,
			MessageIds: []int32{message_id},
		}
	} else {
		if len(d.MessageIds) >= FRIEND_UNREAD_MESSAGE_MAX_NUM {
			return int32(msg_client_message.E_ERR_FRIEND_MESSAGE_NUM_MAX)
		}
		d.MessageIds = append(d.MessageIds, message_id)
	}

	this.m_changed = true

	return 1
}

func (this *dbPlayerFriendChatUnreadIdColumn) GetUnreadMessageNum(friend_id int32) int32 {
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadIdColumn.GetUnreadMessageNum")
	defer this.m_row.m_lock.UnSafeRUnlock()

	d := this.m_data[friend_id]
	if d == nil {
		return 0
	}

	return int32(len(d.MessageIds))
}

func (this *dbPlayerFriendChatUnreadIdColumn) ConfirmUnreadIds(friend_id, unread_num int32) (res int32, remove_ids []int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadIdColumn.ConfirmUnreadIds")
	defer this.m_row.m_lock.UnSafeUnlock()

	d := this.m_data[friend_id]
	if d == nil {
		log.Error("Player[%v] no unread message from friend[%v]", this.m_row.m_PlayerId, friend_id)
		res = int32(msg_client_message.E_ERR_FRIEND_NO_UNREAD_MESSAGE)
		return
	}

	if unread_num == 0 || len(d.MessageIds) <= int(unread_num) {
		remove_ids = d.MessageIds
		d.MessageIds = make([]int32, 0)
	} else {
		remove_ids = d.MessageIds[:unread_num]
		d.MessageIds = d.MessageIds[unread_num:]
	}

	this.m_changed = true

	res = 1
	return
}

func (this *dbPlayerFriendChatUnreadMessageColumn) RemoveMessages(friend_id int32, message_ids []int32) {
	if message_ids == nil || len(message_ids) == 0 {
		return
	}

	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadMessageColumn.RemoveMessages")
	defer this.m_row.m_lock.UnSafeUnlock()

	for i := 0; i < len(message_ids); i++ {
		player_message_id := utils.Int64From2Int32(friend_id, message_ids[i])
		delete(this.m_data, player_message_id)
	}
	this.m_changed = true
}

func (this *Player) search_friend(key string) int32 {
	result := this.rpc_search_friend_by_key(key)
	if result == nil {
		log.Error("查找玩家[%v]数据失败", key)
		return -1
	}

	send_search_player_msg(this, result.Players)

	log.Info("Player[%v] searched friend with key[%v]", this.Id, key)

	return 1
}

func (this *Player) add_friend_by_name(name string) int32 {
	result := this.rpc_add_friend_by_name(name)
	if result == nil {
		log.Error("Player[%v] request add friend[%v] failed", this.Id, name)
		return -1
	}

	/*response := &msg_client_message.S2CAddFriendResult{}
	response.PlayerId = proto.Int32(result.AddPlayerId)
	this.Send(response)*/

	log.Info("Player[%v] requested add friend[%v]", this.Id, name)

	return 1
}

func (this *Player) add_friend_by_id(id int32) int32 {
	// 已是好友
	if this.db.Friends.HasIndex(id) {
		log.Error("!!! Player[%v] already added player[%v] to friend", this.Id, id)
		return int32(msg_client_message.E_ERR_FRIEND_THE_PLAYER_ALREADY_FRIEND)
	}

	add_player := player_mgr.GetPlayerById(id)
	if add_player != nil {
		res := add_player.db.FriendReqs.CheckAndAdd(this.Id, this.db.GetName())
		if res < 0 {
			log.Error("!!! Player[%v] request add friend to other player[%v] already exist", this.Id, id)
			return res
		}
	} else {
		// rpc调用
		result := this.rpc_add_friend(id)
		if result == nil {
			log.Error("!!! Player[%v] request add friend to other player[%v] failed", this.Id, id)
			return -1
		}
		if result.Error < 0 {
			return result.Error
		}
	}

	/*response := &msg_client_message.S2CAddFriendResult{}
	response.PlayerId = proto.Int32(id)
	this.Send(response)*/

	log.Info("Player[%v] requested add friend[%v]", this.Id, id)

	return 1
}

func (this *Player) agree_add_friend(id int32) int32 {
	// 该玩家已是好友
	if this.db.Friends.HasIndex(id) {
		log.Error("Player[%v] already have friend[%v]", this.Id, id)
		return -1
	}

	if !this.db.FriendReqs.HasIndex(id) {
		log.Error("Player[%v] no player[%v]'s friend request", this.Id, id)
		return -1
	}

	var data dbPlayerFriendData
	agree_player := player_mgr.GetPlayerById(id)
	if agree_player == nil {
		result := this.rpc_agree_add_friend(id, true)
		if result == nil {
			log.Error("Player[%v] agree add friend with player[%v] failed", this.Id, id)
			return -1
		}

		// 加到自己的好友列表
		data.PlayerId = id
	} else {
		// 加到对方的好友列表
		data.PlayerId = this.Id
		agree_player.db.Friends.Add(&data)

		// 加到自己的好友列表
		data.PlayerId = id
	}

	this.db.Friends.Add(&data)

	// request remove
	this.db.FriendReqs.Remove(id)

	/*response := &msg_client_message.S2CAgreeFriendResult{}
	response.PlayerId = proto.Int32(id)
	response.Name = proto.String(data.FriendName)
	this.Send(response)*/

	log.Debug("Player[%v] agree add friend request of player[%v][%v]", this.Id, id)

	return 1
}

func (this *Player) refuse_add_friend(player_id int32) int32 {

	return 1
}

func (this *Player) remove_friend_data(friend_id int32) {
	this.db.Friends.Remove(friend_id)
	this.db.FriendPoints.Remove(friend_id)
	message_ids, o := this.db.FriendChatUnreadIds.GetMessageIds(friend_id)
	if o {
		this.db.FriendChatUnreadMessages.RemoveMessages(friend_id, message_ids)
		this.db.FriendChatUnreadIds.Remove(friend_id)
	}
}

func (this *Player) remove_friend(player_id int32) int32 {
	if !this.db.Friends.HasIndex(player_id) {
		log.Error("Player[%v] have not friend[%v], remove failed", this.Id, player_id)
		return -1
	}

	p := player_mgr.GetPlayerById(player_id)
	if p == nil {
		result := this.rpc_remove_friend2(player_id)
		if result == nil {
			log.Error("Player[%v] remove friend[%v] failed", this.Id, player_id)
			return int32(msg_client_message.E_ERR_FRIEND_REMOVE_FRIEND_FAILED)
		}
	} else {
		p.remove_friend_data(this.Id)
	}

	this.remove_friend_data(player_id)

	response := &msg_client_message.S2CFriendRemoveResponse{}
	//response.PlayerId = player_id
	this.Send(1, response)

	log.Debug("Player[%v] removed friend[%v]", this.Id, player_id)

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
	this.db.Friends.FillAllListMsg(response)
	this.db.FriendReqs.FillAllListMsg(response)

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
	is_full, next_id := this.db.FriendChatUnreadIds.CheckUnreadNumFull(friend_id)
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

	log.Debug("Player[%v] add friend[%v] chat message[id:%v, long_id:%v, content:%v]", this.Id, friend_id, next_id, new_long_id, message)

	return 1
}

func (this *Player) friend_chat(friend_id int32, message []byte) int32 {
	if !this.db.Friends.HasIndex(friend_id) {
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

	/*response := &msg_client_message.S2CFriendChatResult{}
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
	if !this.db.Friends.HasIndex(friend_id) {
		log.Error("Player[%v] no friend[%v], pull unread message failed", this.Id, friend_id)
		return int32(msg_client_message.E_ERR_FRIEND_NO_THE_FRIEND)
	}

	/*c := 0
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
	if !this.db.Friends.HasIndex(friend_id) {
		log.Error("Player[%v] no friend[%v], confirm unread message failed", this.Id, friend_id)
		return int32(msg_client_message.E_ERR_FRIEND_NO_THE_FRIEND)
	}

	res, remove_ids := this.db.FriendChatUnreadIds.ConfirmUnreadIds(friend_id, message_num)
	if res < 0 {
		return res
	}

	this.db.FriendChatUnreadMessages.RemoveMessages(friend_id, remove_ids)

	/*response := &msg_client_message.S2CFriendConfirmUnreadMessageResult{}
	response.FriendId = proto.Int32(friend_id)
	response.MessageNum = proto.Int32(message_num)
	this.Send(response)*/

	log.Debug("Player[%v] confirm friend[%v] unread message num[%v]", this.Id, friend_id, message_num)

	return 1
}

// ------------------------------------------------------
