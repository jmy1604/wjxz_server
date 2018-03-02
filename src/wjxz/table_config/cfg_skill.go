package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

type XmlSkillItem struct {
	Id              int32  `xml:"Id,attr"`
	SkillValue1sStr string `xml:"SkillValue,attr"`
	SkillValue1s    []int32
	SkillValue2sStr string `xml:"SkillValue2,attr"`
	SkillValue2s    []int32
	EnergyCost      int32 `xml:"EnergyCost,attr"`
}

type XmlSkillConfig struct {
	Items []XmlSkillItem `xml:"item"`
}

type CfgSkillMgr struct {
	Map map[int32]*XmlSkillItem
}

func (this *CfgSkillMgr) Init() bool {
	data, err := ioutil.ReadFile("../game_data/Other.xml")
	if nil != err {
		log.Error("CfgSkillMgr Init read file err[%s] !", err.Error())
		return false
	}

	tmp_cfg := &XmlSkillConfig{}
	err = xml.Unmarshal(data, tmp_cfg)
	if nil != err {
		log.Error("CfgSkillMgr Init xml Unmarshal failed error [%s] !")
		return false
	}

	this.Map = make(map[int32]*XmlSkillItem)
	tmp_len := int32(len(tmp_cfg.Items))
	var tmp_item *XmlSkillItem
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]
		tmp_item.SkillValue1s = parse_xml_str_arr(tmp_item.SkillValue1sStr, ",")
		tmp_item.SkillValue2s = parse_xml_str_arr(tmp_item.SkillValue2sStr, ",")

		this.Map[tmp_item.Id] = tmp_item
	}

	return true
}
