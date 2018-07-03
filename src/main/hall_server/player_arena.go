package main

import (
	"libs/log"
	"libs/utils"
	"main/table_config"
	_ "math/rand"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	"sync/atomic"
	"time"

	"github.com/golang/protobuf/proto"
)

const (
	ARENA_RANK_MAX = 100000
)

type ArenaRankItem struct {
	SaveTime    int32
	PlayerScore int32
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
	this.PlayerScore = n.PlayerScore
	this.SaveTime = n.SaveTime
}

func (this *ArenaRankItem) CopyDataTo(node interface{}) {
	n := node.(*ArenaRankItem)
	if n == nil {
		return
	}
	n.PlayerId = this.PlayerId
	n.PlayerScore = this.PlayerScore
	n.SaveTime = this.SaveTime
}

type ArenaRobot struct {
	robot_data   *table_config.XmlArenaRobotItem
	defense_team *BattleTeam
	power        int32
}

func (this *ArenaRobot) Init(robot *table_config.XmlArenaRobotItem) {
	this.robot_data = robot
	this.defense_team = &BattleTeam{}
	this._calculate_power()
}

func (this *ArenaRobot) _calculate_power() {
	card_list := this.robot_data.RobotCardList
	if card_list == nil {
		return
	}

	for i := 0; i < len(card_list); i++ {
		for j := 0; j < len(card_list[i].EquipID); j++ {
			equip_item := item_table_mgr.Get(card_list[i].EquipID[j])
			if equip_item == nil {
				continue
			}
			this.power += equip_item.BattlePower
		}
	}
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

	now_time := int32(time.Now().Unix())
	this.robots = make(map[int32]*ArenaRobot)
	for _, r := range array {
		robot := &ArenaRobot{}
		robot.Init(r)
		this.robots[r.Id] = robot
		// 加入排行榜
		var d = ArenaRankItem{
			SaveTime:    now_time,
			PlayerScore: r.RobotScore,
			PlayerId:    r.Id,
		}
		rank_list_mgr.UpdateItem(RANK_LIST_TYPE_ARENA, &d)
	}
}

func (this *ArenaRobotManager) Get(robot_id int32) *ArenaRobot {
	return this.robots[robot_id]
}

func (this *Player) _update_arena_score(data *ArenaRankItem) {
	rank_list_mgr.UpdateItem(RANK_LIST_TYPE_ARENA, data)
}

func (this *Player) LoadArenaScore() {
	score := this.db.Arena.GetScore()
	if score <= 0 {
		return
	}
	var data = ArenaRankItem{
		SaveTime:    this.db.Arena.GetUpdateScoreTime(),
		PlayerScore: score,
		PlayerId:    this.Id,
	}
	this._update_arena_score(&data)
}

func (this *Player) UpdateArenaScore(is_win bool) bool {
	var add_score int32
	now_score := this.db.Arena.GetScore()
	division := arena_division_table_mgr.GetByScore(now_score)
	if division == nil {
		log.Error("Arena division table data with score[%v] is not found", now_score)
		return false
	}

	if is_win {
		add_score = division.WinScore
		if this.db.Arena.GetRepeatedWinNum() >= global_config_mgr.GetGlobalConfig().ArenaRepeatedWinNum {
			add_score += division.WinningStreakScoreBonus
		}
	} else {
		add_score = division.LoseScore
	}

	if add_score > 0 {
		now_time := int32(time.Now().Unix())
		score := this.db.Arena.IncbyScore(add_score)
		this.db.Arena.SetUpdateScoreTime(now_time)
		top_score := this.db.Arena.GetHistoryTopRank()
		if score > top_score {
			this.db.Arena.SetHistoryTopRank(score)
		}

		var data = ArenaRankItem{
			SaveTime:    now_time,
			PlayerScore: score,
			PlayerId:    this.Id,
		}
		this._update_arena_score(&data)
	}

	return true
}

func (this *Player) OutputArenaRankItems(rank_start, rank_num int32) {
	rank_items, self_rank, self_value := rank_list_mgr.GetItemsByRange(RANK_LIST_TYPE_ARENA, this.Id, rank_start, rank_num)
	if rank_items == nil {
		log.Warn("Player[%v] get rank list with range[%v,%v] failed", this.Id, rank_start, rank_num)
		return
	}

	l := int32(len(rank_items))
	for rank := rank_start; rank < l; rank++ {
		item := (rank_items[rank-rank_start]).(*ArenaRankItem)
		if item == nil {
			log.Error("Player[%v] get arena rank list by rank[%v] item failed")
			continue
		}
		log.Debug("Rank: %v   Player[%v] Score[%v]", rank, item.PlayerId, item.PlayerScore)
	}

	if self_value != nil && self_rank > 0 {
		log.Debug("Player[%v] score[%v] rank[%v]", this.Id, self_value.(int32), self_rank)
	}
}

// 匹配对手
func (this *Player) MatchArenaPlayer() (player_id int32) {
	rank := rank_list_mgr.GetRankByPlayerId(RANK_LIST_TYPE_ARENA, this.Id)
	if rank < 0 {
		log.Error("Player[%v] get arena rank list rank failed", this.Id)
		return
	}

	var start_rank, rank_num, last_rank int32
	match_num := global_config_mgr.GetGlobalConfig().ArenaMatchPlayerNum
	if rank == 0 {
		start_rank, rank_num = rank_list_mgr.GetLastRankRange(RANK_LIST_TYPE_ARENA, match_num)
		if start_rank < 0 {
			log.Error("Player[%v] match arena player failed", this.Id)
			return
		}
	} else {
		last_rank, _ = rank_list_mgr.GetLastRankRange(RANK_LIST_TYPE_ARENA, 1)
		half_num := match_num / 2
		var left, right int32
		if this.db.Arena.GetRepeatedWinNum() >= global_config_mgr.GetGlobalConfig().ArenaRepeatedWinNum {
			right = rank - 1
			left = rank - match_num
		} else if this.db.Arena.GetRepeatedLoseNum() >= global_config_mgr.GetGlobalConfig().ArenaLoseRepeatedNum {
			right = rank + match_num
			left = rank + 1
		} else {
			right = rank + half_num
			left = rank - half_num
		}

		if left < 1 {
			left = 1
		}
		if right > last_rank {
			right = last_rank
		}

		start_rank = left
		rank_num = right - start_rank + 1
	}

	_, r := rand31n_from_range(start_rank, start_rank+rank_num)
	// 如果是自己
	if rank == r {
		r += 1
		if r > last_rank {
			r -= 2
		}
	}
	item := rank_list_mgr.GetItemByRank(RANK_LIST_TYPE_ARENA, r)
	if item == nil {
		log.Error("Player[%v] match arena player by random rank[%v] get empty item", this.Id, r)
		return
	}

	player_id = item.(*ArenaRankItem).PlayerId

	log.Debug("Player[%v] match arena players rank range [start:%v, num:%v], rand the rank %v, match player[%v]", this.Id, start_rank, rank_num, r, player_id)

	return
}

func (this *Player) send_arena_data() int32 {
	response := &msg_client_message.S2CArenaDataResponse{}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ARENA_DATA_RESPONSE), response)
	log.Debug("Player[%v] arena data: %v", this.Id, response)
	return 1
}

func (this *Player) arena_player_defense_team(player_id int32) int32 {
	p := player_mgr.GetPlayerById(player_id)
	if p == nil {
		log.Error("Player[%v] not found", player_id)
		return int32(msg_client_message.E_ERR_PLAYER_NOT_EXIST)
	}
	defense_team := p.db.BattleTeam.GetDefenseMembers()
	team := make(map[int32]*msg_client_message.PlayerTeamRole)
	if defense_team != nil {
		for i := 0; i < len(defense_team); i++ {
			m := defense_team[i]
			if m <= 0 {
				continue
			}
			table_id, _ := this.db.Roles.GetTableId(m)
			level, _ := this.db.Roles.GetLevel(m)
			team[m] = &msg_client_message.PlayerTeamRole{
				TableId: table_id,
				Pos:     int32(i),
				Level:   level,
			}
		}
	}
	response := &msg_client_message.S2CArenaPlayerDefenseTeamResponse{
		DefenseTeam: team,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ARENA_PLAYER_DEFENSE_TEAM_RESPONSE), response)
	log.Debug("Player[%v] get arena player[%v] defense team[%v]", this.Id, player_id, team)
	return 1
}

func (this *Player) arena_match() int32 {
	pid := this.MatchArenaPlayer()

	var robot *ArenaRobot
	p := player_mgr.GetPlayerById(pid)
	if p == nil {
		robot = arena_robot_mgr.Get(pid)
		if robot == nil {
			log.Error("Player[%v] matched id[%v] is not player and not robot", this.Id, pid)
			return int32(msg_client_message.E_ERR_PLAYER_ARENA_MATCH_PLAYER_FAILED)
		}
	}

	// 当前匹配到的玩家
	this.db.Arena.SetMatchedPlayerId(pid)

	var level, head, score, power, grade int32
	var name string
	if p != nil {
		name = p.db.GetName()
		level = p.db.Info.GetLvl()
		score = p.db.Arena.GetScore()
		head = 0
		power = p.get_defense_team_power()
	} else {
		name = robot.robot_data.RobotName
		level = 1
		score = robot.robot_data.RobotScore
		head = robot.robot_data.RobotHead
		power = robot.power
	}

	response := &msg_client_message.S2CArenaMatchPlayerResponse{
		PlayerId:    pid,
		PlayerName:  name,
		PlayerLevel: level,
		PlayerHead:  head,
		PlayerScore: score,
		PlayerGrade: grade,
		PlayerPower: power,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ARENA_MATCH_PLAYER_RESPONSE), response)
	log.Debug("Player[%v] matched arena player[%v]", this.Id, response)
	return 1
}

func C2SArenaDataHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SArenaDataRequest
	err := proto.Unmarshal(msg_data, &req)
	if nil != err {
		log.Error("Unmarshal msg failed err(%s)", err.Error())
		return -1
	}
	return p.send_arena_data()
}

func C2SArenaPlayerDefenseTeamHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SArenaPlayerDefenseTeamRequest
	err := proto.Unmarshal(msg_data, &req)
	if nil != err {
		log.Error("Unmarshal msg failed err(%s)", err.Error())
		return -1
	}
	return p.arena_player_defense_team(req.GetPlayerId())
}

func C2SArenaMatchPlayerHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SArenaMatchPlayerRequest
	err := proto.Unmarshal(msg_data, &req)
	if nil != err {
		log.Error("Unmarshal msg failed err(%s)", err.Error())
		return -1
	}
	return p.arena_match()
}

// 竞技场赛季管理
type ArenaSeasonMgr struct {
	state      int32 // 0 结束  1 开始
	start_time int32
}

var arena_season_mgr ArenaSeasonMgr

func (this *ArenaSeasonMgr) Init() {

}

func (this *ArenaSeasonMgr) Start() bool {
	return atomic.CompareAndSwapInt32(&this.state, 0, 1)
}

func (this *ArenaSeasonMgr) End() bool {
	return atomic.CompareAndSwapInt32(&this.state, 1, 0)
}

func (this *ArenaSeasonMgr) IsStart() bool {
	return atomic.LoadInt32(&this.state) == 1
}

func (this *ArenaSeasonMgr) IsEnd() bool {
	return atomic.LoadInt32(&this.state) == 0
}

func (this *ArenaSeasonMgr) GetRemainSeconds() (day_remain int32, season_remain int32) {
	now_time := int32(time.Now().Unix())
	day_set := dbc.ArenaSeason.GetRow().Data.GetLastDayResetTime()
	if day_set == 0 {
		dbc.ArenaSeason.GetRow().Data.SetLastDayResetTime(now_time)
		day_set = now_time
	}
	season_set := dbc.ArenaSeason.GetRow().Data.GetLastSeasonResetTime()
	if season_set == 0 {
		dbc.ArenaSeason.GetRow().Data.SetLastSeasonResetTime(now_time)
		season_set = now_time
	}

	day_time_string := global_config_mgr.GetGlobalConfig().ArenaSeasonDayResetTime
	day_remain = utils.GetRemainSeconds2NextDayTime(day_set, day_time_string)
	days := global_config_mgr.GetGlobalConfig().ArenaSeasonDays
	season_remain = utils.GetRemainSeconds2NextSeveralDaysTime(season_set, day_time_string, days)

	return
}

func (this *ArenaSeasonMgr) Reset() {
	for {
		if atomic.LoadInt32(&this.state) == 1 {
			time.Sleep(time.Second * 1)
			continue
		}
		break
	}

	rank_list := rank_list_mgr.GetRankList(RANK_LIST_TYPE_ARENA)
	if rank_list == nil {
		return
	}
	rank_num := rank_list.RankNum()
	for rank := int32(1); rank <= rank_num; rank++ {
		item := rank_list.GetItemByRank(rank)
		if item == nil {
			log.Warn("Cant found rank[%v] item in arena rank list with reset", rank)
			continue
		}
		arena_item := item.(*ArenaRankItem)
		if arena_item == nil {
			continue
		}
		division := arena_division_table_mgr.GetByScore(arena_item.PlayerScore)
		if division == nil {
			continue
		}
		arena_item.PlayerScore = division.NewSeasonScore
		p := player_mgr.GetPlayerById(arena_item.PlayerId)
		if p != nil {
			p.db.Arena.SetScore(arena_item.PlayerScore)
		}
	}
	atomic.StoreInt32(&this.state, 0)
}

func (this *ArenaSeasonMgr) Run() {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	for {
		// 检测时间
		day_remain, season_remain := this.GetRemainSeconds()
		if day_remain < 0 {
			log.Warn("Arena season check day reset time error")
			time.Sleep(time.Second * 2)
			continue
		}
		if season_remain < 0 {
			log.Warn("Arena season check season reset time error")
			time.Sleep(time.Second * 2)
			continue
		}

		now_time := int32(time.Now().Unix())

		// 每天领奖
		if day_remain == 0 {
			dbc.ArenaSeason.GetRow().Data.SetLastDayResetTime(now_time)
		}
		// 赛季结束，发奖重置积分
		if season_remain == 0 {
			dbc.ArenaSeason.GetRow().Data.SetLastSeasonResetTime(now_time)
			this.Reset()
		}

		time.Sleep(time.Second * 2)
	}
}
