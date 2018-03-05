package main

import (
	"libs/log"
	"public_message/gen_go/server_message"
	"time"
)

func (this *GmMgr) add_gm_server_reward(items []*msg_server_message.IdNum, content, channel string, last_sec int32) int32 {
	tmp_len := int32(len(items))
	if nil == items || tmp_len < 1 {
		log.Error("GmMgr add_gm_server_reward items nil or empty !")
		return -1
	}

	if nil == this.server_reward_db {
		log.Error("GmMgr add_gm_server_reward server_reward_db nil !")
		return -1
	}

	not_nil_count := int32(0)

	for idx := int32(0); idx < tmp_len; idx++ {
		if nil == items[idx] {
			continue
		}

		not_nil_count++
	}

	new_reward := &dbServerRewardRewardInfoData{}
	new_reward.RewardId = this.server_reward_db.IncGetNextRewardId()
	new_reward.Channel = channel
	new_reward.EndUnix = int32(time.Now().Unix()) + last_sec
	new_reward.Items = make([]dbIdNumData, not_nil_count)
	cur_count := int32(0)
	for idx := int32(0); idx < tmp_len; idx++ {
		if nil == items[idx] {
			continue
		}

		new_reward.Items[cur_count].Id = items[idx].GetId()
		new_reward.Items[cur_count].Num = items[idx].GetNum()
	}

	this.server_reward_db.RewardInfos.Add(new_reward)

	return new_reward.RewardId
}
