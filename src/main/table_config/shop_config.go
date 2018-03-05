package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
)

const (
	MAX_SHOP_LATTICE_COUNT = 10
)

type XmlNormalShopItem struct {
	SellChestID      int32 `xml:"SellChestID,attr"`
	Camp             int32 `xml:"Camp,attr"`
	ArenaRatingLimit int32 `xml:"ArenaRatingLimit,attr"`
	ShopType         int32 `xml:"ShopType,attr"`
	Lattice          int32 `xml:"Lattice,attr"`
	Odds             int32 `xml:"Odds,attr"`
}

type XmlNormalShopConfig struct {
	Items []XmlNormalShopItem `xml:"item"`
}

type XmlCardPriceItem struct {
	CardQuality  int32 `xml:"CardQuality,attr"`
	CardBuyTimes int32 `xml:"CardBuyTimes,attr"`
	CardPrice    int32 `xml:"CardPrice,attr"`
}

type XmlCardPriceConfig struct {
	Items []XmlCardPriceItem `xml:"item"`
}

type QualityCardPriceCfg struct {
	MaxCount    int32           // 最大累计购买次数
	Count2Price map[int32]int32 // 购买次数对应的价格
}

type XmlCardShopLibItem struct {
	SellCardID       int32 `xml:"SellCardID,attr"`
	Camp             int32 `xml:"Camp,attr"`
	ArenaRatingLimit int32 `xml:"ArenaRatingLimit,attr"`
	ShopType         int32 `xml:"ShopType,attr"`
	Lattice          int32 `xml:"Lattice,attr"`
	Odds             int32 `xml:"Odds,attr"`
}

type XmlCardShopLibConfig struct {
	Items []XmlCardShopLibItem `xml:"item"`
}

type ShopAreanLvlCardLib struct {
	card_count      int32 // 卡片数量
	total_odds      int32 // 总权重
	b_card_lib_init bool
	card_lib        []*XmlCardShopLibItem
}

type ShopLatticeCardLib struct {
	arena2arenacardlib map[int32]*ShopAreanLvlCardLib
}

type ShopCampCardLibCfg struct {
	pos_array   []int32
	pos2cardlib map[int32]*ShopLatticeCardLib
}

type CampShopChests struct {
	Id2ChestItem map[int32]*XmlNormalShopItem
}

type ShopConfigMgr struct {
	camp2shopchests map[int32]*CampShopChests

	qua2QuaCardPice  map[int32]*QualityCardPriceCfg
	camp2ShopCardlib map[int32]*ShopCampCardLibCfg
}

var shop_cfg_mgr ShopConfigMgr

func (this *ShopConfigMgr) Init() bool {
	if !this.LoadNormalShop() {
		return false
	}

	if !this.LoadCardPrice() {
		return false
	}

	if !this.LoadCardLib() {
		return false
	}

	return true
}

func (this *ShopConfigMgr) LoadNormalShop() bool {
	content, err := ioutil.ReadFile("../game_data/ShopConfig.xml")
	if nil != err {
		log.Error("ShopConfigMgr LoadNormalShop read file error !")
		return false
	}

	tmp_cfg := &XmlNormalShopConfig{}
	err = xml.Unmarshal(content, tmp_cfg)
	if nil != err {
		log.Error("ShopConfigMgr LoadNormalShop Unmarshal failed !")
		return false
	}

	this.camp2shopchests = make(map[int32]*CampShopChests)
	for idx := int32(0); idx < int32(len(tmp_cfg.Items)); idx++ {
		val := &tmp_cfg.Items[idx]
		if nil == val {
			continue
		}

		if nil == this.camp2shopchests[val.Camp] {
			this.camp2shopchests[val.Camp] = &CampShopChests{}
			this.camp2shopchests[val.Camp].Id2ChestItem = make(map[int32]*XmlNormalShopItem)
		}

		this.camp2shopchests[val.Camp].Id2ChestItem[val.SellChestID] = val
	}

	log.Info("宝箱信息 %v", this.GetChestCfgById(1, 6501))

	return true
}

func (this *ShopConfigMgr) LoadCardPrice() bool {
	content, err := ioutil.ReadFile("../game_data/CardPriceConfig.xml")
	if nil != err {
		log.Error("ShopConfigMgr LoadCardPrice read file error (%s)", err.Error())
		return false
	}

	tmp_cfg := &XmlCardPriceConfig{}
	err = xml.Unmarshal(content, tmp_cfg)
	if nil != err {
		log.Error("ShopConfigMgr LoadCardPrice unmarshal error (%s)", err.Error())
		return false
	}

	this.qua2QuaCardPice = make(map[int32]*QualityCardPriceCfg)
	for _, val := range tmp_cfg.Items {
		if nil == this.qua2QuaCardPice[val.CardQuality] {
			this.qua2QuaCardPice[val.CardQuality] = &QualityCardPriceCfg{}
			this.qua2QuaCardPice[val.CardQuality].Count2Price = make(map[int32]int32)
		}

		if val.CardBuyTimes > this.qua2QuaCardPice[val.CardQuality].MaxCount {
			this.qua2QuaCardPice[val.CardQuality].MaxCount = val.CardBuyTimes
		}

		this.qua2QuaCardPice[val.CardQuality].Count2Price[val.CardBuyTimes] = val.CardPrice
	}

	return true
}

func (this *ShopConfigMgr) LoadCardLib() bool {
	content, err := ioutil.ReadFile("../game_data/CardShopConfig.xml")
	if nil != err {
		log.Error("ShopConfigMgr LoadCardLib read file error(%s)", err.Error())
		return false
	}

	tmp_cfg := &XmlCardShopLibConfig{}
	err = xml.Unmarshal(content, tmp_cfg)
	if nil != err {
		log.Error("ShopConfigMgr LoadCardLib unmarshal error(%s)", err.Error())
		return false
	}

	this.camp2ShopCardlib = make(map[int32]*ShopCampCardLibCfg)
	for _, val := range tmp_cfg.Items {
		if nil == this.camp2ShopCardlib[val.Camp] {
			this.camp2ShopCardlib[val.Camp] = &ShopCampCardLibCfg{}
			this.camp2ShopCardlib[val.Camp].pos2cardlib = make(map[int32]*ShopLatticeCardLib)
			this.camp2ShopCardlib[val.Camp].pos_array = make([]int32, 0, MAX_SHOP_LATTICE_COUNT)
		}

		if nil == this.camp2ShopCardlib[val.Camp].pos2cardlib[val.Lattice] {
			this.camp2ShopCardlib[val.Camp].pos_array = append(this.camp2ShopCardlib[val.Camp].pos_array, val.Lattice)
			this.camp2ShopCardlib[val.Camp].pos2cardlib[val.Lattice] = &ShopLatticeCardLib{}
			this.camp2ShopCardlib[val.Camp].pos2cardlib[val.Lattice].arena2arenacardlib = make(map[int32]*ShopAreanLvlCardLib)
		}

		if nil == this.camp2ShopCardlib[val.Camp].pos2cardlib[val.Lattice].arena2arenacardlib[val.ArenaRatingLimit] {
			this.camp2ShopCardlib[val.Camp].pos2cardlib[val.Lattice].arena2arenacardlib[val.ArenaRatingLimit] = &ShopAreanLvlCardLib{}
		}
	}

	for _, val := range tmp_cfg.Items {
		for arena_lvl, arena_lib := range this.camp2ShopCardlib[val.Camp].pos2cardlib[val.Lattice].arena2arenacardlib {
			if val.ArenaRatingLimit <= arena_lvl {
				arena_lib.card_count++
				arena_lib.total_odds = arena_lib.total_odds + val.Odds
			}
		}
	}

	for idx := int32(0); idx < int32(len(tmp_cfg.Items)); idx++ {
		val := &tmp_cfg.Items[idx]
		if nil == val {
			continue
		}

		for arena_lvl, arena_lib := range this.camp2ShopCardlib[val.Camp].pos2cardlib[val.Lattice].arena2arenacardlib {
			if val.ArenaRatingLimit <= arena_lvl {
				if !arena_lib.b_card_lib_init {
					arena_lib.b_card_lib_init = true
					arena_lib.card_lib = make([]*XmlCardShopLibItem, 0, arena_lib.card_count)
				}

				arena_lib.card_lib = append(arena_lib.card_lib, val)
			}
		}
	}

	return true
}

func (this *ShopConfigMgr) GetChestCfgById(camp, chestid int32) *XmlNormalShopItem {
	camp_cfg := this.camp2shopchests[camp]
	if nil == camp_cfg {
		return nil
	}

	return camp_cfg.Id2ChestItem[chestid]
}

func (this *ShopConfigMgr) GetBuyCost(cur_count, qua, buy_count int32) (bret bool, total_price, real_buy_count int32) {
	qua_price := this.qua2QuaCardPice[qua]
	if nil == qua_price || cur_count >= qua_price.MaxCount {
		return false, 0, 0
	}

	log.Info("ShopConfigMgr GetBuyCost ", cur_count, qua, buy_count, qua_price.MaxCount)

	total_price = int32(0)
	tmp_count := int32(0)
	for i := int32(1); i <= buy_count; i++ {
		tmp_count = cur_count + i
		if tmp_count > qua_price.MaxCount {
			break
			//tmp_count = qua_price.MaxCount
		}

		real_buy_count++
		total_price += qua_price.Count2Price[tmp_count]
	}

	bret = true
	return
}
