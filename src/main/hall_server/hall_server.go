package main

import (
	"errors"
	"libs/log"
	"libs/rpc"
	"libs/socket"
	"libs/timer"
	"libs/utils"
	"main/table_config"
	"sync"
	"time"

	_ "3p/code.google.com.protobuf/proto"
	_ "github.com/golang/protobuf/proto"
)

type HallServer struct {
	start_time         time.Time
	net                *socket.Node
	quit               bool
	shutdown_lock      *sync.Mutex
	shutdown_completed bool
	ticker             *timer.TickTimer
	initialized        bool
	last_gc_time       int32
	rpc_client         *rpc.Client  // 连接到rpc服务
	rpc_service        *rpc.Service // 接受rpc连接
	redis_conn         *utils.RedisConn

	server_info_row *dbServerInfoRow
}

var hall_server HallServer

func (this *HallServer) Init() (ok bool) {
	this.start_time = time.Now()
	this.shutdown_lock = &sync.Mutex{}
	this.net = socket.NewNode(&hall_server, time.Duration(config.RecvMaxMSec), time.Duration(config.SendMaxMSec), 5000, nil) //(this, 0, 0, 5000, 0, 0, 0, 0, 0)

	this.redis_conn = &utils.RedisConn{}
	if !this.redis_conn.Connect(config.RedisServerIP) {
		return
	}

	// rpc初始化
	if !this.init_rpc_service() {
		return
	}
	if !this.init_rpc_client() {
		return
	}

	world_chat_mgr.Init()
	anouncement_mgr.Init()

	err := this.OnInit()
	if err != nil {
		log.Error("服务器初始化失败[%s]", err.Error())
		return
	}

	this.initialized = true

	ok = true
	return
}

func (this *HallServer) OnInit() (err error) {
	team_member_pool.Init()
	battle_report_pool.Init()

	reg_player_base_info_msg()
	reg_player_guide_msg()
	reg_player_friend_msg()
	reg_player_draw_msg()

	player_mgr.RegMsgHandler()

	if !cfg_position.Init() {
		return errors.New("cfg_positioin init failed")
	} else {
		log.Info("cfg_position init succeed")
	}

	/*if !item_table_mgr.Init() {
		return errors.New("cfg_item_mgr init failed!")
	} else {
		log.Info("cfg_item_mgr init succeed!")
	}

	if stage_table_mgr.Init() {
		return errors.New("cfg_stage_mgr init failed !")
	} else {
		log.Info("cfg_stage_mgr init succeed !")
	}

	if !task_table_mgr.Init() {
		log.Error("task_mgr init failed")
		return errors.New("task_mgr init failed !")
	} else {
		log.Info("task_mgr init succeed !")
	}

	if !cfg_drop_card_mgr.Init() {
		log.Error("cfg_drop_card_mgr init failed !")
		return errors.New("cfg_drop_card_mgr init failed !")
	} else {
		log.Info("cfg_drop_card_mgr init succeed !")
	}

	if !extract_table_mgr.Init() {
		return errors.New("extract_table_mgr init failed")
	} else {
		log.Info("extract_table_mgr init succeed")
	}*/

	if !gm_command_mgr.Init() {
		log.Error("gm_command_mgr init failed")
		return errors.New("gm_command_mgr init failed !")
	} else {
		log.Info("gm_command_mgr init succeed !")
	}

	/*if !cfg_player_level_mgr.Init() {
		log.Error("cfg_player_level_mgr init failed")
		return errors.New("cfg_player_level_mgr init failed!")
	} else {
		log.Info("cfg_player_level_mgr init succeed!")
	}

	if !shop_table_mgr.Init() {
		log.Error("shop_mgr init failed")
		return errors.New("shop_mgr init failed")
	} else {
		log.Info("shop_mgr init succeed!")
	}

	if !box_table_mgr.Init() {
		log.Error("box_mgr init failed")
		return errors.New("box_mgr init failed")
	} else {
		log.Info("box_mgr init succeed!")
	}

	if !cfg_chapter_mgr.Init() {
		log.Error("chapter_mgr init failed")
		return errors.New("chapter_mgr init failed")
	} else {
		log.Info("chapter_mgr init succeed!")
	}

	if !level_table_mgr.Init() {
		log.Error("level_table_mgr init failed")
		return errors.New("level_table_mgr init failed")
	} else {
		log.Info("level_table_mgr init succeed")
	}

	if !handbook_table_mgr.Init() {
		log.Error("handbook_table_mgr init failed")
		return errors.New("handbook_table_mgr init failed")
	} else {
		log.Info("handbook_table_mgr init succeed")
	}

	if !suit_table_mgr.Init() {
		log.Error("suit_table_mgr init failed")
		return errors.New("suit_table_mgr init failed")
	} else {
		log.Info("suit_table_mgr init succeed")
	}

	pay_mgr.init()
	if !pay_mgr.load_google_pay_pub() {
		return errors.New("load google pay pub failed")
	} else {
		log.Info("google pay pub load succeed")
	}

	if !pay_mgr.load_google_pay_db() {
		return errors.New("load google pay db failed")
	} else {
		log.Info("google pay db load succeed")
	}

	os_player_mgr.Init()*/

	if !card_table_mgr.Init() {
		log.Error("card_table_mgr init failed")
		return errors.New("card_table_mgr init failed")
	} else {
		log.Info("card_table_mgr init succeed")
	}

	if !skill_table_mgr.Init() {
		log.Error("skill_table_mgr init failed")
		return errors.New("skill_table_mgr init failed")
	} else {
		log.Info("skill_table_mgr init succeed")
	}

	if !buff_table_mgr.Init() {
		log.Error("buff_table_mgr init failed")
		return errors.New("buff_table_mgr init failed")
	} else {
		log.Info("buff_table_mgr init succeed")
	}

	return
}

func (this *HallServer) Start(use_https bool) (err error) {
	log.Event("服务器已启动", nil, log.Property{"IP", config.ListenClientInIP})
	log.Trace("**************************************************")

	go this.Run()

	if use_https {
		msg_handler_mgr.StartHttps("../conf/server.crt", "../conf/server.key")
	} else {
		msg_handler_mgr.StartHttp()
	}

	return
}

func (this *HallServer) Run() {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}

		this.shutdown_completed = true
	}()

	this.ticker = timer.NewTickTimer(1000)
	this.ticker.Start()
	defer this.ticker.Stop()

	go this.redis_conn.Run(100)

	for {
		select {
		case d, ok := <-this.ticker.Chan:
			{
				if !ok {
					return
				}

				begin := time.Now()
				this.OnTick(d)
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

func (this *HallServer) Shutdown() {
	if !this.initialized {
		return
	}

	this.shutdown_lock.Lock()
	defer this.shutdown_lock.Unlock()

	if this.quit {
		return
	}
	this.quit = true

	this.redis_conn.Close()

	log.Trace("关闭游戏主循环")

	begin := time.Now()

	if this.ticker != nil {
		this.ticker.Stop()
		for {
			if this.shutdown_completed {
				break
			}

			time.Sleep(time.Millisecond * 100)
		}
	}

	log.Trace("等待 shutdown_completed 完毕")
	center_conn.client_node.Shutdown()
	this.net.Shutdown()
	if nil != msg_handler_mgr.msg_http_listener {
		msg_handler_mgr.msg_http_listener.Close()
	}

	this.uninit_rpc_service()
	this.uninit_rpc_client()

	log.Trace("关闭游戏主循环耗时 %v 秒", time.Now().Sub(begin).Seconds())

	dbc.Save(false)
	dbc.Shutdown()
}

func (this *HallServer) OnTick(t timer.TickTime) {
	player_mgr.OnTick()
}

func (this *HallServer) OnAccept(c *socket.TcpConn) {
	log.Info("HallServer OnAccept [%s]", c.GetAddr())
}

func (this *HallServer) OnConnect(c *socket.TcpConn) {

}

func (this *HallServer) OnDisconnect(c *socket.TcpConn, reason socket.E_DISCONNECT_REASON) {
	if c.T > 0 {
		cur_p := player_mgr.GetPlayerById(int32(c.T))
		if nil != cur_p {
			player_mgr.PlayerLogout(cur_p)
		}
	}
	log.Trace("玩家[%d] 断开连接[%v]", c.T, c.GetAddr())
}

func (this *HallServer) CloseConnection(c *socket.TcpConn, reason socket.E_DISCONNECT_REASON) {
	if c == nil {
		log.Error("参数为空")
		return
	}

	c.Close(reason)
}

func (this *HallServer) OnUpdate(c *socket.TcpConn, t timer.TickTime) {

}

var global_config_mgr table_config.GlobalConfigManager
var task_table_mgr table_config.TaskTableMgr
var item_table_mgr table_config.CfgItemManager
var cfg_drop_card_mgr table_config.CfgDropCardManager
var shop_table_mgr table_config.ShopTableManager
var handbook_table_mgr table_config.HandbookTableMgr
var suit_table_mgr table_config.SuitTableMgr
var extract_table_mgr table_config.ExtractTableManager
var cfg_position table_config.CfgPosition

var card_table_mgr table_config.CardTableMgr
var skill_table_mgr table_config.SkillTableMgr
var buff_table_mgr table_config.StatusTableMgr

var team_member_pool TeamMemberPool
var battle_report_pool BattleReportPool
