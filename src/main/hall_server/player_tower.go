package main

import (
	"libs/log"
	_ "main/table_config"
	_ "math/rand"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"

	"time"

	"github.com/golang/protobuf/proto"
)

func get_tower_fight_id(tower_id, i int32) int32 {
	return tower_id*10 + i
}

func (this *Player) get_tower_data() int32 {
	curr_id := this.db.TowerCommon.GetCurrId()
	towers := this.db.Towers.GetAllIndex()
	if towers == nil {
		towers = make([]int32, 0)
	}
	response := &msg_client_message.S2CTowerDataResponse{
		CurrTowerId: curr_id,
		TowerIds:    towers,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_TOWER_DATA_RESPONSE), response)
	return 1
}

func (this *Player) fight_tower(tower_id int32) int32 {
	// 是否时当前层的下一层
	/////////////////////////
	if this.db.Towers.HasIndex(tower_id) {
		log.Error("Player[%v] already fighted tower[%v]", this.Id, tower_id)
		return int32(msg_client_message.E_ERR_PLAYER_TOWER_ALREADY_FIGHTED)
	}

	stage_id := int32(1)
	stage := stage_table_mgr.Get(stage_id)
	if stage == nil {
		log.Error("Tower[%v] stage[%v] not found", tower_id, stage_id)
		return int32(msg_client_message.E_ERR_PLAYER_TOWER_NOT_FOUND)
	}

	is_win, my_team, target_team, enter_reports, rounds, _ := this.FightInStage(stage)
	this.db.TowerCommon.SetCurrId(tower_id)
	this.db.Towers.Add(&dbPlayerTowerData{
		Id: tower_id,
	})

	response := &msg_client_message.S2CBattleResultResponse{
		IsWin:        is_win,
		MyTeam:       my_team,
		TargetTeam:   target_team,
		EnterReports: enter_reports,
		Rounds:       rounds,
	}
	data := this.Send(uint16(msg_client_message_id.MSGID_S2C_BATTLE_RESULT_RESPONSE), response)

	if is_win {
		for i := int32(1); i <= 3; i++ {
			tower_fight_id := get_tower_fight_id(tower_id, i)
			row := dbc.TowerFightSaves.GetRow(tower_fight_id)
			if row == nil {
				row = dbc.TowerFightSaves.AddRow(tower_fight_id)
				data = compress_battle_record_data(data)
				if data != nil {
					row.Data.SetData(data)
					row.SetAttacker(this.Id)
					row.SetSaveTime(int32(time.Now().Unix()))
				}
				break
			}
		}
	}

	return 1
}

func (this *Player) get_tower_records_info(tower_id int32) int32 {
	var records []*msg_client_message.TowerFightRecord
	for i := int32(1); i <= 3; i++ {
		tower_fight_id := get_tower_fight_id(tower_id, i)
		row := dbc.TowerFightSaves.GetRow(tower_fight_id)
		if row == nil {
			continue
		}
		attacker_id := row.GetAttacker()
		attacker := player_mgr.GetPlayerById(attacker_id)
		if attacker == nil {
			continue
		}
		records = append(records, &msg_client_message.TowerFightRecord{
			TowerFightId: tower_fight_id,
			AttackerId:   attacker_id,
			AttackerName: attacker.db.GetName(),
			CreateTime:   row.GetSaveTime(),
		})
	}
	response := &msg_client_message.S2CTowerRecordsInfoResponse{
		Records: records,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_TOWER_RECORDS_INFO_RESPONSE), response)
	return 1
}

func (this *Player) get_tower_record_data(tower_fight_id int32) int32 {
	row := dbc.TowerFightSaves.GetRow(tower_fight_id)
	if row == nil {
		log.Error("Tower fight record[%v] not found", tower_fight_id)
		return int32(msg_client_message.E_ERR_PLAYER_TOWER_FIGHT_RECORD_NOT_FOUND)
	}

	record_data := row.Data.GetData()
	record_data = decompress_battle_record_data(record_data)
	if record_data == nil {
		return int32(msg_client_message.E_ERR_PLAYER_BATTLE_RECORD_DATA_INVALID)
	}
	response := &msg_client_message.S2CTowerRecordDataResponse{
		RecordData: row.Data.GetData(),
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_TOWER_RECORD_DATA_RESPONSE), response)

	return 1
}

func C2STowerRecordsInfoHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2STowerRecordsInfoRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}

	return p.get_tower_records_info(req.GetTowerId())
}

func C2STowerRecordDataHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2STowerRecordDataRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}

	return p.get_tower_record_data(req.GetTowerFightId())
}

func C2STowerDataHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2STowerDataRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}

	return p.get_tower_data()
}
