package table_config

import (
	"encoding/xml"
	"io/ioutil"
	"libs/log"
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

type ShopItemTableManager struct {
	items_map   map[int32]*XmlShopItemItem
	items_array []*XmlShopItemItem
}

func (this *ShopItemTableManager) Init() bool {
	data, err := ioutil.ReadFile("../game_data/Shop.xml")
	if nil != err {
		log.Error("ShopItemTableManager Load read file err[%s] !", err.Error())
		return false
	}

	tmp_cfg := &XmlShopConfig{}
	err = xml.Unmarshal(data, tmp_cfg)
	if nil != err {
		log.Error("ShopItemTableManager Load xml Unmarshal failed error [%s] !", err.Error())
		return false
	}

	tmp_len := int32(len(tmp_cfg.Items))

	this.items_map = make(map[int32]*XmlShopItemItem)
	this.items_array = []*XmlShopItemItem{}
	for i := int32(0); i < tmp_len; i++ {
		//c := tmp_cfg.Items[i]

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
