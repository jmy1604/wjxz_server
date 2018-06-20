package main

import (
	"libs/log"
	_ "libs/timer"
	_ "libs/utils"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	"strconv"
	_ "time"

	_ "3p/code.google.com.protobuf/proto"
	"github.com/golang/protobuf/proto"
	_ "github.com/yuin/gopher-lua"
)

/*

func use_item_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数数量[%v]不够", len(args))
		return -1
	}

	var item_id int
	var err error
	item_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("物品ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	item := item_table_mgr.Map[int32(item_id)]
	if item == nil {
		log.Error("没有物品[%v]配置", item_id)
		return -1
	}

	item_num := 1
	if len(args) > 1 {
		item_num, err = strconv.Atoi(args[1])
		if err != nil {
			log.Error("物品数量[%v]转换失败[%v]", args[1], err.Error())
			return -1
		}
	}
	return p.use_item(int32(item_id), int32(item_num))
}

func get_shop_items_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	shop_id, err := strconv.Atoi(args[0])
	if err != nil {
		log.Error("商店配置ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.fetch_shop_limit_items(int32(shop_id), true)
}

func buy_shop_item_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	item_id, err := strconv.Atoi(args[0])
	if err != nil {
		log.Error("商品配置[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	if p.check_shop_limited_days_items_refresh_by_shop_itemid(int32(item_id), true) {
		log.Info("商店刷新")
	}

	return p.buy_item(int32(item_id), 1, true)
}

func refresh_shop_cmd(p *Player, args []string) int32 {
	if p.check_all_shop_items_refresh(true) {
		return 1
	}
	return -1
}

func add_friend_points_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var friend_points int
	var err error
	friend_points, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("友情点[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.AddFriendPoints(int32(friend_points), "test_add_friend_points", "test")
}

func get_dailys_cmd(p *Player, args []string) int32 {
	return p.get_daily_task_info()
}

func get_achieves_cmd(p *Player, args []string) int32 {
	return p.get_achieve_info()
}

func complete_task_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var task_id int
	var err error
	task_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("任务ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.complete_task(int32(task_id))
}

func get_daily_reward_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var task_id int
	var err error
	task_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("日常任务ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.get_daily_reward(int32(task_id))
}

func get_achieve_reward_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var task_id int
	var err error
	task_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("成就ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}
	return p.get_achieve_reward(int32(task_id))
}

func search_friend_id_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var friend_id int
	var err error
	friend_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("好友ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.search_friend_by_id(int32(friend_id))
}

func search_friend_name_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	return p.search_friend(args[0])
}

func add_friend_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var friend_id int
	var err error
	friend_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("好友ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.add_friend_by_id(int32(friend_id))
}

func agree_friend_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var friend_id int
	var err error
	friend_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("好友ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.agree_add_friend(int32(friend_id))
}

func refuse_friend_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var friend_id int
	var err error
	friend_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("好友ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.refuse_add_friend(int32(friend_id))
}

func remove_friend_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var friend_id int
	var err error
	friend_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("好友[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.remove_friend(int32(friend_id))
}

func get_friends_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var has_foster int
	var err error
	has_foster, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("是否寄养[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	is_foster := false
	if has_foster > 0 {
		is_foster = true
	}
	return p.get_friend_list(is_foster)
}

func get_friend_info_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var friend_id int
	var err error
	friend_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("好友[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	result := p.rpc_get_friend_info2(int32(friend_id))
	if result == nil {
		return -1
	}
	return 1
}

func give_friend_points_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	friend_id := make([]int32, len(args))
	var err error
	var fid int
	for i, _ := range args {
		fid, err = strconv.Atoi(args[i])
		if err != nil {
			log.Error("好友ID[%v]转换失败[%v]", args[i], err.Error())
			return -1
		}
		friend_id[i] = int32(fid)
	}
	return p.give_friend_points(friend_id)
}

func get_friend_points_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	friend_id := make([]int32, len(args))
	var err error
	var fid int
	for i, _ := range args {
		fid, err = strconv.Atoi(args[i])
		if err != nil {
			log.Error("好友ID[%v]转换失败[%v]", args[i], err.Error())
			return -1
		}
		friend_id[i] = int32(fid)
	}
	return p.get_friend_points(friend_id)
}

func friend_chat_cmd(p *Player, args []string) int32 {
	if len(args) < 2 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var fid int
	fid, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("好友ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.friend_chat(int32(fid), []byte(args[1]))
}

func friend_get_unread_message_num_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var fid int
	fid, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("好友ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.friend_get_unread_message_num([]int32{int32(fid)})
}

func friend_pull_unread_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var fid int
	fid, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("好友ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.friend_pull_unread_message(int32(fid))
}

func friend_confirm_unread_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var fid int
	fid, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("好友Id[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	num := 0
	if len(args) > 1 {
		num, err = strconv.Atoi(args[1])
		if err != nil {
			log.Error("确认未读消息数目[%v]转换失败[%v]", args[1], err.Error())
			return -1
		}
	}

	return p.friend_confirm_unread_message(int32(fid), int32(num))
}

func send_test_mail(p *Player, args []string) int32 {
	tmp_len := int32(len(args))
	if tmp_len < 2 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	mail_type, err := strconv.Atoi(args[0])
	if nil != err {
		log.Error("转换邮件类型[%s]失败[%s]", args[0], err.Error())
		return -1
	}

	sender_id, err := strconv.Atoi(args[1])
	if nil != err {
		log.Error("转换发送者Id[%s]失败[%s] !", err.Error())
		return -1
	}

	sender_name := args[2]

	var obj_ids, obj_nums []int32
	if tmp_len >= 5 {
		var obj_len, obj_id, obj_num int
		obj_len = (int(tmp_len) - 3) / 2
		obj_ids = make([]int32, 0, obj_len)
		obj_nums = make([]int32, 0, obj_len)
		for idx := 0; idx < obj_len; idx++ {
			obj_id, err = strconv.Atoi(args[idx*2+3])
			if nil != err {
				log.Error("转换对象Id[%s]失败[%s]", args[idx*2+3])
				return -1
			}

			obj_num, err = strconv.Atoi(args[idx*2+4])
			if nil != err {
				log.Error("转换对象Num[%s]失败[%s]", args[idx*2+4])
				return -1
			}

			obj_ids = append(obj_ids, int32(obj_id))
			obj_nums = append(obj_nums, int32(obj_num))
		}
	}

	p.SendTestMail(int32(mail_type), int32(sender_id), sender_name, obj_ids, obj_nums)

	return 1
}

func activity_finished_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var act_type int
	var err error
	act_type, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("活动测试[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	switch act_type {
	case PLAYER_ACTIVITY_TYPE_FIRST_PAY:
		{
			if len(args) < 2 {
				log.Error("PLAYER_ACTIVITY_TYPE_FIRST_PAY 参数[%v]不够", args)
				return -1
			}

			state, err := strconv.Atoi(args[1])
			if nil != err {
				log.Error("PLAYER_ACTIVITY_TYPE_FIRST_PAY 转换 参数[%v]失败[%s]", args[1], err.Error())
				return -1
			}

			log.Info("设置首冲状态%d", state)
			p.db.Info.SetFirstPayState(int32(state))

			p.OnActivityValAdd(PLAYER_ACTIVITY_TYPE_FIRST_PAY, 1)
		}
	case PLAYER_ACTIVITY_TYPE_VIP_CARD:
		{
			p.AddMonthCard(30)
		}

	}

	return 1
}

func rank_test_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var count int
	count, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换节点数量[%v]错误[%v]", args[0], err.Error())
		return -1
	}
	utils.SkiplistTest(int32(count))
	return 1
}

func rank_test2_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var count int
	count, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换节点数量[%v]错误[%v]", args[0], err.Error())
		return -1
	}
	utils.SkiplistTest2(int32(count))
	return 1
}

func ranklist_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var rank_type int
	rank_type, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换排行榜类型[%v]错误[%v]", args[0], err.Error())
		return -1
	}

	var param, rank_start, rank_num int
	if len(args) > 1 {
		rank_start, err = strconv.Atoi(args[1])
		if err != nil {
			log.Error("转换开始排名[%v]错误[%v]", args[1], err.Error())
			return -1
		}
	}

	if len(args) > 2 {
		rank_num, err = strconv.Atoi(args[2])
		if err != nil {
			log.Error("转换排名数[%v]错误[%v]", args[2], err.Error())
			return -1
		}
	}

	if len(args) > 3 {
		param, err = strconv.Atoi(args[3])
		if err != nil {
			log.Error("转换排名参数[%v]错误[%v]", args[3], err.Error())
			return -1
		}
	}

	if rank_type == 1 {
		return p.get_stage_total_score_rank_list(int32(rank_start), int32(rank_num))
	} else if rank_type == 2 {
		return p.get_stage_score_rank_list(int32(param), int32(rank_start), int32(rank_num))
	} else if rank_type == 3 {
		return p.get_charm_rank_list(int32(rank_start), int32(rank_num))
	} else if rank_type == 4 {

	} else if rank_type == 5 {
		return p.get_zaned_rank_list(int32(rank_start), int32(rank_num))
	} else {
		log.Error("rank_type[%v] is invalid")
		return -1
	}
	return 1
}

func world_chat_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	return p.world_chat([]byte(args[0]))
}

func pull_world_chat_cmd(p *Player, args []string) int32 {
	return p.pull_world_chat()
}

func push_sysmsg_cmd(p *Player, args []string) int32 {
	if len(args) < 2 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var msg_type, param int
	msg_type, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换公告类型[%v]错误[%v]", args[0], err.Error())
		return -1
	}

	param, err = strconv.Atoi(args[1])
	if err != nil {
		log.Error("转换参数[%v]错误[%v]", args[1], err.Error())
		return -1
	}

	if !anouncement_mgr.PushNew(int32(msg_type), true, p.Id, int32(param), 0, 0, "") {
		return -1
	}

	return 1
}

func push_sysmsg_text_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	if !anouncement_mgr.PushNew(ANOUNCEMENT_TYPE_TEXT, true, p.Id, 0, 0, 0, args[0]) {
		return -1
	}

	return 1
}

func reset_day_sign_reward(p *Player, args []string) int32 {
	if nil == p {
		log.Error("reset_day_sign_reward p nil !")
		return -1
	}

	cur_month_day := int32(time.Now().Day())
	p.db.Activitys.SetStates0(2001, PLAYER_ACTIVITY_STATE_FINISHED)
	p.db.Activitys.RemoveValsVal(2001, cur_month_day)

	res2cil := p.db.Activitys.FillAllClientMsg(p.db.Info.GetVipCardEndDay() - timer.GetDayFrom1970WithCfg(0))
	if nil != res2cil {
		p.Send(res2cil)
		return 1
	}

	return 1
}
*/

func test_lua_cmd(p *Player, args []string) int32 {
	/*L := lua.NewState(lua.Options{SkipOpenLibs: true})
	defer L.Close()
	for _, pair := range []struct {
		n string
		f lua.LGFunction
	}{
		{lua.LoadLibName, lua.OpenPackage}, // Must be first
		{lua.BaseLibName, lua.OpenBase},
		{lua.TabLibName, lua.OpenTable},
	} {
		if err := L.CallByParam(lua.P{
			Fn:      L.NewFunction(pair.f),
			NRet:    0,
			Protect: true,
		}, lua.LString(pair.n)); err != nil {
			panic(err)
		}
	}
	if err := L.DoFile("main.lua"); err != nil {
		panic(err)
	}*/
	return 1
}

func rand_role_cmd(p *Player, args []string) int32 {
	role_id := p.rand_role()
	if role_id <= 0 {
		log.Warn("Cant rand role")
	} else {
		log.Debug("Rand role: %v", role_id)
	}
	return 1
}

func new_role_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var table_id, num int
	var err error
	table_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换角色配置ID[%v]错误[%v]", args[0], err.Error())
		return -1
	}

	if len(args) > 1 {
		num, err = strconv.Atoi(args[1])
		if err != nil {
			log.Error("转换角色数量[%v]错误[%v]", args[1], err.Error())
			return -1
		}
	}

	if num == 0 {
		num = 1
	}
	for i := 0; i < num; i++ {
		p.new_role(int32(table_id), 1, 1)
	}
	return 1
}

func list_role_cmd(p *Player, args []string) int32 {
	var camp, typ, star int
	var err error
	if len(args) > 0 {
		camp, err = strconv.Atoi(args[0])
		if err != nil {
			log.Error("转换阵营[%v]错误[%v]", args[0], err.Error())
			return -1
		}
		if len(args) > 1 {
			typ, err = strconv.Atoi(args[1])
			if err != nil {
				log.Error("转换卡牌类型[%v]错误[%v]", args[1], err.Error())
				return -1
			}
			if len(args) > 2 {
				star, err = strconv.Atoi(args[2])
				if err != nil {
					log.Error("转换卡牌星级[%v]错误[%v]", args[2], err.Error())
					return -1
				}
			}
		}
	}
	all := p.db.Roles.GetAllIndex()
	if all != nil {
		for i := 0; i < len(all); i++ {
			table_id, o := p.db.Roles.GetTableId(all[i])
			if !o {
				continue
			}

			level, _ := p.db.Roles.GetLevel(all[i])
			rank, _ := p.db.Roles.GetRank(all[i])

			card := card_table_mgr.GetRankCard(table_id, rank)
			if card == nil {
				continue
			}

			if camp > 0 && card.Camp != int32(camp) {
				continue
			}
			if typ > 0 && card.Type != int32(typ) {
				continue
			}
			if star > 0 && card.Rarity != int32(star) {
				continue
			}
			log.Debug("role_id:%v, table_id:%v, level:%v, rank:%v, camp:%v, type:%v, star:%v", all[i], table_id, level, rank, card.Camp, card.Type, card.Rarity)
		}
	}
	return 1
}

func create_battle_team_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var battle_type int
	battle_type, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换阵型类型[%v]错误[%v]", args[0], err.Error())
		return -1
	}

	if battle_type == 0 {

	}

	return 1
}

func set_attack_team_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var role_id int
	var team []int32
	for i := 0; i < len(args); i++ {
		role_id, err = strconv.Atoi(args[i])
		if err != nil {
			log.Error("转换角色ID[%v]错误[%v]", role_id, err.Error())
			return -1
		}
		team = append(team, int32(role_id))
	}

	if p.SetAttackTeam(team) < 0 {
		log.Error("设置玩家[%v]攻击阵容失败", p.Id)
		return -1
	}

	return 1
}

func set_defense_team_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var role_id int
	var team []int32
	for i := 0; i < len(args); i++ {
		role_id, err = strconv.Atoi(args[i])
		if err != nil {
			log.Error("转换角色ID[%v]错误[%v]", role_id, err.Error())
			return -1
		}
		team = append(team, int32(role_id))
	}

	if p.SetDefenseTeam(team) < 0 {
		log.Error("设置玩家[%v]防守阵容失败", p.Id)
		return -1
	}

	return 1
}

func list_teams_cmd(p *Player, args []string) int32 {
	log.Debug("attack team: %v", p.db.BattleTeam.GetAttackMembers())
	log.Debug("defense team: %v", p.db.BattleTeam.GetDefenseMembers())
	log.Debug("campaign team: %v", p.db.BattleTeam.GetCampaignMembers())
	return 1
}

func pvp_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var player_id int
	player_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换玩家ID[%v]失败[%v]", args[0], err.Error())
		return -1
	}

	player := player_mgr.GetPlayerById(int32(player_id))
	if player == nil {
		log.Error("玩家[%v]找不到", player_id)
		return -1
	}

	p.Fight2Player(int32(player_id))

	log.Debug("玩家[%v]pvp玩家[%v]", p.Id, player.Id)
	return 1
}

func fight_stage_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var stage_id int
	stage_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换关卡[%v]失败[%v]", args[0], err.Error())
		return -1
	}

	stage := stage_table_mgr.Get(int32(stage_id))
	if stage == nil {
		log.Error("关卡[%v]不存在", stage_id)
		return -1
	}
	is_win, my_team, target_team, enter_reports, rounds, has_next_wave := p.FightInStage(1, stage)
	response := &msg_client_message.S2CBattleResultResponse{}
	response.IsWin = is_win
	response.MyTeam = my_team
	response.TargetTeam = target_team
	response.EnterReports = enter_reports
	response.Rounds = rounds
	response.HasNextWave = has_next_wave
	p.Send(uint16(msg_client_message_id.MSGID_S2C_BATTLE_RESULT_RESPONSE), response)
	Output_S2CBattleResult(p, response)
	log.Debug("玩家[%v]挑战了关卡[%v]", p.Id, stage_id)
	return 1
}

func fight_campaign_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var campaign_id int
	campaign_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换关卡ID[%v]失败[%v]", args[0], err.Error())
		return -1
	}

	res := p.FightInCampaign(int32(campaign_id))
	if res < 0 {
		log.Error("玩家[%v]挑战战役关卡[%v]失败[%v]", p.Id, campaign_id, res)
	} else {
		log.Debug("玩家[%v]挑战了战役关卡[%v]", p.Id, campaign_id)
	}
	return res
}

func start_hangup_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var campaign_id int
	campaign_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换战役ID[%v]失败[%v]", args[0], err.Error())
		return -1
	}

	res := p.set_hangup_campaign_id(int32(campaign_id))
	if res < 0 {
		return res
	}

	log.Debug("玩家[%v]设置了挂机战役关卡[%v]", p.Id, campaign_id)
	return 1
}

func hangup_income_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var income_type int
	income_type, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换收益类型[%v]失败[%v]", args[0], err.Error())
		return -1
	}

	p.hangup_income_get(int32(income_type), false)

	log.Debug("玩家[%v]获取了类型[%v]挂机收益", p.Id, income_type)

	return 1
}

func campaign_data_cmd(p *Player, args []string) int32 {
	p.send_campaigns()
	return 1
}

func leave_game_cmd(p *Player, args []string) int32 {
	p.OnLogout()
	return 1
}

func add_item_cmd(p *Player, args []string) int32 {
	if len(args) < 2 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var item_id, item_num int
	item_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换物品ID[%v]失败[%v]", args[0], err.Error())
		return -1
	}
	item_num, err = strconv.Atoi(args[1])
	if err != nil {
		log.Error("转换物品数量[%v]失败[%v]", args[1], err.Error())
		return -1
	}

	if !p.add_resource(int32(item_id), int32(item_num)) {
		return -1
	}

	log.Debug("玩家[%v]增加了资源[%v,%v]", p.Id, item_id, item_num)
	return 1
}

func role_levelup_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var role_id int
	role_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换角色ID[%v]失败[%v]", args[0], err.Error())
		return -1
	}

	res := p.levelup_role(int32(role_id))
	if res > 0 {
		log.Debug("玩家[%v]升级了角色[%v]等级[%v]", p.Id, role_id, res)
	}

	return res
}

func role_rankup_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var role_id int
	role_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换角色ID[%v]失败[%v]", args[0], err.Error())
		return -1
	}

	res := p.rankup_role(int32(role_id))
	if res > 0 {
		log.Debug("玩家[%v]升级了角色[%v]品阶[%v]", p.Id, role_id, res)
	}

	return res
}

func role_decompose_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var role_id int
	role_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换角色ID[%v]失败[%v]", args[0], err.Error())
		return -1
	}

	res := p.decompose_role(int32(role_id))
	if res > 0 {
		log.Debug("玩家[%v]分解了角色[%v]", p.Id, role_id)
	}

	return res
}

func item_fusion_cmd(p *Player, args []string) int32 {
	if len(args) < 2 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var piece_id, fusion_num int
	piece_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换碎片ID[%v]失败[%v]", args[0], err.Error())
		return -1
	}
	fusion_num, err = strconv.Atoi(args[1])
	if err != nil {
		log.Error("转换合成次数[%v]失败[%v]", args[1], err.Error())
		return -1
	}

	return p.fusion_item(int32(piece_id), int32(fusion_num))
}

func item_sell_cmd(p *Player, args []string) int32 {
	if len(args) < 2 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var item_id, item_num int
	item_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换物品ID[%v]失败[%v]", args[0], err.Error())
		return -1
	}
	item_num, err = strconv.Atoi(args[1])
	if err != nil {
		log.Error("转换物品数量[%v]失败[%v]", args[1], err.Error())
		return -1
	}

	return p.sell_item(int32(item_id), int32(item_num))
}

func fusion_role_cmd(p *Player, args []string) int32 {
	if len(args) < 3 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var fusion_id, main_card_id int
	fusion_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("转换合成角色ID[%v]失败[%v]", args[0], err.Error())
		return -1
	}
	main_card_id, err = strconv.Atoi(args[1])
	if err != nil {
		log.Error("转换主卡ID[%v]失败[%v]", args[1], err.Error())
		return -1
	}

	var cost1_ids, cost2_ids, cost3_ids []int32
	cost1_ids = parse_xml_str_arr2(args[2], "|")
	if cost1_ids == nil || len(cost1_ids) == 0 {
		log.Error("消耗角色1系列转换错误")
		return -1
	}
	if len(args) > 3 {
		cost2_ids = parse_xml_str_arr2(args[3], "|")
		if cost2_ids == nil || len(cost2_ids) == 0 {
			log.Error("消耗角色2系列转换错误")
			return -1
		}
		if len(args) > 4 {
			cost3_ids = parse_xml_str_arr2(args[4], "|")
			if cost3_ids == nil || len(cost3_ids) == 0 {
				log.Error("消耗角色3系列转换错误")
				return -1
			}
		}
	}

	return p.fusion_role(int32(fusion_id), int32(main_card_id), [][]int32{cost1_ids, cost2_ids, cost3_ids})
}

func send_mail_cmd(p *Player, args []string) int32 {
	if len(args) < 4 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var receiver_id, mail_type int
	//var title, content string
	var err error
	receiver_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("接收者ID[%v]转换失败[%v]", receiver_id, err.Error())
		return -1
	}
	mail_type, err = strconv.Atoi(args[1])
	if err != nil {
		log.Error("邮件类型[%v]转换失败[%v]", mail_type, err.Error())
		return -1
	}

	return SendMail(p, int32(receiver_id), int32(mail_type), args[2], args[3], nil)
}

func mail_list_cmd(p *Player, args []string) int32 {
	return p.GetMailList()
}

func mail_detail_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	mail_ids := parse_xml_str_arr2(args[0], "|")
	return p.GetMailDetail(mail_ids)
}

func mail_items_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var mail_ids []int32
	for i := 0; i < len(args); i++ {
		var mail_id int
		var err error
		mail_id, err = strconv.Atoi(args[0])
		if err != nil {
			log.Error("邮件ID[%v]转换失败[%v]", args[0], err.Error())
			return -1
		}
		mail_ids = append(mail_ids, int32(mail_id))
	}

	return p.GetMailAttachedItems(mail_ids)
}

func delete_mail_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	mail_ids := parse_xml_str_arr2(args[0], "|")
	return p.DeleteMails(mail_ids)
}

func up_talent_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var talent_id int
	var err error
	talent_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("天赋ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.up_talent(int32(talent_id))
}

func tower_data_cmd(p *Player, args []string) int32 {
	return p.send_tower_data(true)
}

func fight_tower_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var tower_id int
	var err error
	tower_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("爬塔ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.fight_tower(int32(tower_id))
}

func get_tower_key_cmd(p *Player, args []string) int32 {
	tower_key_max := global_config_mgr.GetGlobalConfig().TowerKeyMax
	p.db.TowerCommon.SetKeys(tower_key_max)
	return p.send_tower_data(false)
}

func tower_records_info_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var tower_id int
	var err error
	tower_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("爬塔ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.get_tower_records_info(int32(tower_id))
}

func tower_record_data_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var tower_fight_id int
	var err error
	tower_fight_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("爬塔战斗ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.get_tower_record_data(int32(tower_fight_id))
}

func tower_ranklist_cmd(p *Player, args []string) int32 {
	rank_list := tower_ranking_list.player_list[:tower_ranking_list.player_num]
	log.Debug("TowerRankList: %v", rank_list)
	return 1
}

func battle_recordlist_cmd(p *Player, args []string) int32 {
	return p.GetBattleRecordList()
}

func battle_record_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var record_id int
	var err error
	record_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("战斗录像ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.GetBattleRecord(int32(record_id))
}

type test_cmd_func func(*Player, []string) int32

var test_cmd2funcs = map[string]test_cmd_func{
	"test_lua":           test_lua_cmd,
	"rand_role":          rand_role_cmd,
	"new_role":           new_role_cmd,
	"list_role":          list_role_cmd,
	"set_attack_team":    set_attack_team_cmd,
	"set_defense_team":   set_defense_team_cmd,
	"list_teams":         list_teams_cmd,
	"pvp":                pvp_cmd,
	"fight_stage":        fight_stage_cmd,
	"fight_campaign":     fight_campaign_cmd,
	"start_hangup":       start_hangup_cmd,
	"hangup_income":      hangup_income_cmd,
	"campaign_data":      campaign_data_cmd,
	"leave_game":         leave_game_cmd,
	"add_item":           add_item_cmd,
	"role_levelup":       role_levelup_cmd,
	"role_rankup":        role_rankup_cmd,
	"role_decompose":     role_decompose_cmd,
	"item_fusion":        item_fusion_cmd,
	"fusion_role":        fusion_role_cmd,
	"item_sell":          item_sell_cmd,
	"send_mail":          send_mail_cmd,
	"mail_list":          mail_list_cmd,
	"mail_detail":        mail_detail_cmd,
	"mail_items":         mail_items_cmd,
	"delete_mail":        delete_mail_cmd,
	"up_talent":          up_talent_cmd,
	"tower_data":         tower_data_cmd,
	"get_tower_key":      get_tower_key_cmd,
	"fight_tower":        fight_tower_cmd,
	"tower_records_info": tower_records_info_cmd,
	"tower_record_data":  tower_record_data_cmd,
	"tower_ranklist":     tower_ranklist_cmd,
	"battle_recordlist":  battle_recordlist_cmd,
	"battle_record":      battle_record_cmd,
}

func C2STestCommandHandler(w http.ResponseWriter, r *http.Request, p *Player /*msg proto.Message*/, msg_data []byte) int32 {
	var req msg_client_message.C2S_TEST_COMMAND
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("client_msg_handler unmarshal sub msg failed err(%s) !", err.Error())
		return -1
	}

	cmd := req.GetCmd()
	args := req.GetArgs()
	res := int32(0)

	fun := test_cmd2funcs[cmd]
	if fun != nil {
		res = fun(p, args)
	} else {
		log.Warn("不支持的测试命令[%v]", cmd)
	}

	return res
}
