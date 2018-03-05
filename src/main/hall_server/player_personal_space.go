package main

import (
	//"fmt"
	"libs/log"
	"net/http"
	"public_message/gen_go/client_message"
	"time"
	"youma/rpc_common"
	//"youma/table_config"

	"3p/code.google.com.protobuf/proto"
)

func reg_player_personl_space_msg() {
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SGetPersonalSpace, C2SGetPersonalSpaceHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SPersonalSpaceModifySignature, C2SPersonalSpaceModifySignatureHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SPersonalSpacePullLeaveMsg, C2SPersonalSpacePullLeaveMsgHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SPersonalSpaceSendLeaveMsg, C2SPersonalSpaceSendLeaveMsgHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SPersonalSpaceDelLeaveMsg, C2SPersonalSpaceDelLeaveMsgHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SPersonalSpaceSendLeaveMsgComment, C2SPersonalSpaceSendLeaveMsgCommentHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SPersonalSpaceDelLeaveMsgComment, C2SPersonalSpaceDelLeaveMsgCommentHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SPersonalSpacePullLeaveMsgComment, C2SPersonalSpacePullLeaveMsgCommentHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SPersonalSpaceZan, C2SPersonalSpaceZanHandler)
}

func rpc_leave_msg_comments_to_proto_leave_msg_comments(rpc_data []*rpc_common.H2R_PSLeaveMessageCommentData) (proto_data []*msg_client_message.PSLeaveMsgComment) {
	proto_data = make([]*msg_client_message.PSLeaveMsgComment, len(rpc_data))
	for j := 0; j < len(rpc_data); j++ {
		c := rpc_data[j]
		proto_data[j] = &msg_client_message.PSLeaveMsgComment{
			CommentId:    proto.Int32(c.Id),
			Content:      c.Content,
			SendTime:     proto.Int32(c.SendTime),
			SendPlayerId: proto.Int32(c.SendPlayerId),
		}
	}
	return
}

func rpc_leave_msgs_to_proto_leave_msgs(rpc_data []*rpc_common.H2R_PSLeaveMessageData) (proto_data []*msg_client_message.PSLeaveMsg) {
	proto_data = make([]*msg_client_message.PSLeaveMsg, len(rpc_data))
	for i := 0; i < len(rpc_data); i++ {
		m := rpc_data[i]
		proto_data[i] = &msg_client_message.PSLeaveMsg{
			MsgId:        proto.Int32(m.Id),
			Content:      m.Content,
			SendTime:     proto.Int32(m.SendTime),
			SendPlayerId: proto.Int32(m.SendPlayerId),
		}
		proto_data[i].Comments = rpc_leave_msg_comments_to_proto_leave_msg_comments(rpc_data[i].Comments)
	}
	return
}

func (this *Player) get_personal_space(player_id int32) int32 {
	if player_id == 0 {
		player_id = this.Id
	}

	result := this.rpc_player_get_personal_space(player_id)
	if result == nil {
		return -1
	}

	if result.Error < 0 {
		log.Error("Player[%v] get personal space error[%v]", this.Id, result.Error)
		return result.Error
	}

	response := &msg_client_message.S2CGetPersonalSpaceResult{}
	response.PlayerId = proto.Int32(player_id)
	response.Signature = proto.String(result.Signature)
	response.Pics = make([]*msg_client_message.PSPicData, len(result.Pictures))
	for i := 0; i < len(result.Pictures); i++ {
		p := result.Pictures[i]
		response.Pics[i] = &msg_client_message.PSPicData{
			PicId:        proto.Int32(p.Id),
			ThumbNailUrl: proto.String(p.Url),
			Zaned:        proto.Int32(p.Zaned),
			MsgNum:       proto.Int32(p.LeaveMsgNum),
		}
	}
	response.LeaveMsgs = rpc_leave_msgs_to_proto_leave_msgs(result.LeaveMsgs)
	var is_more int32
	if result.LeaveMsgIsMore {
		is_more = 1
	}
	response.IsMoreMsg = proto.Int32(is_more)

	this.Send(response)

	log.Debug("Player[%v] get personal space", this.Id)

	return 1
}

func (this *Player) personal_space_modify_signature(signature string) int32 {
	result := this.rpc_personal_space_modify_signature(signature)
	if result == nil {
		return -1
	}

	if result.Error < 0 {
		log.Error("Player[%v] modify personal space signature error[%v]", this.Id, result.Error)
		return result.Error
	}

	response := &msg_client_message.S2CPersonalSpaceModifySignatureResult{
		Signature: proto.String(signature),
	}

	this.Send(response)

	log.Debug("Player[%v] modified personal space signature to %v", this.Id, signature)

	return 1
}

func (this *Player) personal_space_zan(player_id, pic_id int32) int32 {
	result := this.rpc_personal_space_zan(player_id, pic_id)
	if result == nil {
		return -1
	}
	if result.Error < 0 {
		log.Error("Player[%v] zan player[%v] personal space pic[%v] error[%v]", this.Id, player_id, pic_id, result.Error)
		return result.Error
	}

	response := &msg_client_message.S2CPersonalSpaceZanResult{
		PlayerId: proto.Int32(player_id),
		PicId:    proto.Int32(pic_id),
		Zaned:    proto.Int32(result.Zaned),
	}
	this.Send(response)
	return 1
}

func (this *Player) personal_space_pull_leave_msg(player_id, pic_id, start_index, msg_num int32) int32 {
	if player_id == 0 {
		player_id = this.Id
	}
	result := this.rpc_pull_personal_space_leave_msg(player_id, pic_id, start_index, msg_num)
	if result == nil {
		return -1
	}

	if result.Error < 0 {
		log.Error("Player[%v] pull player[%v] pic[%v] leave msg with range[%v,%v] error[%v]", this.Id, player_id, pic_id, start_index, msg_num, result.Error)
		return result.Error
	}

	response := &msg_client_message.S2CPersonalSpacePullLeaveMsgResult{}
	response.PlayerId = proto.Int32(player_id)
	response.PicId = proto.Int32(pic_id)
	response.StartIndex = proto.Int32(start_index)
	response.MsgNum = proto.Int32(msg_num)
	response.LeaveMsgs = rpc_leave_msgs_to_proto_leave_msgs(result.LeaveMsgs)

	this.Send(response)

	log.Debug("Player[%v] pulled player[%v] pic[%v] leave msg with range[%v,%v]", this.Id, player_id, pic_id, start_index, msg_num)

	return 1
}

func (this *Player) personal_space_send_leave_msg(player_id, pic_id int32, content []byte) int32 {
	if player_id == 0 {
		player_id = this.Id
	}
	result := this.rpc_personal_space_send_leave_msg(player_id, pic_id, content)
	if result == nil {
		return -1
	}

	if result.Error < 0 {
		log.Error("Player[%v] send pic[%v] leave msg[%v] to player[%v] personal space", this.Id, pic_id, content, player_id)
		return result.Error
	}

	response := &msg_client_message.S2CPersonalSpaceSendLeaveMsgResult{}
	response.PlayerId = proto.Int32(player_id)
	response.PicId = proto.Int32(pic_id)
	response.MsgId = proto.Int32(result.MsgId)
	response.LeaveMsg = content
	response.SendTime = proto.Int32(int32(time.Now().Unix()))
	this.Send(response)

	log.Debug("Player[%v] send pic[%v] leave msg[%v] to player[%v] personal space", this.Id, pic_id, content, player_id)

	return 1
}

func (this *Player) personal_space_delete_leave_msg(player_id, pic_id, msg_id int32) int32 {
	result := this.rpc_personal_space_delete_leave_msg(player_id, pic_id, msg_id)
	if result == nil {
		return -1
	}

	if result.Error < 0 {
		log.Error("Player[%v] delete pic_id[%v] msg_id[%v] from player[%v] personal space", this.Id, pic_id, msg_id, player_id)
		return result.Error
	}

	response := &msg_client_message.S2CPersonalSpaceDelLeaveMsgResult{}
	response.PlayerId = proto.Int32(player_id)
	response.MsgId = proto.Int32(msg_id)
	this.Send(response)

	log.Debug("Player[%v] deleted pic_id[%v] leave msg_id[%v] in player[%v] personal space", this.Id, pic_id, msg_id, player_id)

	return 1
}

func (this *Player) personal_space_send_leave_msg_comment(player_id, pic_id, msg_id int32, comment []byte) int32 {
	result := this.rpc_personal_space_send_leave_msg_comment(player_id, pic_id, msg_id, comment)
	if result == nil {
		return -1
	}

	if result.Error < 0 {
		log.Error("Player[%v] send pic_id[%v] leave msg[%v] comment[%v] to player[%v] error[%v]", this.Id, pic_id, msg_id, comment, player_id, result.Error)
		return result.Error
	}

	response := &msg_client_message.S2CPersonalSpaceSendLeaveMsgCommentResult{}
	response.PlayerId = proto.Int32(player_id)
	response.PicId = proto.Int32(pic_id)
	response.MsgId = proto.Int32(msg_id)
	response.Comment = comment
	response.CommentId = proto.Int32(result.CommentId)

	this.Send(response)

	log.Debug("Player[%v] send pic_id[%v] leave msg_id[%v] comment[%v] to player[%v] personal space pic[%v]", this.Id, pic_id, msg_id, comment, player_id, pic_id)
	return 1
}

func (this *Player) personal_space_delete_leave_msg_comment(player_id, pic_id, msg_id, comment_id int32) int32 {
	result := this.rpc_personal_space_delete_leave_msg_comment(player_id, pic_id, msg_id, comment_id)
	if result == nil {
		return -1
	}

	if result.Error < 0 {
		log.Error("Player[%v] delete leave msg_id[%v] comment_id[%v] in player[%v] personal space pic[%v] error[%v]", this.Id, msg_id, comment_id, player_id, pic_id, result.Error)
		return result.Error
	}

	response := &msg_client_message.S2CPersonalSpaceDelLeaveMsgCommentResult{}
	response.PlayerId = proto.Int32(player_id)
	response.PicId = proto.Int32(pic_id)
	response.MsgId = proto.Int32(msg_id)
	response.CommentId = proto.Int32(comment_id)
	this.Send(response)

	log.Debug("Player[%v] deleted leave msg_id[%v] comment_id[%v] in player[%v] personal space pic[%v]", this.Id, msg_id, comment_id, player_id, pic_id)

	return 1
}

func (this *Player) personal_space_pull_leave_msg_comment(player_id, pic_id, msg_id, start_index, comment_num int32) int32 {
	result := this.rpc_personal_space_pull_leave_msg_comment(player_id, pic_id, msg_id, start_index, comment_num)
	if result == nil {
		return -1
	}

	if result.Error < 0 {
		log.Error("Player[%v] pull player[%v] personal space pic[%v] leave msg_id[%v] with range[%v,%v] error[%v]", this.Id, player_id, pic_id, msg_id, start_index, comment_num, result.Error)
		return result.Error
	}

	response := &msg_client_message.S2CPersonalSpacePullLeaveMsgCommentResult{}
	response.PlayerId = proto.Int32(player_id)
	response.PicId = proto.Int32(pic_id)
	response.MsgId = proto.Int32(msg_id)
	response.StartIndex = proto.Int32(start_index)
	response.CommentNum = proto.Int32(comment_num)
	response.Comments = rpc_leave_msg_comments_to_proto_leave_msg_comments(result.Comments)
	this.Send(response)

	log.Debug("Player[%v] pull player[%v] pic_id[%v] leave msg_id[%v] comments with range[%v,%v]", this.Id, player_id, pic_id, msg_id, start_index, comment_num)
	return 1
}

func C2SGetPersonalSpaceHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SGetPersonalSpace)
	if req == nil || p == nil {
		log.Error("C2SGetPersonalSpaceHandler proto invalid")
		return -1
	}
	return p.get_personal_space(req.GetPlayerId())
}

func C2SPersonalSpaceModifySignatureHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SPersonalSpaceModifySignature)
	if req == nil || p == nil {
		log.Error("C2SPersonalSpaceModifySignatureHandler proto invalid")
		return -1
	}
	return p.personal_space_modify_signature(req.GetSignature())
}

func C2SPersonalSpacePullLeaveMsgHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SPersonalSpacePullLeaveMsg)
	if req == nil || p == nil {
		log.Error("C2SPersonalSpacePullLeaveMsgHandler proto invalid")
		return -1
	}
	return p.personal_space_pull_leave_msg(req.GetPlayerId(), req.GetPicId(), req.GetStartIndex(), req.GetMsgNum())
}

func C2SPersonalSpaceSendLeaveMsgHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SPersonalSpaceSendLeaveMsg)
	if req == nil || p == nil {
		log.Error("C2SPersonalSpaceSendLeaveMsgHandler proto invalid")
		return -1
	}
	return p.personal_space_send_leave_msg(req.GetPlayerId(), req.GetPicId(), req.GetLeaveMsg())
}

func C2SPersonalSpaceDelLeaveMsgHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SPersonalSpaceDelLeaveMsg)
	if req == nil || p == nil {
		return -1
	}
	return p.personal_space_delete_leave_msg(req.GetPlayerId(), req.GetPicId(), req.GetMsgId())
}

func C2SPersonalSpaceSendLeaveMsgCommentHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SPersonalSpaceSendLeaveMsgComment)
	if req == nil || p == nil {
		return -1
	}
	return p.personal_space_send_leave_msg_comment(req.GetPlayerId(), req.GetPicId(), req.GetMsgId(), req.GetComment())
}

func C2SPersonalSpaceDelLeaveMsgCommentHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SPersonalSpaceDelLeaveMsgComment)
	if req == nil || p == nil {
		return -1
	}
	return p.personal_space_delete_leave_msg_comment(req.GetPlayerId(), req.GetPicId(), req.GetMsgId(), req.GetCommentId())
}

func C2SPersonalSpacePullLeaveMsgCommentHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SPersonalSpacePullLeaveMsgComment)
	if req == nil || p == nil {
		return -1
	}
	return p.personal_space_pull_leave_msg_comment(req.GetPlayerId(), req.GetPicId(), req.GetMsgId(), req.GetStartIndex(), req.GetCommentNum())
}

func C2SPersonalSpaceZanHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SPersonalSpaceZan)
	if req == nil || p == nil {
		return -1
	}
	return p.personal_space_zan(req.GetPlayerId(), req.GetPicId())
}
