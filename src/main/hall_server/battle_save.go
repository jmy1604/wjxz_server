package main

type BattleSaveManager struct {
	saves *dbBattleSaveTable
}

func (this *BattleSaveManager) Init() {
	this.saves = dbc.BattleSaves
}

func (this *BattleSaveManager) SaveNew(attacker_id, defenser_id int32, data []byte) {

}
