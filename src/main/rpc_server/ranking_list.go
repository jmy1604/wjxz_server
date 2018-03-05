package main

import (
	"libs/log"
	"libs/utils"
	"sync"
	"youma/rpc_common"
)

const (
	RANKING_LIST_TYPE_STAGE_TOTAL_SCORE = 1
	RANKING_LIST_TYPE_STAGE_SCORE       = 2
	RANKING_LIST_TYPE_CHARM             = 3
	RANKING_LIST_TYPE_CAT_OUQI          = 4
	RANKING_LIST_TYPE_ZANED             = 5
)

// 关卡总分排行项
type RankStageTotalScoreItem struct {
	PlayerId        int32
	PlayerLevel     int32
	StageTotalScore int32
	SaveTime        int32
}

// 每关排行项
type RankStageScoreItem struct {
	PlayerId    int32
	PlayerLevel int32
	StageId     int32
	StageScore  int32
	SaveTime    int32
}

// 魅力排行项
type RankCharmItem struct {
	PlayerId    int32
	PlayerLevel int32
	Charm       int32
	SaveTime    int32
}

// 猫欧气值排行项
type RankOuqiItem struct {
	PlayerId    int32
	PlayerLevel int32
	CatId       int32
	CatTableId  int32
	CatOuqi     int32
	CatLevel    int32
	CatStar     int32
	CatNick     string
	SaveTime    int32
}

// 被赞排行项
type RankZanedItem struct {
	PlayerId    int32
	PlayerLevel int32
	Zaned       int32
	SaveTime    int32
}

/* 关卡总积分 */
func (this *RankStageTotalScoreItem) Less(value interface{}) bool {
	item := value.(*RankStageTotalScoreItem)
	if item == nil {
		return false
	}
	if this.StageTotalScore < item.StageTotalScore {
		return true
	}
	if this.StageTotalScore == item.StageTotalScore {
		if this.SaveTime > item.SaveTime {
			return true
		}
		if this.SaveTime == item.SaveTime {
			if this.PlayerLevel < item.PlayerLevel {
				return true
			}
			if this.PlayerLevel == item.PlayerLevel {
				if this.PlayerId < item.PlayerId {
					return true
				}
			}
		}
	}
	return false
}

func (this *RankStageTotalScoreItem) Greater(value interface{}) bool {
	item := value.(*RankStageTotalScoreItem)
	if item == nil {
		return false
	}
	if this.StageTotalScore > item.StageTotalScore {
		return true
	}
	if this.StageTotalScore == item.StageTotalScore {
		if this.SaveTime < item.SaveTime {
			return true
		}
		if this.SaveTime == item.SaveTime {
			if this.PlayerLevel > item.PlayerLevel {
				return true
			}
			if this.PlayerLevel == item.PlayerLevel {
				if this.PlayerId > item.PlayerId {
					return true
				}
			}
		}
	}
	return false
}

func (this *RankStageTotalScoreItem) KeyEqual(value interface{}) bool {
	item := value.(*RankStageTotalScoreItem)
	if item == nil {
		return false
	}
	if item == nil {
		return false
	}
	if this.PlayerId == item.PlayerId {
		return true
	}
	return false
}

func (this *RankStageTotalScoreItem) GetKey() interface{} {
	return this.PlayerId
}

func (this *RankStageTotalScoreItem) GetValue() interface{} {
	return this.StageTotalScore
}

func (this *RankStageTotalScoreItem) New() utils.SkiplistNodeValue {
	return &RankStageTotalScoreItem{}
}

func (this *RankStageTotalScoreItem) Assign(node utils.SkiplistNodeValue) {
	n := node.(*RankStageTotalScoreItem)
	if n == nil {
		return
	}
	this.PlayerId = n.PlayerId
	this.PlayerLevel = n.PlayerLevel
	this.StageTotalScore = n.StageTotalScore
	this.SaveTime = n.SaveTime
}

func (this *RankStageTotalScoreItem) CopyDataTo(node interface{}) {
	n := node.(*rpc_common.H2R_RankStageTotalScore)
	if n == nil {
		return
	}
	n.PlayerId = this.PlayerId
	n.PlayerLevel = this.PlayerLevel
	n.TotalScore = this.StageTotalScore
}

/* 关卡积分*/
func (this *RankStageScoreItem) Less(value interface{}) bool {
	item := value.(*RankStageScoreItem)
	if item == nil {
		return false
	}
	if this.StageScore < item.StageScore {
		return true
	}
	if this.StageScore == item.StageScore {
		if this.SaveTime > item.SaveTime {
			return true
		}
		if this.SaveTime == item.SaveTime {
			if this.PlayerLevel < item.PlayerLevel {
				return true
			}
			if this.PlayerLevel == item.PlayerLevel {
				if this.PlayerId < item.PlayerId {
					return true
				}
			}
		}
	}
	return false
}

func (this *RankStageScoreItem) Greater(value interface{}) bool {
	item := value.(*RankStageScoreItem)
	if item == nil {
		return false
	}
	if this.StageScore > item.StageScore {
		return true
	}
	if this.StageScore == item.StageScore {
		if this.SaveTime < item.SaveTime {
			return true
		}
		if this.SaveTime == item.SaveTime {
			if this.PlayerLevel > item.PlayerLevel {
				return true
			}
			if this.PlayerLevel == item.PlayerLevel {
				if this.PlayerId > item.PlayerId {
					return true
				}
			}
		}
	}
	return false
}

func (this *RankStageScoreItem) KeyEqual(value interface{}) bool {
	item := value.(*RankStageScoreItem)
	if item == nil {
		return false
	}
	if this.PlayerId == item.PlayerId {
		return true
	}
	return false
}

func (this *RankStageScoreItem) GetKey() interface{} {
	return this.PlayerId
}

func (this *RankStageScoreItem) GetValue() interface{} {
	return this.StageScore
}

func (this *RankStageScoreItem) New() utils.SkiplistNodeValue {
	return &RankStageScoreItem{}
}

func (this *RankStageScoreItem) Assign(node utils.SkiplistNodeValue) {
	n := node.(*RankStageScoreItem)
	if n != nil {
		this.PlayerId = n.PlayerId
		this.PlayerLevel = n.PlayerLevel
		this.StageId = n.StageId
		this.StageScore = n.StageScore
		this.SaveTime = n.SaveTime
	}
}

func (this *RankStageScoreItem) CopyDataTo(node interface{}) {
	n := node.(*rpc_common.H2R_RankStageScore)
	if n == nil {
		return
	}
	n.PlayerId = this.PlayerId
	n.PlayerLevel = this.PlayerLevel
	n.StageId = this.StageId
	n.StageScore = this.StageScore
}

/*魅力值*/
func (this *RankCharmItem) Less(value interface{}) bool {
	item := value.(*RankCharmItem)
	if item == nil {
		return false
	}
	if this.Charm < item.Charm {
		return true
	}
	if this.Charm == item.Charm {
		if this.SaveTime > item.SaveTime {
			return true
		}
		if this.SaveTime == item.SaveTime {
			if this.PlayerLevel < item.PlayerLevel {
				return true
			}
			if this.PlayerLevel == item.PlayerLevel {
				if this.PlayerId < item.PlayerId {
					return true
				}
			}
		}
	}
	return false
}

func (this *RankCharmItem) Greater(value interface{}) bool {
	item := value.(*RankCharmItem)
	if item == nil {
		return false
	}
	if this.Charm > item.Charm {
		return true
	}
	if this.Charm == item.Charm {
		if this.SaveTime < item.SaveTime {
			return true
		}
		if this.SaveTime == item.SaveTime {
			if this.PlayerLevel > item.PlayerLevel {
				return true
			}
			if this.PlayerLevel == item.PlayerLevel {
				if this.PlayerId > item.PlayerId {
					return true
				}
			}
		}
	}
	return false
}

func (this *RankCharmItem) KeyEqual(value interface{}) bool {
	item := value.(*RankCharmItem)
	if item == nil {
		return false
	}
	if this.PlayerId == item.PlayerId {
		return true
	}
	return false
}

func (this *RankCharmItem) GetKey() interface{} {
	return this.PlayerId
}

func (this *RankCharmItem) GetValue() interface{} {
	return this.Charm
}

func (this *RankCharmItem) New() utils.SkiplistNodeValue {
	return &RankCharmItem{}
}

func (this *RankCharmItem) Assign(node utils.SkiplistNodeValue) {
	n := node.(*RankCharmItem)
	if n != nil {
		this.PlayerId = n.PlayerId
		this.PlayerLevel = n.PlayerLevel
		this.Charm = n.Charm
		this.SaveTime = n.SaveTime
	}
}

func (this *RankCharmItem) CopyDataTo(node interface{}) {
	n := node.(*rpc_common.H2R_RankCharm)
	if n == nil {
		return
	}
	n.PlayerId = this.PlayerId
	n.PlayerLevel = this.PlayerLevel
	n.Charm = this.Charm
}

/*欧气值*/
func (this *RankOuqiItem) Less(value interface{}) bool {
	item := value.(*RankOuqiItem)
	if item == nil {
		return false
	}
	if this.CatOuqi < item.CatOuqi {
		return true
	}
	if this.CatOuqi == item.CatOuqi {
		if this.SaveTime > item.SaveTime {
			return true
		}
		if this.SaveTime == item.SaveTime {
			if this.PlayerId < item.PlayerId {
				return true
			}
			if this.PlayerId == item.PlayerId {
				if this.CatId < item.CatId {
					return true
				}
			}
		}
	}
	return false
}

func (this *RankOuqiItem) Greater(value interface{}) bool {
	item := value.(*RankOuqiItem)
	if item == nil {
		return false
	}
	if this.CatOuqi > item.CatOuqi {
		return true
	}
	if this.CatOuqi == item.CatOuqi {
		if this.SaveTime < item.SaveTime {
			return true
		}
		if this.SaveTime == item.SaveTime {
			if this.PlayerId > item.PlayerId {
				return true
			}
			if this.PlayerId == item.PlayerId {
				if this.CatId > item.CatId {
					return true
				}
			}
		}
	}
	return false
}

func (this *RankOuqiItem) KeyEqual(value interface{}) bool {
	item := value.(*RankOuqiItem)
	if item == nil {
		return false
	}
	if this.PlayerId == item.PlayerId && this.CatId == item.CatId {
		return true
	}
	return false
}

func (this *RankOuqiItem) GetKey() interface{} {
	return utils.Int64From2Int32(this.PlayerId, this.CatId)
}

func (this *RankOuqiItem) GetValue() interface{} {
	return this.CatOuqi
}

func (this *RankOuqiItem) New() utils.SkiplistNodeValue {
	return &RankOuqiItem{}
}

func (this *RankOuqiItem) Assign(node utils.SkiplistNodeValue) {
	n := node.(*RankOuqiItem)
	if n != nil {
		this.PlayerId = n.PlayerId
		this.PlayerLevel = n.PlayerLevel
		this.CatId = n.CatId
		this.CatTableId = n.CatTableId
		this.CatLevel = n.CatLevel
		this.CatStar = n.CatStar
		this.CatNick = n.CatNick
		this.CatOuqi = n.CatOuqi
		this.SaveTime = n.SaveTime
	}
}

func (this *RankOuqiItem) CopyDataTo(node interface{}) {
	n := node.(*rpc_common.H2R_RankCatOuqi)
	if n == nil {
		return
	}
	n.PlayerId = this.PlayerId
	n.PlayerLevel = this.PlayerLevel
	n.CatId = this.CatId
	n.CatLevel = this.CatLevel
	n.CatTableId = this.CatTableId
	n.CatStar = this.CatStar
	n.CatNick = this.CatNick
	n.CatOuqi = this.CatOuqi
}

/*被赞*/
func (this *RankZanedItem) Less(value interface{}) bool {
	item := value.(*RankZanedItem)
	if item == nil {
		return false
	}
	if this.Zaned < item.Zaned {
		return true
	}
	if this.Zaned == item.Zaned {
		if this.SaveTime > item.SaveTime {
			return true
		}
		if this.SaveTime == item.SaveTime {
			if this.PlayerLevel < item.PlayerLevel {
				return true
			}
			if this.PlayerLevel == item.PlayerLevel {
				if this.PlayerId < item.PlayerId {
					return true
				}
			}
		}
	}
	return false
}

func (this *RankZanedItem) Greater(value interface{}) bool {
	item := value.(*RankZanedItem)
	if item == nil {
		return false
	}
	if this.Zaned > item.Zaned {
		return true
	}
	if this.Zaned == item.Zaned {
		if this.SaveTime < item.SaveTime {
			return true
		}
		if this.SaveTime == item.SaveTime {
			if this.PlayerLevel > item.PlayerLevel {
				return true
			}
			if this.PlayerLevel == item.PlayerLevel {
				if this.PlayerId > item.PlayerId {
					return true
				}
			}
		}
	}
	return false
}

func (this *RankZanedItem) KeyEqual(value interface{}) bool {
	item := value.(*RankZanedItem)
	if item == nil {
		return false
	}
	if this.PlayerId == item.PlayerId {
		return true
	}
	return false
}

func (this *RankZanedItem) GetKey() interface{} {
	return this.PlayerId
}

func (this *RankZanedItem) GetValue() interface{} {
	return this.Zaned
}

func (this *RankZanedItem) New() utils.SkiplistNodeValue {
	return &RankZanedItem{}
}

func (this *RankZanedItem) Assign(node utils.SkiplistNodeValue) {
	n := node.(*RankZanedItem)
	if n != nil {
		this.PlayerId = n.PlayerId
		this.PlayerLevel = n.PlayerLevel
		this.Zaned = n.Zaned
		this.SaveTime = n.SaveTime
	}
}

func (this *RankZanedItem) CopyDataTo(node interface{}) {
	n := node.(*rpc_common.H2R_RankZaned)
	if n == nil {
		return
	}
	n.PlayerId = this.PlayerId
	n.PlayerLevel = this.PlayerLevel
	n.Zaned = this.Zaned
}

// -----------------------------------------------------------------------------
/* 通用排行榜 */
// -----------------------------------------------------------------------------
type CommonRankingList struct {
	ranking_items *utils.Skiplist
	key2item      map[interface{}]utils.SkiplistNodeValue
	root_node     utils.SkiplistNodeValue
	max_rank      int32
	items_pool    *sync.Pool
	locker        *sync.RWMutex
}

func NewCommonRankingList(root_node utils.SkiplistNodeValue, max_rank int32) *CommonRankingList {
	ranking_list := &CommonRankingList{
		ranking_items: utils.NewSkiplist(),
		key2item:      make(map[interface{}]utils.SkiplistNodeValue),
		root_node:     root_node,
		max_rank:      max_rank,
		items_pool: &sync.Pool{
			New: func() interface{} {
				return root_node.New()
			},
		},
		locker: &sync.RWMutex{},
	}

	return ranking_list
}

func (this *CommonRankingList) GetByRank(rank int32) utils.SkiplistNodeValue {
	this.locker.RLock()
	defer this.locker.RUnlock()

	item := this.ranking_items.GetByRank(rank)
	if item == nil {
		return nil
	}
	new_item := this.items_pool.Get().(utils.SkiplistNodeValue)
	new_item.Assign(item)
	return new_item
}

func (this *CommonRankingList) GetByKey(key interface{}) utils.SkiplistNodeValue {
	this.locker.RLock()
	defer this.locker.RUnlock()

	item, o := this.key2item[key]
	if !o || item == nil {
		return nil
	}
	new_item := this.items_pool.Get().(utils.SkiplistNodeValue)
	new_item.Assign(item)
	return new_item
}

func (this *CommonRankingList) insert(key interface{}, item utils.SkiplistNodeValue, is_lock bool) bool {
	if is_lock {
		this.locker.Lock()
		defer this.locker.Unlock()
	}
	this.ranking_items.Insert(item)
	this.key2item[key] = item
	return true
}

func (this *CommonRankingList) Insert(key interface{}, item utils.SkiplistNodeValue) bool {
	return this.insert(key, item, true)
}

func (this *CommonRankingList) delete(key interface{}, is_lock bool) bool {
	if is_lock {
		this.locker.Lock()
		defer this.locker.Unlock()
	}

	item, o := this.key2item[key]
	if !o {
		log.Debug("CommonRankingList key[%v] not found", key)
		return false
	}
	if !this.ranking_items.Delete(item) {
		log.Debug("CommonRankingList delete key[%v] value[%v] in ranking list failed", key, item.GetValue())
		return false
	}
	if is_lock {
		this.items_pool.Put(item)
	}
	//delete(this.key2item, key)
	return true
}

func (this *CommonRankingList) Delete(key interface{}) bool {
	return this.delete(key, true)
}

func (this *CommonRankingList) Update(item utils.SkiplistNodeValue) bool {
	this.locker.Lock()
	defer this.locker.Unlock()

	old_item, o := this.key2item[item.GetKey()]
	if o {
		if !this.delete(item.GetKey(), false) {
			log.Error("Update key[%v] for StageTotalScoreRankingList failed", item)
			return false
		}
		old_item.Assign(item)
		return this.insert(item.GetKey(), old_item, false)
	} else {
		new_item := this.items_pool.Get().(utils.SkiplistNodeValue)
		new_item.Assign(item)
		return this.insert(item.GetKey(), new_item, false)
	}
}

func (this *CommonRankingList) GetRangeNodes(rank_start, rank_num int32, nodes []interface{}) (num int32) {
	if rank_start <= int32(0) || rank_start > this.max_rank {
		log.Warn("Ranking list rank_start[%v] invalid", rank_start)
		return
	}

	this.locker.RLock()
	defer this.locker.RUnlock()

	if int(rank_start) > len(this.key2item) {
		log.Debug("Ranking List rank range[1,%v], rank_start[%v] over rank list", len(this.key2item), rank_start)
		return
	}

	real_num := int32(len(this.key2item)) - rank_start + 1
	if real_num < rank_num {
		rank_num = real_num
	}

	items := make([]utils.SkiplistNodeValue, rank_num)
	b := this.ranking_items.GetByRankRange(rank_start, rank_num, items)
	if !b {
		log.Warn("Ranking List rank range[%v,%v] is empty", rank_start, rank_num)
		return
	}

	for i := int32(0); i < rank_num; i++ {
		item := items[i]
		if item == nil {
			log.Error("Get Rank[%v] for Ranking List failed")
			continue
		}
		node := nodes[i]
		item.CopyDataTo(node)
		num += 1
	}
	return
}

func (this *CommonRankingList) GetRank(key interface{}) int32 {
	this.locker.RLock()
	defer this.locker.RUnlock()

	item, o := this.key2item[key]
	if !o {
		return 0
	}
	return this.ranking_items.GetRank(item)
}

func (this *CommonRankingList) GetRankAndValue(key interface{}) (rank int32, value interface{}) {
	this.locker.RLock()
	defer this.locker.RUnlock()

	item, o := this.key2item[key]
	if !o {
		return 0, nil
	}

	return this.ranking_items.GetRank(item), item.GetValue()
}

func (this *CommonRankingList) GetRankRange(start, num int32) (int32, int32) {
	this.locker.RLock()
	defer this.locker.RUnlock()

	l := int32(len(this.key2item))
	if this.key2item == nil || l == 0 {
		return 0, 0
	}

	if start > l {
		return 0, 0
	}

	if l-start+1 < num {
		num = l - start + 1
	}
	return start, num
}
