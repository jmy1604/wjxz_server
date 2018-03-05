package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"libs/log"
	"libs/timer"
	"net"
	"net/http"
	"public_message/gen_go/server_message"
	"strings"
	"sync"
	"time"

	"3p/code.google.com.protobuf/proto"
)

type WaitCenterInfo struct {
	res_chan    chan *msg_server_message.C2LPlayerAccInfo
	create_time int32
}

type LoginServer struct {
	start_time         time.Time
	quit               bool
	shutdown_lock      *sync.Mutex
	shutdown_completed bool
	ticker             *timer.TickTimer
	initialized        bool

	login_http_listener net.Listener

	acc2c_wait      map[string]*WaitCenterInfo
	acc2c_wait_lock *sync.RWMutex
}

var server *LoginServer

func (this *LoginServer) Init() (ok bool) {
	this.start_time = time.Now()
	this.shutdown_lock = &sync.Mutex{}
	this.acc2c_wait = make(map[string]*WaitCenterInfo)
	this.acc2c_wait_lock = &sync.RWMutex{}

	this.initialized = true

	return true
}

func (this *LoginServer) Start() (err error) {
	go this.StartHttp()

	log.Event("服务器已启动", nil, log.Property{"IP", config.ListenClientIP})
	log.Trace("**************************************************")

	this.Run()

	return
}

func (this *LoginServer) Run() {
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

func (this *LoginServer) Shutdown() {
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

	this.login_http_listener.Close()
	center_conn.ShutDown()
	hall_agent_manager.net.Shutdown()
	log.Trace("关闭游戏主循环耗时 %v 秒", time.Now().Sub(begin).Seconds())
}

func (this *LoginServer) OnTick(t timer.TickTime) {
}

func (this *LoginServer) add_to_c_wait(acc string, c_wait *WaitCenterInfo) {
	this.acc2c_wait_lock.Lock()
	defer this.acc2c_wait_lock.Unlock()

	this.acc2c_wait[acc] = c_wait
}

func (this *LoginServer) remove_c_wait(acc string) {
	this.acc2c_wait_lock.Lock()
	defer this.acc2c_wait_lock.Unlock()

	delete(this.acc2c_wait, acc)
}

func (this *LoginServer) get_c_wait_by_acc(acc string) *WaitCenterInfo {
	this.acc2c_wait_lock.RLock()
	defer this.acc2c_wait_lock.RUnlock()

	return this.acc2c_wait[acc]
}

func (this *LoginServer) pop_c_wait_by_acc(acc string) *WaitCenterInfo {
	this.acc2c_wait_lock.Lock()
	defer this.acc2c_wait_lock.Unlock()

	cur_wait := this.acc2c_wait[acc]
	if nil != cur_wait {
		delete(this.acc2c_wait, acc)
		return cur_wait
	}

	return nil
}

//=================================================================================

type LoginHttpHandle struct{}

func (this *LoginServer) StartHttp() {
	var err error
	this.reg_http_mux()

	this.login_http_listener, err = net.Listen("tcp", config.ListenClientIP)
	if nil != err {
		log.Error("LoginServer StartHttp Failed %s", err.Error())
		return
	}

	login_http_server := http.Server{
		Handler:     &LoginHttpHandle{},
		ReadTimeout: 6 * time.Second,
	}

	err = login_http_server.Serve(this.login_http_listener)
	if err != nil {
		log.Error("启动Login Http Server %s", err.Error())
		return
	}
}

var login_http_mux map[string]func(http.ResponseWriter, *http.Request)

func (this *LoginServer) reg_http_mux() {
	login_http_mux = make(map[string]func(http.ResponseWriter, *http.Request))
	login_http_mux["/login"] = login_http_handler
}

func (this *LoginHttpHandle) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var act_str, url_str string
	url_str = r.URL.String()
	idx := strings.Index(url_str, "?")
	if -1 == idx {
		act_str = url_str
	} else {
		act_str = string([]byte(url_str)[:idx])
	}
	log.Info("ServeHTTP actstr(%s)", act_str)
	if h, ok := login_http_mux[act_str]; ok {
		h(w, r)
	}

	return
}

type JsonLoginRes struct {
	Code          int32
	Account       string
	Token         string
	HallIP        string
	ForbidEndTime string
}

func login_http_handler(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
			return
		}
	}()

	account_str := r.URL.Query().Get("account")
	if "" == account_str {
		log.Error("login_http_handler get account failed")
		return
	}

	/*
		cur_wait := server.get_c_wait_by_acc(account_str)
		if nil != cur_wait {
			log.Error("login_http_handler already in login !!")
			return
		}
	*/

	res_2c := &msg_server_message.L2CGetPlayerAccInfo{}
	res_2c.Account = proto.String(account_str)
	center_conn.Send(res_2c)

	log.Info("login_http_handler account(%s)", account_str)
	new_c_wait := &WaitCenterInfo{}
	new_c_wait.res_chan = make(chan *msg_server_message.C2LPlayerAccInfo)
	new_c_wait.create_time = int32(time.Now().Unix())
	server.add_to_c_wait(account_str, new_c_wait)

	c2l_res, ok := <-new_c_wait.res_chan
	if !ok || nil == c2l_res {
		log.Error("login_http_handler wait chan failed", ok)
		return
	}

	// 检查是否被封号
	if 1 == c2l_res.GetIfForbidLogin() {
		log.Info("login_http_handler account %d forbid end time %s", account_str, c2l_res.GetForbidEndTime())
		http_res := &JsonLoginRes{Code: -1, Account: account_str, ForbidEndTime: c2l_res.GetForbidEndTime()}
		data, err := json.Marshal(http_res)
		if nil != err {
			log.Error("login_http_handler json mashal error")
			return
		}
		w.Write(data)
		return
	} else {
		token_str := r.URL.Query().Get("token")
		if "" == token_str {
			log.Error("login_http_handler token empty")
			return
		}

		hall_id := c2l_res.GetHallId()
		hall_agent := hall_agent_manager.GetAgentByID(hall_id)
		if nil == hall_agent {
			log.Error("login_http_handler get hall_agent failed")
			http_res := &JsonLoginRes{Code: -1}
			data, err := json.Marshal(http_res)
			if nil == err {
				w.Write(data)
			} else {
				log.Error("login_http_handler return -1 marshal json error !")
			}

			return
		}

		inner_token := fmt.Sprintf("%d", time.Now().Unix())
		req_2h := &msg_server_message.L2HSyncAccountToken{}
		req_2h.Account = proto.String(account_str)
		req_2h.Token = proto.String(inner_token)
		req_2h.PlayerId = proto.Int64(c2l_res.GetPlayerId())
		hall_agent.Send(req_2h)

		http_res := &JsonLoginRes{Code: 0, Account: account_str, Token: inner_token, HallIP: c2l_res.GetHallIP()}
		data, err := json.Marshal(http_res)
		if nil != err {
			log.Error("login_http_handler json mashal error")
			return
		}
		w.Write(data)
	}

	return
}

func Google_Login_Verify(token string) bool {
	if "" == token {
		log.Error("Apple_Login_verify param token(%s) empty !", token)
		return false
	}

	url_str := global_config.GoogleLoginVerifyUrl + "?id_token=" + token

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(url_str)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	log.Info("%v", body)
	if 200 != resp.StatusCode {
		log.Error("Apple_Login_verify token failed(%d)", resp.StatusCode)
		return false
	}

	return true
}

func Apple_Login_Verify(token string) bool {
	return true
}
