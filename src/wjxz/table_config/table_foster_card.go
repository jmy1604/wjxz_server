package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

type XmlFosterCardItem struct {
	Id         int32  `xml:"Id,attr"`
	RewardStr  string `xml:"Reward,attr"`
	Rewards    []int32
	FosterTime int32 `xml:"FosterTime,attr"`
	ItemId     int32 `xml:"ItemId,attr"`
}

type XmlFosterCardConfig struct {
	Items []XmlFosterCardItem `xml:"item"`
}

type FosterCardTypeItems struct {
	Items []*XmlFosterCardItem
}

type FosterCardTableMgr struct {
	Map        map[int32]*XmlFosterCardItem
	Array      []*XmlFosterCardItem
	type2items map[int32]*FosterCardTypeItems
}

func (this *FosterCardTableMgr) Init() bool {
	if !this.Load() {
		log.Error("FosterCardTableMgr Init load failed !")
		return false
	}
	return true
}

func (this *FosterCardTableMgr) Load() bool {
	data, err := ioutil.ReadFile("../game_data/fostercard.xml")
	if nil != err {
		log.Error("FosterCardTableMgr read file err[%s] !", err.Error())
		return false
	}

	tmp_cfg := &XmlFosterCardConfig{}
	err = xml.Unmarshal(data, tmp_cfg)
	if nil != err {
		log.Error("FosterCardTableMgr xml Unmarshal failed error [%s] !", err.Error())
		return false
	}

	if this.Map == nil {
		this.Map = make(map[int32]*XmlFosterCardItem)
	}

	if this.Array == nil {
		this.Array = make([]*XmlFosterCardItem, 0)
	}

	if this.type2items == nil {
		this.type2items = make(map[int32]*FosterCardTypeItems)
	}

	tmp_len := int32(len(tmp_cfg.Items))
	var tmp_item *XmlFosterCardItem
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]

		rewards := parse_xml_str_arr(tmp_item.RewardStr, ",")
		if rewards == nil || len(rewards)%2 != 0 {
			log.Error("foster table parse field Reward[%v] error", tmp_item.RewardStr)
			return false
		}
		tmp_item.Rewards = rewards

		this.Map[tmp_item.Id] = tmp_item
		this.Array = append(this.Array, tmp_item)
	}

	return true
}

func (this *FosterCardTableMgr) Has(id int32) bool {
	if d := this.Map[id]; d == nil {
		return false
	}
	return true
}

func (this *FosterCardTableMgr) Get(id int32) *XmlFosterCardItem {
	return this.Map[id]
}

func (this *FosterCardTableMgr) GetByItemId(item_id int32) *XmlFosterCardItem {
	if this.Array == nil {
		return nil
	}

	for _, a := range this.Array {
		if a.ItemId == item_id {
			return a
		}
	}

	return nil
}
