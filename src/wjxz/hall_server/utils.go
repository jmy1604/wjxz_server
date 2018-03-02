package main

import (
	"math/rand"
	"time"
)

func rand31n_from_range(min, max int32) (bool, int32) {
	if min > max {
		return false, 0
	} else if min == max {
		return true, min
	}
	return true, (rand.Int31n(max-min) + min)
}

func GetRemainSeconds(start_time int32, duration int32) int32 {
	now := time.Now().Unix()
	if duration <= (int32(now) - start_time) {
		return 0
	}
	return duration - (int32(now) - start_time)
}

func GetRoundValue(value float32) int32 {
	v := int32(value)
	if value-float32(v) >= float32(0.5) {
		return v + 1
	} else {
		return v
	}
}

func GetPlayerBaseInfo(player_id int32) (name string, level int32, head string) {
	player := player_mgr.GetPlayerById(player_id)
	if player != nil {
		name = player.db.GetName()
		level = player.db.Info.GetLvl()
		head = player.db.Info.GetIcon()
	} else {
		row := os_player_mgr.GetPlayer(player_id)
		if row != nil {
			name = row.GetName()
			level = row.GetLevel()
			head = row.GetHead()
		}
	}
	return
}
