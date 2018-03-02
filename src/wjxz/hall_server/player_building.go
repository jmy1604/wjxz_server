package main

import (
	"libs/log"
	"net/http"
	"public_message/gen_go/client_message"
	"time"
	"youma/table_config"

	"3p/code.google.com.protobuf/proto"
)

const (
	PLAYER_BUILDING_TAG_1 = 1

	PLAYER_BUILDING_UNLOCK_TYPE_P_LVL   = 1 // 玩家等级
	PLAYER_BUILDING_UNLOCK_TYPE_VIP_LVL = 2 // VIP等级
	PLAYER_BUILDING_UNLOCK_TYPE_FORMULA = 3 // 配方

	PLAYER_BUILDING_TYPE_FARMLAND = 1  // 农田
	PLAYER_BUILDING_TYPE_CAT_HOME = 2  // 猫舍
	PLAYER_BUILDING_TYPE_FOSTER   = 11 // 寄养所
)

func (this *Player) AddDepotBuilding(building_config_id int32, num int32, reason, mod string, send_msg bool) *msg_client_message.DepotBuildingInfo {
	build := cfg_building_mgr.Map[building_config_id]
	if build == nil {
		return nil
	}

	building := this.db.BuildingDepots.Get(building_config_id)
	if building == nil {
		if num <= 0 {
			return nil
		}
		var d dbPlayerBuildingDepotData
		d.CfgId = building_config_id
		d.Num = num
		this.db.BuildingDepots.Add(&d)
	} else {
		if building.Num+num < 0 {
			return nil
		}
		this.db.BuildingDepots.IncbyNum(building_config_id, num)
	}

	this.item_cat_building_change_info.depot_building_update(this, building_config_id)

	b := &msg_client_message.DepotBuildingInfo{}
	if send_msg {
		b.CfgId = proto.Int32(building_config_id)
		b.Num = proto.Int32(num)
		this.Send(b)
	}

	// 图鉴
	this.AddHandbookItem(building_config_id)

	// 公告
	if build != nil && build.Rarity >= 4 {
		anouncement_mgr.PushNew(ANOUNCEMENT_TYPE_GET_BUILDING, true, this.Id, building_config_id, 0, 0, "")
	}

	return b
}

func (this *Player) RemoveDepotBuilding(building_config_id int32, num int32, reason, mod string) bool {
	n, o := this.db.BuildingDepots.GetNum(building_config_id)
	if !o {
		return false
	}
	if n < num {
		return false
	}

	if n == num {
		this.db.BuildingDepots.Remove(building_config_id)
	} else {
		this.db.BuildingDepots.IncbyNum(building_config_id, -num)
	}
	this.item_cat_building_change_info.depot_building_update(this, building_config_id)
	return true
}

func (this *Player) IfCurMapNotInit() bool {

	if !this.b_cur_building_map_init {
		this.b_cur_building_map_init = true
		return true
	}

	return false
}

func (this *Player) ChkUpdateMyBuildingAreas() {
	this.b_cur_building_map_init_lock.Lock()
	defer this.b_cur_building_map_init_lock.Unlock()

	if this.IfCurMapNotInit() {
		//this.cur_area_use_count = make(map[int32]int32)
		this.cur_building_map, this.cur_areablocknum_map = this.db.Buildings.GetAllBuildingPos() // this.cur_area_use_count
		myidxs := this.db.Areas.GetAllIdxs()
		log.Info("==========当前开放区域 %v", myidxs)
		for _, tmp_idx := range myidxs {
			tmp_area := cfg_build_area_mgr.Getidx2area()[tmp_idx]
			if nil == tmp_area {
				continue
			}

			for tmp_xy, _ := range tmp_area.ArenaXYsMap {
				this.cur_open_pos_map[tmp_xy] = 1
				//log.Info("==========当前开放区域 坐标 %d (%d %d)", tmp_xy, tmp_xy>>16, tmp_xy&0x0000FFFF)
			}
		}

	}

	return
}

func (this *Player) set_area_building(pos_x, pos_y, width, height, building_id int32) {

	for tmp_x := int32(0); tmp_x < width; tmp_x++ {
		for tmp_y := int32(0); tmp_y < height; tmp_y++ {
			arena_xy := (pos_x+tmp_x)<<16 | (pos_y+tmp_y)&0x0000FFFF
			this.cur_building_map[arena_xy] = building_id
		}
	}
}

func (this *Player) remove_area_building(pos_x, pos_y, width, height int32) {
	for tmp_x := int32(0); tmp_x < width; tmp_x++ {
		for tmp_y := int32(0); tmp_y < height; tmp_y++ {
			arena_xy := (pos_x+tmp_x)<<16 | (pos_y+tmp_y)&0x0000FFFF
			delete(this.cur_building_map, arena_xy)
		}
	}
}

func (this *Player) RemoveAreaBuilding(building_id, pos_x, pos_y, width, height int32) {
	this.cur_area_map_lock.Lock()
	defer this.cur_area_map_lock.Unlock()

	for tmp_x := int32(0); tmp_x < width; tmp_x++ {
		for tmp_y := int32(0); tmp_y < height; tmp_y++ {
			arena_xy := (pos_x+tmp_x)<<16 | (pos_y+tmp_y)&0x0000FFFF
			delete(this.cur_building_map, arena_xy)
		}
	}

	this.db.Buildings.Remove(building_id)
}

func (this *Player) if_pos_can_set_building(x, y, width, height, extra_id, area_type int32) int32 {
	//myidxs := this.db.Areas.GetAllIdxs()
	//log.Info("ChkIfPosCanSetBuilding %d %d %d %d %d %d [%v]", x, y, width, height, extra_id, area_type, myidxs)

	var cur_building_id int32
	for tmp_x := int32(0); tmp_x < width; tmp_x++ {
		for tmp_y := int32(0); tmp_y < height; tmp_y++ {
			arena_xy := ((x + tmp_x) << 16) | (y+tmp_y)&0x0000FFFF
			if cfg_build_area_mgr.AreaXY2Type[arena_xy]&area_type <= 0 {
				//log.Info("ChkIfPosCanSetBuilding 位置%d[%d+%d, %d+%d]的地理类型不匹配 %d %d", arena_xy, x, tmp_x, y, tmp_y, cfg_build_area_mgr.AreaXY2Type[arena_xy], area_type)
				return int32(msg_client_message.E_ERR_BUILDING_AREA_TYPE_NOT_MATCH)
			}

			cur_building_id = this.cur_building_map[arena_xy]
			if cur_building_id > 0 && cur_building_id != extra_id {
				//log.Info("ChkIfPosCanSetBuilding 位置%d[%d+%d, %d+%d]上已经有建筑 %d", arena_xy, x, tmp_x, y, tmp_y, cur_building_id)
				return int32(msg_client_message.E_ERR_BUILDING_POS_FORBIDEN)
			}

			if 1 != this.cur_open_pos_map[arena_xy] {
				log.Info("ChkIfPosCanSetBuilding 没有在当前开放区域中找到位置%d[%d+%d, %d+%d]", arena_xy, x, tmp_x, y, tmp_y)
				return int32(msg_client_message.E_ERR_BUILDING_AREA_TYPE_NO_POS)
			}
		}
	}

	return 0
}

func (this *Player) TrySetMapBuildingDefDir(cfgid int32) *dbPlayerBuildingData {

	building_cfg := cfg_building_mgr.Map[cfgid]
	if nil == building_cfg {
		log.Error("Player TrySetBuilding failed to find building_cfg !")
		return nil
	}

	this.cur_area_map_lock.Lock()
	defer this.cur_area_map_lock.Unlock()

	var pos_x, pos_y int32
	var pos_y16 int16
	var iret int32
	for arena_xy, _ := range this.cur_open_pos_map {
		if this.IfXYAreaBlockFull(arena_xy) {
			continue
		}

		pos_x = arena_xy >> 16
		pos_y16 = int16(arena_xy)
		pos_y = int32(pos_y16)
		iret = this.if_pos_can_set_building(pos_x, pos_y, building_cfg.MapSizes[0], building_cfg.MapSizes[1], 0, building_cfg.Geography)
		if iret >= 0 {
			new_building_db := &dbPlayerBuildingData{}
			new_building_db.Id = this.db.Info.IncbyNextBuildingId(1)
			new_building_db.CfgId = cfgid
			new_building_db.X = pos_x
			new_building_db.Y = int32(pos_y)
			new_building_db.Dir = table_config.BUILDING_DIR_BIG_X_DIR
			new_building_db.CreateUnix = int32(time.Now().Unix())

			this.db.Buildings.Add(new_building_db)
			this.set_area_building(pos_x, pos_y, building_cfg.MapSizes[0], building_cfg.MapSizes[1], new_building_db.Id)
			return new_building_db
		}
	}

	return nil
}

func (this *Player) SetMapBuilding(cfgid, pos_x, pos_y, dir int32, is_buy int32, bslience bool) int32 {
	building_cfg := cfg_building_mgr.Map[cfgid]
	if nil == building_cfg {
		log.Error("Player TrySetBuilding failed to find building_cfg !")
		return 0
	}

	this.cur_area_map_lock.Lock()
	defer this.cur_area_map_lock.Unlock()

	var width, height int32
	if table_config.BUILDING_DIR_BIG_X_DIR == dir {
		width, height = building_cfg.MapSizes[0], building_cfg.MapSizes[1]
	} else {
		width, height = building_cfg.MapSizes[1], building_cfg.MapSizes[0]
	}

	log.Info("设置地图建筑[%d, %v] 当前开放区域 %v", cfgid, building_cfg.MapSizes, this.db.Areas.GetAllIdxs())
	iret := this.if_pos_can_set_building(pos_x, pos_y, width, height, 0, building_cfg.Geography)
	if iret >= 0 {
		is_done := true
		done := int32(1)
		if is_buy > 0 && building_cfg.BuildTime > 0 {
			is_done = false
			done = 0
		}

		new_building_db := &dbPlayerBuildingData{}
		new_building_db.Id = this.db.Info.IncbyNextBuildingId(1)
		new_building_db.CfgId = cfgid
		new_building_db.X = pos_x
		new_building_db.Y = pos_y
		new_building_db.Dir = table_config.BUILDING_DIR_BIG_X_DIR
		new_building_db.CreateUnix = int32(time.Now().Unix())
		new_building_db.IsDone = done

		this.db.Buildings.Add(new_building_db)

		this.set_area_building(pos_x, pos_y, building_cfg.MapSizes[0], building_cfg.MapSizes[1], new_building_db.Id)
		if !bslience {
			res2cli := &msg_client_message.S2CSetBuilding{}
			res2cli.BuildingCfgId = proto.Int32(cfgid)
			res2cli.Dir = proto.Int32(table_config.BUILDING_DIR_BIG_X_DIR)
			res2cli.X = proto.Int32(pos_x)
			res2cli.Y = proto.Int32(pos_y)
			res2cli.LeftBuildTime = proto.Int32(building_cfg.BuildTime)
			res2cli.IsDone = proto.Bool(is_done)
		}

		return new_building_db.Id
	} else {
		return iret
	}

	return 0
}

// 成功返回建筑配置Id, 失败返回小于零的错误码
func (this *Player) MoveMapBuilding(building_id, x, y, dir int32) int32 {
	building_db := this.db.Buildings.Get(building_id)
	if nil == building_db {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_EXIST)
	}

	building_cfg := cfg_building_mgr.Map[building_db.CfgId]
	if nil == building_cfg {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_CFG)
	}

	var old_width, old_height, new_width, new_height int32
	if table_config.BUILDING_DIR_BIG_X_DIR == building_db.Dir {
		old_width, old_height = building_cfg.MapSizes[0], building_cfg.MapSizes[1]
	} else {
		old_width, old_height = building_cfg.MapSizes[1], building_cfg.MapSizes[0]
	}
	if table_config.BUILDING_DIR_BIG_X_DIR == dir {
		new_width, new_height = building_cfg.MapSizes[0], building_cfg.MapSizes[1]
	} else {
		new_width, new_height = building_cfg.MapSizes[1], building_cfg.MapSizes[0]
	}

	iret := this.RelocateMapBuilding(building_cfg, building_db.Id, building_db.X, building_db.Y, old_width, old_height, x, y, new_width, new_height, dir)
	if iret < 0 {
		return iret
	}
	/*
		iret := this.if_pos_can_set_building(x, y, width, height, building_id, building_cfg.Geography)
		if iret < 0 {
			return iret
		}

		this.remove_area_building(building_db.X, building_db.Y, width, height)
		this.set_area_building(x, y, width, height, building_db.Id)
		this.db.Buildings.SetX(building_id, x)
		this.db.Buildings.SetY(building_id, y)
	*/

	return building_db.CfgId
}

func (this *Player) RelocateMapBuilding(building_cfg *table_config.XmlBuildingItem, building_id, old_x, old_y, old_w, old_h, new_x, new_y, new_w, new_h, new_dir int32) int32 {

	if nil == building_cfg {
		log.Error("Player RelocateMapBuilding building_cfg nil !")
		return -1
	}

	this.cur_area_map_lock.Lock()
	defer this.cur_area_map_lock.Unlock()

	iret := this.if_pos_can_set_building(new_x, new_y, new_w, new_h, building_id, building_cfg.Geography)
	if iret < 0 {
		return iret
	}

	this.remove_area_building(old_x, old_y, old_w, old_h)
	this.set_area_building(new_x, new_y, new_w, new_h, building_id)
	this.db.Buildings.SetX(building_id, new_x)
	this.db.Buildings.SetY(building_id, new_y)
	this.db.Buildings.SetDir(building_id, new_dir)

	return building_cfg.Id

}

// 成功返回建筑配置Id, 失败返回小于零的错误码
func (this *Player) ChgMapBuildingDir(building_id int32) int32 {
	building_db := this.db.Buildings.Get(building_id)
	if nil == building_db {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_EXIST)
	}

	building_cfg := cfg_building_mgr.Map[building_db.CfgId]
	if nil == building_cfg {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_CFG)
	}

	var width, height, new_dir int32
	if table_config.BUILDING_DIR_BIG_X_DIR == building_db.Dir {
		width, height = building_cfg.MapSizes[0], building_cfg.MapSizes[1]
		new_dir = table_config.BUILDING_DIR_BIG_Y_DIR
	} else {
		width, height = building_cfg.MapSizes[1], building_cfg.MapSizes[0]
		new_dir = table_config.BUILDING_DIR_BIG_X_DIR
	}

	iret := this.RelocateMapBuilding(building_cfg, building_db.Id, building_db.X, building_db.Y, width, height, building_db.X, building_db.Y, height, width, new_dir)
	if iret < 0 {
		return iret
	}
	/*
		iret := this.if_pos_can_set_building(building_db.X, building_db.Y, height, width, building_id, building_cfg.Geography)
		if iret < 0 {
			return iret
		}

		this.remove_area_building(building_db.X, building_db.Y, width, height)
		this.set_area_building(building_db.X, building_db.Y, height, width, building_db.Id)
	*/

	return building_db.CfgId
}

// 成功返回建筑配置Id，失败返回小于零的错误码
func (this *Player) ReomveMapBlock(block_id int32) (int32, *msg_client_message.S2CRemoveBlock) {
	block_db := this.db.Buildings.Get(block_id)
	if nil == block_db {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_EXIST), nil
	}

	building_cfg := cfg_building_mgr.Map[block_db.CfgId]
	if nil == building_cfg {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_CFG), nil
	}

	block_cfg := cfg_block_mgr.Map[block_db.CfgId]
	if nil == block_cfg {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_CFG), nil
	}

	log.Info("Player ReomveMapBlock 检查[%d]删除物品[%v]是否足够", block_cfg.Id, block_cfg.RemoveItems)
	if !this.ChkItemsEnough(block_cfg.RemoveItems) {
		return int32(msg_client_message.E_ERR_BUILDING_REMOVE_LESS_ITEM), nil
	}

	this.RemoveItems(block_cfg.RemoveItems, "open_mapchest", "building")

	var width, height int32
	if table_config.BUILDING_DIR_BIG_X_DIR == block_db.Dir {
		width, height = building_cfg.MapSizes[0], building_cfg.MapSizes[1]
	} else {
		width, height = building_cfg.MapSizes[1], building_cfg.MapSizes[0]
	}

	area_id := cfg_build_area_mgr.AreaXY2AreaId[(block_db.X<<16)|(block_db.Y&0X0000FFFF)]
	if area_id > 0 {
		if this.cur_areablocknum_map[area_id] > 0 {
			this.cur_areablocknum_map[area_id] = this.cur_areablocknum_map[area_id] - 1
		}
	}

	this.ChkUpdateMyBuildingAreas()
	this.RemoveAreaBuilding(block_db.Id, block_db.X, block_db.Y, width, height)
	//this.db.Buildings.Remove(block_id)

	res2cli := &msg_client_message.S2CRemoveBlock{}
	res2cli.BuildingId = proto.Int32(block_id)
	var bret bool
	bret, res2cli.Items, res2cli.Cats, res2cli.DepotBuildings = this.DropItems2(block_cfg.DropIds, true)
	if !bret {
		res2cli = nil
	}

	return block_db.CfgId, res2cli
}

func (this *Player) open_chest_result(chest_cfg_id int32) *msg_client_message.S2COpenMapChest {
	chest_cfg := cfg_mapchest_mgr.Map[chest_cfg_id]
	if nil == chest_cfg {
		log.Error("Player OpenMapChest no chest_cfg[%d] !", chest_cfg_id)
		return nil
	}
	this.AddExp(chest_cfg.Exp, "mapchest", "building")

	//this.db.Buildings.Remove(chest_id)

	res2cli := &msg_client_message.S2COpenMapChest{}

	var bret bool
	bret, res2cli.Items, res2cli.Cats, res2cli.DepotBuildings = this.DropItems2(chest_cfg.Rewards, true)
	if !bret {
		log.Error("Player[%v] open chest[%v] result by DropItem2[%v] failed", this.Id, chest_cfg_id, chest_cfg.Rewards)
		res2cli = nil
	}
	if res2cli.Items != nil {
		this.SendItemsUpdate()
	}
	if res2cli.Cats != nil {
		this.SendCatsUpdate()
	}
	if res2cli.DepotBuildings != nil {
		this.SendDepotBuildingUpdate()
	}
	return res2cli
}

func (this *Player) open_chest_cost(chest_cfg_id int32) int32 {
	chest_cfg := cfg_mapchest_mgr.Map[chest_cfg_id]
	if nil == chest_cfg {
		log.Error("Player OpenMapChest no chest_cfg[%d] !", chest_cfg_id)
		return int32(msg_client_message.E_ERR_BUILDING_NOT_CFG)
	}

	log.Info("Player OpenMapChest 检查[%d]删除物品[%v]是否足够", chest_cfg.Id, chest_cfg.RemoveCost)
	if !this.ChkResEnough(chest_cfg.RemoveCost) {
		return int32(msg_client_message.E_ERR_BUILDING_OPEN_MAP_CHEST_LESS_RES)
	}

	this.RemoveResources(chest_cfg.RemoveCost, "openmapchest", "Building")

	return 1
}

func (this *Player) return_chest_cost(chest_cfg_id int32) int32 {
	chest_cfg := cfg_mapchest_mgr.Map[chest_cfg_id]
	if nil == chest_cfg {
		log.Error("Player OpenMapChest no chest_cfg[%d] !", chest_cfg_id)
		return int32(msg_client_message.E_ERR_BUILDING_NOT_CFG)
	}
	this.AddResources(chest_cfg.RemoveCost, "return_chest_cost", "building")
	return 1
}

func (this *Player) return_chest_cost_by_id(chest_id int32) int32 {
	cfg_id, o := this.db.Buildings.GetCfgId(chest_id)
	if !o {
		return -1
	}
	return this.return_chest_cost(cfg_id)
}

// 成功返回建筑配置Id，失败返回小于零的错误码
func (this *Player) OpenMapChest(chest_id int32, is_add bool) (int32, *msg_client_message.S2COpenMapChest) {
	chest_cfg_id, o := this.db.Buildings.GetCfgId(chest_id)
	if !o {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_EXIST), nil
	}

	building_cfg := cfg_building_mgr.Map[chest_cfg_id]
	if nil == building_cfg {
		log.Error("Player OpenMapChest no building_cfg[%d] !", chest_cfg_id)
		return int32(msg_client_message.E_ERR_BUILDING_NOT_CFG), nil
	}

	if is_add {
		res := this.open_chest_cost(chest_cfg_id)
		if res < 0 {
			return res, nil
		}
	}

	dir, _ := this.db.Buildings.GetDir(chest_id)
	var width, height int32
	if table_config.BUILDING_DIR_BIG_X_DIR == dir {
		width, height = building_cfg.MapSizes[0], building_cfg.MapSizes[1]
	} else {
		width, height = building_cfg.MapSizes[1], building_cfg.MapSizes[0]
	}

	/*
		for tmp_x := int32(0); tmp_x < width; tmp_x++ {
			for tmp_y := int32(0); tmp_y < height; tmp_y++ {
				arena_xy := (chest_db.X+tmp_x)<<16 | (chest_db.Y+tmp_y)&0x0000FFFF
				delete(this.cur_building_map, arena_xy)
			}
		}
	*/
	this.ChkUpdateMyBuildingAreas()
	x, _ := this.db.Buildings.GetX(chest_id)
	y, _ := this.db.Buildings.GetY(chest_id)
	this.RemoveAreaBuilding(chest_id, x, y, width, height)
	if !is_add {
		this.return_chest_cost(chest_cfg_id)
		return chest_cfg_id, nil
	}

	res2cli := this.open_chest_result(chest_cfg_id)
	if res2cli == nil {
		this.return_chest_cost(chest_cfg_id)
	}
	res2cli.BuildingId = proto.Int32(chest_id)
	return chest_cfg_id, res2cli
}

// 成功返回建筑配置Id，失败返回小于零的错误码
func (this *Player) GetBackMapBuilding(building_id int32) int32 {
	building := this.db.Buildings.Get(building_id)
	if nil == building {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_EXIST)
	}

	build_cfg := cfg_building_mgr.Map[building.CfgId]
	if nil == build_cfg {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_CFG)
	}

	// 特殊判断猫舍
	if build_cfg.Type == PLAYER_BUILDING_TYPE_CAT_HOME {
		r := this.cathouse_can_remove(building_id)
		if r < 0 {
			return r
		}
	}

	var width, height int32
	if table_config.BUILDING_DIR_BIG_X_DIR == building.Dir {
		width, height = build_cfg.MapSizes[0], build_cfg.MapSizes[1]
	} else {
		width, height = build_cfg.MapSizes[1], build_cfg.MapSizes[0]
	}

	this.RemoveAreaBuilding(building_id, building.X, building.Y, width, height)
	//this.db.Buildings.Remove(building_id)

	this.item_cat_building_change_info.building_remove(this, building_id)
	tmp_count, _ := this.db.BuildingDepots.GetNum(building.CfgId)
	log.Info("GetBackMapBuilding增加物品前%d %d", building.CfgId, tmp_count)
	this.AddDepotBuilding(building.CfgId, 1, "getback", "building", false)
	tmp_count, _ = this.db.BuildingDepots.GetNum(building.CfgId)
	log.Info("GetBackMapBuilding增加物品后%d %d", building.CfgId, tmp_count)
	this.item_cat_building_change_info.depot_building_update(this, building.CfgId)

	// 特殊判断农田和猫舍
	if build_cfg.Type == PLAYER_BUILDING_TYPE_CAT_HOME {
		this.cathouse_remove(building_id, false)
	} else if build_cfg.Type == PLAYER_BUILDING_TYPE_FARMLAND {
		this.remove_crop(building_id)
	}

	// 减少魅力
	if build_cfg.Charm > 0 {
		this.SubCharmVal(build_cfg.Charm, "get_back_building", "building")
	}

	return building.CfgId
}

// ----------------------------------------------------------------------------

func (this *Player) IfXYAreaBlockFull(area_xy int32) bool {
	area_id := cfg_build_area_mgr.AreaXY2AreaId[area_xy]
	if area_id < 0 {
		log.Error("Player[%d] IfXYAreaBlockFull failed to find area_id for xy[%d] !", this.Id, area_xy)
		return true
	}

	area_un_cfg := cfg_areaunlock_mgr.Map[area_id]
	if nil == area_un_cfg {
		log.Error("Player[%d] IfXYAreaBlockFull failed to find un_cfg for area_id[%d] !", this.Id, area_id)
		return true
	}

	if area_un_cfg.MaxObstacle <= this.cur_areablocknum_map[area_id] {
		//log.Error("Player[%d] IfXYAreaBlockFull area[%d] block num[%d] over max !", this.Id, area_id, this.cur_areablocknum_map[area_id], area_un_cfg.MaxObstacle)
		return true
	}

	return false
}

func (this *Player) ChkMapBlock() (count int32) {
	cur_unix := int32(time.Now().Unix())
	last_up_unix := this.db.Info.GetLastMapBlockUpUnix()
	if 0 >= last_up_unix {
		this.db.Info.SetLastMapBlockUpUnix(cur_unix)
		return
	}

	this.ChkUpdateMyBuildingAreas()

	var tmp_block *table_config.XmlBlockItem
	var new_building *dbPlayerBuildingData
	var block_xy, area_id int32
	log.Info("玩家[%d]自动刷新障碍检查 %d %d", this.Id, time.Unix(int64(last_up_unix), 0).Format("2006-01-02 15:04:05.999999999"), time.Unix(int64(cur_unix), 0).Format("2006-01-02 15:04:05.999999999"))
	for tmp_unix := last_up_unix; tmp_unix+global_config_mgr.GetGlobalConfig().MapBlockRefleshSec < cur_unix; tmp_unix += global_config_mgr.GetGlobalConfig().MapBlockRefleshSec {
		tmp_block = cfg_block_mgr.RandBlock()
		//log.Info("玩家[%d]自动刷新障碍检查时间递增 %s 刷出障碍[%s]", this.Id, time.Unix(int64(tmp_unix), 0).Format("2006-01-02 15:04:05.999999999"), tmp_block.Id)
		if nil == tmp_block {
			log.Error("Player ChkMapBlock failed !")
			continue
		}

		new_building = this.TrySetMapBuildingDefDir(tmp_block.Id)
		if nil != new_building {
			log.Info("玩家[%d]自动刷新障碍检查尝试增加障碍 %d  %d", this.Id, tmp_block.Id, new_building.Id)
			this.item_cat_building_change_info.building_add(this, new_building.Id)
			count++
			this.db.Info.SetLastMapBlockUpUnix(tmp_unix + global_config_mgr.GetGlobalConfig().MapBlockRefleshSec)
			block_xy = (new_building.X)<<16 | (new_building.Y)&0x0000FFFF
			area_id = cfg_build_area_mgr.AreaXY2AreaId[block_xy]
			if area_id > 0 {
				this.cur_areablocknum_map[area_id] = this.cur_areablocknum_map[area_id] + 1
			}
		} else {
			//log.Info("玩家[%d]自动刷新障碍检查尝试增加障碍[%d]失败", this.Id, tmp_block.Id)
		}
	}

	return
}

func (this *Player) ChkMapChest() (count int32) {
	cur_unix := int32(time.Now().Unix())
	last_up_unix := this.db.Info.GetLastMapChestUpUnix()
	if 0 >= last_up_unix {
		this.db.Info.SetLastMapChestUpUnix(cur_unix)
		return
	}

	this.ChkUpdateMyBuildingAreas()

	var tmp_chest *table_config.XmlMapChestItem
	var new_building *dbPlayerBuildingData
	log.Info("玩家[%d]自动刷新宝箱检查 %s %s", this.Id, time.Unix(int64(last_up_unix), 0).Format("2006-01-02 15:04:05.999999999"), time.Unix(int64(cur_unix), 0).Format("2006-01-02 15:04:05.999999999"))
	for tmp_unix := last_up_unix; tmp_unix+global_config_mgr.GetGlobalConfig().MapChestRefleshSec < cur_unix; tmp_unix += global_config_mgr.GetGlobalConfig().MapChestRefleshSec {
		//log.Info("玩家[%d]自动刷新宝箱时间递增 %s", this.Id, time.Unix(int64(tmp_unix), 0).Format("2006-01-02 15:04:05.999999999"))
		if cur_unix-tmp_unix > cfg_mapchest_mgr.MaxBoxLastSec {
			continue
		}
		tmp_chest = cfg_mapchest_mgr.RandMapChest()
		if nil == tmp_chest {
			log.Error("Player ChkMapBlock failed !")
			continue
		}

		if cur_unix-tmp_unix > tmp_chest.LastSec {
			log.Info("玩家[%d]自动刷新宝箱时间递增尝试增加宝箱 但是直接过期了", this.Id, tmp_chest.Id, tmp_chest.LastSec)
			continue
		}

		new_building = this.TrySetMapBuildingDefDir(tmp_chest.Id)

		if nil != new_building {
			log.Info("玩家[%d]自动刷新宝箱时间递增尝试增加宝箱 %d %d 超时时间[%d + %d]", this.Id, tmp_chest.Id, new_building.Id, tmp_unix, tmp_chest.LastSec)
			this.db.Buildings.SetOverUnix(new_building.Id, tmp_unix+tmp_chest.LastSec)
			this.item_cat_building_change_info.building_add(this, new_building.Id)
			this.db.Info.SetLastMapChestUpUnix(tmp_unix + global_config_mgr.GetGlobalConfig().MapChestRefleshSec)
			count++
		} else {
			log.Info("玩家[%d]自动刷新宝箱", this.Id)
		}
	}

	return
}

func (this *Player) VisitPlayerBuildings(player_id int32) int32 {
	if player_id == this.Id {
		log.Error("no need to visit self buildings")
		return -1
	}

	var player_name, player_head string
	var player_level, player_vip_level, player_gold, player_diamond, player_charm int32
	var buildings_info []*msg_client_message.ViewBuildingInfo
	var area []*msg_client_message.AreaInfo
	player := player_mgr.GetPlayerById(player_id)
	if player != nil {
		building_ids := player.db.Buildings.GetAllIndex()
		if building_ids == nil {
			buildings_info = make([]*msg_client_message.ViewBuildingInfo, 0)
		} else {
			for i := 0; i < len(building_ids); i++ {
				cfg_id, o := player.db.Buildings.GetCfgId(building_ids[i])
				if !o {
					continue
				}
				building := cfg_building_mgr.Map[cfg_id]
				if building == nil {
					continue
				}
				x, _ := player.db.Buildings.GetX(building_ids[i])
				y, _ := player.db.Buildings.GetY(building_ids[i])
				dir, _ := player.db.Buildings.GetDir(building_ids[i])
				base_data := &msg_client_message.BuildingInfo{
					Id:    proto.Int32(building_ids[i]),
					CfgId: proto.Int32(cfg_id),
					X:     proto.Int32(x),
					Y:     proto.Int32(y),
					Dir:   proto.Int32(dir),
				}
				building_info := &msg_client_message.ViewBuildingInfo{
					BaseData: base_data,
				}
				var crop_data *msg_client_message.CropInfo
				var cathouse_data *msg_client_message.CatHouseInfo
				if building.Type == PLAYER_BUILDING_TYPE_FARMLAND {
					// 农田
					crop_data = player.db.Crops.GetCropInfo(building_ids[i])
					if crop_data != nil {
						building_info.CropData = crop_data
					}
				} else if building.Type == PLAYER_BUILDING_TYPE_CAT_HOME {
					// 猫舍
					if player.db.CatHouses.HasIndex(building_ids[i]) {
						level, _ := player.db.CatHouses.GetLevel(building_ids[i])
						cat_ids, _ := player.db.CatHouses.GetCatIds(building_ids[i])
						if cat_ids != nil && len(cat_ids) > 0 {
							for n := 0; n < len(cat_ids); n++ {
								cat_ids[n], _ = player.db.Cats.GetCfgId(cat_ids[n])
							}
						}
						is_done, _ := player.db.CatHouses.GetIsDone(building_ids[i])
						cathouse_data = &msg_client_message.CatHouseInfo{
							Level:  proto.Int32(level),
							CatIds: cat_ids,
							IsDone: proto.Int32(is_done),
						}
						building_info.CatHouseData = cathouse_data
					}
				}

				buildings_info = append(buildings_info, building_info)
			}
		}
		area = this.db.Areas.GetAllAreaInfo()
		player_name = player.db.GetName()
		player_head = player.db.Info.GetIcon()
		player_level = player.db.Info.GetLvl()
		player_vip_level = player.db.Info.GetVipLvl()
		player_gold = player.GetCoin()
		player_diamond = player.db.Info.GetDiamond()
		player_charm = player.db.Info.GetCharmVal()
	} else {
		result := this.rpc_visit_player(player_id)
		if result == nil {
			log.Error("### Player[%v] remote visit player[%v] failed", this.Id, player_id)
			return -1
		}
		buildings_info = make([]*msg_client_message.ViewBuildingInfo, len(result.Buildings))
		for i := 0; i < len(result.Buildings); i++ {
			r := result.Buildings[i]
			base_data := &msg_client_message.BuildingInfo{
				Id:    proto.Int32(r.BuildingId),
				CfgId: proto.Int32(r.BuildingTableId),
				X:     proto.Int32(r.CordX),
				Y:     proto.Int32(r.CordY),
				Dir:   proto.Int32(r.Dir),
			}
			buildings_info[i] = &msg_client_message.ViewBuildingInfo{
				BaseData: base_data,
			}
			if r.CropData != nil {
				buildings_info[i].CropData = &msg_client_message.CropInfo{
					CropId:        proto.Int32(r.CropData.CropId),
					RemainSeconds: proto.Int32(r.CropData.RemainSeconds),
				}
			} else if r.CatHouseData != nil {
				is_done := int32(0)
				if r.CatHouseData.IsDone {
					is_done = int32(1)
				}
				buildings_info[i].CatHouseData = &msg_client_message.CatHouseInfo{
					Level:  proto.Int32(r.CatHouseData.CatHouseLevel),
					CatIds: r.CatHouseData.CatIds,
					IsDone: proto.Int32(is_done),
				}
			}
		}

		player_name, player_level, player_head = GetPlayerBaseInfo(player_id)
		player_vip_level = result.ToPlayerVipLevel
		player_gold = result.ToPlayerGold
		player_diamond = result.ToPlayerDiamond
		player_charm = result.ToPlayerCharm
		if result.Areas == nil {
			area = make([]*msg_client_message.AreaInfo, 0)
		} else {
			for i := 0; i < len(result.Areas); i++ {
				a := &msg_client_message.AreaInfo{
					CfgId: proto.Int32(result.Areas[i].TableId),
				}
				area = append(area, a)
			}
		}
	}

	response := &msg_client_message.S2CVisitPlayerResult{
		Buildings:      buildings_info,
		PlayerId:       proto.Int32(player_id),
		PlayerName:     proto.String(player_name),
		PlayerLevel:    proto.Int32(player_level),
		PlayerVipLevel: proto.Int32(player_vip_level),
		PlayerHead:     proto.String(player_head),
		PlayerGold:     proto.Int32(player_gold),
		PlayerDiamond:  proto.Int32(player_diamond),
		PlayerCharm:    proto.Int32(player_charm),
		Areas:          area,
	}
	this.Send(response)

	this.TaskUpdate(table_config.TASK_FINISH_VISIT_FRIEND_NUM, false, 0, 1)

	return 1
}

func (this *Player) check_and_fill_buildings_msg() []*msg_client_message.BuildingInfo {
	var msg_builds []*msg_client_message.BuildingInfo
	building_ids := this.db.Buildings.GetAllIndex()
	if building_ids != nil {
		for i := 0; i < len(building_ids); i++ {
			cfg_id, o := this.db.Buildings.GetCfgId(building_ids[i])
			if !o {
				log.Error("Player[%v] cant get building[%v] config id", this.Id, building_ids[i])
				continue
			}
			building := cfg_building_mgr.Map[cfg_id]
			if building == nil {
				log.Error("building config id: %v invalid on player[%v] check data", cfg_id, this.Id)
				continue
			}

			x, _ := this.db.Buildings.GetX(building_ids[i])
			y, _ := this.db.Buildings.GetY(building_ids[i])
			dir, _ := this.db.Buildings.GetDir(building_ids[i])
			is_done, _ := this.db.Buildings.GetIsDone(building_ids[i])
			done := false
			if is_done > 0 {
				done = true
			}
			if building.Type == PLAYER_BUILDING_TYPE_CAT_HOME {
				if !this.db.CatHouses.HasIndex(building_ids[i]) {
					var width, height int32
					if table_config.BUILDING_DIR_BIG_X_DIR == dir {
						width, height = building.MapSizes[0], building.MapSizes[1]
					} else {
						width, height = building.MapSizes[1], building.MapSizes[0]
					}

					this.RemoveAreaBuilding(building_ids[i], x, y, width, height)
					//this.db.Buildings.Remove(building_ids[i])
					log.Error("Player[%v] building[%v] not found its cat house data", this.Id, building_ids[i])
					continue
				}
			}

			msg_build := &msg_client_message.BuildingInfo{
				Id:     proto.Int32(building_ids[i]),
				CfgId:  proto.Int32(cfg_id),
				X:      proto.Int32(x),
				Y:      proto.Int32(y),
				Dir:    proto.Int32(dir),
				IsDone: proto.Bool(done),
			}
			msg_builds = append(msg_builds, msg_build)
		}
	}
	return msg_builds
}

func (this *Player) BuildingSpeedup(building_id int32) int32 {
	if !this.db.Buildings.HasIndex(building_id) {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_EXIST)
	}
	building_tid, _ := this.db.Buildings.GetCfgId(building_id)
	building := cfg_building_mgr.Map[building_tid]
	if building == nil {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_CFG)
	}

	// 猫舍
	if building.Type == PLAYER_BUILDING_TYPE_CAT_HOME {
		return this.cathouse_speed_levelup(building_id)
	}

	is_done, _ := this.db.Buildings.GetIsDone(building_id)
	if is_done > 0 {
		return int32(msg_client_message.E_ERR_BUILDING_ALREDY_SET_DONE)
	}

	create_unix, _ := this.db.Buildings.GetCreateUnix(building_id)
	remain_seconds := GetRemainSeconds(create_unix, building.BuildTime)
	if remain_seconds == 0 {
		log.Warn("Player[%v] building[%v] already built", this.Id, building_id)
		return 1
	}

	cost_diamond := (remain_seconds + (global_id.CatHouseSpeedupLevelCostDiamond_18 - 1)) / global_id.CatHouseSpeedupLevelCostDiamond_18

	if this.GetDiamond() < cost_diamond {
		log.Error("Player[%v] building[%v] speedup level up not enough diamond, need[%v]", this.Id, building_id, cost_diamond)
		return int32(msg_client_message.E_ERR_CATHOUSE_SPEEDUP_LEVELUP_NOT_ENOUGH_DIAMOND)
	}

	this.db.Buildings.SetIsDone(building_id, 1)

	this.SubDiamond(cost_diamond, "build_speedup", "buildings")

	response := &msg_client_message.S2CBuildingSpeedupResult{
		BuildingId: proto.Int32(building_id),
	}
	this.Send(response)

	log.Debug("Player[%v] speed up building[%v]", this.Id, building_id)

	return 1
}

// ----------------------------------------------------------------------------

func reg_player_building_msg() {
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SGetBuildingInfos, C2SGetBuildingInfosHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SSetBuilding, C2SSetBuildingHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SGetBackBuilding, C2SGetBackBuildingHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SSellBuilding, C2SSellBuildingHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SRemoveBlock, C2SRemoveBlockHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SOpenMapChest, C2SOpenMapChestHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SMoveBuilding, C2SMoveBuildingHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SChgBuildingDir, C2SChgBuildingDirHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SVisitPlayer, C2SVisitPlayerHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SBuildingSpeedup, C2SBuildingSpeedupHandler)
}

func C2SGetBuildingInfosHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SGetBuildingInfos)
	if nil == req {
		log.Error("C2SGetItemInfosHandler req nil[%v] !", nil == req)
		return -1
	}

	res2cli := &msg_client_message.S2CRetBuildingInfos{}
	//p.db.Buildings.FillAllMsg(res2cli)
	res2cli.Builds = p.check_and_fill_buildings_msg()
	p.Send(res2cli)

	if p.ChkMapBlock() > 0 {
		p.item_cat_building_change_info.send_buildings_update(p)
	}

	if p.ChkMapChest() > 0 {
		p.item_cat_building_change_info.send_buildings_update(p)
	}

	return 1
}

func C2SSetBuildingHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SSetBuilding)
	if nil == p || nil == req {
		log.Error("C2SSetBuildingHandler req or p nil !")
		return -1
	}

	building_cfgid := req.GetBuildingCfgId()
	building_cfg := cfg_building_mgr.Map[building_cfgid]
	if nil == building_cfg {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_FIND_ITEM)
	}

	cur_building_count := p.db.Buildings.GetCountByType(building_cfg.Type)
	building_count_max := int32(9999)
	if 1 == building_cfg.IfFunction {
		building_count_max = 1
		if PLAYER_BUILDING_TYPE_CAT_HOME == building_cfg.Type {
			lvl_cfg := cfg_player_level_mgr.Map[p.db.Info.GetLvl()]
			if nil != lvl_cfg {
				building_count_max = lvl_cfg.MaxCattery
			}
		} else if PLAYER_BUILDING_TYPE_FARMLAND == building_cfg.Type {
			lvl_cfg := cfg_player_level_mgr.Map[p.db.Info.GetLvl()]
			if nil != lvl_cfg {
				building_count_max = lvl_cfg.MaxFarm
			}
		}
	}

	if cur_building_count >= building_count_max {
		log.Info("C2SSetBuildingHandler max_count %d %d", cur_building_count, building_count_max)
		return int32(msg_client_message.E_ERR_BUILDING_SET_MAX_COUNT)
	}

	switch building_cfg.UnlockType {
	case PLAYER_BUILDING_UNLOCK_TYPE_P_LVL:
		{
			if p.db.Info.GetLvl() < building_cfg.UnlockLevel {
				return int32(msg_client_message.E_ERR_BUILDING_BUYSET_LESS_P_LVL)
			}
		}
	case PLAYER_BUILDING_UNLOCK_TYPE_VIP_LVL:
		{
			if p.db.Info.GetVipLvl() < building_cfg.UnlockLevel {
				return int32(msg_client_message.E_ERR_BUILDING_BUYSET_LESS_VIP_LVL)
			}
		}
	}

	if 1 == req.GetIfBuy() {
		if PLAYER_BUILDING_TAG_1 != building_cfg.Tag {
			return int32(msg_client_message.E_ERR_BUILDING_SET_CANNOT_BUY)
		}
		if !p.ChkResEnough(building_cfg.UnlockCosts) {
			return int32(msg_client_message.E_ERR_BUILDING_NO_ENOUGH_COIN)
		}
	} else {
		building_db := p.db.BuildingDepots.Get(building_cfgid)
		if nil == building_db || building_db.Num <= 0 {
			return int32(msg_client_message.E_ERR_BUILDING_NO_DEPOT_BUILDING)
		}
	}

	p.ChkUpdateMyBuildingAreas()
	new_building_id := p.SetMapBuilding(building_cfgid, req.GetX(), req.GetY(), req.GetDir(), req.GetIfBuy(), true)
	if new_building_id <= 0 {
		return new_building_id
	}

	if 1 == req.GetIfBuy() {
		//p.SubCoin(building_cfg.SaleCoin, "set_building", "building")
		p.RemoveResources(building_cfg.UnlockCosts, "set_building", "building")
	} else {
		p.RemoveDepotBuilding(building_cfgid, 1, "set_building", "building")
	}

	p.item_cat_building_change_info.building_add(p, new_building_id)

	// 刷新当前区域
	p.item_cat_building_change_info.send_buildings_update(p)
	p.item_cat_building_change_info.send_items_update(p)
	p.item_cat_building_change_info.send_depot_building_update(p)

	// 猫舍
	if building_cfg.Type == PLAYER_BUILDING_TYPE_CAT_HOME {
		p.create_cathouse(new_building_id)
	} else if building_cfg.Type == PLAYER_BUILDING_TYPE_FOSTER { // 寄养所
		//p.db.Foster.SetBuildingId(new_building_id)
	}

	// 魅力
	if building_cfg.Charm > 0 {
		p.AddCharmVal(building_cfg.Charm, "set_building", "building")
	}

	return 1
}

func C2SGetBackBuildingHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SGetBackBuilding)
	if nil == p || nil == req {
		return int32(msg_client_message.E_ERR_INTERNAL)
	}

	iret := p.GetBackMapBuilding(req.GetBuildingId())
	if iret <= 0 {
		return iret
	}

	p.item_cat_building_change_info.send_buildings_update(p)
	p.item_cat_building_change_info.send_depot_building_update(p)

	return 1
}

func C2SSellBuildingHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SSellBuilding)
	if nil == p || nil == req {
		return int32(msg_client_message.E_ERR_INTERNAL)
	}

	building_id := req.GetBuildingId()

	building := p.db.Buildings.Get(building_id)
	if nil == building {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_EXIST)
	}

	build_cfg := cfg_building_mgr.Map[building.CfgId]
	if nil == build_cfg {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_CFG)
	}

	// 猫舍
	if build_cfg.Type == PLAYER_BUILDING_TYPE_CAT_HOME {
		r := p.cathouse_can_remove(building_id)
		if r < 0 {
			return r
		}
	} else if build_cfg.Type == PLAYER_BUILDING_TYPE_FARMLAND {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_FOR_SELL)
	}

	if build_cfg.Type != PLAYER_BUILDING_TYPE_CAT_HOME && build_cfg.SaleCoin <= 0 {
		return int32(msg_client_message.E_ERR_BUILDING_NOT_FOR_SELL)
	}

	var width, height int32
	if table_config.BUILDING_DIR_BIG_X_DIR == building.Dir {
		width, height = build_cfg.MapSizes[0], build_cfg.MapSizes[1]
	} else {
		width, height = build_cfg.MapSizes[1], build_cfg.MapSizes[0]
	}

	p.ChkUpdateMyBuildingAreas()
	p.RemoveAreaBuilding(building_id, building.X, building.Y, width, height)
	//p.db.Buildings.Remove(req.GetBuildingId())

	// 猫舍
	if build_cfg.Type == PLAYER_BUILDING_TYPE_CAT_HOME {
		p.cathouse_remove(req.GetBuildingId(), true)
	} else if build_cfg.Type == PLAYER_BUILDING_TYPE_FARMLAND {
		if build_cfg.Type == PLAYER_BUILDING_TYPE_FARMLAND {
			p.remove_crop(building_id)
		}
		p.AddCoin(build_cfg.SaleCoin, "sell", "building")
	}

	p.item_cat_building_change_info.building_remove(p, building_id)
	p.item_cat_building_change_info.send_buildings_update(p)

	// 减少魅力
	if build_cfg.Charm > 0 {
		p.SubCharmVal(build_cfg.Charm, "sell_building", "building")
	}

	//response := &msg_client_message.S2CSellBuilding{}
	//response.BuildingId = proto.Int32(req.GetBuildingId())

	return 1
}

func C2SRemoveBlockHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SRemoveBlock)

	block_id := req.GetBuildingId()

	iret, res2cli := p.ReomveMapBlock(block_id)
	if iret <= 0 {
		return iret
	}

	p.item_cat_building_change_info.building_remove(p, block_id)
	p.item_cat_building_change_info.send_buildings_update(p)
	var tmp_item *msg_client_message.ItemInfo
	for idx := 0; idx < len(res2cli.Items); idx++ {
		tmp_item = res2cli.Items[idx]
		if nil == tmp_item {
			continue
		}

		p.item_cat_building_change_info.item_update(p, tmp_item.GetItemCfgId())
	}

	p.item_cat_building_change_info.send_items_update(p)

	if nil != res2cli {
		p.Send(res2cli)
	}

	return 1
}

func C2SOpenMapChestHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SOpenMapChest)

	chest_id := req.GetBuildingId()
	friend_id := req.GetFriendId()

	if friend_id > 0 {
		return p.open_friend_chest(friend_id, chest_id)
	}

	iret, res2cli := p.OpenMapChest(chest_id, true)
	if iret <= 0 {
		return iret
	}

	p.item_cat_building_change_info.building_remove(p, chest_id)
	p.item_cat_building_change_info.send_buildings_update(p)

	if nil != res2cli {
		p.Send(res2cli)
	}

	return 1
}

func C2SMoveBuildingHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {

	req := msg.(*msg_client_message.C2SMoveBuilding)

	building_id := req.GetBuildingId()
	p.ChkUpdateMyBuildingAreas()
	iret := p.MoveMapBuilding(building_id, req.GetX(), req.GetY(), req.GetDir())
	if iret <= 0 {
		return iret
	}

	p.item_cat_building_change_info.building_update(p, building_id)
	p.item_cat_building_change_info.send_buildings_update(p)

	return 1
}

func C2SChgBuildingDirHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {

	req := msg.(*msg_client_message.C2SChgBuildingDir)

	building_id := req.GetBuildingId()

	iret := p.ChgMapBuildingDir(building_id)
	if iret <= 0 {
		return iret
	}

	p.item_cat_building_change_info.building_update(p, building_id)
	p.item_cat_building_change_info.send_buildings_update(p)

	return 1
}

func C2SVisitPlayerHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SVisitPlayer)
	if req == nil {
		return -1
	}
	return p.VisitPlayerBuildings(req.GetPlayerId())
}

func C2SBuildingSpeedupHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SBuildingSpeedup)
	if req == nil {
		return -1
	}
	return p.BuildingSpeedup(req.GetBuildingId())
}
