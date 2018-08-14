package utils

import (
	"libs/log"
	"sync"
)

const (
	SHORT_RANK_ITEM_MAX_NUM = 100
)

type ShortRankItem interface {
	Less(item ShortRankItem) bool
	Greater(item ShortRankItem) bool
	GetKey() interface{}
	GetValue() interface{}
	Assign(item ShortRankItem)
}

type ShortRankList struct {
	items    []ShortRankItem
	max_num  int32
	curr_num int32
	keys_map map[interface{}]int32
	locker   *sync.RWMutex
}

func (this *ShortRankList) Init(max_num int32) bool {
	if max_num <= 0 {
		return false
	}

	this.items = make([]ShortRankItem, max_num)
	this.max_num = max_num
	this.keys_map = make(map[interface{}]int32)
	this.locker = &sync.RWMutex{}
	return true
}

func (this *ShortRankList) GetLength() int32 {
	this.locker.RLock()
	defer this.locker.RUnlock()
	return this.curr_num
}

func (this *ShortRankList) Update(item ShortRankItem) bool {
	this.locker.Lock()
	defer this.locker.Unlock()

	idx, o := this.keys_map[item.GetKey()]
	if !o && this.curr_num >= this.max_num {
		log.Error("Short Rank List length %v is max, cant insert new item", this.curr_num)
		return false
	}

	if !o {
		i := this.curr_num
		for ; i >= 0; i-- {
			if !item.Greater(this.items[i]) {
				break
			}
		}
		if i >= 0 {
			for n := this.curr_num - 1; n >= i; n-- {
				this.items[n+1] = this.items[n]
				this.keys_map[this.items[n+1]] = n + 1
			}
			this.items[i] = item
		} else {
			this.items[this.curr_num] = item
		}
		this.keys_map[item.GetKey()] = this.curr_num
		this.curr_num += 1
	} else {
		var i, b, e int32
		if item.Greater(this.items[idx]) {
			i = idx - 1
			for ; i >= 0; i-- {
				if !item.Greater(this.items[i]) {
					break
				}
			}
			b = i
			e = idx - 1
		} else if item.Less(this.items[idx]) {
			i = idx + 1
			for ; i < this.curr_num; i++ {
				if item.Greater(this.items[i]) {
					break
				}
			}
			b = idx + 1
			e = i
		} else {
			return false
		}

		for i = e; i >= b; i-- {
			this.items[i+1] = this.items[i]
		}

		this.items[i].Assign(item)
	}

	return true
}

func (this *ShortRankList) GetRank(key interface{}) (rank int32) {
	this.locker.RLock()
	defer this.locker.RUnlock()

	rank, _ = this.keys_map[key]
	rank += 1
	return
}

func (this *ShortRankList) GetByRank(rank int32) (key interface{}, value interface{}) {
	this.locker.RLock()
	defer this.locker.RUnlock()

	if this.curr_num < rank {
		return
	}
	item := this.items[rank-1]
	if item == nil {
		return
	}
	key = item.GetKey()
	value = item.GetValue()
	return
}
