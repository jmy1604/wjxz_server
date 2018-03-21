package main

import (
	"libs/log"
	"main/table_config"
	_ "math/rand"
	_ "public_message/gen_go/client_message"
	"sync"
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

type MemberBuff struct {
	id   int32
	next *MemberBuff
}

type TeamMember struct {
	id      int32
	card    *table_config.XmlCardItem
	hp      int32
	energy  int32
	attack  int32
	defense int32
	act_num int32 // 行动次数
	attrs   []int32
	enable  bool
}

type TeamMemberPool struct {
	mems_pool *sync.Pool
	locker    *sync.RWMutex
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

func (this *TeamMember) init(id int32, level int32, role_card *table_config.XmlCardItem) {
	this.id = id
	this.card = role_card
	this.hp = role_card.BaseHP + (level-1)*role_card.GrowthHP/100
	this.attack = role_card.BaseAttack + (level-1)*role_card.GrowthAttack/100
	this.defense = role_card.BaseDefence + (level-1)*role_card.GrowthDefence/100
	this.energy = BATTLE_TEAM_MEMBER_INIT_ENERGY

	if this.attrs == nil {
		this.attrs = make([]int32, ATTR_COUNT_MAX)
	}
	this.attrs[ATTR_HP_MAX] = this.hp
	this.attrs[ATTR_HP] = this.hp
	this.attrs[ATTR_ATTACK] = this.attack
	this.attrs[ATTR_DEFENSE] = this.defense
	this.enable = true
}

func (this *TeamMember) round_start() {
	this.act_num += 1
	this.energy += BATTLE_TEAM_MEMBER_ADD_ENERGY
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

func (this *TeamMemberPool) Init() {
	this.mems_pool = &sync.Pool{
		New: func() interface{} {
			m := &TeamMember{}
			return m
		},
	}
	this.locker = &sync.RWMutex{}
}

func (this *TeamMemberPool) Get() *TeamMember {
	return this.mems_pool.Get().(*TeamMember)
}

func (this *TeamMemberPool) Put(m *TeamMember) {
	this.mems_pool.Put(m)
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
		for i := 0; i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
			this.members[i] = team_member_pool.Get()
		}
	}

	for i := 0; i < len(members); i++ {
		if members[i] <= 0 {
			this.members[i].enable = false
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

		this.members[i].init(members[i], level, role_card)

		// 装备BUFF增加属性

		m := this.members[i]
		log.Debug("mem[%v]: role_id[%v] role_rank[%v] hp[%v] energy[%v] attack[%v] defense[%v]", i, m.card.Id, m.card.Rank, m.hp, m.energy, m.attack, m.defense)
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

// 使用技能
func (this *BattleTeam) UseSkill(self_index int32, target_team *BattleTeam, target_pos []int32, skill *table_config.XmlSkillItem) {
	for i := 0; i < len(target_pos); i++ {
		skill_effect(this, self_index, target_team, target_pos, skill)
	}
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
			if mem == nil || !mem.enable {
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
			if mem == nil || !mem.enable || mem.get_use_skill() == 0 {
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
}

// 开打
func (this *BattleTeam) Fight(target_team *BattleTeam, end_type int32, end_param int32) {
	round_max := end_param
	if end_type == BATTLE_END_BY_ALL_DEAD {
		round_max = BATTLE_ROUND_MAX_NUM
	} else if end_type == BATTLE_END_BY_ROUND_OVER {
	}

	for c := int32(0); c < round_max; c++ {
		log.Debug("------ Round[%v]", c+1)
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
		if this.members[i] == nil || !this.members[i].enable {
			continue
		}
		if this.members[i].attrs[ATTR_HP] > 0 {
			all_dead = false
			break
		}
	}
	return all_dead
}
