package main

import (
	_ "fmt"
	_ "libs/log"
	"public_message/gen_go/server_message"
	_ "strconv"
	_ "strings"
	"sync"
	_ "time"

	_ "3p/code.google.com.protobuf/proto"
	_ "github.com/golang/protobuf/proto"
)

const (
	GM_SERVER_REWARD_CHK_SEC = 300
)

type HALL_GM_FUNC func(params []string) *msg_server_message.H2CGmResult

type ServerReward struct {
	RewardId    int32                       // 奖励Id
	RewardItems []*msg_server_message.IdNum // 奖励内容
	EndUnix     int32                       // 结束时间
	Channel     string
	ContentId   string
}

type GmCommandMgr struct {
	gm_func_map map[string]HALL_GM_FUNC

	last_chk_server_reward_sec int32

	id2serverreward_lock *sync.RWMutex
	id2serverreward      map[int32]*ServerReward
}

var gm_command_mgr GmCommandMgr

func (this *GmCommandMgr) Init() bool {
	this.id2serverreward_lock = &sync.RWMutex{}
	this.id2serverreward = make(map[int32]*ServerReward)

	this.RegGmFunc()
	this.RegGmMsg()
	return true
}

func (this *GmCommandMgr) AfterCenterMatchConn() {
	// 向中心服务器请求
	//center_conn.Send(&msg_server_message.H2CGetServerReward{})
}

// ==============================================

func (this *GmCommandMgr) RegGmFunc() {
	this.gm_func_map = make(map[string]HALL_GM_FUNC)
	//this.gm_func_map["give_one_item"] = this.give_one_item // 给予多个人物品
	//this.gm_func_map["forbid_talk"] = this.forbid_talk
	this.gm_func_map["broadcast"] = this.gm_broadcast
}

/*func (this *GmCommandMgr) give_one_item(params []string) *msg_server_message.H2CGmResult {
	if len(params) < 1 {
		return nil
	}

	pids_str := params[1]
	pids_arry := strings.Split(pids_str, " ")

	tmp_len := int32(len(pids_arry))
	if tmp_len < 1 {
		log.Error("GmCommandMgr give_one_item pids empty !")
		return nil
	}

	mail_content := params[2]

	items := parse_id_nums_string(params[3])

	var pid int32
	var p *Player
	for idx := int32(0); idx < tmp_len; idx++ {
		ival, err := strconv.Atoi(pids_arry[idx])
		if nil != err {
			log.Error("GmCommandMgr give_one_item failed to convert pid[%s]", pids_arry[idx])
			continue
		}

		pid = int32(ival)
		p = player_mgr.GetPlayerById(pid)
		if nil == p {
			continue
		}

		//p.SendGmItemMail(mail_content, items, 24*3600*7, true)
	}

	ret_msg := &msg_server_message.H2CGmResult{}
	ret_msg.Result = "{result:\"Succeed !\"}"

	return ret_msg
}*/

func (this *GmCommandMgr) gm_broadcast(params []string) *msg_server_message.H2CGmResult {

	// 调用全服公告接口
	anouncement_mgr.PushNew(ANOUNCEMENT_TYPE_TEXT, true, 0, 0, 0, 0, params[1])

	ret_msg := &msg_server_message.H2CGmResult{}
	ret_msg.Result = "{result:\"Succeed !\"}"

	return ret_msg
}

// =============================================

func (this *GmCommandMgr) RegGmMsg() {
	//center_conn.SetMessageHandler(uint16(msg_server_message.MSGID_), C2HGmCommandHandler)
}

/*func C2HGmCommandHandler(c *CenterConnection, msg proto.Message) {
	req := msg.(*msg_server_message.C2HGmCommand)
	if nil == c || nil == req {
		log.Error("C2HGmCommandHandler c or req nil !")
		return
	}

	params_str := req.GetCommand()
	params := strings.Split(params_str, " ")
	if len(params) < 1 {
		log.Error("C2HGmCommandHandler params nil [%s]", params_str)
		return
	}

	gm_func := gm_command_mgr.gm_func_map[params[0]]
	if nil == gm_func {
		log.Error("C2HGmCommandHandler failed to find gm_func[%s] !", params[0])
		return
	}

	ret_msg := gm_func(params)
	if nil != ret_msg {
		ret_msg.SessionId = proto.Int32(req.GetSessionId())
		c.Send(ret_msg)
	}

	return
}

func C2HItemQueryHandler(c *CenterConnection, msg proto.Message) {
	req := msg.(*msg_server_message.C2HItemQuery)
	if nil == c || nil == req {
		log.Error("C2HItemQueryHandler c or req nil !")
		return
	}

	log.Info("C2HItemQueryHandler %s", req.String())

	pid := req.GetPlayerId()
	p := player_mgr.GetPlayerById(pid)
	if nil == p {
		log.Error("C2HItemQueryHandler failed to find player[%d]", pid)
		ret_msg := &msg_server_message.H2CGmResult{}

		ret_msg.SessionId = proto.Int32(req.GetSessionId())
		ret_msg.Result = proto.String(fmt.Sprintf("failed to find player[%d]", pid))
		c.Send(ret_msg)
		return
	}

	ret_msg := &msg_server_message.H2CItemQuery{}
	p.db.Items.FillAllGmQueryMsg(ret_msg)

	ret_msg.SessionId = proto.Int32(req.GetSessionId())
	c.Send(ret_msg)

	return
}*/
