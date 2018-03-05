package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

type XmlMailItem struct {
	MailId       int32 `xml:"MailID,attr"`
	MailDuration int32 `xml:"MailDuration,attr"`
	Reward1      int32 `xml:"Reward1,attr"`
	Reward1Count int32 `xml:"Reward1Count,attr"`
	Reward2      int32 `xml:"Reward2,attr"`
	Reward2Count int32 `xml:"Reward2Count,attr"`
	Reward3      int32 `xml:"Reward3,attr"`
	Reward3Count int32 `xml:"Reward3Count,attr"`
	Reward4      int32 `xml:"Reward4,attr"`
	Reward4Count int32 `xml:"Reward4Count,attr"`
	Reward5      int32 `xml:"Reward5,attr"`
	Reward5Count int32 `xml:"Reward5Count,attr"`
	Reward6      int32 `xml:"Reward6,attr"`
	Reward6Count int32 `xml:"Reward7Count,attr"`
	RewardIds    []int32
	RewardNums   []int32
}

type XmlMailConfifg struct {
	Items []XmlMailItem `xml:"item"`
}

type MailConfigManager struct {
	Map map[int32]*XmlMailItem
}

func (this *MailConfigManager) Init() bool {
	if !this.Load() {
		return false
	}
	return true
}

func (this *MailConfigManager) Load() bool {
	data, err := ioutil.ReadFile("../game_data/MailConfig.xml")
	if nil != err {
		log.Error("MailConfigManager Load failed to read file [%s] !", err.Error())
		return false
	}

	tmp_cfg := &XmlMailConfifg{}
	err = xml.Unmarshal(data, tmp_cfg)
	if nil != err {
		log.Error("MailConfigManager Load failed to unmarshal data [%s]", err.Error())
		return false
	}

	tmp_len := int32(len(tmp_cfg.Items))
	if tmp_len < 1 {
		log.Error("MailConfigManager load cfg items empty !")
		return false
	}

	var tmp_item *XmlMailItem
	this.Map = make(map[int32]*XmlMailItem)
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]
		tmp_item.RewardIds = make([]int32, 0, 6)
		tmp_item.RewardNums = make([]int32, 0, 6)
		this.Map[tmp_item.MailId] = tmp_item
		if tmp_item.Reward1 > 0 && tmp_item.Reward1Count > 0 {
			tmp_item.RewardIds = append(tmp_item.RewardIds, tmp_item.Reward1)
			tmp_item.RewardNums = append(tmp_item.RewardNums, tmp_item.Reward1Count)
		}
		if tmp_item.Reward2 > 0 && tmp_item.Reward2Count > 0 {
			tmp_item.RewardIds = append(tmp_item.RewardIds, tmp_item.Reward2)
			tmp_item.RewardNums = append(tmp_item.RewardNums, tmp_item.Reward2Count)
		}
		if tmp_item.Reward3 > 0 && tmp_item.Reward3Count > 0 {
			tmp_item.RewardIds = append(tmp_item.RewardIds, tmp_item.Reward3)
			tmp_item.RewardNums = append(tmp_item.RewardNums, tmp_item.Reward3Count)
		}
		if tmp_item.Reward4 > 0 && tmp_item.Reward4Count > 0 {
			tmp_item.RewardIds = append(tmp_item.RewardIds, tmp_item.Reward4)
			tmp_item.RewardNums = append(tmp_item.RewardNums, tmp_item.Reward4Count)
		}
		if tmp_item.Reward5 > 0 && tmp_item.Reward5Count > 0 {
			tmp_item.RewardIds = append(tmp_item.RewardIds, tmp_item.Reward5)
			tmp_item.RewardNums = append(tmp_item.RewardNums, tmp_item.Reward5Count)
		}
		if tmp_item.Reward6 > 0 && tmp_item.Reward6Count > 0 {
			tmp_item.RewardIds = append(tmp_item.RewardIds, tmp_item.Reward6)
			tmp_item.RewardNums = append(tmp_item.RewardNums, tmp_item.Reward6Count)
		}
	}

	return true
}
