package main

import (
	"libs/log"
	_ "main/table_config"
	_ "public_message/gen_go/client_message"
	_ "public_message/gen_go/server_message"
	_ "time"

	_ "3p/code.google.com.protobuf/proto"
	_ "github.com/golang/protobuf/proto"
)

func (this *DBC) on_preload() (err error) {
	var p *Player
	for _, db := range this.Players.m_rows {
		if nil == db {
			log.Error("DBC on_preload Players have nil db !")
			continue
		}

		p = new_player_with_db(db.m_PlayerId, db)
		if nil == p {
			continue
		}

		player_mgr.Add2IdMap(p)
		player_mgr.Add2AccMap(p)
	}

	return
}

func (this *dbGlobalRow) GetNextPlayerId() int32 {
	this.m_lock.UnSafeLock("dbGlobalRow.SetdbGlobalCurrentPlayerIdColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_CurrentPlayerId += 1
	new_id := ((config.ServerId << 24) & 0x7f000000) | this.m_CurrentPlayerId
	this.m_CurrentPlayerId_changed = true
	return new_id
}
