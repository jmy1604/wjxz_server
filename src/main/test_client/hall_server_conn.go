package main

import (
	"bytes"
	"io/ioutil"
	"libs/log"
	"net/http"
	"public_message/gen_go/client_message"
	"reflect"
	//"strings"

	"3p/code.google.com.protobuf/proto"
)

const (
	HALL_CONN_STATE_DISCONNECT  = 0
	HALL_CONN_STATE_CONNECTED   = 1
	HALL_CONN_STATE_FORCE_CLOSE = 2
)

// ========================================================================================

type HallConnection struct {
	state          int32
	last_conn_time int32
	acc            string
	token          string
	hall_ip        string
	playerid       int32

	blogin bool

	last_send_time int64
}

var hall_conn HallConnection

func new_hall_connect(hall_ip, acc, token string) *HallConnection {
	ret_conn := &HallConnection{}
	ret_conn.acc = acc
	ret_conn.hall_ip = hall_ip
	ret_conn.token = token

	log.Info("new hall connection to ip %v", hall_ip)

	return ret_conn
}

func (this *HallConnection) Send(msg proto.Message) {
	data, err := proto.Marshal(msg)
	if nil != err {
		log.Error("login unmarshal failed err[%s]", err.Error())
		return
	}

	C2S_MSG := &msg_client_message.C2S_MSG_DATA{}
	C2S_MSG.PlayerId = proto.Int32(this.playerid)
	C2S_MSG.Token = proto.String(this.token)
	C2S_MSG.MsgCode = proto.Int32(int32(msg.MessageTypeId()))
	C2S_MSG.Data = data

	data, err = proto.Marshal(C2S_MSG)
	if nil != err {
		log.Error("login C2S_MSG Marshal err(%s) !", err.Error())
		return
	}

	resp, err := http.Post(this.hall_ip+"/client_msg", "application/x-www-form-urlencoded", bytes.NewReader(data))
	if nil != err {
		log.Error("login C2S_MSG http post[%s] error[%s]", this.hall_ip+"/client_msg", err.Error())
		return
	}

	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if nil != err {
		log.Error("HallConnection Send read resp body err [%s]", err.Error())
		return
	}

	log.Info("接收到的二进制流 长度[%v] 数据[%v]", len(data), data)
	if len(data) < 0 {
		return
	}

	S2C_MSG := &msg_client_message.S2C_MSG_DATA{}
	err = proto.Unmarshal(data, S2C_MSG)
	if nil != err {
		log.Error("HallConnection unmarshal resp data err(%s) !", err.Error())
		return
	}

	if S2C_MSG.GetErrorCode() < 0 {
		log.Error("服务器返回错误码[%d]", S2C_MSG.GetErrorCode())
		return
	}

	var msg_code uint16
	var cur_len, sub_len int32
	total_data_len := int32(len(S2C_MSG.Data))
	for cur_len < total_data_len {
		msg_code = uint16(S2C_MSG.Data[cur_len])<<8 + uint16(S2C_MSG.Data[cur_len+1])
		sub_len = int32(S2C_MSG.Data[cur_len+2])<<8 + int32(S2C_MSG.Data[cur_len+3])
		sub_data := S2C_MSG.Data[cur_len+4 : cur_len+4+sub_len]
		cur_len = cur_len + 4 + sub_len

		handler_info := msg_handler_mgr.msgid2handler[int32(msg_code)]
		if nil == handler_info {
			log.Warn("HallConnection failed to get msg_handler_info[%d] !", msg_code)
			continue
		}

		new_msg := reflect.New(handler_info.typ).Interface().(proto.Message)
		log.Info("玩家[%d:%s]收到服务器返回%s:[%s]", this.playerid, this.acc, new_msg.MessageTypeName(), new_msg.String())
		err = proto.Unmarshal(sub_data, new_msg)
		if nil != err {
			log.Error("HallConnection failed unmarshal msg data !", msg_code)
			return
		}

		handler_info.msg_handler(this, new_msg)
	}

	return
}

//========================================================================

type CLIENT_MSG_HANDLER func(*HallConnection, proto.Message)

type NEW_MSG_FUNC func() proto.Message

type MsgHanlderInfo struct {
	typ         reflect.Type
	msg_handler CLIENT_MSG_HANDLER
}

type MsgHandlerMgr struct {
	msgid2handler map[int32]*MsgHanlderInfo
}

var msg_handler_mgr MsgHandlerMgr

func (this *MsgHandlerMgr) Init() bool {
	this.msgid2handler = make(map[int32]*MsgHanlderInfo)
	this.RegisterMsgHandler()
	return true
}

func (this *MsgHandlerMgr) SetMsgHandler(msg_code uint16, msg_handler CLIENT_MSG_HANDLER) {
	log.Info("set msg [%d] handler !", msg_code)
	this.msgid2handler[int32(msg_code)] = &MsgHanlderInfo{typ: msg_client_message.MessageTypes[msg_code], msg_handler: msg_handler}
}

func (this *MsgHandlerMgr) RegisterMsgHandler() {
	this.SetMsgHandler(msg_client_message.ID_S2CLoginResponse, S2CLoginResponseHandler)
	this.SetMsgHandler(msg_client_message.ID_S2CRetOptions, S2CRetOptionsHandler)
	this.SetMsgHandler(msg_client_message.ID_S2CRetBaseInfo, S2CRetBaseInfo)
	this.SetMsgHandler(msg_client_message.ID_S2CRetItemInfos, S2CRetItemInfos)
	this.SetMsgHandler(msg_client_message.ID_S2CRetCatInfos, S2CRetCatInfos)
	this.SetMsgHandler(msg_client_message.ID_S2CDrawResult, S2CDrawCatHandler)
	this.SetMsgHandler(msg_client_message.ID_S2CComposeCatResult, S2CComposeCatHandler)
}

func S2CLoginResponseHandler(hall_conn *HallConnection, m proto.Message) {
	res := m.(*msg_client_message.S2CLoginResponse)
	cur_hall_conn := hall_conn_mgr.GetHallConnByAcc(res.GetAcc())
	if nil == cur_hall_conn {
		log.Error("S2CLoginResponseHandler failed to get cur hall[%s]", res.GetAcc())
		return
	}

	hall_conn.playerid = res.GetPlayerId()
	hall_conn.blogin = true
	log.Info("player[%d:%s]收到服务器登录返回 %v", res.GetPlayerId(), res.GetAcc(), res)

	return
}

func S2CRetBaseInfo(hall_conn *HallConnection, m proto.Message) {
	res := m.(*msg_client_message.S2CRetBaseInfo)
	log.Info("收到服务器返回的玩家基本数据 %v", res.String())
}

func S2CRetItemInfos(hall_conn *HallConnection, m proto.Message) {
	res := m.(*msg_client_message.S2CRetItemInfos)
	log.Info("收到服务器返回的物品数据 %v", res.String())
}

func S2CRetCatInfos(hall_conn *HallConnection, m proto.Message) {
	res := m.(*msg_client_message.S2CRetCatInfos)
	log.Info("收到服务器返回的猫数据 %v", res.String())
}

func S2CRetOptionsHandler(hall_conn *HallConnection, m proto.Message) {
	res := m.(*msg_client_message.S2CRetOptions)
	log.Info("player[%d:%s]收到服务器选项返回 %v", hall_conn.playerid, hall_conn.acc, res)

	return
}

func S2CDrawCatHandler(hall_conn *HallConnection, m proto.Message) {
	res := m.(*msg_client_message.S2CDrawResult)
	log.Info("player[%d:%s]收到服务器抽卡返回%s, 猫[%v], 物品[%v]", hall_conn.playerid, hall_conn.acc, res.String(), res.Cats, res.Items)
	return
}

func S2CComposeCatHandler(hall_conn *HallConnection, m proto.Message) {
	res := m.(*msg_client_message.S2CComposeCatResult)
	log.Info("收到服务器返回的合成结果[%v]", res.String())
}
