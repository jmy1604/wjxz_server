package main

import (
	"libs/log"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	"time"

	"github.com/golang/protobuf/proto"
)

type BattleSaveManager struct {
	saves *dbBattleSaveTable
}

var battle_record_mgr BattleSaveManager

func (this *BattleSaveManager) Init() {
	this.saves = dbc.BattleSaves
}

func (this *BattleSaveManager) SaveNew(attacker_id, defenser_id int32, data []byte) bool {
	attacker := player_mgr.GetPlayerById(attacker_id)
	if attacker == nil {
		return false
	}
	defenser := player_mgr.GetPlayerById(defenser_id)
	if defenser == nil {
		return false
	}
	row := this.saves.AddRow()
	if row != nil {
		row.SetAttacker(attacker_id)
		row.SetDefenser(defenser_id)
		row.Data.SetData(data)
		row.SetSaveTime(int32(time.Now().Unix()))
		attacker.db.BattleSaves.Add(&dbPlayerBattleSaveData{
			Id: row.GetId(),
		})
		defenser.db.BattleSaves.Add(&dbPlayerBattleSaveData{
			Id: row.GetId(),
		})
		log.Debug("Battle Record[%v] saved with attacker[%v] and defenser[%v]", row.GetId(), attacker_id, defenser_id)
	}
	return true
}

func (this *BattleSaveManager) GetRecord(requester_id, record_id int32) (attacker_id, defenser_id int32, record_data []byte, save_time int32) {
	row := this.saves.GetRow(record_id)
	if row == nil {
		return
	}

	delete_state := row.GetDeleteState()

	if delete_state == 1 && attacker_id == requester_id {
		log.Error("Player[%v] is attacker, had deleted record", requester_id)
		return
	} else if delete_state == 2 && defenser_id == requester_id {
		log.Error("Player[%v] is defenser, had deleted record", requester_id)
		return
	}

	attacker_id = row.GetAttacker()
	defenser_id = row.GetDefenser()
	record_data = row.Data.GetData()
	save_time = row.GetSaveTime()
	return
}

func (this *BattleSaveManager) DeleteRecord(requester_id, record_id int32) int32 {
	row := this.saves.GetRow(record_id)
	if row == nil {
		log.Error("Not found battle record[%v]", record_id)
		return int32(msg_client_message.E_ERR_PLAYER_BATTLE_RECORD_NOT_FOUND)
	}

	attacker_id := row.GetAttacker()
	defenser_id := row.GetDefenser()
	if requester_id != attacker_id && requester_id != defenser_id {
		log.Error("Battle record[%v] cant delete by player[%v]", record_id, requester_id)
		return int32(msg_client_message.E_ERR_PLAYER_BATTLE_RECORD_FORBIDDEN_DELETE)
	}

	delete_state := row.GetDeleteState()
	if requester_id == attacker_id && delete_state == 1 {
		log.Error("Player[%v] already deleted battle record[%v] as attacker", requester_id, record_id)
		return int32(msg_client_message.E_ERR_PLAYER_BATTLE_RECORD_FORBIDDEN_DELETE)
	}

	if requester_id == defenser_id && delete_state == 2 {
		log.Error("Player[%v] already deleted battle record[%v] as defenser", requester_id, record_id)
		return int32(msg_client_message.E_ERR_PLAYER_BATTLE_RECORD_FORBIDDEN_DELETE)
	}

	attacker := player_mgr.GetPlayerById(attacker_id)
	defenser := player_mgr.GetPlayerById(defenser_id)
	if delete_state == 0 {
		if requester_id == attacker_id {
			row.SetDeleteState(1)
			if attacker != nil {
				attacker.db.BattleSaves.Remove(record_id)
			}
		} else {
			row.SetDeleteState(2)
			if defenser != nil {
				defenser.db.BattleSaves.Remove(record_id)
			}
		}
	} else if delete_state == 1 {
		this.saves.RemoveRow(record_id)
		if defenser != nil {
			defenser.db.BattleSaves.Remove(record_id)
		}
	} else if delete_state == 2 {
		this.saves.RemoveRow(record_id)
		if attacker != nil {
			attacker.db.BattleSaves.Remove(record_id)
		}
	} else {
		return int32(msg_client_message.E_ERR_PLAYER_BATTLE_RECORD_FORBIDDEN_DELETE)
	}

	return 1
}

func (this *Player) GetBattleRecordList() int32 {
	var record_list []*msg_client_message.BattleRecordData
	all := this.db.BattleSaves.GetAllIndex()
	if all != nil {
		for i := 0; i < len(all); i++ {
			row := dbc.BattleSaves.GetRow(all[i])
			if row != nil {
				record := &msg_client_message.BattleRecordData{}
				record.RecordId = row.GetId()
				record.RecordTime = row.GetSaveTime()
				record.AttackerId = row.GetAttacker()
				attacker := player_mgr.GetPlayerById(record.AttackerId)
				if attacker != nil {
					record.AttackerName = attacker.db.GetName()
				}
				record.DefenserId = row.GetDefenser()
				defenser := player_mgr.GetPlayerById(record.DefenserId)
				if defenser != nil {
					record.DefenserName = defenser.db.GetName()
				}
				record_list = append(record_list, record)
			}
		}
	}

	if record_list == nil {
		record_list = make([]*msg_client_message.BattleRecordData, 0)
	}

	response := &msg_client_message.S2CBattleRecordListResponse{
		Records: record_list,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_BATTLE_RECORD_LIST_RESPONSE), response)

	return 1
}

func (this *Player) GetBattleRecord(record_id int32) int32 {
	attacker_id, defenser_id, record_data, record_time := battle_record_mgr.GetRecord(this.Id, record_id)
	if attacker_id == 0 {
		return int32(msg_client_message.E_ERR_PLAYER_BATTLE_RECORD_NOT_FOUND)
	}

	var attacker_name, defenser_name string
	attacker := player_mgr.GetPlayerById(attacker_id)
	if attacker != nil {
		attacker_name = attacker.db.GetName()
	}
	defenser := player_mgr.GetPlayerById(defenser_id)
	if defenser != nil {
		defenser_name = defenser.db.GetName()
	}

	response := &msg_client_message.S2CBattleRecordResponse{
		Id:           record_id,
		AttackerId:   attacker_id,
		AttackerName: attacker_name,
		DefenserId:   defenser_id,
		DefenserName: defenser_name,
		RecordData:   record_data,
		RecordTime:   record_time,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_BATTLE_RECORD_RESPONSE), response)

	return 1
}

func C2SBattleRecordListHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SBattleRecordListRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}
	return p.GetBattleRecordList()
}

func C2SBattleRecordHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SBattleRecordRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}
	return p.GetBattleRecord(req.GetId())
}

func C2SBattleRecordDeleteHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SBattleRecordRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}
	return battle_record_mgr.DeleteRecord(p.Id, req.GetId())
}
