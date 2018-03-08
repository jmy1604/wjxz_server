package main

import (
	"libs/socket"
	"public_message/gen_go/client_message"

	"3p/code.google.com.protobuf/proto"
)

const (
	PLAYER_FIRST_PAY_NOT_ACT  = 0 // 首充奖励未激活
	PLAYER_FIRST_PAY_ACT      = 1 // 首充奖励未领取
	PLAYER_FIRST_PAY_REWARDED = 2 // 首充奖励已经领取
)

func (this *Player) SyncPlayerFirstPayState() {
	this.OnActivityValSet(PLAYER_ACTIVITY_TYPE_FIRST_PAY, 1)
	/*
		res2cli := &msg_client_message.S2CSyncFirstPayState{}
		res2cli.CurState = proto.Int32(this.db.Info.GetFirstPayState())
		this.Send(res2cli)
	*/
}

// ----------------------------------------------------------------------------

func reg_player_first_pay_msg() {
	hall_server.SetMessageHandler(msg_client_message.ID_C2SGetFirstPayReward, C2SGetFirstPayRewardHandler)
}

func C2SGetFirstPayRewardHandler(c *socket.TcpConn, msg proto.Message) {
	return
}
