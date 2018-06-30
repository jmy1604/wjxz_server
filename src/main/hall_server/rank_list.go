package main

import (
	"libs/log"
	"libs/utils"
	_ "main/table_config"
	_ "math/rand"
	_ "net/http"
	"public_message/gen_go/client_message"
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
}

func (this *RankList) Init(root_node utils.SkiplistNode) {
	this.rank_list = utils.NewCommonRankingList(root_node, ARENA_RANK_MAX)
	this.item_pool = &sync.Pool{
		New: func() interface{} {
			return &ArenaRankItem{}
		},
	}
}

// 获取排名
func (this *RankList) GetItemsByRank(player_id, start_rank, rank_num int32) (rank_items []*msg_client_message.RankItemInfo, self_rank int32, self_value interface{}) {
	start_rank, rank_num = this.rank_list.GetRankRange(start_rank, rank_num)
	if start_rank == 0 {
		log.Error("Get rank list range with [%v,%v] failed", start_rank, rank_num)
		return make([]*msg_client_message.RankItemInfo, 0), 0, nil
	}

	nodes := make([]interface{}, rank_num)
	for i := int32(0); i < rank_num; i++ {
		nodes[i] = this.item_pool.Get().(*msg_client_message.RankItemInfo)
	}

	num := this.rank_list.GetRangeNodes(start_rank, rank_num, nodes)
	if num == 0 {
		log.Error("Get rank list nodes failed")
		return make([]*msg_client_message.RankItemInfo, 0), 0, nil
	}

	rank_items = make([]*msg_client_message.RankItemInfo, num)
	for i := int32(0); i < num; i++ {
		rank_items[i] = nodes[i].(*msg_client_message.RankItemInfo)
	}

	self_rank, self_value = this.rank_list.GetRankAndValue(player_id)
	return

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

func (this *RankListManager) GetItemsByRange(rank_type, player_id, start_rank, rank_num int32) (rank_items []*msg_client_message.RankItemInfo, self_rank int32, self_value interface{}) {
	if int(rank_type) >= len(this.rank_lists) {
		return nil, 0, nil
	}
	return this.rank_lists[rank_type].GetItemsByRank(player_id, start_rank, rank_num)
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

func (this *RankListManager) GetItemsByRange2(rank_type, player_id, start_rank, rank_num int32) (rank_items []*msg_client_message.RankItemInfo, self_rank int32, self_value interface{}) {
	this.locker.RLock()
	rank_list := this.rank_map[rank_type]
	if rank_list == nil {
		return
	}
	this.locker.RUnlock()
	return rank_list.GetItemsByRank(player_id, start_rank, rank_num)
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
