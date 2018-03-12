package main

import (
	"libs/log"
	"libs/timer"
	"main/table_config"
	"math/rand"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/server_message"
	"sync"
	"time"

	"3p/code.google.com.protobuf/proto"
)

const (
	FINISHE_ALL_STAR = 3
)

type StagePassSession struct {
	SessionId    int32 // 会话Id
	ret          *msg_client_message.S2CStagePass
	b_center_ret bool
}

type StagePassMgr struct {
	sessions_lock *sync.RWMutex
	id2session    map[int32]*StagePassSession

	session_id_lock *sync.Mutex
	nex_session_id  int32
}

var stage_pass_mgr StagePassMgr

func (this *StagePassMgr) Init() bool {
	this.sessions_lock = &sync.RWMutex{}
	this.id2session = make(map[int32]*StagePassSession)

	return true
}

func (this *StagePassMgr) GetNextSessionId() int32 {
	this.sessions_lock.Lock()
	defer this.sessions_lock.Unlock()

	this.nex_session_id++
	return this.nex_session_id
}

func (this *StagePassMgr) AddSession(session *StagePassSession) {
	if nil == session {
		log.Error("StagePassMgr AddSession session nil !")
		return
	}

	this.sessions_lock.Lock()
	defer this.sessions_lock.Unlock()

	this.id2session[session.SessionId] = session

	return
}

func (this *StagePassMgr) PopSessionById(sid int32) *StagePassSession {
	this.sessions_lock.Lock()
	defer this.sessions_lock.Unlock()

	cur_ssession := this.id2session[sid]
	if nil == cur_ssession {
		return nil
	}

	delete(this.id2session, sid)

	return cur_ssession
}

// ============================================================================

func (this *Player) GetDayBuyTiLiCount() int32 {
	cur_unix_day := timer.GetDayFrom1970WithCfg(0)
	if cur_unix_day != this.db.Info.GetDayBuyTiLiUpDay() {
		this.db.Info.SetDayBuyTiLiCount(0)
		this.db.Info.SetDayBuyTiLiUpDay(cur_unix_day)
		return 0
	}

	return this.db.Info.GetDayBuyTiLiCount()
}

// ============================================================================

func (this *Player) CheckBeginStage(data *StageBeginData) bool {
	level := level_table_mgr.GetLevel(data.stage_id)
	if level == nil {
		return false
	}

	if level.NeedPower > this.CalcSpirit() {
		log.Error("Player[%v] not enough stamina to begin stage[%v]", this.Id, data.stage_id)
		return false
	}

	if this.stage_state == 1 {
		log.Warn("Player[%v] already begin stage[%v]", this.Id, data.stage_id)
	}

	this.SubSpirit(level.NeedPower, "pass_stage", "stage")

	if data.item_ids != nil {
		items := make(map[int32]int32)
		for i := 0; i < len(data.item_ids); i++ {
			if items[data.item_ids[i]] == 0 {
				items[data.item_ids[i]] = 1
			} else {
				items[data.item_ids[i]] += 1
			}
		}
		for k, v := range items {
			num := this.GetItemResourceValue(k)
			if num < v {
				log.Error("Player[%v] begin stage[%v] with item[%v,%v] not enough", this.Id, data.stage_id, k, num)
				return false
			}
		}
		for i := 0; i < len(data.item_ids); i++ {
			this.RemoveItemResource(data.item_ids[i], 1, "begin_stage", "stage")
		}
		this.SendItemsUpdate()
	}

	this.stage_id = data.stage_id
	this.stage_cat_id = data.cat_id
	this.stage_state = 1

	return true
}

func (this *Player) ChkFinishStage(stageid, star, score int32, ret_msg *msg_client_message.S2CStagePass, bforce bool) (top_score int32) {
	if this.stage_id != stageid {
		return int32(msg_client_message.E_ERR_STAGE_NO_MATCH_WITH_END)
	}
	if this.stage_state == 0 {
		return int32(msg_client_message.E_ERR_STAGE_ALREADY_FINISHED)
	}

	if !bforce {
		if stageid > this.db.Info.GetMaxUnlockStage() {
			return int32(msg_client_message.E_ERR_STAGE_PASS_NOT_UNLOCK)
		}

		cur_max_stage_id := this.db.Info.GetCurMaxStage()
		if cur_max_stage_id > 0 && stageid > cur_max_stage_id {
			cur_max_stage_cfg := level_table_mgr.Map[cur_max_stage_id]
			if nil == cur_max_stage_cfg {
				log.Error("Player ChkFinishStage faild to find cur_max_stage_cfg[%d] !", cur_max_stage_id)
				return int32(msg_client_message.E_ERR_STAGE_PASS_NOT_UNLOCK)
			}

			if cur_max_stage_cfg.NextLevel != stageid {
				return int32(msg_client_message.E_ERR_STAGE_PASS_OVER_NEXT_STATE)
			}
		}
	}

	stagecfg := level_table_mgr.Map[stageid]
	if nil == stagecfg {
		log.Error("Player ChkFinishStage failed to find stage[%d]", stageid)
		return int32(msg_client_message.E_ERR_STAGE_TABLE_DATA_NOT_FOUND)
	}

	bfirst := false
	bfirst_3star := false
	cur_stage_db := this.db.Stages.Get(stageid)
	old_top_score := int32(0)
	if nil == cur_stage_db {
		new_db := &dbPlayerStageData{}
		new_db.LastFinishedUnix = int32(time.Now().Unix())
		new_db.StageId = stageid
		new_db.Stars = star
		new_db.TopScore = score
		this.db.Stages.Add(new_db)
		bfirst = true
		top_score = score
		this.AddStar(star, "pass_stage", "stage")
		if star >= FINISHE_ALL_STAR {
			bfirst_3star = true
		}
	} else {
		old_top_score, _ = this.db.Stages.GetTopScore(stageid)
		top_score = this.db.Stages.ChkSetTopScore(stageid, score)
		if star > cur_stage_db.Stars {
			this.db.Stages.SetStars(stageid, star)
			this.AddStar(star-cur_stage_db.Stars, "pass_stage", "stage")
			if cur_stage_db.Stars < FINISHE_ALL_STAR && star >= FINISHE_ALL_STAR {
				bfirst_3star = true
			}
		}
	}

	// update ranking list, the score is top score
	if score > old_top_score {
		if this.rpc_rank_update_stage_total_score(this.db.Stages.GetTotalScore()) == nil {
			log.Warn("Player[%v] update stages total score failed", this.Id)
		}
		if this.rpc_rank_update_stage_score(stageid, top_score) == nil {
			log.Warn("Player[%v] update stage[%v] top score[%v] failed", this.Id, stageid, score)
		}
	}

	if this.db.Info.ChkSetCurMaxStage(stageid) {
		this.b_base_prop_chg = true
	}

	// 给予首次通关奖励
	var tmp_item *msg_client_message.ItemInfo

	ret_msg.Getitems = make([]*msg_client_message.ItemInfo, 0)
	if bfirst {
		ret_msg.GetitemsFirst = make([]*msg_client_message.ItemInfo, 0, len(stagecfg.FirstClearReward)/2)
		log.Info("首次通关[%d]给予奖励 %v", stageid, stagecfg.FirstClearReward)
		for i := 0; i < len(stagecfg.FirstClearReward)/2; i++ {
			tmp_item = &msg_client_message.ItemInfo{}
			tmp_item.ItemCfgId = proto.Int32(stagecfg.FirstClearReward[2*i])
			tmp_item.ItemNum = proto.Int32(stagecfg.FirstClearReward[2*i+1])
			ret_msg.GetitemsFirst = append(ret_msg.GetitemsFirst, tmp_item)

			this.AddItemResource(stagecfg.FirstClearReward[2*i], stagecfg.FirstClearReward[2*i+1], "FirstClearReward", "Stage")
		}
	}
	if /*FINISHE_ALL_STAR == star && (bfirst || cur_stage_db.Stars < FINISHE_ALL_STAR)*/ bfirst_3star {
		ret_msg.Getitems3Star = make([]*msg_client_message.ItemInfo, 0, len(stagecfg.FirstAllStarReward)/2)
		for i := 0; i < len(stagecfg.FirstAllStarReward)/2; i++ {
			tmp_item = &msg_client_message.ItemInfo{}
			tmp_item.ItemCfgId = proto.Int32(stagecfg.FirstAllStarReward[2*i])
			tmp_item.ItemNum = proto.Int32(stagecfg.FirstAllStarReward[2*i+1])
			ret_msg.Getitems3Star = append(ret_msg.Getitems3Star, tmp_item)

			this.AddItemResource(stagecfg.FirstAllStarReward[2*i], stagecfg.FirstAllStarReward[2*i+1], "StageFirstAllStar", "Stage")
		}
		log.Debug("Player[%v] First Finish Stage[%v], Stamina[%v]", this.Id, stageid, this.db.Info.GetSpirit())
	}

	// 额外增加的金币
	extra_coin := int32(0)
	ret_msg.CatExtraAddCoin = proto.Int32(extra_coin)
	log.Debug("@@@@ stage_cat_id[%v] extra_coin[%v]", this.stage_cat_id, extra_coin)

	// 给予普通奖励
	this.AddCoin(stagecfg.CoinReward+extra_coin, "StagePass", "Stage")
	ret_msg.GetCoin = proto.Int32(stagecfg.CoinReward + extra_coin)
	coin_item := &msg_client_message.ItemInfo{
		ItemCfgId: proto.Int32(ITEM_RESOURCE_ID_GOLD),
		ItemNum:   proto.Int32(stagecfg.CoinReward + extra_coin),
	}
	ret_msg.Getitems = append(ret_msg.Getitems, coin_item)

	var b bool
	if len(stagecfg.ExtraReward1) == 2 && rand.Int31n(100) < stagecfg.ExtraReward1[1] {
		//this.AddItem(stagecfg.ExtraReward1, 1, "StagePass", "Stage", true)
		b, tmp_item = this.drop_item_by_id(stagecfg.ExtraReward1[0], false)
		if b {
			if tmp_item != nil {
				ret_msg.Getitems = append(ret_msg.Getitems, tmp_item)
			}
		}
	}

	if len(stagecfg.ExtraReward2) == 2 && rand.Int31n(100) < stagecfg.ExtraReward2[1] {
		//this.AddItem(stagecfg.ExtraReward2, 1, "StagePass", "Stage", true)
		b, tmp_item = this.drop_item_by_id(stagecfg.ExtraReward2[0], false)
		if b {
			if tmp_item != nil {
				ret_msg.Getitems = append(ret_msg.Getitems, tmp_item)
			}
		}
	}

	this.stage_state = 0

	return
}

// ============================================================================

func reg_player_stage_msg() {
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SStageBegin, C2SStagePassBeginHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SStagePass, C2SStagePassHandler)
	center_conn.SetMessageHandler(msg_server_message.ID_GetFriendStageInfo, C2HGetFriendStageInfoHandler)
	center_conn.SetMessageHandler(msg_server_message.ID_RetPlayerStagePass, C2HRetPlayerStagePassHandler)
}

type StageBeginData struct {
	stage_id int32
	cat_id   int32
	item_ids []int32
}

func C2SStagePassBeginHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SStageBegin)
	if req == nil {
		log.Error("C2SStageBegin proto is invalid")
		return -1
	}

	var data = StageBeginData{
		stage_id: req.GetStageId(),
		cat_id:   req.GetCatId(),
		item_ids: req.GetItemIds(),
	}
	if !p.CheckBeginStage(&data) {
		return -1
	}

	response := &msg_client_message.S2CStageBeginResult{}
	response.StageId = proto.Int32(req.GetStageId())
	p.Send(response)
	return 1
}

func (this *dbPlayerStageColumn) GetTotalScore() int32 {
	this.m_row.m_lock.UnSafeRLock("dbPlayerStageColumn.GetTotalScore")
	defer this.m_row.m_lock.UnSafeRUnlock()

	total_score := int32(0)
	for _, v := range this.m_data {
		total_score += v.TopScore
	}
	return total_score
}

func (p *Player) stage_pass(result int32, stageid int32, score int32, stars int32, items []*msg_client_message.ItemInfo, bforce bool) int32 {
	new_session := &StagePassSession{}
	new_session.SessionId = stage_pass_mgr.GetNextSessionId()
	tmp_ret := &msg_client_message.S2CStagePass{}
	tmp_ret.StageId = proto.Int32(stageid)
	tmp_ret.Score = proto.Int32(score)
	tmp_ret.Stars = proto.Int32(stars)
	tmp_ret.UseItems = items
	tmp_ret.Result = proto.Int32(result)

	// 未过关
	if result == 0 {
		tmp_ret.FriendItems = make([]*msg_client_message.PlayerStageInfo, 0)
		tmp_ret.GetCoin = proto.Int32(0)
		tmp_ret.Getitems = make([]*msg_client_message.ItemInfo, 0)
		tmp_ret.RankItems = make([]*msg_client_message.PlayerStageInfo, 0)
		p.Send(tmp_ret)
		return 1
	}

	top_score := p.ChkFinishStage(stageid, stars, score, tmp_ret, bforce)
	if top_score < 0 {
		return top_score
	}

	if result > 0 {
		p.db.Stages.IncbyPassCount(stageid, 1)
	}
	p.db.Stages.IncbyPlayedCount(stageid, 1)

	new_session.ret = tmp_ret
	new_session.ret.TopScore = proto.Int32(top_score)

	// 物品消耗
	if items != nil {
		for i := 0; i < len(items); i++ {
			p.RemoveItem(items[i].GetItemCfgId(), items[i].GetItemNum(), true)
		}
	}

	// 更新任务
	p.TaskUpdate(table_config.TASK_FINISH_PASS_NUM, false, 0, 1)
	level := level_table_mgr.GetLevel(stageid)
	if level != nil {
		chapter_levels := level_table_mgr.GetChapter(level.Chapter)
		if chapter_levels != nil {
			c := true
			for _, v := range chapter_levels.Levels {
				if !p.db.Stages.HasIndex(v.Id) {
					c = false
					break
				}
			}
			if c {
				p.TaskUpdate(table_config.TASK_FINISH_PASS_CHAPTER, false, stageid, 1)
			}
		}
	}

	// 获取好友关卡数据并排序
	r := p.rpc_get_friends_stage_info(stageid)
	if r != nil {
		for _, info := range r.StageInfos {
			item := &msg_client_message.PlayerStageInfo{
				PlayerId: proto.Int32(info.PlayerId),
				Name:     proto.String(info.Name),
				Lvl:      proto.Int32(info.Level),
				Icon:     proto.String(info.Head),
				Score:    proto.Int32(info.TopScore),
			}
			new_session.ret.FriendItems = append(new_session.ret.FriendItems, item)
		}
	}

	log.Info("Stage Pass res %v", new_session.ret)

	p.SendItemsUpdate()
	p.Send(new_session.ret)
	p.send_stage_info()

	return 1
}

func C2SStagePassHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SStagePass)
	if nil == req {
		log.Error("C2SStagePassHandler p[%d:%v] or req nil", nil == p)
		return -1
	}

	stageid := req.GetStageId()
	score := req.GetScore()
	stars := req.GetStars()

	return p.stage_pass(req.GetResult(), stageid, score, stars, req.GetItems(), false)
}

func C2SDayBuyTiLiHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	global_cfg := global_config_mgr.GetGlobalConfig()
	cur_count := p.GetDayBuyTiLiCount()
	if cur_count >= global_cfg.MaxDayBuyTiLiCount {
		return int32(msg_client_message.E_ERR_DAYBUY_TILI_MAX_COUNT)
	}

	if p.GetDiamond() < global_cfg.DayBuyTiLiCost {
		return int32(msg_client_message.E_ERR_DAYBUY_TILI_LESS_DIAMOND)
	}

	p.db.Info.SetDayBuyTiLiCount(cur_count + 1)
	p.SubDiamond(global_cfg.DayBuyTiLiCost, "DayBuyTiLi", "Stage")
	p.AddSpirit(global_cfg.DayBuyTiliAdd, "DayBuyTiLi", "Stage")

	return 1
}

// -------------------------

func C2HGetFriendStageInfoHandler(c *CenterConnection, msg proto.Message) {
	res := msg.(*msg_server_message.GetFriendStageInfo)
	if nil == res || nil == c {
		log.Error("C2HGetPlayerStageInfoHandler res or c nil !")
		return
	}

	pid := res.GetFriendId()
	tmp_p := player_mgr.GetPlayerById(pid)
	if nil == tmp_p {
		log.Error("C2HGetPlayerStageInfoHandler get p[%d] failed !", pid)
		return
	}

	/*res2co := &msg_server_message.RetFriendStageInfo{}
	tmp_info := &msg_server_message.PlayerStageInfo{}
	tmp_info.PlayerId = proto.Int32(tmp_p.Id)
	tmp_info.Name = proto.String(tmp_p.db.GetName())
	tmp_info.Icon = proto.String(tmp_p.db.Info.GetIcon())
	tmp_info.CustomIcon = proto.String(tmp_p.db.Info.GetCustomIcon())
	tmp_info.Score = proto.Int32(tmp_p.db.Stages.ChkGetTopScore(res.GetStageId()))
	res2co.Info = tmp_info
	res2co.ReqPId = proto.Int32(res.GetReqPId())

	c.Send(res2co)*/

	return
}

func C2HRetPlayerStagePassHandler(c *CenterConnection, msg proto.Message) {
	res := msg.(*msg_server_message.RetPlayerStagePass)
	if nil == res || nil == c {
		log.Error("C2HRetPlayerStagePassHandler c or res nil !")
		return
	}

	pid := res.GetPlayerId()
	p := player_mgr.GetPlayerById(res.GetPlayerId())
	if nil == p {
		log.Error("C2HRetPlayerStagePassHandler failed to find playerid [%d]", pid)
		return
	}

	cur_ssession := stage_pass_mgr.PopSessionById(res.GetSessionId())
	if nil == cur_ssession {
		log.Error("C2HRetPlayerStagePassHandler failed to find session [%d]", res.GetSessionId())
		return
	}

	res2cli := cur_ssession.ret

	rank_items := res.GetTopPlayers()
	rank_len := int32(len(rank_items))
	var tmp_svr_pinfo *msg_server_message.PlayerStageInfo
	var tmp_cli_pinfo *msg_client_message.PlayerStageInfo
	if rank_len > 0 {
		res2cli.RankItems = make([]*msg_client_message.PlayerStageInfo, 0, rank_len)
		for idx := int32(0); idx < rank_len; idx++ {
			tmp_svr_pinfo = rank_items[idx]
			tmp_cli_pinfo = &msg_client_message.PlayerStageInfo{}
			tmp_cli_pinfo.PlayerId = proto.Int32(tmp_svr_pinfo.GetPlayerId())
			tmp_cli_pinfo.Name = proto.String(tmp_svr_pinfo.GetName())
			tmp_cli_pinfo.Lvl = proto.Int32(tmp_svr_pinfo.GetLvl())
			tmp_cli_pinfo.Score = proto.Int32(tmp_svr_pinfo.GetScore())
			tmp_cli_pinfo.Icon = proto.String(tmp_svr_pinfo.GetIcon())
			tmp_cli_pinfo.CustomIcon = proto.String(tmp_svr_pinfo.GetCustomIcon())
			res2cli.RankItems = append(res2cli.RankItems, tmp_cli_pinfo)
		}
	}

	friend_items := res.GetFriendInfos()
	friend_len := int32(len(friend_items))
	if friend_len > 0 {
		res2cli.FriendItems = make([]*msg_client_message.PlayerStageInfo, 0, friend_len)
		for idx := int32(0); idx < friend_len; idx++ {
			tmp_svr_pinfo = friend_items[idx]
			tmp_cli_pinfo = &msg_client_message.PlayerStageInfo{}
			tmp_cli_pinfo.PlayerId = proto.Int32(tmp_svr_pinfo.GetPlayerId())
			tmp_cli_pinfo.Name = proto.String(tmp_svr_pinfo.GetName())
			tmp_cli_pinfo.Lvl = proto.Int32(tmp_svr_pinfo.GetLvl())
			tmp_cli_pinfo.Score = proto.Int32(tmp_svr_pinfo.GetScore())
			tmp_cli_pinfo.Icon = proto.String(tmp_svr_pinfo.GetIcon())
			tmp_cli_pinfo.CustomIcon = proto.String(tmp_svr_pinfo.GetCustomIcon())
			res2cli.FriendItems = append(res2cli.FriendItems, tmp_cli_pinfo)
		}
	}

	//p.Send(res2cli)

	cur_ssession.b_center_ret = true

	return
}
