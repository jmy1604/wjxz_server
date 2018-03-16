package main

import (
	_ "libs/log"
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
	SKILL_RANGE_TYPE_LINE      = 3 // 竖排
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
