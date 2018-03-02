package main

import (
	"libs/log"
	"libs/utils"
	"net/http"
	"public_message/gen_go/client_message"
	"time"
	"youma/rpc_common"
	"youma/table_config"

	"3p/code.google.com.protobuf/proto"
)

func reg_player_foster_new_msg() {
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SPullFriendFoster, C2SPullFriendFosterHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SFosterCat, C2SFosterCatHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SFosterCatFinish, C2SFosterCatFinishHandler)
	msg_handler_mgr.SetPlayerMsgHandler(msg_client_message.ID_C2SFosterIncomes, C2SFosterIncomesHandler)
}

func get_foster_card_remain_seconds_new(card_id int32, card_start_time int32) int32 {
	foster_card := foster_card_table_mgr.GetByItemId(card_id)
	if foster_card == nil {
		return 0
	}
	return GetRemainSeconds(card_start_time, foster_card.FosterTime)
}

func (this *Player) foster_cat_num_to_friend(friend_id int32) (num int32) {
	all := this.db.FosterCatOnFriends.GetAllIndex()
	for _, cid := range all {
		fid, _ := this.db.FosterCatOnFriends.GetFriendId(cid)
		if fid == friend_id {
			num += 1
		}
	}
	return
}

func (this *dbPlayerFosteredFriendCatColumn) CheckAndAddFriendCat(friend_id, cat_id, cat_table_id, cat_level, cat_star int32, foster_friend_num int32, card *table_config.XmlFosterCardItem) int32 {
	this.m_row.m_lock.UnSafeLock("dbPlayerFosterFriendCatColumn.CheckAndAddFriendCat")
	defer this.m_row.m_lock.UnSafeUnlock()

	l := len(this.m_data)
	if int32(l) >= foster_friend_num {
		log.Error("Player[%v] foster friend[%v] cat[%v] no enough space", this.m_row.m_PlayerId, friend_id, cat_id)
		return int32(msg_client_message.E_ERR_FOSTER_FRIEND_NO_SPACE_TO_FOSTER)
	}

	card_id := int32(0)
	if card != nil {
		card_id = card.ItemId
	}
	friend_cat_id := utils.Int64From2Int32(friend_id, cat_id)
	d := this.m_data[friend_cat_id]
	if d != nil {
		d.CatTableId = cat_table_id
		d.CatLevel = cat_level
		d.CatStar = cat_star
		d.StartCardId = card_id
		d.StartTime = int32(time.Now().Unix())
		log.Warn("Player[%v] foster set friend[%v] cat[%v] already exists", this.m_row.m_PlayerId, friend_id, cat_id)
		//return int32(msg_client_message.E_ERR_FOSTER_ALREADY_CAT_IN_THE_FRIEND)
	} else {
		this.m_data[friend_cat_id] = &dbPlayerFosteredFriendCatData{
			PlayerCatId: friend_cat_id,
			CatTableId:  cat_table_id,
			CatLevel:    cat_level,
			CatStar:     cat_star,
			StartCardId: card_id,
			StartTime:   int32(time.Now().Unix()),
		}
	}

	this.m_changed = true

	log.Debug("Player[%v] was fostered friend[%v] cat[%v]", this.m_row.m_PlayerId, friend_id, cat_id)

	return 1
}

// 猫放入好友的寄养所
func (this *Player) foster_cat(friend_id int32, cat_id int32, foster_card_id int32) int32 {
	// 寄养卡是否存在
	if !this.db.Items.HasIndex(foster_card_id) {
		log.Error("Player[%v] not found foster card[%v]", this.Id, foster_card_id)
		return int32(msg_client_message.E_ERR_ITEM_NOT_FOUND)
	}

	foster_card := foster_card_table_mgr.GetByItemId(foster_card_id)
	if foster_card == nil {
		log.Error("Player[%v] cant get foster card by item id[%v]", this.Id, foster_card_id)
		return int32(msg_client_message.E_ERR_FOSTER_CARD_NOT_FOUND)
	}

	// 达到寄存数上限
	if this.foster_slot_num_to_friend() <= this.db.FosterCatOnFriends.NumAll() {
		log.Error("Player[%v] cant set cat to friend[%v], no empty slot, [%v<=%v]", this.Id, friend_id, this.foster_slot_num_to_friend(), this.db.FosterCatOnFriends.NumAll())
		return int32(msg_client_message.E_ERR_FOSTER_MAX_FRIEND_NUM_TO_FOSTER)
	}

	// 猫不存在
	if !this.db.Cats.HasIndex(cat_id) {
		log.Error("Player[%v] not found cat[%v]", this.Id, cat_id)
		return int32(msg_client_message.E_ERR_CAT_NOT_FOUND)
	}

	// 猫是否忙
	if this.IfCatBusy(cat_id) {
		log.Error("Player[%v] cat[%v] is busy, cant set to friend[%v] foster", this.Id, cat_id, friend_id)
		return int32(msg_client_message.E_ERR_CAT_IS_BUSY)
	}

	friend := player_mgr.GetPlayerById(friend_id)
	if friend != nil {
		// 先结算该好友寄养所中其他玩家的猫
		friend.foster_settlement_friends_cat()
		// 放猫
		cat_table_id, _ := this.db.Cats.GetCfgId(cat_id)
		cat_level, _ := this.db.Cats.GetLevel(cat_id)
		cat_star, _ := this.db.Cats.GetStar(cat_id)
		res := friend.db.FosteredFriendCats.CheckAndAddFriendCat(this.Id, cat_id, cat_table_id, cat_level, cat_star, friend.fostered_slot_num_for_friend(), foster_card)
		if res < 0 {
			log.Error("Player[%v] set cat[%v] to local friend[%v] foster failed", this.Id, cat_id, friend_id)
			return res
		}
	} else {
		// 先结算再放猫
		result := this.rpc_foster_cat_to_friend(friend_id, cat_id)
		if result == nil {
			log.Error("Player[%v] set cat[%v] to remote friend[%v] foster failed", this.Id, cat_id, friend_id)
			return int32(msg_client_message.E_ERR_FOSTER_SET_CAT_TO_FRIEND_FAILED)
		}
	}

	this.db.FosterCatOnFriends.Add(&dbPlayerFosterCatOnFriendData{
		CatId:        cat_id,
		FriendId:     friend_id,
		FosterCardId: foster_card_id,
	})

	this.RemoveItem(foster_card_id, 1, false)

	this.SendCatUpdate(cat_id)

	response := &msg_client_message.S2CFosterSetCat2FriendResult{
		FriendId:     proto.Int32(friend_id),
		CatId:        proto.Int32(cat_id),
		FosterCardId: proto.Int32(foster_card_id),
	}
	this.Send(response)

	log.Debug("Player[%v] set cat[%v] to friend[%v] foster with card[%v]", this.Id, cat_id, friend_id, foster_card_id)

	return 1
}

func (this *Player) foster_finish_cat_on_friend(friend_id int32, cat_ids []int32) int32 {
	friend := player_mgr.GetPlayerById(friend_id)
	if friend != nil {
		for _, cat_id := range cat_ids {
			fcid := utils.Int64From2Int32(this.Id, cat_id)
			friend.db.FosteredFriendCats.Remove(fcid)
		}
	} else {
		this.rpc_foster_finish_cats(friend_id, cat_ids)
	}
	return 1
}

// 中断寄养
func (this *Player) foster_cat_finish(cat_ids []int32) int32 {
	if cat_ids == nil || len(cat_ids) == 0 {
		return 1
	}

	var finish_friend_cats map[int32][]int32
	for _, cat_id := range cat_ids {
		if !this.db.FosterCatOnFriends.HasIndex(cat_id) {
			log.Error("Player[%v] not foster cat[%v]", this.Id, cat_id)
			continue
			//return int32(msg_client_message.E_ERR_FOSTER_NO_CAT_IN_THE_FRIEND)
		}

		friend_id, _ := this.db.FosterCatOnFriends.GetFriendId(cat_id)
		this.db.FosterCatOnFriends.Remove(cat_id)

		if finish_friend_cats == nil {
			finish_friend_cats = make(map[int32][]int32)
		}
		if _, o := finish_friend_cats[friend_id]; !o {
			finish_friend_cats[friend_id] = []int32{cat_id}
		} else {
			cats := finish_friend_cats[friend_id]
			cats = append(cats, cat_id)
			finish_friend_cats[friend_id] = cats
		}
	}

	log.Debug("@@@@@@@@@@@ finish_friend_cats: %v", finish_friend_cats)
	if finish_friend_cats != nil {
		for friend_id, cat_ids := range finish_friend_cats {
			this.foster_finish_cat_on_friend(friend_id, cat_ids)
		}
	}

	response := &msg_client_message.S2CFosterCatFinishResult{
		CatIds: cat_ids,
	}
	this.Send(response)

	log.Debug("Player[%v] finish foster cats[%v]", this.Id, cat_ids)
	return 1
}

// 结算寄养所中的好友的猫
func (this *Player) foster_settlement() (out_cats map[int32][]int32) {
	all := this.db.FosteredFriendCats.GetAllIndex()
	if all == nil {
		return
	}

	for _, fcid := range all {
		if !this.db.FosteredFriendCats.HasIndex(fcid) {
			continue
		}
		card_id, _ := this.db.FosteredFriendCats.GetStartCardId(fcid)
		start_time, _ := this.db.FosteredFriendCats.GetStartTime(fcid)
		remain_seconds := get_foster_card_remain_seconds_new(card_id, start_time)
		if remain_seconds <= 0 {
			cat_table_id, _ := this.db.FosteredFriendCats.GetCatTableId(fcid)
			// 收益
			this.db.FosterIncomeForFriends.Add(&dbPlayerFosterIncomeForFriendData{
				PlayerCatId:  fcid,
				FosterCardId: card_id,
				CatTableId:   cat_table_id,
			})
			// 删除
			this.db.FosteredFriendCats.Remove(fcid)

			friend_id, cat_id := utils.TwoInt32FromInt64(fcid)
			if out_cats == nil {
				out_cats = make(map[int32][]int32)
			}
			if _, o := out_cats[friend_id]; !o {
				out_cats[friend_id] = []int32{cat_id}
			} else {
				friend_cats := out_cats[friend_id]
				friend_cats = append(friend_cats, cat_id)
				out_cats[friend_id] = friend_cats
			}
		}
	}
	return
}

// 寄养所数据格式化为消息
func (this *Player) foster_format_data_to_msg() (foster_cats []*msg_client_message.FosterCat, out_cats map[int32][]int32) {
	out_cats = this.foster_settlement()
	all := this.db.FosteredFriendCats.GetAllIndex()
	if all == nil || len(all) == 0 {
		return
	}
	for _, fcid := range all {
		friend_id, cat_id := utils.TwoInt32FromInt64(fcid)
		cat_table_id, _ := this.db.FosteredFriendCats.GetCatTableId(fcid)
		cat_level, _ := this.db.FosteredFriendCats.GetCatLevel(fcid)
		cat_star, _ := this.db.FosteredFriendCats.GetCatStar(fcid)
		cat_nick, _ := this.db.FosteredFriendCats.GetCatNick(fcid)
		card_id, _ := this.db.FosteredFriendCats.GetStartCardId(fcid)
		start_time, _ := this.db.FosteredFriendCats.GetStartTime(fcid)
		remain_seconds := get_foster_card_remain_seconds_new(card_id, start_time)

		friend_name, friend_level, friend_head := GetPlayerBaseInfo(friend_id)
		foster_cat := &msg_client_message.FosterCat{
			FriendId:      proto.Int32(friend_id),
			FriendName:    proto.String(friend_name),
			FriendLevel:   proto.Int32(friend_level),
			FriendHead:    proto.String(friend_head),
			CatId:         proto.Int32(cat_id),
			CatTableId:    proto.Int32(cat_table_id),
			CatLevel:      proto.Int32(cat_level),
			CatStar:       proto.Int32(cat_star),
			CatNick:       proto.String(cat_nick),
			FosterCardId:  proto.Int32(card_id),
			RemainSeconds: proto.Int32(remain_seconds),
		}
		foster_cats = append(foster_cats, foster_cat)
	}
	return
}

// 寄养所数据格式化为RPC协议
func (this *Player) foster_format_data_to_rpc() (foster_cats []*rpc_common.H2H_FosteredCat, out_cats map[int32][]int32) {
	out_cats = this.foster_settlement()
	all := this.db.FosteredFriendCats.GetAllIndex()
	for _, fcid := range all {
		friend_id, _ := utils.TwoInt32FromInt64(fcid)
		cat_table_id, _ := this.db.FosteredFriendCats.GetCatTableId(fcid)
		cat_level, _ := this.db.FosteredFriendCats.GetCatLevel(fcid)
		cat_star, _ := this.db.FosteredFriendCats.GetCatStar(fcid)
		cat_nick, _ := this.db.FosteredFriendCats.GetCatNick(fcid)
		card_id, _ := this.db.FosteredFriendCats.GetStartCardId(fcid)
		start_time, _ := this.db.FosteredFriendCats.GetStartTime(fcid)
		remain_seconds := get_foster_card_remain_seconds_new(card_id, start_time)

		foster_cat := &rpc_common.H2H_FosteredCat{
			StartCardId:   card_id,
			RemainSeconds: remain_seconds,
			CatTableId:    cat_table_id,
			CatLevel:      cat_level,
			CatStar:       cat_star,
			CatNick:       cat_nick,
			PlayerId:      friend_id,
		}
		foster_cats = append(foster_cats, foster_cat)
	}
	return
}

func (this *Player) remove_foster_cats_on_friends(cats map[int32][]int32) {
	for pid, cids := range cats {
		player := player_mgr.GetPlayerById(pid)
		if player != nil {
			if cids != nil {
				for _, cid := range cids {
					player.db.FosterCatOnFriends.Remove(cid)
				}
			}
		} else {
			if this.rpc_foster_clear_finished_cats(pid, cids) == nil {
				log.Error("Player[%v] remove friend[%v] finished cats[%v] failed", pid, cids)
			}
		}
	}
}

// 获取玩家的寄养所
func (this *Player) foster_data_by_player_id(player_id int32) int32 {
	if player_id == this.Id {
		return -1
	}

	player := player_mgr.GetPlayerById(player_id)
	var foster_cats []*msg_client_message.FosterCat
	var fostered_num int32
	if player != nil {
		//var out_cats map[int32][]int32
		foster_cats /*out_cats*/, _ = player.foster_format_data_to_msg()
		fostered_num = player.fostered_slot_num_for_friend()
		//this.remove_foster_cats_on_friends(out_cats)
	} else {
		result := this.rpc_foster_get_friend_cats(player_id)
		if result == nil {
			log.Error("Player[%v] get player[%v] foster data failed", this.Id, player_id)
			return -1
		}
		//this.remove_foster_cats_on_friends(result.OutFriendCats)
		for i := 0; i < len(result.Cats); i++ {
			c := result.Cats[i]
			friend_name, friend_level, friend_head := GetPlayerBaseInfo(player_id)
			d := &msg_client_message.FosterCat{
				FosterCardId:  proto.Int32(c.StartCardId),
				RemainSeconds: proto.Int32(c.RemainSeconds),
				CatTableId:    proto.Int32(c.CatTableId),
				CatNick:       proto.String(c.CatNick),
				CatLevel:      proto.Int32(c.CatLevel),
				CatStar:       proto.Int32(c.CatStar),
				FriendId:      proto.Int32(c.PlayerId),
				FriendName:    proto.String(friend_name),
				FriendLevel:   proto.Int32(friend_level),
				FriendHead:    proto.String(friend_head),
			}
			foster_cats = append(foster_cats, d)
		}
		fostered_num = result.CanFosteredSlot
	}

	response := &msg_client_message.S2CPullFriendFosterResult{
		FriendId:         proto.Int32(player_id),
		Cats:             foster_cats,
		CanFosterSlotNum: proto.Int32(fostered_num),
	}
	this.Send(response)

	log.Debug("Player[%v] get the player[%v] foster data", this.Id, player_id)

	return 1
}

// 取出某个好友寄养收益
func (this *Player) foster_get_income_from_friend(friend_id int32) (cat_ids, card_ids []int32) {
	all := this.db.FosterIncomeForFriends.GetAllIndex()
	if all == nil {
		return
	}

	for _, fcid := range all {
		if !this.db.FosterIncomeForFriends.HasIndex(fcid) {
			continue
		}

		fid, cid := utils.TwoInt32FromInt64(fcid)
		if friend_id == fid {
			card_id, _ := this.db.FosterIncomeForFriends.GetFosterCardId(fcid)
			cat_ids = append(cat_ids, cid)
			card_ids = append(card_ids, card_id)
			this.db.FosterIncomeForFriends.Remove(fcid)
		}
	}

	return
}

func (this *Player) foster_add_income_items(cat_id, card_id int32) {
	foster_card := foster_card_table_mgr.GetByItemId(card_id)
	if foster_card == nil {
		return
	}

	update_items := false
	for i := 0; i < len(foster_card.Rewards)/2; i++ {
		if foster_card.Rewards[2*i] == ITEM_RESOURCE_ID_CAT_EXP {
			this.feed_cat(cat_id, 0, foster_card.Rewards[2*i+1]*foster_card.FosterTime/3600, false)
		} else {
			if this.AddItemResource(foster_card.Rewards[2*i], foster_card.Rewards[2*i+1]*foster_card.FosterTime/3600, "foster_income", "foster") > 0 {
				update_items = true
			}
		}
	}
	if update_items {
		this.SendItemsUpdate()
		this.SendCatsUpdate()
		this.SendDepotBuildingUpdate()
	}
}

// 获取所有寄养收益
func (this *Player) foster_get_incomes() (incomes []*msg_client_message.FosterCatIncome) {
	all := this.db.FosterCatOnFriends.GetAllIndex()
	if all == nil || len(all) == 0 {
		incomes = make([]*msg_client_message.FosterCatIncome, 0)
		return
	}

	friends := make(map[int32]int32)
	for _, cid := range all {
		friend_id, _ := this.db.FosterCatOnFriends.GetFriendId(cid)
		friends[friend_id] = friend_id
	}

	for _, fid := range friends {
		var cat_ids, card_ids []int32
		friend := player_mgr.GetPlayerById(fid)
		if friend != nil {
			friend.foster_settlement()
			cat_ids, card_ids = friend.foster_get_income_from_friend(this.Id)
		} else {
			result := this.rpc_foster_get_income_from_friend(fid)
			if result == nil {
				incomes = make([]*msg_client_message.FosterCatIncome, 0)
				return
			}
			cat_ids = result.CatIds
			card_ids = result.CardIds
		}

		if cat_ids != nil && card_ids != nil && len(cat_ids) == len(card_ids) {
			friend_name, friend_level, friend_head := GetPlayerBaseInfo(fid)
			for i := 0; i < len(cat_ids); i++ {
				// 收益加到玩家身上
				this.foster_add_income_items(cat_ids[i], card_ids[i])
				this.db.FosterCatOnFriends.Remove(cat_ids[i])
				income := &msg_client_message.FosterCatIncome{
					FriendId:    proto.Int32(fid),
					FriendName:  proto.String(friend_name),
					FriendLevel: proto.Int32(friend_level),
					FriendHead:  proto.String(friend_head),
					CatId:       proto.Int32(cat_ids[i]),
					CardId:      proto.Int32(card_ids[i]),
				}
				incomes = append(incomes, income)
			}
		}
	}

	return
}

func C2SPullFriendFosterHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SPullFriendFoster)
	if req == nil {
		log.Error("C2SPullFosterDataHandler proto invalid!")
		return -1
	}
	return p.foster_data_by_player_id(req.GetFriendId())
}

func C2SFosterCatHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SFosterCat)
	if req == nil {
		log.Error("C2SPullFosterDataWithFriendHandler proto invalid")
		return -1
	}
	return p.foster_cat(req.GetFriendId(), req.GetCatId(), req.GetCardItemId())
}

func C2SFosterCatFinishHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SFosterCatFinish)
	if req == nil {
		log.Error("C2SFosterEquipCardHandler proto invalid")
		return -1
	}
	return p.foster_cat_finish(req.GetCatIds())
}

func C2SFosterIncomesHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2SFosterIncomes)
	if req == nil {
		log.Error("C2SFosterIncomesHandler proto invalid")
		return -1
	}
	response := &msg_client_message.S2CFosterIncomesResult{
		Incomes: p.foster_get_incomes(),
	}
	p.Send(response)
	return 1
}
