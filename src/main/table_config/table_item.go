package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

type XmlItemItem struct {
	Id       int32  `xml:"Id,attr"`
	Type     int32  `xml:"Type,attr"`
	MaxCount string `xml:"MaxCount,attr"`
}

type XmlItemConfig struct {
	Items []XmlItemItem `xml:"item"`
}

type ItemTableMgr struct {
	Map   map[int32]*XmlItemItem
	Array []*XmlItemItem
}

func (this *ItemTableMgr) Init() bool {
	if !this.Load() {
		log.Error("ItemTableMgr Init load failed !")
		return false
	}
	return true
}

func (this *ItemTableMgr) Load() bool {
	data, err := ioutil.ReadFile("../game_data/item.xml")
	if nil != err {
		log.Error("ItemTableMgr read file err[%s] !", err.Error())
		return false
	}

	tmp_cfg := &XmlItemConfig{}
	err = xml.Unmarshal(data, tmp_cfg)
	if nil != err {
		log.Error("ItemTableMgr xml Unmarshal failed error [%s] !", err.Error())
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

		this.Map[tmp_item.Id] = tmp_item
		this.Array = append(this.Array, tmp_item)
	}

	return true
}

func (this *ItemTableMgr) Get(id int32) *XmlItemItem {
	return this.Map[id]
}
