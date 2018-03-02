package main

import (
	"libs/log"
	"libs/server_conn"
	"libs/timer"
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
}

var center_conn CenterConnection

func (this *CenterConnection) Init() {
	this.client_node = server_conn.NewNode(this, 0, 0, 100, 0, 0, 0, 0, 0)
	this.client_node.SetDesc("中心服务器", "")

	this.state = CENTER_CONN_STATE_DISCONNECT
	this.RegisterMsgHandler()
}

func (this *CenterConnection) Start() {
	if this.Connect(CENTER_CONN_STATE_DISCONNECT) {
		log.Event("连接CenterServer成功", nil, log.Property{"IP", config.CenterServerIP})
	}
	for {
		state := atomic.LoadInt32(&this.state)
		if state == CENTER_CONN_STATE_CONNECTED {
			time.Sleep(time.Second * 2)
			continue
		}

		if state == CENTER_CONN_STATE_FORCE_CLOSE {
			this.client_node.ClientDisconnect()
			log.Event("与CenterServer的连接被强制关闭", nil)
			break
		}
		if this.Connect(state) {
			log.Event("连接CenterServer成功", nil, log.Property{"IP", config.CenterServerIP})
		}
	}
}

func (this *CenterConnection) Connect(state int32) (ok bool) {
	if CENTER_CONN_STATE_DISCONNECT == state {
		var err error
		for CENTER_CONN_STATE_FORCE_CLOSE != this.state {
			log.Trace("连接CenterServer %v", config.CenterServerIP)
			err = this.client_node.ClientConnect(config.CenterServerIP, time.Second*10)
			if nil == err {
				break
			}

			// 每隔30秒输出一次连接信息
			now := time.Now().Unix()
			if int32(now)-this.last_conn_time >= 30 {
				log.Trace("CenterServer连接中...")
				this.last_conn_time = int32(now)
			}
			time.Sleep(time.Second * 5)
		}
	}

	if CENTER_CONN_STATE_FORCE_CLOSE != this.state && atomic.CompareAndSwapInt32(&this.state, state, CENTER_CONN_STATE_CONNECTED) {
		go this.client_node.ClientRun()
		ok = true
	}
	return
}

func (this *CenterConnection) OnAccept(c *server_conn.ServerConn) {
	log.Error("Impossible accept")
}

func (this *CenterConnection) OnConnect(c *server_conn.ServerConn) {
	if CENTER_CONN_STATE_FORCE_CLOSE != this.state {
		log.Trace("LoginServer[%v][%v] on CenterServer connect", config.ServerId, config.ServerName)
		notify := &msg_server_message.L2CLoginServerRegister{}
		notify.ServerId = proto.Int32(config.ServerId)
		notify.ServerName = proto.String(config.ServerName)
		notify.ListenMatchIP = proto.String(config.ListenMatchIP)
		c.Send(notify, true)
	} else {
		log.Trace("LoginServer[%v][%v] force closed on CenterServer connect", config.ServerId, config.ServerName)
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
	log.Event("与CenterServer连接断开", nil)
}

func (this *CenterConnection) ShutDown() {
	this.state = CENTER_CONN_STATE_FORCE_CLOSE
	if nil != this.client_node {
		this.client_node.Shutdown()
	}

}

func (this *CenterConnection) set_ih(type_id uint16, h server_conn.Handler) {
	t := msg_server_message.MessageTypes[type_id]
	if t == nil {
		log.Error("设置消息句柄失败，不存在的消息类型 %v", type_id)
		return
	}

	this.client_node.SetHandler(type_id, t, h)
}

type AssistMessageHandler func(a *CenterConnection, m proto.Message)

func (this *CenterConnection) SetMessageHandler(type_id uint16, h AssistMessageHandler) {
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
		log.Info("CenterServer未连接!!!")
		return
	}

	if nil == this.client_node {
		return
	}

	this.client_node.GetClient().Send(msg, false)
}

//========================================================================

func (this *CenterConnection) RegisterMsgHandler() {
	this.SetMessageHandler(msg_server_message.ID_C2LPlayerAccInfo, C2LPlayerAccInfoHandler)
}

func C2LPlayerAccInfoHandler(conn *CenterConnection, msg proto.Message) {
	res := msg.(*msg_server_message.C2LPlayerAccInfo)
	if nil == conn || nil == res {
		log.Error("C2LPlayerAccInfoHandler param error !")
		return
	}

	hallid := res.GetHallId()
	hall_agent := hall_agent_manager.GetAgentByID(hallid)
	if nil == hall_agent {
		log.Error("C2LPlayerAccInfoHandler can not find hall(%d)", hallid)
		return
	}

	acc := res.GetAccount()
	if "" == acc {
		log.Error("C2LPlayerAccInfoHandler acc empty")
		return
	}

	c_wait := server.pop_c_wait_by_acc(acc)
	if nil == c_wait {
		log.Error("C2LPlayerAccInfoHandler failed to get c_wait by acc(%s) !", acc)
		return
	}

	go send_res_to_wait(res, c_wait)

	return
}

func send_res_to_wait(res *msg_server_message.C2LPlayerAccInfo, c_wait *WaitCenterInfo) {
	c_wait.res_chan <- res
	log.Trace("C2MAccountInfoResponseHandler %v", res)
}
