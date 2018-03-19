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

	ATTR_TOTAL_DAMAGE_ADD    = 15 // 总增伤
	ATTR_CLOSE_DAMAGE_ADD    = 16 // 近战增伤
	ATTR_REMOTE_DAMAGE_ADD   = 17 // 远程增伤
	ATTR_NORMAL_DAMAGE_ADD   = 18 // 普攻增伤
	ATTR_RAGE_DAMAGE_ADD     = 19 // 怒气增伤
	ATTR_TOTAL_DAMAGE_SUB    = 20 // 总减伤
	ATTR_CLOSE_DAMAGE_SUB    = 21 // 近战减伤
	ATTR_REMOTE_DAMAGE_SUB   = 22 // 远程减伤
	ATTR_NORMAL_DAMAGE_SUB   = 23 // 普攻减伤
	ATTR_RAGE_DAMAGE_SUB     = 24 // 怒气减伤
	ATTR_CLOSE_VAMPIRE       = 25 // 近战吸血
	ATTR_REMOTE_VAMPIRE      = 26 // 远程吸血
	ATTR_CURE_RATE_CORRECT   = 27 // 治疗率修正
	ATTR_CURED_RATE_CORRECT  = 28 // 被治疗率修正
	ATTR_CLOSE_REFLECT       = 29 // 近战反击系数
	ATTR_REMOTE_REFLECT      = 30 // 远程反击系数
	ATTR_ARMOR_ADD           = 31 // 护甲增益
	ATTR_BREAK_ARMOR         = 32 // 破甲
	ATTR_POISON_INJURED_RATE = 33 // 毒气受伤率
	ATTR_BURN_INJURED_RATE   = 34 // 点燃受伤率
	ATTR_BLEED_INJURED_RATE  = 35 // 流血受伤率
	ATTR_COUNT_MAX           = 40
)

const (
	BATTLE_TEAM_MEMBER_INIT_ENERGY       = 1 // 初始能量
	BATTLE_TEAM_MEMBER_MAX_ENERGY        = 4 // 最大能量
	BATTLE_TEAM_MEMBER_ADD_ENERGY        = 2 // 能量增加量
	BATTLE_TEAM_MEMBER_MAX_NUM           = 9 // 最大人数
	BATTLE_FORMATION_LINE_NUM            = 3 // 阵型列数
	BATTLE_FORMATION_ONE_LINE_MEMBER_NUM = 3 // 每列人数
)

type TeamMember struct {
	id      int32
	card    *table_config.XmlCardItem
	hp      int32
	energy  int32
	attack  int32
	defense int32
	attrs   []int32
}

type BattleTeam struct {
	curr_attack int32 // 当前进攻的索引
	members     []*TeamMember
}

type ReportItem struct {
	attacker    *TeamMember
	be_attacker *TeamMember
	skill_id    int32
	damage      int32
	next        *ReportItem
}

// 利用玩家初始化
func (this *BattleTeam) Init(p *Player, is_attack bool) {
	var members []int32
	if is_attack {
		members = p.db.BattleTeam.GetAttackMembers()
	} else {
		members = p.db.BattleTeam.GetDefenseMembers()
	}
	if members == nil {
		return
	}

	this.members = make([]*TeamMember, BATTLE_TEAM_MEMBER_MAX_NUM)
	for i := 0; i < len(members); i++ {
		if members[i] <= 0 {
			continue
		}

		var table_id, rank, level int32
		var o bool
		table_id, o = p.db.Roles.GetTableId(members[i])
		if !o {
			log.Error("Cant get table id by battle team member id[%v]", members[i])
			return
		}
		rank, o = p.db.Roles.GetRank(members[i])
		if !o {
			log.Error("Cant get rank by battle team member id[%v]", members[i])
			return
		}
		level, o = p.db.Roles.GetLevel(members[i])
		if !o {
			log.Error("Cant get level by battle team member id[%v]", members[i])
			return
		}
		role_card := card_table_mgr.GetRankCard(table_id, rank)
		if role_card == nil {
			log.Error("Cant get card by role_id[%v] and rank[%v]", table_id, rank)
			return
		}

		m := &TeamMember{}
		m.id = members[i]
		m.card = role_card
		m.hp = role_card.BaseHP + (level-1)*role_card.GrowthHP/100
		m.attack = role_card.BaseAttack + (level-1)*role_card.GrowthAttack/100
		m.defense = role_card.BaseDefence + (level-1)*role_card.GrowthDefence/100
		m.energy = BATTLE_TEAM_MEMBER_INIT_ENERGY

		m.attrs = make([]int32, ATTR_COUNT_MAX)
		m.attrs[ATTR_HP_MAX] = m.hp
		m.attrs[ATTR_HP] = m.hp
		m.attrs[ATTR_ATTACK] = m.attack
		m.attrs[ATTR_DEFENSE] = m.defense

		// 装备BUFF增加属性

		this.members[i] = m
	}
	this.curr_attack = 0
}

// find targets
func (this *BattleTeam) FindTargets(self_index int32, team *BattleTeam) (is_enemy bool, pos []int32, skill *table_config.XmlSkillItem) {
	skill_id := int32(0)
	m := this.members[self_index]

	// 能量满用绝杀
	if m.energy >= BATTLE_TEAM_MEMBER_MAX_ENERGY {
		skill_id = m.card.SuperSkillID
	} else {
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
		//
	} else if skill.SkillEnemy == SKILL_ENEMY_TYPE_OUR {
		team = this
		is_enemy = true
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
		pos = []int32{this.curr_attack}
	} else {
		log.Error("Invalid skill target type: %v", skill.SkillTarget)
		return
	}

	return
}

// do action
func (this *BattleTeam) DoAction2Targets(target_team *BattleTeam, target_pos []int32, skill *table_config.XmlSkillItem) {
	for i := 0; i < len(target_pos); i++ {
		skill_effect(this, this.curr_attack, target_team, target_pos, skill)
		this.members[this.curr_attack].energy += BATTLE_TEAM_MEMBER_ADD_ENERGY
	}
}

// attack
func (this *BattleTeam) Attack(target_team *BattleTeam) {
	for i := 0; i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
		if this.members[i] == nil {
			continue
		}
		is_enemy, target_pos, skill := this.FindTargets(int32(i), target_team)
		if target_pos == nil {
			log.Warn("Cant find targets to attack")
			return
		}

		if !is_enemy {
			target_team = this
		}

		this.DoAction2Targets(target_team, target_pos, skill)
	}
}
