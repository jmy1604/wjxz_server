package main

import (
	"libs/log"
	"main/table_config"
	"math/rand"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"

	_ "time"

	"github.com/golang/protobuf/proto"
)

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

func (this *Player) delete_role(role_id int32) bool {
	if !this.db.Roles.HasIndex(role_id) {
		return false
	}
	this.db.Roles.Remove(role_id)
	this.roles_id_change_info.id_remove(role_id)
	return true
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

func (this *Player) send_roles() {
	msg := &msg_client_message.S2CRolesResponse{}
	msg.Roles = this.db.Roles.BuildMsg()
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ROLES_RESPONSE), msg)
}

func (this *Player) levelup_role(role_id int32) int32 {
	lvl, o := this.db.Roles.GetLevel(role_id)
	if !o {
		log.Error("Player[%v] not have role[%v]", this.Id, role_id)
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_NOT_FOUND)
	}

	if len(levelup_table_mgr.Array) <= int(lvl) {
		log.Error("Player[%v] is already max level[%v]", this.Id, lvl)
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_LEVEL_IS_MAX)
	}

	levelup_data := levelup_table_mgr.Get(lvl)
	if levelup_data == nil {
		log.Error("cant found level[%v] data", lvl)
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_LEVEL_DATA_NOT_FOUND)
	}

	if levelup_data.CardLevelUpRes != nil {
		for i := 0; i < len(levelup_data.CardLevelUpRes)/2; i++ {
			resource_id := levelup_data.CardLevelUpRes[2*i]
			resource_num := levelup_data.CardLevelUpRes[2*i+1]
			if this.get_resource(resource_id) < resource_num {
				return int32(msg_client_message.E_ERR_PLAYER_ITEM_NUM_NOT_ENOUGH)
			}
		}
		for i := 0; i < len(levelup_data.CardLevelUpRes)/2; i++ {
			resource_id := levelup_data.CardLevelUpRes[2*i]
			resource_num := levelup_data.CardLevelUpRes[2*i+1]
			this.add_resource(resource_id, -resource_num)
		}

		this.db.Roles.SetLevel(role_id, lvl+1)
		lvl += 1
	}

	response := &msg_client_message.S2CRoleLevelUpResponse{
		RoleId: role_id,
		Level:  lvl,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ROLE_LEVELUP_RESPONSE), response)

	log.Debug("Player[%v] role[%v] up to level[%v]", this.Id, role_id, lvl)

	return lvl
}

func (this *Player) rankup_role(role_id int32) int32 {
	rank, o := this.db.Roles.GetRank(role_id)
	if !o {
		log.Error("Player[%v] not have role[%v]", this.Id, role_id)
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_NOT_FOUND)
	}

	table_id, _ := this.db.Roles.GetTableId(role_id)
	cards := card_table_mgr.GetCards(table_id)
	if len(cards) <= int(rank) {
		log.Error("Player[%v] is already max rank[%v]", this.Id, rank)
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_RANK_IS_MAX)
	}

	card_data := card_table_mgr.GetRankCard(table_id, rank)
	if card_data == nil {
		log.Error("Cant found card[%v,%v] data", table_id, rank)
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_TABLE_ID_NOT_FOUND)
	}

	rank_data := rankup_table_mgr.Get(rank)
	if rank_data == nil {
		log.Error("Cant found rankup[%v] data", rank)
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_RANKUP_DATA_NOT_FOUND)
	}
	var cost_resources []int32
	if card_data.Type == 1 {
		cost_resources = rank_data.Type1RankUpRes
	} else if card_data.Type == 2 {
		cost_resources = rank_data.Type2RankUpRes
	} else if card_data.Type == 3 {
		cost_resources = rank_data.Type3RankUpRes
	} else {
		log.Error("Card[%v,%v] type[%v] invalid", table_id, rank, card_data.Type)
		return -1
	}

	for i := 0; i < len(cost_resources)/2; i++ {
		resource_id := cost_resources[2*i]
		resource_num := cost_resources[2*i+1]
		rn := this.get_resource(resource_id)
		if rn < resource_num {
			log.Error("Player[%v] rank[%] up failed, resource[%v] num[%v] not enough", this.Id, rank, resource_id, rn)
			return int32(msg_client_message.E_ERR_PLAYER_ITEM_NUM_NOT_ENOUGH)
		}
	}

	for i := 0; i < len(cost_resources)/2; i++ {
		resource_id := cost_resources[2*i]
		resource_num := cost_resources[2*i+1]
		this.add_resource(resource_id, -resource_num)
	}

	rank += 1
	this.db.Roles.SetRank(role_id, rank)

	response := &msg_client_message.S2CRoleRankUpResponse{
		RoleId: role_id,
		Rank:   rank,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ROLE_RANKUP_RESPONSE), response)

	log.Debug("Player[%v] role[%v] up rank[%v]", this.Id, role_id, rank)

	return rank
}

func get_decompose_rank_res(table_id, rank int32) []int32 {
	rank_data := rankup_table_mgr.Get(rank)
	if rank_data == nil {
		log.Error("Cant get rankup[%v] data", rank)
		return nil
	}
	var resources []int32
	card_data := card_table_mgr.GetRankCard(table_id, rank)
	if card_data == nil {
		log.Error("Cant found card[%v,%v] data", table_id, rank)
		return nil
	}
	if card_data.Type == 1 {
		resources = rank_data.Type1DecomposeRes
	} else if card_data.Type == 2 {
		resources = rank_data.Type2DecomposeRes
	} else if card_data.Type == 3 {
		resources = rank_data.Type3DecomposeRes
	} else {
		log.Error("Card[%v,%v] type[%v] invalid", table_id, rank, card_data.Type)
		return nil
	}

	return resources
}

func (this *Player) team_has_role(team_id int32, role_id int32) bool {
	var members []int32
	if team_id == BATTLE_ATTACK_TEAM {
		members = this.db.BattleTeam.GetAttackMembers()
	} else if team_id == BATTLE_DEFENSE_TEAM {
		members = this.db.BattleTeam.GetDefenseMembers()
	}
	if members != nil {
		for _, m := range members {
			if role_id == m {
				return true
			}
		}
	}
	return false
}

func (this *Player) decompose_role(role_id int32) int32 {
	level, o := this.db.Roles.GetLevel(role_id)
	if !o {
		log.Error("Player[%v] not have role[%v]", this.Id, role_id)
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_NOT_FOUND)
	}

	/*if this.team_member_mgr[role_id] != nil {
		log.Error("Player[%v] team has role[%v], cant decompose", this.Id, role_id)
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_IN_TEAM_CANT_DECOMPOSE)
	}*/

	if this.team_has_role(BATTLE_ATTACK_TEAM, role_id) {
		log.Error("Player[%v] attack team has role[%v], cant decompose", this.Id, role_id)
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_IN_TEAM_CANT_DECOMPOSE)
	}

	if this.team_has_role(BATTLE_DEFENSE_TEAM, role_id) {
		log.Error("Player[%v] defense team has role[%v], cant decompose", this.Id, role_id)
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_IN_TEAM_CANT_DECOMPOSE)
	}

	rank, _ := this.db.Roles.GetRank(role_id)
	table_id, _ := this.db.Roles.GetTableId(role_id)

	card_data := card_table_mgr.GetRankCard(table_id, rank)
	if card_data == nil {
		log.Error("Not found card data by table_id[%v] and rank[%v]", table_id, rank)
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_TABLE_ID_NOT_FOUND)
	}

	for i := 0; i < len(card_data.DecomposeRes)/2; i++ {
		item_id := card_data.DecomposeRes[2*i]
		item_num := card_data.DecomposeRes[2*i+1]
		this.add_resource(item_id, item_num)
		if this.tmp_cache_items == nil {
			this.tmp_cache_items = make(map[int32]int32)
		}
		this.tmp_cache_items[item_id] += item_num
	}

	levelup_data := levelup_table_mgr.Get(level)
	if levelup_data == nil {
		log.Error("Not found levelup[%v] data", level)
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_LEVEL_DATA_NOT_FOUND)
	}
	if levelup_data.CardDecomposeRes != nil {
		for i := 0; i < len(levelup_data.CardDecomposeRes)/2; i++ {
			item_id := levelup_data.CardDecomposeRes[2*i]
			item_num := levelup_data.CardDecomposeRes[2*i+1]
			this.add_resource(item_id, item_num)
			if this.tmp_cache_items == nil {
				this.tmp_cache_items = make(map[int32]int32)
			}
			this.tmp_cache_items[item_id] += item_num
		}
	}

	rank_res := get_decompose_rank_res(table_id, rank)
	if rank_res != nil {
		for i := 0; i < len(rank_res)/2; i++ {
			this.add_resource(rank_res[2*i], rank_res[2*i+1])
			if this.tmp_cache_items == nil {
				this.tmp_cache_items = make(map[int32]int32)
			}
			this.tmp_cache_items[rank_res[2*i]] += rank_res[2*i+1]
		}
	}

	this.db.Roles.Remove(role_id)
	role := this.team_member_mgr[role_id]
	if role != nil {
		team_member_pool.Put(role)
		delete(this.team_member_mgr, role_id)
	}

	response := &msg_client_message.S2CRoleDecomposeResponse{
		RoleId: role_id,
	}
	if this.tmp_cache_items != nil {
		for k, v := range this.tmp_cache_items {
			response.GetItems = append(response.GetItems, &msg_client_message.ItemInfo{
				ItemCfgId: k,
				ItemNum:   v,
			})
		}
		this.tmp_cache_items = nil
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ROLE_DECOMPOSE_RESPONSE), response)

	log.Debug("Player[%v] decompose role[%v]", this.Id, role_id)

	return 1
}

func (this *Player) check_fusion_role_cond(cost_role_ids []int32, cost_cond *table_config.FusionCostCond) int32 {
	for i := 0; i < len(cost_role_ids); i++ {
		if !this.db.Roles.HasIndex(cost_role_ids[i]) {
			log.Error("Player[%v] fusion role need role[%v] not found", this.Id, cost_role_ids[i])
			return int32(msg_client_message.E_ERR_PLAYER_FUSION_NEED_ROLE_NOT_FOUND)
		}

		table_id, _ := this.db.Roles.GetTableId(cost_role_ids[i])
		if cost_cond.CostId > 0 && table_id != cost_cond.CostId {
			log.Error("Player[%v] fusion cost role[%v] invalid", this.Id, cost_role_ids[i])
			return int32(msg_client_message.E_ERR_PLAYER_FUSION_ROLE_INVALID)
		} else {
			rank, _ := this.db.Roles.GetRank(cost_role_ids[i])
			card := card_table_mgr.GetRankCard(table_id, rank)
			if card == nil {
				log.Error("Player[%v] fusion role[%v] not found card[%v] with rank[%v]", this.Id, cost_role_ids[i], table_id, rank)
				return int32(msg_client_message.E_ERR_PLAYER_FUSION_ROLE_INVALID)
			}
			if card.Camp != cost_cond.CostCamp {
				log.Error("Player[%v] fusion role[%v] camp[%v] invalid", this.Id, cost_role_ids[i], card.Camp)
				return int32(msg_client_message.E_ERR_PLAYER_FUSION_ROLE_INVALID)
			}
			if card.Type != cost_cond.CostType {
				log.Error("Player[%v] fusion role[%v] type[%v] invalid", this.Id, cost_role_ids[i], card.Type)
				return int32(msg_client_message.E_ERR_PLAYER_FUSION_ROLE_INVALID)
			}
			if card.Rarity != cost_cond.CostStar {
				log.Error("Player[%v] fusion role[%v] star[%v] invalid", this.Id, cost_role_ids[i], card.Type)
				return int32(msg_client_message.E_ERR_PLAYER_FUSION_ROLE_INVALID)
			}
		}
	}
	return 1
}

func (this *Player) fusion_role(fusion_id, main_role_id int32, cost_role_ids [][]int32) int32 {
	fusion := fusion_table_mgr.Get(fusion_id)
	if fusion == nil {
		log.Error("Fusion[%v] table data not found", fusion_id)
		return int32(msg_client_message.E_ERR_PLAYER_FUSION_ROLE_TABLE_DATA_NOT_FOUND)
	}

	// 固定配方
	if fusion.FusionType == 1 {
		if !this.db.Roles.HasIndex(main_role_id) {
			log.Error("Player[%v] fusion[%v] not found main role[%v]", this.Id, fusion_id, main_role_id)
			return int32(msg_client_message.E_ERR_PLAYER_FUSION_MAIN_ROLE_NOT_FOUND)
		}
	}

	for i := 0; i < len(cost_role_ids); i++ {
		if i >= len(fusion.CostConds) {
			break
		}
		if cost_role_ids[i] != nil && int(fusion.CostConds[i].CostNum) > len(cost_role_ids[i]) {
			log.Error("Player[%v] fusion[%v] cost[%v] not enough", this.Id, fusion_id, fusion.CostConds[i].CostId)
			return int32(msg_client_message.E_ERR_PLAYER_FUSION_ROLE_MATERIAL_NOT_ENOUGH)
		}
		res := this.check_fusion_role_cond(cost_role_ids[i], fusion.CostConds[i])
		if res < 0 {
			return res
		}
	}

	var item *msg_client_message.ItemInfo
	var o bool
	if o, item = this.drop_item_by_id(fusion.ResultDropID, true, true); !o {
		log.Error("Player[%v] fusion[%v] drop new card failed", this.Id, fusion_id)
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_FUSION_FAILED)
	}

	for i := 0; i < len(cost_role_ids); i++ {
		for j := 0; j < len(cost_role_ids[i]); j++ {
			this.delete_role(cost_role_ids[i][j])
		}
	}

	new_role_id := int32(0)
	if fusion.FusionType == 1 {
		this.db.Roles.SetTableId(main_role_id, item.ItemCfgId)
		this.roles_id_change_info.id_update(main_role_id)
	} else {
		new_role_id = this.new_role(item.ItemCfgId, 1, 1)
	}

	response := &msg_client_message.S2CRoleFusionResponse{
		NewCardId: item.ItemCfgId,
		RoleId:    new_role_id,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ROLE_FUSION_RESPONSE), response)

	log.Debug("Player[%v] fusion[%v] main_card[%v] get new role[%v] new card[%v], cost cards[%v]", this.Id, fusion_id, main_role_id, new_role_id, item.ItemCfgId, cost_role_ids)

	return 1
}

func C2SRoleLevelUpHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SRoleLevelUpRequest
	err := proto.Unmarshal(msg_data, &req)
	if nil != err {
		log.Error("Unmarshal msg failed err(%s) !", err.Error())
		return -1
	}
	return p.levelup_role(req.GetRoleId())
}

func C2SRoleRankUpHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SRoleRankUpRequest
	err := proto.Unmarshal(msg_data, &req)
	if nil != err {
		log.Error("Unmarshal msg failed err(%s) !", err.Error())
		return -1
	}
	return p.rankup_role(req.GetRoleId())
}

func C2SRoleDecomposeHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SRoleDecomposeRequest
	err := proto.Unmarshal(msg_data, &req)
	if nil != err {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}
	return p.decompose_role(req.GetRoleId())
}

func C2SRoleFusionHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SRoleFusionRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}
	return p.fusion_role(req.GetFusionId(), req.GetMainCardId(), [][]int32{req.GetCost1CardIds(), req.GetCost2CardIds(), req.GetCost3CardIds()})
}