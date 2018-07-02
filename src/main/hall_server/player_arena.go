package main

import (
	"libs/log"
	"libs/utils"
	"main/table_config"
	_ "math/rand"
	_ "net/http"
	_ "public_message/gen_go/client_message"
	_ "public_message/gen_go/client_message_id"
	_ "sync"
	"time"

	_ "github.com/golang/protobuf/proto"
)

const (
	ARENA_RANK_MAX = 100000
)

type ArenaRankItem struct {
	SaveTime    int32
	PlayerScore int32
	PlayerLevel int32
	PlayerId    int32
}

func (this *ArenaRankItem) Less(value interface{}) bool {
	item := value.(*ArenaRankItem)
	if item == nil {
		return false
	}
	if this.PlayerScore < item.PlayerScore {
		return true
	}
	if this.PlayerScore == item.PlayerScore {
		if this.SaveTime > item.SaveTime {
			return true
		}
		if this.SaveTime == item.SaveTime {
			if this.PlayerLevel < item.PlayerLevel {
				return true
			}
		}
	}
	return false
}

func (this *ArenaRankItem) Greater(value interface{}) bool {
	item := value.(*ArenaRankItem)
	if item == nil {
		return false
	}
	if this.PlayerScore > item.PlayerScore {
		return true
	}
	if this.PlayerScore == item.PlayerScore {
		if this.SaveTime < item.SaveTime {
			return true
		}
		if this.SaveTime == item.SaveTime {
			if this.PlayerLevel > item.PlayerLevel {
				return true
			}
		}
	}
	return false
}

func (this *ArenaRankItem) KeyEqual(value interface{}) bool {
	item := value.(*ArenaRankItem)
	if item == nil {
		return false
	}
	if item == nil {
		return false
	}
	if this.PlayerId == item.PlayerId {
		return true
	}
	return false
}

func (this *ArenaRankItem) GetKey() interface{} {
	return this.PlayerId
}

func (this *ArenaRankItem) GetValue() interface{} {
	return this.PlayerScore
}

func (this *ArenaRankItem) New() utils.SkiplistNode {
	return &ArenaRankItem{}
}

func (this *ArenaRankItem) Assign(node utils.SkiplistNode) {
	n := node.(*ArenaRankItem)
	if n == nil {
		return
	}
	this.PlayerId = n.PlayerId
	this.PlayerLevel = n.PlayerLevel
	this.PlayerScore = n.PlayerScore
	this.SaveTime = n.SaveTime
}

func (this *ArenaRankItem) CopyDataTo(node interface{}) {
	return
}

type ArenaRobot struct {
	robot_data   *table_config.XmlArenaRobotItem
	defense_team *BattleTeam
}

func (this *ArenaRobot) Init(robot *table_config.XmlArenaRobotItem) {
	this.robot_data = robot
	this.defense_team = &BattleTeam{}
}

type ArenaRobotManager struct {
	robots map[int32]*ArenaRobot
}

var arena_robot_mgr ArenaRobotManager

func (this *ArenaRobotManager) Init() {
	array := arena_robot_table_mgr.Array
	if array == nil {
		return
	}

	this.robots = make(map[int32]*ArenaRobot)
	for _, r := range array {
		robot := &ArenaRobot{}
		robot.Init(r)
		this.robots[r.Id] = robot
	}
}

func (this *ArenaRobotManager) Get(robot_id int32) *ArenaRobot {
	return this.robots[robot_id]
}

func (this *Player) _update_arena_score(data *ArenaRankItem) {
	rank_list_mgr.UpdateItem(RANK_LIST_TYPE_ARENA, data)
}

func (this *Player) LoadArenaScore() {
	var data = ArenaRankItem{
		SaveTime:    this.db.Arena.GetUpdateScoreTime(),
		PlayerScore: this.db.Arena.GetScore(),
		PlayerLevel: this.db.Info.GetLvl(),
		PlayerId:    this.Id,
	}
	this._update_arena_score(&data)
}

func (this *Player) UpdateArenaScore(is_win bool) {
	var add_score int32
	if is_win {
		add_score = global_config_mgr.GetGlobalConfig().ArenaWinAddScore
		if this.db.Arena.GetRepeatedWinNum() >= global_config_mgr.GetGlobalConfig().ArenaRepeatedWinNum {
			add_score += global_config_mgr.GetGlobalConfig().ArenaRepeatedWinAddExtraScore
		}
	} else {
		add_score = global_config_mgr.GetGlobalConfig().ArenaLoseAddScoreOnLowGrade
	}

	if add_score > 0 {
		now_time := int32(time.Now().Unix())
		score := this.db.Arena.IncbyScore(add_score)
		this.db.Arena.SetUpdateScoreTime(now_time)

		var data = ArenaRankItem{
			SaveTime:    now_time,
			PlayerScore: score,
			PlayerLevel: this.db.Info.GetLvl(),
			PlayerId:    this.Id,
		}
		this._update_arena_score(&data)
	}
}

func (this *Player) OutputArenaRankItems(rank_start, rank_num int32) {
	rank_items, self_rank, self_value := rank_list_mgr.GetItemsByRange(RANK_LIST_TYPE_ARENA, this.Id, rank_start, rank_num)
	if rank_items == nil {
		log.Warn("Player[%v] get rank list with range[%v,%v] failed", this.Id, rank_start, rank_num)
		return
	}

	for rank := rank_start; rank < rank_start+rank_num; rank++ {
		item := rank_items[rank-rank_start]
		pid := item.GetKey().(int32)
		score := item.GetValue().(int32)
		log.Debug("Rank: %v   Player[%v] Score[%v]", rank, pid, score)
	}

	log.Debug("Player[%v] score[%v] rank[%v]", this.Id, self_value.(int32), self_rank)
}

// 匹配对手
func (this *Player) MatchArenaPlayer() (p *Player) {
	rank := rank_list_mgr.GetRankByPlayerId(RANK_LIST_TYPE_ARENA, this.Id)
	if rank < 0 {
		log.Error("Player[%v] get arena rank list rank failed", this.Id)
		return
	}

	var start_rank, rank_num int32
	match_num := global_config_mgr.GetGlobalConfig().ArenaMatchPlayerNum
	if rank == 0 {
		start_rank, rank_num = rank_list_mgr.GetLastRankRange(RANK_LIST_TYPE_ARENA, match_num)
		if start_rank < 0 {
			log.Error("Player[%v] match arena player failed", this.Id)
			return
		}
	} else {
		last_rank, _ := rank_list_mgr.GetLastRankRange(RANK_LIST_TYPE_ARENA, 1)
		half_num := match_num / 2
		// 一般情况
		right_need := (rank + half_num) - last_rank
		if right_need > 0 {
			rank_num = last_rank - rank
		} else {
			rank_num = half_num
		}
		left_need := rank - half_num + 1
		if left_need <= 0 {
			rank_num += half_num
		} else {
			rank_num += (half_num - left_need)
		}
		if right_need > 0 && left_need < 0 {
			if right_need < -left_need {

			}
		} else if right_need < 0 && left_need > 0 {

		}
	}

	_, r := rand31n_from_range(start_rank, start_rank+rank_num)
	item := rank_list_mgr.GetItemByRank(RANK_LIST_TYPE_ARENA, r)
	if item == nil {
		log.Error("Player[%v] match arena player by random rank[%v] get empty item", this.Id, r)
		return
	}

	player_id := item.(*ArenaRankItem).PlayerId
	p = player_mgr.GetPlayerById(player_id)
	return
}
