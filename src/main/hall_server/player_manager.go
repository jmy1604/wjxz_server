package main

import (
	"libs/log"
	"libs/socket"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/server_message"
	"strings"
	"sync"

	"3p/code.google.com.protobuf/proto"
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
	msg_handler_mgr.SetMsgHandler(msg_client_message.ID_C2SLoginRequest, C2SLoginRequestHandler)
}

func C2SLoginRequestHandler(w http.ResponseWriter, r *http.Request, msg proto.Message) (int32, *Player) {
	var p *Player
	req := msg.(*msg_client_message.C2SLoginRequest)

	acc := req.GetAcc()
	if "" == acc {
		log.Error("PlayerLoginRequestHandler acc empty !")
		return -1, p
	}

	token_info := login_token_mgr.GetTockenByAcc(acc)
	if nil == token_info {
		log.Error("PlayerLoginRequestHandler account[%v] no token info!", acc)
		return -2, p
	}

	if req.GetToken() != token_info.token {
		log.Error("PlayerLoginRequestHandler token check failed !(%s) != (%s)", req.GetToken(), token_info.token)
		return -3, p
	}

	playerid := token_info.playerid
	pdb := dbc.Players.GetRow(playerid)
	p = player_mgr.GetPlayerById(playerid)
	if nil == p {
		if nil == pdb {
			pdb = dbc.Players.AddRow(playerid)
			if nil == pdb {
				log.Error("player_db_to_msg AddRow pid(%d) failed !", playerid)
				return -4, p
			}

			pdb.SetAccount(token_info.acc)
			p = new_player(playerid, token_info.acc, token_info.token, pdb)

			p.OnCreate()

			log.Info("player_db_to_msg new player(%d) !", playerid)
		} else {
			p = new_player(playerid, token_info.acc, token_info.token, pdb)
		}
		player_mgr.Add2AccMap(p)
		player_mgr.Add2IdMap(p)
	} else {
		p.Account = token_info.acc
		p.Token = token_info.token
	}

	ip_port := strings.Split(r.RemoteAddr, ":")
	if len(ip_port) >= 2 {
		p.pos = cfg_position.GetPosByIP(ip_port[0])
	}

	p.bhandling = true

	res := &msg_client_message.S2CLoginResponse{}
	res.Acc = proto.String(req.GetAcc())
	res.PlayerId = proto.Int32(playerid)
	res.Name = proto.String(p.db.GetName())

	log.Info("PlayerLoginRequestHandler %s %s %s", req.GetAcc(), req.GetToken())

	p.Send(res)
	//p.Send(res)

	p.OnLogin()

	return 1, p
}

func HeartBeatHandler(conn *socket.TcpConn, msg proto.Message) {

	return
}

func C2SC2SGetPlayerInfoHandler(conn *socket.TcpConn, msg proto.Message) {
	req := msg.(*msg_client_message.C2SGetPlayerInfo)
	if nil == conn || nil == req {
		log.Error("C2SC2SGetPlayerInfoHandler conn or req nil [%d]", nil == req)
		return
	}

	p := player_mgr.GetPlayerById(int32(conn.T))
	if nil == p {
		log.Error("C2SC2SGetPlayerInfoHandler not login [%d]", conn.T)
		return
	}

	req2co := &msg_server_message.GetPlayerInfo{}
	req2co.PlayerId = proto.Int32(p.Id)
	req2co.TgtPlayerId = proto.Int32(req.GetPlayerId())

	center_conn.Send(req2co)

	return
}

// ----------------------------------------------------------------------------
func C2HGetPlayerInfoHandler(c *CenterConnection, msg proto.Message) {
	req := msg.(*msg_server_message.GetPlayerInfo)
	if nil == c || nil == req {
		log.Error("C2HGetPlayerInfoHandler c or req nil [%v]", nil == req)
		return
	}

	tgt_pid := req.GetTgtPlayerId()

	tgp := player_mgr.GetPlayerById(tgt_pid)
	if nil == tgp {
		log.Error("C2HGetPlayerInfoHandler")
		return
	}

	res2co := &msg_server_message.RetPlayerInfo{}
	res2co.TgtPlayerId = proto.Int32(tgt_pid)
	res2co.PlayerId = proto.Int32(req.GetPlayerId())

	c.Send(res2co)
}

func C2HRetPlayerInfoHandler(c *CenterConnection, msg proto.Message) {
	res := msg.(*msg_server_message.RetPlayerInfo)
	if nil == c || nil == res {
		log.Error("C2HRetPlayerInfoHandler c or res nil [%v]", nil == res)
		return
	}

	p := player_mgr.GetPlayerById(res.GetPlayerId())
	if nil == p {
		log.Error("C2HRetPlayerInfoHandler p[%d] nil ", res.GetPlayerId())
		return
	}

	tmp_bi := &msg_client_message.OtherPlayerBaseInfo{}
	res_bi := res.GetBaseInfo()
	if nil != res_bi {
		tmp_bi.PlayerId = proto.Int32(res_bi.GetPlayerId())
		tmp_bi.MatchScore = proto.Int32(res_bi.GetMatchScore())
		tmp_bi.Coins = proto.Int32(res_bi.GetCoins())
		tmp_bi.Diamonds = proto.Int32(res_bi.GetDiamonds())
		tmp_bi.ArenaLvl = proto.Int32(res_bi.GetArenaLvl())
		tmp_bi.MyLvl = proto.Int32(res_bi.GetMyLvl())
		tmp_bi.WinCount = proto.Int32(res_bi.GetWinCount())
		tmp_bi.CurLegBestScore = proto.Int32(res_bi.GetCurLegBestScore())
		tmp_bi.LastLegBestScore = proto.Int32(res_bi.GetLastLegBestScore())
		tmp_bi.OfenCardCfgId = proto.Int32(res_bi.GetOfenCardCfgId())
		tmp_bi.DonateCount = proto.Int32(res_bi.GetDonateCount())
		tmp_bi.CheModWinCount = proto.Int32(res_bi.GetCheModWinCount())
		tmp_bi.CheModeOfenCardCfg = proto.Int32(res_bi.GetCheModeOfenCardCfg())
		tmp_bi.Camp = proto.Int32(res_bi.GetCamp())
		tmp_bi.CurLegScore = proto.Int32(res_bi.GetMatchScore())
		tmp_bi.TongIcon = proto.Int32(res_bi.GetTongIcon())
		tmp_bi.TongName = proto.String(res_bi.GetTongName())
		tmp_bi.FightCardIds = res_bi.GetFightCardIds()
		tmp_bi.FightCardLvls = res_bi.GetFightCardLvls()
		tmp_bi.CurCardGetNum = proto.Int32(res_bi.GetCurCardGetNum())
		tmp_bi.IfCaptain = proto.Int32(res_bi.GetIfCaptain())
	} else {
		log.Error("C2HRetPlayerInfoHandler tmp_bi nil")
		return
	}

	res2cli := &msg_client_message.S2CRetPlayerInfo{}
	res2cli.BaseInfo = tmp_bi

	p.Send(res2cli)

	return
}
