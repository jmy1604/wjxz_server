package main

import (
	"libs/log"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/server_message"
	"time"

	"3p/code.google.com.protobuf/proto"
)

const (
	PLAYER_MAIL_SENDER_ID_SYSTEM = 1 // 系统邮件

	PLAYER_MAIL_SENDER_NAME_SYSYTEM = "System" // 系统邮件发送名称

	PLAYER_MAIL_TYPE_NORMAL   = 0 // 普通邮件
	PLAYER_MAIL_TYPE_REQ_HELP = 1 // 求助邮件

	PLAYER_MAIL_OP_TYPE_SYNC   = 0 // 邮件同步
	PLAYER_MAIL_OP_TYPE_ADD    = 1 // 邮件增加
	PLAYER_MAIL_OP_TYPE_REMOVE = 2 // 邮件删除
	PLAYER_MAIL_OP_TYPE_UPDATE = 3 // 邮件同步

	PLAYER_MAIL_STATE_INIT     = 0 // 邮件初始状态
	PLAYER_MAIL_STATE_READED   = 1 // 邮件已经读状态
	PLAYER_MAIL_STATE_HAVEDONE = 2 // 邮件已经领取附件或者确认帮助
)

func (this *Player) AddMail(msg *msg_server_message.MailAdd) {
	if nil == msg {
		log.Error("Player AddSystemMail msg nil")
		return
	}

	new_mail_id := this.db.Mails.GetAviMailId()

	new_mail := &dbPlayerMailData{}
	new_mail.MailId = new_mail_id
	new_mail.MailType = int8(msg.GetMailType())
	new_mail.SenderId = PLAYER_MAIL_SENDER_ID_SYSTEM
	new_mail.SenderName = msg.GetSenderName()
	new_mail.Content = msg.GetContent()
	new_mail.MailTitle = msg.GetTitle()
	new_mail.ObjIds = msg.GetObjIds()
	new_mail.ObjNums = msg.GetObjNums()
	new_mail.SendUnix = msg.GetSendUnix()
	new_mail.OverUnix = msg.GetOverUnix()

	this.db.Mails.Add(new_mail)

	return
}

func (this *Player) SendTestMail(mail_type, sender_id int32, sender_name string, objids, nums []int32) {
	new_mail := &dbPlayerMailData{}
	new_mail.MailId = this.db.Mails.GetAviMailId()
	new_mail.MailType = int8(mail_type)
	new_mail.MailTitle = "test_title"
	new_mail.Content = "dadadadadadadasdsadsadasdsadsadsadasdsadsa"
	new_mail.SenderId = sender_id
	new_mail.SenderName = sender_name
	new_mail.ObjIds = objids
	new_mail.ObjNums = nums
	new_mail.SendUnix = int32(time.Now().Unix())
	new_mail.OverUnix = int32(time.Now().Unix()) + global_config_mgr.GetGlobalConfig().NormalMailLastSec

	log.Info("SendTestMail cur_unix[%d] over_unix[%d] cfg_over_sec[%d]", new_mail.SendUnix, new_mail.OverUnix, global_config_mgr.GetGlobalConfig().NormalMailLastSec)
	log.Info("SendTestMail mail_id[%d] sender_id[%d] sender_name[%s] objids%v objnums", new_mail.MailId, sender_id, sender_name, objids, nums)

	this.db.Mails.Add(new_mail)

	return
}

func (this *Player) SendHelpMail(sender_id int32, sender_name string, chapter_id int32) {
	new_mail := &dbPlayerMailData{}
	new_mail.MailId = this.db.Mails.GetAviMailId()
	new_mail.MailType = PLAYER_MAIL_TYPE_REQ_HELP
	new_mail.SenderId = sender_id
	new_mail.SenderName = sender_name
	new_mail.ExtraDatas = make([]int32, 1)
	new_mail.ExtraDatas[0] = chapter_id
	new_mail.SendUnix = int32(time.Now().Unix())
	new_mail.OverUnix = int32(time.Now().Unix()) + global_config_mgr.GetGlobalConfig().ReqHelpMailLastSec
	this.db.Mails.Add(new_mail)

	return
}

func (this *Player) SendRewardMail(title_id, content_id string, rewards []int32, bslience bool) {
	new_mail := &dbPlayerMailData{}
	new_mail.MailId = this.db.Mails.GetAviMailId()
	new_mail.MailType = PLAYER_MAIL_TYPE_NORMAL
	new_mail.MailTitle = title_id
	new_mail.Content = content_id
	tmp_len := int32(len(rewards))
	new_mail.ObjIds = make([]int32, 0, tmp_len/2)
	new_mail.ObjNums = make([]int32, 0, tmp_len/2)
	var tmp_mail_add *msg_client_message.MailInfo

	log.Info("SendRewardMail", title_id, content_id, rewards, bslience)

	for idx := int32(0); idx+1 < tmp_len; idx += 2 {
		new_mail.ObjIds = append(new_mail.ObjIds, rewards[idx])
		new_mail.ObjNums = append(new_mail.ObjNums, rewards[idx+1])

	}

	new_mail.SendUnix = int32(time.Now().Unix())
	new_mail.OverUnix = int32(time.Now().Unix()) + global_config_mgr.GetGlobalConfig().ReqHelpMailLastSec
	this.db.Mails.Add(new_mail)

	if !bslience {
		res2cli := &msg_client_message.S2CMailList{}
		res2cli.MailList = make([]*msg_client_message.MailInfo, 1)
		tmp_mail_add = &msg_client_message.MailInfo{}
		tmp_mail_add.MailId = proto.Int32(new_mail.MailId)
		tmp_mail_add.MailType = proto.Int32(PLAYER_MAIL_TYPE_NORMAL)
		tmp_mail_add.Title = proto.String(new_mail.MailTitle)
		tmp_mail_add.Content = proto.String(new_mail.Content)
		tmp_mail_add.LeftSec = proto.Int32(global_config_mgr.GetGlobalConfig().ReqHelpMailLastSec)
		tmp_mail_add.OpType = proto.Int32(PLAYER_MAIL_OP_TYPE_SYNC)
		tmp_mail_add.ObjIds = new_mail.ObjIds
		tmp_mail_add.ObjNums = new_mail.ObjNums

		res2cli.MailList[0] = tmp_mail_add
		this.Send(res2cli)

	}

	return
}

func (this *Player) SendGmItemMail(content_id string, item_infos []*ItemInfo, last_sec int32, bslience bool) {
	new_mail := &dbPlayerMailData{}
	new_mail.MailId = this.db.Mails.GetAviMailId()
	new_mail.MailType = PLAYER_MAIL_TYPE_NORMAL
	new_mail.MailTitle = ""
	new_mail.Content = content_id

	log.Info("Player SendGmItemMail ", content_id, item_infos, bslience)

	tmp_len := int32(len(item_infos))
	new_mail.ObjIds = make([]int32, 0, tmp_len)
	new_mail.ObjNums = make([]int32, 0, tmp_len)
	var tmp_mail_add *msg_client_message.MailInfo

	for idx := int32(0); idx < tmp_len; idx++ {
		new_mail.ObjIds = append(new_mail.ObjIds, item_infos[idx].Id)
		new_mail.ObjNums = append(new_mail.ObjNums, item_infos[idx].Num)

	}

	new_mail.SendUnix = int32(time.Now().Unix())
	new_mail.OverUnix = int32(time.Now().Unix()) + last_sec
	this.db.Mails.Add(new_mail)

	if bslience {
		res2cli := &msg_client_message.S2CMailList{}
		res2cli.MailList = make([]*msg_client_message.MailInfo, 1)
		tmp_mail_add = &msg_client_message.MailInfo{}
		tmp_mail_add.MailId = proto.Int32(new_mail.MailId)
		tmp_mail_add.MailType = proto.Int32(PLAYER_MAIL_TYPE_NORMAL)
		tmp_mail_add.Title = proto.String(new_mail.MailTitle)
		tmp_mail_add.Content = proto.String(new_mail.Content)
		tmp_mail_add.LeftSec = proto.Int32(last_sec)
		tmp_mail_add.OpType = proto.Int32(PLAYER_MAIL_OP_TYPE_SYNC)
		tmp_mail_add.ObjIds = new_mail.ObjIds
		tmp_mail_add.ObjNums = new_mail.ObjNums

		res2cli.MailList[0] = tmp_mail_add

		this.Send(res2cli)
	}

	return
}

func (this *Player) SendGmRewardItemMail(content_id string, item_infos []*msg_server_message.IdNum, last_sec int32, msg *msg_client_message.S2CMailList) {
	new_mail := &dbPlayerMailData{}
	new_mail.MailId = this.db.Mails.GetAviMailId()
	new_mail.MailType = PLAYER_MAIL_TYPE_NORMAL
	new_mail.MailTitle = ""
	new_mail.Content = content_id

	tmp_len := int32(len(item_infos))
	new_mail.ObjIds = make([]int32, 0, tmp_len)
	new_mail.ObjNums = make([]int32, 0, tmp_len)
	var tmp_mail_add *msg_client_message.MailInfo

	for idx := int32(0); idx < tmp_len; idx++ {
		new_mail.ObjIds = append(new_mail.ObjIds, item_infos[idx].GetId())
		new_mail.ObjNums = append(new_mail.ObjNums, item_infos[idx].GetNum())
	}

	new_mail.SendUnix = int32(time.Now().Unix())
	new_mail.OverUnix = int32(time.Now().Unix()) + last_sec
	this.db.Mails.Add(new_mail)

	tmp_mail_add = &msg_client_message.MailInfo{}
	tmp_mail_add.MailId = proto.Int32(new_mail.MailId)
	tmp_mail_add.MailType = proto.Int32(PLAYER_MAIL_TYPE_NORMAL)
	tmp_mail_add.Title = proto.String(new_mail.MailTitle)
	tmp_mail_add.Content = proto.String(new_mail.Content)
	tmp_mail_add.LeftSec = proto.Int32(last_sec)
	tmp_mail_add.OpType = proto.Int32(PLAYER_MAIL_OP_TYPE_SYNC)
	tmp_mail_add.ObjIds = new_mail.ObjIds
	tmp_mail_add.ObjNums = new_mail.ObjNums

	msg.MailList = append(msg.MailList, tmp_mail_add)

	return
}

// ===============================================================

func reg_player_mail_msg() {
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SGetMailList, C2SGetMailListHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SGetMailAttach, C2SGetMailAttachHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SSetMailRead, C2SSetMailReadHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SMailRemove, C2SMailRemoveHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SAgreeMailHelpReq, C2SAgreeMailHelpReqHandler)
	//hall_server.SetMessageHandler(msg_client_message.ID_C2SGetMailList, C2SGetMailListHandler)
	//hall_server.SetMessageHandler(msg_client_message.ID_C2SGetMailAttach, C2SGetMailAttachHandler)
	//hall_server.SetMessageHandler(msg_client_message.ID_C2SMailRemove, C2SMailRemoveHandler)
	//hall_server.SetMessageHandler(msg_client_message.ID_C2SSetMailRead, C2SSetMailReadHandler)

	center_conn.SetMessageHandler(msg_server_message.ID_MailAdd, C2HMailAddHandler)
}

func C2SGetMailListHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SGetMailList)
	if nil == req {
		log.Error("C2SGetMailListHandler c or req nil[%v]", nil == req)
		return -1
	}

	res2cli := p.db.Mails.FillMsgList()
	if nil != res2cli {
		p.Send(res2cli)
		return 1
	}

	return 0
}

func C2SGetMailAttachHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SGetMailAttach)
	if nil == req {
		log.Error("C2SGetMailAttachHandler c or req nil[%v]", nil == req)
		return -1
	}

	mail_id := req.GetMailId()
	var db_mail *dbPlayerMailData
	var tmp_len int
	var obj_id, obj_num int32
	res2cli := &msg_client_message.S2CMailList{}
	if mail_id == -1 {
		all_mail_ids := p.db.Mails.GetAllIndex()
		res2cli.MailList = make([]*msg_client_message.MailInfo, 0, len(all_mail_ids))
		for _, mail_id := range all_mail_ids {
			db_mail = p.db.Mails.Get(mail_id)
			if nil == db_mail {
				log.Error("C2SGetMailListHandler failed to get mail[%d]", mail_id)
				return int32(msg_client_message.E_ERR_MAIL_FAILED_TO_FIND_MAIL)
			}

			tmp_len = len(db_mail.ObjIds)
			if tmp_len != len(db_mail.ObjNums) {
				return int32(msg_client_message.E_ERR_MAIL_ATTACH_ERROR)
			}

			for idx := 0; idx < tmp_len; idx++ {
				obj_id = db_mail.ObjIds[idx]
				obj_num = db_mail.ObjNums[idx]

				if nil != item_table_mgr.Map[obj_id] {
					p.AddItem(obj_id, obj_num, "expedition_finish", "expedition", true)
				} else if nil != cfg_building_mgr.Map[obj_id] {
					p.AddDepotBuilding(obj_id, obj_num, "expedition_finish", "expedition", true)
				} else if nil != cat_table_mgr.Map[obj_id] {
					p.AddCat(obj_id, "expedition_finish", "expedition", true)
				} else {
					p.AddItemResource(obj_id, obj_num, "expedition_finish", "expedition")
				}
			}

			p.db.Mails.SetState(mail_id, PLAYER_MAIL_STATE_HAVEDONE)

			tmp_info := &msg_client_message.MailInfo{}
			tmp_info.OpType = proto.Int32(PLAYER_MAIL_OP_TYPE_UPDATE)
			tmp_info.MailId = proto.Int32(db_mail.MailId)
			tmp_info.MailType = proto.Int32(int32(db_mail.MailType))
			tmp_info.Title = proto.String(db_mail.MailTitle)
			tmp_info.SenderId = proto.Int32(db_mail.SenderId)
			tmp_info.SenderName = proto.String(db_mail.SenderName)
			tmp_info.Content = proto.String(db_mail.Content)
			tmp_info.ObjIds = db_mail.ObjIds
			tmp_info.ObjNums = db_mail.ObjNums
			tmp_info.SendUnix = proto.Int32(db_mail.SendUnix)
			tmp_info.LeftSec = proto.Int32(db_mail.OverUnix - int32(time.Now().Unix()))
			tmp_info.State = proto.Int32(PLAYER_MAIL_STATE_HAVEDONE)

			res2cli.MailList = append(res2cli.MailList, tmp_info)

			p.Send(res2cli)
		}
	} else {
		res2cli.MailList = make([]*msg_client_message.MailInfo, 1)
		db_mail = p.db.Mails.Get(mail_id)
		if nil == db_mail {
			log.Error("C2SGetMailListHandler failed to get mail[%d]", mail_id)
			return int32(msg_client_message.E_ERR_MAIL_FAILED_TO_FIND_MAIL)
		}

		tmp_len = len(db_mail.ObjIds)
		if tmp_len != len(db_mail.ObjNums) {
			return int32(msg_client_message.E_ERR_MAIL_ATTACH_ERROR)
		}

		for idx := 0; idx < tmp_len; idx++ {
			obj_id = db_mail.ObjIds[idx]
			obj_num = db_mail.ObjNums[idx]

			if nil != item_table_mgr.Map[obj_id] {
				p.AddItem(obj_id, obj_num, "expedition_finish", "expedition", true)
			} else if nil != cfg_building_mgr.Map[obj_id] {
				p.AddDepotBuilding(obj_id, obj_num, "expedition_finish", "expedition", true)
			} else if nil != cat_table_mgr.Map[obj_id] {
				p.AddCat(obj_id, "expedition_finish", "expedition", true)
			} else {
				p.AddItemResource(obj_id, obj_num, "expedition_finish", "expedition")
			}
		}

		p.db.Mails.SetState(mail_id, PLAYER_MAIL_STATE_HAVEDONE)

		tmp_info := &msg_client_message.MailInfo{}
		tmp_info.OpType = proto.Int32(PLAYER_MAIL_OP_TYPE_UPDATE)
		tmp_info.MailId = proto.Int32(db_mail.MailId)
		tmp_info.MailType = proto.Int32(int32(db_mail.MailType))
		tmp_info.Title = proto.String(db_mail.MailTitle)
		tmp_info.SenderId = proto.Int32(db_mail.SenderId)
		tmp_info.SenderName = proto.String(db_mail.SenderName)
		tmp_info.Content = proto.String(db_mail.Content)
		tmp_info.ObjIds = db_mail.ObjIds
		tmp_info.ObjNums = db_mail.ObjNums
		tmp_info.SendUnix = proto.Int32(db_mail.SendUnix)
		tmp_info.LeftSec = proto.Int32(db_mail.OverUnix - int32(time.Now().Unix()))
		tmp_info.State = proto.Int32(PLAYER_MAIL_STATE_HAVEDONE)
		res2cli := &msg_client_message.S2CMailList{}
		res2cli.MailList = make([]*msg_client_message.MailInfo, 1)
		res2cli.MailList[0] = tmp_info

		p.Send(res2cli)
	}

	p.SendItemsUpdate()
	p.SendDepotBuildingUpdate()
	p.SendCatsUpdate()

	return 1
}

func C2SSetMailReadHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SSetMailRead)
	if nil == req {
		log.Error("C2SSetMailReadHandler c or req nil [%v]", nil == req)
		return -1
	}

	mail_id := req.GetMailId()
	db_mail := p.db.Mails.Get(mail_id)
	if nil == db_mail {
		return int32(msg_client_message.E_ERR_MAIL_FAILED_TO_FIND_MAIL)
	}

	tmp_info := &msg_client_message.MailInfo{}
	tmp_info.OpType = proto.Int32(PLAYER_MAIL_OP_TYPE_UPDATE)
	tmp_info.MailId = proto.Int32(db_mail.MailId)
	tmp_info.MailType = proto.Int32(int32(db_mail.MailType))
	tmp_info.Title = proto.String(db_mail.MailTitle)
	tmp_info.SenderId = proto.Int32(db_mail.SenderId)
	tmp_info.SenderName = proto.String(db_mail.SenderName)
	tmp_info.Content = proto.String(db_mail.Content)
	tmp_info.ObjIds = db_mail.ObjIds
	tmp_info.ObjNums = db_mail.ObjNums
	tmp_info.SendUnix = proto.Int32(db_mail.SendUnix)
	tmp_info.LeftSec = proto.Int32(db_mail.OverUnix - int32(time.Now().Unix()))
	tmp_info.State = proto.Int32(PLAYER_MAIL_STATE_READED)

	res2cli := &msg_client_message.S2CMailList{}
	res2cli.MailList = make([]*msg_client_message.MailInfo, 1)
	res2cli.MailList[0] = tmp_info

	p.db.Mails.SetState(mail_id, PLAYER_MAIL_STATE_READED)

	p.Send(res2cli)

	return 1
}

func C2SMailRemoveHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SMailRemove)
	if nil == req {
		log.Error("C2SMailRemoveHandler c or req nil[%v]", nil == req)
		return -1
	}

	mail_id := req.GetMailId()
	tmp_info := &msg_client_message.MailInfo{}
	tmp_info.OpType = proto.Int32(PLAYER_MAIL_OP_TYPE_REMOVE)
	tmp_info.MailId = proto.Int32(mail_id)
	res2cli := &msg_client_message.S2CMailList{}
	res2cli.MailList = make([]*msg_client_message.MailInfo, 1)
	res2cli.MailList[0] = tmp_info
	p.db.Mails.Remove(mail_id)

	p.Send(res2cli)

	log.Info("玩家[%d] 删除邮件[%d]", p.Id, mail_id)
	return 1
}

func C2SAgreeMailHelpReqHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SAgreeMailHelpReq)
	if nil == req {
		log.Error("C2SAgreeMailHelpReqHandler c or req nil[%v]", nil == req)
		return -1
	}

	return 0
}

// -----------------------------------------------------

func C2HMailAddHandler(c *CenterConnection, msg proto.Message) {
	req := msg.(*msg_server_message.MailAdd)
	if nil == c || nil == req {
		log.Error("C2HMailAddHandler c or req nil [%v]", nil == req)
		return
	}

	p := player_mgr.GetPlayerById(req.GetPlayerId())
	if nil == p {
		log.Error("C2HMailAddHandler failed to get player(%d)", req.GetPlayerId())
		return
	}

	p.AddMail(req)

	return
}
