package main

import (
	"libs/log"
	"libs/utils"
	_ "main/table_config"
	_ "math/rand"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	_ "strconv"
	_ "sync"
	_ "time"

	"github.com/golang/protobuf/proto"
)

func (this *dbGuildStageDamageItemData) Less(item utils.ShortRankItem) bool {
	it := item.(*dbGuildStageDamageItemData)
	if it == nil {
		return false
	}
	if this.Damage < it.Damage {
		return true
	}
	return false
}

func (this *dbGuildStageDamageItemData) Greater(item utils.ShortRankItem) bool {
	it := item.(*dbGuildStageDamageItemData)
	if it == nil {
		return false
	}
	if this.Damage > it.Damage {
		return true
	}
	return false
}

func (this *dbGuildStageDamageItemData) GetKey() interface{} {
	return this.AttackerId
}

func (this *dbGuildStageDamageItemData) GetValue() interface{} {
	return this.Damage
}

func (this *dbGuildStageDamageItemData) Assign(item utils.ShortRankItem) {
	it := item.(*dbGuildStageDamageItemData)
	if it == nil {
		return
	}
	this.AttackerId = it.AttackerId
	this.Damage = it.Damage
}

// ----------------------------------------------------------------------------

func (this *dbGuildStageAttackLogColumn) GetDamageList2(id int32) (v []dbGuildStageDamageItemData, has bool) {
	this.m_row.m_lock.UnSafeRLock("dbGuildStageAttackLogColumn.GetDamageList")
	defer this.m_row.m_lock.UnSafeRUnlock()

	d := this.m_data[id]
	if d == nil {
		return
	}
	v = make([]dbGuildStageDamageItemData, len(d.DamageList))
	for _ii, _vv := range d.DamageList {
		_vv.clone_to(&v[_ii])
	}
	return v, true
}

func (this *dbGuildStageAttackLogColumn) SetDamageList2(id int32, v []*dbGuildStageDamageItemData) (has bool) {
	this.m_row.m_lock.UnSafeLock("dbGuildStageAttackLogColumn.SetDamageList")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d == nil {
		log.Error("not exist %v %v", this.m_row.GetId(), id)
		return
	}
	d.DamageList = make([]dbGuildStageDamageItemData, len(v))
	for _ii, _vv := range v {
		_vv.clone_to(&d.DamageList[_ii])
	}
	this.m_changed = true
	return true
}

// ----------------------------------------------------------------------------

// 获得公会副本伤害排名
func guild_stage_damage_list(guild_id, boss_id int32) (damage_list_msg []*msg_client_message.GuildStageDamageItem) {
	damage_list := guild_manager.GetStageDamageList(guild_id, boss_id)
	if damage_list == nil {
		return
	}
	length := damage_list.GetLength()
	if length > 0 {
		for r := int32(1); r <= length; r++ {
			k, v := damage_list.GetByRank(r)
			attacker_id := k.(int32)
			if attacker_id <= 0 {
				continue
			}
			name, level, head := GetPlayerBaseInfo(attacker_id)
			damage := v.(int32)
			damage_list_msg = append(damage_list_msg, &msg_client_message.GuildStageDamageItem{
				Rank:       r,
				MemberId:   attacker_id,
				MemberName: name,
				Level:      level,
				Head:       head,
				Damage:     damage,
			})
		}
	}
	return
}

// 初始化公会副本
func guild_stage_data_init(guild *dbGuildRow, boss_id int32) int32 {
	guild_stage := guild_boss_table_mgr.Get(boss_id)
	if guild_stage == nil {
		log.Error("guild stage %v not found", boss_id)
		return -1
	}
	stage_id := guild_boss_table_mgr.Array[0].StageId
	stage := stage_table_mgr.Get(stage_id)
	if stage == nil {
		log.Error("Stage %v table data not found", stage_id)
		return -1
	}
	if stage.Monsters == nil || len(stage.Monsters) == 0 {
		log.Error("Stage[%v] monster list is empty", stage_id)
		return -1
	}
	monster := stage.Monsters[0]
	if monster.Slot < 1 || monster.Slot > BATTLE_TEAM_MEMBER_MAX_NUM {
		log.Error("Stage[%v] monster[%v] pos %v invalid", stage_id, monster.MonsterID, monster.Slot)
		return -1
	}
	guild.Stage.SetBossId(boss_id)
	guild.Stage.SetBossPos(monster.Slot - 1)
	guild.Stage.SetHpPercent(100)
	return 1
}

// 公会副本数据
func (this *Player) send_guild_stage_data() int32 {
	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		log.Error("Player[%v] get guild failed or guild not found", this.Id)
		return -1
	}
	boss_id := guild.Stage.GetBossId()
	if boss_id == 0 {
		boss_id = guild_boss_table_mgr.Array[0].Id
		res := guild_stage_data_init(guild, boss_id)
		if res < 0 {
			return res
		}
	}

	response := &msg_client_message.S2CGuildStageDataResponse{
		CurrBossId:           boss_id,
		HpPercent:            guild.Stage.GetHpPercent(),
		RespawnNum:           this.db.GuildStage.GetRespawnNum(),
		RefreshRemainSeconds: utils.GetRemainSeconds2NextDayTime(guild.GetLastStageRefreshTime(), global_config.GuildStageRefreshTime),
		StageState:           this.db.GuildStage.GetRespawnState(),
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_STAGE_DATA_RESPONSE), response)
	log.Debug("Player[%v] send guild data %v", this.Id, response)
	return 1
}

// 公会副本排行榜
func (this *Player) guild_stage_rank_list(boss_id int32) int32 {
	guild_id := this.db.Guild.GetId()
	if guild_id <= 0 {
		log.Error("Player[%v] no joined one guild")
		return -1
	}
	damage_list := guild_stage_damage_list(guild_id, boss_id)
	response := &msg_client_message.S2CGuildStageRankListResponse{
		DmgList: damage_list,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_STAGE_RANK_LIST_RESPONSE), response)
	log.Debug("Player[%v] guild stage %v rank list %v", this.Id, boss_id, response)
	return 1
}

const (
	GUILD_STAGE_STATE_CAN_FIGHT    = iota
	GUILD_STAGE_STATE_WAIT_RESPAWN = 1
	GUILD_STAGE_STATE_CAN_RESPAWN  = 2
)

func (this *Player) guild_stage_fight(boss_id int32) int32 {
	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		log.Error("Player[%v] get guild failed or guild not found", this.Id)
		return -1
	}

	curr_boss_id := guild.Stage.GetBossId()
	if boss_id != curr_boss_id {
		// 返回排行榜
		return this.guild_stage_rank_list(boss_id)
	}

	stage_state := this.db.GuildStage.GetRespawnState()
	if stage_state == GUILD_STAGE_STATE_WAIT_RESPAWN {
		log.Error("Player[%v] waiting to respawn for guild stage", this.Id)
		return -1
	} else if stage_state == GUILD_STAGE_STATE_CAN_RESPAWN {
		log.Error("Player[%v] can respawn for guild stage", this.Id)
		return -1
	} else if stage_state != GUILD_STAGE_STATE_CAN_FIGHT {
		log.Error("Player[%v] guild stage stage %v invalid", this.Id, stage_state)
		return -1
	}

	guild_stage := guild_boss_table_mgr.Get(boss_id)
	if guild_stage == nil {
		log.Error("guild stage %v table data not found", boss_id)
		return -1
	}
	stage := stage_table_mgr.Get(guild_stage.StageId)
	if stage == nil {
		log.Error("stage %v table data not found", guild_stage.StageId)
		return -1
	}

	err, is_win, my_team, target_team, enter_reports, rounds, has_next_wave := this.FightInStage(9, stage, nil, guild)
	if err < 0 {
		log.Error("Player[%v] fight guild stage %v failed, team is empty", this.Id, boss_id)
		return err
	}

	if is_win {
		next_guild_stage := guild_boss_table_mgr.GetNext(boss_id)
		if next_guild_stage != nil {
			// 下一副本
			err := guild_stage_data_init(guild, next_guild_stage.Id)
			if err < 0 {
				log.Error("Player[%v] fight guild stage %v win, init next stage %v failed %v", this.Id, boss_id, next_guild_stage.Id, err)
				return err
			}
		}
	} else {
		// 状态置成等待复活
		stage_state = GUILD_STAGE_STATE_WAIT_RESPAWN
		this.db.GuildStage.SetRespawnState(stage_state)
	}

	member_damages := this.guild_stage_team.common_data.members_damage
	member_cures := this.guild_stage_team.common_data.members_cure
	response := &msg_client_message.S2CBattleResultResponse{
		IsWin:               is_win,
		EnterReports:        enter_reports,
		Rounds:              rounds,
		MyTeam:              my_team,
		TargetTeam:          target_team,
		MyMemberDamages:     member_damages[this.guild_stage_team.side],
		TargetMemberDamages: member_damages[this.target_stage_team.side],
		MyMemberCures:       member_cures[this.guild_stage_team.side],
		TargetMemberCures:   member_cures[this.target_stage_team.side],
		HasNextWave:         has_next_wave,
		BattleType:          9,
		BattleParam:         boss_id,
		ExtraValue:          stage_state,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_BATTLE_RESULT_RESPONSE), response)

	if is_win && !has_next_wave {
		// 关卡奖励
		this.send_stage_reward(stage, 7)
	}

	Output_S2CBattleResult(this, response)

	return 1
}

// 公会副本玩家复活
func (this *Player) guild_stage_player_respawn() int32 {
	return 1
}

// 公会副本重置
func (this *Player) guild_stage_reset() int32 {
	return 1
}

func C2SGuildStageDataHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildStageDataRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)", err.Error())
		return -1
	}
	return p.send_guild_stage_data()
}

func C2SGuildStageRankListHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildStageRankListRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%v)", err.Error())
		return -1
	}
	return p.guild_stage_rank_list(req.GetBossId())
}

func C2SGuildStagePlayerRespawnHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildStagePlayerRespawnRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%v)", err.Error())
		return -1
	}
	return p.guild_stage_player_respawn()
}

func C2SGuildStageResetHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildStageResetRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%v)", err.Error())
		return -1
	}
	return p.guild_stage_reset()
}
