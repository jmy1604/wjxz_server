package main

import (
	"libs/log"
	_ "main/table_config"
	"math"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	_ "sync"
	_ "time"

	_ "3p/code.google.com.protobuf/proto"
	"github.com/golang/protobuf/proto"
)

// 物品类型
const (
	ITEM_TYPE_NONE              = iota
	ITEM_TYPE_RESOURCE          = 1  // 资源类
	ITEM_TYPE_DRAW              = 2  // 抽卡券
	ITEM_TYPE_STAMINA           = 8  // 体力道具
	ITEM_TYPE_GIFT              = 12 // 礼包
	ITEM_TYPE_DECORATION_RECIPE = 13 // 装饰物配方
)

// 其他属性
const (
	ITEM_RESOURCE_ID_GOLD        = 1  // 金币
	ITEM_RESOURCE_ID_DIAMOND     = 3  // 钻石
	ITEM_RESOURCE_ID_EXP         = 7  // 经验值
	ITEM_RESOURCE_ID_STAMINA     = 8  // 体力
	ITEM_RESOURCE_ID_FRIENDPOINT = 9  // 友情点
	ITEM_RESOURCE_ID_HEROCOIN    = 10 // 英雄币
)

// 装备类型
const (
	EQUIP_TYPE_WEAPON   = 1 // 武器
	EQUIP_TYPE_BODY     = 2 // 衣服
	EQUIP_TYPE_LEG      = 3 // 腿
	EQUIP_TYPE_ACCESORY = 4 // 饰品
	EQUIP_TYPE_GEM      = 5 // 宝石 不能卸
	EQUIP_TYPE_RELIC    = 6 // 神器 不能强化
	EQUIP_TYPE_MAX      = 7 //
)

func (this *dbPlayerItemColumn) BuildMsg() (items []*msg_client_message.ItemInfo) {
	this.m_row.m_lock.UnSafeRLock("dbPlayerItemColumn.BuildMsg")
	defer this.m_row.m_lock.UnSafeRUnlock()

	for _, v := range this.m_data {
		item := &msg_client_message.ItemInfo{
			ItemCfgId: v.Id,
			ItemNum:   v.Count,
		}
		items = append(items, item)
	}

	return
}

func (this *Player) add_item(id int32, count int32) bool {
	item := item_table_mgr.Get(id)
	if item == nil {
		log.Error("item %v not found in table", id)
		return false
	}

	if !this.db.Items.HasIndex(id) {
		this.db.Items.Add(&dbPlayerItemData{
			Id:    id,
			Count: count,
		})
	} else {
		this.db.Items.IncbyCount(id, count)
	}

	if this.items_changed_info == nil {
		this.items_changed_info = make(map[int32]int32)
	}
	if d, o := this.items_changed_info[id]; !o {
		this.items_changed_info[id] = count
	} else {
		this.items_changed_info[id] = d + count
	}

	return true
}

func (this *Player) del_item(id int32, count int32) bool {
	c, o := this.db.Items.GetCount(id)
	if !o {
		return false
	}

	if c < count {
		return false
	}

	if c == count {
		this.db.Items.Remove(id)
	} else {
		this.db.Items.IncbyCount(id, -count)
	}
	if d, o := this.items_changed_info[id]; !o {
		this.items_changed_info[id] = -count
	} else {
		this.items_changed_info[id] = d - count
	}
	return true
}

func (this *Player) get_item(id int32) int32 {
	c, _ := this.db.Items.GetCount(id)
	return c
}

func (this *Player) add_all_items() {
}

func (this *Player) send_items() {
	msg := &msg_client_message.S2CItemsSync{}
	msg.Items = this.db.Items.BuildMsg()
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ITEMS_SYNC), msg)
}

func (this *Player) add_gold(add int32) int32 {
	result := int32(0)
	gold := add + this.db.Info.GetGold()
	if gold >= 0 {
		if gold < add || gold < this.db.Info.GetGold() {
			this.db.Info.SetGold(math.MaxInt32)
			result = math.MaxInt32
		} else {
			result = this.db.Info.IncbyGold(add)
		}
	} else {
		if gold < 0 {
			this.db.Info.SetGold(0)
			result = 0
		} else {
			result = this.db.Info.IncbyGold(add)
		}
	}
	if add != 0 {
		this.b_base_prop_chg = true
	}
	return result
}

func (this *Player) get_gold() int32 {
	return this.db.Info.GetGold()
}

func (this *Player) add_diamond(add int32) int32 {
	result := int32(0)
	diamond := add + this.db.Info.GetDiamond()
	if diamond >= 0 {
		if diamond < add || diamond < this.db.Info.GetDiamond() {
			result = math.MaxInt32
			this.db.Info.SetDiamond(result)
		} else {
			result = this.db.Info.IncbyDiamond(add)
		}
	} else {
		if diamond < 0 {
			result = 0
			this.db.Info.SetDiamond(result)
		} else {
			result = this.db.Info.IncbyDiamond(add)
		}
	}
	if add != 0 {
		this.b_base_prop_chg = true
	}
	return result
}

func (this *Player) get_diamond() int32 {
	return this.db.Info.GetDiamond()
}

func (this *Player) add_exp(add int32) (level, exp int32) {
	if add < 0 {
		return
	}

	exp = add + this.db.Info.GetExp()
	if exp < add || exp < this.db.Info.GetExp() {
		exp = math.MaxInt32
	}

	max_level := 200
	level = this.db.Info.GetLvl()
	for {
		lvl_data := levelup_table_mgr.Get(level)
		if lvl_data == nil {
			break
		}
		if lvl_data.PlayerLevelUpExp > exp {
			break
		}
		exp -= lvl_data.PlayerLevelUpExp
		level += 1
		if int(level) > max_level {
			break
		}
	}

	if exp != this.db.Info.GetExp() {
		this.db.Info.SetExp(exp)
		this.b_base_prop_chg = true
	}
	if level != this.db.Info.GetLvl() {
		this.db.Info.SetLvl(level)
		this.b_base_prop_chg = true
	}

	return
}

func (this *Player) get_exp() int32 {
	return this.db.Info.GetExp()
}

func (this *Player) add_resource(id, count int32) bool {
	res := true
	if id == ITEM_RESOURCE_ID_GOLD {
		this.add_gold(count)
	} else if id == ITEM_RESOURCE_ID_DIAMOND {
		this.add_diamond(count)
	} else if id == ITEM_RESOURCE_ID_EXP {
		this.add_exp(count)
	} else {
		if !this.add_item(id, count) {
			res = false
		}
	}
	return res
}

func (this *Player) get_resource(id int32) int32 {
	if id == ITEM_RESOURCE_ID_GOLD {
		return this.get_gold()
	} else if id == ITEM_RESOURCE_ID_DIAMOND {
		return this.get_diamond()
	} else if id == ITEM_RESOURCE_ID_EXP {
		return this.get_exp()
	} else {
		return this.get_item(id)
	}
}

func (this *Player) equip(role_id, equip_id int32) int32 {
	var n int32
	var o bool
	if n, o = this.db.Items.GetCount(equip_id); !o {
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_NOT_FOUND)
	}

	if n <= 0 {
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_NUM_NOT_ENOUGH)
	}

	item_tdata := item_table_mgr.Get(equip_id)
	if item_tdata == nil {
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_TABLE_ID_NOT_FOUND)
	}

	if item_tdata.EquipType < EQUIP_TYPE_WEAPON || item_tdata.EquipType >= EQUIP_TYPE_MAX {
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_TYPE_NOT_MATCH)
	}

	var equips []int32
	equips, o = this.db.Roles.GetEquip(role_id)
	if !o {
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_NOT_FOUND)
	}

	if equips == nil || len(equips) == 0 {
		equips = make([]int32, EQUIP_TYPE_MAX)
		this.db.Roles.SetEquip(role_id, equips)
	}

	if equips[item_tdata.EquipType] > 0 {
		this.add_item(equips[item_tdata.EquipType], 1)
	}
	equips[item_tdata.EquipType] = equip_id
	this.del_item(equip_id, 1)

	return 1
}

func (this *Player) unequip(role_id, equip_type int32) int32 {
	equips, o := this.db.Roles.GetEquip(role_id)
	if !o {
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_NOT_FOUND)
	}
	if equips == nil || len(equips) == 0 {
		return int32(msg_client_message.E_ERR_PLAYER_EQUIP_SLOT_EMPTY)
	}

	if equip_type < EQUIP_TYPE_WEAPON || equip_type >= EQUIP_TYPE_MAX {
		return int32(msg_client_message.E_ERR_PLAYER_EQUIP_TYPE_INVALID)
	}

	if equips[equip_type] <= 0 {
		return int32(msg_client_message.E_ERR_PLAYER_EQUIP_SLOT_EMPTY)
	}

	this.add_item(equips[equip_type], 1)
	equips[equip_type] = 0
	this.db.Roles.SetEquip(role_id, equips)

	return 1
}

func (this *Player) fusion_item(piece_id int32, fusion_num int32) int32 {
	piece_num := this.get_item(piece_id)
	if piece_num <= 0 {
		log.Error("Player[%v] no piece[%v], cant fusion", this.Id, piece_id)
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_NOT_FOUND)
	}

	piece := item_table_mgr.Get(piece_id)
	if piece == nil {
		log.Error("Cant found item[%v] table data", piece_id)
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_TABLE_ID_NOT_FOUND)
	}

	if piece.ComposeType != 2 {
		log.Error("Cant fusion item with piece[%v]", piece_id)
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_FUSION_FAILED)
	}

	if piece.ComposeNum*fusion_num > piece_num {
		log.Error("Player[%v] piece[%v] not enough to fusion", this.Id, piece_id)
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_COUNT_NOT_ENOUGH_TO_FUSION)
	}

	o, item := this.drop_item_by_id(piece.ComposeDropID, true, true)
	if !o {
		log.Error("Player[%v] fusion item with piece[%v] failed", this.Id, piece_id)
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_FUSION_FAILED)
	}

	this.del_item(piece_id, fusion_num*piece.ComposeNum)

	response := &msg_client_message.S2CItemFusionResponse{
		Item: item,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ITEM_FUSION_RESPONSE), response)

	log.Debug("Player[%v] fusioned item[%v] with piece[%v,%v]", this.Id, item.ItemCfgId)

	return 1
}

func (this *Player) sell_item(item_id, item_num int32) int32 {
	item := item_table_mgr.Get(item_id)
	if item == nil {
		log.Error("Cant found item[%v] table data", item_id)
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_TABLE_ID_NOT_FOUND)
	}

	if this.get_item(item_id) < item_num {
		log.Error("Player[%v] item[%v] not enough", this.Id, item_id)
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_NUM_NOT_ENOUGH)
	}

	this.del_item(item_id, item_num)
	this.add_gold(item.SellPrice)

	response := &msg_client_message.S2CItemSellResponse{
		ItemId:  item_id,
		ItemNum: item_num,
		Gold:    item.SellPrice,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ITEM_SELL_RESPONSE), response)

	log.Debug("Player[%v] sell item[%v,%v], get gold[%v]", this.Id, item_id, item_num, item.SellPrice)

	return 1
}

/*
func (this *Player) add_handbook_data(item_id int32) {
	var d dbPlayerHandbookItemData
	d.Id = item_id
	this.db.HandbookItems.Add(&d)

	msg := &msg_client_message.S2CNewHandbookItemNotify{}
	msg.ItemId = proto.Int32(item_id)
	this.Send(msg)
}

func (this *Player) AddHandbookItem(item_id int32) {
	if handbook_table_mgr.Get(item_id) == nil {
		return
	}
	if this.db.HandbookItems.HasIndex(item_id) {
		return
	}

	this.add_handbook_data(item_id)
}

func (this *Player) AddHead(item_id int32) {
	if this.db.HeadItems.HasIndex(item_id) {
		return
	}
	var d dbPlayerHeadItemData
	d.Id = item_id
	this.db.HeadItems.Add(&d)
	msg := &msg_client_message.S2CNewHeadNotify{}
	msg.ItemId = proto.Int32(item_id)
	this.Send(msg)
}

func (this *Player) use_item(item_id int32, item_count int32) int32 {
	if item_count <= 0 {
		return -1
	}

	item := item_table_mgr.Map[item_id]
	if item == nil {
		log.Error("没有ID为%v的物品配置", item_id)
		return -1
	}

	num, o := this.db.Items.GetItemNum(item_id)
	if !o {
		log.Error("没有物品[%v]", item_id)
		return -1
	}

	if num < item_count {
		log.Error("物品[%v]数量[%v]不够", item_id, item_count)
		return -1
	}

	// 先判断是否为限时道具
	if item.ValidTime > 0 {
		item_data := this.db.Items.Get(item_id)
		if item_data != nil {
			if get_time_item_remain_seconds(item_data) == 0 {
				log.Error("玩家[%v]限时道具[%v]已过期", this.Id, item_id)
				return -1
			}
		}
	}

	// 体力道具
	if item.Type == ITEM_TYPE_STAMINA {
		if len(item.Numbers) < 2 {
			log.Error("物品[%v]数据配置错误", item_id)
			return -1
		}
		//this.AddSpirit(item.Numbers[1], "use_spirit_item", "use_item")
		this.RemoveItem(item_id, item_count, false)
		this.AddItemResource(item.Numbers[0], item.Numbers[1]*item_count, "use_spirit_item", "use_item")
	}

	// 发送物品变化
	this.item_change_info.send_items_update(this)

	msg := &msg_client_message.S2CUseItem{}
	msg.CostItem = &msg_client_message.ItemInfo{}
	msg.CostItem.ItemCfgId = proto.Int32(item_id)
	msg.CostItem.ItemNum = proto.Int32(item_count)
	this.Send(msg)

	return 1
}

func (this *Player) sell_item(item_id int32, item_count int32) int32 {
	item := item_table_mgr.Map[item_id]
	if item == nil {
		log.Error("没有ID为%v的物品", item_id)
		return -1
	}

	if this.RemoveItem(item_id, item_count, false) == nil {
		return -1
	}

	// 发送物品变化
	this.item_change_info.send_items_update(this)

	this.AddCoin(item.SaleCoin*item_count, "sell item", "item")

	msg := &msg_client_message.S2CSellItemResult{}
	msg.ItemId = proto.Int32(item_id)
	msg.ItemNum = proto.Int32(item_count)
	this.Send(msg)

	return 1
}

func (this *Player) ChkItemsEnough(itemidnums []int32) bool {
	tmp_len := int32(len(itemidnums))
	var item_id, item_num, db_num int32
	for idx := int32(0); idx < tmp_len; idx += 2 {
		item_id = itemidnums[idx]
		item_num = itemidnums[idx+1]
		db_num, _ = this.db.Items.GetItemNum(item_id)
		if db_num < item_num {
			return false
		}
	}
	return true
}

func (this *Player) RemoveItems(itemidnums []int32, reason, mod string) {
	tmp_len := int32(len(itemidnums))
	var item_id, item_num int32
	for idx := int32(0); idx < tmp_len; idx += 2 {
		item_id = itemidnums[idx]
		item_num = itemidnums[idx+1]
		this.RemoveItem(item_id, item_num, true)
	}
	return
}
*/

func C2SItemFusionHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SItemFusionRequest
	err := proto.Unmarshal(msg_data, &req)
	if nil != err {
		log.Error("Unmarshal msg failed err(%s) !", err.Error())
		return -1
	}
	return p.fusion_item(req.GetPieceId(), req.GetFusionNum())
}

func C2SItemSellHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SItemSellRequest
	err := proto.Unmarshal(msg_data, &req)
	if nil != err {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}
	return p.sell_item(req.GetItemId(), req.GetItemNum())
}
