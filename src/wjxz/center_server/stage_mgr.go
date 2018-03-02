package main

import (
	"libs/log"
	"libs/server_conn"
	"public_message/gen_go/server_message"
	"sync"
	"time"

	"3p/code.google.com.protobuf/proto"
)

type StagePassSession struct {
	player_id   int32
	session_id  int32
	res_2h      *msg_server_message.RetPlayerStagePass
	create_unix int32
}

type StageRankCache struct {
	small_rank      *SmallRankService
	last_cache_unix int32
	rank_msgs       []*msg_server_message.PlayerStageInfo
	db              *dbStageScoreRankRow
}

func (this *StageRankCache) ChkUpdateRankCache() {
	cur_unix := int32(time.Now().Unix())
	if this.last_cache_unix <= 0 {
		all_rds := this.small_rank.GetTopN(10)
		tmp_len := int32(len(all_rds))
		if tmp_len > 0 {
			var tmp_info *msg_server_message.PlayerStageInfo
			var tmp_item *SmallRankRecord
			tmp_rank_msgs := make([]*msg_server_message.PlayerStageInfo, 0, tmp_len)
			for idx := int32(0); idx < tmp_len; idx++ {
				tmp_item = all_rds[idx]
				if nil == tmp_item {
					continue
				}

				tmp_info = &msg_server_message.PlayerStageInfo{}
				tmp_info.PlayerId = proto.Int32(tmp_item.id)
				tmp_info.Name = proto.String(tmp_item.name)
				tmp_info.Lvl = proto.Int32(tmp_item.lvl)
				tmp_info.Icon = proto.String(tmp_item.icon)
				tmp_info.CustomIcon = proto.String(tmp_item.custom_icon)
				tmp_rank_msgs = append(tmp_rank_msgs, tmp_info)
			}
			this.rank_msgs = tmp_rank_msgs
		}
		this.last_cache_unix = cur_unix

	} else if cur_unix-this.last_cache_unix > 1 {
		bchg, all_rds := this.small_rank.GetAllIfChanged()
		tmp_len := int32(len(all_rds))
		if bchg {
			var tmp_info *msg_server_message.PlayerStageInfo
			var tmp_item *SmallRankRecord
			tmp_rank_msgs := make([]*msg_server_message.PlayerStageInfo, 0, tmp_len)
			for idx := int32(0); idx < tmp_len; idx++ {
				tmp_item = all_rds[idx]
				if nil == tmp_item {
					continue
				}

				tmp_info = &msg_server_message.PlayerStageInfo{}
				tmp_info.PlayerId = proto.Int32(tmp_item.id)
				tmp_info.Name = proto.String(tmp_item.name)
				tmp_info.Lvl = proto.Int32(tmp_item.lvl)
				tmp_info.Icon = proto.String(tmp_item.icon)
				tmp_info.CustomIcon = proto.String(tmp_item.custom_icon)
				tmp_rank_msgs = append(tmp_rank_msgs, tmp_info)
			}
			this.rank_msgs = tmp_rank_msgs
		}

		this.last_cache_unix = cur_unix
	}
}

type StageMgr struct {
	stage2rank map[int32]*StageRankCache

	pid2session_lock *sync.RWMutex
	pid2session      map[int32]*StagePassSession
}

var stage_mgr StageMgr

func (this *StageMgr) Init() bool {
	this.pid2session_lock = &sync.RWMutex{}
	this.pid2session = make(map[int32]*StagePassSession)
	// 初始化排行榜
	this.stage2rank = make(map[int32]*StageRankCache)
	var tmp_db *dbStageScoreRankRow
	var tmp_cache *StageRankCache
	for stageid, cfg := range cfg_stage_mgr.Map {
		if nil == cfg {
			continue
		}

		tmp_db = dbc.StageScoreRanks.GetRow(stageid)
		if nil == tmp_db {
			tmp_db = dbc.StageScoreRanks.AddRow(stageid)
		}

		if nil == tmp_db {
			log.Error("StageMgr Init tmp_db nil !")
			return false
		}

		tmp_cache = &StageRankCache{}
		tmp_cache.db = tmp_db
		tmp_cache.small_rank = NewSmallRankService(10, SMALL_RANK_TYPE_STAGE_SCORE, SMALL_RANK_SORT_TYPE_B)
		this.stage2rank[stageid] = tmp_cache
	}

	this.RegMsgHandler()

	return true
}

func (this *StageMgr) IfPlayerHaveSession(pid int32) bool {
	this.pid2session_lock.RLock()
	defer this.pid2session_lock.RUnlock()
	if nil != this.pid2session[pid] {
		return true
	}

	return false
}

func (this *StageMgr) AddSessionByMsg(msg *msg_server_message.PlayerStagePass) bool {
	if nil == msg {
		return false
	}

	if this.IfPlayerHaveSession(msg.GetSessionId()) {
		log.Error("StageMgr AddSessionByMsg already exist !")
		return false
	}

	new_session := &StagePassSession{}
	new_session.session_id = msg.GetSessionId()
	new_session.player_id = msg.GetPlayerId()
	new_session.res_2h = &msg_server_message.RetPlayerStagePass{}
	new_session.res_2h.PlayerId = proto.Int32(new_session.player_id)
	new_session.res_2h.FriendInfos = make([]*msg_server_message.PlayerStageInfo, 0, len(msg.GetFriendIds()))
	new_session.res_2h.SessionId = proto.Int32(new_session.session_id)
	stage_id := msg.GetStageId()
	tmp_rank := this.stage2rank[stage_id]
	if nil != tmp_rank {
		top_score := msg.GetTopScore()
		if top_score > 0 {
			tmp_rank.small_rank.SetUpdateRank(new_session.player_id, top_score, msg.GetLvl(), msg.GetIcon(), msg.GetName(), msg.GetCustomIcon())
		}
		tmp_rank.ChkUpdateRankCache()
		new_session.res_2h.TopPlayers = tmp_rank.rank_msgs
	}

	new_session.create_unix = int32(time.Now().Unix())

	this.pid2session_lock.Lock()
	defer this.pid2session_lock.Unlock()

	this.pid2session[msg.GetSessionId()] = new_session

	return true
}

func (this *StageMgr) AddFriendInfoToSession(msg *msg_server_message.RetFriendStageInfo) {
	if nil == msg || nil == msg.GetInfo() {
		log.Error("StageMgr AddFriendInfoToSession msg nil")
		return
	}

	req_pid := msg.GetReqPId()
	this.pid2session_lock.Lock()
	defer this.pid2session_lock.Unlock()
	cur_session := this.pid2session[req_pid]
	if nil == cur_session {
		return
	}

	cur_session.res_2h.FriendInfos = append(cur_session.res_2h.FriendInfos, msg.GetInfo())

	return
}

func (this *StageMgr) PopSessionById(pid int32) *StagePassSession {
	this.pid2session_lock.Lock()
	defer this.pid2session_lock.Unlock()

	return this.pid2session[pid]
}

// ===========================================================================

func (this *StageMgr) RegMsgHandler() {
	hall_agent_mgr.SetMessageHandler(msg_server_message.ID_PlayerStagePass, this.H2CPlayerStagePassHandler)
	hall_agent_mgr.SetMessageHandler(msg_server_message.ID_ChapterUnlockHelp, this.H2CChapterUnlockHelpHandler)
}

func (this *StageMgr) PlayerStagePassWait(c *server_conn.ServerConn, pid, stage_id int32, friendids []int32) {
	tmp_len := int32(len(friendids))
	if tmp_len > 0 {
		var tmp_svr_info *SingleHallCfg
		var tmp_hall *HallAgent
		res_2h := &msg_server_message.GetFriendStageInfo{}
		res_2h.FriendId = proto.Int32(0)
		res_2h.ReqPId = proto.Int32(pid)
		res_2h.StageId = proto.Int32(stage_id)
		var tmp_pid int32
		for idx := int32(0); idx < tmp_len; idx++ {
			tmp_pid = friendids[idx]
			if tmp_pid <= 0 {
				continue
			}

			tmp_svr_info = hall_group_mgr.GetHallCfgByPlayerId(tmp_pid)
			if nil == tmp_svr_info {
				continue
			}

			tmp_hall = hall_agent_mgr.GetAgentById(tmp_pid)
			if nil == tmp_hall {
				continue
			}

			*res_2h.FriendId = tmp_pid

			tmp_hall.Send(res_2h)

		}
		time.Sleep(time.Second)
	}

	return
}

func (this *StageMgr) H2CPlayerStagePassHandler(c *server_conn.ServerConn, msg proto.Message) {
	req := msg.(*msg_server_message.PlayerStagePass)
	if nil == c || nil == req {
		log.Error("H2CPlayerStagePassHandler c or req nil [%v]", nil == req)
		return
	}

	this.AddSessionByMsg(req)

	this.PlayerStagePassWait(c, req.GetPlayerId(), req.GetStageId(), req.GetFriendIds())

	tmp_session := this.PopSessionById(req.GetSessionId())
	if nil != tmp_session {
		c.Send(tmp_session.res_2h, true)
	} else {
		log.Error("H2CPlayerStagePassHandler PopSessionById[%d] nil !", req.GetSessionId())
	}

	return
}

func (this *StageMgr) H2CRetFriendStageInfoHandler(c *server_conn.ServerConn, msg proto.Message) {
	req := msg.(*msg_server_message.RetFriendStageInfo)
	if nil == c || nil == req {
		log.Error("H2CRetFriendStageInfoHandler c or req nil ")
		return
	}

	stage_mgr.AddFriendInfoToSession(req)

	return
}

func (this *StageMgr) H2CChapterUnlockHelpHandler(c *server_conn.ServerConn, msg proto.Message) {
	req := msg.(*msg_server_message.ChapterUnlockHelp)
	if nil == c || nil == req {
		log.Error("H2CChapterUnlockHelpHandler c or req nil !")
		return
	}

	req_pids := req.GetHelpPlayerIds()

	for _, pid := range req_pids {

		hall_svrinfo := hall_group_mgr.GetHallCfgByPlayerId(pid)
		if nil == hall_svrinfo {
			log.Error("H2CChapterUnlockHelpHandler failed to get hall_svrinfo[%d]", pid)
			continue
		}

		hall_svr := hall_agent_mgr.GetAgentById(hall_svrinfo.ServerId)
		if nil == hall_svr {
			log.Error("H2CChapterUnlockHelpHandler failed to get hall_svr [%d]", hall_svrinfo.ServerId)
			return
		}

		hall_svr.Send(req)
	}

	return
}
