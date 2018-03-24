package main

import (
	//"fmt"
	"libs/log"
	_ "main/table_config"
	"math/rand"
	_ "net/http"
	"public_message/gen_go/client_message"
	_ "public_message/gen_go/server_message"
	"sync"
	"time"

	_ "3p/code.google.com.protobuf/proto"
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

	team_member_mgr map[int32]*TeamMember
	attack_team     BattleTeam
	defense_team    BattleTeam
	team_changed    map[int32]bool

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

func (this *Player) new_role(role_id int32, rank int32, level int32) bool {
	if this.db.Roles.HasIndex(role_id) {
		log.Error("Player[%v] already has role[%v]", this.Id, role_id)
		return false
	}

	card := card_table_mgr.GetRankCard(role_id, rank)
	if card == nil {
		log.Error("Cant get role card by id[%] rank[%v]", role_id, rank)
		return false
	}
	var role dbPlayerRoleData
	role.Id = role_id
	role.Rank = rank
	role.Level = level
	this.db.Roles.Add(&role)
	return true
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
		log.Debug("Player[%v] rand role[%v]", this.Id, table_id)
	}

	return id
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

func (this *Player) SendBaseInfo() {
}
func (this *Player) ChkSendNotifyState() {
}

func (this *Player) PopCurMsgData() []byte {
	if this.b_base_prop_chg {
		this.SendBaseInfo()
	}
	this.ChkSendNotifyState()
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
	// 新任务
	this.UpdateNewTasks(1, false)

	return
}

func (this *Player) OnLogin() {
	gm_command_mgr.OnPlayerLogin(this)
	this.ChkPlayerDialyTask()
	this.db.Info.SetLastLogin(int32(time.Now().Unix()))
	this.team_member_mgr = make(map[int32]*TeamMember)
	this.team_changed = make(map[int32]bool)
}

func (this *Player) OnLogout() {
	log.Info("玩家[%d] 登出 ！！", this.Id)
}

// ----------------------------------------------------------------------------

// ======================================================================

func reg_player_base_info_msg() {
}

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

func (this *Player) SetAttackTeam(team []int32) bool {
	if team == nil {
		return false
	}
	for i := 0; i < len(team); i++ {
		if i >= BATTLE_TEAM_MEMBER_MAX_NUM {
			break
		}
		if !this.db.Roles.HasIndex(team[i]) {
			log.Warn("Player[%v] not has role[%v] for set attack team", this.Id, team[i])
			return false
		}
	}
	this.db.BattleTeam.SetAttackMembers(team)
	this.team_changed[BATTLE_ATTACK_TEAM] = true
	if !this.attack_team.Init(this, BATTLE_ATTACK_TEAM) {
		log.Warn("Player[%v] init attack team failed", this.Id)
	}
	return true
}

func (this *Player) SetDefenseTeam(team []int32) bool {
	if team == nil {
		return false
	}
	for i := 0; i < len(team); i++ {
		if i >= BATTLE_TEAM_MEMBER_MAX_NUM {
			break
		}
		if !this.db.Roles.HasIndex(team[i]) {
			log.Warn("Player[%v] not has role[%v] for set defense team", this.Id, team[i])
			return false
		}
	}
	this.db.BattleTeam.SetDefenseMembers(team)
	this.team_changed[BATTLE_ATTACK_TEAM] = false
	if !this.defense_team.Init(this, BATTLE_DEFENSE_TEAM) {
		log.Warn("Player[%v] init defense team failed", this.Id)
	}
	return true
}

func (this *Player) Fight2Player(player_id int32) int32 {
	p := player_mgr.GetPlayerById(player_id)
	if p == nil {
		return int32(msg_client_message.E_ERR_PLAYER_NOT_EXIST)
	}

	//changed, o := this.team_changed[BATTLE_ATTACK_TEAM]
	//if changed || !o {
	if !this.attack_team.Init(this, BATTLE_ATTACK_TEAM) {
		log.Error("Player[%v] init attack team failed", this.Id)
		return -1
	}
	//}
	//changed, o = p.team_changed[BATTLE_ATTACK_TEAM]
	//if changed || !o {
	if !p.defense_team.Init(this, BATTLE_DEFENSE_TEAM) {
		log.Error("Player[%v] init defense team failed", player_id)
		return -1
	}
	//}

	this.attack_team.Fight(&p.defense_team, BATTLE_END_BY_ALL_DEAD, 0)

	return 1
}
