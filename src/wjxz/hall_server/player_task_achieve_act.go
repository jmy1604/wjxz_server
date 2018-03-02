package main

/*
import (
	"libs/log"
	"libs/socket"
	"libs/timer"
	"public_message/gen_go/client_message"
	"time"

	"3p/code.google.com.protobuf/proto"
)

// 任务类型
const (
	TASK_ACHIEVE_FINISH_CON_ARENA_WIN    = 1 // 竞技场获胜
	TASK_ACHIEVE_FINISH_CON_ARENA_LVL    = 2 // 竞技场等级
	TASK_ACHIEVE_FINISH_CON_CARD_T_NUM   = 3 // 搜集卡牌种类
	TASK_ACHIEVE_FINISH_WIN_FRIEND_FIGHT = 4 // 赢得的友谊战
	TASK_ACHIEVE_FINISH_LEGEND_SCORES    = 5 // 应得的传奇奖杯数目
	TASK_ACHIEVE_FINISH_DIAMOND_COST     = 6 // 钻石消耗数目
	TASK_ACHIEVE_FINISH_COIN_COST        = 7 // 金币消耗数目
)

func (this *Player) ChkPlayerDialyTask() {
	//cur_unix := int32(time.Now().Unix())
	last_up_unix := this.db.Info.GetLastDialyTaskUpUinx()

	cur_unix_day := timer.GetDayFrom1970WithCfg(0)
	last_up_unix_day := timer.GetDayFrom1970WithCfgAndSec(0, last_up_unix)
	if cur_unix_day != last_up_unix_day {
		this.db.DialyTasks.ChkResetDialyTask()
		this.db.Info.SetLastDialyTaskUpUinx(int32(time.Now().Unix()))
	}
}

func (this *Player) SyncPlayerDialyTask() {
	this.ChkPlayerDialyTask()
	res2cli := this.db.DialyTasks.FillDialyTaskMsg()
	if nil == res2cli || len(res2cli.TaskList) < 1 {
		return
	}

	this.Send(res2cli)
	return
}

func (this *Player) SyncPlayerAchieve() {
	res2cli := this.db.Achieves.FillAchieveMsg()
	if nil == res2cli || len(res2cli.AchieveList) < 1 {
		return
	}

	this.Send(res2cli)
	return
}

func (this *Player) SyncPlayerActivity() {
	res2cli := this.db.SevenActivitys.FillSevenMsg(this)
	this.Send(res2cli)
	return
}

// ============================================================================

func (this *Player) TaskAchieveOnConditionAdd(con int32, add_val int32) {
	log.Info("进入任务成就触发add函数con[%d] add_val[%d]", con, add_val)
	var idx int32
	var cur_val int32
	var notify_task *msg_client_message.S2CNotifyTaskValueChg
	sub_dialys := achieve_task_mgr.Type2SubDialyTasks[con]
	if nil != sub_dialys && sub_dialys.Count > 0 {
		var tmp_taskcfg *XmlDialyTaskItem
		for idx = 0; idx < sub_dialys.Count; idx++ {
			tmp_taskcfg = sub_dialys.Array[idx]
			cur_dialy := this.db.DialyTasks.Get(tmp_taskcfg.DailyTasksID)
			if nil != cur_dialy {

				if cur_dialy.RewardUnix > 0 {
					continue
				}

				cur_val = this.db.DialyTasks.IncbyValue(tmp_taskcfg.DailyTasksID, add_val)
			} else {
				new_dialy := &dbPlayerDialyTaskData{}
				new_dialy.TaskId = tmp_taskcfg.DailyTasksID
				new_dialy.Value = add_val
				this.db.DialyTasks.Add(new_dialy)
				cur_val = add_val
			}

			notify_task = &msg_client_message.S2CNotifyTaskValueChg{}
			notify_task.TaskId = proto.Int32(tmp_taskcfg.DailyTasksID)
			notify_task.TaskValue = proto.Int32(cur_val)
			this.Send(notify_task)
		}
	} else {
		log.Error("Player TaskAchieveOnConditionAdd sub dialy nil or empty [%v]", nil == sub_dialys)
	}

	var notify_achieve *msg_client_message.S2CNotifyAchieveValueChg
	sub_achieves := achieve_task_mgr.Typ2SubAchieves[con]
	if nil != sub_achieves && sub_achieves.Count > 0 {
		var tmp_achievecfg *XmlAchieveItem
		for idx = 0; idx < sub_achieves.Count; idx++ {

			tmp_achievecfg = sub_achieves.Array[idx]

			cur_achieve := this.db.Achieves.Get(tmp_achievecfg.AchievementTaskID)
			if nil != cur_achieve {
				if cur_achieve.RewardUnix > 0 {
					continue
				}

				cur_val = this.db.Achieves.IncbyValue(tmp_achievecfg.AchievementTaskID, add_val)
				log.Info("成就[%d]已经有值[%d]", tmp_achievecfg.AchievementTaskID, cur_val)
			} else {
				tmp_achieve := &dbPlayerAchieveData{}
				tmp_achieve.AchieveId = tmp_achievecfg.AchievementTaskID
				tmp_achieve.Value = add_val
				this.db.Achieves.Add(tmp_achieve)
				cur_val = add_val
			}

			notify_achieve = &msg_client_message.S2CNotifyAchieveValueChg{}
			notify_achieve.AchieveId = proto.Int32(tmp_achievecfg.AchievementTaskID)
			notify_achieve.AchieveValue = proto.Int32(cur_val)
			this.Send(notify_achieve)

			log.Info("=====成就[%d]触发add函数con[%d] add_val[%d] cur_val[%d]", tmp_achievecfg.AchievementTaskID, con, add_val, cur_val)
		}
	} else {
		log.Error("Player TaskAchieveOnConditionAdd sub achieve nil or empty")
	}

	create_unix_day := timer.GetDayFrom1970WithCfgAndSec(0, this.db.Info.GetCreateUnix())
	cur_unix_day := timer.GetDayFrom1970WithCfg(0)

	var notify_seven *msg_client_message.S2CNotifySevenActValueChg
	sub_sevens := cfg_player_act_mgr.type2sevendayacts[con]
	if nil != sub_sevens && sub_sevens.Count > 0 {
		var tmp_sevencfg *XmlSevenDayItem
		for idx = 0; idx < sub_sevens.Count; idx++ {

			tmp_sevencfg = sub_sevens.Array[idx]
			if cfg_player_act_mgr.GetSevenDayLeftSec(create_unix_day, cur_unix_day, tmp_sevencfg.TasksID) <= 0 {
				log.Info("七天日活动[%d] 不在开启状态", tmp_sevencfg.TasksID, create_unix_day, cur_unix_day)
				continue
			}

			cur_seven := this.db.SevenActivitys.Get(tmp_sevencfg.TasksID)
			if nil != cur_seven {
				if cur_seven.RewardUnix > 0 {
					log.Info("七天日活动[%d] 已经领奖", tmp_sevencfg.TasksID)
					continue
				}

				cur_val = this.db.SevenActivitys.IncbyValue(tmp_sevencfg.TasksID, add_val)
				log.Info("七天活动[%d]已经有值[%d]", tmp_sevencfg.TasksID, cur_val)
			} else {
				tmp_seven := &dbPlayerSevenActivityData{}
				tmp_seven.ActivityId = tmp_sevencfg.TasksID
				tmp_seven.Value = add_val
				this.db.SevenActivitys.Add(tmp_seven)
				cur_val = add_val
			}

			notify_seven = &msg_client_message.S2CNotifySevenActValueChg{}
			notify_seven.ActivityId = proto.Int32(tmp_sevencfg.TasksID)
			notify_seven.ActivityValue = proto.Int32(cur_val)
			this.Send(notify_seven)

			log.Info("=====七天乐触发[%d]触发add函数con[%d] add_val[%d] cur_val[%d]", tmp_sevencfg.TasksID, con, add_val, cur_val)
		}
	} else {
		log.Error("Player TaskAchieveOnConditionAdd sub seven nil or empty")
	}
}

func (this *Player) TaskAchieveOnConditionSet(con int32, new_val int32, bslience bool) {
	log.Info("进入任务成就触发Set函数con[%d] new_val[%d]", con, new_val)
	var idx int32
	sub_dialys := achieve_task_mgr.Type2SubDialyTasks[con]
	var notify_task *msg_client_message.S2CNotifyTaskValueChg
	if nil != sub_dialys && sub_dialys.Count > 0 {
		var tmp_taskcfg *XmlDialyTaskItem
		for idx = 0; idx < sub_dialys.Count; idx++ {
			tmp_taskcfg = sub_dialys.Array[idx]
			cur_dialy := this.db.DialyTasks.Get(tmp_taskcfg.DailyTasksID)
			if nil != cur_dialy {
				if cur_dialy.RewardUnix > 0 {
					continue
				}
				this.db.DialyTasks.SetValue(tmp_taskcfg.DailyTasksID, new_val)
			} else {
				new_dialy := &dbPlayerDialyTaskData{}
				new_dialy.TaskId = tmp_taskcfg.DailyTasksID
				new_dialy.Value = new_val
				this.db.DialyTasks.Add(new_dialy)
			}

			if !bslience {
				notify_task = &msg_client_message.S2CNotifyTaskValueChg{}
				notify_task.TaskId = proto.Int32(tmp_taskcfg.DailyTasksID)
				notify_task.TaskValue = proto.Int32(new_val)
				this.Send(notify_task)
			}
		}
	} else {
		log.Error("Player TaskAchieveOnConditionSet sub dialy nil or empty [%v]", nil == sub_dialys)
	}

	sub_achieves := achieve_task_mgr.Typ2SubAchieves[con]
	var notify_achieve *msg_client_message.S2CNotifyAchieveValueChg
	if nil != sub_achieves && sub_achieves.Count > 0 {
		var tmp_achievecfg *XmlAchieveItem
		for idx = 0; idx < sub_achieves.Count; idx++ {
			tmp_achievecfg = sub_achieves.Array[idx]
			cur_achieve := this.db.Achieves.Get(tmp_achievecfg.AchievementTaskID)
			if nil != cur_achieve {
				if cur_achieve.RewardUnix > 0 {
					//log.Info("成就[%d]已经领取不能触发 %v", tmp_achievecfg.AchievementTaskID, cur_achieve.RewardUnix)
					continue
				}

				this.db.Achieves.SetValue(tmp_achievecfg.AchievementTaskID, new_val)
			} else {
				tmp_achieve := &dbPlayerAchieveData{}
				tmp_achieve.AchieveId = tmp_achievecfg.AchievementTaskID
				tmp_achieve.Value = new_val
				this.db.Achieves.Add(tmp_achieve)
			}

			if !bslience {
				notify_achieve = &msg_client_message.S2CNotifyAchieveValueChg{}
				notify_achieve.AchieveId = proto.Int32(tmp_achievecfg.AchievementTaskID)
				notify_achieve.AchieveValue = proto.Int32(new_val)
				this.Send(notify_achieve)
			}

			log.Info("=====成就[%d]触发Set函数con[%d] new_val[%d]", tmp_achievecfg.AchievementTaskID, con, new_val)
		}
	} else {
		log.Error("Player TaskAchieveOnConditionSet sub achieve nil or empty")
	}

	create_unix_day := timer.GetDayFrom1970WithCfgAndSec(0, this.db.Info.GetCreateUnix())
	cur_unix_day := timer.GetDayFrom1970WithCfg(0)

	var notify_seven *msg_client_message.S2CNotifySevenActValueChg
	sub_sevens := cfg_player_act_mgr.type2sevendayacts[con]
	if nil != sub_sevens && sub_sevens.Count > 0 {
		var tmp_sevencfg *XmlSevenDayItem
		for idx = 0; idx < sub_sevens.Count; idx++ {

			tmp_sevencfg = sub_sevens.Array[idx]
			if cfg_player_act_mgr.GetSevenDayLeftSec(create_unix_day, cur_unix_day, tmp_sevencfg.TasksID) <= 0 {
				log.Info("七天日活动[%d] 不在开启状态", tmp_sevencfg.TasksID)
				continue
			}

			cur_seven := this.db.SevenActivitys.Get(tmp_sevencfg.TasksID)
			if nil != cur_seven {
				if cur_seven.RewardUnix > 0 {
					log.Info("七天日活动[%d] 已经领奖", tmp_sevencfg.TasksID)
					continue
				}

				this.db.SevenActivitys.SetValue(tmp_sevencfg.TasksID, new_val)
				log.Info("七天活动[%d]已经有值[%d]", tmp_sevencfg.TasksID, new_val)
			} else {
				tmp_seven := &dbPlayerSevenActivityData{}
				tmp_seven.ActivityId = tmp_sevencfg.TasksID
				tmp_seven.Value = new_val
				this.db.SevenActivitys.Add(tmp_seven)
			}

			if !bslience {
				notify_seven = &msg_client_message.S2CNotifySevenActValueChg{}
				notify_seven.ActivityId = proto.Int32(tmp_sevencfg.TasksID)
				notify_seven.ActivityValue = proto.Int32(new_val)
				this.Send(notify_seven)
			}

			log.Info("=====七天乐触发[%d]触发add函数con[%d] new_val[%d]", tmp_sevencfg.TasksID, con, new_val)
		}
	} else {
		log.Error("Player TaskAchieveOnConditionSet sub seven nil or empty")
	}
}

// ============================================================================

func reg_player_task_achieve_msg() {
	hall_server.SetMessageHandler(msg_client_message.ID_C2SGetDialyTaskInfo, C2SGetDialyTaskInfoHanlder)
	hall_server.SetMessageHandler(msg_client_message.ID_C2SGetAchieve, C2SGetAchieveHandler)
	hall_server.SetMessageHandler(msg_client_message.ID_C2SGetTaskReward, C2SGetTaskRewardHandler)
	hall_server.SetMessageHandler(msg_client_message.ID_C2SGetAchieveReward, C2SGetAchieveRewardHandler)
	hall_server.SetMessageHandler(msg_client_message.ID_C2SGetSevenActReward, C2SGetSevenActRewardHandler)
}

func C2SGetDialyTaskInfoHanlder(c *socket.TcpConn, msg proto.Message) {
	req := msg.(*msg_client_message.C2SGetDialyTaskInfo)
	if nil == c || nil == req {
		log.Error("C2SGetDialyTaskInfoHanlder req nil [%v]", nil == req)
		return
	}

	p := player_mgr.GetPlayerById(int32(c.T))
	if nil == p {
		log.Error("C2SGetDialyTaskInfoHanlder not login !")
		return
	}

	p.SyncPlayerDialyTask()

	return
}

func C2SGetAchieveHandler(c *socket.TcpConn, msg proto.Message) {
	p := player_mgr.GetPlayerById(int32(c.T))
	if nil == p {
		log.Error("C2SGetAchieveHandler failed to find p ！")
		return
	}

	p.SyncPlayerAchieve()

	return
}

func C2SGetTaskRewardHandler(c *socket.TcpConn, msg proto.Message) {
	req := msg.(*msg_client_message.C2SGetTaskReward)
	if nil == c || nil == req {
		log.Error("C2SGetTaskRewardHandler c or req nil !", nil == req)
		return
	}

	p := player_mgr.GetPlayerById(int32(c.T))
	if nil == p {
		log.Error("C2SGetTaskRewardHandler not login !")
		return
	}

	task_id := req.GetTaskId()

	curreward_unix, _ := p.db.DialyTasks.GetRewardUnix(task_id)
	if curreward_unix > 0 {
		log.Error("C2SGetTaskRewardHandler already finished !")
		return
	}

	task_cfg := achieve_task_mgr.DialyTaskMap[task_id]
	if nil == task_cfg {
		log.Error("C2SGetTaskRewardHandler not find in cfg[%d]", task_id)
		return
	}

	cur_val, _ := p.db.DialyTasks.GetValue(task_id)
	if cur_val < task_cfg.FinishNeedCount {
		log.Error("C2SGetTaskRewardHandler not finished(%d < %d)", cur_val, task_cfg.FinishNeedCount)
		return
	}

	p.db.DialyTasks.SetRewardUnix(task_id, int32(time.Now().Unix()))

	cur_coin := p.AddCoin(task_cfg.RewardGold, "gettaskreward", "dialytask")
	cur_lvl, cur_exp := p.AddExp(task_cfg.RewardExp, "gettaskreward", "dialytask")

	res2cli := &msg_client_message.S2CRetTaskReward{}
	res2cli.Coin = proto.Int32(cur_coin)
	res2cli.CurLvl = proto.Int32(cur_lvl)
	res2cli.Exp = proto.Int32(cur_exp)
	res2cli.Diamond = proto.Int32(p.db.Info.GetDiamond())

	p.Send(res2cli)

	return
}

func C2SGetAchieveRewardHandler(c *socket.TcpConn, msg proto.Message) {
	req := msg.(*msg_client_message.C2SGetAchieveReward)
	if nil == c || nil == req {
		log.Error("C2SGetAchieveRewardHandler c or req nil !", nil == req)
		return
	}

	p := player_mgr.GetPlayerById(int32(c.T))
	if nil == p {
		log.Error("C2SGetAchieveRewardHandler not login !")
		return
	}

	achieve_id := req.GetAchieveReward()
	curreward_unix, _ := p.db.Achieves.GetRewardUnix(achieve_id)
	if curreward_unix > 0 {
		log.Error("C2SGetAchieveRewardHandler already finished !")
		return
	}

	achieve_cfg := achieve_task_mgr.AchieveMap[achieve_id]
	if nil == achieve_cfg {
		log.Error("C2SGetTaskRewardHandler not find in cfg[%d]", achieve_id)
		return
	}

	pre_reward_unix, pre_has := p.db.DialyTasks.GetRewardUnix(achieve_cfg.FrontTask)
	if pre_has && pre_reward_unix <= 0 {
		log.Error("C2SGetTaskRewardHandler pre task[achieve_cfg.FrontTask] not finished !")
		return
	}

	cur_val, _ := p.db.Achieves.GetValue(achieve_id)
	if cur_val < achieve_cfg.FinishNeedCount {
		log.Error("C2SGetTaskRewardHandler not finished(%d < %d)", cur_val, achieve_cfg.FinishNeedCount)
		return
	}

	p.db.Achieves.SetRewardUnix(achieve_id, int32(time.Now().Unix()))

	cur_coin := p.AddCoin(achieve_cfg.RewardGold, "gettaskreward", "dialytask")
	cur_lvl, cur_exp := p.AddExp(achieve_cfg.RewardExp, "gettaskreward", "dialytask")

	res2cli := &msg_client_message.S2CRetAchieveReward{}
	res2cli.TaskId = proto.Int32(achieve_id)
	res2cli.Coin = proto.Int32(cur_coin)
	res2cli.CurLvl = proto.Int32(cur_lvl)
	res2cli.Exp = proto.Int32(cur_exp)
	res2cli.Diamond = proto.Int32(p.db.Info.GetDiamond())

	p.Send(res2cli)

	return
}

func C2SGetSevenActRewardHandler(c *socket.TcpConn, msg proto.Message) {
	req := msg.(*msg_client_message.C2SGetSevenActReward)
	if nil == c || nil == req {
		log.Error("C2SGetSevenActRewardHandler c or req nil !", nil == req)
		return
	}

	p := player_mgr.GetPlayerById(int32(c.T))
	if nil == p {
		log.Error("C2SGetSevenActRewardHandler not login !")
		return
	}

	act_id := req.GetActivityId()

	seven_cfg := cfg_player_act_mgr.id2sevendayact[act_id]
	if nil == seven_cfg {
		log.Error("C2SGetSevenActRewardHandler not find in cfg[%d]", act_id)
		return
	}

	if cfg_player_act_mgr.GetSevenDayLeftSec(timer.GetDayFrom1970WithCfgAndSec(0, p.db.Info.GetCreateUnix()), timer.GetDayFrom1970WithCfg(int32(time.Now().Unix())), act_id) < 0 {
		log.Error("C2SGetSevenActRewardHandler act not open !")
		return
	}

	curreward_unix, _ := p.db.SevenActivitys.GetRewardUnix(act_id)
	if curreward_unix > 0 {
		log.Error("C2SGetSevenActRewardHandler already finished !")
		return
	}

	cur_val, _ := p.db.SevenActivitys.GetValue(act_id)
	if cur_val < seven_cfg.FinishNeedCount {
		log.Error("C2SGetSevenActRewardHandler not finished(%d < %d)", cur_val, seven_cfg.FinishNeedCount)
		return
	}

	p.db.SevenActivitys.SetRewardUnix(act_id, int32(time.Now().Unix()))

	res2cli := &msg_client_message.S2CRetSevenActReward{}
	res2cli.ActivityId = proto.Int32(act_id)
	res2cli.Rewards = p.OpenChest(seven_cfg.Reward, "get_act_reward", "seven_act", 0, -1, true)

	p.Send(res2cli)

	return
}*/
