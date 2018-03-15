package main

import (
	"libs/log"
	"math/rand"
)

const (
	BATTLE_TEAM_MEMBER_INIT_ENERGY       = 1
	BATTLE_TEAM_MEMBER_MAX_ENERGY        = 4
	BATTLE_TEAM_MEMBER_ADD_ENERGY        = 2
	BATTLE_TEAM_MEMBER_MAX_NUM           = 9
	BATTLE_FORMATION_LINE_NUM            = 3
	BATTLE_FORMATION_ONE_LINE_MEMBER_NUM = 3
)

type AttrItem struct {
	id    int32
	value int32
	add   int32
}

type TeamMember struct {
	id      int32
	hp      int32
	energy  int32
	attack  int32
	defense int32
}

type BattleTeam struct {
	curr_attack int32
	members     []*TeamMember
	in_chess    int32
}

type ReportItem struct {
	attacker    *TeamMember
	be_attacker *TeamMember
	skill_id    int32
	damage      int32
	next        *ReportItem
}

func (this *Player) SetAttackTeam(team []int32) bool {
	if team == nil {
		return false
	}
	for i := 0; i < len(team); i++ {
		if !this.db.Roles.HasIndex(team[i]) {
			log.Warn("Player[%v] not has role[%v] for set attack team", this.Id, team[i])
			return false
		}
	}
	this.db.BattleTeam.SetAttackMembers(team)
	return true
}

func (this *Player) SetDefenseTeam(team []int32) bool {
	if team == nil {
		return false
	}
	for i := 0; i < len(team); i++ {
		if !this.db.Roles.HasIndex(team[i]) {
			log.Warn("Player[%v] not has role[%v] for set defense team", this.Id, team[i])
			return false
		}
	}
	this.db.BattleTeam.SetDefenseMembers(team)
	return true
}

func (this *Player) Fight2Player(player_id int32) {

}

func BattleRound(team []*BattleTeam) (report *ReportItem) {
	for pos := 0; pos < BATTLE_TEAM_MEMBER_MAX_NUM; pos++ {
		m1 := team[0].members[pos]
		if m1 == nil {
			continue
		}
		// find opponent
		var m2 *TeamMember
		if team[1].members[pos] != nil {
			m2 = team[1].members[pos]
		} else {
			for i := 0; i < BATTLE_FORMATION_LINE_NUM; i++ {
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
			}
		}
		if m2 == nil {
			return
		}

		if m1.attack <= m2.defense {

		}
	}
	return
}

func BattleGenerateReport(team1 *BattleTeam, team2 *BattleTeam) (report *ReportItem) {
	team := make([]*BattleTeam, 2)
	if team1.in_chess > team2.in_chess {
		team[0] = team1
		team[1] = team2
	} else if team1.in_chess < team2.in_chess {
		team[1] = team2
		team[0] = team1
	} else {
		if rand.Intn(2) == 0 {
			team[0] = team1
			team[1] = team2
		} else {
			team[1] = team2
			team[0] = team1
		}
	}

	return
}
