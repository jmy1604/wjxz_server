package main

import (
	"public_message/gen_go/client_message"
	"sync"
)

// 战报
type BattleReportPool struct {
	pool *sync.Pool
}

func (this *BattleReportPool) Init() {
	this.pool = &sync.Pool{
		New: func() interface{} {
			m := &msg_client_message.BattleReportItem{}
			return m
		},
	}
}

func (this *BattleReportPool) Get() *msg_client_message.BattleReportItem {
	return this.pool.Get().(*msg_client_message.BattleReportItem)
}

func (this *BattleReportPool) Put(m *msg_client_message.BattleReportItem) {
	this.pool.Put(m)
}

// 阵型成员
type TeamMemberPool struct {
	pool   *sync.Pool
	locker *sync.RWMutex
}

func (this *TeamMemberPool) Init() {
	this.pool = &sync.Pool{
		New: func() interface{} {
			m := &TeamMember{}
			return m
		},
	}
	this.locker = &sync.RWMutex{}
}

func (this *TeamMemberPool) Get() *TeamMember {
	return this.pool.Get().(*TeamMember)
}

func (this *TeamMemberPool) Put(m *TeamMember) {
	this.pool.Put(m)
}

// BUFF
type BuffPool struct {
	pool *sync.Pool
}

func (this *BuffPool) Init() {
	this.pool = &sync.Pool{
		New: func() interface{} {
			return &Buff{}
		},
	}
}

func (this *BuffPool) Get() *Buff {
	return this.pool.Get().(*Buff)
}

func (this *BuffPool) Put(b *Buff) {
	this.pool.Put(b)
}
