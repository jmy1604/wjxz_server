package table_config

import (
	"encoding/json"
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

type XmlArenaRobotItem struct {
	Id               int32  `xml:"RobotID,attr"`
	RobotHead        int32  `xml:"RobotHead,attr"`
	RobotName        string `xml:"RobotName,attr"`
	RobotCardListStr string `xml:"RobotCardList,attr"`
	RobotScore       int32  `xml:"RobotScore,attr"`
	RobotCardList    []*JsonMonster
}

type XmlArenaRobotConfig struct {
	Items []XmlArenaRobotItem `xml:"item"`
}

type ArenaRobotTableMgr struct {
	Map   map[int32]*XmlArenaRobotItem
	Array []*XmlArenaRobotItem
}

func (this *ArenaRobotTableMgr) Init() bool {
	if !this.Load() {
		log.Error("ArenaRobotTableMgr Init load failed !")
		return false
	}
	return true
}

func (this *ArenaRobotTableMgr) Load() bool {
	data, err := ioutil.ReadFile("../game_data/ArenaRobot.xml")
	if nil != err {
		log.Error("ArenaRobotTableMgr read file err[%s] !", err.Error())
		return false
	}

	tmp_cfg := &XmlArenaRobotConfig{}
	err = xml.Unmarshal(data, tmp_cfg)
	if nil != err {
		log.Error("ArenaRobotTableMgr xml Unmarshal failed error [%s] !", err.Error())
		return false
	}

	if this.Map == nil {
		this.Map = make(map[int32]*XmlArenaRobotItem)
	}
	if this.Array == nil {
		this.Array = make([]*XmlArenaRobotItem, 0)
	}

	tmp_len := int32(len(tmp_cfg.Items))

	var tmp_item *XmlArenaRobotItem
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]
		if err = json.Unmarshal([]byte(tmp_item.RobotCardListStr), &tmp_item.RobotCardList); err != nil {
			log.Error("Parse RobotCardList[%v] error[%v]", tmp_item.RobotCardList, err.Error())
			return false
		}
		this.Map[tmp_item.Id] = tmp_item
		this.Array = append(this.Array, tmp_item)
	}

	return true
}

func (this *ArenaRobotTableMgr) Get(id int32) *XmlArenaRobotItem {
	return this.Map[id]
}
