package main

import (
	"libs/log"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	_ "public_message/gen_go/server_message"
	"strings"
	"sync"

	_ "3p/code.google.com.protobuf/proto"
	"github.com/golang/protobuf/proto"
)

const (
	PLAYER_DATA_UP_FRAME_COUNT = 30
	PLAYER_CAMP_1              = 1 // 玩家阵营1
	PLAYER_CAMP_2              = 2 // 玩家阵营2
	DEFAULT_PLAYER_ARRAY_MAX   = 1
	PLAYER_ARRAY_MAX_ADD_STEP  = 1
)

type PlayerManager struct {
	id2players      map[int32]*Player
	id2players_lock *sync.RWMutex

	acc2players      map[string]*Player
	acc2Players_lock *sync.RWMutex

	all_player_array []*Player
	cur_all_count    int32
	cur_all_max      int32

	ol_player_array []*Player
	cur_ol_count    int32
	cur_ol_max      int32
}

var player_mgr PlayerManager

func (this *PlayerManager) Init() bool {
	this.id2players = make(map[int32]*Player)
	this.id2players_lock = &sync.RWMutex{}
	this.acc2players = make(map[string]*Player)
	this.acc2Players_lock = &sync.RWMutex{}

	this.ol_player_array = make([]*Player, DEFAULT_PLAYER_ARRAY_MAX)
	this.cur_ol_count = 0
	this.cur_ol_max = DEFAULT_PLAYER_ARRAY_MAX

	this.all_player_array = make([]*Player, DEFAULT_PLAYER_ARRAY_MAX)
	this.cur_all_count = 0
	this.cur_all_max = DEFAULT_PLAYER_ARRAY_MAX

	return true
}

func (this *PlayerManager) GetPlayerById(id int32) *Player {
	this.id2players_lock.Lock()
	defer this.id2players_lock.Unlock()

	return this.id2players[id]
}

func (this *PlayerManager) GetAllPlayers() []*Player {
	this.id2players_lock.RLock()
	defer this.id2players_lock.RUnlock()

	ret_ps := make([]*Player, 0, len(this.id2players))
	for _, p := range this.id2players {
		ret_ps = append(ret_ps, p)
	}

	return ret_ps
}

func (this *PlayerManager) Add2IdMap(p *Player) {
	if nil == p {
		log.Error("Player_agent_mgr Add2IdMap p nil !")
		return
	}
	this.id2players_lock.Lock()
	defer this.id2players_lock.Unlock()

	if nil != this.id2players[p.Id] {
		log.Error("PlayerManager Add2IdMap already have player(%d)", p.Id)
	}

	this.id2players[p.Id] = p

	if this.cur_all_count >= this.cur_all_max {
		this.cur_all_max = this.cur_all_max + PLAYER_ARRAY_MAX_ADD_STEP
		new_all_array := make([]*Player, this.cur_all_max)
		for idx := int32(0); idx < this.cur_all_count; idx++ {
			new_all_array[idx] = this.all_player_array[idx]
		}

		this.all_player_array = new_all_array
	}

	this.all_player_array[this.cur_all_count] = p
	p.all_array_idx = this.cur_all_count
	this.cur_all_count++

	return
}

func (this *PlayerManager) RemoveFromIdMap(id int32) {
	this.id2players_lock.Lock()
	defer this.id2players_lock.Unlock()

	cur_p := this.id2players[id]
	if nil != cur_p {
		delete(this.id2players, id)
	}

	if -1 != cur_p.all_array_idx {
		if cur_p.all_array_idx != this.cur_all_count-1 {
			this.all_player_array[cur_p.all_array_idx] = this.all_player_array[this.cur_all_count-1]
			this.all_player_array[cur_p.all_array_idx].all_array_idx = cur_p.all_array_idx
		}
		this.cur_all_count--
	}

	return
}

func (this *PlayerManager) GetAllPlayerNum() int32 {
	return this.cur_all_count
}

func (this *PlayerManager) Add2AccMap(p *Player) {
	if nil == p {
		log.Error("PlayerManager Add2AccMap p nil !")
		return
	}

	this.acc2Players_lock.RLock()
	defer this.acc2Players_lock.RUnlock()
	if nil != this.acc2players[p.Account] {
		log.Info("PlayerManager Add2AccMap old_p not nil")
		return
	}

	this.acc2players[p.Account] = p

	if this.cur_ol_count >= this.cur_ol_max {
		tmp_player_array := make([]*Player, this.cur_ol_max+PLAYER_ARRAY_MAX_ADD_STEP)
		for idx := int32(0); idx < this.cur_ol_max; idx++ {
			tmp_player_array[idx] = this.ol_player_array[idx]
		}

		this.cur_ol_max = this.cur_ol_count + PLAYER_ARRAY_MAX_ADD_STEP
		this.ol_player_array = tmp_player_array
	}

	this.ol_player_array[this.cur_ol_count] = p
	p.ol_array_idx = this.cur_ol_count
	this.cur_ol_count++

	return
}

func (this *PlayerManager) RemoveFromAccMap(acc string) {
	if "" == acc {
		log.Error("PlayerManager RemoveFromAccMap acc empty !")
		return
	}

	this.acc2Players_lock.Lock()
	defer this.acc2Players_lock.Unlock()
	cur_p := this.acc2players[acc]
	if nil != cur_p {
		if cur_p.ol_array_idx != -1 {
			if cur_p.ol_array_idx != this.cur_ol_count-1 {
				if nil != this.ol_player_array[this.cur_ol_count-1] {
					this.ol_player_array[this.cur_ol_count-1].ol_array_idx = cur_p.ol_array_idx
				}
				this.ol_player_array[cur_p.ol_array_idx] = this.ol_player_array[this.cur_ol_count-1]
			}
			this.cur_ol_count = this.cur_ol_count - 1
		}
		delete(this.acc2players, acc)
	}

	return
}

func (this *PlayerManager) GetCurOnlineNum() int32 {
	return this.cur_ol_count
}

func (this *PlayerManager) GetPlayerByAcc(acc string) *Player {
	if "" == acc {
		return nil
	}

	this.acc2Players_lock.Lock()
	defer this.acc2Players_lock.Unlock()

	return this.acc2players[acc]
}

func (this *PlayerManager) PlayerLogout(p *Player) {
	if nil == p {
		log.Error("PlayerManager PlayerLogout p nil !")
		return
	}

	this.RemoveFromAccMap(p.Account)

	p.OnLogout()
}

func (this *PlayerManager) OnTick() {

}

func (this *PlayerManager) SendMsgToAllPlayers(msg proto.Message) {
	if nil == msg {
		log.Error("PlayerManager SendMsgToAllPlayers msg nil !")
		return
	}
}

//==============================================================================
func (this *PlayerManager) RegMsgHandler() {
	msg_handler_mgr.SetMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ENTER_GAME_REQUEST), C2SEnterGameRequestHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_TEST_COMMAND), C2STestCommandHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_BATTLE_RESULT_REQUEST), C2SFightHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_SET_TEAM_REQUEST), C2SSetTeamHandler)
}

func base_msgid2msg(msg_id uint16) proto.Message {
	if msg_id == uint16(msg_client_message_id.MSGID_C2S_ENTER_GAME_REQUEST) {
		return &msg_client_message.C2SEnterGameRequest{}
	} else if msg_id == uint16(msg_client_message_id.MSGID_C2S_TEST_COMMAND) {
		return &msg_client_message.C2S_TEST_COMMAND{}
	} else if msg_id == uint16(msg_client_message_id.MSGID_C2S_BATTLE_RESULT_REQUEST) {
		return &msg_client_message.C2SBattleResultRequest{}
	} else if msg_id == uint16(msg_client_message_id.MSGID_C2S_SET_TEAM_REQUEST) {
		return &msg_client_message.C2SSetTeamRequest{}
	} else {
		log.Warn("Cant get base proto message by msg_id[%v]", msg_id)
	}
	return nil
}

func C2SEnterGameRequestHandler(w http.ResponseWriter, r *http.Request, msg proto.Message) (int32, *Player) {
	var p *Player
	req := msg.(*msg_client_message.C2SEnterGameRequest)

	acc := req.GetAcc()
	if "" == acc {
		log.Error("PlayerEnterGameHandler acc empty !")
		return -1, p
	}

	token_info := login_token_mgr.GetTockenByAcc(acc)
	if nil == token_info {
		log.Error("PlayerEnterGameHandler account[%v] no token info!", acc)
		return -2, p
	}

	if req.GetToken() != token_info.token {
		log.Error("PlayerEnterGameHandler token check failed !(%s) != (%s)", req.GetToken(), token_info.token)
		return -3, p
	}

	//p = player_mgr.GetPlayerById(player_id)
	is_new := false
	p = player_mgr.GetPlayerByAcc(acc)
	if nil == p {
		//pdb := dbc.Players.GetRow(p.Id)
		//if nil == pdb {
		global_row := dbc.Global.GetRow()
		player_id := global_row.GetNextPlayerId()
		pdb := dbc.Players.AddRow(player_id)
		if nil == pdb {
			log.Error("player_db_to_msg AddRow pid(%d) failed !", player_id)
			return -4, p
		}

		pdb.SetAccount(token_info.acc)
		p = new_player(player_id, token_info.acc, token_info.token, pdb)

		p.OnCreate()

		log.Info("player_db_to_msg new player(%d) !", player_id)
		//} else {
		//	p = new_player(p.Id, token_info.acc, token_info.token, pdb)
		//}
		player_mgr.Add2AccMap(p)
		player_mgr.Add2IdMap(p)
		is_new = true
	} else {
		p.Account = token_info.acc
		p.Token = token_info.token
	}

	ip_port := strings.Split(r.RemoteAddr, ":")
	if len(ip_port) >= 2 {
		p.pos = position_table.GetPosByIP(ip_port[0])
	}

	p.bhandling = true

	p.send_enter_game(acc, p.Id)
	p.OnLogin()
	if !is_new {
		p.send_roles()
	}
	p.notify_enter_complete()

	log.Info("PlayerEnterGameHandler account[%s] token[%s]", req.GetAcc(), req.GetToken())

	return 1, p
}
