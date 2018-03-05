package main

import (
	"libs/log"
	"time"
)

type RedisGlobalData struct {
	inited             bool
	nick_id_set        *NickIdSet
	shop_limited_items *ShopLimitedItemSet
	redis_conn         *RedisConn // redis连接
}

var global_data RedisGlobalData

func (this *RedisGlobalData) Init() bool {
	this.redis_conn = &RedisConn{}
	if this.redis_conn == nil {
		log.Error("redis客户端未初始化")
		return false
	}

	if !this.redis_conn.Connect(rpc_config.RedisServerIP) {
		return false
	}

	// 昵称集合生成
	this.nick_id_set = &NickIdSet{}
	if !this.nick_id_set.Init() {
		return false
	}

	// 商店限时商品
	this.shop_limited_items = &ShopLimitedItemSet{}
	if !this.shop_limited_items.Init() {
		return false
	}

	/*--------------- 排行榜 --------------*/
	// 关卡总分
	if this.LoadStageTotalScoreRankItems() < 0 {
		return false
	}
	// 关卡积分
	if this.LoadStageScoreRankItems() < 0 {
		return false
	}
	// 魅力值
	if this.LoadCharmRankItems() < 0 {
		return false
	}
	// 欧气值
	if this.LoadCatOuqiRankItems() < 0 {
		return false
	}
	// 被赞
	if this.LoadZanedRankItems() < 0 {
		return false
	}
	/*--------------------------------------*/

	// 个人空间
	ps_mgr.Init()
	ps_leave_messages_mgr.Init()
	ps_pic_mgr.Init()
	ps_pic_leave_messages_mgr.Init()
	ps_pic_zan_mgr.Init()

	if this.LoadPersonalSpaceBaseData() < 0 {
		return false
	}
	if this.LoadPersonalSpacePictures() < 0 {
		return false
	}
	if this.LoadPersonalSpaceLeaveMessages() < 0 {
		return false
	}
	if this.LoadPersonalSpacePicLeaveMessages() < 0 {
		return false
	}
	if this.LoadPersonalSpacePicZan() < 0 {
		return false
	}

	this.inited = true
	log.Info("全局数据GlobalData载入完成")
	return true
}

func (this *RedisGlobalData) Close() {
	this.redis_conn.Close()
}

func (this *RedisGlobalData) RunRedis() {
	this.redis_conn.Run(1000)
}

func (this *RedisGlobalData) AddIdNick(id int32, nick string) bool {
	if !this.nick_id_set.AddIdNick(id, nick) {
		return false
	}

	err := this.redis_conn.Post("HSET", ID_NICK_SET, id, nick)
	if err != nil {
		log.Error("redis增加集合[%v]数据[%v,%v]错误[%v]", ID_NICK_SET, id, nick, err.Error())
		return false
	}

	log.Debug("加入昵称[%v]成功", nick)

	return true
}

func (this *RedisGlobalData) RenameNick(player_id int32, new_nick string) int32 {
	errcode := this.nick_id_set.RenameNick(player_id, new_nick)
	if errcode < 0 {
		return errcode
	}

	/*err := this.redis_conn.Post("HDEL", ID_NICK_SET, player_id)
	if err != nil {
		log.Error("redis删除集合[%v]数据[%v]错误[%v]", ID_NICK_SET, old_nick, err.Error())
		return -1
	}*/
	err := this.redis_conn.Post("HSET", ID_NICK_SET, player_id, new_nick)
	if err != nil {
		log.Error("redis增加集合[%v]数据[%v,%v]错误[%v]", ID_NICK_SET, player_id, new_nick, err.Error())
		return -1
	}
	log.Info("修改昵称到[%v]成功", new_nick)
	return 1
}

func (this *RedisGlobalData) GetNickById(id int32) (nick string, ok bool) {
	return this.nick_id_set.GetNickById(id)
}

func (this *RedisGlobalData) GetIdsByNick(nick string) []int32 {
	return this.nick_id_set.GetIdsByNick(nick)
}

func (this *RedisGlobalData) GetShopLimitedItemLeftNum(item_id int32) (bool, int32) {
	return this.shop_limited_items.GetItemLeftNum(item_id)
}

func (this *RedisGlobalData) GetShopLimitedItemSaveTime(days int32) (bool, int32) {
	return this.shop_limited_items.GetSaveTime(days)
}

func (this *RedisGlobalData) SetShopLimitedItemLeftNum(item_id int32) bool {
	id2num := this.shop_limited_items.Refresh()
	if id2num == nil || len(id2num) == 0 {
		return true
	}
	for id, num := range id2num {
		err := this.redis_conn.Post("HSET", SHOP_LIMITED_ITEM_SET, id, num)
		if err != nil {
			log.Error("redis修改集合[%v]数据[%v,%v]错误[%v]", SHOP_LIMITED_ITEM_SET, id, num, err.Error())
			return false
		}
	}
	return true
}

func (this *RedisGlobalData) AddShopLimitedItem(item_id int32, num int32) int32 {
	num = this.shop_limited_items.AddItem(item_id, num)
	if num < 0 {
		return -1
	}
	err := this.redis_conn.Post("HSET", SHOP_LIMITED_ITEM_SET, item_id, num)
	if err != nil {
		log.Error("redis添加集合[%v]数据[%v,%v]错误[%v]", SHOP_LIMITED_ITEM_SET, item_id, num, err.Error())
		return -1
	}
	log.Info("添加redis集合[%v]数据[%v,%v]成功", SHOP_LIMITED_ITEM_SET, item_id, num)
	return num
}

func (this *RedisGlobalData) BuyShopLimitedItem(item_id, item_num int32) int32 {
	b, left_num := this.shop_limited_items.RemoveItem(item_id, item_num)
	if !b {
		log.Error("redis集合[%v]不存在商品[%v]", SHOP_LIMITED_ITEM_SET, item_id)
		return -1
	}
	err := this.redis_conn.Post("HSET", SHOP_LIMITED_ITEM_SET, item_id, left_num)
	if err != nil {
		log.Error("redis设置集合[%v]数据[%v,%v]错误[%v]", SHOP_LIMITED_ITEM_SET, item_id, left_num, err.Error())
		return -1
	}
	log.Info("设置redis集合[%v]数据[%v,%v]成功", SHOP_LIMITED_ITEM_SET, item_id, left_num)
	return left_num
}

func (this *RedisGlobalData) RefreshShop() int32 {
	items := this.shop_limited_items.Refresh()
	if items == nil {
		return -1
	}
	for k, v := range items {
		err := this.redis_conn.Post("HSET", SHOP_LIMITED_ITEM_SET, k, v)
		if err != nil {
			log.Error("redis设置集合[%v]数据[%v,%v]错误[%v]", SHOP_LIMITED_ITEM_SET, k, v, err.Error())
			return -1
		}
		log.Info("设置redis集合[%v]数据[%v,%v]成功")
	}
	return 1
}

func (this *RedisGlobalData) RefreshSomeShopItems(items []int32) bool {
	if !this.shop_limited_items.RefreshSome(items) {
		return false
	}
	for _, item_id := range items {
		item := shop_mgr.GetItem(item_id)
		if item == nil {
			continue
		}
		err := this.redis_conn.Post("HSET", SHOP_LIMITED_ITEM_SET, item_id, item.LimitedNumber)
		if err != nil {
			log.Error("redis设置集合[%v]数据[%v,%v]错误[%v]", SHOP_LIMITED_ITEM_SET, item_id, item.LimitedNumber)
			return false
		}
		log.Info("设置redis集合[%v]数据[%v,%v]成功", SHOP_LIMITED_ITEM_SET, item_id, item.LimitedNumber)
	}
	return true
}

func (this *RedisGlobalData) SetTimeForLimitedDaysItems(days int32, now_time int32) int32 {
	err := this.redis_conn.Post("HSET", SHOP_LIMITED_ITEM_LAST_SAVE_TIME, days, now_time)
	if err != nil {
		log.Error("redis设置集合[%v]数据[%v,%v]错误[%v]", SHOP_LIMITED_ITEM_LAST_SAVE_TIME, days, now_time)
		return -1
	}
	log.Info("设置redis集合[%v]数据[%v,%v]成功", SHOP_LIMITED_ITEM_LAST_SAVE_TIME, days, now_time)
	return 1
}

func (this *RedisGlobalData) CheckRefreshShop4Days(days int32) (int32, int32) {
	limited := shop_mgr.GetLimitedItems4Days(days)
	if limited == nil || limited.GlobalItems == nil || len(limited.GlobalItems) == 0 {
		log.Warn("没有[%v]天对应的限时商品", days)
		return 0, 0
	}

	now_time := time.Now()
	o, save_time := this.shop_limited_items.CheckRefreshShop4Days(days, limited, now_time)
	if !o {
		return 0, 0
	}

	for _, v := range limited.GlobalItems {
		err := this.redis_conn.Post("HSET", SHOP_LIMITED_ITEM_SET, v.Id, v.LimitedNumber)
		if err != nil {
			log.Error("redis设置集合[%v]数据[%v,%v]错误[%v]", SHOP_LIMITED_ITEM_SET, v.Id, v.LimitedNumber, err.Error())
			return -1, 0
		}
		log.Info("设置redis集合[%v]数据[%v,%v]成功", SHOP_LIMITED_ITEM_SET, v.Id, v.LimitedNumber)
	}

	res := this.SetTimeForLimitedDaysItems(days, int32(now_time.Unix()))
	if res < 0 {
		return -1, 0
	}

	return res, save_time
}
