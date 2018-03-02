package main

import (
	"io/ioutil"
	"libs/log"
	"net"
	"net/http"
	"public_message/gen_go/client_message"
	"reflect"
	"strings"
	"time"

	"3p/code.google.com.protobuf/proto"
)

var msg_handler_http_mux map[string]func(http.ResponseWriter, *http.Request)

type MsgHttpHandle struct{}

func (this *MsgHttpHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var act_str, url_str string
	url_str = r.URL.String()
	idx := strings.Index(url_str, "?")
	if -1 == idx {
		act_str = url_str
	} else {
		act_str = string([]byte(url_str)[:idx])
	}
	log.Info("hall msg handler ServeHTTP actstr(%s)", act_str)
	if h, ok := msg_handler_http_mux[act_str]; ok {
		h(w, r)
	}

	return
}

//=======================================================

type CLIENT_MSG_HANDLER func(http.ResponseWriter, *http.Request, proto.Message) (int32, *Player)

type CLIENT_PLAYER_MSG_HANDLER func(http.ResponseWriter, *http.Request, *Player, proto.Message) int32

type MsgHanlderInfo struct {
	typ                reflect.Type
	msg_handler        CLIENT_MSG_HANDLER
	player_msg_handler CLIENT_PLAYER_MSG_HANDLER
	if_player_msg      bool
}

type MsgHandlerMgr struct {
	msg_http_listener net.Listener
	msgid2handler     map[int32]*MsgHanlderInfo
}

var msg_handler_mgr MsgHandlerMgr

func (this *MsgHandlerMgr) Init() bool {
	this.msgid2handler = make(map[int32]*MsgHanlderInfo)
	return true
}

func (this *MsgHandlerMgr) SetMsgHandler(msg_code uint16, msg_handler CLIENT_MSG_HANDLER) {
	log.Info("set msg [%d] handler !", msg_code)
	this.msgid2handler[int32(msg_code)] = &MsgHanlderInfo{typ: msg_client_message.MessageTypes[msg_code], msg_handler: msg_handler, if_player_msg: false}
}

func (this *MsgHandlerMgr) SetPlayerMsgHandler(msg_code uint16, msg_handler CLIENT_PLAYER_MSG_HANDLER) {
	log.Info("set msg [%d] handler !", msg_code)
	this.msgid2handler[int32(msg_code)] = &MsgHanlderInfo{typ: msg_client_message.MessageTypes[msg_code], player_msg_handler: msg_handler, if_player_msg: true}
}

func (this *MsgHandlerMgr) StartHttp() {
	var err error
	this.reg_http_mux()

	this.msg_http_listener, err = net.Listen("tcp", config.ListenClientInIP)
	if nil != err {
		log.Error("Center StartHttp Failed %s", err.Error())
		return
	}

	signal_mgr.RegCloseFunc("msg_handler_mgr", this.CloseFunc)

	msg_http_server := http.Server{
		Handler:     &MsgHttpHandle{},
		ReadTimeout: 6 * time.Second,
	}

	log.Info("启动消息处理服务 IP:%s", config.ListenClientInIP)
	err = msg_http_server.Serve(this.msg_http_listener)
	if err != nil {
		log.Error("启动消息处理服务失败 %s", err.Error())
		return
	}

}

func (this *MsgHandlerMgr) CloseFunc(info *SignalRegRecod) {
	if nil != this.msg_http_listener {
		this.msg_http_listener.Close()
	}

	info.close_flag = true
	return
}

//=========================================================

func (this *MsgHandlerMgr) reg_http_mux() {
	msg_handler_http_mux = make(map[string]func(http.ResponseWriter, *http.Request))
	msg_handler_http_mux["/client_msg"] = client_msg_handler
}

func client_msg_handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	defer r.Body.Close()

	data, err := ioutil.ReadAll(r.Body)
	if nil != err {
		log.Error("client_msg_handler ReadAll err[%s]", err.Error())
		return
	}
	log.Info("客户端发送过来的二进制流  %v", data)

	tmp_msg := &msg_client_message.C2S_MSG_DATA{}
	err = proto.Unmarshal(data, tmp_msg)
	if nil != err {
		log.Error("client_msg_handler proto Unmarshal err[%s]", err.Error())
		return
	}

	handlerinfo := msg_handler_mgr.msgid2handler[tmp_msg.GetMsgCode()]
	if nil == handlerinfo {
		log.Error("client_msg_handler msg_handler_mgr[%d] nil ", tmp_msg.GetMsgCode())
		return
	}

	req := reflect.New(handlerinfo.typ).Interface().(proto.Message)
	err = proto.Unmarshal(tmp_msg.Data, req)
	if nil != err {
		log.Error("client_msg_handler unmarshal sub msg failed err(%s) !", err.Error())
		return
	}

	log.Info("[接收] [玩家%d:%s] [%s] ", tmp_msg.GetPlayerId(), req.MessageTypeName(), req.String())

	/*
		pid := tmp_msg.GetPlayerId()
		if nil == player_mgr.GetPlayerById(pid) {
			log.Error("client_msg_handler player_mgr failed to getplayerbyid[%d] !", pid)
			return
		}
	*/

	//log.Info("handlerinfo.if_player_msg %d", handlerinfo.if_player_msg)

	var p *Player
	var ret_code int32
	if handlerinfo.if_player_msg {
		pid := tmp_msg.GetPlayerId()

		p = player_mgr.GetPlayerById(pid)
		if nil == p {
			log.Error("client_msg_handler failed to GetPlayerById [%d]", tmp_msg.GetPlayerId())
			return
		}

		tokeninfo := login_token_mgr.GetTockenByAcc(p.Account)
		if nil == tokeninfo || tokeninfo.token != tmp_msg.GetToken() {
			ret_code = int32(msg_client_message.E_ERR_PLAYER_OTHER_PLACE_LOGIN)
		} else {
			p.bhandling = true
			p.b_base_prop_chg = false
			ret_code = handlerinfo.player_msg_handler(w, r, p, req)
		}

	} else {
		ret_code, p = handlerinfo.msg_handler(w, r, req)
	}

	if ret_code <= 0 {
		log.Error("client_msg_handler exec msg_handler ret error_code %d", ret_code)
		res2cli := &msg_client_message.S2C_MSG_DATA{}
		res2cli.ErrorCode = proto.Int32(ret_code)

		final_data, err := proto.Marshal(res2cli)
		if nil != err {
			log.Error("client_msg_handler marshal 1 client msg failed err(%s)", err.Error())
			return
		}

		iret, err := w.Write(final_data)
		if nil != err {
			log.Error("client_msg_handler write data 1 failed err[%s] ret %d", err.Error(), iret)
			return
		}
		//log.Info("write http resp data error %v", final_data)
		return
	} else {
		if nil == p {
			log.Error("client_msg_handler after handle p nil")
			return
		}

		res2cli := &msg_client_message.S2C_MSG_DATA{}
		data := p.PopCurMsgData()
		if nil == data || len(data) < 4 {
			//log.Error("client_msg_handler PopCurMsgDataError nil or len[%d] error", len(data))
			//return
			res2cli.ErrorCode = proto.Int32(ret_code)
		} else {
			//log.Trace("client_msg_handler pop data %v", data)
			res2cli.Data = data
		}

		final_data, err := proto.Marshal(res2cli)
		if nil != err {
			log.Error("client_msg_handler marshal 2 client msg failed err(%s)", err.Error())
			return
		}

		iret, err := w.Write(final_data)
		if nil != err {
			log.Error("client_msg_handler write data 2 failed err[%s] ret %d", err.Error(), iret)
			return
		}

		test_msg := &msg_client_message.S2C_MSG_DATA{}

		err = proto.Unmarshal(final_data, test_msg)
		if nil != err {
			log.Error("test unmarshal failed err[%d]", err.Error())
		}
		//log.Info("write http resp data normal %v len", final_data, len(final_data))

	}

	return
}
