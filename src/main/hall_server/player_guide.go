package main

import (
	_ "libs/log"
	_ "libs/socket"
	_ "public_message/gen_go/client_message"

	_ "3p/code.google.com.protobuf/proto"
	_ "github.com/golang/protobuf/proto"
)

func (this *Player) SyncPlayerGuideData() {
	/*res2cli := &msg_client_message.S2CSyncGuideData{}
	this.Send(res2cli)*/
}

// ----------------------------------------------------------------------------

func reg_player_guide_msg() {
	//hall_server.SetMessageHandler(msg_client_message.ID_C2SSaveGuideData, C2SSaveGuideDataHandler)
}

/*func C2SSaveGuideDataHandler(c *socket.TcpConn, msg proto.Message) {
	req := msg.(*msg_client_message.C2SSaveGuideData)
	if nil == c || nil == req {
		log.Error("C2SSaveGuideDataHandler c or req nil[%d]", nil == req)
		return
	}

	p := player_mgr.GetPlayerById(int32(c.T))
	if nil == p {
		log.Error("C2SSaveGuideDataHandler not login [%d]", c.T)
		return
	}

	guide_id := int32(req.GetGuideId())
	p.db.Guidess.ForceAdd(guide_id)

	res2cli := &msg_client_message.S2CRetSaveGuideData{}
	res2cli.GuideId = proto.Int32(guide_id)

	p.Send(res2cli)

	return
}*/
