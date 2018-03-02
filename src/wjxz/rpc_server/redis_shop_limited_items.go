package main

import (
	"libs/log"
	"strconv"
	"sync"
	"time"
	"youma/table_config"

	"github.com/garyburd/redigo/redis"
)

//const SHOP_LIMITED_ITEMS_FIRST_LOAD = "mm:shop_limited_items_first_load"
const SHOP_LIMITED_ITEM_SET = "mm:shop_limited_item_set"
const SHOP_LIMITED_ITEM_LAST_SAVE_TIME = "mm:shop_limited_item_last_refresh_save_time"

type ShopLimitedItem struct {
	item_id  int32
	left_num int32
}

type ShopLimitedItemSet struct {
	last_save_time []int32
	items          map[int32]*ShopLimitedItem
	mtx            *sync.RWMutex
}

func (this *ShopLimitedItemSet) save_limited_days_time(days int32, now_time int32) {
	if this.last_save_time[days] == 0 {
		global_data.SetTimeForLimitedDaysItems(days, now_time)
		this.last_save_time[days] = int32(now_time)
	}
}

func (this *ShopLimitedItemSet) Init() bool {
	this.last_save_time = make([]int32, 512) // 限制为512天
	this.items = make(map[int32]*ShopLimitedItem)
	this.mtx = &sync.RWMutex{}

	var err error
	var int_map map[string]int

	// 限时刷新时间载入
	int_map, err = redis.IntMap(global_data.redis_conn.Do("HGETALL", SHOP_LIMITED_ITEM_LAST_SAVE_TIME))
	if err != nil {
		log.Error("redis获取数据[%v]错误[%v]", SHOP_LIMITED_ITEM_LAST_SAVE_TIME, err.Error())
		return false
	}

	for s, t := range int_map {
		days, e := strconv.Atoi(s)
		if e != nil {
			log.Warn("取出集合[%v]物品[%v]数据错误[%v]", SHOP_LIMITED_ITEM_LAST_SAVE_TIME, s, err.Error())
			continue
		}
		this.last_save_time[days] = int32(t)
		log.Debug("取出集合[%v]限时[%v]天对应最近刷新时间[%v]", SHOP_LIMITED_ITEM_LAST_SAVE_TIME, s, t)
	}

	// 限时商品载入
	int_map, err = redis.IntMap(global_data.redis_conn.Do("HGETALL", SHOP_LIMITED_ITEM_SET))
	if err != nil {
		log.Error("redis获取集合[%v]数据失败[%v]", SHOP_LIMITED_ITEM_SET, err.Error())
		return false
	}

	c := int32(0)
	now_time := time.Now().Unix()
	for s, id := range int_map {
		item_id, e := strconv.Atoi(s)
		if e != nil {
			log.Warn("取出集合[%v]物品[%v]数据失败[%v]", SHOP_LIMITED_ITEM_SET, s, err.Error())
			continue
		}

		// 判断是否合法
		item := shop_mgr.GetItem(int32(item_id))
		if item == nil {
			log.Warn("配置文件没有商品[%v]", item_id)
			continue
		}

		if item.LimitedType != 1 {
			log.Warn("配置物品[%v]不是全局限时商品", id)
			continue
		}

		if this.AddItem(int32(item_id), int32(id)) < 0 {
			log.Warn("载入集合[%v]数据[%v,%v]失败", SHOP_LIMITED_ITEM_SET, s, id)
		}

		this.save_limited_days_time(item.LimitedTime, int32(now_time))

		c += 1
	}

	// 第一次载入，生成新的限时商品
	if c == 0 {
		for _, v := range shop_mgr.GetItems() {
			if v.LimitedType == 1 {
				add_num := global_data.AddShopLimitedItem(v.Id, v.LimitedNumber)
				if add_num < 0 {
					log.Warn("首次加入商品[%v]失败", v.Id)
				} else {
					// 首次刷新时间
					this.save_limited_days_time(v.LimitedTime, int32(now_time))
					log.Info("首次加入商品[%v,%v]", v.Id, v.LimitedNumber)
				}
			}
		}
	}

	return true
}

func (this *ShopLimitedItemSet) AddItem(item_id int32, left_num int32) int32 {
	this.mtx.Lock()
	defer this.mtx.Unlock()

	if this.items == nil {
		this.items = make(map[int32]*ShopLimitedItem)
	}

	item := this.items[item_id]
	if item == nil {
		item = &ShopLimitedItem{}
		item.item_id = item_id
		item.left_num = left_num
		this.items[item_id] = item
	} else {
		item.left_num += left_num
	}

	return item.left_num
}

func (this *ShopLimitedItemSet) RemoveItem(item_id, item_num int32) (bool, int32) {
	this.mtx.Lock()
	defer this.mtx.Unlock()

	if this.items == nil {
		return false, 0
	}

	item := this.items[item_id]
	if item == nil {
		log.Warn("没有限时商品[%v]，删除失败")
		return false, 0
	}

	it := shop_mgr.GetItem(item_id)
	if it == nil {
		log.Warn("没有限时商品[%v]配置，删除失败")
		return false, 0
	}

	if item.left_num < it.Number*item_num {
		log.Info("限时商品[%v]数量[%v]不足", item_id, item.left_num)
		return false, 0
	}

	item.left_num -= (it.Number * item_num)

	return true, item.left_num
}

func (this *ShopLimitedItemSet) GetItemLeftNum(item_id int32) (bool, int32) {
	this.mtx.RLock()
	defer this.mtx.RUnlock()

	if this.items == nil {
		return false, 0
	}

	item := this.items[item_id]
	if item == nil {
		return false, 0
	}

	return true, item.left_num
}

func (this *ShopLimitedItemSet) GetSaveTime(days int32) (bool, int32) {
	this.mtx.RLock()
	defer this.mtx.RUnlock()

	if int(days) <= 0 || int(days) >= len(this.last_save_time) {
		return false, 0
	}

	return true, this.last_save_time[days]
}

func (this *ShopLimitedItemSet) Refresh() map[int32]int32 {
	this.mtx.Lock()
	defer this.mtx.Unlock()

	if this.items == nil {
		return nil
	}

	id2num := make(map[int32]int32)
	for k, v := range this.items {
		item := shop_mgr.GetItem(k)
		if item == nil {
			log.Warn("找不到商品[%v]配置", k)
			continue
		}
		v.left_num = item.LimitedNumber
		id2num[k] = item.LimitedNumber
	}
	return id2num
}

func (this *ShopLimitedItemSet) RefreshSome(items []int32) bool {
	this.mtx.Lock()
	defer this.mtx.Unlock()

	if this.items == nil {
		return false
	}

	for _, v := range items {
		titem := shop_mgr.GetItem(v)
		if titem == nil {
			log.Warn("找不到商品[%v]配置", v)
			continue
		}
		it := this.items[v]
		if it == nil {
			log.Warn("商品[%v]不存在", v)
			continue
		}
		it.left_num = titem.LimitedNumber
	}

	return true
}

func (this *ShopLimitedItemSet) CheckRefreshShop4Days(days int32, limited *table_config.ShopLimitedItems, now_time time.Time) (bool, int32) {
	this.mtx.Lock()
	defer this.mtx.Unlock()

	if days >= int32(len(this.last_save_time)) {
		return false, 0
	}

	last_save := this.last_save_time[days]
	if !global_config_mgr.GetShopTimeChecker().IsArrival(last_save, days) {
		return false, 0
	}

	for _, v := range limited.GlobalItems {
		citem := shop_mgr.GetItem(v.Id)
		if citem == nil {
			log.Warn("商品[%v]配置不存在，无法刷新", v.Id)
			continue
		}
		item := this.items[v.Id]
		if item == nil {
			log.Warn("商品[%v]不存在，无法刷新", v.Id)
			continue
		}
		item.left_num = citem.LimitedNumber
	}

	this.last_save_time[days] = int32(now_time.Unix())

	log.Debug("限时[%v]天商店刷新", days)

	return true, this.last_save_time[days]
}
