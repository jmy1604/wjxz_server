package main

import (
	"libs/log"
	"libs/utils"
	_ "main/table_config"
	_ "math/rand"
	_ "net/http"
	_ "public_message/gen_go/client_message"
	_ "public_message/gen_go/client_message_id"
	"sync"
	_ "time"

	_ "github.com/golang/protobuf/proto"
)

const (
	RANK_LIST_TYPE_NONE  = iota
	RANK_LIST_TYPE_ARENA = 1
	RANK_LIST_TYPE_MAX   = 16
)

type RankList struct {
	rank_list *utils.CommonRankingList
	item_pool *sync.Pool
	root_node utils.SkiplistNode
}

func (this *RankList) Init(root_node utils.SkiplistNode) {
	this.root_node = root_node
	this.rank_list = utils.NewCommonRankingList(this.root_node, ARENA_RANK_MAX)
	this.item_pool = &sync.Pool{
		New: func() interface{} {
			return this.root_node.New()
		},
	}
}

func (this *RankList) GetItemByPlayerId(player_id int32) (item utils.SkiplistNode) {
	return this.rank_list.GetByKey(player_id)
}

func (this *RankList) GetRankByPlayerId(player_id int32) int32 {
	return this.rank_list.GetRank(player_id)
}

func (this *RankList) GetItemByRank(rank int32) (item utils.SkiplistNode) {
	return this.rank_list.GetByRank(rank)
}

// 获取排名项
func (this *RankList) GetItemsByRank(player_id, start_rank, rank_num int32) (rank_items []utils.SkiplistNode, self_rank int32, self_value interface{}) {
	start_rank, rank_num = this.rank_list.GetRankRange(start_rank, rank_num)
	if start_rank == 0 {
		log.Error("Get rank list range with [%v,%v] failed", start_rank, rank_num)
		return nil, 0, nil
	}

	nodes := make([]interface{}, rank_num)
	for i := int32(0); i < rank_num; i++ {
		nodes[i] = this.item_pool.Get().(utils.SkiplistNode)
	}

	num := this.rank_list.GetRangeNodes(start_rank, rank_num, nodes)
	if num == 0 {
		log.Error("Get rank list nodes failed")
		return nil, 0, nil
	}

	rank_items = make([]utils.SkiplistNode, num)
	for i := int32(0); i < num; i++ {
		rank_items[i] = nodes[i].(utils.SkiplistNode)
	}

	self_rank, self_value = this.rank_list.GetRankAndValue(player_id)
	return

}

// 获取最后的几个排名
func (this *RankList) GetLastRankRange(rank_num int32) (int32, int32) {
	return this.rank_list.GetLastRankRange(rank_num)
}

// 更新排行榜
func (this *RankList) UpdateItem(item utils.SkiplistNode) bool {
	//before_first_item := this.rank_list.GetByRank(1)
	if !this.rank_list.Update(item) {
		log.Error("Update rank item[%v] failed", item)
		return false
	}

	//this.anouncement_stage_total_score_first_rank(before_first_item)

	log.Debug("Updated rank list item[%v]", item)

	return true
}

// 删除指定值
func (this *RankList) DeleteItem(key interface{}) bool {
	return this.DeleteItem(key)
}

var root_rank_item []utils.SkiplistNode = []utils.SkiplistNode{
	nil,
	&ArenaRankItem{},
}

type RankListManager struct {
	rank_lists []*RankList
	rank_map   map[int32]*RankList
	locker     *sync.RWMutex
}

var rank_list_mgr RankListManager

func (this *RankListManager) Init() {
	this.rank_lists = make([]*RankList, RANK_LIST_TYPE_MAX)
	for i := int32(1); i < RANK_LIST_TYPE_MAX; i++ {
		if int(i) >= len(root_rank_item) {
			break
		}
		this.rank_lists[i] = &RankList{}
		this.rank_lists[i].Init(root_rank_item[i])
	}
	this.rank_map = make(map[int32]*RankList)
	this.locker = &sync.RWMutex{}
}

func (this *RankListManager) GetItemByPlayerId(rank_type, player_id int32) (item utils.SkiplistNode) {
	if int(rank_type) >= len(this.rank_lists) {
		return nil
	}
	return this.rank_lists[rank_type].GetItemByPlayerId(player_id)
}

func (this *RankListManager) GetRankByPlayerId(rank_type, player_id int32) int32 {
	if int(rank_type) >= len(this.rank_lists) {
		return -1
	}
	return this.rank_lists[rank_type].GetRankByPlayerId(player_id)
}

func (this *RankListManager) GetItemByRank(rank_type, rank int32) (item utils.SkiplistNode) {
	if int(rank_type) >= len(this.rank_lists) {
		return nil
	}
	return this.rank_lists[rank_type].GetItemByRank(rank)
}

func (this *RankListManager) GetItemsByRange(rank_type, player_id, start_rank, rank_num int32) (rank_items []utils.SkiplistNode, self_rank int32, self_value interface{}) {
	if int(rank_type) >= len(this.rank_lists) {
		return nil, 0, nil
	}
	return this.rank_lists[rank_type].GetItemsByRank(player_id, start_rank, rank_num)
}

func (this *RankListManager) GetLastRankRange(rank_type, rank_num int32) (int32, int32) {
	if int(rank_type) >= len(this.rank_lists) {
		return -1, -1
	}
	return this.rank_lists[rank_type].GetLastRankRange(rank_num)
}

func (this *RankListManager) UpdateItem(rank_type int32, item utils.SkiplistNode) bool {
	if int(rank_type) >= len(this.rank_lists) {
		return false
	}
	return this.rank_lists[rank_type].UpdateItem(item)
}

func (this *RankListManager) DeleteItem(rank_type int32, key interface{}) bool {
	if int(rank_type) >= len(this.rank_lists) {
		return false
	}
	return this.rank_lists[rank_type].DeleteItem(key)
}

func (this *RankListManager) GetItemByPlayerId2(rank_type, player_id int32) (item utils.SkiplistNode) {
	this.locker.RLock()
	rank_list := this.rank_map[rank_type]
	if rank_list == nil {
		return
	}
	this.locker.RUnlock()
	return rank_list.GetItemByPlayerId(player_id)
}

func (this *RankListManager) GetRankByPlayerId2(rank_type, player_id int32) int32 {
	this.locker.RLock()
	rank_list := this.rank_map[rank_type]
	if rank_list == nil {
		return 0
	}
	this.locker.RUnlock()
	return rank_list.GetRankByPlayerId(player_id)
}

func (this *RankListManager) GetItemByRank2(rank_type, rank int32) (item utils.SkiplistNode) {
	this.locker.RLock()
	rank_list := this.rank_map[rank_type]
	if rank_list == nil {
		return
	}
	this.locker.RUnlock()
	return rank_list.GetItemByRank(rank)
}

func (this *RankListManager) GetItemsByRange2(rank_type, player_id, start_rank, rank_num int32) (rank_items []utils.SkiplistNode, self_rank int32, self_value interface{}) {
	this.locker.RLock()
	rank_list := this.rank_map[rank_type]
	if rank_list == nil {
		return
	}
	this.locker.RUnlock()
	return rank_list.GetItemsByRank(player_id, start_rank, rank_num)
}

func (this *RankListManager) GetLastRankRange2(rank_type, rank_num int32) (int32, int32) {
	this.locker.RLock()
	rank_list := this.rank_map[rank_type]
	if rank_list == nil {
		return -1, -1
	}
	this.locker.RUnlock()
	return rank_list.GetLastRankRange(rank_num)
}

func (this *RankListManager) UpdateItem2(rank_type int32, item utils.SkiplistNode) bool {
	this.locker.Lock()
	rank_list := this.rank_map[rank_type]
	if rank_list == nil {
		rank_list = &RankList{}
		this.rank_map[rank_type] = rank_list
	}
	this.locker.Unlock()
	return rank_list.UpdateItem(item)
}

func (this *RankListManager) DeleteItem2(rank_type int32, key interface{}) bool {
	this.locker.Lock()
	rank_list := this.rank_map[rank_type]
	if rank_list == nil {
		rank_list = &RankList{}
		this.rank_map[rank_type] = rank_list
	}
	this.locker.Unlock()
	return rank_list.DeleteItem(key)
}
