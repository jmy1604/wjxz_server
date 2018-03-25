package main

import (
	"libs/log"
	"main/table_config"
	"math"
	"math/rand"
	"public_message/gen_go/client_message"
	"time"
)

// 技能类型
const (
	SKILL_TYPE_NORMAL  = 1 // 普攻
	SKILL_TYPE_SUPER   = 2 // 绝杀
	SKILL_TYPE_PASSIVE = 3 // 被动
	SKILL_TYPE_NEXT    = 4 // 连携
)

// 技能战斗类型
const (
	SKILL_FIGHT_TYPE_NONE   = iota
	SKILL_FIGHT_TYPE_MELEE  = 1 // 近战
	SKILL_FIGHT_TYPE_REMOTE = 2 // 远程
)

// 技能敌我类型
const (
	SKILL_ENEMY_TYPE_OUR   = 1 // 我方
	SKILL_ENEMY_TYPE_ENEMY = 2 // 敌方
)

// 技能攻击范围
const (
	SKILL_RANGE_TYPE_SINGLE    = 1 // 单个
	SKILL_RANGE_TYPE_ROW       = 2 // 横排
	SKILL_RANGE_TYPE_COLUMN    = 3 // 竖排
	SKILL_RANGE_TYPE_MULTI     = 4 // 多个
	SKILL_RANGE_TYPE_CROSS     = 5 // 十字
	SKILL_RANGE_TYPE_BIG_CROSS = 6 // 大十字
	SKILL_RANGE_TYPE_ALL       = 7 // 全体
)

// 技能目标类型
const (
	SKILL_TARGET_TYPE_DEFAULT = 1 // 默认
	SKILL_TARGET_TYPE_BACK    = 2 // 后排
	SKILL_TARGET_TYPE_HP_MIN  = 3 // 血最少
	SKILL_TARGET_TYPE_RANDOM  = 4 // 随机
	SKILL_TARGET_TYPE_SELF    = 5 // 自身
)

// BUFF效果类型
const (
	BUFF_EFFECT_TYPE_DAMAGE                = 1
	BUFF_EFFECT_TYPE_DISABLE_NORMAL_ATTACK = 2
	BUFF_EFFECT_TYPE_DISABLE_SUPER_ATTACK  = 3
	BUFF_EFFECT_TYPE_DISABLE_ACTION        = 4
	BUFF_EFFECT_TYPE_MODIFY_ATTR           = 5
	BUFF_EFFECT_TYPE_DODGE                 = 6
	BUFF_EFFECT_TYPE_COUNT                 = 8
)

// 获取行数顺序
func _get_rows_order(self_pos int32) (rows_order []int32) {
	if self_pos%BATTLE_FORMATION_ONE_LINE_MEMBER_NUM == 0 {
		rows_order = []int32{0, 1, 2}
	} else if self_pos%BATTLE_FORMATION_ONE_LINE_MEMBER_NUM == 1 {
		rows_order = []int32{1, 0, 2}
	} else if self_pos%BATTLE_FORMATION_ONE_LINE_MEMBER_NUM == 2 {
		rows_order = []int32{2, 1, 0}
	} else {
		log.Warn("not impossible self_pos[%v]", self_pos)
	}
	return
}

// 行是否为空
func _check_team_row(row_index int32, target_team *BattleTeam) (is_empty bool, pos []int32) {
	is_empty = true
	for i := 0; i < BATTLE_FORMATION_ONE_LINE_MEMBER_NUM; i++ {
		p := row_index + int32(BATTLE_FORMATION_LINE_NUM*i)
		if target_team.members[p] != nil {
			pos = append(pos, p)
			if is_empty {
				is_empty = false
			}
		}
	}
	return
}

// 列是否为空
func _check_team_column(column_index int32, target_team *BattleTeam) (is_empty bool, pos []int32) {
	is_empty = true
	for i := 0; i < BATTLE_FORMATION_LINE_NUM; i++ {
		p := int(column_index)*BATTLE_FORMATION_ONE_LINE_MEMBER_NUM + i
		if target_team.members[p] != nil {
			pos = append(pos, int32(p))
			if is_empty {
				is_empty = false
			}
		}
	}
	return
}

// 十字攻击范围
func _get_team_cross_targets() [][]int32 {
	return [][]int32{
		[]int32{0, 2, 3},
		[]int32{1, 0, 2, 4},
		[]int32{2, 1, 5},
		[]int32{3, 0, 4},
		[]int32{4, 1, 3, 5, 7},
		[]int32{5, 2, 4, 8},
		[]int32{6, 3, 7},
		[]int32{7, 4, 6, 8},
		[]int32{8, 5, 7},
	}
}

// 大十字攻击范围
func _get_team_big_cross_targets() [][]int32 {
	return [][]int32{
		[]int32{0, 1, 2, 3, 6},
		[]int32{1, 0, 2, 4, 7},
		[]int32{2, 1, 0, 5, 8},
		[]int32{3, 0, 6, 4, 5},
		[]int32{4, 1, 3, 5, 7},
		[]int32{5, 2, 4, 3, 8},
		[]int32{6, 3, 0, 7, 8},
		[]int32{7, 4, 1, 6, 8},
		[]int32{8, 5, 3, 7, 6},
	}
}

// 单个默认目标
func _get_single_default_target(self_pos int32, target_team *BattleTeam) (pos int32) {
	pos = int32(-1)
	m := target_team.members[self_pos]
	if m != nil {
		pos = self_pos
	} else {
		rows_order := _get_rows_order(self_pos)
		for l := 0; l < len(rows_order); l++ {
			for i := 0; i < BATTLE_FORMATION_ONE_LINE_MEMBER_NUM; i++ {
				p := int(rows_order[l])*BATTLE_FORMATION_ONE_LINE_MEMBER_NUM + i
				if target_team.members[p] != nil {
					pos = int32(p)
					break
				}
			}
		}
	}
	return
}

// 默认目标选择
func skill_get_default_targets(self_pos int32, target_team *BattleTeam, skill_data *table_config.XmlSkillItem) (pos []int32) {
	if skill_data.RangeType == SKILL_RANGE_TYPE_SINGLE { // 单体
		pos = []int32{_get_single_default_target(self_pos, target_team)}
	} else if skill_data.RangeType == SKILL_RANGE_TYPE_ROW { //横排
		rows := _get_rows_order(self_pos)
		if rows == nil {
			log.Warn("get rows failed")
			return
		}
		is_empty := false
		for i := 0; i < len(rows); i++ {
			is_empty, pos = _check_team_row(rows[i], target_team)
			if !is_empty {
				break
			}
		}
	} else if skill_data.RangeType == SKILL_RANGE_TYPE_COLUMN { // 竖排
		for c := 0; c < BATTLE_FORMATION_LINE_NUM; c++ {
			is_empty := false
			is_empty, pos = _check_team_column(int32(c), target_team)
			if !is_empty {
				break
			}
		}
	} else if skill_data.RangeType == SKILL_RANGE_TYPE_MULTI { // 多体
		// 默认多体不存在
		log.Warn("Cant get target pos on default multi members")
	} else if skill_data.RangeType == SKILL_RANGE_TYPE_CROSS { // 十字
		p := _get_single_default_target(self_pos, target_team)
		if p < 0 {
			log.Error("Get single target pos by self_pos[%v] failed", self_pos)
			return
		}
		ps := _get_team_cross_targets()[p]
		for i := 0; i < len(ps); i++ {
			if target_team.members[ps[i]] != nil {
				pos = append(pos, ps[i])
			}
		}
	} else if skill_data.RangeType == SKILL_RANGE_TYPE_BIG_CROSS { // 大十字
		p := _get_single_default_target(self_pos, target_team)
		if p < 0 {
			log.Error("Get single target pos by self_pos[%v] failed", self_pos)
			return
		}
		ps := _get_team_big_cross_targets()[p]
		for i := 0; i < len(ps); i++ {
			if target_team.members[ps[i]] != nil {
				pos = append(pos, ps[i])
			}
		}
	} else if skill_data.RangeType == SKILL_RANGE_TYPE_ALL { // 全体
		for i := 0; i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
			if target_team.members[i] != nil {
				pos = append(pos, int32(i))
			}
		}
	} else {
		log.Error("Unknown skill range type: %v", skill_data.RangeType)
	}
	return
}

// 后排目标选择
func skill_get_back_targets(self_pos int32, target_team *BattleTeam, skill_data *table_config.XmlSkillItem) (pos []int32) {
	if skill_data.RangeType == SKILL_RANGE_TYPE_SINGLE { // 单体
		for i := BATTLE_FORMATION_LINE_NUM - 1; i >= 0; i-- {
			for j := 0; j < BATTLE_FORMATION_ONE_LINE_MEMBER_NUM; j++ {
				p := int32(i)*BATTLE_FORMATION_ONE_LINE_MEMBER_NUM + int32(j)
				if target_team.members[p] != nil {
					pos = append(pos, p)
					break
				}
			}
		}
	} else if skill_data.RangeType == SKILL_RANGE_TYPE_COLUMN { // 竖排
		is_empty := false
		for i := BATTLE_FORMATION_LINE_NUM - 1; i >= 0; i-- {
			is_empty, pos = _check_team_column(int32(i), target_team)
			if !is_empty {
				break
			}
		}
	} else {
		log.Warn("Range type %v cant get back targets", skill_data.RangeType)
	}
	return
}

// 血最少选择
func skill_get_hp_min_targets(self_pos int32, target_team *BattleTeam, skill_data *table_config.XmlSkillItem) (pos []int32) {
	if skill_data.RangeType == SKILL_RANGE_TYPE_SINGLE {
		hp := int32(0)
		p := int32(-1)
		for i := 0; i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
			m := target_team.members[i]
			if m != nil {
				if hp == 0 || hp > m.hp {
					hp = m.hp
					p = int32(i)
				}
			}
		}
		pos = append(pos, p)
	} else {
		log.Warn("Range type %v cant get hp min targets", skill_data.RangeType)
	}
	return
}

// 随机一个目标
func _random_one_target(self_pos int32, target_team *BattleTeam, except_pos []int32) (pos int32) {
	pos = int32(-1)
	c := int32(0)
	r := rand.Int31n(BATTLE_TEAM_MEMBER_MAX_NUM)
	for {
		used := false
		if except_pos != nil {
			for i := 0; i < len(except_pos); i++ {
				if r == except_pos[i] {
					used = true
					break
				}
			}
		}
		if !used && (self_pos < 0 || self_pos != r) && target_team.members[r] != nil {
			pos = r
			break
		}
		r = (r + 1) % BATTLE_TEAM_MEMBER_MAX_NUM
		c += 1
		if c >= BATTLE_TEAM_MEMBER_MAX_NUM {
			break
		}
	}
	return
}

// 随机选择
func skill_get_random_targets(self_pos int32, target_team *BattleTeam, skill_data *table_config.XmlSkillItem) (pos []int32) {
	if skill_data.RangeType == SKILL_RANGE_TYPE_SINGLE {
		rand.Seed(time.Now().Unix())
		p := _random_one_target(self_pos, target_team, pos)
		if p < 0 {
			log.Error("Cant get random one target with self_pos %v", self_pos)
			return
		}
		pos = append(pos, p)
	} else if skill_data.RangeType == SKILL_RANGE_TYPE_MULTI {
		rand.Seed(time.Now().Unix())
		for i := int32(0); i < skill_data.MaxTarget; i++ {
			p := _random_one_target(self_pos, target_team, pos)
			if p >= 0 {
				pos = append(pos, p)
			}
		}
	} else {
		log.Warn("Range type %v cant get random targets", skill_data.RangeType)
	}
	return
}

// 技能条件
const (
	SKILL_COND_TYPE_NONE           = iota
	SKILL_COND_TYPE_HAS_LABEL      = 1
	SKILL_COND_TYPE_HAS_BUFF       = 2
	SKILL_COND_TYPE_HP_NOT_LESS    = 3
	SKILL_COND_TYPE_HP_GREATER     = 4
	SKILL_COND_TYPE_HP_NOT_GREATER = 5
	SKILL_COND_TYPE_HP_LESS        = 6
	SKILL_COND_TYPE_MP_NOT_LESS    = 7
	SKILL_COND_TYPE_MP_NOT_GREATER = 8
	SKILL_COND_TYPE_TEAM_HAS_ROLE  = 9
	SKILL_COND_TYPE_IS_TYPE        = 10
	SKILL_COND_TYPE_IS_CAMP        = 11
)

func _skill_check_cond(mem *TeamMember, effect_cond []int32) bool {
	if len(effect_cond) > 0 {
		if effect_cond[0] == SKILL_COND_TYPE_NONE {
			return true
		}
		if effect_cond[0] == SKILL_COND_TYPE_HAS_LABEL {
			if mem.card.Label != effect_cond[1] {
				return true
			}
		} else if effect_cond[0] == SKILL_COND_TYPE_HAS_BUFF {
			if mem.has_buff(effect_cond[1]) {
				return true
			}
		} else if effect_cond[0] == SKILL_COND_TYPE_HP_NOT_LESS {
			if mem.attrs[ATTR_HP] >= effect_cond[1] {
				return true
			}
		} else if effect_cond[0] == SKILL_COND_TYPE_HP_GREATER {
			if mem.attrs[ATTR_HP] > effect_cond[1] {
				return true
			}
		} else if effect_cond[0] == SKILL_COND_TYPE_HP_NOT_GREATER {
			if mem.attrs[ATTR_HP] <= effect_cond[1] {
				return true
			}
		} else if effect_cond[0] == SKILL_COND_TYPE_HP_LESS {
			if mem.attrs[ATTR_HP] < effect_cond[1] {
				return true
			}
		} else if effect_cond[0] == SKILL_COND_TYPE_MP_NOT_LESS {
			if mem.attrs[ATTR_MP] >= effect_cond[1] {
				return true
			}
		} else if effect_cond[0] == SKILL_COND_TYPE_MP_NOT_GREATER {
			if mem.attrs[ATTR_MP] <= effect_cond[1] {
				return true
			}
		} else if effect_cond[0] == SKILL_COND_TYPE_TEAM_HAS_ROLE {
			if mem.team.HasRole(effect_cond[1]) {
				return true
			}
		} else if effect_cond[0] == SKILL_COND_TYPE_IS_TYPE {
			if mem.card.Type == effect_cond[1] {
				return true
			}
		} else if effect_cond[0] == SKILL_COND_TYPE_IS_CAMP {
			if mem.card.Camp == effect_cond[1] {
				return true
			}
		} else {
			log.Warn("skill effect cond %v unknown", effect_cond[0])
		}
	}
	return true
}

func skill_check_cond(self *TeamMember, target *TeamMember, effect_cond1 []int32, effect_cond2 []int32) bool {
	if len(effect_cond1) == 0 && len(effect_cond2) == 0 {
		return true
	}

	if !_skill_check_cond(self, effect_cond1) {
		return false
	}

	if !_skill_check_cond(target, effect_cond2) {
		return false
	}

	return true
}

// 技能效果类型
const (
	SKILL_EFFECT_TYPE_DIRECT_INJURY         = 1  // 直接伤害
	SKILL_EFFECT_TYPE_CURE                  = 2  // 治疗
	SKILL_EFFECT_TYPE_ADD_BUFF              = 3  // 施加BUFF
	SKILL_EFFECT_TYPE_SUMMON                = 4  // 召唤技能
	SKILL_EFFECT_TYPE_MODIFY_ATTR           = 5  // 改变下次计算时的角色参数
	SKILL_EFFECT_TYPE_MODIFY_NORMAL_SKILL   = 6  // 改变普通攻击技能ID
	SKILL_EFFECT_TYPE_MODIFY_RAGE_SKILL     = 7  // 改变怒气攻击技能ID
	SKILL_EFFECT_TYPE_ADD_NORMAL_ATTACK_NUM = 8  // 增加普攻次数
	SKILL_EFFECT_TYPE_MODIFY_RAGE           = 9  // 改变怒气
	SKILL_EFFECT_TYPE_AURA                  = 10 // 光环
)

// 技能直接伤害
func skill_effect_direct_injury(self *TeamMember, target *TeamMember, skill_type, skill_fight_type int32, effect []int32) (target_damage, self_damage int32) {
	if len(effect) < 4 {
		log.Error("skill effect length %v not enough", len(effect))
		return
	}

	// 增伤减伤总和
	damage_add := self.attrs[ATTR_TOTAL_DAMAGE_ADD]
	damage_sub := target.attrs[ATTR_TOTAL_DAMAGE_SUB]

	// 类型
	if skill_type == SKILL_TYPE_NORMAL {
		damage_add += self.attrs[ATTR_NORMAL_DAMAGE_ADD]
		damage_sub += target.attrs[ATTR_NORMAL_DAMAGE_SUB]
	} else if skill_type == SKILL_TYPE_SUPER {
		damage_add += self.attrs[ATTR_RAGE_DAMAGE_ADD]
		damage_sub += target.attrs[ATTR_RAGE_DAMAGE_SUB]
	} else {
		log.Error("Invalid skill type: %v", skill_type)
		return
	}

	// 战斗类型
	if skill_fight_type == SKILL_FIGHT_TYPE_MELEE {
		damage_add += self.attrs[ATTR_CLOSE_DAMAGE_ADD]
		damage_sub += target.attrs[ATTR_CLOSE_DAMAGE_SUB]
	} else if skill_fight_type == SKILL_FIGHT_TYPE_REMOTE {
		damage_add += self.attrs[ATTR_REMOTE_DAMAGE_ADD]
		damage_sub += target.attrs[ATTR_REMOTE_DAMAGE_SUB]
	} else {
		log.Error("Invalid skill melee type: %v", skill_fight_type)
		return
	}

	is_reflect_damage := false
	// 角色类型克制
	if self.card.Type == table_config.CARD_ROLE_TYPE_ATTACK && target.card.Type == table_config.CARD_ROLE_TYPE_SKILL {
		damage_add += 1500
		is_reflect_damage = true
	} else if self.card.Type == table_config.CARD_ROLE_TYPE_SKILL && target.card.Type == table_config.CARD_ROLE_TYPE_DEFENSE {
		damage_add += 1500
		is_reflect_damage = true
	} else if self.card.Type == table_config.CARD_ROLE_TYPE_DEFENSE && target.card.Type == table_config.CARD_ROLE_TYPE_ATTACK {
		damage_add += 1500
		is_reflect_damage = true
	}

	// 反伤
	if is_reflect_damage {
		var reflect_damage int32
		if skill_fight_type == SKILL_FIGHT_TYPE_MELEE {
			reflect_damage = target.attrs[ATTR_ATTACK] * target.attrs[ATTR_CLOSE_REFLECT] / 10000
		} else if skill_fight_type == SKILL_FIGHT_TYPE_REMOTE {
			reflect_damage = target.attrs[ATTR_ATTACK] * target.attrs[ATTR_REMOTE_REFLECT] / 10000
		}
		if self.attrs[ATTR_HP] < reflect_damage {
			self.attrs[ATTR_HP] = 0
		} else {
			self.attrs[ATTR_HP] -= reflect_damage
		}
	}

	// 防御力
	defense := target.attrs[ATTR_DEFENSE] * (10000 - self.attrs[ATTR_BREAK_ARMOR] + self.attrs[ATTR_ARMOR_ADD]) / 10000
	if defense < 0 {
		defense = 0
	}
	attack := self.attrs[ATTR_ATTACK] - defense
	attack1 := self.attrs[ATTR_ATTACK] * self.attrs[ATTR_ATTACK] / (self.attrs[ATTR_ATTACK] + defense) / 5
	if attack < attack1 {
		attack = attack1
	}
	if attack < 1 {
		attack = 1
	}

	// 基础技能伤害
	base_skill_damage := attack * effect[1] / 10000
	target_damage = int32(float64(base_skill_damage) * math.Max(0.1, float64((10000+damage_add-damage_sub)/10000)) * float64(10000+self.attrs[ATTR_DAMAGE_PERCENT_BONUS]) / 10000)
	if target_damage < 1 {
		target_damage = 1
	}

	// 实际暴击率
	is_critical := false
	critical := self.attrs[ATTR_CRITICAL] - self.attrs[ATTR_ANTI_CRITICAL]
	if critical < 0 {
		critical = 0
	} else {
		// 触发暴击
		if critical > rand.Int31n(10000) {
			target_damage *= int32(math.Max(1.5, float64((20000+self.attrs[ATTR_CRITICAL_MULTI])/10000)))
			is_critical = true
			log.Debug("####### target_damage[%v]", target_damage)
		}
	}
	if !is_critical {
		// 实际格挡率
		block := target.attrs[ATTR_BLOCK_RATE] - self.attrs[ATTR_BREAK_BLOCK_RATE]
		if block > rand.Int31n(10000) {
			target_damage = int32(math.Max(1, float64(target_damage)*math.Max(0.1, math.Min(0.9, float64((5000-self.attrs[ATTR_BLOCK_DEFENSE_RATE]))/10000))))
			log.Debug("@@@@@@@ target_damage[%v]", target_damage)
		}
	}

	// 贯通
	if effect[3] > 0 {
		if target.attrs[ATTR_SHIELD] < target_damage {
			target.attrs[ATTR_SHIELD] = 0
		} else {
			target.attrs[ATTR_SHIELD] -= target_damage
		}
	} else {
		if target.attrs[ATTR_SHIELD] < target_damage {
			target.attrs[ATTR_SHIELD] = 0
			target_damage -= target.attrs[ATTR_SHIELD]
		} else {
			target.attrs[ATTR_SHIELD] -= target_damage
			target_damage = 0
		}
	}

	// 状态伤害

	return
}

// 技能治疗效果
func skill_effect_cure(self_mem *TeamMember, target_mem *TeamMember, effect []int32) (cure int32) {
	if len(effect) < 2 {
		log.Error("cure skill effect length %v not enough", len(effect))
		return
	}
	cure = self_mem.attrs[ATTR_ATTACK]
	cure = cure * effect[1] / 10000
	cure = int32(math.Max(0, float64(cure*(10000+self_mem.attrs[ATTR_CURE_RATE_CORRECT]+target_mem.attrs[ATTR_CURED_RATE_CORRECT])/10000)))
	return
}

// 施加BUFF
func skill_effect_add_buff(self_mem *TeamMember, target_mem *TeamMember, effect []int32) (buff_id int32) {
	if len(effect) < 5 {
		log.Error("add buff skill effect length %v not enough", len(effect))
		return
	}
	buff_id = target_mem.add_buff(self_mem, effect)
	return
}

// 技能效果
func skill_effect(self_team *BattleTeam, self_pos int32, target_team *BattleTeam, target_pos []int32, skill_data *table_config.XmlSkillItem) (report *msg_client_message.BattleReportItem) {
	effects := skill_data.Effects
	self := self_team.members[self_pos]
	if self == nil {
		return
	}

	for i := 0; i < len(target_pos); i++ {
		target := target_team.members[target_pos[i]]
		if target == nil {
			continue
		}
		for i := 0; i < len(effects); i++ {
			if effects[i] == nil || len(effects[i]) < 1 {
				continue
			}
			if !skill_check_cond(self, target, skill_data.EffectsCond1s[i], skill_data.EffectsCond2s[i]) {
				log.Debug("self[%v] cant use skill[%v] to target[%v]")
				continue
			}
			effect_type := effects[i][0]
			if effect_type == SKILL_EFFECT_TYPE_DIRECT_INJURY {
				// 直接伤害
				target_dmg, self_dmg := skill_effect_direct_injury(self, target, skill_data.Type, skill_data.SkillMelee, effects[i])
				if target_dmg != 0 {
					target.add_hp(-target_dmg)
				}
				if self_dmg != 0 {
					self.add_hp(-self_dmg)
				}
				report = battle_report_pool.Get()
				log.Debug("role[%v] use skill[%v] to enemy target[%v] with dmg[%v], target hp[%v], reflect self dmg[%v], self hp[%v]", self.id, skill_data.Id, target.id, target_dmg, target.hp, self_dmg, self.hp)
			} else if effect_type == SKILL_EFFECT_TYPE_CURE {
				// 治疗
				cure := skill_effect_cure(self, target, effects[i])
				if cure != 0 {
					target.add_hp(cure)
				}
				log.Debug("role[%v] use cure skill[%v] to self target[%v] with resume hp[%v]", self.id, skill_data.Id, target.id, cure)
			} else if effect_type == SKILL_EFFECT_TYPE_ADD_BUFF {
				// 施加BUFF
				buff_id := skill_effect_add_buff(self, target, effects[i])
				if buff_id > 0 {
					log.Debug("role[%v] use skill[%v] to target[%v] add buff[%v]", self.id, skill_data.Id, target.id, buff_id)
				}
			} else if effect_type == SKILL_EFFECT_TYPE_SUMMON {
				// 召唤
			} else if effect_type == SKILL_EFFECT_TYPE_MODIFY_ATTR {
				// 改变下次计算时的角色参数
				for i := 0; i < (len(effects[i])-1)/2; i++ {
					target.attrs[effects[i][1+2*i]] = effects[i][1+2*i+1]
				}
			} else if effect_type == SKILL_EFFECT_TYPE_MODIFY_NORMAL_SKILL {
				// 改变普通攻击技能ID
			} else if effect_type == SKILL_EFFECT_TYPE_MODIFY_RAGE_SKILL {
				// 改变必杀技ID
			} else if effect_type == SKILL_EFFECT_TYPE_MODIFY_RAGE {
				// 改变怒气
				if effects[i][3] > 0 {
					if rand.Int31n(10000) > effects[i][3] {
						target.energy += effects[i][1]
						self.energy += effects[i][2]
					}
				}
			} else if effect_type == SKILL_EFFECT_TYPE_ADD_NORMAL_ATTACK_NUM {
				// 增加行动次数
				target.act_num += effects[i][1]
			} else if effect_type == SKILL_EFFECT_TYPE_AURA {
				// 产生光环
			}
		}
	}

	return
}

type Buff struct {
	buff      *table_config.XmlStatusItem
	attack    int32
	dmg_add   int32
	param     int32
	round_num int32
	next      *Buff
	prev      *Buff
}

type BuffList struct {
	owner *TeamMember
	head  *Buff
	tail  *Buff
	//tmp_buffs map[*Buff]bool
}

func (this *BuffList) clear() {
	b := this.head
	for b != nil {
		next := b.next
		buff_pool.Put(b)
		b = next
	}
	this.head = nil
	this.tail = nil
	this.owner = nil
}

func (this *BuffList) remove_buff(buff *Buff) {
	if buff.prev != nil {
		buff.prev.next = buff.next
	}
	if buff.next != nil {
		buff.next.prev = buff.prev
	}
	if buff == this.head {
		this.head = buff.next
	}
	if buff == this.tail {
		this.tail = buff.prev
	}
	this.owner.remove_buff_effect(buff)

	buff_pool.Put(buff)
	/*if this.tmp_buffs == nil {
		this.tmp_buffs = make(map[*Buff]bool)
	}
	this.tmp_buffs[buff] = true*/

	log.Debug("@@@@@@@@@ remove buff[%v][%p][%v]", buff.buff.Id, buff, buff)
}

func (this *BuffList) check_buff_mutex(b *table_config.XmlStatusItem) bool {
	hh := this.head
	for hh != nil {
		next := hh.next
		for j := 0; j < len(hh.buff.ResistMutexTypes); j++ {
			if b.MutexType == hh.buff.ResistMutexTypes[j] {
				log.Debug("BUFF[%v]类型[%v]排斥BUFF[%v]类型[%v]", hh.buff.Id, hh.buff.MutexType, b.Id, b.MutexType)
				return true
			}
		}
		for j := 0; j < len(hh.buff.ResistMutexIDs); j++ {
			if b.Id == hh.buff.ResistMutexIDs[j] {
				log.Debug("BUFF[%v]排斥BUFF[%v]", hh.buff.Id, b.Id)
				return true
			}
		}
		for j := 0; j < len(hh.buff.CancelMutexTypes); j++ {
			if b.MutexType == hh.buff.CancelMutexTypes[j] {
				this.remove_buff(hh)
				log.Debug("BUFF[%v]类型[%v]驱散了BUFF[%v]类型[%v]", b.Id, b.MutexType, hh.buff.Id, hh.buff.MutexType)
			}
		}
		for j := 0; j < len(hh.buff.CancelMutexIDs); j++ {
			if b.Id == hh.buff.CancelMutexIDs[j] {
				this.remove_buff(hh)
				log.Debug("BUFF[%v]驱散了BUFF[%v]", b.Id, hh.buff.Id)
			}
		}
		hh = next
	}
	return false
}

func (this *BuffList) add_buff(attacker *TeamMember, b *table_config.XmlStatusItem, skill_effect []int32) (buff_id int32) {
	buff := buff_pool.Get()
	buff.buff = b
	buff.attack = attacker.attrs[ATTR_ATTACK]
	buff.dmg_add = attacker.attrs[ATTR_TOTAL_DAMAGE_ADD]
	buff.param = skill_effect[3]
	buff.round_num = skill_effect[4]

	if this.head == nil {
		buff.prev = nil
		this.head = buff
	} else {
		buff.prev = this.tail
		this.tail.next = buff
	}
	this.tail = buff
	this.tail.next = nil

	buff_id = b.Id

	/*if this.tmp_buffs == nil {
		this.tmp_buffs = make(map[*Buff]bool)
	}
	delete(this.tmp_buffs, buff)*/

	log.Debug("######### add buff[%v] [%p] [%v]", b.Id, buff, buff)
	return
}

func (this *BuffList) on_round_end() {
	bf := this.head
	for bf != nil {
		next := bf.next
		if bf.round_num > 0 {
			if bf.buff.Effect[0] == BUFF_EFFECT_TYPE_DAMAGE {
				dmg := buff_effect_damage(bf.attack, bf.dmg_add, bf.param, bf.buff.Effect[1], this.owner)
				this.owner.add_hp(-dmg)
				log.Debug("role[%v] hp damage[%v] on buff[%v] left round[%v] end", this.owner.id, dmg, bf.buff.Id, bf.round_num)
			}

			bf.round_num -= 1
			if bf.round_num <= 0 {
				this.remove_buff(bf)
				log.Debug("role[%v] buff[%v] round over", this.owner.id, bf.buff.Id)
			}
		}
		bf = next
	}
}

// 状态伤害效果
func buff_effect_damage(user_attack, user_damage_add, skill_damage_coeff, attr int32, target *TeamMember) (damage int32) {
	base_damage := user_attack * skill_damage_coeff / 10000
	f := float64(10000 + user_damage_add - target.attrs[ATTR_TOTAL_DAMAGE_SUB] + target.attrs[attr])
	damage = int32(math.Max(1, float64(base_damage)*math.Max(0.1, f)/10000))
	return
}
