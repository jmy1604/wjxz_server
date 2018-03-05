package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"libs/log"
	"libs/timer"
	"net/http"
	"public_message/gen_go/client_message"
	"strings"
	"sync"
	"time"

	"3p/code.google.com.protobuf/proto"
)

const (
	TONG_MEMBER_MENBER  = 0 // 帮会成员
	TONG_MEMBER_CAPTAIN = 1 // 帮会帮主

	TONG_JOIN_TYPE_EVERYONE = 0 // 任何人都可以加入
	TONG_JOIN_TYPE_NOONE    = 1 // 不允许任何人加入
	TONG_JOIN_TYPE_NEED_CHK = 2 // 需要确认

	TONG_AGREE_AVI_SEC = 5 // 同意占位有效时间
)

type TestClient struct {
	start_time         time.Time
	quit               bool
	shutdown_lock      *sync.Mutex
	shutdown_completed bool
	ticker             *timer.TickTimer
	initialized        bool
}

var test_client TestClient

func (this *TestClient) Init() (ok bool) {
	this.start_time = time.Now()
	this.shutdown_lock = &sync.Mutex{}
	this.initialized = true

	return true
}

func (this *TestClient) Start() (err error) {

	log.Event("客户端已启动", nil)
	log.Trace("**************************************************")

	this.Run()

	return
}

func (this *TestClient) Run() {
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

				this.OnTick(d)
			}
		}
	}
}

func (this *TestClient) Shutdown() {
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

	log.Trace("关闭游戏主循环耗时 %v 秒", time.Now().Sub(begin).Seconds())
}

const (
	CMD_TYPE_LOGIN = 1 // 登录命令
)

type JsonLoginRes struct {
	Code    int32
	Account string
	Token   string
	HallIP  string
}

var cur_hall_conn *HallConnection

func (this *TestClient) cmd_login() {
	var acc string
	fmt.Printf("请输入账号：")
	fmt.Scanf("%s\n", &acc)
	cur_hall_conn = hall_conn_mgr.GetHallConnByAcc(acc)
	if nil != cur_hall_conn && cur_hall_conn.blogin {
		log.Info("[%s] already login", acc)
		return
	}

	if config.AccountNum == 0 {
		config.AccountNum = 1
	}
	for i := int32(0); i < config.AccountNum; i++ {
		account := acc
		if config.AccountNum > 1 {
			account = fmt.Sprintf("%s_%v", acc, i)
		}
		url_str := fmt.Sprintf(config.LoginUrl, account)
		//url_str := fmt.Sprintf("http://123.207.182.67:15000/login?account=%s&token=0000", acc)
		//url_str := fmt.Sprintf("http://192.168.10.113:35000/login?account=%s&token=0000", acc)
		fmt.Println("Url str %s", url_str)
		http.Get(url_str)
		http.Get(url_str)
		http.Get(url_str)
		resp, err := http.Get(url_str)
		if nil != err {
			log.Error("login http get err (%s)", err.Error())
			return
		}

		data, err := ioutil.ReadAll(resp.Body)
		if nil != err {
			log.Error("login ioutil readall failed err(%s) !", err.Error())
			return
		}

		res := &JsonLoginRes{}
		err = json.Unmarshal(data, res)
		if nil != err {
			log.Error("login ummarshal failed err(%s)", err.Error())
			return
		}

		log.Info("login before connect hall, HallIP(%v), Account(%v), Token(%v)", res.HallIP, res.Account, res.Token)

		cur_hall_conn = new_hall_connect(res.HallIP, res.Account, res.Token)
		hall_conn_mgr.AddHallConn(cur_hall_conn)

		req2s := &msg_client_message.C2SLoginRequest{}
		req2s.Acc = proto.String(res.Account)
		req2s.Token = proto.String(res.Token)
		req2s.Channel = proto.String("test_channel")

		cur_hall_conn.Send(req2s)

		if config.AccountNum > 1 {
			log.Debug("Account[%v] logined, total count[%v]", res.Account, i+1)
		}
	}
}

func (this *TestClient) cmd_options_get() {
	if nil == cur_hall_conn || !cur_hall_conn.blogin {
		log.Error("当前连接未登陆", cur_hall_conn)
	} else {
		req := &msg_client_message.C2SGetOptions{}
		log.Info("获取选项保存值 [%v]", req)
		cur_hall_conn.Send(req)
	}
}

func (this *TestClient) cmd_stage_pass() {
	if nil == cur_hall_conn || !cur_hall_conn.blogin {
		log.Error("当前连接未登陆", cur_hall_conn)
	} else {
		req := &msg_client_message.C2SStagePass{}
		fmt.Println("请输入关卡Id:")
		var stageid int32
		fmt.Scanf("%d\n", &stageid)
		req.StageId = proto.Int32(stageid)
		fmt.Println("请输入积分:")
		var score int32
		fmt.Scanf("%d\n", &score)
		req.Score = proto.Int32(score)
		fmt.Println("请输入星星:")
		var stars int32
		fmt.Scanf("%d\n", &stars)
		req.Stars = proto.Int32(stars)
		cur_hall_conn.Send(req)
	}
}

func (this *TestClient) cmd_get_items() {
	if nil == cur_hall_conn || !cur_hall_conn.blogin {
		log.Error("当前连接未登陆", cur_hall_conn)
	} else {
		req := &msg_client_message.C2SGetItemInfos{}
		cur_hall_conn.Send(req)
	}
}

func (this TestClient) cmd_draw_cards() {
	if nil == cur_hall_conn || !cur_hall_conn.blogin {
		log.Error("当前连接未登陆", cur_hall_conn)
	} else {
		req := &msg_client_message.C2SDraw{}
		fmt.Println("请输入抽取类型:")
		var drawtype int32
		fmt.Scanf("%d\n", &drawtype)
		req.DrawType = proto.Int32(drawtype)
		fmt.Println("请输入抽取次数:")
		var drawcount int32
		fmt.Scanf("%d\n", &drawcount)
		req.DrawCount = proto.Int32(drawcount)
		cur_hall_conn.Send(req)
	}
}

func (this *TestClient) cmd_set_building() {
	if nil == cur_hall_conn || !cur_hall_conn.blogin {
		log.Error("当前连接未登陆", cur_hall_conn)
	} else {
		req := &msg_client_message.C2SSetBuilding{}
		fmt.Println("请输入建筑配置Id:")
		var building_cfgid int32
		fmt.Scanf("%d\n", &building_cfgid)
		req.BuildingCfgId = proto.Int32(building_cfgid)
		fmt.Println("请输入方向:")
		var dir int32
		fmt.Scanf("%d\n", &dir)
		fmt.Println("请输入X坐标:")
		var X int32
		fmt.Scanf("%d\n", &X)
		req.X = proto.Int32(X)
		fmt.Println("请输入Y坐标:")
		var Y int32
		fmt.Scanf("%d\n", &Y)
		req.Y = proto.Int32(Y)
		cur_hall_conn.Send(req)
	}
}

func (this *TestClient) cmd_get_cats() {
	req := &msg_client_message.C2SGetCatInfos{}
	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_get_base_info() {
	req := &msg_client_message.C2SGetBaseInfo{}
	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_change_nick() {
	var new_nick string
	fmt.Println("请输入新昵称:")
	fmt.Scanf("%s\n", &new_nick)
	req := &msg_client_message.C2SChgName{}
	req.Name = proto.String(new_nick)
	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_get_all_expeditions() {
	req := &msg_client_message.C2SGetAllExpedition{}
	fmt.Println("请求所有探险任务！")
	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_get_expedition_reward() {
	var task_id int32
	fmt.Println("请输入任务Id:")
	fmt.Scanf("%s\n", &task_id)
	req := &msg_client_message.C2SGetExpeditionReward{}
	req.Id = proto.Int32(task_id)
	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_start_expedition() {
	var task_id int32
	fmt.Println("请输入任务Id:")
	fmt.Scanf("%d\n", &task_id)
	req := &msg_client_message.C2SStartExpedition{}
	req.Id = proto.Int32(task_id)

	var cat_num, cat_id int32
	fmt.Println("请输入猫数量:")
	fmt.Scanf("%d\n", &cat_num)
	req.CatIds = make([]int32, 0, cat_num)
	for idx := int32(0); idx < cat_num; idx++ {

		fmt.Println("请输入猫Id:")
		fmt.Scanf("%d\n", &cat_id)
		req.CatIds = append(req.CatIds, cat_id)
	}

	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_reward_expedition() {
	var task_id int32
	fmt.Println("请输入任务Id:")
	fmt.Scanf("%d\n", &task_id)
	req := &msg_client_message.C2SGetExpeditionReward{}
	req.Id = proto.Int32(task_id)

	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_chg_expedition() {
	var task_id int32
	fmt.Println("请输入任务Id:")
	fmt.Scanf("%d\n", &task_id)
	req := &msg_client_message.C2SChgExpedition{}
	req.Id = proto.Int32(task_id)

	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_unlock_area() {
	var area_id int32
	fmt.Println("请输入区域Id:")
	fmt.Scanf("%d\n", &area_id)
	var if_quick int32
	fmt.Println("是否快速解锁:")
	fmt.Scanf("%d\n", &if_quick)
	req := &msg_client_message.C2SUnlockArea{}
	req.AreaId = proto.Int32(area_id)
	req.IfQuick = proto.Int32(if_quick)

	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_get_all_buildings() {
	cur_hall_conn.Send(&msg_client_message.C2SGetBuildingInfos{})
}

func (this *TestClient) cmd_heart_beat() {
	log.Info("心跳")
	cur_hall_conn.Send(&msg_client_message.HeartBeat{})
}

func (this *TestClient) cmd_remove_block() {
	var building_id int32
	fmt.Println("请输入建筑Id:")
	fmt.Scanf("%d\n", &building_id)
	req := &msg_client_message.C2SRemoveBlock{}
	req.BuildingId = proto.Int32(building_id)

	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_open_chest() {
	var building_id int32
	fmt.Println("请输入建筑Id:")
	fmt.Scanf("%d\n", &building_id)
	req := &msg_client_message.C2SOpenMapChest{}
	req.BuildingId = proto.Int32(building_id)

	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_get_info() {
	fmt.Println("获取基本信息")
	req := &msg_client_message.C2SGetInfo{}
	req.Stage = proto.Bool(true)

	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_unlock_chapter() {
	fmt.Println("请输入章节Id:")
	var chapter_id int32
	fmt.Scanf("%d\n", &chapter_id)

	fmt.Println("请输入解锁类型(0时间/1星星/2钻石/3请求好友):")
	var unlock_type int32
	fmt.Scanf("%d\n", &unlock_type)

	req := &msg_client_message.C2SChapterUnlock{}
	req.ChapterId = proto.Int32(chapter_id)
	req.UnLockType = proto.Int32(unlock_type)

	if 3 == unlock_type {
		fmt.Println("请输入请求好友的数目:")
		var friend_num int32
		fmt.Scanf("%d\n", &friend_num)
		if friend_num <= 0 {
			fmt.Println("你输入的好友数目[%d]由错误", friend_num)
			return
		}
		req.FriendIds = make([]int32, 0, friend_num)

		var friend_id int32
		for idx := int32(0); idx < friend_num; idx++ {
			fmt.Println("请输入第%d个好友的Id:", idx)
			fmt.Scanf("%d\n", &friend_id)

			req.FriendIds = append(req.FriendIds, friend_id)
		}
	}

	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_get_all_act() {
	req := &msg_client_message.C2SGetAllActivityInfos{}
	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_get_act_reward() {
	fmt.Println("获取活动奖励")
	req := &msg_client_message.C2SGetActivityReward{}
	fmt.Println("请输入活动Id:")
	var act_id int32
	fmt.Scanf("%d\n", &act_id)
	req.ActivityCfgId = proto.Int32(act_id)
	fmt.Println("请输入附加参数:")
	var extra_param int32
	fmt.Scanf("%d\n", &extra_param)
	req.ExtraParams = make([]int32, 1)
	req.ExtraParams[0] = extra_param

	cur_hall_conn.Send(req)
}

func (this *TestClient) cmd_get_area_info() {
	fmt.Println("获取活动奖励")
	req := &msg_client_message.C2SGetAreasInfos{}

	cur_hall_conn.Send(req)
}

var is_test bool

func (this *TestClient) OnTick(t timer.TickTime) {
	if !is_test {
		fmt.Printf("请输入命令:\n")
		var cmd_str string
		fmt.Scanf("%s\n", &cmd_str)
		switch cmd_str {
		case "login":
			{
				this.cmd_login()
				is_test = true
			}
		case "options_get":
			{
				this.cmd_options_get()
			}
		case "stage_pass":
			{
				this.cmd_stage_pass()
			}
		case "get_items":
			{
				this.cmd_get_items()
			}
		case "draw_card":
			{
				this.cmd_draw_cards()
			}
		case "set_building":
			{
				this.cmd_set_building()
			}
		case "get_all_buildings":
			{
				this.cmd_get_all_buildings()
			}
		case "get_cats":
			{
				this.cmd_get_cats()
			}
		case "get_base_info":
			{
				this.cmd_get_base_info()
			}
		case "rename_nick":
			{
				this.cmd_change_nick()
			}
		case "enter_test":
			{
				is_test = true
			}
		case "get_all_expedition":
			{
				this.cmd_get_all_expeditions()
			}
		case "get_expedition_reward":
			{
				this.cmd_get_expedition_reward()
			}
		case "start_expedition":
			{
				this.cmd_start_expedition()
			}
		case "reward_expedition":
			{
				this.cmd_reward_expedition()
			}
		case "chg_expedition":
			{
				this.cmd_chg_expedition()
			}
		case "unlock_area":
			{
				this.cmd_unlock_area()
			}
		case "heart_beat":
			{
				this.cmd_heart_beat()
			}
		case "remove_block":
			{
				this.cmd_remove_block()
			}
		case "get_info":
			{
				this.cmd_get_info()
			}
		case "unlock_chapter":
			{
				this.cmd_unlock_chapter()
			}
		case "get_all_acts":
			{
				this.cmd_get_all_act()
			}
		case "get_act_reward":
			{
				this.cmd_get_act_reward()
			}
		case "get_area_info":
			{
				this.cmd_get_area_info()
			}
		}
	} else {
		fmt.Printf("请输入测试命令:\n")
		var cmd_str string
		fmt.Scanln(&cmd_str, "\n")
		switch cmd_str {
		case "leave_test":
			{
				is_test = false
			}
		default:
			{
				if cmd_str != "" {
					strs := strings.Split(cmd_str, ",")
					fmt.Printf("strs[%v] length is %v\n", strs, len(strs))
					if len(strs) == 1 {
						//fmt.Printf("命令[%v]参数不够，至少一个\n", strs[0])
						//return
					} else if len(strs) == 0 {
						fmt.Printf("没有输入命令\n")
						return
					}
					req := &msg_client_message.C2S_TEST_COMMAND{}
					req.Cmd = proto.String(strs[0])
					req.Args = strs[1:]
					if config.AccountNum > 1 {
						n := (config.AccountNum + 100 - 1) / 100
						for i := int32(0); i < n; i++ {
							go func() {
								for j := i * 100; j < i*(100+1); j++ {
									if int(j) >= len(hall_conn_mgr.acc_arr) {
										break
									}
									conn := hall_conn_mgr.acc_arr[j]
									conn.Send(req)
								}
							}()
						}
					} else {
						cur_hall_conn.Send(req)
					}
				}
			}
		}
	}
}

//=================================================================================
