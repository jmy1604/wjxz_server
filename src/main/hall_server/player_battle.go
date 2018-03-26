package main

import (
	"libs/log"
	"main/table_config"
	"math/rand"
	"net/http"
	"public_message/gen_go/client_message"

	"github.com/golang/protobuf/proto"
)

// 基础属性
const (
	ATTR_HP_MAX             = 1  // 最大血量
	ATTR_HP                 = 2  // 当前血量
	ATTR_MP                 = 3  // 气势
	ATTR_ATTACK             = 4  // 攻击
	ATTR_DEFENSE            = 5  // 防御
	ATTR_DODGE_COUNT        = 6  // 闪避次数
	ATTR_INJURED_MAX        = 7  // 受伤上限
	ATTR_SHIELD             = 8  // 护盾
	ATTR_CRITICAL           = 9  // 暴击率
	ATTR_CRITICAL_MULTI     = 10 // 暴击伤害倍率
	ATTR_ANTI_CRITICAL      = 11 // 抗暴率
	ATTR_BLOCK_RATE         = 12 // 格挡率
	ATTR_BLOCK_DEFENSE_RATE = 13 // 格挡减伤率
	ATTR_BREAK_BLOCK_RATE   = 14 // 破格率

	ATTR_TOTAL_DAMAGE_ADD      = 15 // 总增伤
	ATTR_CLOSE_DAMAGE_ADD      = 16 // 近战增伤
	ATTR_REMOTE_DAMAGE_ADD     = 17 // 远程增伤
	ATTR_NORMAL_DAMAGE_ADD     = 18 // 普攻增伤
	ATTR_RAGE_DAMAGE_ADD       = 19 // 怒气增伤
	ATTR_TOTAL_DAMAGE_SUB      = 20 // 总减伤
	ATTR_CLOSE_DAMAGE_SUB      = 21 // 近战减伤
	ATTR_REMOTE_DAMAGE_SUB     = 22 // 远程减伤
	ATTR_NORMAL_DAMAGE_SUB     = 23 // 普攻减伤
	ATTR_RAGE_DAMAGE_SUB       = 24 // 怒气减伤
	ATTR_CLOSE_VAMPIRE         = 25 // 近战吸血
	ATTR_REMOTE_VAMPIRE        = 26 // 远程吸血
	ATTR_CURE_RATE_CORRECT     = 27 // 治疗率修正
	ATTR_CURED_RATE_CORRECT    = 28 // 被治疗率修正
	ATTR_CLOSE_REFLECT         = 29 // 近战反击系数
	ATTR_REMOTE_REFLECT        = 30 // 远程反击系数
	ATTR_ARMOR_ADD             = 31 // 护甲增益
	ATTR_BREAK_ARMOR           = 32 // 破甲
	ATTR_POISON_INJURED_RATE   = 33 // 毒气受伤率
	ATTR_BURN_INJURED_RATE     = 34 // 点燃受伤率
	ATTR_BLEED_INJURED_RATE    = 35 // 流血受伤率
	ATTR_HP_PERCENT_BONUS      = 36 // 血量百分比
	ATTR_ATTACK_PERCENT_BONUS  = 37 // 攻击百分比
	ATTR_DEFENSE_PERCENT_BONUS = 38 // 防御百分比
	ATTR_DAMAGE_PERCENT_BONUS  = 39 // 伤害百分比
	ATTR_COUNT_MAX             = 50
)

// 战斗结束类型
const (
	BATTLE_END_BY_ALL_DEAD   = 1 // 一方全死
	BATTLE_END_BY_ROUND_OVER = 2 // 回合用完
)

// 最大回合数
const (
	BATTLE_ROUND_MAX_NUM = 30
)

const (
	BATTLE_TEAM_MEMBER_INIT_ENERGY       = 1 // 初始能量
	BATTLE_TEAM_MEMBER_MAX_ENERGY        = 4 // 最大能量
	BATTLE_TEAM_MEMBER_ADD_ENERGY        = 2 // 能量增加量
	BATTLE_TEAM_MEMBER_MAX_NUM           = 9 // 最大人数
	BATTLE_FORMATION_LINE_NUM            = 3 // 阵型列数
	BATTLE_FORMATION_ONE_LINE_MEMBER_NUM = 3 // 每列人数
)

// 阵容类型
const (
	BATTLE_ATTACK_TEAM  = 1
	BATTLE_DEFENSE_TEAM = 2
)

type MemberPassiveTriggerData struct {
	skill      *table_config.XmlSkillItem
	battle_num int32
	round_num  int32
}

type TeamMember struct {
	team             *BattleTeam
	id               int32
	level            int32
	card             *table_config.XmlCardItem
	hp               int32
	energy           int32
	attack           int32
	defense          int32
	act_num          int32 // 行动次数
	attrs            []int32
	bufflist_arr     []BuffList
	passive_triggers map[int32][]*MemberPassiveTriggerData
}

func (this *TeamMember) add_skill_attr(skill_id int32) {
	skill := skill_table_mgr.Get(skill_id)
	if skill == nil {
		return
	}
	for i := 0; i < len(skill.SkillAttr)/2; i++ {
		attr := skill.SkillAttr[2*i]
		if attr == ATTR_HP {
			this.add_hp(skill.SkillAttr[2*i+1])
		} else if attr == ATTR_HP_MAX {
			this.add_max_hp(skill.SkillAttr[2*i+1])
		} else {
			this.attrs[skill.SkillAttr[2*i]] = skill.SkillAttr[2*i+1]
		}
	}
}

func (this *TeamMember) init_passive_data(role_card *table_config.XmlCardItem) {
	if role_card.PassiveSkillIds == nil {
		return
	}
	for i := 0; i < len(role_card.PassiveSkillIds); i++ {
		skill := skill_table_mgr.Get(role_card.PassiveSkillIds[i])
		if skill == nil || skill.Type != SKILL_TYPE_PASSIVE {
			continue
		}

		if skill.TriggerBattleMax <= 0 || skill.TriggerRoundMax <= 0 {
			continue
		}

		if this.passive_triggers == nil {
			this.passive_triggers = make(map[int32][]*MemberPassiveTriggerData)
		}

		d := passive_trigger_data_pool.Get()
		d.skill = skill
		d.battle_num = skill.TriggerBattleMax
		d.round_num = skill.TriggerRoundMax
		datas := this.passive_triggers[skill.SkillTriggerType]
		if datas == nil {
			this.passive_triggers[skill.SkillTriggerType] = []*MemberPassiveTriggerData{d}
		} else {
			this.passive_triggers[skill.SkillTriggerType] = append(datas, d)
		}
	}
}

func (this *TeamMember) can_passive_trigger(trigger_event int32, skill_id int32) (trigger bool) {
	d, o := this.passive_triggers[trigger_event]
	if !o || d == nil {
		return
	}

	for i := 0; i < len(d); i++ {
		if d[i] == nil {
			continue
		}
		if d[i].skill.Id != skill_id {
			continue
		}
		if d[i].battle_num > 0 && d[i].round_num > 0 {
			trigger = true
		}
		break
	}

	return
}

func (this *TeamMember) used_passive_trigger_count(trigger_event int32, skill_id int32) {
	d, o := this.passive_triggers[trigger_event]
	if !o || d == nil {
		return
	}

	for i := 0; i < len(d); i++ {
		if d[i] == nil {
			continue
		}
		if d[i].skill.Id != skill_id {
			continue
		}
		if d[i].battle_num > 0 {
			d[i].battle_num -= 1
		}
		if d[i].round_num > 0 {
			d[i].round_num -= 1
		}
		if d[i].battle_num <= 0 || d[i].round_num <= 0 {
			passive_trigger_data_pool.Put(d[i])
		}
		break
	}
}

func (this *TeamMember) init(team *BattleTeam, id int32, level int32, role_card *table_config.XmlCardItem) {
	if this.attrs == nil {
		this.attrs = make([]int32, ATTR_COUNT_MAX)
	} else {
		for i := 0; i < len(this.attrs); i++ {
			this.attrs[i] = 0
		}
	}

	if this.bufflist_arr != nil {
		for i := 0; i < len(this.bufflist_arr); i++ {
			this.bufflist_arr[i].clear()
		}
	}

	this.team = team
	this.id = id
	this.level = level
	this.card = role_card
	this.hp = (role_card.BaseHP + (level-1)*role_card.GrowthHP/100) * (10000 + this.attrs[ATTR_HP_PERCENT_BONUS]) / 10000
	this.attack = (role_card.BaseAttack + (level-1)*role_card.GrowthAttack/100) * (10000 + this.attrs[ATTR_ATTACK_PERCENT_BONUS]) / 10000
	this.defense = (role_card.BaseDefence + (level-1)*role_card.GrowthDefence/100) * (10000 + this.attrs[ATTR_DEFENSE_PERCENT_BONUS]) / 10000
	this.energy = BATTLE_TEAM_MEMBER_INIT_ENERGY

	this.attrs[ATTR_HP_MAX] = this.hp
	this.attrs[ATTR_HP] = this.hp
	this.attrs[ATTR_ATTACK] = this.attack
	this.attrs[ATTR_DEFENSE] = this.defense

	// 技能增加属性
	if role_card.NormalSkillID > 0 {
		this.add_skill_attr(role_card.NormalSkillID)
	}
	if role_card.SuperSkillID > 0 {
		this.add_skill_attr(role_card.SuperSkillID)
	}
	for i := 0; i < len(role_card.PassiveSkillIds); i++ {
		this.add_skill_attr(role_card.PassiveSkillIds[i])
	}

	this.init_passive_data(role_card)
}

func (this *TeamMember) add_hp(hp int32) {
	if hp > 0 {
		if this.attrs[ATTR_HP]+hp > this.attrs[ATTR_HP_MAX] {
			this.attrs[ATTR_HP] = this.attrs[ATTR_HP_MAX]
		} else {
			this.attrs[ATTR_HP] += hp
		}
	} else if hp < 0 {
		if this.attrs[ATTR_HP]+hp < 0 {
			this.attrs[ATTR_HP] = 0
		} else {
			this.attrs[ATTR_HP] += hp
		}
	}
	this.hp = this.attrs[ATTR_HP]
}

func (this *TeamMember) add_max_hp(add int32) {
	if add < 0 {
		if this.attrs[ATTR_HP_MAX]+add < this.attrs[ATTR_HP] {
			this.attrs[ATTR_HP] = this.attrs[ATTR_HP_MAX] + add
		}
	}
	this.attrs[ATTR_HP_MAX] += add
}

func (this *TeamMember) round_start() {
	this.act_num += 1
	this.energy += BATTLE_TEAM_MEMBER_ADD_ENERGY
}

func (this *TeamMember) round_end() {
	for i := 0; i < len(this.bufflist_arr); i++ {
		buffs := this.bufflist_arr[i]
		buffs.on_round_end()
	}

	for _, v := range this.passive_triggers {
		if v == nil {
			continue
		}
		for i := 0; i < len(v); i++ {
			if v[i].skill.TriggerRoundMax > 0 {
				v[i].round_num = v[i].skill.TriggerRoundMax
			}
		}
	}
}

func (this *TeamMember) get_use_skill() (skill_id int32) {
	// 能量满用绝杀
	if this.energy >= BATTLE_TEAM_MEMBER_MAX_ENERGY {
		skill_id = this.card.SuperSkillID
	} else if this.act_num > 0 {
		skill_id = this.card.NormalSkillID
	}
	return
}

func (this *TeamMember) used_skill() {
	if this.energy >= BATTLE_TEAM_MEMBER_MAX_ENERGY {
		this.energy -= BATTLE_TEAM_MEMBER_MAX_ENERGY
	}
	if this.act_num > 0 {
		this.act_num -= 1
	}
}

func (this *TeamMember) add_buff(attacker *TeamMember, skill_effect []int32) (buff_id int32) {
	b := buff_table_mgr.Get(skill_effect[1])
	if b == nil {
		return
	}

	if this.bufflist_arr == nil {
		this.bufflist_arr = make([]BuffList, BUFF_EFFECT_TYPE_COUNT)
		for i := 0; i < BUFF_EFFECT_TYPE_COUNT; i++ {
			this.bufflist_arr[i].owner = this
		}
	}

	// 互斥
	for i := 0; i < len(this.bufflist_arr); i++ {
		h := &this.bufflist_arr[i]
		if h.check_buff_mutex(b) {
			return
		}
	}

	if rand.Int31n(10000) >= skill_effect[2] {
		return
	}

	return this.bufflist_arr[b.Effect[0]].add_buff(attacker, b, skill_effect)
}

func (this *TeamMember) has_buff(buff_id int32) bool {
	if this.bufflist_arr != nil {
		for i := 0; i < len(this.bufflist_arr); i++ {
			bufflist := &this.bufflist_arr[i]
			buff := bufflist.head
			for buff != nil {
				if buff.buff.Id == buff_id {
					return true
				}
			}
		}
	}
	return false
}

func (this *TeamMember) remove_buff_effect(buff *Buff) {
	if buff.buff.Effect[0] == BUFF_EFFECT_TYPE_MODIFY_ATTR {
		if this.attrs[buff.buff.Effect[1]] < buff.param {
			this.attrs[buff.buff.Effect[1]] = 0
		} else {
			this.attrs[buff.buff.Effect[1]] -= buff.param
		}

		if buff.buff.Effect[1] == ATTR_HP_MAX && this.attrs[ATTR_HP] > this.attrs[ATTR_HP_MAX] {
			this.attrs[ATTR_HP] = this.attrs[ATTR_HP_MAX]
		}
	}
}

func (this *TeamMember) is_disable_normal_attack() bool {
	if this.bufflist_arr == nil {
		return false
	}
	disable := false
	bufflist := &this.bufflist_arr[BUFF_EFFECT_TYPE_DISABLE_NORMAL_ATTACK]
	if bufflist.head != nil {
		disable = true
	} else {
		bufflist = &this.bufflist_arr[BUFF_EFFECT_TYPE_DISABLE_ACTION]
		if bufflist.head != nil {
			disable = true
		}
	}
	return disable
}

func (this *TeamMember) is_disable_super_attack() bool {
	if this.bufflist_arr == nil {
		return false
	}
	disable := false
	bufflist := &this.bufflist_arr[BUFF_EFFECT_TYPE_DISABLE_SUPER_ATTACK]
	if bufflist.head != nil {
		disable = true
	} else {
		bufflist = &this.bufflist_arr[BUFF_EFFECT_TYPE_DISABLE_ACTION]
		if bufflist.head != nil {
			disable = true
		}
	}
	return disable
}

func (this *TeamMember) is_disable_attack() bool {
	if this.bufflist_arr == nil {
		return false
	}
	disable := false
	bufflist := &this.bufflist_arr[BUFF_EFFECT_TYPE_DISABLE_ACTION]
	if bufflist.head != nil {
		disable = true
	} else {
		bufflist = &this.bufflist_arr[BUFF_EFFECT_TYPE_DISABLE_NORMAL_ATTACK]
		if bufflist.head != nil {
			disable = true
		} else {
			bufflist = &this.bufflist_arr[BUFF_EFFECT_TYPE_DISABLE_SUPER_ATTACK]
			if bufflist.head != nil {
				disable = true
			}
		}
	}
	return disable
}

func (this *TeamMember) is_dead() bool {
	if this.hp > 0 {
		return false
	}
	return true
}

type BattleTeam struct {
	curr_attack int32 // 当前进攻的索引
	side        int32 // 0 左边 1 右边
	members     []*TeamMember
}

// 利用玩家初始化
func (this *BattleTeam) Init(p *Player, team_id int32, side int32) bool {
	var members []int32
	if team_id == BATTLE_ATTACK_TEAM {
		members = p.db.BattleTeam.GetAttackMembers()
	} else if team_id == BATTLE_DEFENSE_TEAM {
		members = p.db.BattleTeam.GetDefenseMembers()
	} else {
		log.Warn("Unknown team id %v", team_id)
		return false
	}

	if this.members == nil {
		this.members = make([]*TeamMember, BATTLE_TEAM_MEMBER_MAX_NUM)
	}

	for i := 0; i < len(members); i++ {
		if members[i] <= 0 {
			this.members[i] = nil
			continue
		}

		//if this.members[i] != nil && members[i] == this.members[i].id {
		//	continue
		//s}

		var table_id, rank, level int32
		var o bool
		table_id, o = p.db.Roles.GetTableId(members[i])
		if !o {
			log.Error("Cant get table id by battle team member id[%v]", members[i])
			return false
		}
		rank, o = p.db.Roles.GetRank(members[i])
		if !o {
			log.Error("Cant get rank by battle team member id[%v]", members[i])
			return false
		}
		level, o = p.db.Roles.GetLevel(members[i])
		if !o {
			log.Error("Cant get level by battle team member id[%v]", members[i])
			return false
		}
		role_card := card_table_mgr.GetRankCard(table_id, rank)
		if role_card == nil {
			log.Error("Cant get card by role_id[%v] and rank[%v]", table_id, rank)
			return false
		}

		m := p.team_member_mgr[members[i]]
		if m == nil {
			m = team_member_pool.Get()
		}
		m.init(this, members[i], level, role_card)
		this.members[i] = m

		// 装备BUFF增加属性
		log.Debug("mem[%v]: id[%v] role_id[%v] role_rank[%v] hp[%v] energy[%v] attack[%v] defense[%v]", i, m.id, m.card.Id, m.card.Rank, m.hp, m.energy, m.attack, m.defense)
	}
	this.curr_attack = 0
	this.side = side

	p.team_changed[team_id] = false

	return true
}

// round start
func (this *BattleTeam) RoundStart() {
	for i := 0; i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
		if this.members[i] != nil {
			this.members[i].round_start()
		}
	}
	this.curr_attack = 0
}

// round end
func (this *BattleTeam) RoundEnd() {
	for i := 0; i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
		if this.members[i] != nil {
			this.members[i].round_end()
		}
	}
}

// find targets
func (this *BattleTeam) FindTargets(self_index int32, target_team *BattleTeam, trigger_skill int32) (is_enemy bool, pos []int32, skill *table_config.XmlSkillItem) {
	skill_id := int32(0)
	m := this.members[self_index]

	if trigger_skill == 0 {
		// 能量满用绝杀
		if m.energy >= BATTLE_TEAM_MEMBER_MAX_ENERGY {
			if m.is_disable_super_attack() {
				return
			}
			skill_id = m.card.SuperSkillID
		} else if m.act_num > 0 {
			if m.is_disable_normal_attack() {
				return
			}
			skill_id = m.card.NormalSkillID
		}
	} else {
		skill_id = trigger_skill
	}

	if m.is_disable_attack() {
		return
	}

	skill = skill_table_mgr.Get(skill_id)
	if skill == nil {
		log.Error("Cant get skill by id[%v]", skill_id)
		return
	}

	if skill.Type == SKILL_TYPE_NORMAL {

	} else if skill.Type == SKILL_TYPE_SUPER {

	} else if skill.Type == SKILL_TYPE_PASSIVE {
		// 被动触发
	} else if skill.Type == SKILL_TYPE_NEXT {

	} else {
		log.Error("Invalid skill type[%v]", skill.Type)
		return
	}

	if skill.SkillEnemy == SKILL_ENEMY_TYPE_ENEMY {
		is_enemy = true
	} else if skill.SkillEnemy == SKILL_ENEMY_TYPE_OUR {
		target_team = this
	} else {
		log.Error("Invalid skill enemy type[%v]", skill.SkillEnemy)
		return
	}

	if skill.SkillTarget == SKILL_TARGET_TYPE_DEFAULT {
		pos = skill_get_default_targets(self_index, target_team, skill)
	} else if skill.SkillTarget == SKILL_TARGET_TYPE_BACK {
		pos = skill_get_back_targets(self_index, target_team, skill)
	} else if skill.SkillTarget == SKILL_TARGET_TYPE_HP_MIN {
		pos = skill_get_hp_min_targets(self_index, target_team, skill)
	} else if skill.SkillTarget == SKILL_TARGET_TYPE_RANDOM {
		pos = skill_get_random_targets(self_index, target_team, skill)
	} else if skill.SkillTarget == SKILL_TARGET_TYPE_SELF {
		pos = []int32{self_index}
	} else if skill.SkillTarget == SKILL_TARGET_TYPE_TRIGGER_OBJECT {

	} else if skill.SkillTarget == SKILL_TARGET_TYPE_CROPSE {

	} else {
		log.Error("Invalid skill target type: %v", skill.SkillTarget)
		return
	}

	return
}

func (this *BattleTeam) UseOnceSkill(self_index int32, target_team *BattleTeam, trigger_skill int32) (skill *table_config.XmlSkillItem) {
	is_enemy, target_pos, skill := this.FindTargets(self_index, target_team, 0)
	if target_pos == nil {
		log.Warn("Self index[%v] Cant find targets to attack with skill[%v]", self_index, skill.Id)
		return nil
	}
	log.Debug("team member[%v] find is_enemy[%v] targets[%v] to use skill[%v]", self_index, is_enemy, target_pos, skill.Id)
	if !is_enemy {
		target_team = this
	}
	skill_effect(this, self_index, target_team, target_pos, skill)

	return skill
}

func (this *BattleTeam) UseSkill(self_index int32, target_team *BattleTeam) int32 {
	mem := this.members[self_index]
	if mem == nil {
		return -1
	}
	for mem.get_use_skill() > 0 {
		if target_team.IsAllDead() {
			return 0
		}
		skill := this.UseOnceSkill(self_index, target_team, 0)
		if skill == nil {
			break
		}
		if skill.ComboSkill > 0 {
			this.UseOnceSkill(self_index, target_team, skill.ComboSkill)
		}
		mem.used_skill()
	}
	return 1
}

// 回合
func (this *BattleTeam) DoRound(target_team *BattleTeam) {
	this.RoundStart()
	target_team.RoundStart()

	// 被动技，回合行动前触发
	for i := int32(0); i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
		passive_skill_effect(EVENT_BEFORE_ROUND, this, i, nil, target_team)
		passive_skill_effect(EVENT_BEFORE_ROUND, target_team, i, nil, this)
	}

	var self_index, target_index int32
	for self_index < BATTLE_TEAM_MEMBER_MAX_NUM && target_index < BATTLE_TEAM_MEMBER_MAX_NUM {
		for ; self_index < BATTLE_TEAM_MEMBER_MAX_NUM; self_index++ {
			if this.UseSkill(self_index, target_team) >= 0 {
				self_index += 1
				break
			}
		}
		for ; target_index < BATTLE_TEAM_MEMBER_MAX_NUM; target_index++ {
			if target_team.UseSkill(target_index, this) >= 0 {
				target_index += 1
				break
			}
		}
	}

	this.RoundEnd()
	target_team.RoundEnd()
}

// 开打
func (this *BattleTeam) Fight(target_team *BattleTeam, end_type int32, end_param int32) {
	round_max := end_param
	if end_type == BATTLE_END_BY_ALL_DEAD {
		round_max = BATTLE_ROUND_MAX_NUM
	} else if end_type == BATTLE_END_BY_ROUND_OVER {
	}

	// 被动技，进场前触发
	for i := int32(0); i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
		passive_skill_effect(EVENT_ENTER_BATTLE, this, i, nil, target_team)
		passive_skill_effect(EVENT_ENTER_BATTLE, target_team, i, nil, this)
	}

	for c := int32(0); c < round_max; c++ {
		log.Debug("------------------------------------ Round[%v] ----------------------------------", c+1)
		this.DoRound(target_team)
		if this.IsAllDead() {
			log.Debug("self all dead")
			break
		}
		if target_team.IsAllDead() {
			log.Debug("target all dead")
			break
		}
	}
}

// 是否全挂
func (this *BattleTeam) IsAllDead() bool {
	all_dead := true
	for i := 0; i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
		if this.members[i] == nil {
			continue
		}
		if !this.members[i].is_dead() {
			all_dead = false
			break
		}
	}
	return all_dead
}

// 是否有某个角色
func (this *BattleTeam) HasRole(role_id int32) bool {
	for i := 0; i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
		if this.members[i] == nil {
			continue
		}
		if this.members[i].card.Id == role_id {
			return true
		}
	}
	return false
}

func C2SFightHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SBattleResultRequest)
	if req == nil || p == nil {
		log.Error("C2SWorldChatMsgPullHandler player[%v] proto is invalid", p.Id)
		return -1
	}

	if !p.SetAttackTeam(req.AttackMembers) {
		log.Error("Player[%v] set attack members[%v] failed", p.Id, req.AttackMembers)
		return int32(msg_client_message.E_ERR_PLAYER_SET_ATTACK_MEMBERS_FAILED)
	}

	p.Fight2Player(req.FightPlayerId)

	return 1
}
