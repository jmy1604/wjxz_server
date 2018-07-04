package main

import (
	"libs/log"
	"libs/utils"
	"public_message/gen_go/client_message"
	"sync"
	"time"

	_ "3p/code.google.com.protobuf/proto"
	_ "github.com/golang/protobuf/proto"
)

const MAX_WORLD_CHAT_ONCE_GET int32 = 50
const MAX_WORLD_CHAT_MSG_NUM int32 = 150

type WorldChatItem struct {
	send_player_id    int32
	send_player_name  string
	send_player_level int32
	send_player_head  string
	content           []byte
	send_time         int32
	prev              *WorldChatItem
	next              *WorldChatItem
}

type WorldChatItemFactory struct {
}

func (this *WorldChatItemFactory) New() interface{} {
	return &WorldChatItem{}
}

type PlayerWorldChatData struct {
	curr_msg       *WorldChatItem
	curr_send_time int32
}

type WorldChatMgr struct {
	msg_num       int32                 // 消息数
	chat_msg_head *WorldChatItem        // 最早的结点
	chat_msg_tail *WorldChatItem        // 最新的节点
	items_pool    *utils.SimpleItemPool // 消息池
	items_factory *WorldChatItemFactory // 对象工厂
	locker        *sync.RWMutex         // 锁
}

var world_chat_mgr WorldChatMgr

func get_world_chat_max_msg_num() int32 {
	max_num := global_config.WorldChatMaxMsgNum
	if max_num == 0 {
		max_num = MAX_WORLD_CHAT_MSG_NUM
	}
	return max_num
}

func (this *WorldChatMgr) Init() {
	this.items_pool = &utils.SimpleItemPool{}
	this.items_factory = &WorldChatItemFactory{}
	this.items_pool.Init(get_world_chat_max_msg_num(), this.items_factory)
	this.locker = &sync.RWMutex{}
	this.chat_msg_head = nil
	this.chat_msg_tail = nil
}

func (this *WorldChatMgr) recycle_old() {
	now_time := int32(time.Now().Unix())
	msg := this.chat_msg_head
	for msg != nil {
		if now_time-msg.send_time >= global_config.WorldChatMsgExistTime*60 {
			if msg == this.chat_msg_head {
				this.chat_msg_head = msg.next
			}
			if msg == this.chat_msg_tail {
				this.chat_msg_tail = nil
			}
			this.items_pool.Recycle(msg)
			if msg.prev != nil {
				msg.prev.next = msg.next
			}
			if msg.next != nil {
				msg.next.prev = msg.prev
			}
		}
		msg = msg.next
	}
}

func (this *WorldChatMgr) push_chat_msg(content []byte, player_id int32, player_level int32, player_name string, player_head string) bool {
	this.locker.Lock()
	defer this.locker.Unlock()

	this.recycle_old()

	if !this.items_pool.HasFree() {
		// 回收最早的节点
		if !this.items_pool.Recycle(this.chat_msg_head) {
			log.Error("###[WorldChatMgr]### Recycle failed")
			return false
		}
		n := this.chat_msg_head.next
		this.chat_msg_head = n
		if n != nil {
			n.prev = nil
		}
	}

	it := this.items_pool.GetFree()
	if it == nil {
		log.Error("###[WorldChatMgr]### No free item")
		return false
	}

	item := it.(*WorldChatItem)
	item.content = content
	item.send_player_id = player_id
	item.send_player_name = player_name
	item.send_player_head = player_head
	item.send_player_level = player_level
	item.send_time = int32(time.Now().Unix())

	item.prev = this.chat_msg_tail
	item.next = nil
	if this.chat_msg_head == nil {
		this.chat_msg_head = item
	}
	if this.chat_msg_tail != nil {
		this.chat_msg_tail.next = item
	}
	this.chat_msg_tail = item
	this.msg_num += 1

	return true
}

func (this *WorldChatMgr) pull_world_chat(player *Player) (chat_items []*msg_client_message.WorldChatItem) {
	this.locker.RLock()
	defer this.locker.RUnlock()

	if this.msg_num <= 0 {
		chat_items = make([]*msg_client_message.WorldChatItem, 0)
		return
	}
	msg_num := MAX_WORLD_CHAT_ONCE_GET
	if msg_num > this.msg_num {
		msg_num = this.msg_num
	}
	msg := player.world_chat_data.curr_msg
	if msg == nil {
		msg = this.chat_msg_head
	} else {
		if msg.send_time != player.world_chat_data.curr_send_time {
			msg = this.chat_msg_head
		} else {
			msg = msg.next
		}
	}

	now_time := int32(time.Now().Unix())

	for n := int32(0); n < msg_num; n++ {
		if msg == nil {
			break
		}
		if now_time-msg.send_time >= global_config.WorldChatMsgExistTime*60 {
			msg = msg.next
			continue
		}
		item := &msg_client_message.WorldChatItem{
			Content:     msg.content,
			PlayerId:    msg.send_player_id,
			PlayerName:  msg.send_player_name,
			PlayerLevel: msg.send_player_level,
			PlayerHead:  msg.send_player_head,
			SendTime:    msg.send_time,
		}
		chat_items = append(chat_items, item)

		player.world_chat_data.curr_msg = msg
		player.world_chat_data.curr_send_time = msg.send_time
		msg = msg.next
	}

	return
}

func (this *Player) world_chat(content []byte) int32 {
	now_time := int32(time.Now().Unix())
	/*if now_time < this.db.TalkForbid.GetEndUnix() {
		log.Error("Player[%v] world chat is forbidden !", this.Id)
		res := &msg_client_message.S2CWorldChatForbid{}
		end_t := time.Unix(int64(this.db.TalkForbid.GetEndUnix()), 0)
		res.EndTime = end_t.Format("2006-01-02 15:04:05.999999999")

		this.Send(uint16(1), res)

		return int32(msg_client_message.E_ERR_WORLDCHAT_SEND_MSG_BE_FORBIDEN)
	}*/
	if now_time-this.db.WorldChat.GetLastChatTime() < 10 /*global_id.WorldChannelChatCooldown_40*/ {
		log.Error("Player[%v] world chat is cooling down !", this.Id)
		return int32(msg_client_message.E_ERR_WORLDCHAT_SEND_MSG_COOLING_DOWN)
	}
	if int32(len(content)) > global_config.WorldChatMsgMaxBytes {
		log.Error("Player[%v] world chat content length is too long !", this.Id)
		return int32(msg_client_message.E_ERR_WORLDCHAT_SEND_MSG_BYTES_TOO_LONG)
	}
	if !world_chat_mgr.push_chat_msg(content, this.Id, this.db.Info.GetLvl(), this.db.GetName(), this.db.Info.GetIcon()) {
		return int32(msg_client_message.E_ERR_WORLDCHAT_CANT_SEND_WITH_NO_FREE)
	}

	if this.rpc_world_chat(content) == nil {
		log.Error("Player[%v] world chat to remote rpc service failed", this.Id)
	}

	this.db.WorldChat.SetLastChatTime(now_time)

	response := &msg_client_message.S2CWorldChatSendResult{
		Content: content,
	}
	this.Send(1, response)
	log.Debug("Player[%v] world chat content[%v]", this.Id, content)

	return 1
}

func (this *Player) pull_world_chat() int32 {
	now_time := int32(time.Now().Unix())
	if now_time-this.db.WorldChat.GetLastPullTime() < global_config.WorldChatPullMsgCooldown {
		log.Error("Player[%v] pull world chat msg is cooling down", this.Id)
		//return int32(msg_client_message.E_ERR_WORLDCHAT_PULL_COOLING_DOWN)
		response := &msg_client_message.S2CWorldChatMsgPullResult{}
		this.Send(1, response)
		return 1
	}
	msgs := world_chat_mgr.pull_world_chat(this)
	this.db.WorldChat.SetLastPullTime(now_time)
	response := &msg_client_message.S2CWorldChatMsgPullResult{
		Items: msgs,
	}
	this.Send(1, response)
	log.Debug("Player[%v] pulled world chat msgs", this.Id)
	return 1
}
