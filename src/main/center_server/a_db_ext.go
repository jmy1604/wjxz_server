package main

import (
	_ "public_message/gen_go/server_message"
	_ "time"

	_ "3p/code.google.com.protobuf/proto"
	_ "github.com/golang/protobuf/proto"
)

func (this *DBC) on_preload() (err error) {

	return
}

func (this *dbPlayerIdMaxRow) Inc() (id int32) {
	this.m_lock.Lock("dbPlayerIdMaxRow.Inc")
	defer this.m_lock.Unlock()

	this.m_PlayerIdMax++
	id = this.m_PlayerIdMax

	this.m_PlayerIdMax_changed = true

	return
}
