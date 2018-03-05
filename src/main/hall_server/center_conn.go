package main

import (
	"libs/log"
	"libs/server_conn"
	"libs/timer"
	"public_message/gen_go/client_message"
	"public_message/gen_go/server_message"
	"sync/atomic"
	"time"

	"3p/code.google.com.protobuf/proto"
)

const (
	CENTER_CONN_STATE_DISCONNECT  = 0
	CENTER_CONN_STATE_CONNECTED   = 1
	CENTER_CONN_STATE_FORCE_CLOSE = 2
)

type CenterConnection struct {
	client_node    *server_conn.Node
	state          int32
	last_conn_time int32

	connect_finished    bool
	connect_finish_chan chan int32
}

var center_conn CenterConnection

func (this *CenterConnection) Init() {
	this.client_node = server_conn.NewNode(this, 0, 0, 100, 0, 0, 0, 0, 0)
	this.client_node.SetDesc("中心服务器", "")
	this.state = CENTER_CONN_STATE_DISCONNECT
	this.RegisterMsgHandler()
	this.connect_finished = false
	this.connect_finish_chan = make(chan int32, 2)
}

func (this *CenterConnection) Start() {
	if this.Connect(CENTER_CONN_STATE_DISCONNECT) {
		log.Event("连接中心服务器成功", nil, log.Property{"IP", config.CenterServerIP})
	}
	for {
		state := atomic.LoadInt32(&this.state)
		if state == CENTER_CONN_STATE_CONNECTED {
			time.Sleep(time.Second * 2)
			continue
		}

		if state == CENTER_CONN_STATE_FORCE_CLOSE {
			this.client_node.ClientDisconnect()
			log.Event("与中心服务器的连接被强制关闭", nil)
			break
		}
		if this.Connect(state) {
			log.Event("连接中心服务器成功", nil, log.Property{"IP", config.CenterServerIP})
		}
	}
}

func (this *CenterConnection) Connect(state int32) (ok bool) {
	if CENTER_CONN_STATE_DISCONNECT == state {
		var err error
		for {
			log.Trace("连接中心服务器 %v", config.CenterServerIP)
			err = this.client_node.ClientConnect(config.CenterServerIP, time.Second*10)
			if nil == err {
				break
			}

			// 每隔30秒输出一次连接信息
			now := time.Now().Unix()
			if int32(now)-this.last_conn_time >= 30 {
				log.Trace("中心服务器连接中...")
				this.last_conn_time = int32(now)
			}
			time.Sleep(time.Second * 5)

			if signal_mgr.IfClosing() {
				this.state = CENTER_CONN_STATE_FORCE_CLOSE
				return
			}
		}
	}

	if atomic.CompareAndSwapInt32(&this.state, state, CENTER_CONN_STATE_CONNECTED) {
		go this.client_node.ClientRun()
		ok = true
	}
	return
}

func (this *CenterConnection) OnAccept(c *server_conn.ServerConn) {
	log.Error("Impossible accept")
}

func (this *CenterConnection) OnConnect(c *server_conn.ServerConn) {
	log.Trace("CenterServer [%v][%v] on CenterServer connect", config.ServerId, config.ServerName)

	notify := &msg_server_message.H2CHallServerRegister{}
	notify.ServerId = proto.Int32(config.ServerId)
	notify.ServerName = proto.String(config.ServerName)
	notify.ListenClientIP = proto.String(config.ListenClientOutIP)
	notify.ListenRoomIP = proto.String(config.ListenRoomServerIP)
	c.Send(notify, true)

}

func (this *CenterConnection) WaitConnectFinished() {
	for {

		if this.connect_finished {
			break
		}

		time.Sleep(time.Microsecond * 50)
	}

}

func (this *CenterConnection) OnUpdate(c *server_conn.ServerConn, t timer.TickTime) {

}

func (this *CenterConnection) OnDisconnect(c *server_conn.ServerConn, reason server_conn.E_DISCONNECT_REASON) {
	if reason == server_conn.E_DISCONNECT_REASON_FORCE_CLOSED {
		this.state = CENTER_CONN_STATE_FORCE_CLOSE
	} else {
		this.state = CENTER_CONN_STATE_DISCONNECT
	}
	log.Event("与中心服务器连接断开", nil)
}

func (this *CenterConnection) set_ih(type_id uint16, h server_conn.Handler) {
	t := msg_server_message.MessageTypes[type_id]
	if t == nil {
		log.Error("设置消息句柄失败，不存在的消息类型 %v", type_id)
		return
	}

	this.client_node.SetHandler(type_id, t, h)
}

type CenterMessageHandler func(a *CenterConnection, m proto.Message)

func (this *CenterConnection) SetMessageHandler(type_id uint16, h CenterMessageHandler) {
	if h == nil {
		this.set_ih(type_id, nil)
		return
	}

	this.set_ih(type_id, func(c *server_conn.ServerConn, m proto.Message) {
		h(this, m)
	})
}

func (this *CenterConnection) Send(msg proto.Message) {
	if CENTER_CONN_STATE_CONNECTED != this.state {
		log.Info("中心服务器未连接!!!")
		return
	}
	if nil == this.client_node {
		return
	}
	this.client_node.GetClient().Send(msg, true)
}

//========================================================================

func (this *CenterConnection) RegisterMsgHandler() {
	this.SetMessageHandler(msg_server_message.ID_C2HLoginServerList, C2HLoginServerListHandler)
	this.SetMessageHandler(msg_server_message.ID_C2HNewLoginServerAdd, C2HNewLoginServerAddHandler)
	this.SetMessageHandler(msg_server_message.ID_C2HLoginServerRemove, C2HLoginServerRemoveHandler)
	this.SetMessageHandler(msg_server_message.ID_PayBackAdd, C2HPayBackDataHandler)
	this.SetMessageHandler(msg_server_message.ID_SyncPayBackDataList, C2HSyncPayBackDataListHandler)
	this.SetMessageHandler(msg_server_message.ID_PayBackRemove, C2HPayBackRemoveHandler)
	this.SetMessageHandler(msg_server_message.ID_NoticeAdd, C2HNoticeAddHandler)
}

func C2HLoginServerListHandler(conn *CenterConnection, msg proto.Message) {
	req := msg.(*msg_server_message.C2HLoginServerList)
	if nil == conn || nil == req {
		log.Error("C2HLoginServerListHandler param error !")
		return
	}

	log.Info("中心服务器同步 登录服务器列表", req.GetServerList())

	login_conn_mgr.DisconnectAll()
	for _, info := range req.GetServerList() {
		login_conn_mgr.AddLogin(info)
	}

	conn.connect_finished = true
}

func C2HNewLoginServerAddHandler(conn *CenterConnection, msg proto.Message) {
	req := msg.(*msg_server_message.C2HNewLoginServerAdd)
	if nil == conn || nil == req || nil == req.GetServer() {
		log.Error("C2HNewLoginServerAddHandler param error !")
		return
	}

	cur_login := login_conn_mgr.GetLoginById(req.GetServer().GetServerId())
	if nil != cur_login {
		cur_login.ForceClose(true)
	}

	login_conn_mgr.AddLogin(req.GetServer())

	conn.connect_finished = true
}

func C2HLoginServerRemoveHandler(conn *CenterConnection, msg proto.Message) {
	req := msg.(*msg_server_message.C2HLoginServerRemove)
	if nil == conn || nil == req {
		log.Error("C2HLoginServerRemoveHandler param error !")
		return
	}

	serverid := req.GetServerId()
	cur_login := login_conn_mgr.GetLoginById(serverid)
	if nil != cur_login {
		log.Info("C2HLoginServerRemoveHandler 登录服务器[%d]连接还在，断开连接", serverid)
		cur_login.ForceClose(true)
		login_conn_mgr.RemoveLogin(serverid)
	}

	log.Info("中心服务器通知 LoginServer[%d] 断开", serverid)

	return
}

func C2HPayBackDataHandler(c *CenterConnection, msg proto.Message) {
	notify := msg.(*msg_server_message.PayBackAdd)
	if nil == c || nil == notify {
		log.Error("C2HPayBackDataHandler nil == c || nil == notify[%v]", nil == notify)
		return
	}

	payback_mgr.AddPayBack(notify)

	log.Info("PayBackMgr 增加 补偿 [%v]", *notify)
}

func C2HSyncPayBackDataListHandler(c *CenterConnection, msg proto.Message) {
	sync := msg.(*msg_server_message.SyncPayBackDataList)
	if nil == c || nil == sync {
		log.Error("C2HSyncPayBackDataListHandler c or sync nil[%v]", nil == sync)
		return
	}

	for _, tmp_pb := range sync.PayBackList {
		if nil == tmp_pb {
			continue
		}

		payback_mgr.AddPayBack(tmp_pb)
	}

	return
}

func C2HPayBackRemoveHandler(c *CenterConnection, msg proto.Message) {
	notify := msg.(*msg_server_message.PayBackRemove)
	if nil == c || nil == notify {
		log.Error("PayBackMgr C2HPayBackRemoveHandler c or notify nil[%v]", nil == notify)
		return
	}

	payback_mgr.RemovePayBack(notify.GetPbId())

	log.Info("PayBackMgr 删除补偿 [%v]", *notify)
	return
}

func C2HNoticeAddHandler(c *CenterConnection, msg proto.Message) {
	notify := msg.(*msg_server_message.NoticeAdd)
	if nil == notify || nil == c {
		log.Error("NoticeMgr C2HNoticeAddHandler c or notify nil [%v]", nil == notify)
		return
	}

	last_sec := notify.GetOverUnix() - int32(time.Now().Unix())

	res2cli := &msg_client_message.S2CNoticeAdd{}
	res2cli.Content = proto.String(notify.GetContent())
	res2cli.NoticeId = proto.Int32(notify.GetNoticeId())
	res2cli.LastSec = proto.Int32(last_sec)

	go player_mgr.SendMsgToAllPlayers(res2cli)

	notice_mgr.AddNotice(notify)
}
