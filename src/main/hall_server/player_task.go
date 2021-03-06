package main

import (
	"libs/log"
	"libs/utils"
	"main/table_config"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	"time"

	"github.com/golang/protobuf/proto"
)

// 任务状态
const (
	TASK_STATE_DOING    = 0 // 正在进行
	TASK_STATE_COMPLETE = 1 // 完成
	TASK_STATE_REWARD   = 2 // 已领奖
)

func (this *dbPlayerTaskColumn) ResetDailyTask() {
	this.m_row.m_lock.UnSafeLock("dbPlayerTaskColumn.ChkResetDailyTask")
	defer this.m_row.m_lock.UnSafeUnlock()

	daily_tasks := task_table_mgr.GetDailyTasks()
	if daily_tasks == nil {
		return
	}

	for id, task := range daily_tasks {
		d := this.m_data[id]
		if d == nil {
			data := &dbPlayerTaskData{}
			data.Id = task.Id
			data.Value = 0
			this.m_data[id] = data
		} else {
			d.Value = 0
			d.State = 0
		}
	}

	this.m_changed = true

	return
}

func (this *dbPlayerTaskColumn) FillTaskMsg(p *Player, task_type int32) *msg_client_message.S2CTaskDataResponse {
	var tmp_item *msg_client_message.TaskData
	this.m_row.m_lock.UnSafeRLock("dbPlayerTaskColumn.FillDialyTaskMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()
	ret_msg := &msg_client_message.S2CTaskDataResponse{}
	ret_msg.TaskType = task_type
	ret_msg.TaskList = make([]*msg_client_message.TaskData, 0, len(this.m_data))
	for _, val := range this.m_data {
		if nil == val {
			continue
		}

		task := task_table_mgr.GetTask(val.Id)
		if task == nil {
			continue
		}

		if task.Type != task_type {
			continue
		}

		tmp_item = &msg_client_message.TaskData{}
		tmp_item.Id = val.Id
		tmp_item.Value = val.Value
		tmp_item.State = val.State
		if tmp_item.Value >= task.CompleteNum && tmp_item.State == TASK_STATE_DOING {
			tmp_item.State = TASK_STATE_COMPLETE
		}
		ret_msg.TaskList = append(ret_msg.TaskList, tmp_item)
	}

	return ret_msg
}

func (this *Player) ChkPlayerDailyTask() int32 {
	remain_seconds := utils.GetRemainSeconds2NextDayTime(this.db.TaskCommon.GetLastRefreshTime(), global_config.DailyTaskRefreshTime)
	if remain_seconds <= 0 {
		this.db.Tasks.ResetDailyTask()
		now_time := int32(time.Now().Unix())
		this.db.TaskCommon.SetLastRefreshTime(now_time)
		remain_seconds = utils.GetRemainSeconds2NextDayTime(now_time, global_config.DailyTaskRefreshTime)
	}
	return remain_seconds
}

func (this *Player) first_gen_achieve_tasks() {
	if this.db.Tasks.NumAll() > 0 {
		this.db.Tasks.Clear()
	}
	achieves := task_table_mgr.GetStartAchieveTasks()
	if achieves != nil {
		for i := 0; i < len(achieves); i++ {
			this.db.Tasks.Add(&dbPlayerTaskData{
				Id: achieves[i].Id,
			})
		}
	}
}

func (this *Player) send_task(task_type int32) int32 {
	if task_type == 0 || task_type == table_config.TASK_TYPE_DAILY {
		remain_seconds := this.ChkPlayerDailyTask()
		response := this.db.Tasks.FillTaskMsg(this, table_config.TASK_TYPE_DAILY)
		response.DailyTaskRefreshRemainSeconds = remain_seconds
		this.Send(uint16(msg_client_message_id.MSGID_S2C_TASK_DATA_RESPONSE), response)
		log.Debug("Player[%v] daily tasks %v", this.Id, response)
	}

	if task_type == 0 || task_type == table_config.TASK_TYPE_ACHIVE {
		response := this.db.Tasks.FillTaskMsg(this, table_config.TASK_TYPE_ACHIVE)
		this.Send(uint16(msg_client_message_id.MSGID_S2C_TASK_DATA_RESPONSE), response)
		log.Debug("Player[%v] achive tasks %v", this.Id, response)
	}

	return 1
}

func (this *Player) IsPrevAchieveReward(task *table_config.XmlTaskItem) bool {
	if task.Prev <= 0 {
		return true
	}
	r, o := this.db.Tasks.GetState(task.Prev)
	if !o || r != TASK_STATE_REWARD {
		return false
	}
	return true
}

func (this *Player) check_add_next_task(task *table_config.XmlTaskItem, add_val int32) {
	if task.Next <= 0 {
		return
	}
	next_task := task_table_mgr.GetTask(task.Next)
	if next_task == nil {
		return
	}
	if this.db.Tasks.HasIndex(task.Next) {
		return
	}

	if next_task.EventId != task.EventId || task.EventId == table_config.TASK_COMPLETE_TYPE_PASS_CAMPAIGN {
		add_val = 0
	}

	update, cur_val, cur_state := this.SingleTaskUpdate(next_task, add_val)
	if update {
		notify := &msg_client_message.S2CTaskValueNotify{}
		this.NotifyTaskValue(notify, task.Next, cur_val, cur_state)
		log.Debug("Player[%v] add new task %v, reuse add_val %v", this.Id, task.Next, add_val)
	}
}

// ============================================================================

func (this *Player) NotifyTaskValue(notify_task *msg_client_message.S2CTaskValueNotify, task_id, value, state int32) {
	notify_task.Data = &msg_client_message.TaskData{}
	notify_task.Data.Id = task_id
	notify_task.Data.Value = value
	notify_task.Data.State = state
	this.Send(uint16(msg_client_message_id.MSGID_S2C_TASK_VALUE_NOTIFY), notify_task)
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
		prev_task_data := this.db.Tasks.Get(task.Prev)
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
	} else if task.Type == table_config.TASK_TYPE_ACHIVE {
		prev_task = task_table_mgr.GetTaskMap()[task.Prev]
		// 前置任务不存在
		if prev_task == nil {
			return true
		}
		prev_task_data := this.db.Tasks.Get(task.Prev)
		// 前置任务未开始
		if prev_task_data == nil {
			//log.Debug("任务(%v)前置任务(%v)未开始", task.Id, prev_task.Id)
			return false
		}
		// 前置任务未完成
		if prev_task.CompleteNum != prev_task_data.Value {
			//log.Debug("任务(%v)前置任务(%v)未完成", task.Id, prev_task.Id)
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
		task_data := this.db.Tasks.Get(task.Id)
		if task_data == nil {
			return false
		}
		if task_data.Value < task.CompleteNum {
			return false
		}
	} else if task.Type == table_config.TASK_TYPE_ACHIVE {
		task_data := this.db.Tasks.Get(task.Id)
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
func (this *Player) SingleTaskUpdate(task *table_config.XmlTaskItem, add_val int32) (updated bool, cur_val int32, cur_state int32) {
	if this.db.Tasks.HasIndex(task.Id) {
		// 已领奖
		state, _ := this.db.Tasks.GetState(task.Id)
		if state == TASK_STATE_REWARD {
			return
		}

		value, _ := this.db.Tasks.GetValue(task.Id)
		if task.CompleteNum > value {
			cur_val = this.db.Tasks.IncbyValue(task.Id, add_val)
			updated = true
		}
	} else {
		this.db.Tasks.Add(&dbPlayerTaskData{
			Id:    task.Id,
			Value: add_val,
		})
		cur_val = add_val
		updated = true
	}

	if cur_val >= task.CompleteNum {
		cur_state = TASK_STATE_COMPLETE
		this.db.Tasks.SetState(task.Id, TASK_STATE_COMPLETE)
	} else {
		cur_state = TASK_STATE_DOING
	}
	return
}

// 完成所有日常任务更新
func (this *Player) WholeDailyTaskUpdate(daily_task *table_config.XmlTaskItem, notify_task *msg_client_message.S2CTaskValueNotify) {
	if task_table_mgr.GetWholeDailyTask() == nil || this.IsTaskComplete(task_table_mgr.GetWholeDailyTask()) {
		return
	}

	if daily_task.EventId != table_config.TASK_COMPLETE_TYPE_ALL_DAILY {
		task := this.db.Tasks.Get(daily_task.Id)
		if task == nil {
			return
		}
		to_send, cur_val, cur_state := this.SingleTaskUpdate(task_table_mgr.GetWholeDailyTask(), 1)
		if to_send {
			this.NotifyTaskValue(notify_task, task_table_mgr.GetWholeDailyTask().Id, cur_val, cur_state)
			log.Info("Player(%v) WholeDailyTask(%v) Update, Progress(%v/%v), Complete(%v)", this.Id, task_table_mgr.GetWholeDailyTask().Id, cur_val, task_table_mgr.GetWholeDailyTask().CompleteNum, cur_state)
		}
	}
}

// 任务更新
func (this *Player) TaskUpdate(complete_type int32, if_not_less bool, event_param int32, value int32) {
	//log.Info("complete_type[%d] event_param[%v] aval[%d]", complete_type, event_param, value)
	var idx int32
	var cur_val, cur_state int32

	notify_task := &msg_client_message.S2CTaskValueNotify{}
	ftasks := task_table_mgr.GetFinishTasks()[complete_type]
	if nil == ftasks || ftasks.GetCount() == 0 {
		log.Error("Task complete type %v no corresponding tasks", complete_type)
		return
	}

	var tmp_taskcfg *table_config.XmlTaskItem
	for idx = 0; idx < ftasks.GetCount(); idx++ {
		tmp_taskcfg = ftasks.GetArray()[idx]

		if !this.db.Tasks.HasIndex(tmp_taskcfg.Id) {
			continue
		}

		// 已完成
		if this.IsTaskComplete(tmp_taskcfg) {
			continue
		}

		// 事件参数
		if tmp_taskcfg.EventParam > 0 {
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
		}

		var updated bool
		if tmp_taskcfg.Type == table_config.TASK_TYPE_DAILY && cur_state == TASK_STATE_COMPLETE {
			// 所有日常任务更新
			this.WholeDailyTaskUpdate(tmp_taskcfg, notify_task)
		} else {
			updated, cur_val, cur_state = this.SingleTaskUpdate(tmp_taskcfg, value)
		}

		if updated {
			this.NotifyTaskValue(notify_task, tmp_taskcfg.Id, cur_val, cur_state)
			log.Info("Player[%v] Task[%v] EventParam[%v] Progress[%v/%v] FinishType(%v) Complete(%v)", this.Id, tmp_taskcfg.Id, event_param, cur_val, tmp_taskcfg.CompleteNum, complete_type, cur_state)
		}
	}
}

func (p *Player) task_get_reward(task_id int32) int32 {
	state, _ := p.db.Tasks.GetState(task_id)
	if state != TASK_STATE_COMPLETE {
		log.Error("Player[%v] task %v state %v cant reward", p.Id, task_id, state)
		if state == TASK_STATE_DOING {
			return int32(msg_client_message.E_ERR_PLAYER_TASK_NOT_COMPLETE)
		} else if state == TASK_STATE_REWARD {
			return int32(msg_client_message.E_ERR_PLAYER_TASK_ALREADY_REWARDED)
		}
	}

	task_cfg := task_table_mgr.GetTaskMap()[task_id]
	if nil == task_cfg {
		log.Error("task %v table data not found", task_id)
		return int32(msg_client_message.E_ERR_PLAYER_TASK_NOT_FOUND)
	}

	/*plvl := p.db.Info.GetLvl()
	if plvl < task_cfg.MinLevel || plvl > task_cfg.MaxLevel {
		log.Error("player level %v is not range for %v-%v", plvl, task_cfg.MinLevel, task_cfg.MaxLevel)
		return int32(msg_client_message.E_ERR_TASK_LEVEL_NOT_ENOUGH)
	}*/

	cur_val, _ := p.db.Tasks.GetValue(task_id)
	if cur_val < task_cfg.CompleteNum {
		log.Error("Player[%v] task %v not finished(%d < %d)", p.Id, task_id, cur_val, task_cfg.CompleteNum)
		return int32(msg_client_message.E_ERR_PLAYER_TASK_NOT_COMPLETE)
	}

	p.db.Tasks.SetState(task_id, TASK_STATE_REWARD)
	notify_task := &msg_client_message.S2CTaskValueNotify{}
	p.NotifyTaskValue(notify_task, task_id, cur_val, TASK_STATE_REWARD)

	response := &msg_client_message.S2CTaskRewardResponse{
		TaskId: task_id,
	}
	p.Send(uint16(msg_client_message_id.MSGID_S2C_TASK_REWARD_RESPONSE), response)

	if task_cfg.Type == table_config.TASK_TYPE_ACHIVE {
		if task_cfg.Next > 0 {
			p.db.Tasks.Remove(task_id)
			var data dbPlayerFinishedTaskData
			data.Id = task_id
			p.db.FinishedTasks.Add(&data)

			// 后置任务
			p.check_add_next_task(task_cfg, cur_val)
		}
	}

	return 1
}

func (this *Player) complete_task(task_id int32) int32 {
	task := task_table_mgr.GetTask(task_id)
	if task == nil {
		log.Error("Task[%v] table data not found", task_id)
		return -1
	}

	task_data := this.db.Tasks.Get(task_id)
	if task_data == nil {
		var data dbPlayerTaskData
		data.Id = task_id
		data.Value = task.CompleteNum
		data.State = TASK_STATE_COMPLETE
		this.db.Tasks.Add(&data)
	} else {
		this.db.Tasks.SetValue(task_id, task.CompleteNum)
	}

	return 0
}

// ============================================================================

func C2STaskDataHanlder(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2STaskDataRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)", err.Error())
		return -1
	}

	return p.send_task(req.GetTaskType())
}

func C2SGetTaskRewardHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2STaskRewardRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)", err.Error())
		return -1
	}

	return p.task_get_reward(req.GetTaskId())
}
