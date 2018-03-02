package main

import (
	"libs/log"
	"libs/utils"
	"math"
	"public_message/gen_go/client_message"
	"sync"
	"time"
	"youma/rpc_common"
)

type PersonalSpaceComment struct {
	id               int32
	content          []byte
	send_player_id   int32
	send_player_time int32
	prev             *PersonalSpaceComment
	next             *PersonalSpaceComment
	subs             *PersonalSpaceComment
}
