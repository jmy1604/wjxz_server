package main

import (
	"libs/log"
	"main/table_config"
	_ "math/rand"
	_ "public_message/gen_go/client_message"
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

type Buff struct {
	buff            *table_config.XmlStatusItem
	attack          int32
	dmg_add         int32
	skill_dmg_coeff int32
	param           int32
	round_num       int32
	next            *Buff
	prev            *Buff
}

type BuffList struct {
	head *Buff
	tail *Buff
}

type TeamMember struct {
	id           int32
	card         *table_config.XmlCardItem
	hp           int32
	energy       int32
	attack       int32
	defense      int32
	act_num      int32 // 行动次数
	attrs        []int32
	bufflist_arr []BuffList
}

func (this *TeamMember) init(id int32, level int32, role_card *table_config.XmlCardItem) {
	if this.attrs == nil {
		this.attrs = make([]int32, ATTR_COUNT_MAX)
	}

	this.id = id
	this.card = role_card
	this.hp = (role_card.BaseHP + (level-1)*role_card.GrowthHP/100) * (10000 + this.attrs[ATTR_HP_PERCENT_BONUS]) / 10000
	this.attack = (role_card.BaseAttack + (level-1)*role_card.GrowthAttack/100) * (10000 + this.attrs[ATTR_ATTACK_PERCENT_BONUS]) / 10000
	this.defense = (role_card.BaseDefence + (level-1)*role_card.GrowthDefence/100) * (10000 + this.attrs[ATTR_DEFENSE_PERCENT_BONUS]) / 10000
	this.energy = BATTLE_TEAM_MEMBER_INIT_ENERGY

	this.attrs[ATTR_HP_MAX] = this.hp
	this.attrs[ATTR_HP] = this.hp
	this.attrs[ATTR_ATTACK] = this.attack
	this.attrs[ATTR_DEFENSE] = this.defense
}

func (this *TeamMember) round_start() {
	this.act_num += 1
	this.energy += BATTLE_TEAM_MEMBER_ADD_ENERGY
}

func (this *TeamMember) round_end() {
	for i := 0; i < len(this.bufflist_arr); i++ {
		buffs := this.bufflist_arr[i]
		bf := buffs.head
		for bf != nil {
			if bf.buff.Effect[0] == BUFF_EFFECT_TYPE_DAMAGE {
				dmg := buff_effect_damage(bf.attack, bf.dmg_add, 0, bf.buff.Effect[1], this)
				this.attrs[ATTR_HP] -= dmg
				if this.attrs[ATTR_HP] < 0 {
					this.attrs[ATTR_HP] = 0
				}
			}
			bf.round_num -= 1
			next := bf.next
			if bf.round_num <= 0 {
				if bf.prev != nil {
					bf.prev.next = bf.next
				}
				if bf.next != nil {
					bf.next.prev = bf.prev
				}
				buff_pool.Put(bf)
			}
			bf = next
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
	}

	// 互斥
	for i := 0; i < len(this.bufflist_arr); i++ {
		h := this.bufflist_arr[i]
		hh := h.head
		for hh != nil {
			for j := 0; j < len(hh.buff.ResistMutexTypes); j++ {
				if b.MutexType == hh.buff.ResistMutexTypes[j] {
					log.Debug("BUFF[%v]类型[%v]排斥BUFF[%v]类型[%v]", hh.buff.Id, hh.buff.MutexType, b.Id, b.MutexType)
					return
				}
			}
			for j := 0; j < len(hh.buff.ResistMutexIDs); j++ {
				if b.Id == hh.buff.ResistMutexIDs[j] {
					log.Debug("BUFF[%v]排斥BUFF[%v]", hh.buff.Id, b.Id)
					return
				}
			}
			for j := 0; j < len(hh.buff.CancelMutexTypes); j++ {
				if b.MutexType == hh.buff.CancelMutexTypes[j] {
					if hh.prev != nil {
						hh.prev.next = hh.next
					}
					if hh.next != nil {
						hh.next.prev = hh.prev
					}
					buff_pool.Put(hh)
					log.Debug("BUFF[%v]类型[%v]驱散了BUFF[%v]类型[%v]", b.Id, b.MutexType, hh.buff.Id, hh.buff.MutexType)
				}
			}
			for j := 0; j < len(hh.buff.CancelMutexIDs); j++ {
				if b.Id == hh.buff.CancelMutexIDs[i] {
					if hh.prev != nil {
						hh.prev.next = hh.next
					}
					if hh.next != nil {
						hh.next.prev = hh.prev
					}
					buff_pool.Put(hh)
					log.Debug("BUFF[%v]驱散了BUFF[%v]", b.Id, hh.buff.Id)
				}
			}
			hh = hh.next
		}
	}

	buffs := &this.bufflist_arr[b.Effect[0]]
	buff := buff_pool.Get()
	buff.buff = b
	buff.attack = attacker.attrs[ATTR_ATTACK]
	buff.dmg_add = attacker.attrs[ATTR_TOTAL_DAMAGE_ADD]
	buff.param = skill_effect[3]
	buff.round_num = skill_effect[4]
	if buffs.head == nil {
		buffs.tail = buff
		buffs.head = buff
	} else {
		buff.prev = buffs.tail
		buffs.tail.next = buff
	}

	return b.Id
}

func (this *TeamMember) remove_buff_effect(buff *Buff) {
	if buff.buff.Effect[0] == BUFF_EFFECT_TYPE_MODIFY_ATTR {
		this.attrs[buff.buff.Effect[1]] -= buff.param
	}
}

type BattleTeam struct {
	curr_attack int32 // 当前进攻的索引
	members     []*TeamMember
}

// 利用玩家初始化
func (this *BattleTeam) Init(p *Player, team_id int32) bool {
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

		if this.members[i] != nil && members[i] == this.members[i].id {
			continue
		}

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
		m.init(members[i], level, role_card)
		this.members[i] = m

		// 装备BUFF增加属性
		log.Debug("mem[%v]: id[%v] role_id[%v] role_rank[%v] hp[%v] energy[%v] attack[%v] defense[%v]", i, m.id, m.card.Id, m.card.Rank, m.hp, m.energy, m.attack, m.defense)
	}
	this.curr_attack = 0

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
func (this *BattleTeam) FindTargets(self_index int32, team *BattleTeam) (is_enemy bool, pos []int32, skill *table_config.XmlSkillItem) {
	skill_id := int32(0)
	m := this.members[self_index]

	// 能量满用绝杀
	if m.energy >= BATTLE_TEAM_MEMBER_MAX_ENERGY {
		skill_id = m.card.SuperSkillID
	} else if m.act_num > 0 {
		skill_id = m.card.NormalSkillID
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
		team = this
	} else {
		log.Error("Invalid skill enemy type[%v]", skill.SkillEnemy)
		return
	}

	if skill.SkillTarget == SKILL_TARGET_TYPE_DEFAULT {
		pos = skill_get_default_targets(self_index, team, skill)
	} else if skill.SkillTarget == SKILL_TARGET_TYPE_BACK {
		pos = skill_get_back_targets(self_index, team, skill)
	} else if skill.SkillTarget == SKILL_TARGET_TYPE_HP_MIN {
		pos = skill_get_hp_min_targets(self_index, team, skill)
	} else if skill.SkillTarget == SKILL_TARGET_TYPE_RANDOM {
		pos = skill_get_random_targets(self_index, team, skill)
	} else if skill.SkillTarget == SKILL_TARGET_TYPE_SELF {
		pos = []int32{self_index}
	} else {
		log.Error("Invalid skill target type: %v", skill.SkillTarget)
		return
	}

	return
}

// 回合
func (this *BattleTeam) DoRound(target_team *BattleTeam) {
	this.RoundStart()
	target_team.RoundStart()

	var self_index, target_index int32
	var mem *TeamMember
	used_skill := false
	for {
		for used_skill = false; self_index < BATTLE_TEAM_MEMBER_MAX_NUM; self_index++ {
			if used_skill {
				break
			}
			mem = this.members[self_index]
			if mem == nil {
				continue
			}
			for mem.get_use_skill() > 0 {
				is_enemy, target_pos, skill := this.FindTargets(self_index, target_team)
				if target_pos == nil {
					log.Warn("Cant find targets to attack")
					return
				}
				log.Debug("team member[%v] find is_enemy[%v] targets[%v] to use skill[%v]", mem.id, is_enemy, target_pos, skill.Id)
				if is_enemy {
					skill_effect(this, self_index, target_team, target_pos, skill)
				} else {
					skill_effect(this, self_index, this, target_pos, skill)
				}
				mem.used_skill()
				used_skill = true
			}
		}

		for used_skill = false; target_index < BATTLE_TEAM_MEMBER_MAX_NUM; target_index++ {
			if used_skill {
				break
			}
			mem = target_team.members[target_index]
			if mem == nil || mem.get_use_skill() == 0 {
				continue
			}
			for mem.get_use_skill() > 0 {
				is_enemy, target_pos, skill := target_team.FindTargets(target_index, this)
				if target_pos == nil {
					log.Warn("Cant find targets to attack")
					return
				}
				log.Debug("target team member[%v] find is_enemy[%v] targets[%v] to use skill[%v]", mem.id, is_enemy, target_pos, skill.Id)
				if is_enemy {
					skill_effect(target_team, target_index, this, target_pos, skill)
				} else {
					skill_effect(target_team, target_index, target_team, target_pos, skill)
				}
				mem.used_skill()
				used_skill = true
			}
		}

		if self_index >= BATTLE_TEAM_MEMBER_MAX_NUM && target_index >= BATTLE_TEAM_MEMBER_MAX_NUM {
			break
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

	for c := int32(0); c < round_max; c++ {
		log.Debug("-------------------- Round[%v] --------------------", c+1)
		this.DoRound(target_team)
		if this.IsAllDead() || target_team.IsAllDead() {
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
		if this.members[i].attrs[ATTR_HP] > 0 {
			all_dead = false
			break
		}
	}
	return all_dead
}
