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

func (this *Player) send_notify_state() {
	var response *msg_client_message.S2CStateNotify

	// 挂机收益
	s := this.check_income_state()
	if s != 0 {
		if response == nil {
			response = &msg_client_message.S2CStateNotify{}
		}
	}
	if s > 0 {
		response.States = append(response.States, int32(msg_client_message.MODULE_STATE_HANGUP_RANDOM_INCOME))
	} else if s < 0 {
		response.CancelStates = append(response.CancelStates, int32(msg_client_message.MODULE_STATE_HANGUP_RANDOM_INCOME))
	}

	// 其他
	if this.states_changed != nil {
		if response == nil {
			response = &msg_client_message.S2CStateNotify{}
		}
		for k, v := range this.states_changed {
			if v == 1 {
				response.States = append(response.States, k)
			} else if v == 2 {
				response.CancelStates = append(response.CancelStates, k)
			}
		}
		this.states_changed = nil
	}

	if response != nil {
		this.Send(uint16(msg_client_message_id.MSGID_S2C_STATE_NOTIFY), response)
	}
}

func (this *Player) notify_state_changed(state int32, change_type int32) {
	if this.states_changed == nil {
		this.states_changed = make(map[int32]int32)
	}
	this.states_changed[state] = change_type
}

//==============================================================================
func (this *PlayerManager) RegMsgHandler() {
	msg_handler_mgr.SetMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ENTER_GAME_REQUEST), C2SEnterGameRequestHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_LEAVE_GAME_REQUEST), C2SLeaveGameRequestHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_TEST_COMMAND), C2STestCommandHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_HEARTBEAT), C2SHeartbeatHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_DATA_SYNC_REQUEST), C2SDataSyncHandler)

	// 战役
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_BATTLE_RESULT_REQUEST), C2SFightHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_SET_TEAM_REQUEST), C2SSetTeamHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_BATTLE_SET_HANGUP_CAMPAIGN_REQUEST), C2SSetHangupCampaignHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_CAMPAIGN_HANGUP_INCOME_REQUEST), C2SCampaignHangupIncomeHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_CAMPAIGN_DATA_REQUEST), C2SCampaignDataHandler)

	// 角色
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ROLE_ATTRS_REQUEST), C2SRoleAttrsHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ROLE_LEVELUP_REQUEST), C2SRoleLevelUpHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ROLE_RANKUP_REQUEST), C2SRoleRankUpHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ROLE_DECOMPOSE_REQUEST), C2SRoleDecomposeHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ROLE_FUSION_REQUEST), C2SRoleFusionHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ROLE_LOCK_REQUEST), C2SRoleLockHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ROLE_HANDBOOK_REQUEST), C2SRoleHandbookHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ROLE_LEFTSLOT_OPEN_REQUEST), C2SRoleLeftSlotOpenHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ROLE_ONEKEY_EQUIP_REQUEST), C2SRoleOneKeyEquipHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ROLE_ONEKEY_UNEQUIP_REQUEST), C2SRoleOneKeyUnequipHandler)

	// 物品
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ITEM_FUSION_REQUEST), C2SItemFusionHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ITEM_SELL_REQUEST), C2SItemSellHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ITEM_EQUIP_REQUEST), C2SItemEquipHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ITEM_UNEQUIP_REQUEST), C2SItemUnequipHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_ITEM_UPGRADE_REQUEST), C2SItemUpgradeHandler)

	// 邮件
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_MAIL_SEND_REQUEST), C2SMailSendHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_MAIL_LIST_REQUEST), C2SMailListHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_MAIL_DETAIL_REQUEST), C2SMailDetailHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_MAIL_GET_ATTACHED_ITEMS_REQUEST), C2SMailGetAttachedItemsHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_MAIL_DELETE_REQUEST), C2SMailDeleteHandler)

	// 录像
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_BATTLE_RECORD_LIST_REQUEST), C2SBattleRecordListHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_BATTLE_RECORD_REQUEST), C2SBattleRecordHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_BATTLE_RECORD_DELETE_REQUEST), C2SBattleRecordDeleteHandler)

	// 天赋
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_TALENT_UP_REQUEST), C2STalentListHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_TALENT_LIST_REQUEST), C2STalentListHandler)

	// 爬塔
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_TOWER_DATA_REQUEST), C2STowerDataHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_TOWER_RECORDS_INFO_REQUEST), C2STowerRecordsInfoHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_TOWER_RECORD_DATA_REQUEST), C2STowerRecordDataHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_TOWER_RANKING_LIST_REQUEST), C2STowerRankingListHandler)

	// 抽卡
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_DRAW_CARD_REQUEST), C2SDrawCardHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_DRAW_DATA_REQUEST), C2SDrawDataHandler)

	// 点金手
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_GOLD_HAND_DATA_REQUEST), C2SGoldHandDataHandler)
	msg_handler_mgr.SetPlayerMsgHandler(uint16(msg_client_message_id.MSGID_C2S_TOUCH_GOLD_REQUEST), C2STouchGoldHandler)
}

func C2SEnterGameRequestHandler(w http.ResponseWriter, r *http.Request, msg_data []byte) (int32, *Player) {
	var p *Player
	var req msg_client_message.C2SEnterGameRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s) !", err.Error())
		return -1, p
	}

	acc := req.GetAcc()
	if "" == acc {
		log.Error("PlayerEnterGameHandler acc empty !")
		return -1, p
	}

	token_info := login_token_mgr.GetTokenByAcc(acc)
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
		pdb.SetCurrReplyMsgNum(0)
		p = new_player(player_id, token_info.acc, token_info.token, pdb)
		p.OnCreate()
		//} else {
		//	p = new_player(p.Id, token_info.acc, token_info.token, pdb)
		//}
		player_mgr.Add2AccMap(p)
		player_mgr.Add2IdMap(p)
		is_new = true
		log.Info("player_db_to_msg new player(%d) !", player_id)
	} else {
		p.Account = token_info.acc
		p.Token = token_info.token
		pdb := dbc.Players.GetRow(p.Id)
		if pdb != nil {
			pdb.SetCurrReplyMsgNum(0)
		}
	}

	ip_port := strings.Split(r.RemoteAddr, ":")
	if len(ip_port) >= 2 {
		p.pos = position_table.GetPosByIP(ip_port[0])
	}

	p.bhandling = true

	p.send_enter_game(acc, p.Id)
	p.OnLogin()
	if !is_new {
		p.send_items()
		p.send_roles()
	}
	p.send_info()
	p.send_teams()
	p.notify_enter_complete()

	log.Info("PlayerEnterGameHandler account[%s] token[%s]", req.GetAcc(), req.GetToken())

	return 1, p
}

func C2SLeaveGameRequestHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SLeaveGameRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s) !", err.Error())
		return -1
	}
	p.OnLogout()
	return 1
}

func C2SHeartbeatHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SHeartbeat
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s) !", err.Error())
		return -1
	}

	if p.IsOffline() {
		log.Error("Player[%v] is offline", p.Id)
		return int32(msg_client_message.E_ERR_PLAYER_IS_OFFLINE)
	}
	if USE_CONN_TIMER_WHEEL == 0 {
		conn_timer_mgr.Insert(p.Id)
	} else {
		conn_timer_wheel.Insert(p.Id)
	}
	p.send_notify_state()
	p.check_and_send_tower_data()

	return 1
}

func C2SDataSyncHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SDataSyncRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}
	if req.Base {
		p.send_info()
	}
	if req.Items {
		p.send_items()
	}
	if req.Roles {
		p.send_roles()
	}
	if req.Teams {
		p.send_teams()
	}
	if req.Campaigns {
		p.send_campaigns()
	}
	return 1
}
