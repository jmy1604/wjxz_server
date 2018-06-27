package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
	"math/rand"
)

const (
	SHOP_TYPE_NONE = iota
	SHOP_TYPE_SPECIAL
	SHOP_TYPE_FRIEND_POINTS
	SHOP_TYPE_CHARM_MEDAL
	SHOP_TYPE_RMB
	SHOP_TYPE_SOUL_STONE
)

type XmlShopItemItem struct {
	Id           int32  `xml:"GoodID,attr"`
	ShopId       int32  `xml:"ShopID,attr"`
	ItemStr      string `xml:"ItemList,attr"`
	Item         []int32
	BuyCostStr   string `xml:"BuyCost,attr"`
	BuyCost      []int32
	StockNum     int32 `xml:"StockNum,attr"`
	RandomWeight int32 `xml:"RandomWeight,attr"`
}

type XmlShopItemConfig struct {
	Items []*XmlShopItemItem `xml:"item"`
}

type ItemsShop struct {
	items        []*XmlShopItemItem
	total_weight int32
}

type ShopItemTableManager struct {
	items_map   map[int32]*XmlShopItemItem
	items_array []*XmlShopItemItem
	shops_map   map[int32]*ItemsShop
}

func (this *ShopItemTableManager) Init() bool {
	data, err := ioutil.ReadFile("../game_data/ShopItem.xml")
	if nil != err {
		log.Error("ShopItemTableManager Load read file err[%s] !", err.Error())
		return false
	}

	tmp_cfg := &XmlShopItemConfig{}
	err = xml.Unmarshal(data, tmp_cfg)
	if nil != err {
		log.Error("ShopItemTableManager Load xml Unmarshal failed error [%s] !", err.Error())
		return false
	}

	tmp_len := int32(len(tmp_cfg.Items))

	this.items_map = make(map[int32]*XmlShopItemItem)
	this.items_array = []*XmlShopItemItem{}
	this.shops_map = make(map[int32]*ItemsShop)
	for i := int32(0); i < tmp_len; i++ {
		c := tmp_cfg.Items[i]
		shop := this.shops_map[c.ShopId]
		if shop == nil {
			shop = &ItemsShop{}
			this.shops_map[c.ShopId] = shop
		}
		shop.items = append(shop.items, c)
		shop.total_weight = c.RandomWeight
	}

	log.Info("Shop table load items count(%v)", tmp_len)

	return true
}

func (this *ShopItemTableManager) GetItem(item_id int32) *XmlShopItemItem {
	return this.items_map[item_id]
}

func (this *ShopItemTableManager) GetItems() map[int32]*XmlShopItemItem {
	return this.items_map
}

func (this *ShopItemTableManager) RandomShopItem(shop_id int32) *XmlShopItemItem {
	shop := this.shops_map[shop_id]
	if shop == nil {
		return nil
	}

	if shop.total_weight <= 0 {
		return nil
	}

	r := rand.Int31n(shop.total_weight)
	for _, item := range shop.items {
		if r <= item.RandomWeight {
			return item
		}
	}
	return nil
}

func (this *ShopItemTableManager) GetItemsShop(shop_id int32) []*XmlShopItemItem {
	shop := this.shops_map[shop_id]
	if shop == nil {
		return nil
	}
	return shop.items
}
