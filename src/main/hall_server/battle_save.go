package main

import (
	"libs/log"
	"time"
)

type BattleSaveManager struct {
	saves *dbBattleSaveTable
}

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

func (this *BattleSaveManager) DeleteRecord(requester_id, record_id int32) bool {
	row := this.saves.GetRow(record_id)
	if row == nil {
		log.Error("Not found battle record[%v]", record_id)
		return false
	}

	attacker_id := row.GetAttacker()
	defenser_id := row.GetDefenser()
	if requester_id != attacker_id && requester_id != defenser_id {
		log.Error("Battle record[%v] cant delete by player[%v]", record_id, requester_id)
		return false
	}

	delete_state := row.GetDeleteState()
	if requester_id == attacker_id && delete_state == 1 {
		log.Error("Player[%v] already deleted battle record[%v] as attacker", requester_id, record_id)
		return false
	}

	if requester_id == defenser_id && delete_state == 2 {
		log.Error("Player[%v] already deleted battle record[%v] as defenser", requester_id, record_id)
		return false
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
		return false
	}

	return true
}
