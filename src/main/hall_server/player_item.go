package main

import (
	"libs/log"
	_ "main/table_config"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	_ "sync"
	_ "time"

	_ "3p/code.google.com.protobuf/proto"
	_ "github.com/golang/protobuf/proto"
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
	ITEM_RESOURCE_ID_RMB          = 1  // 人民币
	ITEM_RESOURCE_ID_GOLD         = 2  // 金币
	ITEM_RESOURCE_ID_DIAMOND      = 3  // 钻石
	ITEM_RESOURCE_ID_STAMINA      = 5  // 体力
	ITEM_RESOURCE_ID_FRIEND_POINT = 6  // 友情值
	ITEM_RESOURCE_ID_CHARM_VALUE  = 7  // 魅力值
	ITEM_RESOURCE_ID_EXP_VALUE    = 8  // 经验值
	ITEM_RESOURCE_ID_TIME         = 11 // 时间
	ITEM_RESOURCE_ID_STAR         = 12 // 星数
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

func (this *Player) send_items() {
	msg := &msg_client_message.S2CItemsSync{}
	msg.Items = this.db.Items.BuildMsg()
	this.Send(uint16(msg_client_message_id.MSGID_S2C_ITEMS_SYNC), msg)
}

func (this *Player) Equip(role_id, equip_id int32) int32 {
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

func (this *Player) Unequip(role_id, equip_type int32) int32 {
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

/*
type ItemChangeInfo struct {
	items_update      map[int32]*msg_client_message.ItemInfo // 物品变化
	items_update_lock *sync.RWMutex                          // 物品变化锁
}

func (this *ItemChangeInfo) init() {
	this.items_update_lock = &sync.RWMutex{}
}

func (this *ItemChangeInfo) item_update(p *Player, item_id int32) {
	//this.items_update_lock.Lock()
	//defer this.items_update_lock.Unlock()

	if this.items_update == nil {
		this.items_update = make(map[int32]*msg_client_message.ItemInfo)
	}
	if this.items_update[item_id] == nil {
		this.items_update[item_id] = &msg_client_message.ItemInfo{}
	}

	this.items_update[item_id].ItemCfgId = proto.Int32(item_id)

	item := p.db.Items.Get(item_id)

	if item == nil {
		this.items_update[item_id].ItemNum = proto.Int32(0)
	} else {
		this.items_update[item_id].ItemNum = proto.Int32(item.ItemNum)
		this.items_update[item_id].RemainSeconds = proto.Int32(get_time_item_remain_seconds(item))
	}
}

// 计算计时物品剩余时间
func get_time_item_remain_seconds(item *dbPlayerItemData) int32 {
	if item.StartTimeUnix == 0 {
		return 0
	}

	now_time := int32(time.Now().Unix())
	cost_seconds := now_time - item.StartTimeUnix
	// 剩余时间小于等于3秒一律算到时
	left_seconds := item.RemainSeconds - cost_seconds
	if left_seconds <= 3 {
		return 0
	}
	return left_seconds
}

func (this *ItemChangeInfo) send_items_update(p *Player) bool {
	//this.items_update_lock.Lock()
	//defer this.items_update_lock.Unlock()

	if this.items_update == nil || len(this.items_update) == 0 {
		return false
	}

	msg := &msg_client_message.S2CItemsInfoUpdate{}
	msg.Items = make([]*msg_client_message.ItemInfo, len(this.items_update))
	i := int32(0)
	for _, v := range this.items_update {
		msg.Items[i] = v
		i += 1
	}

	p.Send(msg)

	this.items_update = nil

	return true
}

//////////////////////////////////////////////////////////////////////////////////
func (this *Player) SendItemsUpdate() {
	this.item_change_info.send_items_update(this)
}

// 体力增长计算
func (this *Player) CalcSpirit() int32 {
	curr_stamina := this.db.Info.GetSpirit()
	cp := cfg_player_level_mgr.Map[this.db.Info.GetLvl()]
	if cp == nil {
		return curr_stamina
	}

	last_save := this.db.Info.GetSaveLastSpiritPointTime()
	now := time.Now().Unix()
	used_seconds := int32(now) - last_save
	if curr_stamina < cp.MaxPower && used_seconds > global_id.SpiritGrowPointNeedMinute_44*60 {
		y := used_seconds % global_id.SpiritGrowPointNeedMinute_44
		grow_points := used_seconds / (global_id.SpiritGrowPointNeedMinute_44 * 60)
		if curr_stamina+grow_points > cp.MaxPower {
			grow_points = cp.MaxPower - curr_stamina
		}
		if grow_points > 0 {
			this.db.Info.IncbySpirit(grow_points)
			this.db.Info.SetSaveLastSpiritPointTime(int32(now) - y)
		}
	}
	return this.db.Info.GetSpirit()
}

func (this *Player) GetItemResourceValue(other_id int32) int32 {
	switch other_id {
	case ITEM_RESOURCE_ID_RMB:
		{
			return 0
		}
	case ITEM_RESOURCE_ID_GOLD:
		{
			return this.db.Info.GetCoin()
		}
	case ITEM_RESOURCE_ID_DIAMOND:
		{
			return this.db.Info.GetDiamond()
		}
	case ITEM_RESOURCE_ID_STAMINA:
		{
			// 体力要即时计算
			return this.CalcSpirit()
		}
	case ITEM_RESOURCE_ID_FRIEND_POINT:
		{
			return this.db.Info.GetFriendPoints()
		}
	case ITEM_RESOURCE_ID_CHARM_VALUE:
		{
			return this.db.Info.GetCharmVal()
		}
	case ITEM_RESOURCE_ID_EXP_VALUE:
		{
			return this.db.Info.GetExp()
		}
	default:
		{
			num, o := this.db.Items.GetItemNum(other_id)
			if !o {
				return 0
			}
			return num
		}
	}
}

func (this *Player) AddItemResource(cid, num int32, reason, mod string) int32 {
	switch cid {
	case ITEM_RESOURCE_ID_RMB:
		{
			log.Debug("rmb is not supported")
		}
	case ITEM_RESOURCE_ID_GOLD:
		{
			return this.AddCoin(num, reason, mod)
		}
	case ITEM_RESOURCE_ID_DIAMOND:
		{
			return this.AddDiamond(num, reason, mod)
		}
	case ITEM_RESOURCE_ID_CHARM_VALUE:
		{
			return this.AddCharmVal(num, reason, mod)
		}
	case ITEM_RESOURCE_ID_EXP_VALUE:
		{
			this.AddExp(num, reason, mod)
		}
	case ITEM_RESOURCE_ID_FRIEND_POINT:
		{
			return this.AddFriendPoints(num, reason, mod)
		}
	case ITEM_RESOURCE_ID_STAMINA:
		{
			return this.AddSpirit(num, reason, mod)
		}
	default:
		{
			if this.AddItem(cid, num, reason, mod, false) == nil {
				return -1
			}
		}
	}
	return 1
}

func (this *Player) RemoveItemResource(cid, num int32, reason, mod string) int32 {
	switch cid {
	case ITEM_RESOURCE_ID_RMB:
		{
			log.Debug("rmb is not supported")
		}
	case ITEM_RESOURCE_ID_GOLD:
		{
			return this.SubCoin(num, reason, mod)
		}
	case ITEM_RESOURCE_ID_DIAMOND:
		{
			return this.SubDiamond(num, reason, mod)
		}
	case ITEM_RESOURCE_ID_CHARM_VALUE:
		{
			return this.SubCharmVal(num, reason, mod)
		}
	case ITEM_RESOURCE_ID_EXP_VALUE:
		{

		}
	case ITEM_RESOURCE_ID_FRIEND_POINT:
		{
			return this.SubFriendPoints(num, reason, mod)
		}
	case ITEM_RESOURCE_ID_STAMINA:
		{
			return this.SubSpirit(num, reason, mod)
		}
	default:
		{
			if this.RemoveItem(cid, num, true) == nil {
				return -1
			}
		}
	}
	return 1
}

func (this *Player) ChkResEnough(resources []int32) bool {
	tmp_len := int32(len(resources))
	var item_type, item_val int32
	for idx := int32(0); idx < tmp_len; idx += 2 {
		item_type = resources[idx]
		item_val = resources[idx+1]
		if this.GetItemResourceValue(item_type) < item_val {
			return false
		}
	}

	return true
}

func (this *Player) RemoveResources(resources []int32, reason, mod string) {
	tmp_len := int32(len(resources))
	var item_type, item_val int32
	for idx := int32(0); idx < tmp_len; idx += 2 {
		item_type = resources[idx]
		item_val = resources[idx+1]
		this.RemoveItemResource(item_type, item_val, reason, mod)
	}

	return
}

func (this *Player) AddResources(resources []int32, reason, mod string) {
	if resources == nil {
		return
	}
	for i := 0; i < len(resources)/2; i++ {
		this.AddItemResource(resources[2*i], resources[2*i+1], reason, mod)
	}
}

// 各个属性设置函数

// 玩家添加猫或者物品或建筑库
func (this *Player) AddObj(objcfgid, addnum int32, reason, mod string, bslience bool) int32 {
	new_num := this.AddItemResource(objcfgid, addnum, reason, mod)
	if new_num >= 0 {
		return new_num
	}

	return new_num
}

// 玩家经验
func (this *Player) AddExp(add_val int32, reason, mod string) (int32, int32) {
	old_lvl := this.db.Info.GetLvl()
	if add_val < 0 {
		log.Error("Player AddExp add_val[%d] < 0 ", add_val)
		return old_lvl, this.db.Info.GetExp()
	}

	old_exp := this.db.Info.GetExp()
	cur_exp := old_exp + add_val
	if old_lvl < 1 {
		return -1, -1
	}
	cur_lvl := old_lvl
	if cur_lvl+1 <= cfg_player_level_mgr.MaxLevel {
		blvl_chg := false
		for i := cur_lvl; i < cfg_player_level_mgr.MaxLevel; i++ {
			next_exp := cfg_player_level_mgr.Array[i-1].MaxExp
			if cur_exp >= next_exp {
				cur_lvl = i + 1
				cur_exp = cur_exp - next_exp
				blvl_chg = true
			} else {
				break
			}
		}

		if blvl_chg {
			log.Info("玩家[%d] 升级了[%d]", this.Id, cur_lvl)
			this.db.Info.SetLvl(cur_lvl)
			this.db.Info.SetExp(cur_exp)
		} else {
			this.db.Info.SetExp(cur_exp)
		}
	}

	this.b_base_prop_chg = true

	if cur_lvl > old_lvl {
		this.UpdateNewTasks(cur_lvl, true)
		result := this.rpc_update_base_info()
		if result.Error < 0 {
			log.Warn("Player[%v] update base info error[%v]", this.Id, result.Error)
		}
	}
	return cur_lvl, cur_exp
}

// 玩家金币 ====================================

func (this *Player) GetCoin() int32 {
	return this.db.Info.GetCoin()
}

func (this *Player) AddCoin(val int32, reason, mod string) int32 {
	if val < 0 {
		log.Error("Player AddCoin %d", val)
		return this.db.Info.GetCoin()
	}

	if this.db.Info.GetCoin()+val < 0 {
		this.db.Info.SetCoin(0x7fffffff)
		return 0x7fffffff
	}

	cur_coin := this.db.Info.IncbyCoin(val)
	this.b_base_prop_chg = true
	return cur_coin
}

func (this *Player) SubCoin(val int32, reason, mod string) int32 {
	if val < 0 {
		log.Error("Player SubCoin %d", val)
		return this.db.Info.GetCoin()
	}

	cur_coin := this.db.Info.SubCoin(val)

	//this.TaskAchieveOnConditionAdd(TASK_ACHIEVE_FINISH_COIN_COST, val)

	this.b_base_prop_chg = true
	return cur_coin
}

// 玩家钻石 ====================================

func (this *Player) GetDiamond() int32 {
	return this.db.Info.GetDiamond()
}

func (this *Player) SubDiamond(sub_val int32, reason, mod string) int32 {
	if sub_val < 0 {
		log.Error("Player SubDiamond sub_val[%d] < 0, reason[%s] mod[%s]", sub_val, reason, mod)
		return this.db.Info.GetDiamond()
	}

	cur_diamond := this.db.Info.SubDiamond(sub_val)

	//this.TaskAchieveOnConditionAdd(TASK_ACHIEVE_FINISH_DIAMOND_COST, sub_val)

	this.b_base_prop_chg = true
	return cur_diamond
}

func (this *Player) AddDiamond(add_val int32, reason, mod string) int32 {
	if add_val < 0 {
		log.Error("Player AddDiamod add_val[%d] < 0, reason[%s] mod[%s]", add_val, reason, mod)
		return this.db.Info.GetDiamond()
	}

	if this.db.Info.GetDiamond()+add_val < 0 {
		this.db.Info.SetDiamond(0x7fffffff)
		return 0x7fffffff
	}

	this.b_base_prop_chg = true
	return this.db.Info.IncbyDiamond(add_val)
}

// 玩家魅力 =====================================

func (this *Player) SubCharmVal(sub_val int32, reason, mod string) int32 {
	if sub_val < 0 {
		log.Error("Player SubCharamVal sub_val(%d) < 0 reason(%s) mod(%s)", sub_val, reason, mod)
		return this.db.Info.GetCharmVal()
	}

	cur_charmval := this.db.Info.IncbyCharmVal(-sub_val)
	this.b_base_prop_chg = true

	// update ranking list
	if this.rpc_rank_update_charm(cur_charmval) == nil {
		log.Warn("Player[%v] update charm[%v] rank list failed", this.Id, cur_charmval)
	}

	return cur_charmval
}

func (this *Player) AddCharmVal(add_val int32, reason, mod string) int32 {
	if add_val < 0 {
		log.Error("Player AddCharmVal add_val(%d)< 0 reason(%s) mod(%s)", add_val, reason, mod)
		return this.db.Info.GetCharmVal()
	}

	if this.db.Info.GetCharmVal()+add_val < 0 {
		this.db.Info.SetCharmVal(0x7fffffff)
		return 0x7fffffff
	}

	cur_charmval := this.db.Info.IncbyCharmVal(add_val)
	this.b_base_prop_chg = true

	// update task
	this.TaskUpdate(table_config.TASK_FINISH_CHARM_VALUE, false, 0, cur_charmval)

	// update ranking list
	if this.rpc_rank_update_charm(cur_charmval) == nil {
		log.Warn("Player[%v] update charm[%v] rank list failed", this.Id, cur_charmval)
	}

	return cur_charmval
}

// 玩家友情点 ====================================
func (this *Player) SubFriendPoints(sub_val int32, reason, mod string) int32 {
	if sub_val < 0 {
		log.Error("Player SubFriendPoints sub_val(%v) < 0 reason(%s) mod(%s)", sub_val, reason, mod)
		return this.db.Info.GetFriendPoints()
	}

	cur_friendpoints := this.db.Info.IncbyFriendPoints(-sub_val)
	this.b_base_prop_chg = true
	return cur_friendpoints
}

func (this *Player) AddFriendPoints(add_val int32, reason, mod string) int32 {
	if add_val < 0 {
		log.Error("Player AddFriendPoints add_val(%d) < 0 reason(%s) mod(%s)", add_val, reason, mod)
		return this.db.Info.GetFriendPoints()
	}

	if this.db.Info.GetFriendPoints()+add_val < 0 {
		this.db.Info.SetFriendPoints(0x7fffffff)
		return 0x7fffffff
	}

	cur_friendpoints := this.db.Info.IncbyFriendPoints(add_val)
	this.b_base_prop_chg = true
	return cur_friendpoints
}

// 玩家体力 =====================================
func (this *Player) AddSpirit(spirit int32, reason, mod string) int32 {
	this.CalcSpirit()
	if spirit < 0 {
		log.Error("Player AddSpirit spirit(%v) < 0  reason(%v) mod(%s)", spirit, reason, mod)
		return this.db.Info.GetSpirit()
	}
	if this.db.Info.GetSpirit()+spirit < 0 {
		this.db.Info.SetSpirit(0x7fffffff)
		return 0x7fffffff
	}
	cur_spirit := this.db.Info.IncbySpirit(spirit)
	this.b_base_prop_chg = true
	return cur_spirit
}

func (this *Player) SubSpirit(spirit int32, reason, mod string) int32 {
	this.CalcSpirit()
	if spirit < 0 {
		log.Error("Player SubSpirit spirit(%v) < 0  reason(%v) mod(%s)", spirit, reason, mod)
		return this.db.Info.GetSpirit()
	}
	cur_spirit := this.db.Info.IncbySpirit(-spirit)
	this.b_base_prop_chg = true
	return cur_spirit
}

// 玩家星数
func (this *Player) AddStar(star int32, reason, mod string) int32 {
	if star < 0 {
		log.Error("Player AddStar star(%v) < 0  reason(%v) mod(%v)", star, reason, mod)
		return this.db.Info.GetTotalStars()
	}
	curr_star := this.db.Info.IncbyTotalStars(star)
	this.b_base_prop_chg = true

	// update task
	this.TaskUpdate(table_config.TASK_FINISH_COLLECT_STAR_NUM, false, 0, curr_star)

	return curr_star
}

func (this *Player) SubStar(star int32, reason, mod string) int32 {
	if star < 0 {
		log.Error("Player SubStar star(%v) < 0  reason(%v) mod(%v)", star, reason, mod)
		return this.db.Info.GetTotalStars()
	}
	curr_star := this.db.Info.IncbyTotalStars(-star)
	this.b_base_prop_chg = true
	return curr_star
}

// 玩家赞数
func (this *Player) AddZan(zan int32, reason, mod string) int32 {
	if zan < 0 {
		log.Error("Player AddZan zan(%v) < 0  reason(%v) mod(%v)", zan, reason, mod)
		return this.db.Info.GetZan()
	}
	if this.db.Info.GetZan()+zan < 0 {
		this.db.Info.SetZan(0x7fffffff)
		return 0x7fffffff
	}
	cur_zan := this.db.Info.IncbyZan(zan)
	this.b_base_prop_chg = true
	return cur_zan
}

func (this *Player) SubZan(zan int32, reason, mod string) int32 {
	if zan < 0 {
		log.Error("Player SubZan zan(%v) < 0  reason(%v) mod(%v)", zan, reason, mod)
		return this.db.Info.GetZan()
	}
	cur_zan := this.db.Info.IncbyZan(-zan)
	this.b_base_prop_chg = true
	return cur_zan
}

// 玩家物品
func (this *Player) AddItem(itemcfgid, addnum int32, reason, mod string, bslience bool) *msg_client_message.ItemInfo {
	itemcfg := item_table_mgr.Map[itemcfgid]
	if nil == itemcfg {
		log.Error("Player AddItem failed to find itemcfg[%v] reason[%v] mod[%v]", itemcfgid, reason, mod)
		return nil
	}

	new_num := this.db.Items.ChkAddItemByNum(itemcfgid, addnum)

	// 更新物品变化状态
	this.item_change_info.item_update(this, itemcfgid)

	res2cli := &msg_client_message.ItemInfo{}
	res2cli.ItemCfgId = proto.Int32(itemcfgid)
	res2cli.ItemNum = proto.Int32(new_num)
	if !bslience {
		this.Send(res2cli)
	}

	return res2cli
}

func (this *Player) RemoveItem(cfgid, num int32, bsilence bool) *msg_client_message.ItemInfo {
	item := item_table_mgr.Map[cfgid]
	if item == nil {
		log.Error("Not found item[%v] in config", cfgid)
		return nil
	}
	o, _ := this.db.Items.ChkRemoveItem(cfgid, num)
	if !o {
		log.Error("remove item[%v,%v] for player[%v] failed", cfgid, num, this.Id)
		return nil
	}

	// 更新物品变化状态
	this.item_change_info.item_update(this, cfgid)

	msg := &msg_client_message.ItemInfo{}
	msg.ItemCfgId = proto.Int32(cfgid)
	msg.ItemNum = proto.Int32(num)
	if !bsilence {
		this.Send(msg)
	}
	return msg
}

func (this *Player) RemoveItemAll(item_id int32, silence bool) {
	n, o := this.db.Items.GetItemNum(item_id)
	if !o {
		return
	}
	this.db.Items.Remove(item_id)

	// 更新物品变化状态
	this.item_change_info.item_update(this, item_id)

	msg := &msg_client_message.ItemInfo{}
	msg.ItemCfgId = proto.Int32(item_id)
	msg.ItemNum = proto.Int32(n)
	if !silence {
		this.Send(msg)
	}
}

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

func (this *Player) is_today_zan(player_id int32, now_time time.Time) bool {
	zan_time, o := this.db.Zans.GetZanTime(player_id)
	if !o {
		return false
	}

	tt := time.Unix(int64(zan_time), 0)

	if tt.Year() != now_time.Year() || tt.YearDay() != now_time.YearDay() {
		return false
	}

	return true
}

func (p *Player) zan_player(player_id int32) int32 {
	now_time := time.Now()
	o := p.db.Zans.HasIndex(player_id)
	if o {
		if p.is_today_zan(player_id, now_time) {
			log.Warn("Player[%v] zan player[%v] today yet", p.Id, player_id)
			return int32(msg_client_message.E_ERR_PLAYER_ALREADY_ZAN_TODAY)
		}
		p.db.Zans.IncbyZanNum(player_id, 1)
		p.db.Zans.SetZanTime(player_id, int32(now_time.Unix()))
	} else {
		d := &dbPlayerZanData{
			PlayerId: player_id,
			ZanTime:  int32(now_time.Unix()),
			ZanNum:   1,
		}
		p.db.Zans.Add(d)
	}
	return 1
}
*/
