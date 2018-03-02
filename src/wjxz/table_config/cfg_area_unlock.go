package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

type XmlAreaUnlockItem struct {
	AreaId             int32  `xml:"AreaId,attr"`
	FrontAreaStr       string `xml:"FrontArea,attr"`
	FrontAreas         []int32
	UnlockLevel        int32  `xml:"UnlockLevel,attr"`
	UnlockStar         int32  `xml:"UnlockStar,attr"`
	UnlockCostStr      string `xml:"UnlockCost,attr"`
	UnlockCost         []int32
	QuickUnlockCostStr string `xml:"QuickUnlock,attr"`
	QuickUnlockCost    []int32
	MaxObstacle        int32 `xml:"MaxObstacle,attr"`
}

type XmlAreaUnlockConfig struct {
	Items []XmlAreaUnlockItem `xml:"item"`
}

type CfgAreaUnlockMgr struct {
	Map         map[int32]*XmlAreaUnlockItem
	InitAreaIds []int32
}

func (this *CfgAreaUnlockMgr) Init() bool {
	this.InitAreaIds = make([]int32, 0, 10)
	data, err := ioutil.ReadFile("../game_data/Area.xml")
	if nil != err {
		log.Error("CfgAreaUnlockMgr Init read file err[%s] !", err.Error())
		return false
	}

	tmp_cfg := &XmlAreaUnlockConfig{}
	err = xml.Unmarshal(data, tmp_cfg)
	if nil != err {
		log.Error("CfgAreaUnlockMgr Init xml Unmarshal failed error [%s] !", err.Error())
		return false
	}

	this.Map = make(map[int32]*XmlAreaUnlockItem)
	tmp_len := int32(len(tmp_cfg.Items))
	var tmp_item *XmlAreaUnlockItem
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]
		tmp_item.FrontAreas = parse_xml_str_arr(tmp_item.FrontAreaStr, ",")
		tmp_item.UnlockCost = parse_xml_str_arr(tmp_item.UnlockCostStr, ",")
		if len(tmp_item.UnlockCost)%2 != 0 {
			log.Error("CfgAreaUnlockMgr Init UnlockCost[%s] error !", tmp_item.UnlockCostStr)
			return false
		}
		tmp_item.QuickUnlockCost = parse_xml_str_arr(tmp_item.QuickUnlockCostStr, ",")
		if len(tmp_item.QuickUnlockCost)%2 != 0 {
			log.Error("CfgAreaUnlockMgr Init QuickUnlockCost[%s] error !", tmp_item.QuickUnlockCostStr)
			return false
		}

		if 0 == tmp_item.UnlockLevel {
			this.InitAreaIds = append(this.InitAreaIds, tmp_item.AreaId)
		}

		this.Map[tmp_item.AreaId] = tmp_item
	}

	log.Info("初始化解锁区域 %v", this.InitAreaIds)

	return true
}
