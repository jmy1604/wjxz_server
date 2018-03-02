package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
	"math/rand"
	"strconv"
	"strings"
)

type XmlChestCycleItem struct {
	ChestOrder   int32 `xml:"ChestOrder,attr"`
	ChestQuelity int32 `xml:"ChestQuelity,attr"`
}

type XmlChestCycleConfig struct {
	Items []XmlChestCycleItem `xml:"item"`
}

type XmlDropIdItem struct {
	DropID    int32 `xml:"DropID,attr"`
	DropGoods int32 `xml:"DropGoods,attr"`
	DropOdds  int32 `xml:"DropOdds,attr"`
}

type XmlDropIdConfig struct {
	Items []XmlDropIdItem `xml:"item"`
}

type DropPackage struct {
	DropID    int32
	TotalOdds int32
	ItemCount int32
	DropItems []*XmlDropIdItem
}

type DropAmountInfo struct {
	Odds  int32 // 权重
	Count int32 // 数量
}

type XmlChestItem struct {
	ChestID      int32  `xml:"ChestID,attr"`
	Camp         int32  `xml:"Camp,attr"`
	ArenaRating  int32  `xml:"ArenaRating,attr"`
	ChestQuality int32  `xml:"ChestQuality,attr"`
	DropOdds     int32  `xml:"DropOdds,attr"`
	GoldCost     int32  `xml:"GoldCost,attr"`
	GemCost      int32  `xml:"GemCost,attr"`
	UnlockSecStr string `xml:"UnlockTime,attr"`
	UnlockSec    int32

	GoldExtractTimes int32 `xml:"GoldExtractTimes,attr"`
	GoldMin          int32 `xml:"GoldMin,attr"`
	GoldMax          int32 `xml:"GoldMax,attr"`

	GemExtractTimes int32 `xml:"GemExtractTimes,attr"`
	GemMin          int32 `xml:"GemMin,attr"`
	GemMax          int32 `xml:"GemMax,attr"`

	CardTokenTotalExtractTimes int32 `xml:"CardTokenTotalExtractTimes,attr"`
	CardTokenTotalMin          int32 `xml:"CardTokenTotalMin,attr"`
	CardTokenTotalMax          int32 `xml:"CardTokenTotalMax,attr"`

	CardToken2DropOdds   int32 `xml:"CardToken2DropOdds,attr"`
	CardToken2ExtraTimes int32 `xml:"CardToken2ExtractTimes,attr"`
	CardToken2Min        int32 `xml:"CardToken2Min,attr"`
	CardToken2Max        int32 `xml:"CardToken2Max,attr"`

	CardToken3DropOdds   int32 `xml:"CardToken3DropOdds,attr"`
	CardToken3ExtraTimes int32 `xml:"CardToken3ExtractTimes,attr"`
	CardToken3Min        int32 `xml:"CardToken3Min,attr"`
	CardToken3Max        int32 `xml:"CardToken3Max,attr"`

	CardToken4DropOdds   int32 `xml:"CardToken4DropOdds,attr"`
	CardToken4ExtraTimes int32 `xml:"CardToken4ExtractTimes,attr"`
	CardToken4Min        int32 `xml:"CardToken4Min,attr"`
	CardToken4Max        int32 `xml:"CardToken4Max,attr"`

	ExtractingTimes int32 `xml:"ExtractingTimes,attr"`
	CardAmount      int32 `xml:"CardAmount,attr"`

	LegendaryExtractTimes string `xml:"LegendaryExtractTimes,attr"`
	LegendaryExtraOdds    int32
	LegendaryExtraLib     []*DropAmountInfo
	LegendaryAmount       string `xml:"LegendaryAmount,attr"`
	LegendaryAmountOdds   int32
	LegendaryAmountLib    []*DropAmountInfo
	LegendaryDropID       int32 `xml:"LegendaryDropID,attr"`

	EpicExtractTimes string `xml:"EpicExtractTimes,attr"`
	EpicExtraOdds    int32
	EpicExtraLib     []*DropAmountInfo
	EpicAmount       string `xml:"EpicAmount,attr"`
	EpicAmountOdds   int32
	EpicAmountLib    []*DropAmountInfo
	EpicDropID       int32 `xml:"EpicDropID,attr"`

	RareExtractTimes string `xml:"RareExtractTimes,attr"`
	RareExtraOdds    int32
	RareExtraLib     []*DropAmountInfo
	RareAmount       string `xml:"RareAmount,attr"`
	RareAmountOdds   int32
	RareAmountLib    []*DropAmountInfo
	RareDropID       int32 `xml:"RareDropID,attr"`

	CommonDropID int32 `xml:"CommonDropID,attr"`
}

type XmlChestConfig struct {
	Items []XmlChestItem `xml:"item"`
}

type QualityChestLib struct {
	TotolOdds  int32
	ChestCount int32
	Chests     []*XmlChestItem
}

type ArenaChestLib struct {
	qua2chestlib map[int32]int32
}

type CampChestDrop struct {
	arenalvl2chestlib map[int32]*ArenaChestLib
}

type ChestConfigMgr struct {
	Map   map[int32]*XmlChestItem
	Array []*XmlChestItem

	id2droppackage map[int32]*DropPackage
	camp2CampChest map[int32]*CampChestDrop

	max_chest_cycle int32
	count2chestqua  map[int32]int32
}

func (this *ChestConfigMgr) Init() bool {
	if !this.LoadChest() {
		return false
	}

	if !this.LoadChestLib() {
		return false
	}

	if !this.LoadChestCycle() {
		return false
	}

	log.Info("DoDrop Test")
	dropmap := make(map[int32]int32)
	this.DoDrop(6013, 1, 1, dropmap)
	for card, num := range dropmap {
		log.Info("== card[%d] num[%d]", card, num)
	}

	for camp, val := range this.camp2CampChest {
		for arena_lvl, chest_lib := range val.arenalvl2chestlib {
			log.Info("阵营【%d】 竞技等级 %d,宝箱掉落[%v]", camp, arena_lvl, chest_lib)
		}
	}

	log.Info("ChestCycle Test 2 [%d][%d][%d][%d]", this.GetDropChestId(1, 1, 2), this.GetDropChestId(1, 2, 3), this.GetDropChestId(2, 3, 1), this.GetDropChestId(2, 3, 4))

	chest_cfg := this.Map[6001]

	log.Info("6001 leg [%v] [%v] [%v]", chest_cfg.LegendaryExtraOdds, len(chest_cfg.LegendaryAmountLib), chest_cfg.LegendaryAmountLib[0])
	log.Info("6001 rare [%v] [%v] [%v] [%v]", chest_cfg.RareAmountOdds, len(chest_cfg.RareAmountLib), chest_cfg.RareAmountLib[0], chest_cfg.RareAmountLib[1])

	return true
}

func (this *ChestConfigMgr) LoadChest() bool {
	content, err := ioutil.ReadFile("../game_data/ChestConfig.xml")
	if nil != err {
		log.Error("ChestConfigMgr Load read file error(%s) !", err.Error())
		return false
	}

	tmp_cfg := &XmlChestConfig{}
	err = xml.Unmarshal(content, tmp_cfg)
	if nil != err {
		log.Error("ChestConfigMgr Load unmarshal error(%s) !", err.Error())
		return false
	}

	this.Map = make(map[int32]*XmlChestItem)
	this.Array = make([]*XmlChestItem, 0, len(tmp_cfg.Items))

	this.camp2CampChest = make(map[int32]*CampChestDrop)

	var tmp_odds, tmp_count int
	var tmp_drop *DropAmountInfo
	var group_arr []string
	var item_arr []string
	for idx := int32(0); idx < int32(len(tmp_cfg.Items)); idx++ {
		val := &tmp_cfg.Items[idx]
		if nil == val {
			continue
		}

		switch val.UnlockSecStr {
		case "WoodChestUnlockTime":
			{
				//val.UnlockSec = global_config.WoodChestUnlockTime
			}
		case "SilverChestUnlockTime":
			{
				//val.UnlockSec = global_config.SilverChestUnlockTime
			}
		case "GoldenChestUnlockTime":
			{
				//val.UnlockSec = global_config.GoldenChestUnlockTime
			}
		case "GiantChestUnlockTime":
			{
				//val.UnlockSec = global_config.GiantChestUnlockTime
			}
		case "MagicChestUnlockTime":
			{
				//val.UnlockSec = global_config.MagicChestUnlockTime
			}
		case "RareChestUnlockTime":
			{
				//val.UnlockSec = global_config.RareChestUnlockTime
			}
		case "EpicChestUnlockTime":
			{
				//val.UnlockSec = global_config.EpicChestUnlockTime
			}
		case "LegendryChestUnlockTime":
			{
				//val.UnlockSec = global_config.LegendryChestUnlockTime
			}
		}

		// 传奇卡片 =============================================================
		if val.LegendaryExtractTimes != "" {
			group_arr = strings.Split(val.LegendaryExtractTimes, "|")
		} else {
			group_arr = nil
		}
		tmp_len := int32(len(group_arr))
		if tmp_len > 0 {
			val.LegendaryExtraLib = make([]*DropAmountInfo, 0, tmp_len)
		}
		for j := int32(0); j < tmp_len; j++ {
			item_arr = strings.Split(group_arr[j], ",")
			if 2 != len(item_arr) {
				log.Error("ChestCfgMgr LoadChest Legendary extra left[%d:%s:%s] format error len(%d)", val.ChestID, val.LegendaryExtractTimes, group_arr[0], tmp_len)
				return false
			}

			tmp_count, err = strconv.Atoi(item_arr[0])
			if nil != err {
				log.Error("ChestCfgMgr failed to convert legendary extra left [%s]", item_arr[0])
				return false
			}
			tmp_odds, err = strconv.Atoi(item_arr[1])
			if nil != err {
				log.Error("ChestCfgMgr failed to convert legendary right [%s]", item_arr[1])
				return false
			}

			tmp_drop = &DropAmountInfo{Odds: int32(tmp_odds), Count: int32(tmp_count)}
			val.LegendaryExtraOdds = val.LegendaryExtraOdds + int32(tmp_odds)
			val.LegendaryExtraLib = append(val.LegendaryExtraLib, tmp_drop)
		}
		if val.LegendaryAmount != "" {
			group_arr = strings.Split(val.LegendaryAmount, "|")
		} else {
			group_arr = nil
		}
		tmp_len = int32(len(group_arr))
		if tmp_len > 0 {
			val.LegendaryAmountLib = make([]*DropAmountInfo, 0, tmp_len)
		}
		for j := int32(0); j < tmp_len; j++ {
			item_arr = strings.Split(group_arr[j], ",")
			if 2 != len(item_arr) {
				log.Error("ChestCfgMgr LoadChest Legendary amount left[%s] format error", group_arr[0])
				return false
			}

			tmp_count, err = strconv.Atoi(item_arr[0])
			if nil != err {
				log.Error("ChestCfgMgr failed to convert legendary left [%s]", item_arr[0])
				return false
			}
			tmp_odds, err = strconv.Atoi(item_arr[1])
			if nil != err {
				log.Error("ChestCfgMgr failed to convert legendary right [%s]", item_arr[1])
				return false
			}

			tmp_drop = &DropAmountInfo{Odds: int32(tmp_odds), Count: int32(tmp_count)}
			val.LegendaryAmountOdds = val.LegendaryAmountOdds + int32(tmp_odds)
			val.LegendaryAmountLib = append(val.LegendaryAmountLib, tmp_drop)
		}

		// 史诗卡片 =============================================================
		if val.EpicExtractTimes != "" {
			group_arr = strings.Split(val.EpicExtractTimes, "|")
		} else {
			group_arr = nil
		}
		tmp_len = int32(len(group_arr))
		if tmp_len > 0 {
			val.EpicExtraLib = make([]*DropAmountInfo, 0, tmp_len)
		}

		for j := int32(0); j < tmp_len; j++ {
			item_arr = strings.Split(group_arr[j], ",")
			if 2 != len(item_arr) {
				log.Error("ChestCfgMgr LoadChest Epic extra left[%d,%s,%s] format error len(%d)", val.ChestID, val.EpicExtractTimes, group_arr[j], len(item_arr))
				return false
			}

			tmp_count, err = strconv.Atoi(item_arr[0])
			if nil != err {
				log.Error("ChestCfgMgr failed to convert Epic extra left [%s]", item_arr[0])
				return false
			}
			tmp_odds, err = strconv.Atoi(item_arr[1])
			if nil != err {
				log.Error("ChestCfgMgr failed to convert Epic extra right [%s]", item_arr[1])
				return false
			}

			tmp_drop = &DropAmountInfo{Odds: int32(tmp_odds), Count: int32(tmp_count)}
			val.EpicExtraOdds = val.EpicExtraOdds + int32(tmp_odds)
			val.EpicExtraLib = append(val.EpicExtraLib, tmp_drop)
		}
		if val.EpicAmount != "" {
			group_arr = strings.Split(val.EpicAmount, "|")
		} else {
			group_arr = nil
		}
		tmp_len = int32(len(group_arr))
		if tmp_len > 0 {
			val.EpicAmountLib = make([]*DropAmountInfo, 0, tmp_len)
		}
		for j := int32(0); j < tmp_len; j++ {
			item_arr = strings.Split(group_arr[j], ",")
			if 2 != len(item_arr) {
				log.Error("ChestCfgMgr LoadChest Epic amount left[%d,%s,%s] format error", val.ChestID, val.EpicAmount, group_arr[j])
				return false
			}

			tmp_count, err = strconv.Atoi(item_arr[0])
			if nil != err {
				log.Error("ChestCfgMgr failed to convert Epic left [%s]", item_arr[0])
				return false
			}
			tmp_odds, err = strconv.Atoi(item_arr[1])
			if nil != err {
				log.Error("ChestCfgMgr failed to convert Epic right [%s]", item_arr[1])
				return false
			}

			tmp_drop = &DropAmountInfo{Odds: int32(tmp_odds), Count: int32(tmp_count)}
			val.EpicAmountOdds = val.EpicAmountOdds + int32(tmp_odds)
			val.EpicAmountLib = append(val.EpicAmountLib, tmp_drop)
		}

		// 稀有卡片 =============================================================
		if val.RareExtractTimes != "" {
			group_arr = strings.Split(val.RareExtractTimes, "|")
		} else {
			group_arr = nil
		}
		tmp_len = int32(len(group_arr))
		if tmp_len > 0 {
			val.RareExtraLib = make([]*DropAmountInfo, 0, tmp_len)
		}

		for j := int32(0); j < tmp_len; j++ {
			item_arr = strings.Split(group_arr[j], ",")
			if 2 != len(item_arr) {
				log.Error("ChestCfgMgr LoadChest Rare extra left[%s] format error", group_arr[0])
				return false
			}

			tmp_count, err = strconv.Atoi(item_arr[0])
			if nil != err {
				log.Error("ChestCfgMgr failed to convert Rare extra left [%s]", item_arr[0])
				return false
			}
			tmp_odds, err = strconv.Atoi(item_arr[1])
			if nil != err {
				log.Error("ChestCfgMgr failed to convert Rare extra right [%s]", item_arr[1])
				return false
			}

			tmp_drop = &DropAmountInfo{Odds: int32(tmp_odds), Count: int32(tmp_count)}
			val.RareExtraOdds = val.RareExtraOdds + int32(tmp_odds)
			val.RareExtraLib = append(val.RareExtraLib, tmp_drop)
		}
		if val.RareAmount != "" {
			group_arr = strings.Split(val.RareAmount, "|")
		} else {
			group_arr = nil
		}
		tmp_len = int32(len(group_arr))
		if tmp_len > 0 {
			val.RareAmountLib = make([]*DropAmountInfo, 0, tmp_len)
		}

		for j := int32(0); j < tmp_len; j++ {
			item_arr = strings.Split(group_arr[j], ",")
			if 2 != len(item_arr) {
				log.Error("ChestCfgMgr LoadChest Rare amount left[%s] format error", group_arr[0])
				return false
			}

			tmp_count, err = strconv.Atoi(item_arr[0])
			if nil != err {
				log.Error("ChestCfgMgr failed to convert Rare left [%s]", item_arr[0])
				return false
			}
			tmp_odds, err = strconv.Atoi(item_arr[1])
			if nil != err {
				log.Error("ChestCfgMgr failed to convert Rare right [%s]", item_arr[1])
				return false
			}

			tmp_drop = &DropAmountInfo{Odds: int32(tmp_odds), Count: int32(tmp_count)}
			val.RareAmountOdds = val.RareAmountOdds + int32(tmp_odds)
			val.RareAmountLib = append(val.RareAmountLib, tmp_drop)
		}

		// ====================================================================

		this.Map[val.ChestID] = val
		this.Array = append(this.Array, val)

		if nil == this.camp2CampChest[val.Camp] {
			this.camp2CampChest[val.Camp] = &CampChestDrop{}
			this.camp2CampChest[val.Camp].arenalvl2chestlib = make(map[int32]*ArenaChestLib)
		}

		if nil == this.camp2CampChest[val.Camp].arenalvl2chestlib[val.ArenaRating] {
			this.camp2CampChest[val.Camp].arenalvl2chestlib[val.ArenaRating] = &ArenaChestLib{}
			this.camp2CampChest[val.Camp].arenalvl2chestlib[val.ArenaRating].qua2chestlib = make(map[int32]int32)
		}

		this.camp2CampChest[val.Camp].arenalvl2chestlib[val.ArenaRating].qua2chestlib[val.ChestQuality] = val.ChestID
	}

	log.Info("6001 的开启时间[%d]", this.Map[6001].UnlockSec)

	return true
}

func (this *ChestConfigMgr) LoadChestLib() bool {
	content, err := ioutil.ReadFile("../game_data/DropID.xml")
	if nil != err {
		log.Error("ChestConfigMgr LoadChestLib readfile DropID failed err(%s) !", err.Error())
		return false
	}

	tmp_cfg := &XmlDropIdConfig{}
	err = xml.Unmarshal(content, tmp_cfg)
	if nil != err {
		log.Error("ChestConfigMgr LoadChestLib unmarshal failed err(%s)!", err.Error())
		return false
	}

	this.id2droppackage = make(map[int32]*DropPackage)
	for _, val := range tmp_cfg.Items {
		if nil == this.id2droppackage[val.DropID] {
			this.id2droppackage[val.DropID] = &DropPackage{}
		}

		this.id2droppackage[val.DropID].ItemCount++

		this.id2droppackage[val.DropID].TotalOdds = this.id2droppackage[val.DropID].TotalOdds + val.DropOdds

	}

	for i := int32(0); i < int32(len(tmp_cfg.Items)); i++ {
		val := &tmp_cfg.Items[i]
		if nil == this.id2droppackage[val.DropID].DropItems {
			this.id2droppackage[val.DropID].DropItems = make([]*XmlDropIdItem, 0, this.id2droppackage[val.DropID].ItemCount)
		}

		this.id2droppackage[val.DropID].DropItems = append(this.id2droppackage[val.DropID].DropItems, val)
	}

	return true
}

func (this *ChestConfigMgr) LoadChestCycle() bool {
	content, err := ioutil.ReadFile("../game_data/chestCycle.xml")
	if nil != err {
		log.Error("ChestConfigMgr LoadChestCycle readfile error(%s) !", err.Error())
		return false
	}

	tmp_cfg := &XmlChestCycleConfig{}
	err = xml.Unmarshal(content, tmp_cfg)
	if nil != err {
		log.Error("ChestConfigMgr LoadChestCycle xml unmarshal failed !")
		return false
	}

	this.count2chestqua = make(map[int32]int32)
	for count, val := range tmp_cfg.Items {
		if count > int(this.max_chest_cycle) {
			this.max_chest_cycle = int32(count)
		}

		this.count2chestqua[val.ChestOrder] = val.ChestQuelity
	}

	//log.Info("ChestConfigMgr LoadChestCycle after load %v", this.count2chestqua)

	return true
}

func (this *ChestConfigMgr) GetDropChestId(camp, arena_lvl, count int32) int32 {
	camp_lib := this.camp2CampChest[camp]
	if nil == camp_lib {
		log.Error("ChestConfigMgr GetDropChestId failed to find camplib[%d]", camp)
		return -1
	}

	arena_lib := camp_lib.arenalvl2chestlib[arena_lvl]
	if nil == arena_lib {
		log.Error("ChestConfigMgr GetDropChestId failed to find arena_lib[%d] !", arena_lvl)
		return -1
	}

	if count > this.max_chest_cycle {
		count = this.max_chest_cycle
	}

	qua := this.count2chestqua[count]
	return arena_lib.qua2chestlib[qua]
}

func (this *ChestConfigMgr) DoDrop(dropid, drop_count, card_count int32, drop_cards map[int32]int32) {
	if dropid <= 0 || card_count <= 0 || drop_count <= 0 || nil == drop_cards {
		log.Error("ChestConfigMgr DoDrop dropid[%d]<0, count[%d] <= 0 or drop_cards nil ", dropid, drop_count, card_count)
		return
	}
	pack := this.id2droppackage[dropid]
	if nil == pack {
		log.Error("ChestConfigMgr DoDrop no pack[%d] !", dropid)
		return
	}

	if pack.TotalOdds <= 0 {
		log.Error("ChestConfigMgr DoDrop[%d] totalodds <=0", dropid)
		return
	}

	if drop_count > int32(len(pack.DropItems)) {
		log.Error("ChestConfigMgr DoDrop over total ")
		return
	}

	total_odds := pack.TotalOdds
	tmp_count := pack.ItemCount
	log.Info("卡库", tmp_count, pack.DropItems)

	cur_switch := make(map[int32]int32, drop_count)
	var val *XmlDropIdItem
	for i := int32(0); i < drop_count; i++ {
		rand_val := rand.Int31n(total_odds)
		for j := int32(0); j < tmp_count; j++ {
			switch_i := cur_switch[j]
			if switch_i > 0 {
				switch_i--
				val = pack.DropItems[switch_i]
			} else {
				val = pack.DropItems[j]
			}

			if rand_val < val.DropOdds {
				log.Info("卡片随机", j, val.DropGoods, cur_switch)
				if i == drop_count-1 {
					drop_cards[val.DropGoods] = drop_cards[val.DropGoods] + card_count
					log.Info("最后一次", card_count)
				} else {
					rand_count := card_count - (drop_count - i) + 1
					if 0 <= rand_count {
						tmp_count := rand.Int31n(rand_count) + 1
						drop_cards[val.DropGoods] += tmp_count
						card_count -= tmp_count
					}
					log.Info("非最后一次", card_count, drop_count, i, rand_count)
				}

				cur_switch[j] = tmp_count
				total_odds = total_odds - val.DropOdds
				tmp_count--
				break
			} else {
				rand_val = rand_val - val.DropOdds
			}
		}
	}

	log.Info("ChestConfigMgr DoDrop dropid[%d] count[%d] cards", dropid, drop_count, card_count, drop_cards)
	return
}
