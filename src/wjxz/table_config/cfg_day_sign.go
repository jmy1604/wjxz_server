package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

type XmlDaySignItem struct {
	Year            int32 `xml:"Year,attr"`
	Month           int32 `xml:"Month,attr"`
	Day             int32 `xml:"Day,attr"`
	Camp1CardID     int32 `xml:"Camp1CardID,attr"`
	Camp1CardCount  int32 `xml:"Camp1CardCount,attr"`
	Camp2CardID     int32 `xml:"Camp2CardID,attr"`
	Camp2CardCount  int32 `xml:"Camp2CardCount,attr"`
	GoldCount       int32 `xml:"GoldCount,attr"`
	GemCount        int32 `xml:"GemCount,attr"`
	CardToken1Count int32 `xml:"CardToken1Count,attr"`
	CardToken2Count int32 `xml:"CardToken2Count,attr"`
	CardToken3Count int32 `xml:"CardToken3Count,attr"`
	CardToken4Count int32 `xml:"CardToken4Count,attr"`
	ChestID         int32 `xml:"ChestID,attr"`
}

type XmlDaySignConfig struct {
	Items []XmlDaySignItem `xml:"item"`
}

type CfgDaySignManager struct {
	Map           map[int32]*XmlDaySignItem
	SunNum2Reward map[int32]*DaySignSumReward
}

func (this *CfgDaySignManager) Init() bool {
	if !this.LoadDaySignReward() || !this.LoadDaySignSumReward() {
		log.Error("CfgDaySignManager Init failed !")
		return false
	}

	log.Info("每日签到配置 %v", this.Map[20170605])

	return true
}

func (this *CfgDaySignManager) LoadDaySignReward() bool {
	content, err := ioutil.ReadFile("../game_data/DailyRegister.xml")
	if nil != err {
		log.Error("CfgDaySignManager Load read file error [%s]", err.Error())
		return false
	}

	tmp_cfg := &XmlDaySignConfig{}
	err = xml.Unmarshal(content, tmp_cfg)
	if nil != err {
		log.Error("CfgDaySignManager Load Unmarshal content error !")
		return false
	}

	this.Map = make(map[int32]*XmlDaySignItem)
	var tmp_item *XmlDaySignItem
	tmp_len := int32(len(tmp_cfg.Items))
	for i := int32(0); i < tmp_len; i++ {
		tmp_item = &tmp_cfg.Items[i]
		if nil == tmp_item {
			continue
		}

		//log.Info("Add DaySign key[%d]", tmp_item.Year*10000+tmp_item.Month*100+tmp_item.Day)
		this.Map[tmp_item.Year*10000+tmp_item.Month*100+tmp_item.Day] = tmp_item
	}

	return true
}

func (this *CfgDaySignManager) LoadDaySignSumReward() bool {
	this.SunNum2Reward = make(map[int32]*DaySignSumReward)

	/*tmp_len := int32(len(global_config.DaySignSumRewards))
	for i := int32(0); i < tmp_len; i++ {
		val := &global_config.DaySignSumRewards[i]
		log.Info("add daysignsum %v", val)
		this.SunNum2Reward[val.SignNum] = val
	}

	log.Info("累计签到奖励配置%v %v %v", this.SunNum2Reward[1], tmp_len, this.SunNum2Reward)*/
	return true
}
