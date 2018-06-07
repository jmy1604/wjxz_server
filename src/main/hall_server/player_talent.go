package main

import (
	"libs/log"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"

	"github.com/golang/protobuf/proto"
)

func (this *Player) get_talent_list() []*msg_client_message.TalentInfo {
	all := this.db.Talents.GetAllIndex()
	if all == nil || len(all) == 0 {
		return make([]*msg_client_message.TalentInfo, 0)
	}

	var talents []*msg_client_message.TalentInfo
	for i := 0; i < len(all); i++ {
		lvl, o := this.db.Talents.GetLevel(all[i])
		if !o {
			continue
		}
		talents = append(talents, &msg_client_message.TalentInfo{
			Id:    all[i],
			Level: lvl,
		})
	}
	return talents
}

func (this *Player) up_talent(talent_id int32) int32 {
	level, _ := this.db.Talents.GetLevel(talent_id)
	talent := talent_table_mgr.GetByIdLevel(talent_id, level)
	if talent == nil {
		log.Error("Talent[%v,%v] data not found", talent_id, level)
		return int32(msg_client_message.E_ERR_PLAYER_TALENT_NOT_FOUND)
	}

	// check cost
	for i := 0; i < len(talent.Next.UpgradeCost)/2; i++ {
		rid := talent.Next.UpgradeCost[2*i]
		rct := talent.Next.UpgradeCost[2*i+1]
		if this.get_resource(rid) < rct {
			log.Error("Player[%v] up talent[%v] not enough resource[%v]", this.Id, talent_id, rid)
			return int32(msg_client_message.E_ERR_PLAYER_TALENT_UP_NOT_ENOUGH_RESOURCE)
		}
	}

	// cost resource
	for i := 0; i < len(talent.Next.UpgradeCost)/2; i++ {
		rid := talent.Next.UpgradeCost[2*i]
		rct := talent.Next.UpgradeCost[2*i+1]
		this.add_resource(rid, -rct)
	}

	if level == 0 {
		level += 1
		this.db.Talents.Add(&dbPlayerTalentData{
			Id:    talent_id,
			Level: level,
		})
	} else {
		level += 1
		this.db.Talents.SetLevel(talent_id, level)
	}

	response := &msg_client_message.S2CTalentUpResponse{
		TalentId: talent_id,
		Level:    level,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_TALENT_UP_RESPONSE), response)

	return 1
}

func (this *Player) add_talent_attr(team *BattleTeam) {
	if team.members == nil {
		return
	}

	all_tid := this.db.Talents.GetAllIndex()
	if all_tid == nil {
		return
	}

	for i := 0; i < len(all_tid); i++ {
		lvl, _ := this.db.Talents.GetLevel(all_tid[i])
		t := talent_table_mgr.GetByIdLevel(all_tid[i], lvl)
		if t == nil {
			continue
		}
		for j := 0; j < len(team.members); j++ {
			m := team.members[j]
			if m != nil && !m.is_dead() {
				if !_skill_check_cond(m, t.TalentEffectCond) {
					continue
				}
				m.add_attrs(t.TalentAttr)
				for k := 0; k < len(t.TalentSkillList); k++ {
					m.add_skill_attr(t.TalentSkillList[k])
				}
			}
		}
	}
}

func C2STalentListHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2STalentListRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}

	talents := p.get_talent_list()
	response := &msg_client_message.S2CTalentListResponse{
		Talents: talents,
	}
	p.Send(uint16(msg_client_message_id.MSGID_S2C_TALENT_LIST_RESPONSE), response)
	return 1
}

func C2STalentUpHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2STalentUpRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}
	return p.up_talent(req.GetTalentId())
}
