package main

import (
	"errors"
	"libs/log"
	"libs/rpc"
	"libs/socket"
	"libs/timer"
	"main/table_config"
	"public_message/gen_go/client_message"
	"sync"
	"time"

	"3p/code.google.com.protobuf/proto"
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

	server_info_row *dbServerInfoRow
}

var hall_server HallServer

func (this *HallServer) Init() (ok bool) {
	this.start_time = time.Now()
	this.shutdown_lock = &sync.Mutex{}
	this.net = socket.NewNode(&hall_server, time.Duration(config.RecvMaxMSec), time.Duration(config.SendMaxMSec), 5000, msg_client_message.MessageNames) //(this, 0, 0, 5000, 0, 0, 0, 0, 0)

	// rpc初始化
	if !this.init_rpc_service() {
		return
	}
	if !this.init_rpc_client() {
		return
	}

	if !global_id.Load("../game_data/global_id.json") {
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
	reg_player_mail_msg()
	reg_player_base_info_msg()
	reg_player_sign_msg()
	reg_player_first_pay_msg()
	reg_player_guide_msg()
	reg_player_friend_msg()
	reg_player_stage_msg()
	reg_player_draw_msg()
	reg_player_chapter_msg()
	reg_player_activity_msg()

	player_mgr.RegMsgHandler()

	if !item_table_mgr.Init() {
		return errors.New("cfg_item_mgr init failed!")
	} else {
		log.Info("cfg_item_mgr init succeed!")
	}

	if stage_table_mgr.Init() {
		return errors.New("cfg_stage_mgr init failed !")
	} else {
		log.Info("cfg_stage_mgr init succeed !")
	}

	if !cfg_position.Init() {
		return errors.New("cfg_position init failed !")
	} else {
		log.Info("cfg_position init succeed !")
	}

	//if !achieve_task_mgr.Init() {
	if !task_table_mgr.Init() {
		log.Error("task_mgr init failed")
		return errors.New("task_mgr init failed !")
	} else {
		log.Info("task_mgr init succeed !")
	}

	if !cfg_build_area_mgr.Init() {
		return errors.New("cfg_build_area_mgr init failed !")
	} else {
		log.Info("cfg_build_area_mgr init succeed !")
	}

	if !cat_table_mgr.Init() {
		log.Error("cfg_character_mgr init failed !")
		return errors.New("cfg_character_mgr init failed !")
	} else {
		log.Info("cfg_character_mgr init succeed !")
	}

	if !cfg_skill_mgr.Init() {
		log.Error("cfg_skill_mgr init failed !")
		return errors.New("cfg_skill_mgr init failed")
	} else {
		log.Info("cfg_skill_mgr init succeed")
	}

	if !cfg_building_mgr.Init() {
		log.Error("cfg_building_mgr init failed !")
		return errors.New("cfg_building_mgr init failed")
	} else {
		log.Info("cfg_building_mgr init succeed")
	}

	if !cfg_drop_card_mgr.Init() {
		log.Error("cfg_drop_card_mgr init failed !")
		return errors.New("cfg_drop_card_mgr init failed !")
	} else {
		log.Info("cfg_drop_card_mgr init succeed !")
	}

	if !cfg_expedition_mgr.Init() {
		log.Error("cfg_expedition_mgr init failed !")
		return errors.New("cfg_expedition_mgr init failed !")
	} else {
		log.Info("cfg_expedition_mgr init succeed")
	}

	if !cfg_block_mgr.Init() {
		log.Error("cfg_block_mgr init failed !")
		return errors.New("cfg_block_mgr init failed !")
	} else {
		log.Info("cfg_block_mgr init succeed")
	}

	if !cfg_mapchest_mgr.Init() {
		log.Error("cfg_mapchest_mgr init failed !")
		return errors.New("cfg_mapchest_mgr init failed !")
	} else {
		log.Info("cfg_mapchest_mgr init succeed")
	}

	if !cfg_areaunlock_mgr.Init() {
		log.Error("cfg_areaunlock_mgr init failed !")
		return errors.New("cfg_areaunlock_mgr init failed !")
	} else {
		log.Info("cfg_areaunlock_mgr init succeed")
	}

	if !cfg_activity_mgr.Init() {
		log.Error("cfg_activity_mgr init failed !")
		return errors.New("cfg_activity_mgr init failed !")
	} else {
		log.Info("cfg_activity_mgr init succeed")
	}

	if !gm_command_mgr.Init() {
		log.Error("gm_command_mgr init failed")
		return errors.New("gm_command_mgr init failed !")
	} else {
		log.Info("gm_command_mgr init succeed !")
	}

	if !stage_pass_mgr.Init() {
		log.Error("stage_pass_mgr init failed")
		return errors.New("stage_pass_mgr init failed !")
	} else {
		log.Info("stage_pass_mgr init succeed !")
	}

	if !cfg_player_level_mgr.Init() {
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

	/*if !cfg_mail_mgr.Init() {
		return errors.New("cfg_mail_mgr init failed !")
	} else {
		log.Info("cfg_mail_mgr init succeed !")
	}*/

	//chest_cfg_mgr.OpenChestTest()

	if !formula_table_mgr.Init() {
		log.Error("formula_mgr init failed")
		return errors.New("formula_mgr init failed")
	} else {
		log.Info("formula_mgr init succeed")
	}

	if !other_table_mgr.Init() {
		log.Error("other_mgr init failed")
		return errors.New("other_mgr init failed")
	} else {
		log.Info("other_mgr init succeed")
	}

	if !crop_table_mgr.Init() {
		log.Error("crop_mgr init failed")
		return errors.New("crop_mgr init failed")
	} else {
		log.Info("crop_mgr init succeed")
	}

	if !cathouse_table_mgr.Init() {
		log.Error("cathouse_mgr init failed")
		return errors.New("cathouse_mgr init failed")
	} else {
		log.Info("cathouse_mgr init succeed")
	}

	if !extract_table_mgr.Init() {
		log.Error("extract_table_mgr init failed")
		return errors.New("extract_table_mgr init failed")
	} else {
		log.Info("extract_table_mgr init succeed")
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

	if !foster_table_mgr.Init() {
		log.Error("foster_table_mgr init failed")
		return errors.New("foster_table_mgr init failed")
	} else {
		log.Info("foster_table_mgr init succeed")
	}

	if !foster_card_table_mgr.Init() {
		log.Error("foster_card_table_mgr init failed")
		return errors.New("foster_card_table_mgr init failed")
	} else {
		log.Info("foster_card_table_mgr init succeed")
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

	os_player_mgr.Init()

	return
}

func (this *HallServer) Start() (err error) {
	log.Event("服务器已启动", nil, log.Property{"IP", config.ListenClientInIP})
	log.Trace("**************************************************")

	go this.Run()

	msg_handler_mgr.StartHttp()

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

type MessageHandler func(conn *socket.TcpConn, m proto.Message)

func (this *HallServer) set_ih(type_id uint16, h socket.Handler) {
	t := msg_client_message.MessageTypes[type_id]
	if t == nil {
		log.Error("设置消息句柄失败，不存在的消息类型 %v", type_id)
		return
	}

	this.net.SetHandler(type_id, t, h)
}

func (this *HallServer) SetMessageHandler(type_id uint16, h MessageHandler) {
	if h == nil {
		this.set_ih(type_id, nil)
		return
	}

	this.set_ih(type_id, func(c *socket.TcpConn, m proto.Message) {
		h(c, m)
	})
}

func (this *HallServer) OnUpdate(c *socket.TcpConn, t timer.TickTime) {

}

var global_id table_config.GlobalId

var global_config_mgr table_config.GlobalConfigManager

var task_table_mgr table_config.TaskTableMgr

var item_table_mgr table_config.CfgItemManager

var cfg_building_mgr table_config.CfgBuildingMgr

var stage_table_mgr table_config.CfgStageManager

var cat_table_mgr table_config.CfgCharacterMgr

var cfg_skill_mgr table_config.CfgSkillMgr

var cfg_drop_card_mgr table_config.CfgDropCardManager

var cfg_player_level_mgr table_config.CfgPlayerLevelManager

var shop_table_mgr table_config.ShopTableManager

var box_table_mgr table_config.BoxTableManager

var cfg_position table_config.CfgPosition

var cfg_chapter_mgr table_config.CfgChapterManager

var cfg_build_area_mgr table_config.CfgBuildAreaMgr

var chest_cfg_mgr table_config.ChestConfigMgr

var cfg_mail_mgr table_config.MailConfigManager

var cfg_day_sign_mgr table_config.CfgDaySignManager

var formula_table_mgr table_config.FormulaTableMgr

var other_table_mgr table_config.OtherTableManager

var crop_table_mgr table_config.CropTableMgr

var cathouse_table_mgr table_config.CatHouseTableMgr

var cfg_expedition_mgr table_config.CfgExpeditionMgr

var extract_table_mgr table_config.ExtractTableManager

var cfg_block_mgr table_config.CfgBlockMgr

var cfg_mapchest_mgr table_config.CfgMapChestMgr

var cfg_areaunlock_mgr table_config.CfgAreaUnlockMgr

var level_table_mgr table_config.LevelTableMgr

var handbook_table_mgr table_config.HandbookTableMgr

var suit_table_mgr table_config.SuitTableMgr

var cfg_activity_mgr table_config.CfgActivityMgr

var foster_table_mgr table_config.FosterTableMgr

var foster_card_table_mgr table_config.FosterCardTableMgr
