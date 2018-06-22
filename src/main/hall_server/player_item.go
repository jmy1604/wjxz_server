package main

import (
	"libs/log"
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
	ITEM_TYPE_NONE     = iota
	ITEM_TYPE_RESOURCE = 1 // 资源类
	ITEM_TYPE_EQUIP    = 2 // 装备
	ITEM_TYPE_COST     = 3 // 消耗品
	ITEM_TYPE_PIECE    = 4 // 碎片
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
	EQUIP_TYPE_HEAD      = 1 // 头
	EQUIP_TYPE_WEAPON    = 2 // 武器
	EQUIP_TYPE_CHEST     = 3 // 胸
	EQUIP_TYPE_BOOT      = 4 // 鞋
	EQUIP_TYPE_LEFT_SLOT = 5 // 左槽
	EQUIP_TYPE_RELIC     = 6 // 神器
	EQUIP_TYPE_MAX       = 7 //
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
	if this.items_changed_info == nil {
		this.items_changed_info = make(map[int32]int32)
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

	if item_tdata.EquipType < 1 || item_tdata.EquipType >= EQUIP_TYPE_MAX {
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_TYPE_NOT_MATCH)
	}

	var equips []int32
	equips, o = this.db.Roles.GetEquip(role_id)
	if !o {
		return int32(msg_client_message.E_ERR_PLAYER_ROLE_NOT_FOUND)
	}

	if equips == nil || len(equips) == 0 {
		equips = make([]int32, EQUIP_TYPE_MAX)
		//this.db.Roles.SetEquip(role_id, equips)
	}

	if equips[item_tdata.EquipType] > 0 {
		this.add_item(equips[item_tdata.EquipType], 1)
	}
	equips[item_tdata.EquipType] = equip_id
	this.db.Roles.SetEquip(role_id, equips)
	this.del_item(equip_id, 1)
	this.roles_id_change_info.id_update(role_id)

	response := &msg_client_message.S2CItemEquipResponse{
		RoleId:    role_id,
		ItemId:    equip_id,
		EquipSlot: item_tdata.EquipType,
	}

	this.Send(uint16(msg_client_message_id.MSGID_S2C_ITEM_EQUIP_RESPONSE), response)

	log.Debug("Player[%v] equip role[%v] item[%v] on equip type[%v]", this.Id, role_id, equip_id, item_tdata.EquipType)

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
	this.roles_id_change_info.id_update(role_id)

	response := &msg_client_message.S2CItemUnequipResponse{
		RoleId:    role_id,
		EquipSlot: equip_type,
	}

	this.Send(uint16(msg_client_message_id.MSGID_S2C_ITEM_UNEQUIP_RESPONSE), response)

	log.Debug("Player[%v] unequip role[%v] equip type[%v]", this.Id, role_id, equip_type)

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

	if piece.ComposeType != 2 && piece.ComposeType != 1 {
		log.Error("Cant fusion item with piece[%v]", piece_id)
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_FUSION_FAILED)
	}

	if piece.ComposeNum*fusion_num > piece_num {
		log.Error("Player[%v] piece[%v] not enough to fusion", this.Id, piece_id)
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_COUNT_NOT_ENOUGH_TO_FUSION)
	}

	var items []*msg_client_message.ItemInfo
	for i := int32(0); i < fusion_num; i++ {
		o, item := this.drop_item_by_id(piece.ComposeDropID, true, true)
		if !o {
			log.Error("Player[%v] fusion item with piece[%v] failed", this.Id, piece_id)
			return int32(msg_client_message.E_ERR_PLAYER_ITEM_FUSION_FAILED)
		}
		if items != nil || len(items) > 0 {
			item_data := item_table_mgr.Get(item.ItemCfgId)
			j := 0
			for ; j < len(items); j++ {
				if item_data != nil && items[j].ItemCfgId == item.ItemCfgId {
					items[j].ItemNum += item.ItemNum
					break
				}
			}
			if j >= len(items) {
				items = append(items, item)
			}
		} else {
			items = []*msg_client_message.ItemInfo{item}
		}
	}

	this.del_item(piece_id, fusion_num*piece.ComposeNum)

	response := &msg_client_message.S2CItemFusionResponse{
		Items: items,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ITEM_FUSION_RESPONSE), response)

	log.Debug("Player[%v] fusioned items[%v] with piece[%v,%v]", this.Id, items, piece_id, fusion_num)

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
	if item.SellReward != nil {
		for i := 0; i < len(item.SellReward)/2; i++ {
			this.add_resource(item.SellReward[2*i], item_num*item.SellReward[2*i+1])
		}
	}

	response := &msg_client_message.S2CItemSellResponse{
		ItemId:  item_id,
		ItemNum: item_num,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ITEM_SELL_RESPONSE), response)

	log.Debug("Player[%v] sell item[%v,%v], get items[%v]", this.Id, item_id, item_num, item.SellReward)

	return 1
}

func (this *Player) item_upgrade(role_id, item_id, upgrade_type int32) int32 {
	item := item_table_mgr.Get(item_id)
	if item == nil {
		log.Error("Player[%v] upgrade role[%v] item[%v] with upgrade_type[%v] failed, because item[%v] table data not found", this.Id, role_id, item_id, upgrade_type, item_id)
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_TABLE_ID_NOT_FOUND)
	}
	if item.Type != ITEM_TYPE_EQUIP {
		log.Error("Player[%v] upgrade item[%v] invalid", this.Id, item_id)
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_UPGRADE_TYPE_INVALID)
	}

	// 左槽 右槽
	var equips []int32
	if item.EquipType == EQUIP_TYPE_LEFT_SLOT || item.EquipType == EQUIP_TYPE_RELIC {
		if !this.db.Roles.HasIndex(role_id) {
			log.Error("Player[%v] upgrade left slot equip[%v] failed, role[%v] not found", this.Id, item_id, role_id)
			return int32(msg_client_message.E_ERR_PLAYER_ROLE_NOT_FOUND)
		}
		equips, _ = this.db.Roles.GetEquip(role_id)
		if equips == nil || len(equips) < EQUIP_TYPE_MAX {
			log.Error("Player[%v] role[%v] no equips", this.Id, role_id)
			return -1
		}
		if equips[item.EquipType] != item_id {
			log.Error("Player[%v] equip pos[%v] no item[%v]", this.Id, item.EquipType, item_id)
			return int32(msg_client_message.E_ERR_PLAYER_ITEM_UPGRADE_TYPE_INVALID)
		}
	} else {
		if this.get_resource(item_id) < 1 {
			log.Error("Player[%v] upgrade item[%v] failed, item[%] not found", this.Id, item_id, item_id)
			return int32(msg_client_message.E_ERR_PLAYER_ITEM_NOT_FOUND)
		}
	}

	item_upgrade := item_upgrade_table_mgr.GetByItemId(item_id)
	if item_upgrade == nil {
		return int32(msg_client_message.E_ERR_PLAYER_ITEM_UPGRADE_DATA_NOT_FOUND)
	}

	// 检测消耗物品
	for i := 0; i < len(item_upgrade.ResCondtion)/2; i++ {
		res_id := item_upgrade.ResCondtion[2*i]
		res_num := item_upgrade.ResCondtion[2*i+1]
		if this.get_resource(res_id) < res_num {
			log.Error("Player[%v] upgrade item[%v] failed, res[%v] not enough", this.Id, item_id, res_id)
			return int32(msg_client_message.E_ERR_PLAYER_ITEM_UPGRADE_RES_NOT_ENOUGH)
		}
	}

	new_item_id := int32(0)
	if item.EquipType == EQUIP_TYPE_LEFT_SLOT || item.EquipType == EQUIP_TYPE_RELIC {
		if item.EquipType == EQUIP_TYPE_LEFT_SLOT {
			for {
				if item_upgrade.UpgradeType == upgrade_type {
					break
				}
				item_upgrade = item_upgrade.Next
				if item_upgrade == nil {
					break
				}
			}
			if item_upgrade == nil {
				log.Error("Player[%v] no upgrade table data for role[%v] item[%v] upgrade_type[%v]", this.Id, role_id, item_id, upgrade_type)
				return int32(msg_client_message.E_ERR_PLAYER_ITEM_UPGRADE_FAILED)
			}
		}

		o, new_item := this.drop_item_by_id(item_upgrade.ResultDropId, true, false)
		if !o {
			log.Error("Player[%v] upgrade item[%v] failed, drop error", this.Id, item_id)
			return int32(msg_client_message.E_ERR_PLAYER_ITEM_UPGRADE_FAILED)
		}
		equips[item.EquipType] = new_item.ItemCfgId
		new_item_id = new_item.ItemCfgId
	} else {
		o, new_item := this.drop_item_by_id(item_upgrade.ResultDropId, true, true)
		if !o {
			log.Error("Player[%v] upgrade item[%v] failed, drop error", this.Id, item_id)
			return int32(msg_client_message.E_ERR_PLAYER_ITEM_UPGRADE_FAILED)
		}
		this.add_resource(item_id, -1)
		new_item_id = new_item.ItemCfgId
	}

	// 消耗物品
	for i := 0; i < len(item_upgrade.ResCondtion)/2; i++ {
		res_id := item_upgrade.ResCondtion[2*i]
		res_num := item_upgrade.ResCondtion[2*i+1]
		this.add_resource(res_id, -res_num)
	}

	response := &msg_client_message.S2CItemUpgradeResponse{
		RoleId:    role_id,
		NewItemId: new_item_id,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ITEM_UPGRADE_RESPONSE), response)

	log.Debug("Player[%v] upgraded item[%v] to new item[%v]", this.Id, item_id, new_item_id)

	return 1
}

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

func C2SItemEquipHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SItemEquipRequest
	err := proto.Unmarshal(msg_data, &req)
	if nil != err {
		log.Error("Unmarshal msg failed err(%s)!", err.Error())
		return -1
	}
	return p.equip(req.GetRoleId(), req.GetItemId())
}

func C2SItemUnequipHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SItemUnequipRequest
	err := proto.Unmarshal(msg_data, &req)
	if nil != err {
		log.Error("Unmarshal msg failed err(%s)", err.Error())
		return -1
	}
	return p.unequip(req.GetRoleId(), req.GetEquipSlot())
}

func C2SItemUpgradeHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SItemUpgradeRequest
	err := proto.Unmarshal(msg_data, &req)
	if nil != err {
		log.Error("Unmarshal msg failed err(%s)", err.Error())
		return -1
	}
	return p.item_upgrade(req.GetRoleId(), req.GetItemId(), req.GetUpgradeType())
}
