package main

import (
	"errors"
	"fmt"
	"libs/log"
	"libs/rpc"
	"libs/utils"
	"public_message/gen_go/client_message"
	"time"
	"youma/rpc_common"
)

// ping RPC服务
type R2H_PingProc struct{}

func (this *R2H_PingProc) Do(args *rpc_common.R2H_Ping, reply *rpc_common.R2H_Pong) error {
	// 不做任何处理
	log.Info("收到rpc服务的ping请求")
	return nil
}

//// 玩家调用
type R2H_PlayerProc struct {
}

// 获取查找的玩家数据
func (this *R2H_PlayerProc) GetInfoToSearch(args *rpc_common.R2H_SearchPlayer, reply *rpc_common.R2H_SearchPlayerResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	p := player_mgr.GetPlayerById(args.Id)
	if p == nil {
		err_str := fmt.Sprintf("RPC R2H_PlayerProc @@@ Not found player[%v], get player info failed", args.Id)
		return errors.New(err_str)
	}
	reply.Head = p.db.Info.GetIcon()
	reply.Nick = p.db.GetName()
	reply.Level = p.db.Info.GetLvl()

	log.Debug("RPC R2H_PlayerProc @@@ Get player[%v] info", args.Id)

	return nil
}

// 好友
type R2H_FriendProc struct {
}

// 申请添加好友
func (this *R2H_FriendProc) AddFriendById(args *rpc_common.R2H_AddFriendById, reply *rpc_common.R2H_AddFriendResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	p := player_mgr.GetPlayerById(args.AddPlayerId)
	if p == nil {
		reply.Error = int32(msg_client_message.E_ERR_PLAYER_NOT_EXIST)
		log.Error("RPC R2H_FriendProc @@@ not found player[%v], cant add player[%v] to friend", args.AddPlayerId, args.PlayerId)
	} else {
		if p.db.Friends.HasIndex(args.PlayerId) {
			reply.Error = int32(msg_client_message.E_ERR_FRIEND_THE_PLAYER_ALREADY_FRIEND)
			log.Error("RPC R2H_FriendProc @@@ player[%v] and player[%v] already friends", args.AddPlayerId, args.PlayerId)
		} else {
			res := p.db.FriendReqs.CheckAndAdd(args.PlayerId, args.PlayerName)
			if res < 0 {
				reply.Error = res
				log.Error("RPC R2H_FriendProc @@@ player[%v] already has player[%v] request to friend", args.AddPlayerId, args.PlayerId)
			}
		}
	}

	reply.AddPlayerId = args.AddPlayerId
	reply.PlayerId = args.PlayerId

	if reply.Error >= 0 {
		log.Debug("RPC R2H_FriendProc @@@ Player[%v] requested add friend[%v]", args.PlayerId, args.AddPlayerId)
	}

	return nil
}

// 同意或拒绝好友申请
func (this *R2H_FriendProc) AgreeAddFriend(args *rpc_common.R2H_AgreeAddFriend, reply *rpc_common.R2H_AgreeAddFriendResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	p := player_mgr.GetPlayerById(args.AgreePlayerId)
	if p == nil {
		err_str := fmt.Sprintf("RPC R2H_FriendProc @@@ Not found player[%v]，player[%v] agree add friend failed", args.AgreePlayerId, args.PlayerId)
		return errors.New(err_str)
	}

	if !args.IsAgree {
		return nil
	}

	d := &dbPlayerFriendData{}
	d.FriendId = args.PlayerId
	d.FriendName = args.PlayerName
	p.db.Friends.Add(d)

	reply.IsAgree = args.IsAgree
	reply.PlayerId = args.PlayerId
	reply.AgreePlayerId = args.AgreePlayerId
	reply.AgreePlayerName = p.db.GetName()
	reply.AgreePlayerLevel = p.db.Info.GetLvl()
	reply.AgreePlayerVipLevel = p.db.Info.GetVipLvl()
	reply.AgreePlayerHead = p.db.Info.GetIcon()
	reply.AgreePlayerLastLogin = p.db.Info.GetLastLogin()

	log.Debug("RPC R2H_FriendProc @@@ Player[%v] agreed add friend[%v]", args.PlayerId, args.AgreePlayerId)

	return nil
}

// 删除好友
func (this *R2H_FriendProc) RemoveFriend(args *rpc_common.R2H_RemoveFriend, reply *rpc_common.R2H_RemoveFriendResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	p := player_mgr.GetPlayerById(args.RemovePlayerId)
	if p == nil {
		err_str := fmt.Sprintf("RPC R2H_FriendProc @@@ Not found player[%v], player[%v] remove friend failed", args.PlayerId)
		return errors.New(err_str)
	}

	p.db.Friends.Remove(args.PlayerId)

	log.Debug("RPC R2H_FriendProc @@@ Player[%v] removed friend[%v]", args.PlayerId, args.RemovePlayerId)

	return nil
}

// 大厅到大厅的好友调用
type H2H_FriendProc struct {
}

// 获取好友数据
func (this *H2H_FriendProc) GetFriendInfo(args *rpc_common.H2H_GetFriendInfo, reply *rpc_common.H2H_GetFriendInfoResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	p := player_mgr.GetPlayerById(args.PlayerId)
	if p == nil {
		err_str := fmt.Sprintf("RPC H2H_FriendProc::GetFriendInfo @@@ Not found Player[%v], get player info failed", args.PlayerId)
		return errors.New(err_str)
	}

	reply.PlayerId = p.Id
	reply.PlayerName = p.db.GetName()
	reply.Level = p.db.Info.GetLvl()
	reply.VipLevel = p.db.Info.GetVipLvl()
	reply.LastLogin = p.db.Info.GetLastLogin()

	log.Debug("RPC H2H_FriendProc @@@ Get player[%v] info", args.PlayerId)

	return nil
}

// 加为好友
func (this *H2H_FriendProc) AddFriend(args *rpc_common.H2H_AddFriend, reply *rpc_common.H2H_AddFriendResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	p := player_mgr.GetPlayerById(args.ToPlayerId)
	if p == nil {
		err_str := fmt.Sprintf("RPC H2H_FriendProc::AddFriend @@@ not found player[%v], add friend failed", args.ToPlayerId)
		return errors.New(err_str)
	}

	if p.db.Friends.HasIndex(args.FromPlayerId) {
		// 已是好友
		reply.Error = int32(msg_client_message.E_ERR_FRIEND_THE_PLAYER_ALREADY_FRIEND)
	}

	// 已有申请
	from_player_name, _, _ := GetPlayerBaseInfo(args.FromPlayerId)
	res := p.db.FriendReqs.CheckAndAdd(args.FromPlayerId, from_player_name)
	if res < 0 {
		reply.Error = int32(msg_client_message.E_ERR_FRIEND_THE_PLAYER_REQUESTED)
	}

	reply.FromPlayerId = args.FromPlayerId
	reply.ToPlayerId = args.ToPlayerId

	log.Debug("RPC H2H_FriendProc @@@ player[%v] added friend[%v]", args.FromPlayerId, args.ToPlayerId)

	return nil
}

// 删除好友
func (this *H2H_FriendProc) RemoveFriend(args *rpc_common.H2H_RemoveFriend, reply *rpc_common.H2H_RemoveFriendResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	p := player_mgr.GetPlayerById(args.ToPlayerId)
	if p == nil {
		err_str := fmt.Sprintf("RPC H2H_FriendProc::RemoveFriend @@@ not found player[%v], player[%v] remove friend failed", args.ToPlayerId, args.FromPlayerId)
		return errors.New(err_str)
	}

	p.remove_friend_data(args.FromPlayerId)

	reply.FromPlayerId = args.FromPlayerId
	reply.ToPlayerId = args.ToPlayerId

	log.Debug("RPC H2H_FriendProc @@@ player[%v] removed friend[%v]", args.FromPlayerId, args.ToPlayerId)

	return nil
}

// 赠送友情点
func (this *H2H_FriendProc) GiveFriendPoints(args *rpc_common.H2H_GiveFriendPoints, reply *rpc_common.H2H_GiveFriendPointsResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	reply.FromPlayerId = args.FromPlayerId
	reply.ToPlayerId = args.ToPlayerId
	var err_str string
	p := player_mgr.GetPlayerById(args.ToPlayerId)
	if p == nil {
		reply.Error = 1
		err_str = fmt.Sprintf("RPC H2H_FriendProc::GiveFriendPoints @@@ not found Player[%v], get player info failed", args.ToPlayerId)
		return errors.New(err_str)
	}

	reply.Error, reply.LastSave, reply.RemainSeconds = p.store_friend_points(args.FromPlayerId)

	log.Debug("RPC H2H_FriendProc @@@ Player[%v] Gived Friend[%v] Points[%v] Error[%v]", args.FromPlayerId, args.ToPlayerId, args.GivePoints, reply.Error)
	return nil
}

// 好友聊天
func (this *H2H_FriendProc) Chat(args *rpc_common.H2H_FriendChat, reply *rpc_common.H2H_FriendChatResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	p := player_mgr.GetPlayerById(args.ToPlayerId)
	if p == nil {
		err_str := fmt.Sprintf("RPC H2H_FriendProc::Chat @@@ not found Player[%v]", args.ToPlayerId)
		return errors.New(err_str)
	}

	res := p.friend_chat_add(args.FromPlayerId, args.Message)
	if res < 0 {
		reply.Error = res
		log.Error("RPC H2H_FriendProc::Chat @@@ player[%v] chat to friend[%v] error[%v]", args.FromPlayerId, args.ToPlayerId, res)
	} else {
		reply.FromPlayerId = args.FromPlayerId
		reply.ToPlayerId = args.ToPlayerId
		reply.Message = args.Message
		log.Debug("RPC H2H_FriendProc @@@ Player[%v] chat friend[%v] message[%v]", args.FromPlayerId, args.ToPlayerId, args.Message)
	}
	return nil
}

// 刷新赠送好友
func (this *H2H_FriendProc) RefreshGivePoints(args *rpc_common.H2H_RefreshGiveFriendPoints, reply *rpc_common.H2H_RefreshGiveFriendPointsResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	friend := player_mgr.GetPlayerById(args.ToPlayerId)
	if friend == nil {
		err_str := fmt.Sprintf("RPC H2H_FriendProc::RefreshGivePoints @@@ not found Player[%v]", args.ToPlayerId)
		return errors.New(err_str)
	}

	friend.refresh_friend_give_points(args.FromPlayerId)
	return nil
}

// 获取玩家宝箱配置ID
func (this *H2H_FriendProc) GetPlayerChestTableId(args *rpc_common.H2H_GetPlayerChestTableId, reply *rpc_common.H2H_GetPlayerChestTableIdResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	player := player_mgr.GetPlayerById(args.ToPlayerId)
	if player == nil {
		reply.Error = int32(msg_client_message.E_ERR_PLAYER_NOT_EXIST)
		return nil
	}

	var o bool
	reply.ChestTableId, o = player.db.Buildings.GetCfgId(args.ChestId)
	if !o {
		reply.Error = int32(msg_client_message.E_ERR_BUILDING_NOT_EXIST)
		return nil
	}
	reply.ChestId = args.ChestId
	reply.FromPlayerId = args.FromPlayerId
	reply.ToPlayerId = args.ToPlayerId
	return nil
}

// 打开好友宝箱
func (this *H2H_FriendProc) OpenChest(args *rpc_common.H2H_OpenFriendChest, reply *rpc_common.H2H_OpenFriendChestResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	friend := player_mgr.GetPlayerById(args.ToPlayerId)
	if friend == nil {
		err_str := fmt.Sprintf("RPC H2H_FriendProc::OpenChest @@@ not found Player[%v]", args.ToPlayerId)
		return errors.New(err_str)
	}

	res, _ := friend.OpenMapChest(args.ChestId, false)
	if res < 0 {
		log.Error("RPC H2H_FriendProc::OpenChest @@@ Player[%v] open friend[%v] chest error[%v]", args.FromPlayerId, args.ToPlayerId, res)
	}

	chest := cfg_mapchest_mgr.Map[res]
	if chest != nil {
		friend.AddFriendPoints(chest.FriPoint, "friend_open_chest", "friend")
	}

	reply.FromPlayerId = args.FromPlayerId
	reply.ToPlayerId = args.ToPlayerId
	reply.ChestTableId = res

	friend.item_cat_building_change_info.building_remove(friend, args.ChestId)
	friend.item_cat_building_change_info.send_buildings_update(friend)

	log.Debug("Player[%v] open friend[%v] chest[%v]", args.FromPlayerId, args.ToPlayerId, args.ChestId)
	return nil
}

// 大厅到大厅玩家调用
type H2H_PlayerProc struct {
}

// 更新玩家基本信息
func (this *H2H_PlayerProc) UpdateBaseInfo(args *rpc_common.H2H_BaseInfo, result *rpc_common.H2H_BaseInfoResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	row := os_player_mgr.GetPlayer(args.FromPlayerId)
	if row == nil {
		err_str := fmt.Sprintf("RPC H2H_PlayerProc::UpdateBaseInfo @@@ not found player[%v]", args.FromPlayerId)
		return errors.New(err_str)
	}

	row.SetName(args.Nick)
	row.SetLevel(args.Level)
	row.SetHead(args.Head)

	log.Debug("RPC H2H_PlayerProc::UpdateBaseInfo @@@ player[%v] updated base info", args.FromPlayerId)

	return nil
}

// 点赞
func (this *H2H_PlayerProc) Zan(args *rpc_common.H2H_ZanPlayer, result *rpc_common.H2H_ZanPlayerResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	p := player_mgr.GetPlayerById(args.ToPlayerId)
	if p == nil {
		err_str := fmt.Sprintf("RPC H2H_PlayerProc::Zan @@@ not found player[%v]", args.ToPlayerId)
		return errors.New(err_str)
	}

	zan := p.db.Info.IncbyZan(1)

	result.FromPlayerId = args.FromPlayerId
	result.ToPlayerId = args.ToPlayerId
	result.ToPlayerZanNum = zan

	log.Debug("RPC H2H_PlayerProc @@@ Player[%v] zan player[%v], total zan[%v]", args.FromPlayerId, args.ToPlayerId, zan)

	return nil
}

// 拜访
func (this *H2H_PlayerProc) VisitPlayer(args *rpc_common.H2H_VisitPlayer, result *rpc_common.H2H_VisitPlayerResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	p := player_mgr.GetPlayerById(args.ToPlayerId)
	if p == nil {
		err_str := fmt.Sprintf("RPC H2H_PlayerProc::VisitPlayer @@@ not found player[%v]", args.ToPlayerId)
		return errors.New(err_str)
	}

	building_ids := p.db.Buildings.GetAllIndex()
	if building_ids == nil {
		result.Buildings = make([]*rpc_common.H2H_BuildingInfo, 0)
	} else {
		for i := 0; i < len(building_ids); i++ {
			if !p.db.Buildings.HasIndex(building_ids[i]) {
				continue
			}
			building_info := &rpc_common.H2H_BuildingInfo{}
			building_info.BuildingId = building_ids[i]
			building_info.BuildingTableId, _ = p.db.Buildings.GetCfgId(building_ids[i])
			building_info.CordX, _ = p.db.Buildings.GetX(building_ids[i])
			building_info.CordY, _ = p.db.Buildings.GetY(building_ids[i])
			building_info.Dir, _ = p.db.Buildings.GetDir(building_ids[i])
			building := cfg_building_mgr.Map[building_info.BuildingTableId]
			if building != nil {
				if building.Type == PLAYER_BUILDING_TYPE_FARMLAND {
					crop_data := p.db.Crops.GetCropInfo4RPC(building_ids[i])
					if crop_data != nil {
						building_info.CropData = crop_data
					}
				} else if building.Type == PLAYER_BUILDING_TYPE_CAT_HOME {
					cathouse_data := p.db.CatHouses.Get4RPC(building_ids[i])
					if cathouse_data != nil {
						building_info.CatHouseData = cathouse_data
					}
					if cathouse_data.CatIds != nil && len(cathouse_data.CatIds) > 0 {
						for i := 0; i < len(cathouse_data.CatIds); i++ {
							cathouse_data.CatIds[i], _ = p.db.Cats.GetCfgId(cathouse_data.CatIds[i])
						}
					}
				}
			}
			result.Buildings = append(result.Buildings, building_info)
		}
		//result.ToPlayerName = p.db.GetName()
		//result.ToPlayerHead = p.db.Info.GetIcon()
		//result.ToPlayerLevel = p.db.Info.GetLvl()
		result.ToPlayerVipLevel = p.db.Info.GetVipLvl()
		result.ToPlayerGold = p.db.Info.GetCoin()
		result.ToPlayerDiamond = p.db.Info.GetDiamond()
		areas := p.db.Areas.GetAllAreaInfo()
		if areas == nil {
			result.Areas = make([]*rpc_common.H2H_AreaInfo, 0)
		} else {
			for i := 0; i < len(areas); i++ {
				a := &rpc_common.H2H_AreaInfo{
					TableId: areas[i].GetCfgId(),
				}
				result.Areas = append(result.Areas, a)
			}
		}
	}
	return nil
}

func (this *H2H_PlayerProc) CatInfo(args *rpc_common.H2H_PlayerCatInfo, result *rpc_common.H2H_PlayerCatInfoResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	player := player_mgr.GetPlayerById(args.ToPlayerId)
	if player == nil {
		result.Error = int32(msg_client_message.E_ERR_PLAYER_NOT_EXIST)
		log.Error("RPC H2H_PlayerProc::CatInfo @@@ not found player[%v]", args.ToPlayerId)
		return nil
	}

	cat_id := args.ToPlayerCatId
	o := false
	result.ToPlayerCatLevel, o = player.db.Cats.GetLevel(cat_id)
	if !o {
		result.Error = int32(msg_client_message.E_ERR_CAT_NOT_FOUND)
		log.Error("Player[%v] no cat[%v]", args.ToPlayerId, args.ToPlayerCatId)
		return nil
	}
	result.ToPlayerCatExp, _ = player.db.Cats.GetExp(cat_id)
	result.ToPlayerCatStar, _ = player.db.Cats.GetStar(cat_id)
	result.ToPlayerCatSkillLevel, _ = player.db.Cats.GetSkillLevel(cat_id)
	result.ToPlayerCatAddCoin, _ = player.db.Cats.GetCoinAbility(cat_id)
	result.ToPlayerCatAddMatch, _ = player.db.Cats.GetMatchAbility(cat_id)
	result.ToPlayerCatAddExplore, _ = player.db.Cats.GetExploreAbility(cat_id)
	return nil
}

// 寄养调用
type H2H_FosterProc struct {
}

func (this *H2H_FosterProc) SetCat2Friend(args *rpc_common.H2H_FosterCat2Friend, result *rpc_common.H2H_FosterCat2FriendResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	friend := player_mgr.GetPlayerById(args.ToFriendId)
	if friend == nil {
		err_str := fmt.Sprintf("RPC H2H_FosterProc::SetCat2Friend @@@ not found player[%v]", args.ToFriendId)
		return errors.New(err_str)
	}

	// 先结算该好友的寄养所
	begin := time.Now().Unix()*1000000 + time.Now().UnixNano()/1000
	friend.foster_settlement_friends_cat()
	end := time.Now().Unix()*10000000 + time.Now().UnixNano()/1000
	log.Debug("RPC Player[%v] foster settlement other players cat cost time: %v us", args.FromPlayerId, end-begin)

	card := foster_card_table_mgr.GetByItemId(args.FromPlayerCardId)
	res := friend.db.FosteredFriendCats.CheckAndAddFriendCat(args.FromPlayerId, args.FromPlayerCatId, args.FromPlayerCatTableId, args.FromPlayerCatLevel, args.FromPlayerCatStar, friend.fostered_slot_num_for_friend(), card)
	if res < 0 {
		err_str := fmt.Sprintf("RPC H2H_FosterProc::SetCat2Friend @@@ player[%v] set cat[id:%v,table_id:%v,level:%v,star:%v] to friend[%v] failed, err[%v]",
			args.FromPlayerId, args.FromPlayerCatId, args.FromPlayerCatTableId, args.FromPlayerCatLevel, args.FromPlayerCatStar, args.ToFriendId, res)
		return errors.New(err_str)
	}

	log.Debug("RPC H2H_FosterProc @@@ player[%v] set cat[%v] to friend[%v]", args.FromPlayerId, args.FromPlayerCatId, args.ToFriendId)

	return nil
}

func (this *H2H_FosterProc) Settlement2Friend(args *rpc_common.H2H_FosterSettlement2Friend, result *rpc_common.H2H_FosterSettlement2FriendResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	friend := player_mgr.GetPlayerById(args.ToPlayerId)
	if friend == nil {
		err_str := fmt.Sprintf("RPC H2H_FosterProc::Settlement2Friend @@@ not found player[%v]", args.ToPlayerId)
		return errors.New(err_str)
	}

	friend.settlement_from_friend_foster(args.ToPlayerCatId, args.ToPlayerCatExp, args.ToPlayerItems)
	friend.db.FosterCatOnFriends.Remove(args.FromPlayerId)

	log.Debug("RPC H2H_FosterProc @@@ player[%v] settlement foster cat[%v] to friend[%v], exp[%v] items[%v]", args.FromPlayerId, args.ToPlayerCatId, args.ToPlayerId, args.ToPlayerCatExp, args.ToPlayerItems)

	return nil
}

func (this *H2H_FosterProc) SettlementPlayersCatWithFriend(args *rpc_common.H2H_FosterSettlementPlayersCatWithFriend, result *rpc_common.H2H_FosterSettlementPlayersCatWithFriendResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	friend := player_mgr.GetPlayerById(args.ToPlayerId)
	if friend == nil {
		err_str := fmt.Sprintf("RPC H2H_FosterProc::SettlementPlayersCatWithFriend @@@ not found player[%v]", args.ToPlayerId)
		return errors.New(err_str)
	}

	friend.foster_settlement_friends_cat()

	log.Debug("RPC H2H_FosterProc @@@ player[%v] foster settlement players cat with friend[%v]", args.FromPlayerId, args.ToPlayerId)

	return nil
}

func (this *H2H_FosterProc) GetCatOnFriend(args *rpc_common.H2H_FosterGetCatInfoOnFriend, result *rpc_common.H2H_FosterGetCatInfoOnFriendResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	player := player_mgr.GetPlayerById(args.ToPlayerId)
	if player == nil {
		err_str := fmt.Sprintf("RPC H2H_FosterProc::GetCatOnFriend @@@ not found player[%v]", args.ToPlayerId)
		return errors.New(err_str)
	}

	card_id, _ := player.db.FosterFriendCats.GetStartCardId(args.FromPlayerId)
	remain_seconds, cat_exp, items := player.db.FosterFriendCats.Settlement(args.FromPlayerId)
	if remain_seconds <= 0 {
		player.db.FosterFriendCats.Remove(args.FromPlayerId)
	}

	result.FromPlayerId = args.FromPlayerId
	result.FromPlayerCatId = args.FromPlayerCatId
	result.ToFriendId = args.ToPlayerId
	result.StartCardId = card_id
	result.RemainSeconds = remain_seconds
	result.FromPlayerCatExp = cat_exp
	result.FromPlayerItems = items

	log.Debug("RPC H2H_FosterProc @@@ player[%v] get cat[%v] from friend[%v]", args.FromPlayerId, args.FromPlayerCatId, args.ToPlayerId)

	return nil
}

func (this *H2H_FosterProc) GetPlayerFosterData(args *rpc_common.H2H_FosterGetPlayerFosterData, result *rpc_common.H2H_FosterGetPlayerFosterDataResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	player := player_mgr.GetPlayerById(args.ToPlayerId)
	if player == nil {
		err_str := fmt.Sprintf("RPC H2H_FosterProc::GetPlayerFosterData @@@ not found player[%v]", args.ToPlayerId)
		return errors.New(err_str)
	}

	result.FromPlayerId = args.FromPlayerId
	result.ToPlayerId = args.ToPlayerId
	result.FosterCardId = player.db.Foster.GetEquippedCardId()
	result.CardRemainSeconds = player.db.Foster.GetCardRemainSeconds()
	result.FosteredSlot = player.fostered_slot_num_for_friend()
	cat_ids := player.db.FosterCats.GetAllIndex()
	result.PlayerCats = make([]rpc_common.H2H_FosterCat, len(cat_ids))
	for i := 0; i < len(cat_ids); i++ {
		result.PlayerCats[i].CatTableId, _ = player.db.Cats.GetCfgId(cat_ids[i])
		result.PlayerCats[i].CatLevel, _ = player.db.Cats.GetLevel(cat_ids[i])
		result.PlayerCats[i].CatStar, _ = player.db.Cats.GetStar(cat_ids[i])
	}
	friend_player_ids := player.db.FosterFriendCats.GetAllIndex()
	result.PlayerFriendCats = make([]rpc_common.H2H_FosteredCat, len(friend_player_ids))
	for i := 0; i < len(friend_player_ids); i++ {
		result.PlayerFriendCats[i].CatTableId, _ = player.db.FosterFriendCats.GetCatTableId(friend_player_ids[i])
		result.PlayerFriendCats[i].CatNick, _ = player.db.FosterFriendCats.GetCatNick(friend_player_ids[i])
		result.PlayerFriendCats[i].CatLevel, _ = player.db.FosterFriendCats.GetCatLevel(friend_player_ids[i])
		result.PlayerFriendCats[i].CatStar, _ = player.db.FosterFriendCats.GetCatStar(friend_player_ids[i])
		result.PlayerFriendCats[i].PlayerId = friend_player_ids[i]
	}

	return nil
}

func (this *H2H_FosterProc) GetEmptySlotFriendInfo(args *rpc_common.H2H_FosterGetEmptySlotFriendInfo, result *rpc_common.H2H_FosterGetEmptySlotFriendInfoResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	player := player_mgr.GetPlayerById(args.ToPlayerId)
	if player == nil {
		err_str := fmt.Sprintf("RPC H2H_FosterProc::GetEmptySlotFriendInfo @@@ not found player[%v]", args.ToPlayerId)
		return errors.New(err_str)
	}

	if !player.foster_has_empty_slot_for_friend(args.FromPlayerId) {
		err_str := fmt.Sprintf("RPC H2H_FosterProc::GetEmptySlotFriendInfo player[%v] @@@ not has empty slot for friend[%v]", args.FromPlayerId, args.ToPlayerId)
		return errors.New(err_str)
	}

	result.FromPlayerId = args.FromPlayerId
	result.ToPlayerId = args.ToPlayerId
	result.ToPlayerVipLevel = player.db.Info.GetVipLvl()
	result.ToPlayerLastLogin = player.db.Info.GetLastLogin()
	result.FosterCardId = player.db.Foster.GetEquippedCardId()
	log.Debug("!!!!!! empty slot friend info: %v, args[%v], result[%v]", *result, args, result)

	return nil
}

func (this *H2H_FosterProc) GetFriendCats(args *rpc_common.H2H_FosterGetFriendCats, result *rpc_common.H2H_FosterGetFriendCatsResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	player := player_mgr.GetPlayerById(args.ToPlayerId)
	if player == nil {
		err_str := fmt.Sprintf("RPC H2H_FosterProc::GetFriendCats @@@ not found player[%v]", args.ToPlayerId)
		return errors.New(err_str)
	}

	result.Cats, result.OutFriendCats = player.foster_format_data_to_rpc()
	result.CanFosteredSlot = player.fostered_slot_num_for_friend()

	log.Debug("!!!!!! Player[%v] foster get friend[%v] cats", args.FromPlayerId, args.ToPlayerId)

	return nil
}

func (this *H2H_FosterProc) GetIncome(args *rpc_common.H2H_FosterIncome, result *rpc_common.H2H_FosterIncomeResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	player := player_mgr.GetPlayerById(args.ToPlayerId)
	if player == nil {
		err_str := fmt.Sprintf("RPC H2H_FosterProc::GetEmptySlotFriendInfo @@@ not found player[%v]", args.ToPlayerId)
		return errors.New(err_str)
	}

	player.foster_settlement()
	result.CatIds, result.CardIds = player.foster_get_income_from_friend(args.FromPlayerId)

	log.Debug("!!!!!! Player[%v] get foster income from friend[%v]", args.FromPlayerId, args.ToPlayerId)
	return nil
}

func (this *H2H_FosterProc) ClearFinishedCats(args *rpc_common.H2H_FosterClearFinishedCats, result *rpc_common.H2H_FosterClearFinishedCatsResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	player := player_mgr.GetPlayerById(args.ToPlayerId)
	if player == nil {
		err_str := fmt.Sprintf("RPC H2H_FosterProc::ClearFinishedCats @@@ not found player[%v]", args.ToPlayerId)
		return errors.New(err_str)
	}

	if args.OutCats != nil {
		for _, cid := range args.OutCats {
			player.db.FosterCatOnFriends.Remove(cid)
		}
	}

	log.Debug("!!!!!! Player[%v] clear friend[%v] finished cats[%v]", args.FromPlayerId, args.ToPlayerId, args.OutCats)

	return nil
}

func (this *H2H_FosterProc) FinishCats(args *rpc_common.H2H_FosterFinishCats, result *rpc_common.H2H_FosterFinishCatsResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	player := player_mgr.GetPlayerById(args.ToPlayerId)
	if player == nil {
		err_str := fmt.Sprintf("RPC H2H_FosterProc::FinishCats @@@ not found player[%v]", args.ToPlayerId)
		return errors.New(err_str)
	}

	if args.Cats != nil {
		for _, cid := range args.Cats {
			player.db.FosteredFriendCats.Remove(utils.Int64From2Int32(args.FromPlayerId, cid))
		}
	}

	log.Debug("!!!!!! Player[%v] finished friend[%v] cats[%v]", args.FromPlayerId, args.ToPlayerId, args.Cats)

	return nil
}

// 向另一个HallServer请求玩家数据
type R2H_PlayerStageInfoProc struct {
}

func (this *R2H_PlayerStageInfoProc) Do(args *rpc_common.R2H_PlayerStageInfoReq, result *rpc_common.R2H_PlayerStageInfoResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	p := player_mgr.GetPlayerById(args.PlayerId)
	if p == nil {
		return errors.New("无法找到玩家[%v]数据")
	}
	result.Head = p.db.Info.GetIcon()
	result.Level = p.db.Info.GetLvl()
	result.Nick = p.db.GetName()
	result.TopScore, _ = p.db.Stages.GetTopScore(args.StageId)
	log.Info("获取玩家[%v]的关卡[%v]信息[%v]", args.PlayerId, args.StageId, *result)
	return nil
}

// 全局调用
type H2H_GlobalProc struct {
}

func (this *H2H_GlobalProc) WorldChat(args *rpc_common.H2H_WorldChat, result *rpc_common.H2H_WorldChatResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	if !world_chat_mgr.push_chat_msg(args.ChatContent, args.FromPlayerId, args.FromPlayerLevel, args.FromPlayerName, args.FromPlayerHead) {
		err_str := fmt.Sprintf("@@@ H2H_GlobalProc::WorldChat Player[%v] world chat content[%v] failed", args.FromPlayerId, args.ChatContent)
		return errors.New(err_str)
	}
	log.Debug("@@@ H2H_GlobalProc::WorldChat Player[%v] world chat content[%v]", args.FromPlayerId, args.ChatContent)
	return nil
}

func (this *H2H_GlobalProc) Anouncement(args *rpc_common.H2H_Anouncement, result *rpc_common.H2H_AnouncementResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	if !anouncement_mgr.PushNew(args.MsgType, true, args.FromPlayerId, args.MsgParam1, args.MsgParam2, args.MsgParam3, args.MsgText) {
		err_str := fmt.Sprintf("@@@ H2H_GlobalProc::Anouncement Player[%v] anouncement msg_type[%v] msg_param[%v] failed", args.FromPlayerId, args.MsgType, args.MsgParam1)
		return errors.New(err_str)
	}
	log.Debug("@@@ H2H_GlobalProc::Anouncement Player[%v] anouncement msg_type[%v] msg_param[%v]", args.FromPlayerId, args.MsgType, args.MsgParam1)
	return nil
}

// 排行榜调用
type R2H_RanklistProc struct {
}

func (this *R2H_RanklistProc) AnouncementFirstRank(args *rpc_common.R2H_RanklistPlayerFirstRank, result *rpc_common.R2H_RanklistPlayerFirstRankResult) error {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	if !anouncement_mgr.PushNew(ANOUNCEMENT_TYPE_RANKING_LIST_FIRST_RANK, false, args.PlayerId /*args.PlayerName, args.PlayerLevel, */, args.RankType, args.RankParam, 0, "") {
		err_str := fmt.Sprintf("@@@ R2H_RanklistProc::AnouncementFirstRank Push Player[%v] first rank in ranklist[%v] failed", args.PlayerId /*args.PlayerName, args.PlayerLevel,*/, args.RankType)
		return errors.New(err_str)
	}
	log.Debug("@@@ R2H_RanklistProc::AnouncementFirstRank Pushed Player[%v] is first rank in ranklist[%v]", args.PlayerId /*args.PlayerName, args.PlayerLevel,*/, args.RankType)
	return nil
}

// 初始化rpc服务
func (this *HallServer) init_rpc_service() bool {
	if this.rpc_service != nil {
		return true
	}
	this.rpc_service = &rpc.Service{}

	// 注册RPC服务
	if !this.rpc_service.Register(&R2H_PingProc{}) {
		return false
	}
	if !this.rpc_service.Register(&R2H_PlayerProc{}) {
		return false
	}
	if !this.rpc_service.Register(&R2H_PlayerStageInfoProc{}) {
		return false
	}
	if !this.rpc_service.Register(&R2H_FriendProc{}) {
		return false
	}
	if !this.rpc_service.Register(&H2H_FriendProc{}) {
		return false
	}
	if !this.rpc_service.Register(&H2H_PlayerProc{}) {
		return false
	}
	if !this.rpc_service.Register(&H2H_FosterProc{}) {
		return false
	}
	if !this.rpc_service.Register(&H2H_GlobalProc{}) {
		return false
	}
	if !this.rpc_service.Register(&R2H_RanklistProc{}) {
		return false
	}

	if this.rpc_service.Listen(config.ListenRpcServerIP) != nil {
		log.Error("监听rpc服务端口[%v]失败", config.ListenRpcServerIP)
		return false
	}
	log.Info("监听rpc服务端口[%v]成功", config.ListenRpcServerIP)
	go this.rpc_service.Serve()
	return true
}

// 反初始化rpc服务
func (this *HallServer) uninit_rpc_service() {
	if this.rpc_service != nil {
		this.rpc_service.Close()
		this.rpc_service = nil
	}
}
