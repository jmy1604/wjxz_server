package main

import (
	"libs/log"
	_ "main/table_config"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	_ "strings"
	"time"

	"github.com/golang/protobuf/proto"
)

func (this *Player) refresh_shop(shop_id int32) (res int32, items []*msg_client_message.ShopItem) {
	shop := shop_table_mgr.Get(shop_id)
	if shop == nil {
		log.Error("Shop %v data not found", shop_id)
		return -1, nil
	}

	if shop.ShopMaxSlot > 0 {
		for i := int32(0); i < shop.ShopMaxSlot; i++ {
			shop_item := shopitem_table_mgr.RandomShopItem(shop_id)
			if shop_item == nil {
				log.Error("Player[%v] random shop[%v] item failed", this.Id, shop_id)
				return -1, nil
			}
			si := &msg_client_message.ShopItem{
				ItemId: shop_item.Id,
				CostResource: &msg_client_message.ItemInfo{
					ItemCfgId: shop_item.Item[0],
					ItemNum:   shop_item.Item[1],
				},
				ItemNum: shop_item.StockNum,
			}
			items = append(items, si)
		}
	} else {
		items_shop := shopitem_table_mgr.GetItemsShop(shop_id)
		if items_shop == nil {
			log.Error("Shop[%v] cant found items", shop_id)
			return -1, nil
		}
		for _, item := range items_shop {
			si := &msg_client_message.ShopItem{
				ItemId: item.Id,
				CostResource: &msg_client_message.ItemInfo{
					ItemCfgId: item.Item[0],
					ItemNum:   item.Item[1],
				},
				ItemNum: item.StockNum,
			}
			items = append(items, si)
		}
	}

	res = 1
	return
}

func (this *Player) send_shop(shop_id int32, auto_refresh bool) (res int32, is_free bool) {
	shop_tdata := shop_table_mgr.Get(shop_id)
	if shop_tdata == nil {
		log.Error("Shop[%v] not found", shop_id)
		return -1, false
	}

	now_time := int32(time.Now().Unix())
	last_refresh, _ := this.db.Shops.GetLastFreeRefreshTime(shop_id)
	if last_refresh <= 0 {
		last_refresh = now_time
	}
	remain_free_refresh := shop_tdata.FreeRefreshTime - (now_time - last_refresh)
	if remain_free_refresh < 0 {
		// 免费
		if !auto_refresh {
			is_free = true
		}
		remain_free_refresh = 0
	}

	var shop_items []*msg_client_message.ShopItem
	item_ids := this.db.ShopItems.GetAllIndex()
	for _, item_id := range item_ids {
		shop_item_tdata := shopitem_table_mgr.GetItem(item_id)
		if shop_item_tdata == nil {
			log.Warn("Player[%v] shop[%v] item[%v] table data not found", this.Id, shop_id, item_id)
			continue
		}
		num, o := this.db.ShopItems.GetLeftNum(item_id)
		if !o {
			continue
		}

		shop_item := &msg_client_message.ShopItem{
			ItemId: item_id,
			CostResource: &msg_client_message.ItemInfo{
				ItemCfgId: shop_item_tdata.Item[0],
				ItemNum:   shop_item_tdata.Item[1],
			},
			ItemNum: num,
		}
		shop_items = append(shop_items, shop_item)
	}

	response := &msg_client_message.S2CShopDataResponse{
		ShopId: shop_id,
		NextFreeRefreshRemainSeconds: remain_free_refresh,
		Items: shop_items,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_SHOP_DATA_RESPONSE), response)

	log.Debug("Player[%v] send shop data: %v", response)

	res = 1
	return
}

func C2SShopDataHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SShopDataRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}
	res, _ := p.send_shop(req.GetShopId(), true)
	return res
}
