package main

import (
	_ "libs/log"
	"public_message/gen_go/client_message"
)

func (this *TeamMember) build_battle_item(pos int32, damage int32) *msg_client_message.BattleMemberItem {
	item := msg_battle_member_item_pool.Get()
	item.Side = this.team.side
	item.Id = this.id
	item.TableId = this.card.ClientId
	item.Pos = pos
	item.HP = this.hp
	item.MaxHP = this.attrs[ATTR_HP_MAX]
	item.Damage = damage
	item.Energy = this.energy
	return item
}

func build_battle_report_item(self_team *BattleTeam, self_pos int32, self_damage int32, skill_id int32, is_block, is_critical bool) *msg_client_message.BattleReportItem {
	item := msg_battle_reports_item_pool.Get()
	item.Side = self_team.side
	item.SkillId = skill_id
	self := self_team.members[self_pos]
	item.User = self.build_battle_item(self_pos, self_damage)
	item.BeHiters = make([]*msg_client_message.BattleMemberItem, 0)
	item.AddBuffs = make([]*msg_client_message.BattleMemberBuff, 0)
	item.RemoveBuffs = make([]*msg_client_message.BattleMemberBuff, 0)
	item.IsBlock = is_block
	item.IsCritical = is_critical
	item.HasCombo = false
	return item
}

func build_battle_report_item_add_target_item(item *msg_client_message.BattleReportItem, target_team *BattleTeam, target_pos int32, target_damage int32) {
	if item == nil {
		return
	}
	target := target_team.members[target_pos]
	if target == nil {
		return
	}
	mem_item := target.build_battle_item(target_pos, target_damage)

	if item.BeHiters == nil {
		item.BeHiters = []*msg_client_message.BattleMemberItem{mem_item}
	} else {
		item.BeHiters = append(item.BeHiters, mem_item)
	}
}

func build_battle_report_add_buff(item *msg_client_message.BattleReportItem, target_team *BattleTeam, target_pos int32, buff_id int32) {
	if item == nil {
		return
	}
	buff := msg_battle_buff_item_pool.Get()
	buff.Side = target_team.side
	buff.Pos = target_pos
	buff.BuffId = buff_id
	mem := target_team.members[target_pos]
	buff.MemId = mem.id
	if item.AddBuffs == nil {
		item.AddBuffs = []*msg_client_message.BattleMemberBuff{buff}
	} else {
		item.AddBuffs = append(item.AddBuffs, buff)
	}
}

func build_battle_report_remove_buff(item *msg_client_message.BattleReportItem, target_team *BattleTeam, target_pos int32, buff_id int32) {
	if item == nil {
		return
	}
	buff := msg_battle_buff_item_pool.Get()
	buff.Side = target_team.side
	buff.Pos = target_pos
	buff.BuffId = buff_id
	mem := target_team.members[target_pos]
	buff.MemId = mem.id
	if item.RemoveBuffs == nil {
		item.RemoveBuffs = []*msg_client_message.BattleMemberBuff{buff}
	} else {
		item.RemoveBuffs = append(item.RemoveBuffs, buff)
	}
}
