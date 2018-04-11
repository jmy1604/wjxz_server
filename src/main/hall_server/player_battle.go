package main

import (
	"libs/log"
	"main/table_config"
	"math/rand"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"

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
	ATTR_COUNT_MAX             = 40
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

type DelaySkill struct {
	trigger_event int32
	skill         *table_config.XmlSkillItem
	trigger_pos   []int32
}

type TeamMember struct {
	team                *BattleTeam
	pos                 int32
	id                  int32
	level               int32
	card                *table_config.XmlCardItem
	hp                  int32
	energy              int32
	attack              int32
	defense             int32
	act_num             int32                                 // 行动次数
	attrs               []int32                               // 属性
	bufflist_arr        []*BuffList                           // BUFF
	passive_triggers    map[int32][]*MemberPassiveTriggerData // 被动技触发事件
	temp_normal_skill   int32                                 // 临时普通攻击
	temp_super_skill    int32                                 // 临时怒气攻击
	temp_changed_attrs  []int32                               // 临时改变的属性
	buff_trigger_skills map[int32]int32                       // BUFF触发的技能
	delay_skills        []*DelaySkill                         // 延迟的技能效果
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
		this.add_passive_trigger(role_card.PassiveSkillIds[i])
	}
}

func (this *TeamMember) add_passive_trigger(skill_id int32) bool {
	skill := skill_table_mgr.Get(skill_id)
	if skill == nil || skill.Type != SKILL_TYPE_PASSIVE {
		return false
	}

	if this.passive_triggers == nil {
		this.passive_triggers = make(map[int32][]*MemberPassiveTriggerData)
	}

	d := passive_trigger_data_pool.Get()
	d.skill = skill
	d.battle_num = skill.TriggerBattleMax
	d.round_num = skill.TriggerRoundMax
	if skill.TriggerBattleMax <= 0 {
		d.battle_num = -1
	}
	if skill.TriggerRoundMax <= 0 {
		d.round_num = -1
	}
	datas := this.passive_triggers[skill.SkillTriggerType]
	if datas == nil {
		this.passive_triggers[skill.SkillTriggerType] = []*MemberPassiveTriggerData{d}
	} else {
		this.passive_triggers[skill.SkillTriggerType] = append(datas, d)
	}

	return true
}

func (this *TeamMember) delete_passive_trigger(skill_id int32) bool {
	skill := skill_table_mgr.Get(skill_id)
	if skill == nil || skill.Type != SKILL_TYPE_PASSIVE {
		return false
	}

	if this.passive_triggers == nil {
		return false
	}

	triggers := this.passive_triggers[skill.SkillTriggerType]
	if triggers == nil {
		return false
	}

	l := len(triggers)
	i := l - 1
	for ; i >= 0; i-- {
		if triggers[i] == nil {
			continue
		}
		if triggers[i].skill.Id == skill_id {
			triggers[i] = nil
			break
		}
	}

	if i >= 0 {
		for n := i; n < l-1; n++ {
			triggers[n] = triggers[n+1]
		}
		if l > 1 {
			this.passive_triggers[skill.SkillTriggerType] = triggers[:l-1]
		} else {
			delete(this.passive_triggers, skill.SkillTriggerType)
		}
	}

	return true
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
		if d[i].battle_num != 0 && d[i].round_num != 0 {
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
		if d[i].battle_num == 0 || d[i].round_num == 0 {
			passive_trigger_data_pool.Put(d[i])
		}
		break
	}
}

func (this *TeamMember) init(team *BattleTeam, id int32, level int32, role_card *table_config.XmlCardItem, pos int32) {
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
			this.bufflist_arr[i].owner = this
		}
	}

	this.team = team
	this.id = id
	this.pos = pos
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

func (this *TeamMember) add_attr(attr int32, value int32) {
	if attr == ATTR_HP {
		this.add_hp(value)
	} else if attr == ATTR_HP_MAX {
		this.add_max_hp(value)
	} else {
		this.attrs[attr] += value
	}
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
		this.bufflist_arr = make([]*BuffList, BUFF_EFFECT_TYPE_COUNT)
		for i := 0; i < BUFF_EFFECT_TYPE_COUNT; i++ {
			this.bufflist_arr[i] = &BuffList{}
			this.bufflist_arr[i].owner = this
		}
	}

	// 互斥
	for i := 0; i < len(this.bufflist_arr); i++ {
		h := this.bufflist_arr[i]
		if h != nil && h.check_buff_mutex(b) {
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
			bufflist := this.bufflist_arr[i]
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
	if buff.buff == nil || buff.buff.Effect == nil {
		return
	}

	if len(buff.buff.Effect) >= 2 {
		effect_type := buff.buff.Effect[0]
		if effect_type == BUFF_EFFECT_TYPE_MODIFY_ATTR {
			if this.attrs[buff.buff.Effect[1]] < buff.param {
				this.attrs[buff.buff.Effect[1]] = 0
			} else {
				this.attrs[buff.buff.Effect[1]] -= buff.param
			}

			if buff.buff.Effect[1] == ATTR_HP_MAX && this.attrs[ATTR_HP] > this.attrs[ATTR_HP_MAX] {
				this.attrs[ATTR_HP] = this.attrs[ATTR_HP_MAX]
			}
		} else if effect_type == BUFF_EFFECT_TYPE_TRIGGER_SKILL {
			if _, o := this.buff_trigger_skills[buff.buff.Effect[1]]; o {
				delete(this.buff_trigger_skills, buff.buff.Effect[1])
				this.delete_passive_trigger(buff.buff.Effect[1])
			}
		}
	}
}

func (this *TeamMember) is_disable_normal_attack() bool {
	if this.bufflist_arr == nil {
		return false
	}
	disable := false
	bufflist := this.bufflist_arr[BUFF_EFFECT_TYPE_DISABLE_NORMAL_ATTACK]
	if bufflist.head != nil {
		disable = true
	} else {
		bufflist = this.bufflist_arr[BUFF_EFFECT_TYPE_DISABLE_ACTION]
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
	bufflist := this.bufflist_arr[BUFF_EFFECT_TYPE_DISABLE_SUPER_ATTACK]
	if bufflist.head != nil {
		disable = true
	} else {
		bufflist = this.bufflist_arr[BUFF_EFFECT_TYPE_DISABLE_ACTION]
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
	bufflist := this.bufflist_arr[BUFF_EFFECT_TYPE_DISABLE_ACTION]
	if bufflist.head != nil {
		disable = true
	} else {
		bufflist = this.bufflist_arr[BUFF_EFFECT_TYPE_DISABLE_NORMAL_ATTACK]
		if bufflist.head != nil {
			disable = true
		} else {
			bufflist = this.bufflist_arr[BUFF_EFFECT_TYPE_DISABLE_SUPER_ATTACK]
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

func (this *TeamMember) has_delay_skills() bool {
	if this.delay_skills == nil {
		return false
	}
	return true
}

func (this *TeamMember) push_delay_skill(trigger_event int32, skill *table_config.XmlSkillItem, trigger_pos []int32) {
	ds := delay_skill_pool.Get()
	ds.trigger_event = trigger_event
	ds.skill = skill
	ds.trigger_pos = trigger_pos
	this.delay_skills = append(this.delay_skills, ds)
}

func (this *TeamMember) delay_skills_effect(target_team *BattleTeam) {
	if this.delay_skills == nil {
		return
	}

	for i := 0; i < len(this.delay_skills); i++ {
		ds := this.delay_skills[i]
		if ds == nil {
			continue
		}

		// 延迟的被动技也要处理为连续技
		reports := this.team.reports.reports
		if reports != nil {
			l := len(reports)
			if l > 0 {
				reports[l-1].HasCombo = true
			}
			log.Debug("########################################### team[%v] member[%v] 后面有延迟被动技 %v", this.team.side, this.pos, ds.skill.Id)
		}

		one_passive_skill_effect(ds.trigger_event, ds.skill, this, target_team, ds.trigger_pos)
	}
}

func (this *TeamMember) clear_delay_skills() {
	if this.delay_skills == nil {
		return
	}
	for i := 0; i < len(this.delay_skills); i++ {
		delay_skill_pool.Put(this.delay_skills[i])
	}
	this.delay_skills = nil
}

type BattleReports struct {
	reports         []*msg_client_message.BattleReportItem
	remove_buffs    []*msg_client_message.BattleMemberBuff
	changed_members []*msg_client_message.BattleMemberItem
}

func (this *BattleReports) Reset() {
	this.reports = make([]*msg_client_message.BattleReportItem, 0)
	this.remove_buffs = make([]*msg_client_message.BattleMemberBuff, 0)
	this.changed_members = make([]*msg_client_message.BattleMemberItem, 0)
}

func (this *BattleReports) Recycle() {
	if this.reports != nil {
		for i := 0; i < len(this.reports); i++ {
			r := this.reports[i]
			if r == nil {
				continue
			}
			// user
			if r.User != nil {
				msg_battle_member_item_pool.Put(r.User)
				r.User = nil
			}
			// behiters
			if r.BeHiters != nil {
				for j := 0; j < len(r.BeHiters); j++ {
					if r.BeHiters[j] != nil {
						msg_battle_member_item_pool.Put(r.BeHiters[j])
					}
				}
				r.BeHiters = nil
			}
			// add buffs
			if r.AddBuffs != nil {
				for j := 0; j < len(r.AddBuffs); j++ {
					if r.AddBuffs[j] != nil {
						msg_battle_buff_item_pool.Put(r.AddBuffs[j])
						r.AddBuffs[j] = nil
					}
				}
				r.AddBuffs = nil
			}
			// remove buffs
			if r.RemoveBuffs != nil {
				for j := 0; j < len(r.RemoveBuffs); j++ {
					if r.RemoveBuffs[j] != nil {
						msg_battle_buff_item_pool.Put(r.RemoveBuffs[j])
						r.RemoveBuffs[j] = nil
					}
				}
				r.RemoveBuffs = nil
			}
			msg_battle_reports_item_pool.Put(r)
			this.reports[i] = nil
		}
		this.reports = nil
	}

	if this.remove_buffs != nil {
		for i := 0; i < len(this.remove_buffs); i++ {
			b := this.remove_buffs[i]
			if b == nil {
				continue
			}
			msg_battle_buff_item_pool.Put(b)
			this.remove_buffs[i] = nil
		}
		this.remove_buffs = nil
	}

	if this.changed_members != nil {
		for i := 0; i < len(this.changed_members); i++ {
			m := this.changed_members[i]
			if m == nil {
				continue
			}
			msg_battle_member_item_pool.Put(m)
			this.changed_members[i] = nil
		}
		this.changed_members = nil
	}
}

type BattleTeam struct {
	curr_attack  int32          // 当前进攻的索引
	side         int32          // 0 左边 1 右边
	temp_curr_id int32          // 临时ID，用于标识召唤的角色
	members      []*TeamMember  // 成员
	reports      *BattleReports // 每回合战报
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

	log.Debug("!@!@!@!@!@!@ members: %v", members)

	for i := 0; i < len(this.members); i++ {
		if (i < len(members) && members[i] <= 0) || i >= len(members) {
			this.members[i] = nil
			continue
		}

		//if this.members[i] != nil && members[i] == this.members[i].id {
		//	continue
		//}

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

		if p.team_member_mgr == nil {
			p.team_member_mgr = make(map[int32]*TeamMember)
		}
		m := p.team_member_mgr[members[i]]
		if m == nil {
			m = team_member_pool.Get()
			p.team_member_mgr[members[i]] = m
		}
		m.init(this, members[i], level, role_card, int32(i))
		this.members[i] = m

		// 装备BUFF增加属性
		log.Debug("mem[%v]: id[%v] role_id[%v] role_rank[%v] hp[%v] energy[%v] attack[%v] defense[%v]", i, m.id, m.card.Id, m.card.Rank, m.hp, m.energy, m.attack, m.defense)
	}
	this.curr_attack = 0
	this.side = side
	this.temp_curr_id = p.db.Global.GetCurrentRoleId() + 1

	//p.team_changed[team_id] = false

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
			if m.temp_normal_skill > 0 {
				skill_id = m.temp_normal_skill
			} else {
				skill_id = m.card.SuperSkillID
			}
		} else if m.act_num > 0 {
			if m.is_disable_normal_attack() {
				return
			}
			if m.temp_super_skill > 0 {
				skill_id = m.temp_super_skill
			} else {
				skill_id = m.card.NormalSkillID
			}
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

	} else if skill.SkillTarget == SKILL_TARGET_TYPE_EMPTY_POS {
		pos = skill_get_empty_pos(target_team, skill)
	} else {
		log.Error("Invalid skill target type: %v", skill.SkillTarget)
		return
	}

	return
}

func (this *BattleTeam) UseOnceSkill(self_index int32, target_team *BattleTeam, trigger_skill int32) (skill *table_config.XmlSkillItem) {
	self := this.members[self_index]
	if self == nil || self.is_dead() {
		return nil
	}

	is_enemy, target_pos, skill := this.FindTargets(self_index, target_team, trigger_skill)
	if target_pos == nil {
		log.Warn("team[%v] member[%v] Cant find targets to attack with skill[%v]", this.side, self_index, skill.Id)
		return nil
	}

	log.Debug("team[%v] member[%v] find is_enemy[%v] targets[%v] to use skill[%v]", this.side, self_index, is_enemy, target_pos, skill.Id)

	if !is_enemy {
		target_team = this
	}

	skill_effect(this, self_index, target_team, target_pos, skill)

	// 清除临时技能
	m := this.members[self_index]
	if m != nil {
		if m.temp_normal_skill > 0 {
			m.temp_normal_skill = 0
		} else if m.temp_super_skill > 0 {
			m.temp_super_skill = 0
		}
	}

	// 是否有combo技能
	if skill.ComboSkill > 0 {
		reports := this.reports.reports
		if reports != nil && len(reports) > 0 {
			r := reports[len(reports)-1]
			r.HasCombo = true
			log.Debug("########################################### Team[%v] member[%v] 后面有组合技 %v", this.side, self_index, skill.ComboSkill)
		}
	}

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
	// 延迟的被动技
	if mem.has_delay_skills() {
		mem.delay_skills_effect(target_team)
		mem.clear_delay_skills()
	}
	return 1
}

// 回合
func (this *BattleTeam) DoRound(target_team *BattleTeam) {
	this.RoundStart()
	target_team.RoundStart()

	// 被动技，回合行动前触发
	for i := int32(0); i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
		passive_skill_effect_with_self_pos(EVENT_BEFORE_ROUND, this, i, target_team, nil)
		passive_skill_effect_with_self_pos(EVENT_BEFORE_ROUND, target_team, i, this, nil)
	}

	var self_index, target_index int32
	for self_index < BATTLE_TEAM_MEMBER_MAX_NUM || target_index < BATTLE_TEAM_MEMBER_MAX_NUM {
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

// 回收战报
func (this *BattleTeam) RecycleReports() {

}

func _recycle_battle_reports(reports []*msg_client_message.BattleReportItem) {
	if reports != nil {
		for i := 0; i < len(reports); i++ {
			if reports[i] != nil {
				msg_battle_reports_item_pool.Put(reports[i])
			}
		}
	}
}

func _recycle_battle_rounds(rounds []*msg_client_message.BattleRoundReports) {
	if rounds == nil {
		return
	}
	for n := 0; n < len(rounds); n++ {
		if rounds[n] == nil {
			continue
		}
		if rounds[n].Reports != nil {
			for i := 0; i < len(rounds[n].Reports); i++ {
				r := rounds[n].Reports[i]
				if r == nil {
					continue
				}
				// user
				if r.User != nil {
					msg_battle_member_item_pool.Put(r.User)
					r.User = nil
				}
				// behiters
				if r.BeHiters != nil {
					for j := 0; j < len(r.BeHiters); j++ {
						if r.BeHiters[j] != nil {
							msg_battle_member_item_pool.Put(r.BeHiters[j])
						}
					}
					r.BeHiters = nil
				}
				// add buffs
				if r.AddBuffs != nil {
					for j := 0; j < len(r.AddBuffs); j++ {
						if r.AddBuffs[j] != nil {
							msg_battle_buff_item_pool.Put(r.AddBuffs[j])
							r.AddBuffs[j] = nil
						}
					}
					r.AddBuffs = nil
				}
				// remove buffs
				if r.RemoveBuffs != nil {
					for j := 0; j < len(r.RemoveBuffs); j++ {
						if r.RemoveBuffs[j] != nil {
							msg_battle_buff_item_pool.Put(r.RemoveBuffs[j])
							r.RemoveBuffs[j] = nil
						}
					}
					r.RemoveBuffs = nil
				}
				msg_battle_reports_item_pool.Put(r)
			}
		}
	}
}

// 开打
func (this *BattleTeam) Fight(target_team *BattleTeam, end_type int32, end_param int32) (is_win bool, enter_reports []*msg_client_message.BattleReportItem, rounds []*msg_client_message.BattleRoundReports) {
	round_max := end_param
	if end_type == BATTLE_END_BY_ALL_DEAD {
		round_max = BATTLE_ROUND_MAX_NUM
	} else if end_type == BATTLE_END_BY_ROUND_OVER {
	}

	// 存放战报
	if this.reports == nil {
		this.reports = &BattleReports{}
		target_team.reports = this.reports
	}
	this.reports.Reset()

	// 被动技，进场前触发
	for i := int32(0); i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
		passive_skill_effect_with_self_pos(EVENT_ENTER_BATTLE, this, i, target_team, nil)
		passive_skill_effect_with_self_pos(EVENT_ENTER_BATTLE, target_team, i, this, nil)
	}

	if this.reports.reports != nil {
		enter_reports = this.reports.reports
		this.reports.reports = make([]*msg_client_message.BattleReportItem, 0)
	}

	for c := int32(0); c < round_max; c++ {
		log.Debug("----------------------------------------------- Round[%v] --------------------------------------------", c+1)

		this.DoRound(target_team)

		round := msg_battle_round_reports_pool.Get()
		round.Reports = this.reports.reports
		round.RemoveBuffs = this.reports.remove_buffs
		round.ChangedMembers = this.reports.changed_members
		round.RoundNum = c + 1
		rounds = append(rounds, round)

		if this.IsAllDead() {
			log.Debug("self all dead")
			break
		}
		if target_team.IsAllDead() {
			is_win = true
			log.Debug("target all dead")
			break
		}

		this.reports.Reset()
	}

	return
}

func (this *BattleTeam) _format_members_for_msg() (members []*msg_client_message.BattleMemberItem) {
	for i := 0; i < len(this.members); i++ {
		if this.members[i] == nil {
			continue
		}
		mem := this.members[i].build_battle_item(int32(i), 0)
		members = append(members, mem)
	}
	return
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
		log.Error("C2SBattleResultRequest player[%v] proto is invalid", p.Id)
		return -1
	}

	if req.GetAttackMembers() != nil && len(req.GetAttackMembers()) > 0 {
		res := p.SetAttackTeam(req.AttackMembers)
		if res < 0 {
			log.Error("Player[%v] set attack members[%v] failed", p.Id, req.AttackMembers)
			return res
		}
	}

	p.Fight2Player(req.FightPlayerId)

	return 1
}

func C2SSetTeamHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SSetTeamRequest)
	if req == nil || p == nil {
		log.Error("C2SSetTeamHandler player[%v] proto is invalid", p.Id)
		return -1
	}

	tt := req.GetTeamType()
	if tt == 0 {
		p.SetAttackTeam(req.TeamMembers)
	} else if tt == 1 {
		p.SetDefenseTeam(req.TeamMembers)
	} else {
		log.Warn("Unknown team type[%v] to player[%v]", tt, p.Id)
	}

	response := &msg_client_message.S2CSetTeamResponse{}
	response.TeamType = tt
	response.TeamMembers = req.TeamMembers
	p.Send(uint16(msg_client_message_id.MSGID_S2C_SET_TEAM_RESPONSE), response)

	return 1
}
