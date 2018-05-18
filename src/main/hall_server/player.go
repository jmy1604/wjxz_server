package main

import (
	"libs/log"
	_ "main/table_config"
	"math/rand"
	_ "net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	"sync"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
)

const (
	INIT_PLAYER_MSG_NUM = 10

	MSG_ITEM_HEAD_LEN = 4

	BUILDING_ADD_MSG_INIT_LEN = 5
	BUILDING_ADD_MSG_ADD_STEP = 2
)

type PlayerMsgItem struct {
	data          []byte
	data_len      int32
	data_head_len int32
	msg_code      uint16
}

type IdChangeInfo struct {
	add    []int32
	remove []int32
	update []int32
}

func (this *IdChangeInfo) is_changed() bool {
	if this.add != nil || this.remove != nil || this.update != nil {
		return true
	}
	return false
}

func (this *IdChangeInfo) id_add(id int32) {
	if this.add == nil {
		this.add = []int32{id}
	} else {
		this.add = append(this.add, id)
	}
}

func (this *IdChangeInfo) id_remove(id int32) {
	if this.remove == nil {
		this.remove = []int32{id}
	} else {
		this.remove = append(this.remove, id)
	}
}

func (this *IdChangeInfo) id_update(id int32) {
	if this.update == nil {
		this.update = []int32{id}
	} else {
		this.update = append(this.update, id)
	}
}

func (this *IdChangeInfo) reset() {
	this.add = nil
	this.remove = nil
	this.update = nil
}

type Player struct {
	Id            int32
	Account       string
	Token         string
	ol_array_idx  int32
	all_array_idx int32
	db            *dbPlayerRow
	pos           int32

	bhandling          bool
	msg_items          []*PlayerMsgItem
	msg_items_lock     *sync.Mutex
	cur_msg_items_len  int32
	max_msg_items_len  int32
	total_msg_data_len int32
	b_base_prop_chg    bool

	new_unlock_chapter_id int32
	used_drop_ids         map[int32]int32       // 抽卡掉落ID统计
	world_chat_data       PlayerWorldChatData   // 世界聊天缓存数据
	anouncement_data      PlayerAnouncementData // 公告缓存数据

	team_member_mgr map[int32]*TeamMember // 成员map
	attack_team     *BattleTeam           // 进攻阵营
	defense_team    *BattleTeam           // 防守阵营
	use_defense     int32                 // 是否防守阵容
	stage_team      *BattleTeam           // 关卡阵营
	stage_id        int32
	stage_wave      int32

	roles_id_change_info IdChangeInfo    // 角色增删更新
	items_changed_info   map[int32]int32 // 物品增删更新
	tmp_cache_items      map[int32]int32

	msg_acts_lock    *sync.Mutex
	cur_msg_acts_len int32
	max_msg_acts_len int32
}

func new_player(id int32, account, token string, db *dbPlayerRow) *Player {
	ret_p := &Player{}
	ret_p.Id = id
	ret_p.Account = account
	ret_p.Token = token
	ret_p.db = db
	ret_p.ol_array_idx = -1
	ret_p.all_array_idx = -1

	ret_p.max_msg_items_len = INIT_PLAYER_MSG_NUM
	ret_p.msg_items_lock = &sync.Mutex{}
	ret_p.msg_items = make([]*PlayerMsgItem, ret_p.max_msg_items_len)

	ret_p.msg_acts_lock = &sync.Mutex{}

	return ret_p
}

func new_player_with_db(id int32, db *dbPlayerRow) *Player {
	if id <= 0 || nil == db {
		log.Error("new_player_with_db param error !", id, nil == db)
		return nil
	}

	ret_p := &Player{}
	ret_p.Id = id
	ret_p.db = db
	ret_p.ol_array_idx = -1
	ret_p.all_array_idx = -1
	ret_p.Account = db.GetAccount()

	ret_p.max_msg_items_len = INIT_PLAYER_MSG_NUM
	ret_p.msg_items_lock = &sync.Mutex{}
	ret_p.msg_items = make([]*PlayerMsgItem, ret_p.max_msg_items_len)

	ret_p.msg_acts_lock = &sync.Mutex{}

	return ret_p
}

func (this *Player) new_role(role_id int32, rank int32, level int32) int32 {
	if this.db.Roles.HasIndex(role_id) {
		log.Error("Player[%v] already has role[%v]", this.Id, role_id)
		return 0
	}

	card := card_table_mgr.GetRankCard(role_id, rank)
	if card == nil {
		log.Error("Cant get role card by id[%v] rank[%v]", role_id, rank)
		return 0
	}
	var role dbPlayerRoleData
	role.TableId = role_id
	role.Id = this.db.Global.IncbyCurrentRoleId(1)
	role.Rank = rank
	role.Level = level
	this.db.Roles.Add(&role)

	this.roles_id_change_info.id_add(role.Id)

	return role.Id
}

func (this *Player) has_role(id int32) bool {
	all := this.db.Roles.GetAllIndex()
	for i := 0; i < len(all); i++ {
		table_id, o := this.db.Roles.GetTableId(all[i])
		if o && table_id == id {
			return true
		}
	}
	return false
}

func (this *Player) rand_role() int32 {
	if card_table_mgr.Array == nil {
		return 0
	}

	c := len(card_table_mgr.Array)
	r := rand.Intn(c)
	cr := r
	table_id := int32(0)
	for {
		table_id = card_table_mgr.Array[r%c].Id
		if !this.has_role(table_id) {
			break
		}
		r += 1
		if r-cr >= c {
			// 允许重复
			//table_id = 0
			break
		}
	}

	id := int32(0)
	if table_id > 0 {
		id = this.db.Global.IncbyCurrentRoleId(1)
		this.db.Roles.Add(&dbPlayerRoleData{
			Id:      id,
			TableId: table_id,
			Rank:    1,
			Level:   1,
		})

		this.roles_id_change_info.id_add(id)
		log.Debug("Player[%v] rand role[%v]", this.Id, table_id)
	}

	return id
}

func (this *Player) check_and_send_roles_change() {
	if this.roles_id_change_info.is_changed() {
		var msg msg_client_message.S2CRolesChangeNotify
		if this.roles_id_change_info.add != nil {
			roles := this.db.Roles.BuildSomeMsg(this.roles_id_change_info.add)
			if roles != nil {
				msg.Adds = roles
			}
		}
		if this.roles_id_change_info.remove != nil {
			msg.Removes = this.roles_id_change_info.remove
		}
		if this.roles_id_change_info.update != nil {
			roles := this.db.Roles.BuildSomeMsg(this.roles_id_change_info.update)
			if roles != nil {
				msg.Updates = roles
			}
		}
		this.roles_id_change_info.reset()
		this.Send(uint16(msg_client_message_id.MSGID_S2C_ROLES_CHANGE_NOTIFY), &msg)
	}
}

func (this *Player) check_and_send_items_change() {
	if this.items_changed_info != nil {
		var msg msg_client_message.S2CItemsUpdate
		for k, v := range this.items_changed_info {
			msg.ItemsAdd = append(msg.ItemsAdd, &msg_client_message.ItemInfo{
				ItemCfgId: k,
				ItemNum:   v,
			})
		}
		this.Send(uint16(msg_client_message_id.MSGID_S2C_ITEMS_UPDATE), &msg)
		this.items_changed_info = nil
	}
}

func (this *Player) add_msg_data(msg_code uint16, data []byte) {
	if nil == data {
		log.Error("Player add_msg_data !")
		return
	}

	this.msg_items_lock.Lock()
	defer this.msg_items_lock.Unlock()

	if this.cur_msg_items_len >= this.max_msg_items_len {
		new_max := this.max_msg_items_len + 5
		new_msg_items := make([]*PlayerMsgItem, new_max)
		for idx := int32(0); idx < this.max_msg_items_len; idx++ {
			new_msg_items[idx] = this.msg_items[idx]
		}

		this.msg_items = new_msg_items
		this.max_msg_items_len = new_max
	}

	new_item := &PlayerMsgItem{}
	new_item.msg_code = msg_code
	new_item.data = data
	new_item.data_len = int32(len(data))
	this.total_msg_data_len += new_item.data_len + MSG_ITEM_HEAD_LEN
	this.msg_items[this.cur_msg_items_len] = new_item

	this.cur_msg_items_len++

	return
}

func (this *Player) PopCurMsgData() []byte {
	if this.b_base_prop_chg {

	}

	this.check_and_send_roles_change()
	this.check_and_send_items_change()
	this.CheckAndAnouncement()

	this.msg_items_lock.Lock()
	defer this.msg_items_lock.Unlock()

	this.bhandling = false

	out_bytes := make([]byte, this.total_msg_data_len)
	tmp_len := int32(0)
	var tmp_item *PlayerMsgItem
	for idx := int32(0); idx < this.cur_msg_items_len; idx++ {
		tmp_item = this.msg_items[idx]
		if nil == tmp_item {
			continue
		}

		out_bytes[tmp_len] = byte(tmp_item.msg_code >> 8)
		out_bytes[tmp_len+1] = byte(tmp_item.msg_code & 0xFF)
		out_bytes[tmp_len+2] = byte(tmp_item.data_len >> 8)
		out_bytes[tmp_len+3] = byte(tmp_item.data_len & 0xFF)
		tmp_len += 4
		copy(out_bytes[tmp_len:], tmp_item.data)
		tmp_len += tmp_item.data_len
	}

	this.cur_msg_items_len = 0
	this.total_msg_data_len = 0
	return out_bytes
}

func (this *Player) Send(msg_id uint16, msg proto.Message) {
	if !this.bhandling {
		log.Error("Player [%d] send msg[%d] no bhandling !", this.Id, msg_id)
		return
	}

	log.Info("[发送] [玩家%d:%v] [%s] !", this.Id, msg_id, msg.String())

	data, err := proto.Marshal(msg)
	if nil != err {
		log.Error("Player Marshal msg failed err[%s] !", err.Error())
		return
	}

	this.add_msg_data(msg_id, data)
}

func (this *Player) add_init_roles() {
	var teams []int32
	init_roles := global_config_mgr.GetGlobalConfig().InitRoles
	for i := 0; i < len(init_roles)/3; i++ {
		iid := this.new_role(init_roles[3*i], init_roles[3*i+1], init_roles[3*i+2])
		if teams == nil {
			teams = []int32{iid}
		} else if len(teams) < BATTLE_TEAM_MEMBER_MAX_NUM {
			teams = append(teams, iid)
		}
	}
	this.db.BattleTeam.SetAttackMembers(teams)
	this.db.BattleTeam.SetDefenseMembers(teams)
}

func (this *Player) add_all_items() {
}

func (this *Player) OnCreate() {
	// 随机初始名称
	tmp_acc := this.Account
	if len(tmp_acc) > 6 {
		tmp_acc = string([]byte(tmp_acc)[0:6])
	}
	this.db.Info.SetLvl(1)
	this.db.Info.SetCreateUnix(int32(time.Now().Unix()))
	this.add_init_roles()

	// 新任务
	this.UpdateNewTasks(1, false)

	return
}

func (this *Player) OnLogin() {
	conn_timer_mgr.Insert(this.Id)

	gm_command_mgr.OnPlayerLogin(this)
	this.ChkPlayerDialyTask()
	this.db.Info.SetLastLogin(int32(time.Now().Unix()))
	this.team_member_mgr = make(map[int32]*TeamMember)
}

func (this *Player) OnLogout() {
	conn_timer_mgr.Remove(this.Id)

	// 离线收益时间开始
	this.db.Info.SetLastLogout(int32(time.Now().Unix()))
	log.Info("玩家[%d] 登出 ！！", this.Id)
}

func (this *Player) IsOffline() bool {
	diff := this.db.Info.GetLastLogout() - this.db.Info.GetLastLogin()
	return diff >= 0
}

func (this *dbPlayerRoleColumn) BuildMsg() (roles []*msg_client_message.Role) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerRoleColumn.BuildMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()

	for _, v := range this.m_data {
		role := &msg_client_message.Role{
			Id:      v.Id,
			TableId: v.TableId,
			Rank:    v.Rank,
			Level:   v.Level,
			Attrs:   v.Attr,
		}
		roles = append(roles, role)
	}
	return
}

func (this *dbPlayerRoleColumn) BuildSomeMsg(ids []int32) (roles []*msg_client_message.Role) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerRoleColumn.BuildOneMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()

	for i := 0; i < len(ids); i++ {
		v, o := this.m_data[ids[i]]
		if !o {
			return
		}

		role := &msg_client_message.Role{
			Id:      v.Id,
			TableId: v.TableId,
			Rank:    v.Rank,
			Level:   v.Level,
			Attrs:   v.Attr,
		}
		roles = append(roles, role)
	}
	return
}

func (this *Player) send_enter_game(acc string, id int32) {
	res := &msg_client_message.S2CEnterGameResponse{}
	res.Acc = acc
	res.PlayerId = id
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ENTER_GAME_RESPONSE), res)
}

func (this *Player) send_roles() {
	msg := &msg_client_message.S2CRolesResponse{}
	msg.Roles = this.db.Roles.BuildMsg()
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ROLES_RESPONSE), msg)
}

func (this *Player) send_teams() {
	msg := &msg_client_message.S2CTeamsResponse{}
	attack_team := &msg_client_message.TeamData{
		TeamType:    0,
		TeamMembers: this.db.BattleTeam.GetAttackMembers(),
	}
	defense_team := &msg_client_message.TeamData{
		TeamType:    1,
		TeamMembers: this.db.BattleTeam.GetDefenseMembers(),
	}
	msg.Teams = []*msg_client_message.TeamData{attack_team, defense_team}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_TEAMS_RESPONSE), msg)
}

func (this *Player) notify_enter_complete() {
	msg := &msg_client_message.S2CEnterGameCompleteNotify{}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ENTER_GAME_COMPLETE_NOTIFY), msg)
}

// ----------------------------------------------------------------------------

// ======================================================================
/*
func C2SShopItemsHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SShopItems)
	if req == nil || nil == p {
		log.Error("C2SShopItems proto is invalid")
		return -1
	}

	if req.GetShopId() == 0 {
		var shop_type = []int32{
			table_config.SHOP_TYPE_SPECIAL,
			table_config.SHOP_TYPE_CHARM_MEDAL,
			table_config.SHOP_TYPE_FRIEND_POINTS,
			table_config.SHOP_TYPE_RMB,
			table_config.SHOP_TYPE_SOUL_STONE,
		}
		for i := 0; i < len(shop_type); i++ {
			if res := p.fetch_shop_limit_items(shop_type[i], true); res < 0 {
				return res
			}
		}
		return 1
	}
	return p.fetch_shop_limit_items(req.GetShopId(), true)
}

func C2SBuyShopItemHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SBuyShopItem)
	if req == nil || nil == p {
		log.Error("C2SBuyShopItem proto is invalid")
		return -1
	}
	if p.check_shop_limited_days_items_refresh_by_shop_itemid(req.GetItemId(), true) {
		log.Info("刷新了商店")
		return 1
	}
	return p.buy_item(req.GetItemId(), req.GetItemNum(), true)
}

func C2SGetInfoHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SGetInfo)
	if req == nil || nil == p {
		log.Error("C2SGetInfo proto is invalid")
		return -1
	}

	if req.GetBase() {
		p.SendBaseInfo()
	}

	if req.GetItem() {
		m := &msg_client_message.S2CRetItemInfos{}
		p.db.Items.FillAllMsg(m)
		p.Send(m)
	}

	if req.GetStage() {
		p.send_stage_info()
	}

	if req.GetGuide() {
		p.SyncPlayerGuideData()
	}

	return 1
}

func C2SHeartHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) (ret_val int32) {

	return 1
}

func C2SGetHandbookHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SGetHandbook)
	if req == nil {
		log.Error("C2SGetHandbook proto is invalid")
		return -1
	}
	if p == nil {
		log.Error("C2SGetHandbook player is nil")
		return -1
	}

	response := &msg_client_message.S2CGetHandbookResult{}
	all_ids := p.db.HandbookItems.GetAllIndex()
	if all_ids == nil || len(all_ids) == 0 {
		response.Items = make([]int32, 0)
	} else {
		n := 0
		response.Items = make([]int32, len(all_ids))
		for i := 0; i < len(all_ids); i++ {
			handbook := handbook_table_mgr.Get(all_ids[i])
			if handbook == nil {
				log.Warn("Player[%v] load handbook[%v] not found", p.Id, all_ids[i])
				continue
			}
			response.Items[n] = all_ids[i]
			n += 1
		}
		response.Items = response.Items[:n]
	}
	suit_ids := p.db.SuitAwards.GetAllIndex()
	if suit_ids == nil || len(suit_ids) == 0 {
		response.AwardSuitId = make([]int32, 0)
	} else {
		response.AwardSuitId = make([]int32, len(suit_ids))
		for i := 0; i < len(suit_ids); i++ {
			response.AwardSuitId[i] = suit_ids[i]
		}
	}
	p.Send(response)
	return 1
}

func C2SGetHeadHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SGetHead)
	if req == nil {
		log.Error("C2SGetHead proto is invalid")
		return -1
	}
	if p == nil {
		log.Error("C2SGetHead player is nil")
		return -1
	}

	response := &msg_client_message.S2CGetHeadResult{}
	all_ids := p.db.HeadItems.GetAllIndex()
	if all_ids == nil || len(all_ids) == 0 {
		response.Items = make([]int32, 0)
	} else {
		response.Items = make([]int32, len(all_ids))
		for i := 0; i < len(all_ids); i++ {
			response.Items[i] = all_ids[i]
		}
	}
	p.Send(response)
	return 1
}

func C2SGetSuitHandbookRewardHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	return 1
}

func (this *Player) get_stage_total_score_rank_list(rank_start, rank_num int32) int32 {
	if rank_num > global_config_mgr.GetGlobalConfig().RankingListOnceGetItemsNum {
		return int32(msg_client_message.E_ERR_RANK_GET_ITEMS_NUM_OVER_MAX)
	}

	result := this.rpc_ranklist_stage_total_score(rank_start, rank_num)
	if result == nil {
		log.Error("Player[%v] rpc get stages total score rank list range[%v,%v] failed", this.Id, rank_start, rank_num)
		return -1
	}

	var items []*msg_client_message.RankingListItemInfo
	if result.RankItems == nil {
		items = make([]*msg_client_message.RankingListItemInfo, 0)
	} else {
		now_time := time.Now()
		items = make([]*msg_client_message.RankingListItemInfo, len(result.RankItems))
		for i := int32(0); i < int32(len(result.RankItems)); i++ {
			r := result.RankItems[i]
			is_friend := this.db.Friends.HasIndex(r.PlayerId)
			is_zaned := this.is_today_zan(r.PlayerId, now_time)
			name, level, head := GetPlayerBaseInfo(r.PlayerId)
			items[i] = &msg_client_message.RankingListItemInfo{
				Rank:                  proto.Int32(rank_start + i),
				PlayerId:              proto.Int32(r.PlayerId),
				PlayerName:            proto.String(name),
				PlayerLevel:           proto.Int32(level),
				PlayerHead:            proto.String(head),
				PlayerStageTotalScore: proto.Int32(r.TotalScore),
				IsFriend:              proto.Bool(is_friend),
				IsZaned:               proto.Bool(is_zaned),
			}
		}
	}

	response := &msg_client_message.S2CPullRankingListResult{}
	response.ItemList = items
	response.RankType = proto.Int32(1)
	response.StartRank = proto.Int32(rank_start)
	response.SelfRank = proto.Int32(result.SelfRank)
	if result.SelfRank == 0 {
		response.SelfValue1 = proto.Int32(this.db.Stages.GetTotalScore())
	} else {
		response.SelfValue1 = proto.Int32(result.SelfTotalScore)
	}
	this.Send(response)

	return 1
}

func (this *Player) get_stage_score_rank_list(stage_id, rank_start, rank_num int32) int32 {
	if rank_num > global_config_mgr.GetGlobalConfig().RankingListOnceGetItemsNum {
		return int32(msg_client_message.E_ERR_RANK_GET_ITEMS_NUM_OVER_MAX)
	}

	result := this.rpc_ranklist_stage_score(stage_id, rank_start, rank_num)
	if result == nil {
		log.Error("Player[%v] rpc get stage[%v] score rank list range[%v,%v] failed", this.Id, stage_id, rank_start, rank_num)
		return -1
	}

	var items []*msg_client_message.RankingListItemInfo
	if result.RankItems == nil {
		items = make([]*msg_client_message.RankingListItemInfo, 0)
	} else {
		now_time := time.Now()
		items = make([]*msg_client_message.RankingListItemInfo, len(result.RankItems))
		for i := int32(0); i < int32(len(result.RankItems)); i++ {
			r := result.RankItems[i]
			is_friend := this.db.Friends.HasIndex(r.PlayerId)
			is_zaned := this.is_today_zan(r.PlayerId, now_time)
			name, level, head := GetPlayerBaseInfo(r.PlayerId)
			items[i] = &msg_client_message.RankingListItemInfo{
				Rank:             proto.Int32(rank_start + i),
				PlayerId:         proto.Int32(r.PlayerId),
				PlayerName:       proto.String(name),
				PlayerLevel:      proto.Int32(level),
				PlayerHead:       proto.String(head),
				PlayerStageId:    proto.Int32(r.StageId),
				PlayerStageScore: proto.Int32(r.StageScore),
				IsFriend:         proto.Bool(is_friend),
				IsZaned:          proto.Bool(is_zaned),
			}
		}
	}

	response := &msg_client_message.S2CPullRankingListResult{}
	response.ItemList = items
	response.RankType = proto.Int32(2)
	response.StartRank = proto.Int32(rank_start)
	response.SelfRank = proto.Int32(result.SelfRank)
	if result.SelfRank == 0 {
		score, _ := this.db.Stages.GetTopScore(stage_id)
		response.SelfValue1 = proto.Int32(score)
	} else {
		response.SelfValue1 = proto.Int32(result.SelfScore)
	}

	this.Send(response)

	return 1
}

func (this *Player) get_charm_rank_list(rank_start, rank_num int32) int32 {
	if rank_num > global_config_mgr.GetGlobalConfig().RankingListOnceGetItemsNum {
		return int32(msg_client_message.E_ERR_RANK_GET_ITEMS_NUM_OVER_MAX)
	}

	result := this.rpc_ranklist_charm(rank_start, rank_num)
	if result == nil {
		log.Error("Player[%v] rpc get charm rank list range[%v,%v] failed", this.Id, rank_start, rank_num)
		return -1
	}

	var items []*msg_client_message.RankingListItemInfo
	if result.RankItems == nil {
		items = make([]*msg_client_message.RankingListItemInfo, 0)
	} else {
		now_time := time.Now()
		items = make([]*msg_client_message.RankingListItemInfo, len(result.RankItems))
		for i := int32(0); i < int32(len(result.RankItems)); i++ {
			r := result.RankItems[i]
			is_friend := this.db.Friends.HasIndex(r.PlayerId)
			is_zaned := this.is_today_zan(r.PlayerId, now_time)
			name, level, head := GetPlayerBaseInfo(r.PlayerId)
			items[i] = &msg_client_message.RankingListItemInfo{
				Rank:        proto.Int32(rank_start + i),
				PlayerId:    proto.Int32(r.PlayerId),
				PlayerName:  proto.String(name),
				PlayerLevel: proto.Int32(level),
				PlayerHead:  proto.String(head),
				PlayerCharm: proto.Int32(r.Charm),
				IsFriend:    proto.Bool(is_friend),
				IsZaned:     proto.Bool(is_zaned),
			}
		}
	}

	response := &msg_client_message.S2CPullRankingListResult{}
	response.ItemList = items
	response.RankType = proto.Int32(3)
	response.StartRank = proto.Int32(rank_start)
	response.SelfRank = proto.Int32(result.SelfRank)
	if result.SelfRank == 0 {
		response.SelfValue1 = proto.Int32(this.db.Info.GetCharmVal())
	} else {
		response.SelfValue1 = proto.Int32(result.SelfCharm)
	}

	this.Send(response)

	return 1
}

func (this *Player) get_zaned_rank_list(rank_start, rank_num int32) int32 {
	if rank_num > global_config_mgr.GetGlobalConfig().RankingListOnceGetItemsNum {
		return int32(msg_client_message.E_ERR_RANK_GET_ITEMS_NUM_OVER_MAX)
	}

	result := this.rpc_ranklist_get_zaned(rank_start, rank_num)
	if result == nil {
		log.Error("Player[%v] rpc get zaned rank list range[%v,%v] failed", this.Id, rank_start, rank_num)
		return -1
	}

	var items []*msg_client_message.RankingListItemInfo
	if result.RankItems == nil {
		items = make([]*msg_client_message.RankingListItemInfo, 0)
	} else {
		now_time := time.Now()
		items = make([]*msg_client_message.RankingListItemInfo, len(result.RankItems))
		for i := int32(0); i < int32(len(result.RankItems)); i++ {
			r := result.RankItems[i]
			is_friend := this.db.Friends.HasIndex(r.PlayerId)
			is_zaned := this.is_today_zan(r.PlayerId, now_time)
			name, level, head := GetPlayerBaseInfo(r.PlayerId)
			items[i] = &msg_client_message.RankingListItemInfo{
				Rank:        proto.Int32(rank_start + i),
				PlayerId:    proto.Int32(r.PlayerId),
				PlayerName:  proto.String(name),
				PlayerLevel: proto.Int32(level),
				PlayerHead:  proto.String(head),
				PlayerZaned: proto.Int32(r.Zaned),
				IsFriend:    proto.Bool(is_friend),
				IsZaned:     proto.Bool(is_zaned),
			}
		}
	}
	response := &msg_client_message.S2CPullRankingListResult{}
	response.ItemList = items
	response.RankType = proto.Int32(5)
	response.StartRank = proto.Int32(rank_start)
	response.SelfRank = proto.Int32(result.SelfRank)
	response.SelfValue1 = proto.Int32(result.SendZaned)
	this.Send(response)

	return 1
}

func C2SPullRankingListHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SPullRankingList)
	if req == nil || p == nil {
		log.Error("C2SPullRankingListHandler Player[%v] proto is invalid", p.Id)
		return -1
	}

	var res int32 = 0
	rank_type := req.GetRankType()
	rank_start := req.GetStartRank()
	if rank_start <= 0 {
		log.Warn("Player[%v] get rank list by type[%v] with rank_start[%v] invalid", p.Id, rank_type, rank_start)
		return -1
	}
	rank_num := req.GetRankNum()
	if rank_num <= 0 {
		log.Warn("Player[%v] get rank list by type[%v] with rank_num[%v] invalid", p.Id, rank_type, rank_num)
		return -1
	}
	param := req.GetParam()
	if rank_type == 1 {
		// 关卡总分
		res = p.get_stage_total_score_rank_list(rank_start, rank_num)
	} else if rank_type == 2 {
		// 关卡积分
		res = p.get_stage_score_rank_list(param, rank_start, rank_num)
	} else if rank_type == 3 {
		// 魅力
		res = p.get_charm_rank_list(rank_start, rank_num)
	} else if rank_type == 4 {

	} else if rank_type == 5 {
		// 被赞
		res = p.get_zaned_rank_list(rank_start, rank_num)
	} else {
		res = -1
		log.Error("Player[%v] pull rank_type[%v] invalid", p.Id, rank_type)
	}

	return res
}

func C2SWorldChatMsgPullHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SWorldChatMsgPull)
	if req == nil || p == nil {
		log.Error("C2SWorldChatMsgPullHandler player[%v] proto is invalid", p.Id)
		return -1
	}
	return p.pull_world_chat()
}

func C2SWorldChatSendHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SWorldChatSend)
	if req == nil || p == nil {
		log.Error("C2SWorldChatSendHandler player[%v] proto is invalid", p.Id)
		return -1
	}
	return p.world_chat(req.GetContent())
}*/

func (this *Player) SetAttackTeam(team []int32) int32 {
	if team == nil {
		return -1
	}

	used_id := make(map[int32]bool)
	for i := 0; i < len(team); i++ {
		if _, o := used_id[team[i]]; o {
			return int32(msg_client_message.E_ERR_PLAYER_SET_ATTACK_MEMBERS_FAILED)
		}
		used_id[team[i]] = true
	}

	for i := 0; i < len(team); i++ {
		if i >= BATTLE_TEAM_MEMBER_MAX_NUM {
			break
		}
		if !this.db.Roles.HasIndex(team[i]) {
			log.Warn("Player[%v] not has role[%v] for set attack team", this.Id, team[i])
			return int32(msg_client_message.E_ERR_PLAYER_SET_ATTACK_MEMBERS_FAILED)
		}
	}
	this.db.BattleTeam.SetAttackMembers(team)
	if !this.attack_team.Init(this, BATTLE_ATTACK_TEAM, 0) {
		log.Warn("Player[%v] init attack team failed", this.Id)
	}
	return 1
}

func (this *Player) SetDefenseTeam(team []int32) int32 {
	if team == nil {
		return -1
	}

	used_id := make(map[int32]bool)
	for i := 0; i < len(team); i++ {
		if _, o := used_id[team[i]]; o {
			return int32(msg_client_message.E_ERR_PLAYER_SET_DEFENSE_MEMBERS_FAILED)
		}
		used_id[team[i]] = true
	}

	for i := 0; i < len(team); i++ {
		if i >= BATTLE_TEAM_MEMBER_MAX_NUM {
			break
		}
		if !this.db.Roles.HasIndex(team[i]) {
			log.Warn("Player[%v] not has role[%v] for set defense team", this.Id, team[i])
			return int32(msg_client_message.E_ERR_PLAYER_SET_DEFENSE_MEMBERS_FAILED)
		}
	}
	this.db.BattleTeam.SetDefenseMembers(team)
	if !this.defense_team.Init(this, BATTLE_DEFENSE_TEAM, 1) {
		log.Warn("Player[%v] init defense team failed", this.Id)
	}
	return 1
}

func (this *Player) IsDefensing() bool {
	return atomic.CompareAndSwapInt32(&this.use_defense, 0, 1)
}

func (this *Player) CancelDefensing() bool {
	return atomic.CompareAndSwapInt32(&this.use_defense, 1, 0)
}

func (this *Player) Fight2Player(player_id int32) int32 {
	p := player_mgr.GetPlayerById(player_id)
	if p == nil {
		return int32(msg_client_message.E_ERR_PLAYER_NOT_EXIST)
	}

	// 是否正在防守
	if !p.IsDefensing() {
		log.Warn("Player[%v] is defensing, player[%v] fight failed", player_id, this.Id)
		return int32(msg_client_message.E_ERR_PLAYER_IS_DEFENSING)
	}

	if this.attack_team == nil {
		this.attack_team = &BattleTeam{}
	}
	if !this.attack_team.Init(this, BATTLE_ATTACK_TEAM, 0) {
		log.Error("Player[%v] init attack team failed", this.Id)
		return -1
	}

	if p.defense_team == nil {
		p.defense_team = &BattleTeam{}
	}
	if !p.defense_team.Init(p, BATTLE_DEFENSE_TEAM, 1) {
		log.Error("Player[%v] init defense team failed", player_id)
		return -1
	}

	my_team := this.attack_team._format_members_for_msg()
	target_team := p.defense_team._format_members_for_msg()
	is_win, enter_reports, rounds := this.attack_team.Fight(p.defense_team, BATTLE_END_BY_ALL_DEAD, 0)

	// 对方防守结束
	p.CancelDefensing()

	if enter_reports == nil {
		enter_reports = make([]*msg_client_message.BattleReportItem, 0)
	}
	if rounds == nil {
		rounds = make([]*msg_client_message.BattleRoundReports, 0)
	}
	response := &msg_client_message.S2CBattleResultResponse{
		IsWin:        is_win,
		EnterReports: enter_reports,
		Rounds:       rounds,
		MyTeam:       my_team,
		TargetTeam:   target_team,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_BATTLE_RESULT_RESPONSE), response)
	Output_S2CBattleResult(this, response)
	return 1
}

func output_report(rr *msg_client_message.BattleReportItem) {
	log.Debug("		 	report: side[%v]", rr.Side)
	log.Debug("					 skill_id: %v", rr.SkillId)
	log.Debug("					 user: Side[%v], Pos[%v], HP[%v], MaxHP[%v], Energy[%v], Damage[%v]", rr.User.Side, rr.User.Pos, rr.User.HP, rr.User.MaxHP, rr.User.Energy, rr.User.Damage)
	if rr.IsSummon {
		if rr.SummonNpcs != nil {
			for n := 0; n < len(rr.SummonNpcs); n++ {
				rrs := rr.SummonNpcs[n]
				if rrs != nil {
					log.Debug("					 summon npc: Side[%v], Pos[%v], Id[%v], TableId[%v], HP[%v], MaxHP[%v], Energy[%v]", rrs.Side, rrs.Pos, rrs.Id, rrs.TableId, rrs.HP, rrs.MaxHP, rrs.Energy)
				}
			}
		}
	} else {
		if rr.BeHiters != nil {
			for n := 0; n < len(rr.BeHiters); n++ {
				rrb := rr.BeHiters[n]
				log.Debug("					 behiter: Side[%v], Pos[%v], HP[%v], MaxHP[%v], Energy[%v], Damage[%v], IsCritical[%v], IsBlock[%v]",
					rrb.Side, rrb.Pos, rrb.HP, rrb.MaxHP, rrb.Energy, rrb.Damage, rrb.IsCritical, rrb.IsBlock)
			}
		}
	}
	if rr.AddBuffs != nil {
		for n := 0; n < len(rr.AddBuffs); n++ {
			log.Debug("					 add buff: Side[%v], Pos[%v], BuffId[%v]", rr.AddBuffs[n].Side, rr.AddBuffs[n].Pos, rr.AddBuffs[n].BuffId)
		}
	}
	if rr.RemoveBuffs != nil {
		for n := 0; n < len(rr.RemoveBuffs); n++ {
			log.Debug("					 remove buff: Side[%v], Pos[%v], BuffId[%v]", rr.RemoveBuffs[n].Side, rr.RemoveBuffs[n].Pos, rr.RemoveBuffs[n].BuffId)
		}
	}

	log.Debug("					 has_combo: %v", rr.HasCombo)
}

func Output_S2CBattleResult(player *Player, m proto.Message) {
	response := m.(*msg_client_message.S2CBattleResultResponse)
	if response.IsWin {
		log.Debug("Player[%v] wins", player.Id)
	} else {
		log.Debug("Player[%v] lost", player.Id)
	}

	if response.MyTeam != nil {
		log.Debug("My team:")
		for i := 0; i < len(response.MyTeam); i++ {
			m := response.MyTeam[i]
			if m == nil {
				continue
			}
			log.Debug("		 Side:%v Id:%v Pos:%v HP:%v MaxHP:%v Energy:%v TableId:%v", m.Side, m.Id, m.Pos, m.HP, m.MaxHP, m.Energy, m.TableId)
		}
	}
	if response.TargetTeam != nil {
		log.Debug("Target team:")
		for i := 0; i < len(response.TargetTeam); i++ {
			m := response.TargetTeam[i]
			if m == nil {
				continue
			}
			log.Debug("		 Side:%v Id:%v Pos:%v HP:%v MaxHP:%v Energy:%v TableId:%v", m.Side, m.Id, m.Pos, m.HP, m.MaxHP, m.Energy, m.TableId)
		}
	}

	if response.EnterReports != nil {
		log.Debug("   before enter:")
		for i := 0; i < len(response.EnterReports); i++ {
			r := response.EnterReports[i]
			output_report(r)
		}
	}

	if response.Rounds != nil {
		log.Debug("Round num: %v", len(response.Rounds))
		for i := 0; i < len(response.Rounds); i++ {
			r := response.Rounds[i]
			log.Debug("	  round[%v]", r.RoundNum)
			if r.Reports != nil {
				for j := 0; j < len(r.Reports); j++ {
					rr := r.Reports[j]
					output_report(rr)
				}
			}
			if r.RemoveBuffs != nil {
				for j := 0; j < len(r.RemoveBuffs); j++ {
					b := r.RemoveBuffs[j]
					log.Debug("		 	remove buffs: Side[%v], Pos[%v], BuffId[%v]", b.Side, b.Pos, b.BuffId)
				}
			}
			if r.ChangedFighters != nil {
				for j := 0; j < len(r.ChangedFighters); j++ {
					m := r.ChangedFighters[j]
					log.Debug("			changed member: Side[%v], Pos[%v], HP[%v], MaxHP[%v], Energy[%v], Damage[%v]", m.Side, m.Pos, m.HP, m.MaxHP, m.Energy, m.Damage)
				}
			}
		}
	}
}
