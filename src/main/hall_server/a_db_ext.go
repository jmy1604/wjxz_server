package main

import (
	"fmt"
	"libs/log"
	"public_message/gen_go/client_message"
	"public_message/gen_go/server_message"
	"time"
	"youma/table_config"

	"3p/code.google.com.protobuf/proto"
)

func (this *DBC) on_preload() (err error) {
	var p *Player
	for _, db := range this.Players.m_rows {
		if nil == db {
			log.Error("DBC on_preload Players have nil db !")
			continue
		}

		p = new_player_with_db(db.m_PlayerId, db)
		if nil == p {
			continue
		}

		player_mgr.Add2IdMap(p)
		player_mgr.Add2AccMap(p)
	}

	return
}

func (this *dbPlayerInfoColumn) FillBaseInfo(bi *msg_client_message.S2CRetBaseInfo) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.FillBaseInfo")
	defer this.m_row.m_lock.UnSafeRUnlock()
	tmp_data := this.m_data
	bi.Coins = proto.Int32(tmp_data.Coin)
	bi.Diamonds = proto.Int32(tmp_data.Diamond)
	return
}

func (this *dbPlayerInfoColumn) SubCoin(v int32) int32 {
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SubCoin")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.Coin = this.m_data.Coin - v
	if this.m_data.Coin < 0 {
		this.m_data.Coin = 0
	}

	this.m_changed = true
	return this.m_data.Coin
}

func (this *dbPlayerInfoColumn) SubDiamond(v int32) int32 {
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SubDiamond")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.Diamond = this.m_data.Diamond - v
	if this.m_data.Diamond < 0 {
		this.m_data.Diamond = 0
	}
	this.m_changed = true
	return this.m_data.Diamond
}

func (this *dbPlayerMailColumn) GetAviMailId() (ret_idx int32) {
	var min_only_txt_id, min_other_id int32
	min_only_txt_sec := int32(time.Now().Unix())
	min_other_sec := min_only_txt_sec

	var cur_mail *dbPlayerMailData
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.GetAviMailId")
	defer this.m_row.m_lock.UnSafeRUnlock()
	for idx := int32(1); idx <= global_config_mgr.GetGlobalConfig().MaxMailCount; idx++ {
		cur_mail = this.m_data[idx]
		if nil == cur_mail {
			return idx
		}

		if len(cur_mail.ObjIds) > 1 || PLAYER_MAIL_TYPE_REQ_HELP == cur_mail.MailType {
			if cur_mail.SendUnix < min_other_sec {
				min_other_id = idx
				min_other_sec = cur_mail.SendUnix
			}
		} else {
			if cur_mail.SendUnix < min_only_txt_sec {
				min_only_txt_id = idx
				min_only_txt_sec = cur_mail.SendUnix
			}
		}

	}

	if min_only_txt_id > 0 {
		delete(this.m_data, min_only_txt_id)
		ret_idx = min_only_txt_id
	} else if min_other_id > 0 {
		delete(this.m_data, min_other_id)
		ret_idx = min_other_id
	}

	return -1
}

func (this *dbPlayerMailColumn) FillMsgList() *msg_client_message.S2CMailList {
	rm_mailids := make(map[int32]int32)
	cur_unix := int32(time.Now().Unix())
	var tmp_mail *msg_client_message.MailInfo
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.FillMsgList")
	defer this.m_row.m_lock.UnSafeUnlock()

	tmp_len := int32(len(this.m_data))
	if tmp_len < 1 {
		return nil
	}

	ret_msg := &msg_client_message.S2CMailList{}
	ret_msg.MailList = make([]*msg_client_message.MailInfo, 0, tmp_len)
	for _, val := range this.m_data {
		if nil == val {
			continue
		}

		tmp_mail = &msg_client_message.MailInfo{}
		tmp_mail.MailId = proto.Int32(val.MailId)
		if val.OverUnix > 0 && cur_unix >= val.OverUnix {
			log.Info("dbPlayerMailColumn mail FillMsgList [%d] over [%d] [%d]", val.MailId, cur_unix, val.OverUnix)
			rm_mailids[val.MailId] = 1
			tmp_mail.OpType = proto.Int32(PLAYER_MAIL_OP_TYPE_REMOVE)
		} else {
			tmp_mail.OpType = proto.Int32(PLAYER_MAIL_OP_TYPE_SYNC)
			tmp_mail.MailType = proto.Int32(int32(val.MailType))
			tmp_mail.SenderId = proto.Int32(val.SenderId)
			tmp_mail.SenderName = proto.String(val.SenderName)
			tmp_mail.Title = proto.String(val.MailTitle)
			tmp_mail.Content = proto.String(val.Content)
			tmp_mail.SendUnix = proto.Int32(val.SendUnix)
			tmp_mail.State = proto.Int32(int32(val.State))
			tmp_mail.ObjIds = val.ObjIds
			tmp_mail.ObjNums = val.ObjNums

			if val.OverUnix > 0 {
				tmp_mail.LeftSec = proto.Int32(val.OverUnix - cur_unix)
			}
		}

		ret_msg.MailList = append(ret_msg.MailList, tmp_mail)
	}

	if len(rm_mailids) > 0 {
		for mail_id, _ := range rm_mailids {
			delete(this.m_data, mail_id)
		}
		this.m_changed = true
	}

	if len(ret_msg.MailList) > 0 {
		return ret_msg
	}

	return nil
}

func (this *dbPlayerDialyTaskColumn) ChkResetDialyTask() {
	//rm_ids := make(map[int32]bool)
	this.m_row.m_lock.UnSafeLock("dbPlayerDialyTaskColumn.ChkResetDialyTask")
	defer this.m_row.m_lock.UnSafeUnlock()

	/*for task_id, val := range this.m_data {
		if nil == val {
			rm_ids[task_id] = true
			continue
		}

		//tmp_cfg := achieve_task_mgr.DialyTaskMap[val.TaskId]
		tmp_cfg := task_table_mgr.GetTaskMap()[val.TaskId]
		if nil == tmp_cfg {
			rm_ids[task_id] = true
			continue
		}

		if val.Value < tmp_cfg.EventParam {
			rm_ids[task_id] = true
			continue
		}

		if val.RewardUnix > 0 {
			rm_ids[task_id] = true
			continue
		}

	}

	for task_id, _ := range rm_ids {
		delete(this.m_data, task_id)
	}*/

	daily_tasks := task_table_mgr.GetDailyTasks()
	if daily_tasks == nil {
		return
	}

	for id, task := range daily_tasks {
		d := this.m_data[id]
		if d == nil {
			data := &dbPlayerDialyTaskData{}
			data.TaskId = task.Id
			data.Value = 0
			data.RewardUnix = 0
			this.m_data[id] = data
		} else {
			if d.RewardUnix > 0 || d.Value < task.CompleteNum {
				d.Value = 0
				d.RewardUnix = 0
			}
		}
	}

	this.m_changed = true

	return
}

func (this *dbPlayerDialyTaskColumn) FillDialyTaskMsg(p *Player) *msg_client_message.S2CSyncDialyTask {
	var tmp_item *msg_client_message.DialyTaskData
	this.m_row.m_lock.UnSafeRLock("dbPlayerDialyTaskColumn.ChkResetDialyTask")
	defer this.m_row.m_lock.UnSafeRUnlock()
	ret_msg := &msg_client_message.S2CSyncDialyTask{}
	ret_msg.TaskList = make([]*msg_client_message.DialyTaskData, 0, len(this.m_data))
	for _, val := range this.m_data {
		if nil == val {
			continue
		}

		tmp_item = &msg_client_message.DialyTaskData{}
		tmp_item.TaskId = proto.Int32(val.TaskId)
		tmp_item.TaskValue = proto.Int32(val.Value)
		if val.RewardUnix > 0 {
			tmp_item.TaskState = proto.Int32(TASK_STATE_REWARD)
		} else if p.IsTaskCompleteById(val.TaskId) {
			tmp_item.TaskState = proto.Int32(TASK_STATE_COMPLETE)
		} else {
			tmp_item.TaskState = proto.Int32(TASK_STATE_DOING)
		}
		//tmp_item.RewardUnix = proto.Int32(val.RewardUnix)
		ret_msg.TaskList = append(ret_msg.TaskList, tmp_item)
	}

	return ret_msg
}

func (this *dbPlayerAchieveColumn) FillAchieveMsg(p *Player) *msg_client_message.S2CSyncAchieveData {
	ret_msg := &msg_client_message.S2CSyncAchieveData{}
	var tmp_item *msg_client_message.AchieveData
	this.m_row.m_lock.UnSafeRLock("dbPlayerAchieveColumn.FillAchieveMsg")
	defer this.m_row.m_lock.RUnlock()

	tmp_len := int32(len(this.m_data))
	if tmp_len < 1 {
		return nil
	}

	ret_msg.AchieveList = make([]*msg_client_message.AchieveData, 0, tmp_len)
	for _, val := range this.m_data {
		if nil == val {
			continue
		}

		tmp_item = &msg_client_message.AchieveData{}
		tmp_item.AchieveId = proto.Int32(val.AchieveId)
		tmp_item.AchieveValue = proto.Int32(val.Value)
		// 已领奖
		if val.RewardUnix > 0 {
			tmp_item.AchieveState = proto.Int32(TASK_STATE_REWARD)
		} else if p.IsTaskCompleteById(val.AchieveId) {
			tmp_item.AchieveState = proto.Int32(TASK_STATE_COMPLETE)
		} else {
			tmp_item.AchieveState = proto.Int32(TASK_STATE_DOING)
		}
		//tmp_item.RewardUnix = proto.Int32(val.RewardUnix)
		ret_msg.AchieveList = append(ret_msg.AchieveList, tmp_item)
	}

	return ret_msg
}

/*
func (this *dbPlayerSevenActivityColumn) FillSevenMsg(p *Player) *msg_client_message.S2CSyncSevenActivity {
	create_unix := p.db.Info.GetCreateUnix()
	ret_msg := &msg_client_message.S2CSyncSevenActivity{}
	var tmp_act *msg_client_message.ActivityData
	this.m_row.m_lock.UnSafeRLock("dbPlayerSevenActivityColumn.FillSevenMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()
	ret_msg.ActivityList = make([]*msg_client_message.ActivityData, 0, len(this.m_data))
	for _, v := range this.m_data {
		if nil == v {
			continue
		}

		tmp_act = &msg_client_message.ActivityData{}
		tmp_act.ActivityId = proto.Int32(v.ActivityId)
		tmp_act.ActivityValue = proto.Int32(v.Value)
		tmp_act.RewardUnix = proto.Int32(v.RewardUnix)
		tmp_act.LeftDays = proto.Int32(cfg_player_act_mgr.GetSevenDayLeftDays(timer.GetDayFrom1970WithCfgAndSec(0, create_unix), timer.GetDayFrom1970WithCfg(int32(time.Now().Unix())), v.ActivityId))
		ret_msg.ActivityList = append(ret_msg.ActivityList, tmp_act)
	}
	return ret_msg
}
*/
func (this *dbPlayerSignInfoColumn) FillSyncMsg(msg *msg_client_message.S2CSyncSignInfo) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerSignInfoColumn.FillSyncMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()

	msg.CurSignSum = proto.Int32(this.m_data.CurSignSum)
	msg.CurSignDays = this.m_data.CurSignDays
	msg.CurGetSignSumRewards = this.m_data.RewardSignSum

	return
}

func (this *dbPlayerGuidesColumn) ForceAdd(guide_id int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerGuidesColumn.ForceAdd")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[guide_id]
	if has {
		return
	}
	d := &dbPlayerGuidesData{}
	d.GuideId = guide_id
	d.SetUnix = int32(time.Now().Unix())
	this.m_data[guide_id] = d
	this.m_changed = true
	return
}

func (this *dbPlayerGuidesColumn) FillSyncMsg(msg *msg_client_message.S2CSyncGuideData) {
	if nil == msg {
		log.Error("dbPlayerGuidesColumn FillSyncMsg msg nil !")
		return
	}

	this.m_row.m_lock.UnSafeRLock("dbPlayerGuidesColumn.FillSyncMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()

	msg.GuideIds = make([]int32, 0, len(this.m_data))
	for _, val := range this.m_data {
		msg.GuideIds = append(msg.GuideIds, val.GuideId)
	}

	return
}

func (this *dbPlayerFriendColumn) FillAllListMsg(msg *msg_client_message.S2CRetFriendListResult) {
	var tmp_info *msg_client_message.FriendInfo
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.FillAllListMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()
	msg.FriendList = make([]*msg_client_message.FriendInfo, 0, len(this.m_data))
	for _, val := range this.m_data {
		if nil == val {
			continue
		}

		tmp_info = &msg_client_message.FriendInfo{}
		tmp_info.PlayerId = proto.Int32(val.FriendId)
		tmp_info.Name = proto.String(val.FriendName)
		tmp_info.Level = proto.Int32(val.Level)
		tmp_info.VipLevel = proto.Int32(val.VipLevel)
		tmp_info.LastLogin = proto.Int32(val.LastLogin)
		tmp_info.Head = proto.String(val.Head)
		tmp_info.IsOnline = proto.Bool(true)
		log.Info("附加值到好友列表 %v", tmp_info)
		msg.FriendList = append(msg.FriendList, tmp_info)
	}

	return
}

func (this *dbPlayerFriendColumn) GetAviFriendId() int32 {
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.GetAviFriendId")
	defer this.m_row.m_lock.UnSafeRUnlock()
	for i := int32(1); i <= global_config_mgr.GetGlobalConfig().MaxFriendNum; i++ {
		if nil == this.m_data[i] {
			return i
		}
	}
	return 0
}

func (this dbPlayerFriendColumn) TryAddFriend(new_friend *dbPlayerFriendData) {
	if nil == new_friend {
		log.Error("dbPlayerFriendColumn TryAddFriend ")
		return
	}

	this.m_row.m_lock.UnSafeLock("dbPlayerFriendColumn.TryAddFriend")
	defer this.m_row.m_lock.UnSafeUnlock()

	if nil == this.m_data[new_friend.FriendId] {
		this.m_data[new_friend.FriendId] = new_friend
		this.m_changed = true
	}

	return
}

func (this *dbPlayerFriendReqColumn) FillAllListMsg(msg *msg_client_message.S2CRetFriendListResult) {

	var tmp_info *msg_client_message.FriendReq
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendReqColumn.FillAllListMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()

	msg.Reqs = make([]*msg_client_message.FriendReq, 0, len(this.m_data))
	for _, val := range this.m_data {
		if nil == val {
			continue
		}

		tmp_info = &msg_client_message.FriendReq{}
		tmp_info.PlayerId = proto.Int32(val.PlayerId)
		tmp_info.Name = proto.String(val.PlayerName)
		msg.Reqs = append(msg.Reqs, tmp_info)
	}

	return
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
	d.PlayerName = player_name
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

/*func (this *dbPlayerFocusPlayerColumn) FillAllListMsg(msg *msg_client_message.S2CRetFriendList) {

	var tmp_info *msg_client_message.FriendInfo
	this.m_row.m_lock.UnSafeRLock("dbPlayerFocusPlayerColumn.FillAllListMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()
	msg.FriendList = make([]*msg_client_message.FriendInfo, 0, len(this.m_data))
	for _, val := range this.m_data {
		if nil == val {
			continue
		}

		tmp_info = &msg_client_message.FriendInfo{}
		tmp_info.PlayerId = proto.Int32(val.FriendId)
		tmp_info.Name = proto.String(val.FriendName)
		msg.FriendList = append(msg.FriendList, tmp_info)
	}

	return
}

func (this *dbPlayerBeFocusPlayerColumn) FillAllListMsg(msg *msg_client_message.S2CRetFriendList) {

	var tmp_info *msg_client_message.FriendInfo
	this.m_row.m_lock.UnSafeRLock("dbPlayerBeFocusPlayerColumn.FillAllListMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()
	msg.FriendList = make([]*msg_client_message.FriendInfo, 0, len(this.m_data))
	for _, val := range this.m_data {
		if nil == val {
			continue
		}

		tmp_info = &msg_client_message.FriendInfo{}
		tmp_info.PlayerId = proto.Int32(val.FriendId)
		tmp_info.Name = proto.String(val.FriendName)
		msg.FriendList = append(msg.FriendList, tmp_info)
	}

	return
}*/

func (this *dbPlayerFriendColumn) GetAllIds() (ret_ids []int32) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.GetAllIds")
	defer this.m_row.m_lock.UnSafeRUnlock()
	tmp_len := len(this.m_data)
	if tmp_len <= 0 {
		return nil
	}

	ret_ids = make([]int32, 0, len(this.m_data))
	for _, v := range this.m_data {
		ret_ids = append(ret_ids, v.FriendId)
	}
	return
}

func (this *dbPlayerFocusPlayerColumn) GetAllIds() (ret_ids []int32) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerFocusPlayerColumn.GetAllIds")
	defer this.m_row.m_lock.UnSafeRUnlock()
	tmp_len := len(this.m_data)
	if tmp_len <= 0 {
		return nil
	}

	ret_ids = make([]int32, 0, len(this.m_data))
	for _, v := range this.m_data {
		ret_ids = append(ret_ids, v.FriendId)
	}

	return
}

func (this *dbPlayerBeFocusPlayerColumn) GetNum() int32 {
	this.m_row.m_lock.UnSafeRLock("dbPlayerBeFocusPlayerColumn.GetNum")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}

func (this *dbPlayerItemColumn) ChkAddItemByNum(cfgid, num int32) int32 {
	this.m_row.m_lock.UnSafeLock("dbPlayerItemColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()

	item := item_table_mgr.Map[cfgid]
	if item == nil {
		log.Error("添加物品时找不到物品配置ID[%v]", cfgid)
		return 0
	}
	d, has := this.m_data[cfgid]
	if has {
		if item.ValidTime == 0 {
			d.ItemNum += num
			if d.ItemNum > item.MaxNumber {
				d.ItemNum = item.MaxNumber
			}
		} else {
			d.ItemNum = 1
			d.StartTimeUnix = int32(time.Now().Unix())
			d.RemainSeconds = item.ValidTime * 3600
		}
	} else {
		d = &dbPlayerItemData{}
		d.ItemCfgId = cfgid
		if item.ValidTime == 0 {
			if num > item.MaxNumber {
				num = item.MaxNumber
			}
			d.ItemNum = num
		} else {
			d.ItemNum = 1
			d.StartTimeUnix = int32(time.Now().Unix())
			d.RemainSeconds = item.ValidTime * 3600
		}
		this.m_data[cfgid] = d
	}
	this.m_changed = true

	return d.ItemNum
}

func (this *dbPlayerItemColumn) ChkRemoveItem(item_id, num int32) (bool, int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerItemColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	item := item_table_mgr.Map[item_id]
	if item == nil {
		log.Error("删除物品[%v]时找不到ID", item_id)
		return false, 0
	}
	d, has := this.m_data[item_id]
	if !has {
		return false, 0
	}
	var left int32
	if d.ItemNum > num {
		d.ItemNum -= num
		left = d.ItemNum
	} else {
		delete(this.m_data, item_id)
		left = 0
	}
	this.m_changed = true
	return true, left
}

func (this *dbPlayerStageColumn) ChkSetTopScore(id int32, v int32) int32 {
	this.m_row.m_lock.UnSafeLock("dbPlayerStageColumn.ChkSetTopScore")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("not exist %v %v", this.m_row.GetPlayerId(), id)
		return d.TopScore
	}
	if d.TopScore < v {
		d.TopScore = v
		this.m_changed = true
	}

	return d.TopScore
}

func (this *dbPlayerStageColumn) GetTotalTopStar() int32 {
	this.m_row.m_lock.UnSafeRLock("dbPlayerStageColumn.GetTotalTopStar")
	defer this.m_row.m_lock.UnSafeRUnlock()

	total_top := int32(0)
	for _, d := range this.m_data {
		if nil == d {
			continue
		}

		total_top += d.Stars
	}

	return total_top
}

func (this *dbPlayerInfoColumn) ChkSetCurMaxStage(v int32) bool {
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.ChkSetCurMaxStage")
	defer this.m_row.m_lock.UnSafeUnlock()
	if this.m_data.CurMaxStage < v {
		this.m_data.CurMaxStage = v
		this.m_changed = true
		return true
	}
	return false
}

func (this *dbPlayerStageColumn) ChkGetTopScore(id int32) int32 {
	this.m_row.m_lock.UnSafeRLock("dbPlayerStageColumn.ChkGetTopScore")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d == nil {
		return 0
	}

	return d.TopScore
}

func (this *dbPlayerItemColumn) FillAllMsg(msg *msg_client_message.S2CRetItemInfos) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerItemColumn.FillAllMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()

	var tmp_item *msg_client_message.ItemInfo
	msg.Items = make([]*msg_client_message.ItemInfo, 0, len(this.m_data))
	for _, v := range this.m_data {
		if nil == v {
			continue
		}

		tmp_item = &msg_client_message.ItemInfo{}
		tmp_item.ItemCfgId = proto.Int32(v.ItemCfgId)
		tmp_item.ItemNum = proto.Int32(v.ItemNum)
		tmp_item.RemainSeconds = proto.Int32(get_time_item_remain_seconds(v))
		msg.Items = append(msg.Items, tmp_item)
	}

	return
}

func (this *dbPlayerBuildingColumn) FillAllMsg(msg *msg_client_message.S2CRetBuildingInfos) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerBuildingColumn.FillAllMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()

	var tmp_build *msg_client_message.BuildingInfo
	msg.Builds = make([]*msg_client_message.BuildingInfo, 0, len(this.m_data))
	for _, v := range this.m_data {
		if nil == v {
			continue
		}

		tmp_build = &msg_client_message.BuildingInfo{}
		tmp_build.Id = proto.Int32(v.Id)
		tmp_build.CfgId = proto.Int32(v.CfgId)
		tmp_build.X = proto.Int32(v.X)
		tmp_build.Y = proto.Int32(v.Y)
		tmp_build.Dir = proto.Int32(v.Dir)
		msg.Builds = append(msg.Builds, tmp_build)
		if nil != cfg_mapchest_mgr.Map[v.CfgId] {
			tmp_time := time.Unix(int64(v.CreateUnix), 0)
			log.Info("宝箱[%d:%d]的开始时间 %s", v.Id, v.CfgId, tmp_time.Format("2006-01-02 15:04:05.999999999 -0700 MST"))
			tmp_time = time.Unix(int64(v.OverUnix), 0)
			log.Info("宝箱[%d:%d]的结束时间 %s", v.Id, v.CfgId, tmp_time.Format("2006-01-02 15:04:05.999999999 -0700 MST"))
		}
	}

	return
}

func (this *dbPlayerBuildingDepotColumn) FillAllMsg(msg *msg_client_message.S2CRetDepotBuildingInfos) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerBuildingDepotColumn.FillAllMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()

	var tmp_build *msg_client_message.DepotBuildingInfo
	msg.DepotBuilds = make([]*msg_client_message.DepotBuildingInfo, 0, len(this.m_data))
	for _, v := range this.m_data {
		if nil == v {
			continue
		}
		tmp_build = &msg_client_message.DepotBuildingInfo{}
		tmp_build.CfgId = proto.Int32(v.CfgId)
		tmp_build.Num = proto.Int32(v.Num)
		msg.DepotBuilds = append(msg.DepotBuilds, tmp_build)
	}
	return
}

func (this *dbPlayerCatColumn) FillAllMsg(msg *msg_client_message.S2CRetCatInfos) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerCatColumn.FillAllMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()

	var tmp_cat *msg_client_message.CatInfo
	msg.Cats = make([]*msg_client_message.CatInfo, 0, len(this.m_data))
	for _, v := range this.m_data {
		if nil == v {
			continue
		}

		tmp_cat = &msg_client_message.CatInfo{}
		tmp_cat.Id = proto.Int32(v.Id)
		tmp_cat.CatCfgId = proto.Int32(v.CfgId)
		tmp_cat.Level = proto.Int32(v.Level)
		tmp_cat.Star = proto.Int32(v.Star)
		tmp_cat.SkillLevel = proto.Int32(v.SkillLevel)
		lock := false
		if v.Locked > 0 {
			lock = true
		}
		tmp_cat.Locked = proto.Bool(lock)
		tmp_cat.Exp = proto.Int32(v.Exp)
		tmp_cat.CoinAbility = proto.Int32(v.CoinAbility)
		tmp_cat.ExploreAbility = proto.Int32(v.ExploreAbility)
		tmp_cat.MatchAbility = proto.Int32(v.MatchAbility)
		tmp_cat.Nick = proto.String(v.Nick)
		if v.CathouseId > 0 {
			tmp_cat.State = proto.Int32(CAT_STATE_IN_CATHOUSE)
		}
		msg.Cats = append(msg.Cats, tmp_cat)
	}

	return
}

func (this *dbPlayerAreaColumn) FillAllMsg(msg *msg_client_message.S2CRetAreasInfos) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerAreaColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()

	var tmp_area *msg_client_message.AreaInfo
	msg.Areas = make([]*msg_client_message.AreaInfo, 0, len(this.m_data))
	for _, v := range this.m_data {
		if nil == v {
			continue
		}

		tmp_area = &msg_client_message.AreaInfo{}
		tmp_area.CfgId = proto.Int32(v.CfgId)
		msg.Areas = append(msg.Areas, tmp_area)
	}
	return
}

func (this *dbPlayerAreaColumn) GetAllAreaInfo() (all_area []*msg_client_message.AreaInfo) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerAreaColumn.GetAllAreaInfo")
	defer this.m_row.m_lock.UnSafeRUnlock()

	var tmp_area *msg_client_message.AreaInfo
	all_area = make([]*msg_client_message.AreaInfo, 0, len(this.m_data))
	for _, v := range this.m_data {
		if nil == v {
			continue
		}

		tmp_area = &msg_client_message.AreaInfo{}
		tmp_area.CfgId = proto.Int32(v.CfgId)
		all_area = append(all_area, tmp_area)
	}
	return
}

func (this *dbPlayerStageColumn) FillAllMsg(msg *msg_client_message.S2CRetStageInfos) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerStageColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()

	var tmp_stage *msg_client_message.StageInfo
	msg.Stages = make([]*msg_client_message.StageInfo, 0, len(this.m_data))
	for stageid, v := range this.m_data {
		if nil == v {
			continue
		}
		tmp_stage = &msg_client_message.StageInfo{}
		tmp_stage.StageId = proto.Int32(stageid)
		tmp_stage.TopScore = proto.Int32(v.TopScore)
		tmp_stage.Star = proto.Int32(v.Stars)
		msg.Stages = append(msg.Stages, tmp_stage)
	}
}

func (this *dbPlayerAreaColumn) GetAllIdxs() (list []int32) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerAreaColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, 0, len(this.m_data))

	for _, v := range this.m_data {
		list = append(list, v.CfgId)
	}
	return
}

func (this *dbPlayerBuildingColumn) GetAllBuildingPos() (pos_map map[int32]int32, cur_area_block_count map[int32]int32) { // , cur_area_use_count map[int32]int32
	this.m_row.m_lock.UnSafeRLock("dbPlayerBuildingColumn.GetAllBuildingPos")
	defer this.m_row.m_lock.UnSafeRUnlock()

	pos_map = make(map[int32]int32, len(this.m_data))
	//cur_area_use_count = make(map[int32]int32)
	cur_area_block_count = make(map[int32]int32)
	var arena_xy, arena_id int32
	var building_cfg *table_config.XmlBuildingItem
	for _, d := range this.m_data {
		if nil == d {
			continue
		}

		building_cfg = cfg_building_mgr.Map[d.CfgId]
		if nil == building_cfg {
			continue
		}

		var width, height int32
		if table_config.BUILDING_DIR_BIG_X_DIR == d.Dir {
			width, height = building_cfg.MapSizes[0], building_cfg.MapSizes[1]
		} else {
			width, height = building_cfg.MapSizes[1], building_cfg.MapSizes[0]
		}

		if nil != cfg_block_mgr.Map[d.CfgId] {
			arena_xy = (d.X)<<16 | (d.Y)&0x0000FFFF
			arena_id = cfg_build_area_mgr.AreaXY2AreaId[arena_xy]
			if arena_id > 0 {
				cur_area_block_count[arena_id] = cur_area_block_count[arena_id] + 1
			}
		}

		for tmp_x := int32(0); tmp_x < width; tmp_x++ {
			for tmp_y := int32(0); tmp_y < height; tmp_y++ {
				arena_xy = (d.X+tmp_x)<<16 | (d.Y+tmp_y)&0x0000FFFF
				pos_map[arena_xy] = d.Id
				//arena_id = cfg_build_area_mgr.AreaXY2AreaId[arena_xy]
				//if arena_id > 0 {
				//cur_area_use_count[arena_id] = cur_area_use_count[arena_id] + 1
				//}
			}
		}
	}

	return
}

func (this *dbPlayerExpeditionColumn) CheckUpdateExpedition(p_lvl int32) (cur_ids map[int32]bool, cur_count int32) {
	var task_cfg *table_config.XmlExpeditionItem
	del_map := make(map[int32]bool)
	this.m_row.m_lock.UnSafeLock("dbPlayerExpeditionsColumn.CheckUpdateExpedition")
	defer this.m_row.m_lock.UnSafeUnlock()

	//log.Info("dbPlayerExpeditionColumn CheckUpdateExpedition, m_data[%v]", this.m_data)

	cur_unix := int32(time.Now().Unix())
	//cur_count := int32(0)
	cur_ids = make(map[int32]bool)
	var pass_sec int32
	for _, val := range this.m_data {
		if nil == val {
			continue
		}

		fmt.Println("CheckUpdateExpedition val [%v]", *val)
		task_cfg = cfg_expedition_mgr.Map[val.TaskId]
		if nil == task_cfg {
			continue
		}
		if PLAYER_EXPEDITION_STATE_DOING == val.CurState && cur_unix > val.EndUnix {
			fmt.Println("====================1", *val)
			if val.Result > 0 {
				val.CurState = PLAYER_EXPEDITION_STATE_SUCCEED
			} else {
				val.CurState = PLAYER_EXPEDITION_STATE_FAILED
			}
		}

		if PLAYER_EXPEDITION_TYPE_TIMELIMIT == task_cfg.TaskType && PLAYER_EXPEDITION_STATE_INIT == val.CurState {
			fmt.Println("====================2", *val)
			pass_sec = cur_unix - val.TaskLeftSecLastUpUnix
			if pass_sec >= val.TaskLeftSec {
				log.Info("限时任务超过期限%d pass_sec[%d-%d=%d]", val.TaskLeftSec, cur_unix, val.TaskLeftSecLastUpUnix, pass_sec)
				del_map[val.Id] = true
				continue
			} else {
				val.TaskLeftSec -= pass_sec
				val.TaskLeftSecLastUpUnix = cur_unix
			}

		}

		fmt.Println("====================3", *val)

		cur_count++
		cur_ids[val.TaskId] = true
		log.Info("CheckUpdateExpedition2 val [%v]", *val)
	}

	for id, _ := range del_map {
		delete(this.m_data, id)
		this.m_changed = true
	}

	log.Info("需要随机%d-%d个任务 删除了%d个任务", global_config_mgr.GetGlobalConfig().ExpeditionTaskCount, cur_count, len(del_map))

	//need_count = global_config_mgr.GetGlobalConfig().ExpeditionTaskCount - cur_count

	/*
		if cur_count < global_config_mgr.GetGlobalConfig().ExpeditionTaskCount {
			new_tasks := cfg_expedition_mgr.RandNWithExistIds(cur_ids, p_lvl, global_config_mgr.GetGlobalConfig().ExpeditionTaskCount-cur_count)
			var tmp_task *dbPlayerExpeditionData

			var rand_val, total_weight int32
			for _, task := range new_tasks {
				if nil == task {
					continue
				}

				tmp_task = &dbPlayerExpeditionData{TaskId: task.Id, StartUnix: int32(time.Now().Unix())}
				if PLAYER_EXPEDITION_TYPE_TIMELIMIT == task.TaskType {
					tmp_task.TaskLeftSecLastUpUnix = cur_unix
					tmp_task.TaskLeftSec = task.LimitTimeSec
					log.Info("设置限时任务的刷新时间 %d", tmp_task.TaskLeftSec)
				}

				// 随机任务条件
				log.Info("随机任务[%d]的条件", task.Id)
				tmp_task.Conditions = make([]dbExpeditionConData, task.NeedConditionNum)
				total_weight = task.TotalConWeight
				cur_map := make(map[int]bool)
				for cur_num := int32(0); cur_num < task.NeedConditionNum; cur_num++ {
					if total_weight <= 0 {
						log.Info("第%d次随机任务条件totalweight[%d]<0退出", cur_num+1, total_weight)
						break
					}

					rand_val = rand.Int31n(total_weight)
					log.Info("第%d次随机任务条件，totol_weight[%d] rand_val[%d] 当前随机好的对象%v", cur_num+1, total_weight, rand_val, cur_map)
					for idx, tmp_con := range task.Conditions {
						if cur_map[idx] {
							continue
						}

						log.Info("	===随机任务条件对比weight[%d] rand_val[%d]", tmp_con.Con_Weight, rand_val)
						if rand_val < tmp_con.Con_Weight {
							total_weight -= tmp_con.Con_Weight
							tmp_task.Conditions[cur_num].ConType = tmp_con.Con_Type
							if PLAYER_EXPEDITION_CON_CAT_COLOR == tmp_con.Con_Type {
								tmp_task.Conditions[cur_num].ConVals = make([]int32, 0, tmp_con.Con_Val)
								sub_total_weight := tmp_con.Ext_val
								sub_cur_map := make(map[int]bool)
								for sub_cur_num := int32(0); sub_cur_num < tmp_con.Con_Val; sub_cur_num++ {
									if sub_total_weight <= 0 {
										break
									}

									sub_rand_val := rand.Int31n(sub_total_weight)
									for sub_idx, color_weight := range tmp_con.Ext_vals {
										if sub_cur_map[sub_idx] {
											continue
										}

										if sub_rand_val < color_weight {
											tmp_task.Conditions[cur_num].ConVals = append(tmp_task.Conditions[cur_num].ConVals, int32(1<<(uint32(sub_idx))))
											sub_cur_map[sub_idx] = true
											sub_total_weight -= color_weight
											break
										} else {
											sub_rand_val -= color_weight
										}
									}
								}
							} else {
								tmp_task.Conditions[cur_num].ConVals = make([]int32, 1)
								tmp_task.Conditions[cur_num].ConVals[0] = tmp_con.Con_Val
							}
							cur_map[idx] = true
							break
						} else {
							rand_val -= tmp_con.Con_Weight
						}
					}

				}

				this.m_data[task.Id] = tmp_task

				log.Info("赋值任务[task.id]给m_data", tmp_task)
			}
		}

	*/

	return
}

func (this *dbPlayerExpeditionColumn) FillAllClientMsg(msg *msg_client_message.S2CRetAllExpedition) {
	if nil == msg {
		log.Error("dbPlayerExpeditionColumn FillAllClientMsg msg nil !")
		return
	}

	this.m_row.m_lock.UnSafeRLock("dbPlayerExpeditionsColumn.FillAllClientMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()

	tmp_len := int32(len(this.m_data))
	if tmp_len <= 0 {
		return
	}

	cur_unix := int32(time.Now().Unix())

	msg.Tasks = make([]*msg_client_message.ExpeditionItem, 0, tmp_len)
	var tmp_item *msg_client_message.ExpeditionItem
	var task_cfg *table_config.XmlExpeditionItem
	for _, val := range this.m_data {
		if nil == val {
			continue
		}

		task_cfg = cfg_expedition_mgr.Map[val.TaskId]
		if nil == task_cfg {
			log.Error("dbPlayerExpeditionColumn FillAllClientMsg no task cfg[%d]", val.TaskId)
			continue
		}

		tmp_item = &msg_client_message.ExpeditionItem{}
		if PLAYER_EXPEDITION_TYPE_TIMELIMIT == task_cfg.TaskType {
			tmp_item.TaskLeftSec = proto.Int32(val.TaskLeftSec)
		}

		tmp_item.Id = proto.Int32(val.Id)
		tmp_item.TaskId = proto.Int32(val.TaskId)
		tmp_item.InCatIds = val.InCatIds
		if val.EndUnix > cur_unix {
			tmp_item.ExpeditionLeftSec = proto.Int32(val.EndUnix - cur_unix)
		}

		tmp_item.ExpeditionPassSec = proto.Int32(cur_unix - val.StartUnix)

		tmp_item.Result = proto.Int32(val.Result)
		tmp_item.CurState = proto.Int32(val.CurState)

		// 条件
		tmp_item.Conditions = make([]*msg_client_message.ExpeditonCondition, 0, len(val.Conditions))
		for _, tmp_con := range val.Conditions {
			msg_con := &msg_client_message.ExpeditonCondition{}
			msg_con.ConditionType = proto.Int32(tmp_con.ConType)
			msg_con.ConVals = tmp_con.ConVals
			tmp_item.Conditions = append(tmp_item.Conditions, msg_con)
		}

		// 事件
		tmp_item.Events = make([]*msg_client_message.ExpeditonEvent, 0, len(val.EventIds))
		for _, tmp_event := range val.EventIds {
			msg_event := &msg_client_message.ExpeditonEvent{}
			msg_event.EventId = proto.Int32(tmp_event.ClientId)
			msg_event.Sec = proto.Int32(tmp_event.Sec)
			msg_event.DropIdNums = tmp_event.DropIdNums
			tmp_item.Events = append(tmp_item.Events, msg_event)
		}

		msg.Tasks = append(msg.Tasks, tmp_item)
	}

	return
}

func (this *dbPlayerExpeditionColumn) IfCatInExpedition(in_catid int32) bool {
	if in_catid <= 0 {
		log.Error("dbPlayerExpeditionColumn IfCatInExpedition")
		return true
	}

	this.m_row.m_lock.UnSafeRLock("dbPlayerExpeditionsColumn.FillAllClientMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()

	for _, val := range this.m_data {
		if nil == val {
			continue
		}

		for _, catid := range val.InCatIds {
			if catid == in_catid {
				return true
			}
		}
	}

	return false
}

func (this *dbPlayerExpeditionColumn) Stop(taskid int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerExpeditionColumn.Stop")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[taskid]
	if d == nil {
		log.Error("dbPlayerExpeditionColumn.Stop not exist %v %v", this.m_row.GetPlayerId(), taskid)
		return
	}

	d.CurState = PLAYER_EXPEDITION_STATE_INIT
	d.TaskLeftSecLastUpUnix = int32(time.Now().Unix())
	d.InCatIds = nil

	this.m_changed = true
	return
}

func (this *dbPlayerBuildingColumn) GetCountByType(b_type int32) (count int32) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerBuildingColumn.GetCountByType")
	defer this.m_row.m_lock.UnSafeRUnlock()

	var building_cfg *table_config.XmlBuildingItem
	for _, val := range this.m_data {
		building_cfg = cfg_building_mgr.Map[val.CfgId]
		if nil == building_cfg {
			continue
		}

		if b_type == building_cfg.Type {
			count++
		}
	}

	return
}

func (this *dbPlayerMailColumn) IfHaveNew() (has bool) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()

	for _, val := range this.m_data {
		if PLAYER_MAIL_STATE_INIT == val.State {
			return true
		}
	}

	return false
}

func (this *dbPlayerChapterUnLockColumn) SetNewUnlockChapter(chapter_id int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerChapterUnLockColumn.SetNewUnlockChapter")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.ChapterId = chapter_id
	this.m_data.PlayerIds = nil
	this.m_data.CurHelpIds = nil
	this.m_data.StartUnix = int32(time.Now().Unix())
	this.m_changed = true
	return
}

func (this *dbPlayerChapterUnLockColumn) Reset() {
	this.m_row.m_lock.UnSafeLock("dbPlayerChapterUnLockColumn.SetNewUnlockChapter")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.ChapterId = 0
	this.m_data.PlayerIds = nil
	this.m_data.CurHelpIds = nil
	this.m_data.StartUnix = 0
	this.m_changed = true
	return
}

func (this *dbPlayerBuildingColumn) ChkBuildingOver() (over_ids map[int32]bool) {
	cur_unix := int32(time.Now().Unix())
	over_ids = make(map[int32]bool)

	//log.Info("dbPlayerBuildingColumn ChkBuildingOver ")

	this.m_row.m_lock.UnSafeLock("dbPlayerBuildingColumn.ChkBuildingOver")
	defer this.m_row.m_lock.UnSafeUnlock()
	for _, d := range this.m_data {
		if nil == d {
			continue
		}

		if d.OverUnix > 0 && d.OverUnix < cur_unix {
			over_ids[d.Id] = true
		}
	}

	if len(over_ids) > 0 {
		for bid, _ := range over_ids {
			delete(this.m_data, bid)
		}
		this.m_changed = true
	}

	return
}

func (this *dbPlayerActivityColumn) IfHaveAct(act_id int32) bool {
	this.m_row.m_lock.UnSafeRLock("dbPlayerActivityColumn.IfHaveAct")
	defer this.m_row.m_lock.UnSafeRUnlock()

	if nil == this.m_data[act_id] {
		return false
	}

	return true
}

func (this *dbPlayerActivityColumn) FillAllClientMsg(vip_left_day int32) (ret_msg *msg_client_message.ActivityInfosUpdate) {

	this.m_row.m_lock.UnSafeLock("dbPlayerActivityColumn.GetAll")
	defer this.m_row.m_lock.UnSafeUnlock()

	tmp_len := int32(len(this.m_data))
	if tmp_len < 1 {
		return nil
	}

	ret_msg = &msg_client_message.ActivityInfosUpdate{}
	ret_msg.Activityinfos = make([]*msg_client_message.ActivityInfo, 0, tmp_len)
	var tmp_info *msg_client_message.ActivityInfo
	var task_cfg *table_config.XmlActivityItem
	//cur_unix_day := timer.GetDayFrom1970WithCfg(0)
	for _, v := range this.m_data {
		log.Info("dbPlayerActivityColumn 处理 活动 [%d] [%v] %v", v.CfgId, this.m_data, &this.m_data)
		task_cfg = cfg_activity_mgr.Map[v.CfgId]
		if nil == task_cfg {
			log.Error("dbPlayerActivityColumn 找不到配置[%d]", v.CfgId)
			continue
		}

		tmp_info = &msg_client_message.ActivityInfo{}
		tmp_info.CfgId = proto.Int32(v.CfgId)
		tmp_info.States = v.States
		tmp_info.Vals = v.Vals

		ret_msg.Activityinfos = append(ret_msg.Activityinfos, tmp_info)
	}

	return
}

func (this *dbPlayerActivityColumn) GetVals0(id int32) int32 {
	this.m_row.m_lock.UnSafeRLock("dbPlayerActivityColumn.GetVals0")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("GetVals0 not exist %v %v", this.m_row.GetPlayerId(), id)
		return 0
	}

	if len(d.Vals) < 1 {
		return 0
	}

	return d.Vals[0]
}

func (this *dbPlayerActivityColumn) GetValsEnd(id int32) int32 {
	this.m_row.m_lock.UnSafeRLock("dbPlayerActivityColumn.GetValsEnd")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("GetValsEnd not exist %v %v", this.m_row.GetPlayerId(), id)
		return 0
	}

	tmp_len := len(d.Vals)
	if tmp_len < 1 {
		return 0
	}

	return d.Vals[tmp_len-1]
}

func (this *dbPlayerActivityColumn) IfValsHave(id, v int32) bool {
	this.m_row.m_lock.UnSafeRLock("dbPlayerActivityColumn.IfValsHave")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("IfStatesHave not exist %v %v", this.m_row.GetPlayerId(), id)
		return false
	}

	for _, val := range d.Vals {
		if val == v {
			return true
		}
	}

	return false
}

func (this *dbPlayerActivityColumn) SetVals0(id int32, v int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerActivityColumn.SetVals0")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("SetVals0 not exist %v %v", this.m_row.GetPlayerId(), id)
		return
	}

	if len(d.Vals) < 1 {
		d.Vals = make([]int32, 1)
	}

	d.Vals[0] = v

	this.m_changed = true
	return
}

func (this *dbPlayerActivityColumn) AddValsVal(id int32, v int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerActivityColumn.AddValsVal")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("AddValsVal not exist %v %v", this.m_row.GetPlayerId(), id)
		return
	}

	tmp_len := int32(len(d.Vals))
	new_vals := make([]int32, tmp_len+1)
	for idx, val := range d.Vals {
		new_vals[idx] = val
	}

	new_vals[tmp_len] = v
	d.Vals = new_vals

	this.m_changed = true
	return
}

func (this *dbPlayerActivityColumn) RemoveValsVal(id int32, v int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerActivityColumn.RemoveValsVal")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("AddValsVal not exist %v %v", this.m_row.GetPlayerId(), id)
		return
	}

	tmp_len := int32(len(d.Vals))
	new_vals := make([]int32, 0, tmp_len)
	for _, val := range d.Vals {
		if val != v {
			new_vals = append(new_vals, val)
		} else {
			this.m_changed = true
		}
	}

	if this.m_changed {
		d.Vals = new_vals
	}

	return
}

func (this *dbPlayerActivityColumn) ClearVals(id int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerActivityColumn.ClearVals")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("ClearVals not exist %v %v", this.m_row.GetPlayerId(), id)
		return
	}

	d.Vals = nil

	this.m_changed = true
	return
}

func (this *dbPlayerActivityColumn) GetStates0(id int32) int32 {
	this.m_row.m_lock.UnSafeRLock("dbPlayerActivityColumn.GetStates0")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("GetStates0 not exist %v %v", this.m_row.GetPlayerId(), id)
		return 0
	}

	if len(d.States) < 1 {
		return 0
	}

	return d.States[0]
}

func (this *dbPlayerActivityColumn) GetStates1(id int32) int32 {
	this.m_row.m_lock.UnSafeRLock("dbPlayerActivityColumn.GetStates1")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("GetStates0 not exist %v %v", this.m_row.GetPlayerId(), id)
		return 0
	}

	if len(d.States) < 2 {
		return 0
	}

	return d.States[1]
}

func (this *dbPlayerActivityColumn) GetStates2(id int32) int32 {
	this.m_row.m_lock.UnSafeRLock("dbPlayerActivityColumn.GetStates2")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("GetStates0 not exist %v %v", this.m_row.GetPlayerId(), id)
		return 0
	}

	if len(d.States) < 3 {
		return 0
	}

	return d.States[2]
}

func (this *dbPlayerActivityColumn) IfStatesHave(id, v int32) bool {
	this.m_row.m_lock.UnSafeRLock("dbPlayerActivityColumn.GetStates0")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("IfStatesHave not exist %v %v", this.m_row.GetPlayerId(), id)
		return false
	}

	for _, val := range d.States {
		if val == v {
			return true
		}
	}

	return false
}

func (this *dbPlayerActivityColumn) SetStates0(id int32, v int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerActivityColumn.SetStates0")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("not exist %v %v", this.m_row.GetPlayerId(), id)
		return
	}

	if len(d.States) < 1 {
		d.States = make([]int32, 1)
	}

	d.States[0] = v

	this.m_changed = true
	return
}

func (this *dbPlayerActivityColumn) IncbyStates0(id int32, v int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerActivityColumn.IncbyStates0")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("not exist %v %v", this.m_row.GetPlayerId(), id)
		return
	}

	if len(d.States) < 1 {
		d.States = make([]int32, 1)
	}

	d.States[0] += v

	this.m_changed = true
	return
}

func (this *dbPlayerActivityColumn) SetStates1(id int32, v int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerActivityColumn.SetStates1")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("not exist %v %v", this.m_row.GetPlayerId(), id)
		return
	}

	cur_len := int32(len(d.States))
	if cur_len < 2 {
		new_states := make([]int32, 2)
		for idx := int32(0); idx < cur_len; idx++ {
			new_states[idx] = d.States[idx]
		}

		d.States = new_states
	}

	d.States[1] = v

	this.m_changed = true
	return
}

func (this *dbPlayerActivityColumn) SetStates2(id int32, v int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerActivityColumn.SetStates2")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("not exist %v %v", this.m_row.GetPlayerId(), id)
		return
	}

	cur_len := int32(len(d.States))
	if cur_len < 3 {
		new_states := make([]int32, 3)
		for idx := int32(0); idx < cur_len; idx++ {
			new_states[idx] = d.States[idx]
		}

		d.States = new_states
	}

	d.States[2] = v

	this.m_changed = true
	return
}

func (this *dbPlayerActivityColumn) AddStateVal(id, v int32, bunique bool) {
	this.m_row.m_lock.UnSafeLock("dbPlayerActivityColumn.AddState")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("AddState not exist %v %v", this.m_row.GetPlayerId(), id)
		return
	}

	tmp_len := len(d.States)
	new_states := make([]int32, tmp_len+1)
	for idx := 0; idx < tmp_len; idx++ {
		if bunique && d.States[idx] == v {
			return
		}

		new_states[idx] = d.States[idx]
	}

	new_states[tmp_len] = v

	d.States = new_states

	this.m_changed = true
	return
}

func (this *dbPlayerInfoColumn) ChkGetNextExpeditionId() (r int32) {
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.ChkGetNextExpeditionId")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.NextExpeditionId += 1
	if this.m_data.NextExpeditionId <= 0 {
		this.m_data.NextExpeditionId = 1
	}
	this.m_changed = true
	return this.m_data.NextExpeditionId
}

func (this *dbPlayerItemColumn) FillAllGmQueryMsg(ret_msg *msg_server_message.H2CItemQuery) {
	if nil == ret_msg {
		log.Error("dbPlayerItemColumn FillAllGmQueryMsg failed !")
		return
	}
	this.m_row.m_lock.UnSafeRLock("dbPlayerItemColumn.FillAllGmQueryMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()

	ret_msg.Items = make([]*msg_server_message.IdNum, 0, len(this.m_data))
	var tmp_idnum *msg_server_message.IdNum
	for _, tmp_val := range this.m_data {
		if nil == tmp_val {
			continue
		}

		tmp_idnum = &msg_server_message.IdNum{}
		tmp_idnum.Id = proto.Int32(tmp_val.ItemCfgId)
		tmp_idnum.Num = proto.Int32(tmp_val.ItemNum)
		ret_msg.Items = append(ret_msg.Items, tmp_idnum)
	}

	return
}
