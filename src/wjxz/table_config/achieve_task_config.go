package table_config

/*
import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

type XmlAchieveItem struct {
	AchievementTaskID int32 `xml:"AchievementTaskID,attr"`
	FrontTask         int32 `xml:"FrontTask,attr"`
	FinishConditionID int32 `xml:"FinishConditionID,attr"`
	FinishNeedCount   int32 `xml:"FinishNeedCount,attr"`
	RewardGold        int32 `xml:"RewardGold,attr"`
	RewardGem         int32 `xml:"RewardGem,attr"`
	RewardExp         int32 `xml:"RewardExp,attr"`
}

type XmlAchieveConfig struct {
	Items []XmlAchieveItem `xml:"item"`
}

type SubTypeAchieves struct {
	Count int32
	Array []*XmlAchieveItem
}

type XmlDialyTaskItem struct {
	DailyTasksID      int32 `xml:"DailyTasksID,attr"`
	FinishConditionID int32 `xml:"FinishConditionID,attr"`
	FinishNeedCount   int32 `xml:"FinishNeedCount,attr"`
	RewardGold        int32 `xml:"RewardGold,attr"`
	RewardExp         int32 `xml:"RewardExp,attr"`
}

type XmlDialyTaskConfig struct {
	Items []XmlDialyTaskItem `xml:"item"`
}

type SubTypeDialyTasks struct {
	Count int32
	Array []*XmlDialyTaskItem
}

type AchieveTaskConfigMgr struct {
	DialyTaskMap       map[int32]*XmlDialyTaskItem
	DialyTaskArray     []*XmlDialyTaskItem
	DialyTaskArray_len int32
	Type2SubDialyTasks map[int32]*SubTypeDialyTasks

	AchieveMap       map[int32]*XmlAchieveItem
	AchieveArray     []*XmlAchieveItem
	AchieveArray_len int32
	Typ2SubAchieves  map[int32]*SubTypeAchieves
}

var achieve_task_mgr AchieveTaskConfigMgr

func (this *AchieveTaskConfigMgr) Init() bool {
	if !this.LoadDialyTask() {
		return false
	}

	if !this.LoadAchieve() {
		return false
	}

	return true
}

func (this *AchieveTaskConfigMgr) LoadDialyTask() bool {
	content, err := ioutil.ReadFile("../game_data/DailyTasksConfig.xml")
	if nil != err {
		log.Error("AchieveTaskConfigMgr LoadDialyTask read file error !")
		return false
	}

	tmp_cfg := &XmlDialyTaskConfig{}
	err = xml.Unmarshal(content, tmp_cfg)
	if nil != err {
		log.Error("AchieveTaskConfigMgr LoadDialyTask unmarshal failed(%s)", err.Error())
		return false
	}

	tmp_len := int32(len(tmp_cfg.Items))
	this.DialyTaskArray = make([]*XmlDialyTaskItem, 0, tmp_len)
	this.DialyTaskMap = make(map[int32]*XmlDialyTaskItem)
	this.Type2SubDialyTasks = make(map[int32]*SubTypeDialyTasks)
	var tmp_item *XmlDialyTaskItem
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]
		this.DialyTaskMap[tmp_item.DailyTasksID] = tmp_item
		this.DialyTaskArray = append(this.DialyTaskArray, tmp_item)
		if nil == this.Type2SubDialyTasks[tmp_item.FinishConditionID] {
			this.Type2SubDialyTasks[tmp_item.FinishConditionID] = &SubTypeDialyTasks{}
		}
		this.Type2SubDialyTasks[tmp_item.FinishConditionID].Count++
	}
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]
		if nil == this.Type2SubDialyTasks[tmp_item.FinishConditionID].Array {
			this.Type2SubDialyTasks[tmp_item.FinishConditionID].Array = make([]*XmlDialyTaskItem, 0, this.Type2SubDialyTasks[tmp_item.FinishConditionID].Count)
		}
		this.Type2SubDialyTasks[tmp_item.FinishConditionID].Array = append(this.Type2SubDialyTasks[tmp_item.FinishConditionID].Array, tmp_item)
	}

	this.DialyTaskArray_len = int32(len(this.DialyTaskArray))

	return true
}

func (this *AchieveTaskConfigMgr) LoadAchieve() bool {
	content, err := ioutil.ReadFile("../game_data/AchievementConfig.xml")
	if nil != err {
		log.Error("AchieveTaskConfigMgr LoadAchieve failed(%s) !", err.Error())
		return false
	}

	tmp_cfg := &XmlAchieveConfig{}
	err = xml.Unmarshal(content, tmp_cfg)
	if nil != err {
		log.Error("AchieveTaskConfigMgr LoadAchieve unmarshal failed(%s) !", err.Error())
		return false
	}

	tmp_len := int32(len(tmp_cfg.Items))
	this.AchieveMap = make(map[int32]*XmlAchieveItem)
	this.AchieveArray = make([]*XmlAchieveItem, 0, tmp_len)
	this.Typ2SubAchieves = make(map[int32]*SubTypeAchieves)
	var tmp_item *XmlAchieveItem
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]
		this.AchieveMap[tmp_item.AchievementTaskID] = tmp_item
		this.AchieveArray = append(this.AchieveArray, tmp_item)
		if nil == this.Typ2SubAchieves[tmp_item.FinishConditionID] {
			this.Typ2SubAchieves[tmp_item.FinishConditionID] = &SubTypeAchieves{}
		}
		this.Typ2SubAchieves[tmp_item.FinishConditionID].Count++
	}

	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]
		if nil == this.Typ2SubAchieves[tmp_item.FinishConditionID].Array {
			this.Typ2SubAchieves[tmp_item.FinishConditionID].Array = make([]*XmlAchieveItem, 0, this.Typ2SubAchieves[tmp_item.FinishConditionID].Count)
		}
		this.Typ2SubAchieves[tmp_item.FinishConditionID].Array = append(this.Typ2SubAchieves[tmp_item.FinishConditionID].Array, tmp_item)
	}

	for tmp_t, subachieve := range this.Typ2SubAchieves {
		log.Info("成就任务类型【%d】", tmp_t)
		for _, val := range subachieve.Array {
			log.Info("	成就任务：%d %d", val.AchievementTaskID, val.FinishNeedCount)
		}
	}

	this.AchieveArray_len = int32(len(this.AchieveArray))

	return true
}*/
