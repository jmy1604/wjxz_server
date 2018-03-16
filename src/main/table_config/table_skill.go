package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

type XmlSkillItem struct {
	Id             int32  `xml:"ID,attr"`
	Type           int32  `xml:"Type,attr"`
	SkillAttr      int32  `xml:"SkillAttr,attr"`
	SkillBuff      int32  `xml:"SkillBuff,attr"`
	SkillTrigger   int32  `xml:"SkillTrigger,attr"`
	SkillMelee     int32  `xml:"SkillMelee,attr"`
	SkillEnemy     int32  `xml:"SkillEnemy,attr"`
	RangeType      int32  `xml:"RangeType,attr"`
	SkillTarget    int32  `xml:"SkillTarget,attr"`
	MaxTarget      int32  `xml:"MaxTarget,attr"`
	SkillCastCount int32  `xml:"SkillCastCount,attr"`
	Effect1        string `xml:"Effect1,attr"`
	Effect2        string `xml:"Effect2,attr"`
	Effect3        string `xml:"Effect3,attr"`
	ComboSKill     int32  `xml:"ComboSKill,attr"`
}

type XmlSkillConfig struct {
	Items []XmlSkillItem `xml:"item"`
}

type SkillTableMgr struct {
	Map   map[int32]*XmlSkillItem
	Array []*XmlSkillItem
}

func (this *SkillTableMgr) Init() bool {
	if !this.Load() {
		log.Error("SkillTableMgr Init load failed !")
		return false
	}
	return true
}

func (this *SkillTableMgr) Load() bool {
	data, err := ioutil.ReadFile("../game_data/skill.xml")
	if nil != err {
		log.Error("SkillTableMgr read file err[%s] !", err.Error())
		return false
	}

	tmp_cfg := &XmlSkillConfig{}
	err = xml.Unmarshal(data, tmp_cfg)
	if nil != err {
		log.Error("SkillTableMgr xml Unmarshal failed error [%s] !", err.Error())
		return false
	}

	if this.Map == nil {
		this.Map = make(map[int32]*XmlSkillItem)
	}
	if this.Array == nil {
		this.Array = make([]*XmlSkillItem, 0)
	}
	tmp_len := int32(len(tmp_cfg.Items))

	var tmp_item *XmlSkillItem
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]

		this.Map[tmp_item.Id] = tmp_item
		this.Array = append(this.Array, tmp_item)
	}

	return true
}

func (this *SkillTableMgr) Get(skill_id int32) *XmlSkillItem {
	return this.Map[skill_id]
}
