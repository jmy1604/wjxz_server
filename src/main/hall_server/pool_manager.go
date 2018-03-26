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

// MemberPassiveTriggerData
type MemberPassiveTriggerDataPool struct {
	pool *sync.Pool
}

func (this *MemberPassiveTriggerDataPool) Init() {
	this.pool = &sync.Pool{
		New: func() interface{} {
			return &MemberPassiveTriggerData{}
		},
	}
}

func (this *MemberPassiveTriggerDataPool) Get() *MemberPassiveTriggerData {
	return this.pool.Get().(*MemberPassiveTriggerData)
}

func (this *MemberPassiveTriggerDataPool) Put(d *MemberPassiveTriggerData) {
	this.pool.Put(d)
}

// MsgBattleMemberItemPool
type MsgBattleMemberItemPool struct {
	pool *sync.Pool
}

func (this *MsgBattleMemberItemPool) Init() {
	this.pool = &sync.Pool{
		New: func() interface{} {
			return &msg_client_message.BattleMemberItem{}
		},
	}
}

func (this *MsgBattleMemberItemPool) Get() *msg_client_message.BattleMemberItem {
	return this.pool.Get().(*msg_client_message.BattleMemberItem)
}

func (this *MsgBattleMemberItemPool) Put(item *msg_client_message.BattleMemberItem) {
	this.pool.Put(item)
}

// MsgBattleReportItemPool
type MsgBattleReportItemPool struct {
	pool *sync.Pool
}

func (this *MsgBattleReportItemPool) Init() {
	this.pool = &sync.Pool{
		New: func() interface{} {
			return &msg_client_message.BattleReportItem{}
		},
	}
}

func (this *MsgBattleReportItemPool) Get() *msg_client_message.BattleReportItem {
	return this.pool.Get().(*msg_client_message.BattleReportItem)
}

func (this *MsgBattleReportItemPool) Put(item *msg_client_message.BattleReportItem) {
	this.pool.Put(item)
}

// MsgBattleRoundReportsPool
type MsgBattleRoundReportsPool struct {
	pool *sync.Pool
}

func (this *MsgBattleRoundReportsPool) Init() {
	this.pool = &sync.Pool{
		New: func() interface{} {
			return &msg_client_message.BattleRoundReports{}
		},
	}
}

func (this *MsgBattleRoundReportsPool) Get() *msg_client_message.BattleRoundReports {
	return this.pool.Get().(*msg_client_message.BattleRoundReports)
}

func (this *MsgBattleRoundReportsPool) Put(reports *msg_client_message.BattleRoundReports) {
	this.pool.Put(reports)
}
