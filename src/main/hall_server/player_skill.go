package main

import (
	"libs/log"
	"main/table_config"
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

// 获取行数顺序
func get_rows_order(self_pos int32) (rows_order []int32) {
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
func check_team_row(row_index int32, target_team *BattleTeam) (is_empty bool, pos []int32) {
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
func check_team_column(column_index int32, target_team *BattleTeam) (is_empty bool, pos []int32) {
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
func get_team_cross_targets() [][]int32 {
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
func get_team_big_cross_targets() [][]int32 {
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
func get_single_default_target(self_pos int32, target_team *BattleTeam) (pos int32) {
	pos = int32(-1)
	m := target_team.members[self_pos]
	if m != nil {
		pos = self_pos
	} else {
		rows_order := get_rows_order(self_pos)
		for l := 0; l < len(rows_order); l++ {
			for i := 1; i < BATTLE_FORMATION_ONE_LINE_MEMBER_NUM; i++ {
				p := l*BATTLE_FORMATION_ONE_LINE_MEMBER_NUM + i
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
func (this *Player) get_default_targets(self_pos int32, target_team *BattleTeam, skill_data *table_config.XmlSkillItem) (pos []int32) {
	if skill_data.RangeType == SKILL_RANGE_TYPE_SINGLE { // 单体
		pos = []int32{get_single_default_target(self_pos, target_team)}
	} else if skill_data.RangeType == SKILL_RANGE_TYPE_ROW { //横排
		rows := get_rows_order(self_pos)
		if rows == nil {
			log.Warn("get rows failed")
			return
		}
		is_empty := false
		for i := 0; i < len(rows); i++ {
			is_empty, pos = check_team_row(rows[i], target_team)
			if !is_empty {
				break
			}
		}
	} else if skill_data.RangeType == SKILL_RANGE_TYPE_COLUMN { // 竖排
		for c := 0; c < BATTLE_FORMATION_LINE_NUM; c++ {
			is_empty := false
			is_empty, pos = check_team_column(int32(c), target_team)
			if !is_empty {
				break
			}
		}
	} else if skill_data.RangeType == SKILL_RANGE_TYPE_MULTI { // 多体
		// 默认多体不存在
		log.Warn("Cant get target pos on default multi members")
	} else if skill_data.RangeType == SKILL_RANGE_TYPE_CROSS { // 十字
		p := get_single_default_target(self_pos, target_team)
		if p < 0 {
			log.Error("Get single target pos by self_pos[%v] failed", self_pos)
			return
		}
		ps := get_team_cross_targets()[p]
		for i := 0; i < len(ps); i++ {
			if target_team.members[ps[i]] != nil {
				pos = append(pos, ps[i])
			}
		}
	} else if skill_data.RangeType == SKILL_RANGE_TYPE_BIG_CROSS { // 大十字
		p := get_single_default_target(self_pos, target_team)
		if p < 0 {
			log.Error("Get single target pos by self_pos[%v] failed", self_pos)
			return
		}
		ps := get_team_big_cross_targets()[p]
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
func (this *Player) get_back_targets(self_pos int32, target_team *BattleTeam, skill_data *table_config.XmlSkillItem) (pos []int32) {
	return
}

// 血最少选择
func (this *Player) get_hp_min_targets(self_pos int32, target_team *BattleTeam, skill_data *table_config.XmlSkillItem) (pos []int32) {
	return
}

// 随机选择
func (this *Player) get_random_targets(self_pos int32, target_team *BattleTeam, skill_data *table_config.XmlSkillItem) (pos []int32) {
	return
}

// 强制自身选择
func (this *Player) get_self_targets(self_pos int32, target_team *BattleTeam, skill_data *table_config.XmlSkillItem) (pos []int32) {
	return
}
