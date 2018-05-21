package main

import (
	"libs/log"
	"main/table_config"
	"math/rand"
	"net/http"
	"public_message/gen_go/client_message"
	"time"

	_ "3p/code.google.com.protobuf/proto"
	"github.com/golang/protobuf/proto"
)

func (this *Player) drop_item_by_id(id int32, check_same bool, add bool) (bool, *msg_client_message.ItemInfo) {
	drop_lib := drop_table_mgr.Map[id]
	if nil == drop_lib {
		return false, nil
	}
	item := this.drop_item(drop_lib, check_same, add)
	return true, item
}

func (this *Player) drop_item(drop_lib *table_config.DropTypeLib, check_same, badd bool) (item *msg_client_message.ItemInfo) {
	log.Info("当前抽取库 总数目%d 总权重%d 详细%v", drop_lib.TotalCount, drop_lib.TotalWeight, drop_lib)

	if check_same {
		if this.used_drop_ids == nil || len(this.used_drop_ids) == int(drop_lib.TotalCount) {
			this.used_drop_ids = make(map[int32]int32)
		}
	}

	get_same := false
	rand_val := rand.Int31n(drop_lib.TotalWeight)
	var tmp_item *table_config.XmlDropItem
	for i := int32(0); i < drop_lib.TotalCount; i++ {
		tmp_item = drop_lib.DropItems[i]
		if nil == tmp_item {
			continue
		}

		if tmp_item.Weight > rand_val || get_same {
			if check_same {
				if _, o := this.used_drop_ids[tmp_item.DropItemID]; o {
					get_same = true
					if i == drop_lib.TotalCount-1 {
						i = 0
					}
					if len(this.used_drop_ids) == int(drop_lib.TotalCount) {
						this.used_drop_ids = make(map[int32]int32)
					}
					log.Debug("!!!!!!!!!!! !!!!!!!! total_count[%v]  used_drop_ids len[%v]  i[%v]", drop_lib.TotalCount, len(this.used_drop_ids), i)
					continue
				}
			}
			_, num := rand31n_from_range(tmp_item.Min, tmp_item.Max)
			if nil != item_table_mgr.Map[tmp_item.DropItemID] {
				if badd {
					if !this.add_item(tmp_item.DropItemID, num) {
						log.Error("Player[%v] rand dropid[%d] not item or cat or building or item resource", this.Id, tmp_item.DropItemID)
						continue
					}
				}
			} else {
				if badd {
					if !this.add_resource(tmp_item.DropItemID, num) {
						log.Error("Player[%v] rand dropid[%d] not item or cat or building or item resource", this.Id, tmp_item.DropItemID)
						continue
					}
				}
			}

			item = &msg_client_message.ItemInfo{ItemCfgId: tmp_item.DropItemID, ItemNum: num}
			if !badd && this.tmp_cache_items != nil {
				n := this.tmp_cache_items[item.ItemCfgId]
				this.tmp_cache_items[item.ItemCfgId] = n + item.ItemNum
			}

			if check_same {
				this.used_drop_ids[tmp_item.DropItemID] = tmp_item.Weight
			}
			break
		} else {
			rand_val -= tmp_item.Weight
		}
	}

	return
}

func (this *Player) DropItems(items_info []*table_config.ItemInfo, draw_count int32, badd bool) (bool, []*msg_client_message.ItemInfo) {
	total_drop_count := int32(0)
	for n := 0; n < len(items_info); n++ {
		total_drop_count += items_info[n].Num
	}

	items := make([]*msg_client_message.ItemInfo, 0, draw_count*total_drop_count)

	this.used_drop_ids = make(map[int32]int32)

	seed := time.Now().Unix() + time.Now().UnixNano()
	rand.Seed(seed + int64(rand.Int31n(100)))
	for count := int32(0); count < draw_count; count++ {
		for i := 0; i < len(items_info); i++ {
			for j := 0; j < int(items_info[i].Num); j++ {
				draw_lib := drop_table_mgr.Map[items_info[i].Id]
				if nil == draw_lib {
					log.Error("Player[%v] draw card not found draw lib[%v]", this.Id, items_info[i].Id)
					return false, nil
				}

				tmp_item := this.drop_item(draw_lib, true, badd)
				if tmp_item != nil {
					items = append(items, tmp_item)
				}
			}
		}
	}

	return true, items
}

func (this *Player) DropItems2(items_info []int32, badd bool) (bool, []*msg_client_message.ItemInfo) {
	total_drop_count := int32(0)
	for n := 0; n < len(items_info)/2; n++ {
		total_drop_count += items_info[2*n+1]
	}

	items := make([]*msg_client_message.ItemInfo, 0, total_drop_count)

	rand.Seed(time.Now().Unix() + time.Now().UnixNano())
	for i := 0; i < len(items_info)/2; i++ {
		drop_lib := drop_table_mgr.Map[items_info[2*i]]
		if nil == drop_lib {
			return false, nil
		}

		for j := 0; j < int(items_info[2*i+1]); j++ {
			tmp_item := this.drop_item(drop_lib, false, badd)
			if tmp_item != nil {
				items = append(items, tmp_item)
			}
		}
	}
	return true, items
}

func (this *Player) DrawCard(draw_type, draw_count int32) int32 {
	extract := extract_table_mgr.Get(draw_type)
	if extract == nil || extract.DropItems == nil {
		log.Error("Player[%v] draw id[%v] not found", this.Id, draw_type)
		return -1
	}

	/*num := this.GetItemResourceValue(extract.CostId)
	if num < extract.CostNum*draw_count {
		log.Error("Player[%v] draw card need item[%v] num[%v] not enough", this.Id, extract.CostId, extract.CostNum*draw_count)
		return int32(msg_client_message.E_ERR_ITEM_NUM_NOT_ENOUGH)
	}

	var b bool
	res2cli := &msg_client_message.S2CDrawResult{}

	// 首抽
	if !this.db.FirstDrawCards.HasIndex(draw_type) && (extract.FirstDropIds != nil && len(extract.FirstDropIds) > 0) {
		b, res2cli.Items = this.DropItems2(extract.FirstDropIds, true)
		if !b {
			log.Error("C2SDrawHandler failed to find draw_lib [%d]", draw_type)
			return int32(msg_client_message.E_ERR_DRAW_WRONG_DRAW_TYPE)
		}
		var d dbPlayerFirstDrawCardData
		d.Id = draw_type
		d.Drawed = 1
		this.db.FirstDrawCards.Add(&d)
		res2cli.IsFirst = proto.Bool(true)
	} else {
		b, res2cli.Items = this.DropItems(extract.DropItems, draw_count, true)
		if !b {
			log.Error("C2SDrawHandler failed to find draw_lib [%d]", draw_type)
			return int32(msg_client_message.E_ERR_DRAW_WRONG_DRAW_TYPE)
		}
		res2cli.IsFirst = proto.Bool(false)
	}

	this.RemoveItemResource(extract.CostId, extract.CostNum*draw_count, "draw_card", "draw")

	this.SendItemsUpdate()
	this.Send(res2cli)*/

	return 1
}

func C2SDrawHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	/*req := msg.(*msg_client_message.C2SDraw)
	if nil == req || nil == p {
		log.Error("C2SDrawHandler req or p nil [%v]", nil == p)
		return -1
	}

	draw_type := req.GetDrawType()
	draw_num := req.GetDrawCount()

	return p.DrawCard(draw_type, draw_num)*/
	return 1
}
