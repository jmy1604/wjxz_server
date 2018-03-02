package main

import (
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"libs/log"
	"libs/server_conn"
	"net"
	"net/http"
	"public_message/gen_go/server_message"
	"strconv"
	"strings"
	"sync"
	"time"

	"3p/code.google.com.protobuf/proto"
)

var gm_http_mux map[string]func(http.ResponseWriter, *http.Request)

type GmHttpHandle struct{}

func (this *GmHttpHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var act_str, url_str string
	url_str = r.URL.String()
	idx := strings.Index(url_str, "?")
	if -1 == idx {
		act_str = url_str
	} else {
		act_str = string([]byte(url_str)[:idx])
	}
	log.Info("ServeHTTP actstr(%s)", act_str)
	if h, ok := gm_http_mux[act_str]; ok {
		h(w, r)
	}

	return
}

//=======================================================

const (
	INIT_GM_CDM_ARRAY_NUM = 50
	GM_CMD_ARRAY_STEP_NUM = 50

	INIT_SERVER_REWARD_NUM = 50
	SERVER_REWARD_STEP_NUM = 50

	GM_REWARD_ROW_ID = 1
)

type GM_CMD_WAIT_ITEM struct {
	SessionId   int32  // 唯一编号
	Result      string // 结果
	IfGetResult bool   // 是否已经设置过结果
}

type ServerReward struct {
	RewardId    int32                       // 奖励Id
	RewardItems []*msg_server_message.IdNum // 奖励内容
	EndUnix     int32                       // 结束时间
	Channel     string                      // 奖励渠道
}

type GmMgr struct {
	id2payback_lock *sync.RWMutex
	id2payback      map[int32]*msg_server_message.PayBackAdd

	id2paynotice_lock *sync.RWMutex
	id2paynotice      map[int32]*msg_server_message.NoticeAdd

	next_gm_session_id      int32
	next_gm_session_id_lock *sync.RWMutex

	cur_wait_gm_session_num int32
	max_wait_gm_session_num int32
	gmsession2waitarr       []*GM_CMD_WAIT_ITEM
	gmsession2waitarr_lock  *sync.RWMutex

	server_reward_db *dbServerRewardRow

	gm_http_listener net.Listener
}

var gm_mgr GmMgr

func (this *GmMgr) Init() bool {
	this.id2paynotice_lock = &sync.RWMutex{}
	this.id2paynotice = make(map[int32]*msg_server_message.NoticeAdd)

	this.next_gm_session_id_lock = &sync.RWMutex{}
	this.max_wait_gm_session_num = INIT_GM_CDM_ARRAY_NUM
	this.gmsession2waitarr = make([]*GM_CMD_WAIT_ITEM, this.max_wait_gm_session_num)
	this.gmsession2waitarr_lock = &sync.RWMutex{}

	center_cmd_func_map = make(map[string]CENTER_GM_FUNC)

	this.server_reward_db = dbc.ServerRewards.GetRow(GM_REWARD_ROW_ID)
	if nil == this.server_reward_db {
		this.server_reward_db = dbc.ServerRewards.AddRow(GM_REWARD_ROW_ID)
		if nil == this.server_reward_db {
			log.Error("GmMgr Init failed to get server_reward_db !")
			return false
		}
	}

	register_center_cmd_func()

	return true
}

type XmlPayBackItem struct {
	Id             int32  `xml:"ID,attr"`
	Title          string `xml:"Title,attr"`
	Content        string `xml:"Content,attr"`
	ChestPosition1 int32  `xml:"ChestPosition1,attr"`
	ChestPosition2 int32  `xml:"ChestPosition2,attr"`
	ChestPosition3 int32  `xml:"ChestPosition3,attr"`
	TermOfValidity int32  `xml:"TermOfValidity,attr"`
	EndDate        string `xml:"EndDate,attr"`
}

type XmlPayBackList struct {
	Items []XmlPayBackItem `xml:"item"`
}

func (this *GmMgr) LoadPayBack() bool {
	this.id2payback_lock = &sync.RWMutex{}
	this.id2payback = make(map[int32]*msg_server_message.PayBackAdd)

	content, err := ioutil.ReadFile("../game_data/compensate.xml")
	if nil != err {
		log.Error("GmMgr LoadPayBack read file error(%s) !", err.Error())
		return false
	}

	tmp_cfg := &XmlPayBackList{}
	err = xml.Unmarshal(content, tmp_cfg)
	if nil != err {
		log.Error("GmMgr LoadPayBack unmarshl error(%s) content(%s)!", err.Error(), string(content))
		return false
	}

	cur_unix := time.Now().Unix()
	var tmp_pb *msg_server_message.PayBackAdd
	for _, val := range tmp_cfg.Items {
		tmp_t, err := time.Parse("2006 Jan 02 15:04:05", val.EndDate)
		if nil != err {
			log.Error("GmMgr LoadPayBack parse date(%s) error(%s) !", val.EndDate, err.Error())
			return false
		}

		if cur_unix > tmp_t.Unix() {
			log.Error("GmMgr LoadPayBack [%d] time over", val.Id)
			continue
		}

		tmp_pb = &msg_server_message.PayBackAdd{}
		tmp_pb.ObjIds = make([]int32, 0, 3)
		if val.ChestPosition1 > 0 {
			tmp_pb.ObjIds = append(tmp_pb.ObjIds, val.ChestPosition1)
		}
		if val.ChestPosition2 > 0 {
			tmp_pb.ObjIds = append(tmp_pb.ObjIds, val.ChestPosition2)
		}
		if val.ChestPosition3 > 0 {
			tmp_pb.ObjIds = append(tmp_pb.ObjIds, val.ChestPosition3)
		}

		//tmp_pb.Coin = proto.Int32(0)
		//tmp_pb.Diamond = proto.Int32(0)
		tmp_pb.MailContent = proto.String(val.Content)
		tmp_pb.MailTitle = proto.String(val.Title)
		tmp_pb.OverUnix = proto.Int32(int32(tmp_t.Unix()))
		tmp_pb.PayBackId = proto.Int32(val.Id)
		this.id2payback[val.Id] = tmp_pb
	}

	return true
}

func (this *GmMgr) AddPayBack(tmp_pb *msg_server_message.PayBackAdd) {
	if nil == tmp_pb {
		log.Error("GmMgr AddPayBack tmp_pb nil")
		return
	}

	pb_id := tmp_pb.GetPayBackId()

	this.id2payback_lock.Lock()
	defer this.id2payback_lock.Unlock()

	cur_pb := this.id2payback[pb_id]
	if nil == cur_pb {
		log.Error("GmMgr AddPayBack cur_pb[%d] already have !", pb_id)
		return
	}

	this.id2payback[pb_id] = tmp_pb
	return
}

func (this *GmMgr) RemovePayBack(pb_id int32) {
	this.id2payback_lock.Lock()
	defer this.id2payback_lock.Unlock()

	if nil != this.id2payback[pb_id] {
		delete(this.id2payback, pb_id)
	}

	return
}

func (this *GmMgr) FillPayBackSyncMsg() *msg_server_message.SyncPayBackDataList {
	this.id2payback_lock.RLock()
	defer this.id2payback_lock.RUnlock()
	tmp_len := int32(len(this.id2payback))
	if tmp_len < 1 {
		return nil
	}

	ret_msg := &msg_server_message.SyncPayBackDataList{}
	ret_msg.PayBackList = make([]*msg_server_message.PayBackAdd, 0, tmp_len)
	for _, pb := range this.id2payback {
		if nil == pb {
			continue
		}

		ret_msg.PayBackList = append(ret_msg.PayBackList, pb)
	}

	if len(ret_msg.PayBackList) > 0 {
		return ret_msg
	}

	return nil
}

func (this *GmMgr) OnHallRegist(hall_svr *HallAgent) {
	if nil == hall_svr {
		log.Error("GmMgr OnHallRegist  hall_svr nil !")
		return
	}

	res2h := this.FillPayBackSyncMsg()
	if nil == res2h {
		log.Info("GmMgr OnHallRegist[%d] payback msg nil", hall_svr.id)
		return
	}

	hall_svr.Send(res2h)

	return
}

func (this *GmMgr) StartHttp() {
	var err error
	this.reg_http_mux()

	this.gm_http_listener, err = net.Listen("tcp", config.GmIP)
	if nil != err {
		log.Error("Center StartHttp Failed %s", err.Error())
		return
	}

	signal_mgr.RegCloseFunc("gm_mgr", this.CloseFunc)

	gm_http_server := http.Server{
		Handler:     &GmHttpHandle{},
		ReadTimeout: 6 * time.Second,
	}

	log.Info("启动Gm服务 IP:%s", config.GmIP)
	err = gm_http_server.Serve(this.gm_http_listener)
	if err != nil {
		log.Error("启动Center gm Http Server %s", err.Error())
		return
	}

}

func (this *GmMgr) CloseFunc(info *SignalRegRecod) {
	if nil != this.gm_http_listener {
		this.gm_http_listener.Close()
	}

	info.close_flag = true
	return
}

//=========================================================

func (this *GmMgr) reg_http_mux() {
	gm_http_mux = make(map[string]func(http.ResponseWriter, *http.Request))
	gm_http_mux["/gm_cmd"] = test_gm_command_http_handler
}

func test_gm_command_http_handler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	if "GET" == r.Method {
		params := make([]string, 0, 5)
		gm_str := r.URL.Query().Get("cmd_str")
		params = append(params, gm_str)
		param1 := r.URL.Query().Get("param1")
		params = append(params, param1)
		param2 := r.URL.Query().Get("param2")
		params = append(params, param2)
		param3 := r.URL.Query().Get("param3")
		params = append(params, param3)
		param4 := r.URL.Query().Get("param4")
		params = append(params, param4)

		log.Info("gm GET cmd_str[%s] param1[%s] param2[%s] param3[%s] param4[%s]", gm_str, param1, param2, param3, param4)

		gm_func := center_cmd_func_map[params[0]]
		var result_str string
		if nil != gm_func {
			result_str = gm_func(params)
			w.Write([]byte(result_str))
		} else {

			new_w_item := &GM_CMD_WAIT_ITEM{}
			new_w_item.SessionId = gm_mgr.get_next_gm_session_id()
			gm_mgr.add_gm_wait_session(new_w_item)

			switch gm_str {
			case "item_query":
				{
					c2h_req := &msg_server_message.C2HItemQuery{}
					ival, err := strconv.Atoi(param1)
					if nil != err {
						log.Error("center gm item_qurey failed to convert player id %s error %s", param1, err.Error())
					} else {
						c2h_req = &msg_server_message.C2HItemQuery{}
						c2h_req.PlayerId = proto.Int32(int32(ival))
						c2h_req.SessionId = proto.Int32(new_w_item.SessionId)
						hall_agent_mgr.Broadcast(c2h_req)
					}

				}
			default:
				{
					tmp_gm := &msg_server_message.C2HGmCommand{}
					tmp_gm.Command = proto.String(gm_str + " " + param1 + " " + param2 + " " + param3 + " " + param4)
					tmp_gm.SessionId = proto.Int32(new_w_item.SessionId)
					hall_agent_mgr.Broadcast(tmp_gm)
				}
			}

			for idx := int32(0); idx < 5; idx++ {
				time.Sleep(1 * time.Second)
				log.Info("session %d wait %d", new_w_item.SessionId, idx)
				if gm_mgr.ChkIfGetResult(new_w_item.SessionId) {
					break
				}
			}

			log.Info("session[%d] before pop", new_w_item.SessionId)
			new_w_item = gm_mgr.pop_gm_wait_session(new_w_item.SessionId)
			log.Info("session[%d] after pop", new_w_item.SessionId)
			if nil != new_w_item && gm_mgr.ChkIfGetResult(new_w_item.SessionId) {
				result_str = new_w_item.Result
			} else {
				result_str = "{result:\"failed to get result !\"}"
			}

			log.Info("session[%d] write result [%s]", new_w_item.SessionId, result_str)
			w.Write([]byte(result_str))
		}

	} else {
		log.Error("test_gm_command_http_handler not support POST Method")
	}
}

// =============================================================

// 中心服务器Gm回复消息监听
func reg_hall_gm_response_msg() {
	hall_agent_mgr.SetMessageHandler(msg_server_message.ID_H2CGmResult, H2CGmResultHanlder)
	hall_agent_mgr.SetMessageHandler(msg_server_message.ID_H2CItemQuery, H2CItemQueryHandler)
	hall_agent_mgr.SetMessageHandler(msg_server_message.ID_H2CGetServerReward, H2CGetServerRewardHandler)
}

func H2CGmResultHanlder(conn *server_conn.ServerConn, msg proto.Message) {
	res := msg.(*msg_server_message.H2CGmResult)
	if nil == conn || nil == res {
		log.Error("H2CGmResultHanlder conn or res nil !")
		return
	}

	log.Info("H2CGmResultHanlder %s", res.String())

	gm_mgr.UpdateResultString(res.GetSessionId(), res.GetResult())

	return
}

func H2CItemQueryHandler(conn *server_conn.ServerConn, m proto.Message) {

	res := m.(*msg_server_message.H2CItemQuery)
	if nil == conn || nil == res {
		log.Error("H2CItemQueryHandler conn or res nil !")
		return
	}

	session_id := res.GetSessionId()
	if !gm_mgr.ChkIfSessionExist(session_id) {
		log.Trace("H2CItemQueryHandler cur_session_id[%s] nil ", session_id)
		return
	}

	var result_str string
	items := res.GetItems()
	tmp_len := len(items)
	if tmp_len < 1 {
		log.Trace("H2CItemQueryHandler item none !")
		gm_mgr.UpdateResultString(session_id, "item_empty !")
		return
	}

	result_str += "{"
	bfirst := true
	for idx := 0; idx < tmp_len; idx++ {

		if bfirst {
			bfirst = false
		} else {
			result_str += ","
		}

		result_str += fmt.Sprintf("{id:%d,num:%d}", items[idx].GetId(), items[idx].GetNum())
	}
	result_str += "}"

	gm_mgr.UpdateResultString(session_id, result_str)

	return
}

func H2CGetServerRewardHandler(conn *server_conn.ServerConn, msg proto.Message) {
	res := gm_mgr.server_reward_db.RewardInfos.ChkFillAllRewardMsg()
	if nil != res {
		conn.Send(res, true)
	}

	return
}

// ---------------------------------------------------------------------------

// 中心服务器的GM会话管理

func (this *GmMgr) get_next_gm_session_id() int32 {
	this.next_gm_session_id_lock.Lock()
	defer this.next_gm_session_id_lock.Unlock()

	this.next_gm_session_id++

	return this.next_gm_session_id
}

// 检查session是否存在
func (this *GmMgr) ChkIfSessionExist(session_id int32) bool {
	this.gmsession2waitarr_lock.RLock()
	defer this.gmsession2waitarr_lock.RUnlock()

	for idx := int32(0); idx < this.cur_wait_gm_session_num; idx++ {
		if this.gmsession2waitarr[idx].SessionId == session_id {
			return true
		}
	}
	return false
}

// 添加Gm命令到等待列表
func (this *GmMgr) add_gm_wait_session(item *GM_CMD_WAIT_ITEM) {
	if nil == item {
		log.Error("GmMgr add_gm_wait_session item nil !")
		return
	}

	this.gmsession2waitarr_lock.Lock()
	defer this.gmsession2waitarr_lock.Unlock()

	if this.cur_wait_gm_session_num >= this.max_wait_gm_session_num {
		new_max := this.max_wait_gm_session_num + GM_CMD_ARRAY_STEP_NUM
		new_arr := make([]*GM_CMD_WAIT_ITEM, new_max)
		for idx := int32(0); idx < this.max_wait_gm_session_num; idx++ {
			new_arr[idx] = this.gmsession2waitarr[idx]
		}

		this.max_wait_gm_session_num = new_max
		this.gmsession2waitarr = new_arr
	}

	this.gmsession2waitarr[this.cur_wait_gm_session_num] = item
	this.cur_wait_gm_session_num++

	return
}

func (this *GmMgr) remove_gm_wait_session(session_id int32) {
	this.gmsession2waitarr_lock.Lock()
	defer this.gmsession2waitarr_lock.Unlock()

	var tmp_item *GM_CMD_WAIT_ITEM
	for idx := int32(0); idx < this.cur_wait_gm_session_num; idx++ {
		tmp_item = this.gmsession2waitarr[idx]
		if nil == tmp_item {
			log.Error("GmMgr remove_gm_wait_session [%d] nil cur[%d] max[%d]", idx, this.cur_wait_gm_session_num, this.max_wait_gm_session_num)
			continue
		}

		if tmp_item.SessionId == session_id {
			if idx != this.cur_wait_gm_session_num-1 {
				this.gmsession2waitarr[idx] = this.gmsession2waitarr[this.cur_wait_gm_session_num-1]
				this.cur_wait_gm_session_num--
			}
			break
		}
	}
}

func (this *GmMgr) pop_gm_wait_session(session_id int32) *GM_CMD_WAIT_ITEM {
	this.gmsession2waitarr_lock.Lock()
	defer this.gmsession2waitarr_lock.Unlock()

	var tmp_item *GM_CMD_WAIT_ITEM
	for idx := int32(0); idx < this.cur_wait_gm_session_num; idx++ {
		tmp_item = this.gmsession2waitarr[idx]
		if nil == tmp_item {
			log.Error("GmMgr remove_gm_wait_session [%d] nil cur[%d] max[%d]", idx, this.cur_wait_gm_session_num, this.max_wait_gm_session_num)
			continue
		}

		if tmp_item.SessionId == session_id {
			if idx != this.cur_wait_gm_session_num-1 {
				this.gmsession2waitarr[idx] = this.gmsession2waitarr[this.cur_wait_gm_session_num-1]
				this.cur_wait_gm_session_num--
			}
			return tmp_item
		}
	}

	return nil
}

func (this *GmMgr) ChkIfGetResult(session_id int32) bool {
	this.gmsession2waitarr_lock.RLock()
	defer this.gmsession2waitarr_lock.RUnlock()

	var tmp_item *GM_CMD_WAIT_ITEM
	for idx := int32(0); idx < this.cur_wait_gm_session_num; idx++ {
		tmp_item = this.gmsession2waitarr[idx]
		if nil == tmp_item {
			continue
		}

		if tmp_item.SessionId == session_id && tmp_item.IfGetResult {
			return true
		}
	}

	return false
}

func (this *GmMgr) UpdateResultString(session_id int32, result_str string) {
	this.gmsession2waitarr_lock.Lock()
	defer this.gmsession2waitarr_lock.Unlock()

	var tmp_item *GM_CMD_WAIT_ITEM
	for idx := int32(0); idx < this.cur_wait_gm_session_num; idx++ {
		tmp_item = this.gmsession2waitarr[idx]
		if nil == tmp_item {
			continue
		}

		if tmp_item.SessionId == session_id {
			tmp_item.IfGetResult = true
			tmp_item.Result = result_str
			break
		}
	}
}

// ======================================================

// 中心服务器就可以处理的Gm命令

type CENTER_GM_FUNC func(params []string) string

var center_cmd_func_map map[string]CENTER_GM_FUNC

func register_center_cmd_func() {
	//center_cmd_func_map["forbid_account_talk"] = forbid_account_talk
	center_cmd_func_map["forbid_account_login"] = forbid_account_login
	center_cmd_func_map["give_all_item"] = give_all_item
	center_cmd_func_map["give_channel_item"] = give_channel_item
	center_cmd_func_map["player_account"] = player_account
}

func forbid_account_talk(params []string) string {
	if len(params) < 4 {
		return "forbid_account_talk failed params less than 1 !"
	}

	pid, err := strconv.Atoi(params[1])
	if nil != err {
		return "forbid_account_talk failed to convert pid " + params[1] + " err:" + err.Error()
	}

	forbid_reason := params[2]

	forbid_sec, err := strconv.Atoi(params[3])
	if nil != err {
		return "forbid_account_talk failed to convert sec " + params[1] + " err:" + err.Error()
	}

	if forbid_sec < 0 {
		return "forbid_account_talk forbidsec < 0 " + params[3]
	}

	cur_db := dbc.ForbidTalks.GetRow(int32(pid))
	if nil != cur_db {
		new_db := dbc.ForbidTalks.AddRow(int32(pid))
		new_db.m_PlayerId = int32(pid)
		new_db.m_Reason = forbid_reason
		new_db.m_EndUnix = int32(time.Now().Unix()) + int32(forbid_sec)
	} else {
		cur_db.SetEndUnix(int32(time.Now().Unix()) + int32(forbid_sec))
	}

	return "ok"
}

func forbid_account_login(params []string) string {
	if len(params) < 4 {
		return "forbid_account_login failed params less than 1 !"
	}

	pid, err := strconv.Atoi(params[1])
	if nil != err {
		return "forbid_account_login failed to convert pid " + params[1] + " err:" + err.Error()
	}

	forbid_reason := params[2]

	forbid_sec, err := strconv.Atoi(params[3])
	if nil != err {
		return "forbid_account_login failed to convert sec " + params[1] + " err:" + err.Error()
	}

	if forbid_sec < 0 {
		return "forbid_account_login forbidsec < 0 " + params[3]
	}

	cur_db := dbc.ForbidLogins.GetRow(int32(pid))
	if nil != cur_db {
		new_db := dbc.ForbidLogins.AddRow(int32(pid))
		new_db.m_PlayerId = int32(pid)
		new_db.m_Reason = forbid_reason
		new_db.m_EndUnix = int32(time.Now().Unix()) + int32(forbid_sec)

	}

	return "ok"
}

func give_all_item(params []string) string {
	if len(params) < 4 {
		return "give_all_item failed params less than 4 !"
	}

	content_id := params[1]

	items := parse_id_nums_string(params[2])
	if len(items) < 1 {
		return "give_all_item items empty " + params[2]
	}

	req2h := &msg_server_message.C2HAddServerReward{}
	req2h.Channel = proto.String("")
	req2h.Content = proto.String(content_id)
	tmp_len := len(items)
	req2h.Items = make([]*msg_server_message.IdNum, 0, tmp_len)
	var tmp_val *ItemInfo
	var tmp_idnum *msg_server_message.IdNum
	for idx := 0; idx < tmp_len; idx++ {
		tmp_val = items[idx]
		tmp_idnum = &msg_server_message.IdNum{}
		tmp_idnum.Id = proto.Int32(tmp_val.Id)
		tmp_idnum.Num = proto.Int32(tmp_val.Num)
		req2h.Items = append(req2h.Items, tmp_idnum)
	}

	last_sec, err := strconv.Atoi(params[3])
	if nil != err {
		log.Error("give_all_item failed to convert last_sec [%s] err[%s] !", params[3], err.Error())
		return fmt.Sprintf("give_all_item failed to convert last_sec [%s] err[%s] !", params[3], err.Error())
	}

	reward_id := gm_mgr.add_gm_server_reward(req2h.Items, content_id, "", int32(last_sec))
	if reward_id > 0 {
		req2h.RewardId = proto.Int32(reward_id)
		req2h.EndUnix = proto.Int32(int32(time.Now().Unix()) + int32(last_sec))
		hall_agent_mgr.Broadcast(req2h)
	}

	return "ok"
}

func give_channel_item(params []string) string {
	if len(params) < 4 {
		return "give_channel_item failed params less than 1 !"
	}

	content_id := params[1]

	channel := params[2]

	items := parse_id_nums_string(params[3])
	tmp_len := len(items)
	if tmp_len < 1 {
		return "give_channel_item items empty " + params[3]
	}

	req2h := &msg_server_message.C2HAddServerReward{}
	req2h.Channel = proto.String(channel)
	req2h.Content = proto.String(content_id)

	req2h.Items = make([]*msg_server_message.IdNum, 0, tmp_len)
	var tmp_val *ItemInfo
	var tmp_idnum *msg_server_message.IdNum
	for idx := 0; idx < tmp_len; idx++ {
		tmp_val = items[idx]
		tmp_idnum = &msg_server_message.IdNum{}
		tmp_idnum.Id = proto.Int32(tmp_val.Id)
		tmp_idnum.Num = proto.Int32(tmp_val.Num)
		req2h.Items = append(req2h.Items, tmp_idnum)
	}

	last_sec, err := strconv.Atoi(params[4])
	if nil != err {
		log.Error("give_channel_item failed to convert last_sec [%s] err[%s] !", params[4], err.Error())
		return fmt.Sprintf("give_channel_item failed to convert last_sec [%s] err[%s] !", params[4], err.Error())
	}

	reward_id := gm_mgr.add_gm_server_reward(req2h.Items, content_id, channel, int32(last_sec))
	if reward_id > 0 {
		req2h.RewardId = proto.Int32(reward_id)
		req2h.EndUnix = proto.Int32(int32(time.Now().Unix()) + int32(last_sec))
		hall_agent_mgr.Broadcast(req2h)
	}

	return "ok"
}

func player_account(params []string) string {
	if len(params) < 2 {
		return "player_account failed params less than 2 !"
	}

	ival, err := strconv.Atoi(params[1])

	if nil != err {
		return fmt.Sprintf("player_account failed to covert ")
	}

	return fmt.Sprintf("account: %s", dbc_account.AccountsMgr.GetAccByPid(int32(ival)))
}
