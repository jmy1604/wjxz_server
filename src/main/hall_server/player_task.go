package main

import (
	"libs/log"
	"libs/timer"
	"net/http"
	"public_message/gen_go/client_message"
	"time"
	"youma/table_config"

	"3p/code.google.com.protobuf/proto"
)

// 任务状态
const (
	TASK_STATE_DOING    = 0 // 正在进行
	TASK_STATE_COMPLETE = 1 // 完成
	TASK_STATE_REWARD   = 2 // 已领奖
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
	res2cli := this.db.DialyTasks.FillDialyTaskMsg(this)
	if nil == res2cli || len(res2cli.TaskList) < 1 {
		return
	}
	this.Send(res2cli)
	return
}

func (this *Player) SyncPlayerAchieve() {
	res2cli := this.db.Achieves.FillAchieveMsg(this)
	if nil == res2cli || len(res2cli.AchieveList) < 1 {
		return
	}
	this.Send(res2cli)
	return
}

func (this *Player) IsPrevAchieveReward(task *table_config.XmlTaskItem) bool {
	if task.Prev <= 0 {
		return true
	}
	r, o := this.db.Achieves.GetRewardUnix(task.Prev)
	if !o || r <= 0 {
		return false
	}
	return true
}

func (this *Player) UpdateNewTasks(level int32, send_msg bool) int32 {
	tasks := task_table_mgr.GetLevelTasks(level)
	if tasks == nil {
		return 0
	}

	notify := &msg_client_message.S2CNotifyAchieveValueChg{}
	for _, task := range tasks {
		if this.db.Achieves.HasIndex(task.Id) {
			continue
		}

		if !this.db.FinishedAchieves.HasIndex(task.Id) && this.IsPrevAchieveReward(task) {
			var data dbPlayerAchieveData
			data.AchieveId = task.Id
			this.db.Achieves.Add(&data)

			if send_msg {
				this.NotifyAchieveValue(notify, data.AchieveId, data.Value, 0)
			}
		}
	}
	return 1
}

func (this *Player) check_add_next_task(task *table_config.XmlTaskItem, add_val int32) {
	if task.Next <= 0 {
		return
	}
	next_task := task_table_mgr.GetTask(task.Next)
	if next_task == nil {
		return
	}
	if this.db.Achieves.HasIndex(task.Next) {
		return
	}
	if next_task.MinLevel > this.db.Info.GetLvl() {
		return
	}

	update, cur_val, cur_state := this.SingleAchieveTaskUpdate(next_task, add_val)
	if update {
		notify := &msg_client_message.S2CNotifyAchieveValueChg{}
		this.NotifyAchieveValue(notify, task.Next, cur_val, cur_state)
	}
}

// ============================================================================

func (this *Player) NotifyTaskValue(notify_task *msg_client_message.S2CNotifyTaskValueChg, task_id, value, state int32) {
	notify_task.TaskId = proto.Int32(task_id)
	notify_task.TaskValue = proto.Int32(value)
	notify_task.TaskState = proto.Int32(state)
	this.Send(notify_task)
}

func (this *Player) NotifyAchieveValue(notify_achieve *msg_client_message.S2CNotifyAchieveValueChg, task_id, value, state int32) {
	notify_achieve.AchieveId = proto.Int32(task_id)
	notify_achieve.AchieveValue = proto.Int32(value)
	notify_achieve.AchieveState = proto.Int32(state)
	this.Send(notify_achieve)
}

// 前置任务是否已完成
func (this *Player) IsPrevTaskComplete(task *table_config.XmlTaskItem) bool {
	// 没有前置任务
	if task.Prev == 0 {
		return true
	}

	var prev_task *table_config.XmlTaskItem
	if task.Type == table_config.TASK_TYPE_DAILY {
		prev_task = task_table_mgr.GetTaskMap()[task.Prev]
		// 前置任务不存在
		if prev_task == nil {
			return true
		}
		prev_task_data := this.db.DialyTasks.Get(task.Prev)
		// 前置任务未开始
		if prev_task_data == nil {
			log.Debug("任务(%v)前置任务(%v)未开始", task.Id, prev_task.Id)
			return false
		}
		// 前置任务未完成
		if prev_task.CompleteNum != prev_task_data.Value {
			log.Debug("任务(%v)前置任务(%v)未完成", task.Id, prev_task.Id)
			return false
		}
	} else if task.Type == table_config.TASK_TYPE_ACHIEVEMENT {
		prev_task = task_table_mgr.GetTaskMap()[task.Prev]
		// 前置任务不存在
		if prev_task == nil {
			return true
		}
		prev_task_data := this.db.Achieves.Get(task.Prev)
		// 前置任务未开始
		if prev_task_data == nil {
			log.Debug("任务(%v)前置任务(%v)未开始", task.Id, prev_task.Id)
			return false
		}
		// 前置任务未完成
		if prev_task.CompleteNum != prev_task_data.Value {
			log.Debug("任务(%v)前置任务(%v)未完成", task.Id, prev_task.Id)
			return false
		}
	} else {
		return false
	}

	return true
}

func (this *Player) IsPrevTaskCompleteById(task_id int32) bool {
	task := task_table_mgr.GetTask(task_id)
	if task == nil {
		return false
	}
	return this.IsPrevTaskComplete(task)
}

// 任务是否完成
func (this *Player) IsTaskComplete(task *table_config.XmlTaskItem) bool {
	if task.Type == table_config.TASK_TYPE_DAILY {
		task_data := this.db.DialyTasks.Get(task.Id)
		if task_data == nil {
			return false
		}
		if task_data.Value < task.CompleteNum {
			return false
		}
	} else if task.Type == table_config.TASK_TYPE_ACHIEVEMENT {
		task_data := this.db.Achieves.Get(task.Id)
		if task_data == nil {
			return false
		}
		if task_data.Value < task.CompleteNum {
			return false
		}
	} else {
		return false
	}
	return true
}

func (this *Player) IsTaskCompleteById(task_id int32) bool {
	task := task_table_mgr.GetTaskMap()[task_id]
	if task == nil {
		return false
	}
	return this.IsTaskComplete(task)
}

// 单个日常任务更新
func (this *Player) SingleDailyTaskUpdate(tmp_taskcfg *table_config.XmlTaskItem, add_val int32) (updated bool, cur_val int32, cur_state int32) {
	cur_dialy := this.db.DialyTasks.Get(tmp_taskcfg.Id)
	if nil != cur_dialy {
		// 已领奖
		if cur_dialy.RewardUnix > 0 {
			return
		}
		if tmp_taskcfg.CompleteNum > cur_dialy.Value {
			/*diff := tmp_taskcfg.CompleteNum - cur_dialy.Value
			if add_val > diff {
				add_val = diff
			}*/
			cur_val = this.db.DialyTasks.IncbyValue(tmp_taskcfg.Id, add_val)
			updated = true
		}
	} else {
		new_dialy := &dbPlayerDialyTaskData{}
		new_dialy.TaskId = tmp_taskcfg.Id
		new_dialy.Value = add_val
		this.db.DialyTasks.Add(new_dialy)
		/*if tmp_taskcfg.CompleteNum < add_val {
			add_val = tmp_taskcfg.CompleteNum
		}*/
		cur_val = add_val
		updated = true
	}
	if cur_val >= tmp_taskcfg.CompleteNum {
		cur_state = TASK_STATE_COMPLETE
	} else {
		cur_state = TASK_STATE_DOING
	}
	return
}

// 单个成就任务更新
func (this *Player) SingleAchieveTaskUpdate(tmp_taskcfg *table_config.XmlTaskItem, add_val int32) (updated bool, cur_val int32, cur_state int32) {
	cur_achieve := this.db.Achieves.Get(tmp_taskcfg.Id)
	if nil != cur_achieve {
		if cur_achieve.RewardUnix > 0 {
			return
		}
		if tmp_taskcfg.CompleteNum > cur_achieve.Value {
			/*diff := tmp_taskcfg.CompleteNum - cur_achieve.Value
			if add_val > diff {
				add_val = diff
			}*/
			cur_val = this.db.Achieves.IncbyValue(tmp_taskcfg.Id, add_val)
			updated = true
		}
	} else {
		new_achieve := &dbPlayerAchieveData{}
		new_achieve.AchieveId = tmp_taskcfg.Id
		new_achieve.Value = add_val
		this.db.Achieves.Add(new_achieve)
		/*if tmp_taskcfg.CompleteNum < add_val {
			add_val = tmp_taskcfg.CompleteNum
		}*/
		cur_val = add_val
		updated = true
	}
	if cur_val >= tmp_taskcfg.CompleteNum {
		cur_state = TASK_STATE_COMPLETE
	} else {
		cur_state = TASK_STATE_DOING
	}
	return
}

// 完成所有日常任务更新
func (this *Player) WholeDailyTaskUpdate(daily_task *table_config.XmlTaskItem, notify_task *msg_client_message.S2CNotifyTaskValueChg) {
	if task_table_mgr.GetWholeDailyTask() == nil || this.IsTaskComplete(task_table_mgr.GetWholeDailyTask()) {
		return
	}

	if daily_task.EventId != table_config.TASK_FINISH_ALL_DAILY {
		task := this.db.DialyTasks.Get(daily_task.Id)
		if task == nil {
			return
		}
		to_send, cur_val, cur_state := this.SingleDailyTaskUpdate(task_table_mgr.GetWholeDailyTask(), 1)
		if to_send {
			this.NotifyTaskValue(notify_task, task_table_mgr.GetWholeDailyTask().Id, cur_val, cur_state)
			log.Info("Player(%v) WholeDailyTask(%v) Update, Progress(%v/%v), Complete(%v)", this.Id, task_table_mgr.GetWholeDailyTask().Id, cur_val, task_table_mgr.GetWholeDailyTask().CompleteNum, cur_state)
		}
	}
}

// 任务更新
func (this *Player) TaskUpdate(finish_type int32, if_not_less bool, event_param int32, add_val int32) {
	log.Info("进入任务成就触发add函数finish_type[%d] event_param[%v] add_val[%d]", finish_type, event_param, add_val)
	var idx int32
	var cur_val, cur_state int32

	notify_task := &msg_client_message.S2CNotifyTaskValueChg{}
	notify_achieve := &msg_client_message.S2CNotifyAchieveValueChg{}
	ftasks := task_table_mgr.GetFinishTasks()[finish_type]
	if nil != ftasks && ftasks.GetCount() > 0 {
		var tmp_taskcfg *table_config.XmlTaskItem
		for idx = 0; idx < ftasks.GetCount(); idx++ {
			tmp_taskcfg = ftasks.GetArray()[idx]
			if tmp_taskcfg.Type == table_config.TASK_TYPE_ACHIEVEMENT {
				/*if !this.db.Achieves.HasIndex(tmp_taskcfg.Id) {
					continue
				}*/
			} else if tmp_taskcfg.Type == table_config.TASK_TYPE_DAILY {
				/*if !this.db.DialyTasks.HasIndex(tmp_taskcfg.Id) {
					continue
				}*/
			} else {
				continue
			}
			// 已完成
			if this.IsTaskComplete(tmp_taskcfg) {
				continue
			}

			// 前置任务未完成
			if !this.IsPrevTaskComplete(tmp_taskcfg) {
				continue
			}

			// 等级不满足
			if tmp_taskcfg.MinLevel > this.db.Info.GetLvl() || tmp_taskcfg.MaxLevel < this.db.Info.GetLvl() {
				continue
			}

			// 事件参数
			if if_not_less {
				if event_param < tmp_taskcfg.EventParam {
					continue
				}
			} else {
				// 参数不一致
				if event_param != tmp_taskcfg.EventParam {
					continue
				}
			}

			var updated bool
			if tmp_taskcfg.Type == table_config.TASK_TYPE_DAILY {
				updated, cur_val, cur_state = this.SingleDailyTaskUpdate(tmp_taskcfg, add_val)
				// 所有日常任务更新
				if cur_state == TASK_STATE_COMPLETE {
					this.WholeDailyTaskUpdate(tmp_taskcfg, notify_task)
				}
			} else if tmp_taskcfg.Type == table_config.TASK_TYPE_ACHIEVEMENT {
				updated, cur_val, cur_state = this.SingleAchieveTaskUpdate(tmp_taskcfg, add_val)
			} else {
				log.Error("not supported task type %v by id %v", tmp_taskcfg.Type, tmp_taskcfg.Id)
				continue
			}

			if updated {
				if tmp_taskcfg.Type == table_config.TASK_TYPE_DAILY {
					this.NotifyTaskValue(notify_task, tmp_taskcfg.Id, cur_val, cur_state)
				} else if tmp_taskcfg.Type == table_config.TASK_TYPE_ACHIEVEMENT {
					this.NotifyAchieveValue(notify_achieve, tmp_taskcfg.Id, cur_val, cur_state)
				}
				log.Info("Player[%v] Task[%v] EventParam[%v] Progress[%v/%v] FinishType(%v) Complete(%v)", this.Id, tmp_taskcfg.Id, event_param, cur_val, tmp_taskcfg.CompleteNum, finish_type, cur_state)
			}
		}
	} else {
		log.Error("Player TaskAchieveOnConditionAdd sub dialy nil or empty [%v]", nil == ftasks)
	}
}

func (this *Player) get_daily_task_info() int32 {
	this.SyncPlayerDialyTask()
	return 1
}

func (this *Player) get_achieve_info() int32 {
	this.SyncPlayerAchieve()
	return 1
}

func (p *Player) get_daily_reward(task_id int32) int32 {
	curreward_unix, _ := p.db.DialyTasks.GetRewardUnix(task_id)
	if curreward_unix > 0 {
		log.Error("C2SGetTaskRewardHandler already finished !")
		return int32(msg_client_message.E_ERR_TASK_ALREADY_REWARDED)
	}

	task_cfg := task_table_mgr.GetTaskMap()[task_id]
	if nil == task_cfg {
		log.Error("C2SGetTaskRewardHandler not find in cfg[%d]", task_id)
		return int32(msg_client_message.E_ERR_TASK_NOT_FOUND)
	}

	plvl := p.db.Info.GetLvl()
	if plvl < task_cfg.MinLevel || plvl > task_cfg.MaxLevel {
		log.Error("player level %v is not range for %v-%v", plvl, task_cfg.MinLevel, task_cfg.MaxLevel)
		return int32(msg_client_message.E_ERR_TASK_LEVEL_NOT_ENOUGH)
	}

	cur_val, _ := p.db.DialyTasks.GetValue(task_id)
	if cur_val < task_cfg.CompleteNum {
		log.Error("C2SGetTaskRewardHandler not finished(%d < %d)", cur_val, task_cfg.CompleteNum)
		return int32(msg_client_message.E_ERR_TASK_NOT_COMPLETE)
	}

	p.db.DialyTasks.SetRewardUnix(task_id, int32(time.Now().Unix()))
	notify_task := &msg_client_message.S2CNotifyTaskValueChg{}
	p.NotifyTaskValue(notify_task, task_id, cur_val, TASK_STATE_REWARD)

	for i := 0; i < len(task_cfg.Rewards); i++ {
		p.AddItemResource(task_cfg.Rewards[i].ItemId, task_cfg.Rewards[i].Num, "gettaskreward", "dailytask")
	}
	cur_lvl, cur_exp := p.AddExp(task_cfg.Exp, "gettaskreward", "dialytask")

	p.SendItemsUpdate()
	p.SendCatsUpdate()
	p.SendDepotBuildingUpdate()

	res2cli := &msg_client_message.S2CRetTaskReward{}
	res2cli.Coin = proto.Int32(p.GetCoin())
	res2cli.CurLvl = proto.Int32(cur_lvl)
	res2cli.Exp = proto.Int32(cur_exp)
	res2cli.Diamond = proto.Int32(p.db.Info.GetDiamond())

	p.Send(res2cli)

	return 1
}

func (p *Player) get_achieve_reward(achieve_id int32) int32 {
	curreward_unix, _ := p.db.Achieves.GetRewardUnix(achieve_id)
	if curreward_unix > 0 {
		log.Error("C2SGetAchieveRewardHandler already finished !")
		return int32(msg_client_message.E_ERR_TASK_ALREADY_REWARDED)
	}

	achieve_cfg := task_table_mgr.GetTaskMap()[achieve_id]
	if nil == achieve_cfg {
		log.Error("C2SGetTaskRewardHandler not find in cfg[%d]", achieve_id)
		return int32(msg_client_message.E_ERR_TASK_NOT_FOUND)
	}

	plvl := p.db.Info.GetLvl()
	if plvl < achieve_cfg.MinLevel || plvl > achieve_cfg.MaxLevel {
		log.Error("player level %v is not range for %v-%v", plvl, achieve_cfg.MinLevel, achieve_cfg.MaxLevel)
		return int32(msg_client_message.E_ERR_TASK_LEVEL_NOT_ENOUGH)
	}

	pre_reward_unix, pre_has := p.db.DialyTasks.GetRewardUnix(achieve_cfg.Prev)
	if pre_has && pre_reward_unix <= 0 {
		log.Error("C2SGetTaskRewardHandler pre task[achieve_cfg.Prev] not finished !")
		return int32(msg_client_message.E_ERR_TASK_PREV_NOT_COMPLETE)
	}

	cur_val, _ := p.db.Achieves.GetValue(achieve_id)
	if cur_val < achieve_cfg.CompleteNum {
		log.Error("C2SGetTaskRewardHandler not finished(%d < %d)", cur_val, achieve_cfg.CompleteNum)
		return int32(msg_client_message.E_ERR_TASK_NOT_COMPLETE)
	}

	p.db.Achieves.SetRewardUnix(achieve_id, int32(time.Now().Unix()))
	notify_achieve := &msg_client_message.S2CNotifyAchieveValueChg{}
	p.NotifyAchieveValue(notify_achieve, achieve_id, cur_val, TASK_STATE_REWARD)

	for i := 0; i < len(achieve_cfg.Rewards); i++ {
		p.AddItemResource(achieve_cfg.Rewards[i].ItemId, achieve_cfg.Rewards[i].Num, "gettaskreward", "dailytask")
	}
	cur_lvl, cur_exp := p.AddExp(achieve_cfg.Exp, "gettaskreward", "dialytask")

	p.SendItemsUpdate()
	p.SendCatsUpdate()
	p.SendDepotBuildingUpdate()

	res2cli := &msg_client_message.S2CRetAchieveReward{}
	res2cli.TaskId = proto.Int32(achieve_id)
	res2cli.Coin = proto.Int32(p.GetCoin())
	res2cli.CurLvl = proto.Int32(cur_lvl)
	res2cli.Exp = proto.Int32(cur_exp)
	res2cli.Diamond = proto.Int32(p.db.Info.GetDiamond())

	p.Send(res2cli)

	p.db.Achieves.Remove(achieve_id)
	var data dbPlayerFinishedAchieveData
	data.AchieveId = achieve_id
	p.db.FinishedAchieves.Add(&data)

	// 后置任务
	p.check_add_next_task(achieve_cfg, 0)

	return 1
}

func (this *Player) complete_task(task_id int32) int32 {
	task := task_table_mgr.GetTask(task_id)
	if task == nil {
		log.Error("Task[%v] table data not found", task_id)
		return -1
	}
	if task.Type == table_config.TASK_TYPE_DAILY {
		task_data := this.db.DialyTasks.Get(task_id)
		if task_data == nil {
			var data dbPlayerDialyTaskData
			data.TaskId = task_id
			data.Value = task.CompleteNum
			this.db.DialyTasks.Add(&data)
		} else {
			this.db.DialyTasks.SetValue(task_id, task.CompleteNum)
		}
	} else if task.Type == table_config.TASK_TYPE_ACHIEVEMENT {
		task_data := this.db.Achieves.Get(task_id)
		if task_data == nil {
			var data dbPlayerAchieveData
			data.AchieveId = task_id
			data.Value = task.CompleteNum
			this.db.Achieves.Add(&data)
		} else {
			this.db.Achieves.SetValue(task_id, task.CompleteNum)
		}
	} else {
		log.Error("task type[%v] unknown", task.Type)
		return -1
	}

	return 0
}

// ============================================================================

func C2SGetDialyTaskInfoHanlder(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SGetDialyTaskInfo)
	if nil == p || nil == req {
		log.Error("C2SGetDialyTaskInfoHanlder req nil [%v]", nil == req)
		return -1
	}

	p.SyncPlayerDialyTask()

	return 1
}

func C2SGetAchieveHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SGetAchieve)
	if nil == req {
		log.Error("C2SGetAchieveHandler req nil [%v]", nil == req)
		return -1
	}

	p.SyncPlayerAchieve()

	return 1
}

func C2SGetTaskRewardHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SGetTaskReward)
	if nil == req {
		log.Error("C2SGetTaskRewardHandler req nil !", nil == req)
		return -1
	}

	task_id := req.GetTaskId()
	return p.get_daily_reward(task_id)
}

func C2SGetAchieveRewardHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SGetAchieveReward)
	if nil == req {
		log.Error("C2SGetAchieveRewardHandler req nil !", nil == req)
		return -1
	}
	achieve_id := req.GetAchieveReward()
	return p.get_achieve_reward(achieve_id)
}
