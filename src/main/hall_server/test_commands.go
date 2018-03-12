package main

import (
	"libs/log"
	"libs/timer"
	"libs/utils"
	"net/http"
	"public_message/gen_go/client_message"
	"strconv"
	"time"

	_ "3p/code.google.com.protobuf/proto"
	"github.com/golang/protobuf/proto"
)

func player_info_cmd(p *Player, args []string) int32 {
	log.Info("### 玩家基础信息如下：")
	log.Info("###### Name: %v", p.db.GetName())
	log.Info("###### Level: %v", p.db.Info.GetLvl())
	log.Info("###### Exp: %v", p.db.Info.GetExp())
	log.Info("###### Diamond: %v", p.db.Info.GetDiamond())
	log.Info("###### Coin: %v", p.db.Info.GetCoin())
	log.Info("###### CharmVal: %v", p.db.Info.GetCharmVal())
	log.Info("###### MaxUnlockStage: %v", p.db.Info.GetMaxUnlockStage())
	log.Info("###### CurMaxStage: %v", p.db.Info.GetCurMaxStage())
	log.Info("###### Star: %v", p.db.Info.GetTotalStars())
	log.Info("###### FriendPoints: %v", p.db.Info.GetFriendPoints())
	log.Info("###### Zan: %v", p.db.Info.GetZan())
	log.Info("###### Spirit: %v", p.CalcSpirit())
	return 0
}

func add_exp_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数数量[%v]不够", len(args))
		return -1
	}

	var exp int
	var err error
	exp, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("经验[%v]转换失败[%v]", exp, err.Error())
		return -1
	}

	p.AddExp(int32(exp), "test_add_exp", "test_command")
	return 1
}

func add_item_cmd(p *Player, args []string) int32 {
	if len(args) < 2 {
		log.Error("参数数量[%v]不够", len(args))
		return -1
	}

	item_id, err := strconv.Atoi(args[0])
	if err != nil {
		log.Error("物品ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	item := item_table_mgr.Map[int32(item_id)]
	if item == nil {
		log.Error("没有物品[%v]配置", item_id)
		return -1
	}

	item_count, err2 := strconv.Atoi(args[1])
	if err2 != nil {
		log.Error("物品数量[%v]转换失败[%v]", args[1], err2.Error())
		return -1
	}

	if item_count < 0 {
		p.RemoveItem(int32(item_id), int32(item_count), false)
	} else {
		p.AddItem(int32(item_id), int32(item_count), "test_add_item", "test_command", false)
	}
	return 1
}

func add_all_item_cmd(p *Player, args []string) int32 {
	p.add_all_items()
	p.SendItemsUpdate()
	return 1
}

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

func list_item_cmd(p *Player, args []string) int32 {
	ids := p.db.Items.GetAllIndex()
	if ids == nil || len(ids) == 0 {
		log.Warn("玩家[%v]没有物品", p.Id)
		return -1
	}
	log.Info("@@@ 玩家[%v]物品列表如下：", p.Id)
	for i, id := range ids {
		item_data := p.db.Items.Get(id)
		if item_data == nil {
			log.Warn("玩家[%v]没有物品[%v]", p.Id, id)
			continue
		}
		log.Info("@@@@@@ [%v] Id[%v] CfgId[%v] Num[%v] RemainSeconds[%v]", i, id, item_data.ItemCfgId, item_data.ItemNum, item_data.RemainSeconds)
	}
	return 0
}

func add_coin_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}
	coin, err := strconv.Atoi(args[0])
	if err != nil {
		log.Error("金币数量[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	p.AddCoin(int32(coin), "test_add_coin", "test_command")
	return 1
}

func set_coin_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}
	coin, err := strconv.Atoi(args[0])
	if err != nil {
		log.Error("金币数量[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}
	if coin < 0 {
		return -1
	}
	p.db.Info.SetCoin(int32(coin))
	return 1
}

func add_diamond_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}
	diamond, err := strconv.Atoi(args[0])
	if err != nil {
		log.Error("钻石数量[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}
	p.AddDiamond(int32(diamond), "test_add_diamond", "test_command")
	return 1
}

func set_diamond_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}
	diamond, err := strconv.Atoi(args[0])
	if err != nil {
		log.Error("钻石数量[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}
	if diamond < 0 {
		return -1
	}
	p.db.Info.SetDiamond(int32(diamond))
	return 1
}

func draw_card_cmd(p *Player, args []string) int32 {
	if len(args) < 2 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}
	draw_type, err := strconv.Atoi(args[0])
	if err != nil {
		log.Error("抽奖类型[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}
	draw_num, err2 := strconv.Atoi(args[1])
	if err2 != nil {
		log.Error("抽奖次数[%v]转换失败[%v]", args[1], err.Error())
		return -1
	}
	return p.DrawCard(int32(draw_type), int32(draw_num))
}

func drop_items_cmd(p *Player, args []string) int32 {
	if len(args)%2 != 0 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var err error
	var n int
	a := make([]int32, len(args))
	for i := 0; i < len(args); i++ {
		n, err = strconv.Atoi(args[i])
		if err != nil {
			log.Error("掉落参数[%v]错误[%v]", args[i], err.Error())
			return -1
		}
		a[i] = int32(n)
	}

	b, items := p.DropItems2([]int32(a), true)
	if !b {
		return -1
	}
	log.Debug("@@@@@@ droped items[%v]", items)
	return 1
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
	/*if p.check_shop_refresh(false) {
		log.Info("商店[%v]刷新", shop_id)
	}*/
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

func sell_item_cmd(p *Player, args []string) int32 {
	if len(args) < 2 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var item_id, item_num int
	var err error
	item_id, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("物品ID[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	item_num, err = strconv.Atoi(args[1])
	if err != nil {
		log.Error("物品[%v]数量转换失败[%v]", args[1], err.Error())
		return -1
	}

	return p.sell_item(int32(item_id), int32(item_num))
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

func add_charm_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var charm int
	var err error
	charm, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("魅力值[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	if charm < 0 {
		return p.SubCharmVal(int32(-charm), "test_add_charm", "test")
	} else {
		return p.AddCharmVal(int32(charm), "test_add_charm", "test")
	}
}

func add_zan_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	zan, err := strconv.Atoi(args[0])
	if err != nil {
		log.Error("赞[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.AddZan(int32(zan), "test_add_charm_medal", "test")
}

func add_star_cmd(p *Player, args []string) int32 {
	if len(args) < 1 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	star, err := strconv.Atoi(args[0])
	if err != nil {
		log.Error("星星数[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	return p.AddStar(int32(star), "test_add_star", "test")
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

func finish_stage_cmd(p *Player, args []string) int32 {
	if len(args) < 2 {
		log.Error("参数[%v]不够", len(args))
		return -1
	}

	var result, stage_id int
	var err error
	result, err = strconv.Atoi(args[0])
	if err != nil {
		log.Error("通关结果[%v]转换失败[%v]", args[0], err.Error())
		return -1
	}

	stage_id, err = strconv.Atoi(args[1])
	if err != nil {
		log.Error("关卡ID[%v]转换失败[%v]", args[1], err.Error())
		return -1
	}

	var stars = 3
	if len(args) >= 3 {
		stars, err = strconv.Atoi(args[2])
		if nil != err {
			log.Info("填写的星星数有问题[%s]")
			stars = 3
		}
	}

	var d StageBeginData
	d.stage_id = int32(stage_id)
	p.CheckBeginStage(&d)

	return p.stage_pass(int32(result), int32(stage_id), int32(99999), int32(stars), make([]*msg_client_message.ItemInfo, 0), true)
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

type test_cmd_func func(*Player, []string) int32

var test_cmd2funcs = map[string]test_cmd_func{
	"player_info":           player_info_cmd,
	"add_exp":               add_exp_cmd,
	"add_item":              add_item_cmd,
	"add_all_item":          add_all_item_cmd,
	"use_item":              use_item_cmd,
	"list_item":             list_item_cmd,
	"add_coin":              add_coin_cmd,
	"set_coin":              set_coin_cmd,
	"add_diamond":           add_diamond_cmd,
	"set_diamond":           set_diamond_cmd,
	"add_friendpoints":      add_friend_points_cmd,
	"add_charm":             add_charm_cmd,
	"add_zan":               add_zan_cmd,
	"add_star":              add_star_cmd,
	"draw_card":             draw_card_cmd,
	"drop_items":            drop_items_cmd,
	"shop_items":            get_shop_items_cmd,
	"refresh_shop":          refresh_shop_cmd,
	"buy_item":              buy_shop_item_cmd,
	"sell_item":             sell_item_cmd,
	"get_dailys":            get_dailys_cmd,
	"get_achieves":          get_achieves_cmd,
	"complete_task":         complete_task_cmd,
	"daily_reward":          get_daily_reward_cmd,
	"achieve_reward":        get_achieve_reward_cmd,
	"search_friend":         search_friend_id_cmd,
	"search_friend_name":    search_friend_name_cmd,
	"add_friend":            add_friend_cmd,
	"agree_friend":          agree_friend_cmd,
	"refuse_friend":         refuse_friend_cmd,
	"remove_friend":         remove_friend_cmd,
	"get_friends":           get_friends_cmd,
	"get_friend_info":       get_friend_info_cmd,
	"give_friend_points":    give_friend_points_cmd,
	"get_friend_points":     get_friend_points_cmd,
	"friend_chat":           friend_chat_cmd,
	"friend_unread_num":     friend_get_unread_message_num_cmd,
	"friend_unread":         friend_pull_unread_cmd,
	"friend_confirm_unread": friend_confirm_unread_cmd,
	"send_test_mail":        send_test_mail,
	"finish_stage":          finish_stage_cmd,
	"act_test":              activity_finished_cmd,
	"rank_test":             rank_test_cmd,
	"rank_test2":            rank_test2_cmd,
	"ranklist":              ranklist_cmd,
	"world_chat":            world_chat_cmd,
	"pull_world_chat":       pull_world_chat_cmd,
	"push_sysmsg":           push_sysmsg_cmd,
	"push_sysmsg_text":      push_sysmsg_text_cmd,
	"reset_day_sign_reward": reset_day_sign_reward,
}

func C2STestCommandHandler(w http.ResponseWriter, r *http.Request, p *Player, msg proto.Message) int32 {
	req := msg.(*msg_client_message.C2S_TEST_COMMAND)
	if p == nil || req == nil {
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
