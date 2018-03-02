package main

import (
	"libs/log"
	"public_message/gen_go/client_message"
	"time"
)

func (this *GmCommandMgr) add_server_reward(item *ServerReward) {
	if nil == item {
		log.Error("GmCommandMgr add_server_reward !")
		return
	}

	this.id2serverreward_lock.Lock()
	defer this.id2serverreward_lock.Unlock()

	this.id2serverreward[item.RewardId] = item

	return
}

func (this *GmCommandMgr) ChkServerRewardOver() {
	cur_unix := int32(time.Now().Unix())
	if this.last_chk_server_reward_sec <= 0 {
		this.last_chk_server_reward_sec = cur_unix
		return
	}

	if cur_unix-this.last_chk_server_reward_sec < GM_SERVER_REWARD_CHK_SEC {
		return
	}

	this.last_chk_server_reward_sec = cur_unix

	del_map := make(map[int32]bool)
	this.id2serverreward_lock.Lock()
	defer this.id2serverreward_lock.Unlock()

	for id, val := range this.id2serverreward {
		if nil == val || val.EndUnix < cur_unix {
			del_map[id] = true
		}
	}

	for id, _ := range del_map {
		delete(this.id2serverreward, id)
	}

	return
}

func (this *GmCommandMgr) PrintCurServerReward() {
	this.id2serverreward_lock.RLock()
	defer this.id2serverreward_lock.RUnlock()

	log.Info("*******************cur_server_reward*******************")
	cur_unix := int32(time.Now().Unix())

	for _, val := range this.id2serverreward {
		if nil == val {
			continue
		}

		log.Info("GmCommandMgr server_reward id[%d] ch[%s] cur_unix[%d] end_unx[%d] if_over[%v]", val.RewardId, val.Channel, cur_unix, val.EndUnix, cur_unix > val.EndUnix)
		for idx := int32(0); idx < int32(len(val.RewardItems)); idx++ {
			log.Info("	reward_item {id:%d, num:%d}", val.RewardItems[idx].GetId(), val.RewardItems[idx].GetNum())
		}
	}

	log.Info("**************************end**************************")
}

func (this *GmCommandMgr) ChkGetServerRewardByPlayer(p *Player) *msg_client_message.S2CMailList {

	log.Info("GmCommandMgr ChkGetServerRewardByPlayer %d", p.Id)
	this.PrintCurServerReward()

	this.ChkServerRewardOver()

	log.Info("GmCommandMgr ChkGetServerRewardByPlayer after chk")
	this.PrintCurServerReward()

	this.id2serverreward_lock.RLock()
	defer this.id2serverreward_lock.RUnlock()

	tmp_len := int32(len(this.id2serverreward))
	if tmp_len < 1 {
		return nil
	}

	cur_unix := int32(time.Now().Unix())
	ret_msg := &msg_client_message.S2CMailList{}
	ret_msg.MailList = make([]*msg_client_message.MailInfo, 0, tmp_len)
	for id, val := range this.id2serverreward {
		if nil == val {
			continue
		}

		if val.EndUnix < cur_unix {
			continue
		}

		if val.Channel != "" && p.db.Info.GetChannel() != val.Channel {
			continue
		}

		if nil == p.db.ServerRewards.Get(id) {
			p.SendGmRewardItemMail(val.ContentId, val.RewardItems, val.EndUnix-cur_unix, ret_msg)
		}
	}

	return ret_msg
}

func (this *GmCommandMgr) OnPlayerLogin(p *Player) {
	this.ChkGetServerRewardByPlayer(p)
}
