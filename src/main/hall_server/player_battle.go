package main

import (
	"libs/log"
	"main/table_config"
	"math/rand"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	"time"

	"github.com/golang/protobuf/proto"
)

type DelaySkillList struct {
	head *DelaySkill
	tail *DelaySkill
}

type BattleCommonData struct {
	reports          []*msg_client_message.BattleReportItem
	remove_buffs     []*msg_client_message.BattleMemberBuff
	changed_fighters []*msg_client_message.BattleFighter
	round_num        int32
	delay_skill_list *DelaySkillList
	members_damage   []map[int32]int32
	members_cure     []map[int32]int32
}

func (this *BattleCommonData) Init() {
	if this.members_damage == nil {
		this.members_damage = make([]map[int32]int32, 2)
	}
	for i := 0; i < len(this.members_damage); i++ {
		this.members_damage[i] = make(map[int32]int32)
	}
	if this.members_cure == nil {
		this.members_cure = make([]map[int32]int32, 2)
	}
	for i := 0; i < len(this.members_cure); i++ {
		this.members_cure[i] = make(map[int32]int32)
	}
}

func (this *BattleCommonData) Reset() {
	this.reports = make([]*msg_client_message.BattleReportItem, 0)
	this.remove_buffs = make([]*msg_client_message.BattleMemberBuff, 0)
	this.changed_fighters = make([]*msg_client_message.BattleFighter, 0)
	if this.delay_skill_list != nil {
		d := this.delay_skill_list.head
		for d != nil {
			n := d.next
			delay_skill_pool.Put(d)
			d = n
		}
	}
}

func (this *BattleCommonData) Recycle() {
	if this.reports != nil {
		for i := 0; i < len(this.reports); i++ {
			r := this.reports[i]
			if r == nil {
				continue
			}
			// user
			if r.User != nil {
				msg_battle_fighter_pool.Put(r.User)
				r.User = nil
			}
			// behiters
			if r.BeHiters != nil {
				for j := 0; j < len(r.BeHiters); j++ {
					if r.BeHiters[j] != nil {
						msg_battle_fighter_pool.Put(r.BeHiters[j])
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

	if this.changed_fighters != nil {
		for i := 0; i < len(this.changed_fighters); i++ {
			m := this.changed_fighters[i]
			if m == nil {
				continue
			}
			msg_battle_fighter_pool.Put(m)
			this.changed_fighters[i] = nil
		}
		this.changed_fighters = nil
	}
}

type BattleTeam struct {
	player       *Player
	team_type    int32
	curr_attack  int32             // 当前进攻的索引
	side         int32             // 0 左边 1 右边
	temp_curr_id int32             // 临时ID，用于标识召唤的角色
	members      []*TeamMember     // 成员
	common_data  *BattleCommonData // 每回合战报
}

// 利用玩家初始化
func (this *BattleTeam) Init(p *Player, team_id int32, side int32) bool {
	var members []int32
	if team_id == BATTLE_ATTACK_TEAM {
		members = p.db.BattleTeam.GetAttackMembers()
	} else if team_id == BATTLE_DEFENSE_TEAM {
		members = p.db.BattleTeam.GetDefenseMembers()
	} else if team_id == BATTLE_CAMPAIN_TEAM {
		members = p.db.BattleTeam.GetCampaignMembers()
	} else if team_id == BATTLE_TOWER_TEAM {
		if p.tmp_teams == nil {
			p.tmp_teams = make(map[int32][]int32)
		}
		if p.tmp_teams[team_id] == nil {
			p.tmp_teams[team_id] = p.db.BattleTeam.GetAttackMembers()
		}
		members = p.tmp_teams[team_id]
	} else {
		log.Warn("Unknown team id %v", team_id)
		return false
	}

	if members == nil {
		return false
	}
	is_empty := true
	// 检测是否为空
	for i := 0; i < len(members); i++ {
		if members[i] > 0 {
			is_empty = false
			break
		}
	}
	if is_empty {
		return false
	}

	if this.members == nil {
		this.members = make([]*TeamMember, BATTLE_TEAM_MEMBER_MAX_NUM)
	}
	this.player = p
	this.team_type = team_id

	for i := 0; i < len(this.members); i++ {
		if (i < len(members) && members[i] <= 0) || i >= len(members) {
			this.members[i] = nil
			continue
		}

		m := p.get_team_member_by_role(members[i], this, int32(i))
		if m == nil {
			log.Error("Player[%v] init battle team get member with role_id[%v] error", p.Id, members[i])
			continue
		}
		this.members[i] = m
		// 装备BUFF增加属性
		log.Debug("mem[%v]: id[%v] role_id[%v] role_rank[%v] hp[%v] energy[%v] attack[%v] defense[%v]", i, m.id, m.card.Id, m.card.Rank, m.hp, m.energy, m.attack, m.defense)
	}
	this.curr_attack = 0
	this.side = side
	this.temp_curr_id = p.db.Global.GetCurrentRoleId() + 1

	return true
}

// init with stage
func (this *BattleTeam) InitWithStage(side int32, stage_id int32, monster_wave int32) bool {
	stage := stage_table_mgr.Get(stage_id)
	if stage == nil {
		log.Warn("Cant found stage %v", stage_id)
		return false
	}
	if stage.Monsters == nil || len(stage.Monsters) == 0 {
		return false
	}

	if this.members == nil {
		this.members = make([]*TeamMember, BATTLE_TEAM_MEMBER_MAX_NUM)
	}

	for i := 0; i < len(this.members); i++ {
		if this.members[i] != nil {
			team_member_pool.Put(this.members[i])
			this.members[i] = nil
		}
	}

	this.side = side
	this.curr_attack = 0

	for i := 0; i < len(stage.Monsters); i++ {
		monster := stage.Monsters[i]
		if monster.Wave-1 == monster_wave {
			pos := monster.Slot - 1
			if pos < 0 || pos >= BATTLE_ROUND_MAX_NUM {
				log.Error("Stage[%v] monster wave[%v] slot[%v] invalid", stage_id, monster_wave, monster.Slot)
				return false
			}

			role_card := card_table_mgr.GetRankCard(monster.MonsterID, monster.Rank)
			if role_card == nil {
				log.Error("Cant get card by role_id[%v] and rank[%v]", monster.MonsterID, monster.Rank)
				return false
			}

			m := team_member_pool.Get()

			m.init_all(this, 0, monster.Level, role_card, pos, monster.EquipID)
			this.members[pos] = m
		}
	}

	return true
}

// init with stage
func (this *BattleTeam) InitWithArenaRobot(robot *table_config.XmlArenaRobotItem, side int32) bool {
	if this.members == nil {
		this.members = make([]*TeamMember, BATTLE_TEAM_MEMBER_MAX_NUM)
	}

	for i := 0; i < len(this.members); i++ {
		if this.members[i] != nil {
			team_member_pool.Put(this.members[i])
			this.members[i] = nil
		}
	}

	this.side = side
	this.curr_attack = 0

	for i := 0; i < len(robot.RobotCardList); i++ {
		monster := robot.RobotCardList[i]
		pos := monster.Slot - 1
		if pos < 0 || pos >= BATTLE_ROUND_MAX_NUM {
			log.Error("Arena Robot[%v] monster slot[%v] invalid", robot.Id, monster.Slot)
			return false
		}

		role_card := card_table_mgr.GetRankCard(monster.MonsterID, monster.Rank)
		if role_card == nil {
			log.Error("Cant get card by role_id[%v] and rank[%v]", monster.MonsterID, monster.Rank)
			return false
		}

		m := team_member_pool.Get()

		m.init_all(this, 0, monster.Level, role_card, pos, monster.EquipID)
		this.members[pos] = m
	}

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
		if this.members[i] != nil && !this.members[i].is_dead() {
			this.members[i].round_end()
		}
	}
}

// find targets
func (this *BattleTeam) FindTargets(self *TeamMember, target_team *BattleTeam, trigger_skill int32) (is_enemy bool, pos []int32, skill *table_config.XmlSkillItem) {
	skill_id := int32(0)
	if trigger_skill == 0 {
		use_normal := true
		// 能量满用绝杀
		if self.energy >= BATTLE_TEAM_MEMBER_MAX_ENERGY {
			if !self.is_disable_super_attack() {
				use_normal = false
			} else {
				log.Debug("@@@@@@@@@@@!!!!!!!!!!!!!!! Team[%v] member[%v] disable super attack", this.side, self.pos)
			}
		} else {
			if self.is_disable_normal_attack() {
				log.Debug("@@@############## Team[%v] member[%v] disable all attack", this.side, self.pos)
				return
			}
		}

		if use_normal {
			if self.temp_normal_skill > 0 {
				skill_id = self.temp_normal_skill
				self.use_temp_skill = true
			} else {
				skill_id = self.card.NormalSkillID
			}
		} else {
			if self.temp_super_skill > 0 {
				skill_id = self.temp_super_skill
				self.use_temp_skill = true
			} else {
				skill_id = self.card.SuperSkillID
			}
		}
	} else {
		skill_id = trigger_skill
	}

	skill = skill_table_mgr.Get(skill_id)
	if skill == nil {
		log.Error("Cant get skill by id[%v]", skill_id)
		return
	}

	if trigger_skill > 0 && self.is_disable_attack() && skill.Type != SKILL_TYPE_PASSIVE {
		log.Debug("############# Team[%v] member[%v] disable combo skill[%v]", this.side, self.pos, trigger_skill)
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
		pos = skill_get_default_targets(self.pos, target_team, skill)
	} else if skill.SkillTarget == SKILL_TARGET_TYPE_BACK {
		pos = skill_get_back_targets(self.pos, target_team, skill)
	} else if skill.SkillTarget == SKILL_TARGET_TYPE_HP_MIN {
		pos = skill_get_hp_min_targets(self.pos, target_team, skill)
	} else if skill.SkillTarget == SKILL_TARGET_TYPE_RANDOM {
		pos = skill_get_random_targets(self.pos, target_team, skill)
	} else if skill.SkillTarget == SKILL_TARGET_TYPE_SELF {
		pos = skill_get_force_self_targets(self.pos, target_team, skill)
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

func (this *BattleTeam) UseSkillOnce(self_index int32, target_team *BattleTeam, trigger_skill int32) (skill *table_config.XmlSkillItem) {
	self := this.members[self_index]
	if self == nil || self.is_dead() {
		return nil
	}

	is_enemy, target_pos, skill := this.FindTargets(self, target_team, trigger_skill)
	if target_pos == nil {
		log.Warn("team[%v] member[%v] Cant find targets to attack", this.side, self_index)
		return nil
	}

	log.Debug("team[%v] member[%v] find is_enemy[%v] targets[%v] to use skill[%v]", this.side, self_index, is_enemy, target_pos, skill.Id)

	if !is_enemy {
		target_team = this
	}

	self.used_skill(skill)
	skill_effect(this, self_index, target_team, target_pos, skill)

	// 清除临时技能
	if self.use_temp_skill {
		if self.temp_normal_skill > 0 {
			log.Debug("!!!!!!!!!!!!!!!!!!! Team[%v] mem[%v] clear temp normal skill[%v]", this.side, self_index, self.temp_normal_skill)
			self.temp_normal_skill = 0
		} else if self.temp_super_skill > 0 {
			log.Debug("!!!!!!!!!!!!!!!!!!! Team[%v] mem[%v] clear temp super skill[%v]", this.side, self_index, self.temp_normal_skill)
			self.temp_super_skill = 0
		}
		self.use_temp_skill = false
	}

	// 是否有combo技能
	if skill.ComboSkill > 0 {
		r := this.GetLastReport()
		if r != nil {
			r.HasCombo = true
			log.Debug("########################################### Team[%v] member[%v] 后面有组合技 %v", this.side, self_index, skill.ComboSkill)
		}
	}

	return skill
}

func (this *BattleTeam) UseSkill(self_index int32, target_team *BattleTeam) int32 {
	mem := this.members[self_index]
	if mem == nil || mem.is_dead() || mem.is_will_dead() {
		return -1
	}
	for mem.get_use_skill() > 0 {
		if target_team.IsAllDead() {
			return 0
		}

		mem.act_done()

		if mem.is_disable_attack() {
			return 0
		}

		if mem.energy >= BATTLE_TEAM_MEMBER_MAX_ENERGY {
			// 被动技，怒气攻击前
			if mem.temp_super_skill == 0 {
				passive_skill_effect_with_self_pos(EVENT_BEFORE_RAGE_ATTACK, this, self_index, target_team, nil, true)
			}
		} else {
			// 被动技，普通攻击前
			if mem.temp_normal_skill == 0 {
				passive_skill_effect_with_self_pos(EVENT_BEFORE_NORMAL_ATTACK, this, self_index, target_team, nil, true)
			}
		}

		skill := this.UseSkillOnce(self_index, target_team, 0)
		if skill == nil {
			break
		}
		if skill.ComboSkill > 0 {
			log.Debug("@@@@@@!!!!!! Team[%v] member[%v] will use combo skill[%v]", this.side, self_index, skill.ComboSkill)
			this.UseSkillOnce(self_index, target_team, skill.ComboSkill)
		}
		this.DelaySkillEffect()
	}

	return 1
}

// 回合
func (this *BattleTeam) DoRound(target_team *BattleTeam) {
	this.RoundStart()
	target_team.RoundStart()

	// 被动技，回合行动前触发
	for i := int32(0); i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
		passive_skill_effect_with_self_pos(EVENT_BEFORE_ROUND, this, i, target_team, nil, false)
		passive_skill_effect_with_self_pos(EVENT_BEFORE_ROUND, target_team, i, this, nil, false)
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

// 结束
func (this *BattleTeam) OnFinish() {
	if this.members == nil {
		return
	}
	for i := 0; i < len(this.members); i++ {
		if this.members[i] != nil {
			this.members[i].on_battle_finish()
		}
	}
}

func (this *BattleTeam) GetLastReport() (last_report *msg_client_message.BattleReportItem) {
	if this.common_data == nil {
		return
	}

	l := len(this.common_data.reports)
	if l > 0 {
		last_report = this.common_data.reports[l-1]
	}
	return
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
					msg_battle_fighter_pool.Put(r.User)
					r.User = nil
				}
				// behiters
				if r.BeHiters != nil {
					for j := 0; j < len(r.BeHiters); j++ {
						if r.BeHiters[j] != nil {
							msg_battle_fighter_pool.Put(r.BeHiters[j])
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

// 人数
func (this *BattleTeam) MembersNum() (num int32) {
	if this.members == nil {
		return
	}
	for i := 0; i < len(this.members); i++ {
		if this.members[i] != nil && !this.members[i].is_dead() {
			num += 1
		}
	}
	return
}

func (this *BattleTeam) GetMembersEnergy() (energy map[int32]int32) {
	energy = make(map[int32]int32)
	for i := int32(0); i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
		if this.members[i] != nil && !this.members[i].is_dead() {
			energy[i] = this.members[i].energy
		}
	}
	return
}

// 开打
func (this *BattleTeam) Fight(target_team *BattleTeam, end_type int32, end_param int32) (is_win bool, enter_reports []*msg_client_message.BattleReportItem, rounds []*msg_client_message.BattleRoundReports) {
	round_max := end_param
	if end_type == BATTLE_END_BY_ALL_DEAD {
		round_max = BATTLE_ROUND_MAX_NUM
	} else if end_type == BATTLE_END_BY_ROUND_OVER {
	}

	// 存放战报
	if this.common_data == nil {
		this.common_data = &BattleCommonData{}
		this.common_data.Init()
	}
	target_team.common_data = this.common_data
	this.common_data.Reset()
	this.common_data.round_num = 0

	// 被动技，进场前触发
	for i := int32(0); i < BATTLE_TEAM_MEMBER_MAX_NUM; i++ {
		passive_skill_effect_with_self_pos(EVENT_ENTER_BATTLE, this, i, target_team, nil, false)
		passive_skill_effect_with_self_pos(EVENT_ENTER_BATTLE, target_team, i, this, nil, false)
	}

	if this.common_data.reports != nil {
		enter_reports = this.common_data.reports
		this.common_data.reports = make([]*msg_client_message.BattleReportItem, 0)
	}

	rand.Seed(time.Now().Unix())
	for c := int32(0); c < round_max; c++ {
		log.Debug("----------------------------------------------- Round[%v] --------------------------------------------", c+1)

		this.common_data.round_num += 1
		this.DoRound(target_team)

		round := msg_battle_round_reports_pool.Get()
		round.MyMembersEnergy = this.GetMembersEnergy()
		round.TargetMembersEnergy = target_team.GetMembersEnergy()
		round.Reports = this.common_data.reports
		round.RemoveBuffs = this.common_data.remove_buffs
		round.ChangedFighters = this.common_data.changed_fighters
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

		this.common_data.Reset()
	}

	this.OnFinish()
	target_team.OnFinish()

	return
}

func (this *BattleTeam) _format_members_for_msg() (members []*msg_client_message.BattleMemberItem) {
	for i := 0; i < len(this.members); i++ {
		if this.members[i] == nil {
			continue
		}
		mem := this.members[i].build_battle_member()
		mem.Side = this.side
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

// 延迟被动技
func (this *BattleTeam) PushDelaySkill(trigger_event int32, skill *table_config.XmlSkillItem, user *TeamMember, target_team *BattleTeam, trigger_pos []int32) {
	if this.common_data == nil {
		return
	}

	ds := delay_skill_pool.Get()
	ds.trigger_event = trigger_event
	ds.skill = skill
	ds.user = user
	ds.target_team = target_team
	ds.trigger_pos = trigger_pos
	ds.next = nil

	dl := this.common_data.delay_skill_list
	if dl == nil {
		dl = &DelaySkillList{}
		this.common_data.delay_skill_list = dl
	}
	if dl.head == nil {
		dl.head = ds
		dl.tail = ds
	} else {
		dl.tail.next = ds
		dl.tail = ds
	}

	log.Debug("############ Team[%v] member[%v] 推入了延迟被动技[%v]", user.team.side, user.pos, skill.Id)
}

// 处理延迟被动技
func (this *BattleTeam) DelaySkillEffect() {
	if this.common_data == nil {
		return
	}
	dl := this.common_data.delay_skill_list
	if dl == nil {
		return
	}

	c := 0
	d := dl.head
	for d != nil {
		//log.Debug("*@*@*@*@*@*@*@*@*@*@*@*@*@*@*@ [%v] To Delay Skill[%v] Effect trigger_event[%v] user[%v,%v] target_team[%v] trigger_pos[%v]",
		//	c+1, d.skill.Id, d.trigger_event, d.user.team.side, d.user.pos, d.target_team.side, d.trigger_pos)
		one_passive_skill_effect(d.trigger_event, d.skill, d.user, d.target_team, d.trigger_pos, true)
		//.Debug("*@*@*@*@*@*@*@*@*@*@*@*@*@*@*@ [%v] Delay Skill[%v] Effected trigger_event[%v] user[%v,%v] target_team[%v] trigger_pos[%v]",
		//	c+1, d.skill.Id, d.trigger_event, d.user.team.side, d.user.pos, d.target_team.side, d.trigger_pos)

		n := d.next
		delay_skill_pool.Put(d)
		d = n
		c += 1
	}
	dl.head = nil
	dl.tail = nil
}

// 是否有延迟技
func (this *BattleTeam) HasDelayTriggerEventSkill(trigger_event int32, behiter *TeamMember) bool {
	if this.common_data == nil {
		return false
	}
	dl := this.common_data.delay_skill_list
	if dl == nil {
		return false
	}
	d := dl.head
	for d != nil {
		if d.trigger_event == trigger_event && d.user == behiter {
			return true
		}
		d = d.next
	}
	return false
}

func C2SFightHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SBattleResultRequest
	err := proto.Unmarshal(msg_data, &req)
	if nil != err {
		log.Error("Unmarshal msg failed err(%s) !", err.Error())
		return -1
	}

	if p.Id == req.GetFightPlayerId() {
		log.Error("Cant fight with self")
		return -1
	}

	if req.GetAttackMembers() != nil && len(req.GetAttackMembers()) > 0 {
		if req.BattleType == 1 {
			res := p.SetAttackTeam(req.AttackMembers)
			if res < 0 {
				log.Error("Player[%v] set attack members[%v] failed", p.Id, req.AttackMembers)
				return res
			}
		} else if req.BattleType == 2 {
			res := p.SetCampaignTeam(req.AttackMembers)
			if res < 0 {
				log.Error("Player[%v] set campaign members[%v] failed", p.Id, req.AttackMembers)
				return res
			}
		} else {
			team_type := int32(-1)
			// 爬塔阵容
			if req.GetBattleType() == 3 {
				team_type = BATTLE_TOWER_TEAM
			} else {
				log.Error("Player[%v] set team[%v] invalid", p.Id, team_type)
				return -1
			}
			res := p.SetTeam(team_type, req.AttackMembers)
			if res < 0 {
				log.Error("Player[%v] set team[%v] failed", p.Id, team_type)
				return res
			}
		}
		p.send_teams()
	}

	var res int32
	if req.FightPlayerId > 0 {
		res = p.Fight2Player(req.FightPlayerId)
	} else if req.CampaignId > 0 {
		res = p.FightInCampaign(req.CampaignId)
	} else {
		if req.BattleType == 1 {
			res = p.Fight2Player(req.BattleParam)
		} else if req.BattleType == 2 {
			res = p.FightInCampaign(req.BattleParam)
		} else if req.BattleType == 3 {
			res = p.fight_tower(req.BattleParam)
		} else {
			res = -1
		}
	}

	if res > 0 {
		if req.BattleType == 1 {
			p.send_battle_team(BATTLE_ATTACK_TEAM, req.GetAttackMembers())
		} else if req.BattleType == 2 {
			p.send_battle_team(BATTLE_CAMPAIN_TEAM, req.GetAttackMembers())
		} else if req.BattleType == 3 {
			p.send_battle_team(BATTLE_TOWER_TEAM, req.GetAttackMembers())
		}
	}

	return res
}

func (this *Player) send_battle_team(tt int32, team_members []int32) {
	response := &msg_client_message.S2CSetTeamResponse{}
	response.TeamType = tt
	response.TeamMembers = team_members
	this.Send(uint16(msg_client_message_id.MSGID_S2C_SET_TEAM_RESPONSE), response)
}

func C2SSetTeamHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SSetTeamRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s) !", err.Error())
		return -1
	}

	var res int32
	tt := req.GetTeamType()
	if tt == 0 {
		res = p.SetAttackTeam(req.TeamMembers)
	} else if tt == 1 {
		res = p.SetDefenseTeam(req.TeamMembers)
	} else {
		log.Warn("Unknown team type[%v] to player[%v]", tt, p.Id)
	}

	p.send_battle_team(tt, req.TeamMembers)

	return res
}

func C2SSetHangupCampaignHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SBattleSetHangupCampaignRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s) !", err.Error())
		return -1
	}

	res := p.set_hangup_campaign_id(req.GetCampaignId())
	if res < 0 {
		log.Debug("Player[%v] set hangup campaign %v failed[%v]", p.Id, req.GetCampaignId(), res)
		return res
	}

	response := &msg_client_message.S2CBattleSetHangupCampaignResponse{}
	response.CampaignId = req.GetCampaignId()
	p.Send(uint16(msg_client_message_id.MSGID_S2C_BATTLE_SET_HANGUP_CAMPAIGN_RESPONSE), response)

	log.Debug("Player[%v] set hangup campaign %v success", p.Id, req.GetCampaignId())

	return 1
}

func C2SCampaignHangupIncomeHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SCampaignHangupIncomeRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s) !", err.Error())
		return -1
	}

	t := req.GetIncomeType()
	p.hangup_income_get(t, false)
	return 1
}

func C2SCampaignDataHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SCampaignDataRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s) !", err.Error())
		return -1
	}
	p.send_campaigns()
	return 1
}
