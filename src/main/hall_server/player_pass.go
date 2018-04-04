package main

import (
	_ "libs/log"
	_ "main/table_config"
	_ "math"
	_ "math/rand"
	_ "public_message/gen_go/client_message"
	_ "time"
)

func (this *Player) begin_pass(pass_id int32) int32 {
	return 1
}

func (this *Player) end_pass(pass_id int32) int32 {
	return 1
}
