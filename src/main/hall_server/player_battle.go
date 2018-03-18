package main

import (
	"libs/log"
	"main/table_config"
	_ "math/rand"
	_ "public_message/gen_go/client_message"
)

// 基础属性
const (
	ATTR_HP_MAX              = 1  // 最大血量
	ATTR_HP                  = 2  // 当前血量
	ATTR_MP                  = 3  // 气势
	ATTR_ATTACK              = 4  // 攻击
	ATTR_DEFENSE             = 5  // 防御
	ATTR_CRITICAL            = 6  // 暴击率
	ATTR_CRITICAL_MULTI      = 7  // 暴击伤害倍率
	ATTR_ANTI_CRITICAL       = 8  // 抗暴率
	ATTR_BLOCK_RATE          = 9  // 格挡率
	ATTR_BLOCK_DEFENSE_RATE  = 10 // 格挡减伤率
	ATTR_BREAK_BLOCK_RATE    = 11 // 破格率
	ATTR_SHIELD              = 12 // 护盾
	ATTR_TOTAL_DAMAGE_ADD    = 13 // 总增伤
	ATTR_CLOSE_DAMAGE_ADD    = 14 // 近战增伤
	ATTR_REMOTE_DAMAGE_ADD   = 15 // 远程增伤
	ATTR_NORMAL_DAMAGE_ADD   = 16 // 普攻增伤
	ATTR_RAGE_DAMAGE_ADD     = 17 // 怒气增伤
	ATTR_TOTAL_DAMAGE_SUB    = 18 // 总减伤
	ATTR_CLOSE_DAMAGE_SUB    = 19 // 近战减伤
	ATTR_REMOTE_DAMAGE_SUB   = 20 // 远程减伤
	ATTR_NORMAL_DAMAGE_SUB   = 21 // 普攻减伤
	ATTR_RAGE_DAMAGE_SUB     = 22 // 怒气减伤
	ATTR_CLOSE_VAMPIRE       = 23 // 近战吸血
	ATTR_REMOTE_VAMPIRE      = 24 // 远程吸血
	ATTR_CURE_RATE_CORRECT   = 25 // 治疗率修正
	ATTR_CURED_RATE_CORRECT  = 26 // 被治疗率修正
	ATTR_CLOSE_COUNTER       = 27 // 近战反击系数
	ATTR_REMOTE_COUNTER      = 28 // 远程反击系数
	ATTR_DODGE_COUNT         = 29 // 闪避次数
	ATTR_INJURED_MAX         = 30 // 受伤上限
	ATTR_POISON_INJURED_RATE = 31 // 毒气受伤率
	ATTR_BURN_INJURED_RATE   = 32 // 点燃受伤率
	ATTR_BLEED_INJURED_RATE  = 33 // 流血受伤率
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

func (this *BattleTeam) get_default_targets(team *BattleTeam) (is_enemy bool, pos []int32) {
	return
}

// find targets
func (this *BattleTeam) FindTargets(team *BattleTeam) (is_enemy bool, pos []int32) {
	skill_id := int32(0)
	m := this.members[this.curr_attack]

	// 能量满用绝杀
	if m.energy >= BATTLE_TEAM_MEMBER_MAX_ENERGY {
		skill_id = m.card.SuperSkillID
	} else {
		skill_id = m.card.NormalSkillID
	}

	skill := skill_table_mgr.Get(skill_id)
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

	} else if skill.SkillTarget == SKILL_TARGET_TYPE_BACK {

	} else if skill.SkillTarget == SKILL_TARGET_TYPE_HP_MIN {

	} else if skill.SkillTarget == SKILL_TARGET_TYPE_RANDOM {

	} else if skill.SkillTarget == SKILL_TARGET_TYPE_SELF {

	} else {
		log.Error("Invalid skill target type: %v", skill.SkillTarget)
		return
	}

	/*for i := 0; i < BATTLE_FORMATION_LINE_NUM; i++ {
		a := pos % BATTLE_FORMATION_ONE_LINE_MEMBER_NUM
		pos2 := i*BATTLE_FORMATION_ONE_LINE_MEMBER_NUM + a
		if team[1].members[pos2] != nil {
			m2 = team[1].members[pos2]
			break
		}
	}
	if m2 == nil {
		for i := 0; i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
			pos2 := (pos + i) % BATTLE_TEAM_MEMBER_MAX_NUM
			if team[1].members[pos2] != nil {
				m2 = team[1].members[pos2]
				break
			}
		}
	}*/
	return
}
