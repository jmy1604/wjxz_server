package main

import (
	_ "libs/log"
	"libs/utils"
	_ "main/table_config"
	_ "math/rand"
	_ "net/http"
	_ "public_message/gen_go/client_message"
	_ "public_message/gen_go/client_message_id"
	_ "sync"
	"time"

	_ "github.com/golang/protobuf/proto"
)

const (
	ARENA_RANK_MAX = 100000
)

type ArenaRankItem struct {
	SaveTime    int32
	PlayerScore int32
	PlayerLevel int32
	PlayerId    int32
}

func (this *ArenaRankItem) Less(value interface{}) bool {
	item := value.(*ArenaRankItem)
	if item == nil {
		return false
	}
	if this.PlayerScore < item.PlayerScore {
		return true
	}
	if this.PlayerScore == item.PlayerScore {
		if this.SaveTime > item.SaveTime {
			return true
		}
		if this.SaveTime == item.SaveTime {
			if this.PlayerLevel < item.PlayerLevel {
				return true
			}
		}
	}
	return false
}

func (this *ArenaRankItem) Greater(value interface{}) bool {
	item := value.(*ArenaRankItem)
	if item == nil {
		return false
	}
	if this.PlayerScore > item.PlayerScore {
		return true
	}
	if this.PlayerScore == item.PlayerScore {
		if this.SaveTime < item.SaveTime {
			return true
		}
		if this.SaveTime == item.SaveTime {
			if this.PlayerLevel > item.PlayerLevel {
				return true
			}
		}
	}
	return false
}

func (this *ArenaRankItem) KeyEqual(value interface{}) bool {
	item := value.(*ArenaRankItem)
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

func (this *ArenaRankItem) GetKey() interface{} {
	return this.PlayerId
}

func (this *ArenaRankItem) GetValue() interface{} {
	return this.PlayerScore
}

func (this *ArenaRankItem) New() utils.SkiplistNode {
	return &ArenaRankItem{}
}

func (this *ArenaRankItem) Assign(node utils.SkiplistNode) {
	n := node.(*ArenaRankItem)
	if n == nil {
		return
	}
	this.PlayerId = n.PlayerId
	this.PlayerLevel = n.PlayerLevel
	this.PlayerScore = n.PlayerScore
	this.SaveTime = n.SaveTime
}

func (this *ArenaRankItem) CopyDataTo(node interface{}) {
	return
}

func (this *Player) UpdateArenaScore(score int32) {
	var data = ArenaRankItem{
		SaveTime:    int32(time.Now().Unix()),
		PlayerScore: score,
		PlayerLevel: this.db.Info.GetLvl(),
		PlayerId:    this.Id,
	}
	rank_list_mgr.UpdateItem(RANK_LIST_TYPE_ARENA, &data)
}
