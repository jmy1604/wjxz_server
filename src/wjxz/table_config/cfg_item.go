package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

const (
	ITEM_CFG_TYPE_BUILDING = 0
	ITEM_CFG_TYPE_PROP     = 1
	ITEM_CFG_TYPE_THINGS   = 2
)

type XmlItemItem struct {
	CfgId        int32 `xml:"Id,attr"`
	ItemType     int32
	SaleCoin     int32 `xml:"SaleCoin,attr"`
	MaxNumber    int32 `xml:"MaxNumber,attr"`
	Cost         int32 `xml:"Cost,attr"`
	Diamond      int32 `xml:"Diamond,attr"`
	Type         int32 `xml:"Type,attr"`
	UseType      int32 `xml:"UseType,attr"`
	ConstantTime int32 `xml:"ConstantTime,attr"`
	ValidTime    int32 `xml:"ValidTime,attr"`
	Numbers      []int32
	NumberStr    string `xml:"Number,attr"`
}

/*
type XmlBuildingItem struct {
	// building 建筑库
	MaxLevel      int32  `xml:"MaxLevel,attr"`
	Type          int32  `xml:"Type,attr"`
	Rarity        int32  `xml:"Rarity,attr"`
	UnlockType    int32  `xml:"UnlockType,attr"`
	UnlockLevel   int32  `xml:"UnlockLevel,attr"`
	UnlockCostStr string `xml:"UnlockCost,attr"`
	UnlockCosts   []int32
	BuildTime     int32  `xml:"BuildTime,attr"`
	Charm         int32  `xml:"Charm,attr"`
	Geography     int32  `xml:"Geography,attr"`
	MapSizeStr    string `xml:"MapSize,attr"`
	MapSizes      []int32
}
*/
type XmlItemConfig struct {
	Items []XmlItemItem `xml:"item"`
}

type CfgItemManager struct {
	Map   map[int32]*XmlItemItem
	Array []*XmlItemItem
}

func (this *CfgItemManager) Init() bool {
	if !this.Load() {
		log.Error("CfgItemManager Init load failed !")
		return false
	}
	return true
}

func (this *CfgItemManager) Load() bool {
	if !this.LoadBuilding() {
		return false
	}

	if !this.LoadProp() {
		return false
	}

	if !this.LoadThings() {
		return false
	}

	return true
}

func (this *CfgItemManager) LoadBuilding() bool {
	/*data, err := ioutil.ReadFile("../game_data/Building.xml")
	if nil != err {
		log.Error("CfgBuildingMgr LoadBuilding read file err[%s] !", err.Error())
		return false
	}

	tmp_cfg := &XmlItemConfig{}
	err = xml.Unmarshal(data, tmp_cfg)
	if nil != err {
		log.Error("CfgBuildingMgr LoadBuilding xml unmarshal error (%s) ", err.Error())
		return false
	}

	tmp_len := int32(len(tmp_cfg.Items))
	var tmp_item *XmlItemItem
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]

		tmp_item.UnlockCosts = parse_xml_str_arr(tmp_item.UnlockCostStr, ",")
		tmp_item.MapSizes = parse_xml_str_arr(tmp_item.MapSizeStr, ",")
		tmp_item.ItemType = ITEM_CFG_TYPE_BUILDING
		this.Map[tmp_item.CfgId] = tmp_item
	}*/

	return true
}

func (this *CfgItemManager) LoadProp() bool {
	data, err := ioutil.ReadFile("../game_data/Prop.xml")
	if nil != err {
		log.Error("CfgItemManager LoadProp read file err[%s] !", err.Error())
		return false
	}

	tmp_cfg := &XmlItemConfig{}
	err = xml.Unmarshal(data, tmp_cfg)
	if nil != err {
		log.Error("CfgItemManager LoadProp xml Unmarshal failed error [%s] !", err.Error())
		return false
	}

	if this.Map == nil {
		this.Map = make(map[int32]*XmlItemItem)
	}
	if this.Array == nil {
		this.Array = make([]*XmlItemItem, 0)
	}
	tmp_len := int32(len(tmp_cfg.Items))
	var tmp_item *XmlItemItem
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]
		tmp_item.ItemType = ITEM_CFG_TYPE_PROP
		this.Map[tmp_item.CfgId] = tmp_item
		this.Array = append(this.Array, tmp_item)
	}

	return true
}

func (this *CfgItemManager) LoadThings() bool {
	data, err := ioutil.ReadFile("../game_data/Things.xml")
	if nil != err {
		log.Error("CfgItemManager LoadThings read file err[%s] !", err.Error())
		return false
	}

	tmp_cfg := &XmlItemConfig{}
	err = xml.Unmarshal(data, tmp_cfg)
	if nil != err {
		log.Error("CfgItemManager LoadThings xml Unmarshal failed error [%s] !", err.Error())
		return false
	}

	if this.Map == nil {
		this.Map = make(map[int32]*XmlItemItem)
	}
	if this.Array == nil {
		this.Array = make([]*XmlItemItem, 0)
	}
	tmp_len := int32(len(tmp_cfg.Items))
	var tmp_item *XmlItemItem
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]
		tmp_item.ItemType = ITEM_CFG_TYPE_THINGS
		tmp_item.Numbers = parse_xml_str_arr(tmp_item.NumberStr, ",")
		this.Map[tmp_item.CfgId] = tmp_item
		this.Array = append(this.Array, tmp_item)
		//log.Info("item[%v] config: %v", tmp_item.CfgId, tmp_item)
	}

	return true
}
