package main

import (
	"public_message/gen_go/server_message"
	"time"

	"3p/code.google.com.protobuf/proto"
)

func (this *DBC) on_preload() (err error) {

	return
}

func (this *dbPlayerIdMaxRow) Inc() (id int32) {
	this.m_lock.Lock("dbPlayerIdMaxRow.Inc")
	defer this.m_lock.Unlock()

	this.m_PlayerIdMax++
	id = this.m_PlayerIdMax

	this.m_PlayerIdMax_changed = true

	return
}

func (this *dbServerRewardRow) IncGetNextRewardId() (r int32) {
	this.m_lock.UnSafeLock("dbServerRewardRow.IncGetNextRewardId")
	defer this.m_lock.UnSafeUnlock()
	this.m_NextRewardId++
	this.m_NextRewardId_changed = true
	return int32(this.m_NextRewardId)
}

func (this *dbServerRewardRewardInfoColumn) ChkFillAllRewardMsg() *msg_server_message.C2HSyncServerReward {
	cur_unix := int32(time.Now().Unix())

	this.m_row.m_lock.UnSafeLock("dbServerRewardRewardInfoColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()

	del_ids := make(map[int32]bool)

	tmp_len := int32(len(this.m_data))
	for rewardid, d := range this.m_data {
		if nil == d || cur_unix > d.EndUnix {
			del_ids[rewardid] = true
			tmp_len--
			continue
		}
	}

	for del_id, _ := range del_ids {
		delete(this.m_data, del_id)
	}

	if tmp_len < 1 {
		return nil
	}

	ret_msg := &msg_server_message.C2HSyncServerReward{}
	ret_msg.Rewards = make([]*msg_server_message.C2HAddServerReward, 0, tmp_len)
	var tmp_re *msg_server_message.C2HAddServerReward
	var items_len int32
	var msg_idnum *msg_server_message.IdNum
	for rewardid, d := range this.m_data {
		tmp_re = &msg_server_message.C2HAddServerReward{}
		tmp_re.Channel = proto.String(d.Channel)
		tmp_re.Content = proto.String(d.Content)
		tmp_re.EndUnix = proto.Int32(d.EndUnix)
		tmp_re.RewardId = proto.Int32(rewardid)
		items_len = int32(len(d.Items))
		if items_len > 0 {
			tmp_re.Items = make([]*msg_server_message.IdNum, 0, items_len)
			for idx := int32(0); idx < items_len; idx++ {
				msg_idnum = &msg_server_message.IdNum{}
				msg_idnum.Id = proto.Int32(d.Items[idx].Id)
				msg_idnum.Num = proto.Int32(d.Items[idx].Num)
				tmp_re.Items = append(tmp_re.Items, msg_idnum)
			}
		}

		ret_msg.Rewards = append(ret_msg.Rewards, tmp_re)
	}

	return ret_msg
}
