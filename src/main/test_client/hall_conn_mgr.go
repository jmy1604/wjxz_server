package main

import (
	"libs/log"
	"sync"
)

type HallConnMgr struct {
	acc2hconn      map[string]*HallConnection
	acc_arr        []*HallConnection
	acc2hconn_lock *sync.RWMutex
}

var hall_conn_mgr HallConnMgr

func (this *HallConnMgr) Init() bool {
	this.acc2hconn = make(map[string]*HallConnection)
	this.acc2hconn_lock = &sync.RWMutex{}
	return true
}

func (this *HallConnMgr) AddHallConn(conn *HallConnection) {
	if nil == conn {
		log.Error("HallConnMgr AddHallConn param error !")
		return
	}

	this.acc2hconn_lock.Lock()
	defer this.acc2hconn_lock.Unlock()

	this.acc2hconn[conn.acc] = conn
	this.acc_arr = append(this.acc_arr, conn)
}

func (this *HallConnMgr) RemoveHallConnByAcc(acc string) {
	this.acc2hconn_lock.Lock()
	defer this.acc2hconn_lock.Unlock()

	conn := this.acc2hconn[acc]
	if conn == nil {
		return
	}
	delete(this.acc2hconn, acc)
	if this.acc_arr != nil {
		n := -1
		for i := 0; i < len(this.acc_arr); i++ {
			if this.acc_arr[i] == conn {
				n = i
				break
			}
		}
		if n >= 0 {
			l := len(this.acc_arr)
			for i := n; i < l-1; i++ {
				this.acc_arr[i] = this.acc_arr[i+1]
			}
			this.acc_arr = this.acc_arr[:l-1]
		}
	}
}

func (this *HallConnMgr) GetHallConnByAcc(acc string) *HallConnection {
	this.acc2hconn_lock.RLock()
	defer this.acc2hconn_lock.RUnlock()

	return this.acc2hconn[acc]
}
