package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"libs/log"
	"os"
	"public_message/gen_go/server_message"
	"time"
)

type ServerConfig struct {
	ServerId             int32
	InnerVersion         string
	ServerName           string
	ListenRoomServerIP   string
	ListenClientInIP     string
	ListenClientOutIP    string
	MaxClientConnections int32
	MaxRoomConnections   int32
	RpcServerIP          string
	ListenRpcServerIP    string

	LogConfigDir   string // 日志配置文件路径
	CenterServerIP string // 中心服务器IP
	MatchServerIP  string // 匹配服务器IP
	RecvMaxMSec    int64  // 接收超时毫秒数
	SendMaxMSec    int64  // 发送超时毫秒数

	MYSQL_NAME    string
	MYSQL_IP      string
	MYSQL_ACCOUNT string
	MYSQL_PWD     string
	DBCST_MIN     int
	DBCST_MAX     int

	MYSQL_COPY_PATH string
}

var config ServerConfig
var shutingdown bool
var dbc DBC

func after_center_match_conn() {

	gm_command_mgr.AfterCenterMatchConn()

	if signal_mgr.IfClosing() {
		return
	}
}

func main() {
	defer func() {
		log.Event("关闭服务器", nil)
		if err := recover(); err != nil {
			log.Stack(err)
		}
		time.Sleep(time.Second * 5)
		hall_server.Shutdown()
	}()

	var temp_i int32

	config_file := "../conf/hall_server.cfg"
	if len(os.Args) > 1 {
		arg_config_file := flag.String("f", "", "config file path")
		if arg_config_file != nil && *arg_config_file == "" {
			flag.Parse()
			fmt.Printf("配置参数 %s", *arg_config_file)
			config_file = *arg_config_file
		}
	}

	data, err := ioutil.ReadFile(config_file)
	if err != nil {
		fmt.Printf("读取配置文件失败 %v", err)
		return
	}
	err = json.Unmarshal(data, &config)
	if err != nil {
		fmt.Printf("解析配置文件失败 %v", err.Error())
		fmt.Scanln(&temp_i)
		return
	}

	// 加载日志配置
	log.Init("", config.LogConfigDir, true)
	log.Event("配置:网络协议版本", int32(msg_server_message.E_VERSION_NUMBER))
	log.Event("配置:服务器监听客户端地址", config.ListenClientInIP)
	log.Event("配置:最大客户端连接数)", config.MaxClientConnections)
	log.Event("连接数据库", config.MYSQL_NAME, log.Property{"地址", config.MYSQL_IP})
	err = dbc.Conn(config.MYSQL_NAME, config.MYSQL_IP, config.MYSQL_ACCOUNT, config.MYSQL_PWD, config.MYSQL_COPY_PATH)
	if err != nil {
		log.Error("连接数据库失败 %v", err)
		return
	} else {
		log.Event("连接数据库成功", nil)
		go dbc.Loop()
	}

	if !signal_mgr.Init() {
		log.Error("signal_mgr init failed")
		return
	}

	// 配置加载
	if !global_config_mgr.Init("../game_data/global.json") {
		log.Error("global_config_load failed !")
		return
	} else {
		log.Info("global_config_load succeed !")
	}

	if !msg_handler_mgr.Init() {
		log.Error("msg_handler_mgr init failed !")
		return
	} else {
		log.Info("msg_handler_mgr init succeed !")
	}

	if !player_mgr.Init() {
		log.Error("player_mgr init failed !")
		return
	} else {
		log.Info("player_mgr init succeed !")
	}

	if nil != dbc.Preload() {
		log.Error("dbc Preload Failed !!")
		return
	} else {
		log.Info("dbc Preload succeed !!")
	}

	if !login_token_mgr.Init() {
		log.Error("启动login_token_mgr失败")
		return
	}

	if !login_conn_mgr.Init() {
		log.Error("login_conn_mgr init failed")
		return
	}

	if !payback_mgr.Init() {
		log.Error("payback_mgr init failed")
		return
	} else {
		log.Info("payback_mgr init succeed !")
	}

	if !notice_mgr.Init() {
		log.Error("notice_mgr init failed")
		return
	} else {
		log.Info("notice_mgr init succeed !")
	}

	// 初始化CenterServer
	center_conn.Init()
	// 初始化大厅
	if !hall_server.Init() {
		log.Error("hall_server init failed !")
		return
	} else {
		log.Info("hall_server init succeed !")
	}

	if signal_mgr.IfClosing() {
		return
	}

	// 连接CenterServer
	log.Info("连接中心服务器！！")
	go center_conn.Start()
	center_conn.WaitConnectFinished()

	after_center_match_conn()

	hall_server.Start()
}
