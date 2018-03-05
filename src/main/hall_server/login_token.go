package main

import (
	"libs/log"
	"libs/socket"
	"sync"
	"time"
)

type LoginTokenInfo struct {
	acc         string
	token       string
	playerid    int32
	create_time int64
	conn        *socket.TcpConn
}

type LoginTokenMgr struct {
	acc2token      map[string]*LoginTokenInfo
	acc2token_lock *sync.RWMutex

	id2acc      map[int32]string
	id2acc_lock *sync.RWMutex
}

var login_token_mgr LoginTokenMgr

func (this *LoginTokenMgr) Init() bool {
	this.acc2token = make(map[string]*LoginTokenInfo)
	this.acc2token_lock = &sync.RWMutex{}

	this.id2acc = make(map[int32]string)
	this.id2acc_lock = &sync.RWMutex{}
	return true
}

func (this *LoginTokenMgr) AddToAcc2Token(acc, token string, playerid int32) {
	if "" == acc {
		log.Error("LoginTokenMgr AddToAcc2Token acc empty")
		return
	}

	this.acc2token_lock.Lock()
	defer this.acc2token_lock.Unlock()

	this.acc2token[acc] = &LoginTokenInfo{acc: acc, token: token, create_time: time.Now().Unix(), playerid: playerid}
	return
}

func (this *LoginTokenMgr) RemoveFromAcc2Token(acc string) {
	if "" == acc {
		log.Error("LoginTokenMgr RemoveFromAcc2Token acc empty !")
		return
	}

	this.acc2token_lock.Lock()
	defer this.acc2token_lock.Unlock()

	if nil != this.acc2token[acc] {
		delete(this.acc2token, acc)
	}

	return
}

func (this *LoginTokenMgr) GetTockenByAcc(acc string) *LoginTokenInfo {
	if "" == acc {
		log.Error("LoginTokenMgr GetTockenByAcc acc empty")
		return nil
	}

	this.acc2token_lock.Lock()
	defer this.acc2token_lock.Unlock()

	return this.acc2token[acc]
}

func (this *LoginTokenMgr) AddToId2Acc(playerid int32, acc string) {
	if "" == acc {
		log.Error("LoginTokenMgr AddToId2Acc acc empty !")
		return
	}

	this.id2acc_lock.Lock()
	defer this.id2acc_lock.Unlock()

	this.id2acc[playerid] = acc
	return
}

func (this *LoginTokenMgr) RemoveFromId2Acc(playerid int32) {
	this.id2acc_lock.Lock()
	defer this.id2acc_lock.Unlock()
	if "" != this.id2acc[playerid] {
		delete(this.id2acc, playerid)
	}

	return
}

func (this *LoginTokenMgr) GetAccById(playerid int32) string {
	this.id2acc_lock.RLock()
	defer this.id2acc_lock.RUnlock()

	return this.id2acc[playerid]
}
