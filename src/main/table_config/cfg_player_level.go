package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

type XmlPlayerLevelItem struct {
	Level        int32 `xml:"Level,attr"`
	MaxExp       int32 `xml:"MaxExp,attr"`
	MaxFarm      int32 `xml:"MaxFarm,attr"`
	MaxCattery   int32 `xml:"MaxCattery,attr"`
	MaxPower     int32 `xml:"MaxPower,attr"`
	FosteredSlot int32 `xml:"BeFriendFosterSlot,attr"` // 被寄养上限
	FosterSlot   int32 `xml:"FriendFosterSlot,attr"`   // 寄养上限
}

type XmlPlayerLevelConfig struct {
	Items []XmlPlayerLevelItem `xml:"item"`
}

type CfgPlayerLevelManager struct {
	Map      map[int32]*XmlPlayerLevelItem
	Array    []*XmlPlayerLevelItem
	MaxLevel int32
}

func (this *CfgPlayerLevelManager) Init() bool {
	if !this.Load() {
		log.Error("CfgPlayerLevelManager Init load failed !")
		return false
	}
	return true
}

func (this *CfgPlayerLevelManager) Load() bool {
	data, err := ioutil.ReadFile("../game_data/PlayerLevel.xml")
	if nil != err {
		log.Error("CfgPlayerLevelManager Load read file err[%s] !", err.Error())
		return false
	}

	tmp_cfg := &XmlPlayerLevelConfig{}
	err = xml.Unmarshal(data, tmp_cfg)
	if nil != err {
		log.Error("CfgPlayerLevelManager Load xml Unmarshal failed error [%s] !", err.Error())
		return false
	}

	if this.Map == nil {
		this.Map = make(map[int32]*XmlPlayerLevelItem)
	}
	if this.Array == nil {
		this.Array = make([]*XmlPlayerLevelItem, 0)
	}
	tmp_len := int32(len(tmp_cfg.Items))
	var tmp_item *XmlPlayerLevelItem
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]
		this.Map[tmp_item.Level] = tmp_item
		this.Array = append(this.Array, tmp_item)
	}

	this.MaxLevel = tmp_len

	return true
}
