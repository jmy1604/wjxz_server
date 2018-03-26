package main

import (
	_ "libs/log"
	"public_message/gen_go/client_message"
)

func (this *TeamMember) build_battle_item(pos int32, damage int32) *msg_client_message.BattleMemberItem {
	item := msg_battle_member_item_pool.Get()
	item.Pos = pos
	item.HP = this.hp
	item.MaxHP = this.attrs[ATTR_HP_MAX]
	item.Damage = damage
	item.Energy = this.energy
	offset := int32(0)
	for i := 0; i < len(this.bufflist_arr); i++ {
		b := this.bufflist_arr[i].head
		c := int32(0)
		for b != nil {
			n := b.next
			if item.BuffIds == nil {
				item.BuffIds = []int32{b.buff.Id}
				c += 1
			} else {
				for j := int(offset); j < len(item.BuffIds); j++ {
					if item.BuffIds[j] != b.buff.Id {
						item.BuffIds = append(item.BuffIds, b.buff.Id)
						c += 1
					}
				}
			}
			b = n
		}
		offset += c
	}

	return item
}

func build_battle_report_item(self_team *BattleTeam, self_pos int32, self_damage int32, skill_id int32, is_passive, is_block, is_critical bool) *msg_client_message.BattleReportItem {
	item := msg_battle_reports_item_pool.Get()
	item.Side = self_team.side
	item.SkillId = skill_id
	self := self_team.members[self_pos]
	item.User = self.build_battle_item(self_pos, self_damage)
	item.IsPassive = is_passive
	item.IsBlock = is_block
	item.IsCritical = is_critical

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
