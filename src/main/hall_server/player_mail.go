package main

import (
	_ "libs/log"
	_ "libs/socket"
	_ "main/table_config"
	"public_message/gen_go/client_message"

	_ "3p/code.google.com.protobuf/proto"
	_ "github.com/golang/protobuf/proto"
)

func (this *Player) get_min_mail_id() int32 {
	all := this.db.Mails.GetAllIndex()
	if all == nil || len(all) == 0 {
		return 0
	}

	min_id := int32(0)
	for i := 0; i < len(all); i++ {
		if min_id == 0 || min_id > all[i] {
			min_id = all[i]
		}
	}
	return min_id
}

func (this *Player) new_mail(typ int32, title, content string) int32 {
	if this.db.Mails.NumAll() >= global_config_mgr.GetGlobalConfig().MailMaxCount {
		min_id := this.get_min_mail_id()
		this.db.Mails.Remove(min_id)
	}
	new_id := this.db.MailCommon.IncbyCurrId(1)
	this.db.Mails.Add(&dbPlayerMailData{
		Id:      new_id,
		Type:    int8(typ),
		Title:   title,
		Content: content,
	})
	return new_id
}

func (this *Player) attach_mail_item(mail_id, item_id, item_num int32) int32 {
	if !this.db.Mails.HasIndex(mail_id) {
		return int32(msg_client_message.E_ERR_PLAYER_MAIL_NOT_FOUND)
	}
	item_ids, _ := this.db.Mails.GetAttachItemIds(mail_id)
	item_nums, _ := this.db.Mails.GetAttachItemNums(mail_id)
	item_ids = append(item_ids, item_id)
	item_nums = append(item_nums, item_num)
	this.db.Mails.SetAttachItemIds(mail_id, item_ids)
	this.db.Mails.SetAttachItemNums(mail_id, item_nums)
	return 1
}

func (this *Player) delete_mail(mail_id int32) int32 {
	if !this.db.Mails.HasIndex(mail_id) {
		return int32(msg_client_message.E_ERR_PLAYER_MAIL_NOT_FOUND)
	}
	this.db.Mails.Remove(mail_id)
	return 1
}
