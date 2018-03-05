package main

import (
	"libs/log"
	"public_message/gen_go/client_message"
	"public_message/gen_go/server_message"
	"sync"
	"time"

	"3p/code.google.com.protobuf/proto"
)

type NoticeMgr struct {
	id2notice_lock *sync.RWMutex
	id2notice      map[int32]*msg_server_message.NoticeAdd
}

var notice_mgr NoticeMgr

func (this *NoticeMgr) Init() bool {
	this.id2notice_lock = &sync.RWMutex{}
	this.id2notice = make(map[int32]*msg_server_message.NoticeAdd)

	return true
}

func (this *NoticeMgr) AddNotice(notice *msg_server_message.NoticeAdd) {
	if nil == notice {
		log.Error("NoticeMgr AddNotice notice nil !")
		return
	}

	this.id2notice_lock.Lock()
	defer this.id2notice_lock.Unlock()

	cur_n := this.id2notice[notice.GetNoticeId()]
	if nil != cur_n {
		log.Info("NoticeMgr AddNotice already exist [%d] !", notice.GetNoticeId())
	}

	this.id2notice[notice.GetNoticeId()] = notice

	return
}

func (this *NoticeMgr) RemoveNotice(notice_id int32) {
	this.id2notice_lock.Lock()
	defer this.id2notice_lock.Unlock()

	if nil != this.id2notice[notice_id] {
		delete(this.id2notice, notice_id)
	}

	return
}

func (this *NoticeMgr) FillNoticeMsg() *msg_client_message.S2CNoticeList {
	cur_unix := int32(time.Now().Unix())
	this.id2notice_lock.Lock()
	defer this.id2notice_lock.Unlock()

	tmp_len := int32(len(this.id2notice))
	if tmp_len <= 0 {
		return nil
	}

	ret_msg := &msg_client_message.S2CNoticeList{}
	var tmp_n *msg_client_message.S2CNoticeAdd
	for _, val := range this.id2notice {
		if nil == val {
			continue
		}

		left_sec := val.GetOverUnix() - cur_unix
		if left_sec <= 0 {
			continue
		}

		tmp_n = &msg_client_message.S2CNoticeAdd{}
		tmp_n.NoticeId = proto.Int32(val.GetNoticeId())
		tmp_n.Content = proto.String(val.GetContent())
		tmp_n.LastSec = proto.Int32(left_sec)
		ret_msg.NoticeList = append(ret_msg.NoticeList, tmp_n)
	}

	if len(ret_msg.NoticeList) > 0 {
		return ret_msg
	}

	return nil
}

func (this *NoticeMgr) OnPlayerLogin(p *Player) {
	if nil == p {
		log.Error("NoticeMgr OnPlayerLogin p nil !")
		return
	}

	res2cli := this.FillNoticeMsg()
	if nil != res2cli && len(res2cli.NoticeList) > 0 {
		p.Send(res2cli)
	}

	return
}
