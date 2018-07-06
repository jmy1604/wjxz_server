package main

import (
	"libs/log"
	"libs/utils"
	_ "main/table_config"
	_ "math/rand"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	_ "sync"
	"time"

	"github.com/golang/protobuf/proto"
)

func (this *Player) _send_active_stage_data() {
	last_refresh := this.db.ActiveStage.GetLastRefreshTime()
	response := &msg_client_message.S2CActiveStageDataResponse{
		CanChallengeNum:            this.db.ActiveStage.GetCanChallengeNum(),
		MaxChallengeNum:            global_config.ActiveStageChallengeNumOfDay,
		RemainSeconds4ChallengeNum: utils.GetRemainSeconds2NextDayTime(last_refresh, global_config.ActiveStageRefreshTime),
		ChallengeNumPrice:          global_config.ActiveStageChallengeNumPrice,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ACTIVE_STAGE_DATA_RESPONSE), response)
}

func (this *Player) check_active_stage_refresh() bool {
	// 固定时间点自动刷新
	if global_config.ActiveStageRefreshTime == "" {
		return false
	}

	now_time := int32(time.Now().Unix())
	last_refresh := this.db.ActiveStage.GetLastRefreshTime()
	if last_refresh == 0 {
		this._send_active_stage_data()
		this.db.ActiveStage.SetLastRefreshTime(now_time)
	} else {
		if !utils.CheckDayTimeArrival(last_refresh, global_config.ActiveStageRefreshTime) {
			return false
		}

		this.db.ActiveStage.SetCanChallengeNum(global_config.ActiveStageChallengeNumOfDay)
		this.db.ActiveStage.SetLastRefreshTime(now_time)

		notify := &msg_client_message.S2CActiveStageRefreshNotify{}
		this.Send(uint16(msg_client_message_id.MSGID_S2C_ACTIVE_STAGE_REFRESH_NOTIFY), notify)
	}

	log.Debug("Player[%v] active stage refreshed", this.Id)
	return true
}

func (this *Player) send_active_stage_data() int32 {
	if this.check_active_stage_refresh() {
		return 1
	}
	this._send_active_stage_data()
	return 1
}

func C2SActiveStageDataHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SActiveStageDataRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}
	return p.send_active_stage_data()
}
