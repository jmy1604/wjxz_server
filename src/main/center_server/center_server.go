package main

import (
	_ "encoding/json"
	"errors"
	_ "io/ioutil"
	"libs/log"
	"libs/server_conn"
	"libs/timer"
	"sync"
	"time"

	"3p/code.google.com.protobuf/proto"
)

type MessageHandler func(conn *server_conn.ServerConn, m proto.Message)

type CenterServer struct {
	quit               bool
	shutdown_lock      *sync.Mutex
	shutdown_completed bool
	ticker             *timer.TickTimer
	initialized        bool
}

var server CenterServer

func (this *CenterServer) Init() (err error) {
	if this.initialized {
		return
	}

	this.shutdown_lock = &sync.Mutex{}

	if !this.OnInit() {
		return errors.New("CenterServer OnInit Failed !")
	}
	this.initialized = true

	return
}

func (this *CenterServer) OnInit() bool {
	if !global_config_load() {
		return false
	} else {
		log.Info("Cfg_position init succeed !")
	}

	if !gm_mgr.Init() {
		log.Error("gm_mgr Init Failed")
		return false
	} else {
		log.Info("gm_mgr init succeed !")
	}

	if !player_mgr.Init() {
		log.Error("player_mgr Init failed !")
		return false
	} else {
		log.Info("player_mgr init succeed !")
	}

	if !cfg_stage_mgr.Init() {
		log.Error("cfg_stage_mgr Init failed !")
		return false
	} else {
		log.Info("cfg_stage_mgr init succeed !")
	}

	if !cfg_chapter_mgr.Init() {
		log.Error("cfg_chapter_mgr Init failed !")
		return false
	} else {
		log.Info("cfg_chapter_mgr init succeed !")
	}

	if !stage_mgr.Init() {
		log.Error("stage_mgr Init failed !")
		return false
	} else {
		log.Info("stage_mgr init succeed !")
	}

	return true
}

func (this *CenterServer) Start() {
	if !this.initialized {
		return
	}

	this.Run()
}

func (this *CenterServer) Run() {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}

		this.shutdown_completed = true
	}()

	this.ticker = timer.NewTickTimer(1000)
	this.ticker.Start()
	defer this.ticker.Stop()

	for {
		select {
		case _, ok := <-this.ticker.Chan:
			{
				if !ok {
					return
				}

				begin := time.Now()
				hall_agent_mgr.OnTick()
				time_cost := time.Now().Sub(begin).Seconds()
				if time_cost > 1 {
					log.Trace("耗时 %v", time_cost)
					if time_cost > 30 {
						log.Error("耗时 %v", time_cost)
					}
				}
			}
		}
	}
}

func (this *CenterServer) Shutdown() {
	if !this.initialized {
		return
	}

	this.shutdown_lock.Lock()
	defer this.shutdown_lock.Unlock()

	if this.quit {
		return
	}
	this.quit = true

	log.Trace("关闭游戏主循环")

	begin := time.Now()

	if this.ticker != nil {
		this.ticker.Stop()
	}

	for {
		if this.shutdown_completed {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	login_agent_mgr.server_node.Shutdown()
	hall_agent_mgr.server_node.Shutdown()

	dbc.Save(false)
	dbc_account.Save(false)

	log.Trace("关闭游戏主循环耗时 %v 秒", time.Now().Sub(begin).Seconds())
}
