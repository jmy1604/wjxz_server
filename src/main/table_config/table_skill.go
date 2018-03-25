package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

type XmlSkillItem struct {
	Id                   int32  `xml:"ID,attr"`
	Type                 int32  `xml:"Type,attr"`
	SkillAttrStr         string `xml:"SkillAttr,attr"`
	SkillAttr            []int32
	SkillTriggerType     int32  `xml:"SkillTriggerType,attr"`
	TriggerCondition1Str string `xml:"TriggerCondition1,attr"`
	TriggerCondition1    []int32
	TriggerCondition2Str string `xml:"TriggerCondition2,attr"`
	TriggerCondition2    []int32
	TriggerRoundMax      int32
	TriggerBattleMax     int32
	SkillBuff            int32 `xml:"SkillBuff,attr"`
	SkillTrigger         int32 `xml:"SkillTrigger,attr"`
	SkillMelee           int32 `xml:"SkillMelee,attr"`
	SkillEnemy           int32 `xml:"SkillEnemy,attr"`
	RangeType            int32 `xml:"RangeType,attr"`
	SkillTarget          int32 `xml:"SkillTarget,attr"`
	MaxTarget            int32 `xml:"MaxTarget,attr"`
	MustHit              int32 `xml:"CertainHit,attr"`
	SkillCastCount       int32 `xml:"SkillCastCount,attr"`

	Effect1Cond1Str string `xml:"Effect1Cond1,attr"`
	Effect1Cond2Str string `xml:"Effect1Cond2,attr"`
	Effect1Str      string `xml:"Effect1,attr"`
	Effect1         []int32
	Effect1Cond1    []int32
	Effect1Cond2    []int32

	Effect2Cond1Str string `xml:"Effect2Cond1,attr"`
	Effect2Cond2Str string `xml:"Effect2Cond2,attr"`
	Effect2Str      string `xml:"Effect2,attr"`
	Effect2         []int32
	Effect2Cond1    []int32
	Effect2Cond2    []int32

	Effect3Cond1Str string `xml:"Effect3Cond1,attr"`
	Effect3Cond2Str string `xml:"Effect3Cond2,attr"`
	Effect3Str      string `xml:"Effect3,attr"`
	Effect3         []int32
	Effect3Cond1    []int32
	Effect3Cond2    []int32

	Effects       [][]int32
	EffectsCond1s [][]int32
	EffectsCond2s [][]int32

	ComboSKill int32 `xml:"ComboSKill,attr"`
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
		tmp_item.SkillAttr = parse_xml_str_arr2(tmp_item.SkillAttrStr, ",")
		if tmp_item.SkillAttr == nil {
			tmp_item.SkillAttr = make([]int32, 0)
		}

		tmp_item.TriggerCondition1 = parse_xml_str_arr2(tmp_item.TriggerCondition1Str, "|")
		if tmp_item.TriggerCondition1 == nil {
			tmp_item.TriggerCondition1 = make([]int32, 0)
		}

		tmp_item.TriggerCondition2 = parse_xml_str_arr2(tmp_item.TriggerCondition2Str, "|")
		if tmp_item.TriggerCondition2 == nil {
			tmp_item.TriggerCondition2 = make([]int32, 0)
		}

		tmp_item.Effect1Cond1 = parse_xml_str_arr2(tmp_item.Effect1Cond1Str, "|")
		if tmp_item.Effect1Cond1 == nil {
			tmp_item.Effect1Cond1 = make([]int32, 0)
		}
		tmp_item.Effect1Cond2 = parse_xml_str_arr2(tmp_item.Effect1Cond2Str, "|")
		if tmp_item.Effect1Cond2 == nil {
			tmp_item.Effect1Cond2 = make([]int32, 0)
		}
		tmp_item.Effect1 = parse_xml_str_arr2(tmp_item.Effect1Str, "|")
		if tmp_item.Effect1 == nil {
			tmp_item.Effect1 = make([]int32, 0)
		}

		tmp_item.Effect2Cond1 = parse_xml_str_arr2(tmp_item.Effect2Cond1Str, "|")
		if tmp_item.Effect2Cond1 == nil {
			tmp_item.Effect2Cond1 = make([]int32, 0)
		}
		tmp_item.Effect2Cond2 = parse_xml_str_arr2(tmp_item.Effect2Cond2Str, "|")
		if tmp_item.Effect2Cond2 == nil {
			tmp_item.Effect2Cond2 = make([]int32, 0)
		}
		tmp_item.Effect2 = parse_xml_str_arr2(tmp_item.Effect2Str, "|")
		if tmp_item.Effect2 == nil {
			tmp_item.Effect2 = make([]int32, 0)
		}

		tmp_item.Effect3Cond1 = parse_xml_str_arr2(tmp_item.Effect3Cond1Str, "|")
		if tmp_item.Effect3Cond1 == nil {
			tmp_item.Effect3Cond1 = make([]int32, 0)
		}
		tmp_item.Effect3Cond2 = parse_xml_str_arr2(tmp_item.Effect3Cond2Str, "|")
		if tmp_item.Effect3Cond2 == nil {
			tmp_item.Effect3Cond2 = make([]int32, 0)
		}
		tmp_item.Effect3 = parse_xml_str_arr2(tmp_item.Effect3Str, "|")
		if tmp_item.Effect3 == nil {
			tmp_item.Effect3 = make([]int32, 0)
		}

		tmp_item.Effects = [][]int32{
			tmp_item.Effect1, tmp_item.Effect2, tmp_item.Effect3,
		}
		tmp_item.EffectsCond1s = [][]int32{
			tmp_item.Effect1Cond1, tmp_item.Effect2Cond1, tmp_item.Effect3Cond1,
		}
		tmp_item.EffectsCond2s = [][]int32{
			tmp_item.Effect1Cond2, tmp_item.Effect2Cond2, tmp_item.Effect3Cond2,
		}

		this.Map[tmp_item.Id] = tmp_item
		this.Array = append(this.Array, tmp_item)
	}

	return true
}

func (this *SkillTableMgr) Get(skill_id int32) *XmlSkillItem {
	return this.Map[skill_id]
}