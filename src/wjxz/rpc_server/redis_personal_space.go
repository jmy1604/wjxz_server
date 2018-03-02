package main

import (
	"encoding/json"
	"fmt"
	"libs/log"
	//"libs/utils"

	"github.com/garyburd/redigo/redis"
)

const (
	REDIS_KEY_PERSONAL_SPACE_BASE_DATA                  = "mm:personal_space_base_data:player_id"
	REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGES             = "mm:personal_space_leave_messages:player_id+msg_id"
	REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGE_COMMENTS     = "mm:personal_space_leave_message_comments:player_id+msg_id+comment_id"
	REDIS_KEY_PERSONAL_SPACE_PICTURE_DATA               = "mm:personal_space_picture_data:player_id+pic_id"
	REDIS_KEY_PERSONAL_SPACE_PIC_LEAVE_MESSAGES         = "mm:personal_space_picture_leave_messages:player_id+pic_id+msg_id"
	REDIS_KEY_PERSONAL_SPACE_PIC_LEAVE_MESSAGE_COMMENTS = "mm:personal_space_picture_leave_message_comments:player_id+pic_id+msg_id+comment_id"
	REDIS_KEY_PERSONAL_SPACE_PIC_ZAN                    = "mm:personal_space_picture_zan:player_id+pic_id"
)

type RedisPSBaseData struct {
	PlayerId       int32   // 玩家ID
	Signature      string  // 签名
	PictureIds     []int32 // 照片
	LeaveMsgIds    []int32 // 留言
	CurrLeaveMsgId int32   // 当前留言ID
	LeaveMsgDsType int32
}

type RedisPSLeaveMessage struct {
	MsgId         int32   // 留言ID
	Content       []byte  // 留言内容
	SendPlayerId  int32   // 留言玩家ID
	CommentIds    []int32 // 评论ID
	CurrCommentId int32   // 当前用到的最新评论ID
	SendTime      int32   // 留言时间
}

type RedisPSComment struct {
	CommentId    int32  // 评论ID
	Content      []byte // 评论内容
	SendPlayerId int32  // 评论玩家ID
	SendTime     int32  // 评论时间
}

type RedisPSPicture struct {
	PicId          int32
	Url            string
	UploadTime     int32
	Zaned          int32
	LeaveMsgIds    []int32 // 留言按时间顺序降序排列
	CurrLeaveMsgId int32   // 当前已用的最大留言ID
	LeaveMsgDsType int32   // 0 表示数组
}

func (this *RedisPSLeaveMessage) load_comments(player_id, pic_id, msg_id int32) (comments []*PSLeaveMessageComment) {
	if this.CommentIds == nil {
		return
	}

	for i := 0; i < len(this.CommentIds); i++ {
		var key, value string
		var err error
		if pic_id == 0 {
			key = fmt.Sprintf("%v_%v_%v", player_id, msg_id, this.CommentIds[i])
			value, err = redis.String(global_data.redis_conn.Do("HGET", REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGE_COMMENTS, key))
			if err != nil {
				log.Error("redis get hashset[%v] key[%v] data error[%v]", REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGE_COMMENTS, key, err.Error())
				continue
			}
		} else {
			key = fmt.Sprintf("%v_%v_%v_%v", player_id, pic_id, msg_id, this.CommentIds[i])
			value, err = redis.String(global_data.redis_conn.Do("HGET", REDIS_KEY_PERSONAL_SPACE_PIC_LEAVE_MESSAGE_COMMENTS, key))
			if err != nil {
				log.Error("redis get hashset[%v] key[%v] data error[%v]", REDIS_KEY_PERSONAL_SPACE_PIC_LEAVE_MESSAGE_COMMENTS, key, err.Error())
				continue
			}
		}

		jitem := &RedisPSComment{}
		if err := json.Unmarshal([]byte(value), jitem); err != nil {
			log.Error("##### unmarshal personal space player[%v] msg_id[%v] comment_id[%v] data[%v] error[%v]", player_id, msg_id, this.CommentIds[i], value, err.Error())
			continue
		}

		c := &PSLeaveMessageComment{
			id:             jitem.CommentId,
			content:        jitem.Content,
			send_player_id: jitem.SendPlayerId,
			send_time:      jitem.SendTime,
		}

		comments = append(comments, c)
	}

	return
}

// 载入个人空间基础数据
func (this *RedisGlobalData) LoadPersonalSpaceBaseData() int32 {
	int_map, err := redis.StringMap(global_data.redis_conn.Do("HGETALL", REDIS_KEY_PERSONAL_SPACE_BASE_DATA))
	if err != nil {
		log.Error("redis获取集合[%v]数据失败[%v]", REDIS_KEY_PERSONAL_SPACE_BASE_DATA, err.Error())
		return -1
	}

	for k, item := range int_map {
		jitem := &RedisPSBaseData{}
		if err := json.Unmarshal([]byte(item), jitem); err != nil {
			log.Error("##### Load Personal Space Base Data item[%v] error[%v]", item, err.Error())
			return -1
		}
		if !ps_mgr.LoadSpace(jitem) {
			log.Warn("载入集合[%v]数据[%v,%v]失败", REDIS_KEY_PERSONAL_SPACE_BASE_DATA, k, item)
		}
	}
	return 1
}

// 载入个人空间留言
func (this *RedisGlobalData) LoadPersonalSpaceLeaveMessages() int32 {
	key_map, err := redis.StringMap(global_data.redis_conn.Do("HGETALL", REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGES))
	if err != nil {
		log.Error("redis get hashset[%v] data error[%v]", REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGES, err.Error())
		return -1
	}

	for key, item := range key_map {
		jitem := &RedisPSLeaveMessage{}
		if err := json.Unmarshal([]byte(item), jitem); err != nil {
			log.Error("##### Load Personal Space Leave Message item[%v] error[%v]", item, err.Error())
			return -1
		}
		var player_id, msg_id int32
		_, err := fmt.Sscanf(key, "%d_%d", &player_id, &msg_id)
		if err != nil {
			log.Error("redis get hashset[%v] key[%v] data error[%v]", REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGES, key, err.Error())
			return -1
		}
		if !ps_leave_messages_mgr.LoadLeaveMessage(player_id, jitem) {
			return -1
		}
	}

	return 1
}

// 载入空间图片
func (this *RedisGlobalData) LoadPersonalSpacePictures() int32 {
	key_map, err := redis.StringMap(global_data.redis_conn.Do("HGETALL", REDIS_KEY_PERSONAL_SPACE_PICTURE_DATA))
	if err != nil {
		log.Error("redis get hashset[%v] data error[%v]", REDIS_KEY_PERSONAL_SPACE_PICTURE_DATA, err.Error())
		return -1
	}

	for key, item := range key_map {
		jitem := &RedisPSPicture{}
		if err := json.Unmarshal([]byte(item), jitem); err != nil {
			log.Error("##### unmarshal personal space picture data[%v] error[%v]", item, err.Error())
			return -1
		}
		var player_id, pic_id int32
		_, err := fmt.Sscanf(key, "%d_%d", &player_id, &pic_id)
		if err != nil {
			log.Error("redis get hashset[%v] key[%v] data error[%v]", REDIS_KEY_PERSONAL_SPACE_PICTURE_DATA, key, err.Error())
			return -1
		}
		if !ps_pic_mgr.LoadPicture(player_id, jitem) {
			return -1
		}
	}
	return 1
}

// 载入空间图片留言
func (this *RedisGlobalData) LoadPersonalSpacePicLeaveMessages() int32 {
	key_map, err := redis.StringMap(global_data.redis_conn.Do("HGETALL", REDIS_KEY_PERSONAL_SPACE_PIC_LEAVE_MESSAGES))
	if err != nil {
		log.Error("redis get hashset[%v] data error[%v]", REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGES)
		return -1
	}

	for key, item := range key_map {
		jitem := &RedisPSLeaveMessage{}
		if err := json.Unmarshal([]byte(item), jitem); err != nil {
			log.Error("##### unmarshal personal space pic leave message item[%v] error[%v]", item, err.Error())
			return -1
		}
		var player_id, pic_id, msg_id int32
		_, err := fmt.Sscanf(key, "%d_%d_%d", &player_id, &pic_id, &msg_id)
		if err != nil {
			log.Error("redis get hashset[%v] key[%v] data error[%v]", REDIS_KEY_PERSONAL_SPACE_PIC_LEAVE_MESSAGES, key, err.Error())
			return -1
		}
		if !ps_pic_leave_messages_mgr.LoadRedisLeaveMessage(player_id, pic_id, jitem) {
			return -1
		}
	}
	return 1
}

// 载入空间图片点赞
func (this *RedisGlobalData) LoadPersonalSpacePicZan() int32 {
	keys, err := redis.Strings(global_data.redis_conn.Do("SMEMBERS", REDIS_KEY_PERSONAL_SPACE_PIC_ZAN))
	if err != nil {
		log.Error("redis get set[%v] data error[%v]", REDIS_KEY_PERSONAL_SPACE_PIC_ZAN, err.Error())
		return -1
	}

	for _, key := range keys {
		var player_id, pic_id, send_player_id int32
		_, err = fmt.Sscanf(key, "%d_%d_%d", &player_id, &pic_id, &send_player_id)
		if err != nil {
			log.Error("redis get set[%v] key[%v] data error[%v]", REDIS_KEY_PERSONAL_SPACE_PIC_ZAN, key, err.Error())
			return -1
		}
		if ps_pic_zan_mgr.CheckAndZan(player_id, pic_id, send_player_id) < 0 {
			log.Warn("add player_id[%v] zan player_id[%v] pic[%v] exists", send_player_id, player_id, pic_id)
			continue
		}
	}
	return 1
}

// 更新个人空间基础数据
func (this *RedisGlobalData) UpdatePersonalSpaceBaseData(player_id int32) int32 {
	data := ps_mgr.ToRedisData(player_id)
	if data == nil {
		log.Error("player[%v] personal space base data to redis data failed", player_id)
		return -1
	}
	bytes, err := json.Marshal(data)
	if err != nil {
		log.Error("##### serialize item[%v] error[%v]", *data, err.Error())
		return -1
	}
	err = this.redis_conn.Post("HSET", REDIS_KEY_PERSONAL_SPACE_BASE_DATA, player_id, string(bytes))
	if err != nil {
		log.Error("redis set hashset[%v] data[%v] error[%v]", REDIS_KEY_PERSONAL_SPACE_BASE_DATA, *data, err.Error())
		return -1
	}
	return 1
}

// 更新个人空间留言
func (this *RedisGlobalData) UpdatePersonalSpaceLeaveMessage(player_id int32, msg_id int32) int32 {
	data := ps_leave_messages_mgr.ToLeaveMsgRedisData(player_id, msg_id)
	if data == nil {
		log.Error("player[%v] personal space leave message[%v] to redis data failed", player_id, msg_id)
		return -1
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Error("##### serialize item[%v] error[%v]", *data, err.Error())
		return -1
	}
	key := fmt.Sprintf("%v_%v", player_id, msg_id)
	err = this.redis_conn.Post("HSET", REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGES, key, string(bytes))
	if err != nil {
		log.Error("redis set hashset[%v] data[%v] error[%v]", REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGES, *data, err.Error())
		return -1
	}
	return 1
}

// 更新个人空间评论
func (this *RedisGlobalData) UpdatePersonalSpaceComment(player_id int32, msg_id int32, comment_id int32) int32 {
	data := ps_leave_messages_mgr.ToCommentRedisData(player_id, msg_id, comment_id)
	if data == nil {
		log.Error("player[%v] personal space leave message[%v] comment[%v] to redis data failed", player_id, msg_id, comment_id)
		return -1
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Error("##### serialize item[%v] error[%v]", *data, err.Error())
		return -1
	}

	key := fmt.Sprintf("%v_%v_%v", player_id, msg_id, comment_id)
	err = this.redis_conn.Post("HSET", REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGE_COMMENTS, key, string(bytes))
	if err != nil {
		log.Error("redis set hashset[%v] data[%v] error[%v]", REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGE_COMMENTS, *data, err.Error())
		return -1
	}

	return 1
}

// 更新个人空间图片数据
func (this *RedisGlobalData) UpdatePersonalSpacePictureData(player_id int32, pic_id int32) int32 {
	data := ps_pic_mgr.ToPicRedisData(player_id, pic_id)
	if data == nil {
		log.Error("player[%v] personal space picture[%v] to redis data failed", player_id, pic_id)
		return -1
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Error("##### serialize item[%v] error[%v]", *data, err.Error())
		return -1
	}

	key := fmt.Sprintf("%v_%v", player_id, pic_id)
	err = this.redis_conn.Post("HSET", REDIS_KEY_PERSONAL_SPACE_PICTURE_DATA, key, string(bytes))
	if err != nil {
		log.Error("redis set hashset[%v] data[%v] error[%v]", REDIS_KEY_PERSONAL_SPACE_PICTURE_DATA, *data, err.Error())
		return -1
	}

	return 1
}

// 更新个人空间图片留言
func (this *RedisGlobalData) UpdatePersonalSpacePicLeaveMsg(player_id, pic_id, msg_id int32) int32 {
	data := ps_pic_leave_messages_mgr.ToLeaveMsgRedisData(player_id, pic_id, msg_id)
	if data == nil {
		log.Error("player[%v] personal space picture[%v] leave msg[%v] to redis data failed", player_id, pic_id, msg_id)
		return -1
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Error("##### serialize item[%v] error[%v]", *data, err.Error())
		return -1
	}

	key := fmt.Sprintf("%v_%v_%v", player_id, pic_id, msg_id)
	err = this.redis_conn.Post("HSET", REDIS_KEY_PERSONAL_SPACE_PIC_LEAVE_MESSAGES, key, string(bytes))
	if err != nil {
		log.Error("redis hashset[%v] data[%v] error[%v]", REDIS_KEY_PERSONAL_SPACE_PIC_LEAVE_MESSAGES, *data, err.Error())
		return -1
	}

	return 1
}

// 更新个人空间图片评论
func (this *RedisGlobalData) UpdatePersonalSpacePicComment(player_id, pic_id, msg_id, comment_id int32) int32 {
	data := ps_pic_leave_messages_mgr.ToCommentRedisData(player_id, pic_id, msg_id, comment_id)
	if data == nil {
		log.Error("player[%v] personal space picture[%v] leave msg[%v] comment[%v] to redis data failed", player_id, pic_id, msg_id, comment_id)
		return -1
	}

	bytes, err := json.Marshal(data)
	if err != nil {
		log.Error("##### serialize item[%v] error[%v]", *data, err.Error())
		return -1
	}

	key := fmt.Sprintf("%v_%v_%v_%v", player_id, pic_id, msg_id, comment_id)
	err = this.redis_conn.Post("HSET", REDIS_KEY_PERSONAL_SPACE_PIC_LEAVE_MESSAGE_COMMENTS, key, string(bytes))
	if err != nil {
		log.Error("redis hashset[%v] data[%v] error[%v]", REDIS_KEY_PERSONAL_SPACE_PIC_LEAVE_MESSAGE_COMMENTS, *data, err.Error())
		return -1
	}

	return 1
}

// 更新个人空间图片赞
func (this *RedisGlobalData) UpdatePersonalSpacePicZan(player_id, pic_id, send_player_id int32) int32 {
	key := fmt.Sprintf("%v_%v_%v", player_id, pic_id, send_player_id)
	err := this.redis_conn.Post("SADD", REDIS_KEY_PERSONAL_SPACE_PIC_ZAN, key)
	if err != nil {
		log.Error("redis set[%v] add data[%v] error[%v]", REDIS_KEY_PERSONAL_SPACE_PIC_ZAN, key, err.Error())
		return -1
	}
	return 1
}

// 删除空间留言
func (this *RedisGlobalData) DeletePersonalSpaceLeaveMsg(player_id, msg_id int32) int32 {
	key := fmt.Sprintf("%v_%v", player_id, msg_id)
	err := this.redis_conn.Post("HDEL", REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGES, key)
	if err != nil {
		log.Error("redis hashset[%v] delete key[%v] error[%v]", REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGES, key, err.Error())
		return -1
	}
	return 1
}

// 删除空间评论
func (this *RedisGlobalData) DeletePersonalSpaceComment(player_id, msg_id, comment_id int32) int32 {
	key := fmt.Sprintf("%v_%v_%v", player_id, msg_id, comment_id)
	err := this.redis_conn.Post("HDEL", REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGE_COMMENTS, key)
	if err != nil {
		log.Error("redis hashset[%v] delete key[%v] error[%v]", REDIS_KEY_PERSONAL_SPACE_LEAVE_MESSAGE_COMMENTS, key, err.Error())
		return -1
	}
	return 1
}

// 删除图片
func (this *RedisGlobalData) DeletePersonalSpacePicture(player_id, pic_id int32) int32 {
	key := fmt.Sprintf("%v_%v", player_id, pic_id)
	err := this.redis_conn.Post("HDEL", REDIS_KEY_PERSONAL_SPACE_PICTURE_DATA, key)
	if err != nil {
		log.Error("redis hashset[%v] delete key[%v] error[%v]", REDIS_KEY_PERSONAL_SPACE_PICTURE_DATA, key, err.Error())
		return -1
	}
	return 1
}

// 删除图片留言
func (this *RedisGlobalData) DeletePersonalSpacePicLeaveMsg(player_id, pic_id, msg_id int32) int32 {
	key := fmt.Sprintf("%v_%v_%v", player_id, pic_id, msg_id)
	err := this.redis_conn.Post("HDEL", REDIS_KEY_PERSONAL_SPACE_PIC_LEAVE_MESSAGES, key)
	if err != nil {
		log.Error("redis hashset[%v] delete key[%v] error[%v]", REDIS_KEY_PERSONAL_SPACE_PIC_LEAVE_MESSAGES, key, err.Error())
		return -1
	}
	return 1
}

// 删除图片评论
func (this *RedisGlobalData) DeletePersonalSpacePicComment(player_id, pic_id, msg_id, comment_id int32) int32 {
	key := fmt.Sprintf("%v_%v_%v_%v", player_id, pic_id, msg_id, comment_id)
	err := this.redis_conn.Post("HDEL", REDIS_KEY_PERSONAL_SPACE_PIC_LEAVE_MESSAGE_COMMENTS, key)
	if err != nil {
		log.Error("redis hashset[%v] delete key[%v] error[%v]", REDIS_KEY_PERSONAL_SPACE_PIC_LEAVE_MESSAGE_COMMENTS, key, err.Error())
		return -1
	}
	return 1
}

// 删除图片赞
func (this *RedisGlobalData) DeletePersonalSpacePicZan(player_id, pic_id, send_player_id int32) int32 {
	key := fmt.Sprintf("%v_%v_%v", player_id, pic_id, send_player_id)
	err := this.redis_conn.Post("SREM", REDIS_KEY_PERSONAL_SPACE_PIC_ZAN, key)
	if err != nil {
		log.Error("redis set[%v] delete key[%v] error[%v]", REDIS_KEY_PERSONAL_SPACE_PIC_ZAN, key, err.Error())
		return -1
	}
	return 1
}
