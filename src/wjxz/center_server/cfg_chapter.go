package main

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

type XmlChapterItem struct {
	ChapterId     int32 `xml:"ChapterId,attr"`     // 配置Id
	MaxId         int32 `xml:"MaxId,attr"`         // 最大Id
	UnlockStarNum int32 `xml:"UnlockStarNum,attr"` // 解锁需要星星
	UnlockTime    int32 `xml:"UnlockTime,attr"`    // 解锁时间
}

type XmlChapterConfig struct {
	Items []XmlChapterItem `xml:"item"`
}

type CfgChapterManager struct {
	InitMaxStage int32
	Map          map[int32]*XmlChapterItem
}

var cfg_chapter_mgr CfgChapterManager

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

		this.Map[tmp_item.ChapterId] = tmp_item
		if 1 == tmp_item.ChapterId {
			this.InitMaxStage = tmp_item.MaxId
		}
	}

	return true
}
