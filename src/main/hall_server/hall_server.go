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
	login_token_mgr.LoadRedisData()

	// rpc初始化
	if !this.init_rpc_service() {
		return
	}
	if !this.init_rpc_client() {
		return
	}

	world_chat_mgr.Init()
	anouncement_mgr.Init()
	battle_record_mgr.Init()
	tower_ranking_list.LoadDB()

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
	buff_pool.Init()
	passive_trigger_data_pool.Init()
	msg_battle_member_item_pool.Init()
	msg_battle_fighter_pool.Init()
	msg_battle_buff_item_pool.Init()
	msg_battle_reports_item_pool.Init()
	msg_battle_round_reports_pool.Init()
	delay_skill_pool.Init()

	player_mgr.RegMsgHandler()

	if !position_table.Init() {
		return errors.New("positioin_table init failed")
	} else {
		log.Info("position_table init succeed")
	}

	if !gm_command_mgr.Init() {
		log.Error("gm_command_mgr init failed")
		return errors.New("gm_command_mgr init failed !")
	} else {
		log.Info("gm_command_mgr init succeed !")
	}

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

	if !stage_table_mgr.Init() {
		log.Error("stage_table_mgr init failed")
		return errors.New("stage_table_mgr init failed")
	} else {
		log.Info("stage_table_mgr init succeed")
	}

	if !item_table_mgr.Init() {
		log.Error("item_table_mgr init failed")
		return errors.New("item_table_mgr init failed")
	} else {
		log.Info("item_table_mgr init succeed")
	}

	if !campaign_table_mgr.Init() {
		log.Error("campaign_table_mgr init failed")
		return errors.New("campaign_table_mgr init failed")
	} else {
		log.Info("campaign_table_gmr init succeed")
	}

	if !drop_table_mgr.Init() {
		log.Error("drop_table_mgr init failed")
		return errors.New("drop_table_mgr init failed")
	} else {
		log.Info("drop_table_mgr init succeed")
	}

	if !levelup_table_mgr.Init() {
		log.Error("levelup_table_mgr init failed")
		return errors.New("levelup_table_mgr init failed")
	} else {
		log.Info("levelup_table_mgr init succeed")
	}

	if !rankup_table_mgr.Init() {
		log.Error("rankup_table_mgr init failed")
		return errors.New("rankup_table_mgr init failed")
	} else {
		log.Info("rankup_table_mgr init succeed")
	}

	if !fusion_table_mgr.Init() {
		log.Error("fusion_table_mgr init failed")
		return errors.New("fusion_table_mgr init failed")
	} else {
		log.Info("fusion_table_mgr init succeed")
	}

	if !talent_table_mgr.Init() {
		log.Error("talent_table_mgr init failed")
		return errors.New("talent_table_mgr init failed")
	} else {
		log.Info("talent_table_mgr init success")
	}

	if !tower_table_mgr.Init() {
		log.Error("tower_table_mgr init failed")
		return errors.New("tower_table_mgr init failed")
	} else {
		log.Info("tower_table_mgr init success")
	}

	if !item_upgrade_table_mgr.Init() {
		log.Error("item_upgrade_table_mgr init failed")
		return errors.New("item_upgrade_table_mgr init failed")
	} else {
		log.Info("item_upgrade_table_mgr init success")
	}

	if !suit_table_mgr.Init() {
		log.Error("suit_table_mgr init failed")
		return errors.New("suit_table_mgr init failed")
	} else {
		log.Info("suit_table_mgr init success")
	}

	if USE_CONN_TIMER_WHEEL == 0 {
		conn_timer_mgr.Init()
	} else {
		conn_timer_wheel.Init()
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

	go this.redis_conn.Run(1000)
	if USE_CONN_TIMER_WHEEL == 0 {
		go conn_timer_mgr.Run()
	} else {
		go conn_timer_wheel.Run()
	}

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
var item_table_mgr table_config.ItemTableMgr
var drop_table_mgr table_config.DropManager
var shop_table_mgr table_config.ShopTableManager
var handbook_table_mgr table_config.HandbookTableMgr
var suit_table_mgr table_config.SuitTableMgr
var extract_table_mgr table_config.ExtractTableManager
var position_table table_config.PositionTable

var card_table_mgr table_config.CardTableMgr
var skill_table_mgr table_config.SkillTableMgr
var buff_table_mgr table_config.StatusTableMgr
var stage_table_mgr table_config.PassTableMgr
var campaign_table_mgr table_config.CampaignTableMgr
var levelup_table_mgr table_config.LevelUpTableMgr
var rankup_table_mgr table_config.RankUpTableMgr
var fusion_table_mgr table_config.FusionTableMgr
var talent_table_mgr table_config.TalentTableMgr
var tower_table_mgr table_config.TowerTableMgr
var item_upgrade_table_mgr table_config.ItemUpgradeTableMgr

var team_member_pool TeamMemberPool
var battle_report_pool BattleReportPool
var buff_pool BuffPool
var passive_trigger_data_pool MemberPassiveTriggerDataPool
var msg_battle_member_item_pool MsgBattleMemberItemPool
var msg_battle_fighter_pool MsgBattleFighterPool
var msg_battle_buff_item_pool MsgBattleMemberBuffPool
var msg_battle_reports_item_pool MsgBattleReportItemPool
var msg_battle_round_reports_pool MsgBattleRoundReportsPool
var delay_skill_pool DelaySkillPool
