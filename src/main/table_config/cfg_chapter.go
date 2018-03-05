package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

type XmlChapterItem struct {
	ChapterId     int32  `xml:"ChapterId,attr"` // 配置Id
	StageRangeStr string `xml:"Range,attr"`     // 关卡范围
	MinStageId    int32
	MaxStageId    int32
	UnlockStarNum int32 `xml:"UnlockStarNum,attr"` // 解锁需要星星
	UnlockTime    int32 `xml:"UnlockTime,attr"`    // 解锁时间
}

type XmlChapterConfig struct {
	Items []XmlChapterItem `xml:"item"`
}

type CfgChapterManager struct {
	InitMaxStage  int32
	InitStageId   int32
	InitChapterId int32
	Map           map[int32]*XmlChapterItem
}

func (this *CfgChapterManager) Init() bool {
	if !this.Load() {
		log.Error("CfgChapterManager Init Load Failed !")
		return false
	}

	return true
}

func (this *CfgChapterManager) Load() bool {
	content, err := ioutil.ReadFile("../game_data/ChapterUnlockConfig.xml")
	if nil != err {
		log.Error("CfgChapterManager Load ReadFile error(%s)", err.Error())
		return false
	}

	tmp_cfg := &XmlChapterConfig{}
	err = xml.Unmarshal(content, tmp_cfg)
	if nil != err {
		log.Error("CfgChapterManager Load unmarshal err (%s)", err.Error())
		return false
	}

	this.Map = make(map[int32]*XmlChapterItem)
	tmp_len := int32(len(tmp_cfg.Items))
	var tmp_item *XmlChapterItem
	for idx := int32(0); idx < tmp_len; idx++ {
		tmp_item = &tmp_cfg.Items[idx]
		if nil == tmp_item {
			continue
		}

		tmp_arr := parse_xml_str_arr(tmp_item.StageRangeStr, ",")
		if 2 != len(tmp_arr) {
			log.Error("CfgChapterManager StageRangeStr[%s] error !", tmp_item.StageRangeStr)
			return false
		}

		tmp_item.MinStageId = tmp_arr[0]
		tmp_item.MaxStageId = tmp_arr[1]

		log.Info("解锁需要星星数目 %d", tmp_item.UnlockStarNum)
		this.Map[tmp_item.ChapterId] = tmp_item
		if 0 == tmp_item.UnlockStarNum {
			this.InitMaxStage = tmp_item.MaxStageId
			this.InitStageId = tmp_item.MinStageId
			this.InitChapterId = tmp_item.ChapterId
			log.Info("!!!!!!! 初始章节最大关卡[%d][%d - %d][%d]", tmp_item.ChapterId, tmp_item.MinStageId, tmp_item.MaxStageId, this.InitStageId)
		}
	}

	return true
}
