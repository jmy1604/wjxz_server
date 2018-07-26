package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

// 任务类型
const (
	TASK_TYPE_DAILY  = 1 // 日常
	TASK_TYPE_ACHIVE = 2 // 成就
)

// 任务完成类型
const (
	TASK_COMPLETE_TYPE_ALL_DAILY              = 101 // 完成所有日常任务
	TASK_COMPLETE_TYPE_GOLD_HAND_NUM          = 102 // 完成N次点金
	TASK_COMPLETE_TYPE_GIVE_POINTS_NUM        = 103 // 赠送N颗爱心
	TASK_COMPLETE_TYPE_EXPLORE_NUM            = 104 // 完成N次日常探索
	TASK_COMPLETE_TYPE_FORGE_EQUIP_NUM        = 105 // 合成N件装备
	TASK_COMPLETE_TYPE_NORMAL_DRAW_NUM        = 106 // 完成N次普通召唤
	TASK_COMPLETE_TYPE_ADVANCED_DRAW_NUM      = 107 // 完成N次高级召唤
	TASK_COMPLETE_TYPE_ARENA_FIGHT_NUM        = 108 // 完成N次竞技场战斗
	TASK_COMPLETE_TYPE_HUANG_UP_NUM           = 109 // 完成N次挂机收获
	TASK_COMPLETE_TYPE_ACTIVE_STAGE_WIN_NUM   = 110 // 获得N次活动副本胜利
	TASK_COMPLETE_TYPE_BUY_ITEM_NUM_ON_SHOP   = 111 // 神秘商店购买N件商品
	TASK_COMPLETE_TYPE_REACH_LEVEL            = 201 // 等级提升到N级
	TASK_COMPLETE_TYPE_GET_FOUR_STAR_ROLES    = 202 // 获得N个四星英雄
	TASK_COMPLETE_TYPE_GET_FIVE_STAR_ROLES    = 203 // 获得N个五星英雄
	TASK_COMPLETE_TYPE_GET_SIX_STAR_ROLES     = 204 // 获得N个六星英雄
	TASK_COMPLETE_TYPE_GET_NINE_STAR_ROLES    = 205 // 获得N个九星英雄
	TASK_COMPLETE_TYPE_GET_TEN_STAR_ROLES     = 206 // 获得N个十星英雄
	TASK_COMPLETE_TYPE_DECOMPOSE_ROLES        = 207 // 分解N个英雄
	TASK_COMPLETE_TYPE_ARENA_WIN_NUM          = 208 // 冠军争夺战进攻获胜N场次
	TASK_COMPLETE_TYPE_ARENA_REACH_SCORE      = 209 // 冠军争夺战积分达到N分
	TASK_COMPLETE_TYPE_PASS_CHAPTERS          = 210 // 通过N个章节
	TASK_COMPLETE_TYPE_PASS_FOUR_STAR_EXPLORE = 211 // 完成N个四星探索任务
	TASK_COMPLETE_TYPE_PASS_FIVE_STAR_EXPLORE = 212 // 完成N个五星探索任务
	TASK_COMPLETE_TYPE_PASS_SIX_STAR_EXPLORE  = 213 // 完成N个六星探索任务
	TASK_COMPLETE_TYPE_REACH_VIP_N_LEVEL      = 214 // 成为VIPN
	TASK_COMPLETE_TYPE_GET_GOLD_EQUIPS_NUM    = 215 // 获得N件金色装备
	TASK_COMPLETE_TYPE_GET_RED_EQUIPS_NUM     = 216 // 获得N件红色装备
	TASK_COMPLETE_TYPE_GET_ORANGE_EQUIPS_NUM  = 217 // 获得N件橙色装备
	TASK_COMPLETE_TYPE_SHARE_GAME_NUM         = 218 // 完成N次分享游戏
)

type TaskReward struct {
	ItemId int32
	Num    int32
}

type XmlTaskItem struct {
	Id   int32 `xml:"Id,attr"`
	Type int32 `xml:"Type,attr"`
	//MinLevel    int32  `xml:"MinLevel,attr"`
	//MaxLevel    int32  `xml:"Maxlevel,attr"`
	EventId     int32  `xml:"EventId,attr"`
	EventParam  int32  `xml:"EventParam,attr"`
	CompleteNum int32  `xml:"CompleteNum,attr"`
	Prev        int32  `xml:"Prev,attr"`
	Next        int32  `xml:"Next,attr"`
	RewardStr   string `xml:"Reward,attr"`
	Rewards     []int32
}

type XmlTaskTable struct {
	Items []XmlTaskItem `xml:"item"`
}

type FinishTypeTasks struct {
	count int32
	array []*XmlTaskItem
}

func (this *FinishTypeTasks) GetCount() int32 {
	return this.count
}

func (this *FinishTypeTasks) GetArray() []*XmlTaskItem {
	return this.array
}

type TaskTableMgr struct {
	task_map            map[int32]*XmlTaskItem     // 任务map
	task_array          []*XmlTaskItem             // 任务数组
	task_array_len      int32                      // 数组长度
	finish_tasks        map[int32]*FinishTypeTasks // 按完成条件组织任务数据
	daily_task_map      map[int32]*XmlTaskItem     // 日常任务MAP
	daily_task_array    []*XmlTaskItem             // 日程任务数组
	all_daily_task      *XmlTaskItem               // 所有日常任务
	start_achieve_tasks []*XmlTaskItem             // 初始成就任务
	//level_tasks      map[int32][]*XmlTaskItem   // 等级对应的任务
}

func (this *TaskTableMgr) Init() bool {
	if !this.LoadTask() {
		return false
	}
	return true
}

func (this *TaskTableMgr) LoadTask() bool {
	content, err := ioutil.ReadFile("../game_data/Mission.xml")
	if nil != err {
		log.Error("TaskTableMgr LoadTask read file error !")
		return false
	}

	tmp_cfg := &XmlTaskTable{}
	err = xml.Unmarshal(content, tmp_cfg)
	if nil != err {
		log.Error("TaskTableMgr LoadTask unmarshal failed(%s)", err.Error())
		return false
	}

	tmp_len := int32(len(tmp_cfg.Items))

	this.task_array = make([]*XmlTaskItem, 0, tmp_len)
	this.task_map = make(map[int32]*XmlTaskItem)
	this.finish_tasks = make(map[int32]*FinishTypeTasks)
	this.daily_task_map = make(map[int32]*XmlTaskItem)
	//this.level_tasks = make(map[int32][]*XmlTaskItem)

	var tmp_item *XmlTaskItem
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]

		rewards := parse_xml_str_arr2(tmp_item.RewardStr, ",")
		if rewards == nil || len(rewards)%2 != 0 {
			log.Error("@@@@@@ Task[%v] Reward[%v] invalid", tmp_item.Id, tmp_item.RewardStr)
			return false
		}

		tmp_item.Rewards = rewards

		this.task_map[tmp_item.Id] = tmp_item
		this.task_array = append(this.task_array, tmp_item)
		if nil == this.finish_tasks[tmp_item.EventId] {
			this.finish_tasks[tmp_item.EventId] = &FinishTypeTasks{}
		}
		this.finish_tasks[tmp_item.EventId].count++
		if tmp_item.Type == TASK_TYPE_DAILY {
			this.daily_task_map[tmp_item.Id] = tmp_item
			this.daily_task_array = append(this.daily_task_array, tmp_item)
			if tmp_item.EventId == TASK_COMPLETE_TYPE_ALL_DAILY {
				this.all_daily_task = tmp_item
			}
		} else {
			if tmp_item.Prev == 0 {
				this.start_achieve_tasks = append(this.start_achieve_tasks, tmp_item)
			}
		}

		/*if tmp_item.Type != TASK_TYPE_DAILY {
			if this.level_tasks[tmp_item.MinLevel] == nil {
				this.level_tasks[tmp_item.MinLevel] = make([]*XmlTaskItem, 0)
			}
			this.level_tasks[tmp_item.MinLevel] = append(this.level_tasks[tmp_item.MinLevel], tmp_item)
		}*/
	}
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]
		if nil == this.finish_tasks[tmp_item.EventId].array {
			this.finish_tasks[tmp_item.EventId].array = make([]*XmlTaskItem, 0, this.finish_tasks[tmp_item.EventId].count)
		}
		this.finish_tasks[tmp_item.EventId].array = append(this.finish_tasks[tmp_item.EventId].array, tmp_item)
		//log.Info("finish type(%v) tasks count %v", tmp_item.EventId, this.finish_tasks[tmp_item.EventId].count)
	}

	this.task_array_len = int32(len(this.task_array))

	// 所有日常任务CompleteNum处理
	if this.all_daily_task != nil {
		for _, d := range this.daily_task_map {
			if d.EventId != TASK_COMPLETE_TYPE_ALL_DAILY {
				this.all_daily_task.CompleteNum += 1
			}
		}
	}

	log.Info("TaskTableMgr Loaded Task table, daily tasks %v", this.daily_task_map)

	return true
}

func (this *TaskTableMgr) GetTaskMap() map[int32]*XmlTaskItem {
	return this.task_map
}

func (this *TaskTableMgr) GetTask(task_id int32) *XmlTaskItem {
	if this.task_map == nil {
		return nil
	}
	return this.task_map[task_id]
}

func (this *TaskTableMgr) GetWholeDailyTask() *XmlTaskItem {
	return this.all_daily_task
}

func (this *TaskTableMgr) GetFinishTasks() map[int32]*FinishTypeTasks {
	return this.finish_tasks
}

/*func (this *TaskTableMgr) GetLevelTasks(level int32) []*XmlTaskItem {
	if this.level_tasks == nil {
		return nil
	}
	return this.level_tasks[level]
}*/

func (this *TaskTableMgr) GetDailyTasks() map[int32]*XmlTaskItem {
	return this.daily_task_map
}

func (this *TaskTableMgr) GetStartAchieveTasks() []*XmlTaskItem {
	return this.start_achieve_tasks
}
