package main

import (
	"libs/log"
	"libs/utils"
	_ "main/table_config"
	"math/rand"
	"net/http"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	"strconv"
	"sync"
	"time"

	"github.com/golang/protobuf/proto"
)

const (
	GUILD_MAX_NUM       = 10000 // 公会最大数目
	GUILD_RECOMMEND_NUM = 5     // 公会推荐数目
	GUILD_LOG_MAX_NUM   = 10    // 公会日志数
)

const (
	GUILD_EXIST_TYPE_NONE        = iota // 未删除
	GUILD_EXIST_TYPE_WILL_DELETE = 1    // 将删除
	GUILD_EXIST_TYPE_DELETED     = 2    // 已删除
)

const (
	GUILD_POSITION_MEMBER    = iota // 成员
	GUILD_POSITION_PRESIDENT = 1    // 会长
	GUILD_POSITION_OFFICER   = 2    // 官员
)

const (
	GUILD_LOG_TYPE_NONE             = iota
	GUILD_LOG_TYPE_CREATE           = 1
	GUILD_LOG_TYPE_MEMBER_JOIN      = 2
	GUILD_LOG_TYPE_MEMBER_QUIT      = 3
	GUILD_LOG_TYPE_MEMBER_KICK      = 4
	GUILD_LOG_TYPE_MEMBER_APPOINT   = 5
	GUILD_LOG_TYPE_OFFICER_DISMISS  = 6
	GUILD_LOG_TYPE_PRESIDENT_CHANGE = 7
)

func _player_get_guild_id(player_id int32) int32 {
	player := player_mgr.GetPlayerById(player_id)
	if player == nil {
		return int32(msg_client_message.E_ERR_PLAYER_NOT_EXIST)
	}
	guild_id := player.db.Guild.GetId()
	return guild_id
}

func _remove_guild(guild *dbGuildRow) {
	guild.SetExistType(GUILD_EXIST_TYPE_DELETED)
	guild.Members.Clear()
	guild.AskLists.Clear()
	guild_manager._remove_guild(guild.GetId(), guild.GetName())
}

type GuildManager struct {
	guilds           *dbGuildTable
	guild_ids        []int32
	guild_num        int32
	guild_id_map     map[int32]int32
	guild_name_map   map[string]int32
	guild_ids_locker *sync.RWMutex
	guild_chat_map   map[int32]*ChatMgr
}

var guild_manager GuildManager

func (this *GuildManager) _add_guild(guild_id int32, guild_name string) bool {
	this.guild_ids_locker.Lock()
	defer this.guild_ids_locker.Unlock()

	if _, o := this.guild_id_map[guild_id]; o {
		return false
	}
	if _, o := this.guild_name_map[guild_name]; o {
		return false
	}
	this.guild_ids[this.guild_num] = guild_id
	this.guild_num += 1
	this.guild_id_map[guild_id] = guild_id
	this.guild_name_map[guild_name] = guild_id

	chat_mgr := &ChatMgr{}
	chat_mgr.Init(CHAT_CHANNEL_GUILD)
	this.guild_chat_map[guild_id] = chat_mgr

	return true
}

func (this *GuildManager) _remove_guild(guild_id int32, guild_name string) bool {
	this.guild_ids_locker.Lock()
	defer this.guild_ids_locker.Unlock()

	if _, o := this.guild_id_map[guild_id]; !o {
		return false
	}
	if _, o := this.guild_name_map[guild_name]; !o {
		return false
	}
	for i := int32(0); i < this.guild_num; i++ {
		if this.guild_ids[i] == guild_id {
			this.guild_ids[i] = this.guild_ids[this.guild_num-1]
			this.guild_num -= 1
			break
		}
	}
	delete(this.guild_id_map, guild_id)
	delete(this.guild_name_map, guild_name)
	delete(this.guild_chat_map, guild_id)
	return true
}

func (this *GuildManager) Init() {
	this.guilds = dbc.Guilds
	this.guild_ids = make([]int32, GUILD_MAX_NUM)
	this.guild_id_map = make(map[int32]int32)
	this.guild_name_map = make(map[string]int32)
	this.guild_chat_map = make(map[int32]*ChatMgr)
	this.guild_ids_locker = &sync.RWMutex{}
	for gid, guild := range this.guilds.m_rows {
		if _guild_get_exist_type(guild) == GUILD_EXIST_TYPE_DELETED {
			continue
		}
		this._add_guild(gid, guild.GetName())
	}
}

func (this *GuildManager) CreateGuild(player_id int32, guild_name string, logo int32) int32 {
	player := player_mgr.GetPlayerById(player_id)
	if player == nil {
		return int32(msg_client_message.E_ERR_PLAYER_NOT_EXIST)
	}
	guild_id := player.db.Guild.GetId()
	if guild_id > 0 && this.GetGuild(guild_id) != nil {
		log.Error("Player[%v] already create guild[%v|%v]", player_id, guild_name, guild_id)
		return -1
	}

	row := this.guilds.AddRow()
	if row == nil {
		log.Error("Player[%v] create guild add db row failed", player_id)
		return -1
	}

	row.SetName(guild_name)
	row.SetCreater(player_id)
	row.SetCreateTime(int32(time.Now().Unix()))
	row.SetLevel(1)
	row.SetLogo(logo)
	row.SetPresident(player_id)
	row.Members.Add(&dbGuildMemberData{
		PlayerId: player_id,
	})
	guild_id = row.GetId()

	player.db.Guild.SetId(guild_id)
	player.db.Guild.SetPosition(GUILD_POSITION_PRESIDENT)

	this._add_guild(guild_id, guild_name)

	return guild_id
}

func (this *GuildManager) GetGuild(guild_id int32) *dbGuildRow {
	guild := this.guilds.GetRow(guild_id)
	if guild == nil {
		return nil
	}
	exist_type := _guild_get_exist_type(guild)
	if exist_type == GUILD_EXIST_TYPE_DELETED {
		return nil
	}
	return guild
}

func (this *GuildManager) GetChatMgr(guild_id int32) *ChatMgr {
	this.guild_ids_locker.RLock()
	defer this.guild_ids_locker.RUnlock()

	return this.guild_chat_map[guild_id]
}

func (this *GuildManager) Recommend(player_id int32) (guild_ids []int32) {
	guild_id := _player_get_guild_id(player_id)
	if guild_id > 0 && this.GetGuild(guild_id) != nil {
		log.Error("Player[%v] already joined one guild", player_id)
		return
	}

	this.guild_ids_locker.RLock()
	defer this.guild_ids_locker.RUnlock()

	if this.guild_num == 0 {
		log.Error("No guild to recommend")
		return
	}

	rand.Seed(time.Now().Unix() + time.Now().UnixNano())

	for i := 0; i < GUILD_RECOMMEND_NUM; i++ {
		r := rand.Int31n(this.guild_num)
		sr := r
		for {
			has := false
			if guild_ids != nil {
				for n := 0; n < len(guild_ids); n++ {
					if guild_ids[n] == this.guild_ids[sr] {
						has = true
						break
					}
				}
			}
			if !has {
				break
			}
			sr = (sr + 1) % this.guild_num
			if sr == r {
				log.Info("GuildManager Recommend guild count[%v] not enough to random for recommend", this.guild_num)
				return
			}
		}
		guild_id = this.guild_ids[sr]
		if guild_id <= 0 {
			break
		}
		guild_ids = append(guild_ids, guild_id)
	}
	return
}

func (this *GuildManager) Search(key string) (guild_ids []int32) {
	guild_id, _ := strconv.Atoi(key)

	this.guild_ids_locker.RLock()
	defer this.guild_ids_locker.RUnlock()

	if this.guild_id_map[int32(guild_id)] > 0 {
		guild_ids = []int32{int32(guild_id)}
	}

	if guild_id, o := this.guild_name_map[key]; o {
		guild_ids = append(guild_ids, guild_id)
	}
	return
}

func (this *GuildManager) _get_guild(player_id int32, is_president bool) (guild *dbGuildRow) {
	player := player_mgr.GetPlayerById(player_id)
	if player == nil {
		return nil
	}
	guild_id := player.db.Guild.GetId()
	if guild_id <= 0 {
		return nil
	}
	guild = this.GetGuild(guild_id)
	if guild == nil || (is_president && guild.GetPresident() != player_id) {
		return nil
	}

	position := player.db.Guild.GetPosition()
	if position <= GUILD_POSITION_MEMBER && guild.GetPresident() == player_id {
		player.db.Guild.SetPosition(GUILD_POSITION_PRESIDENT)
	} else if position == GUILD_POSITION_PRESIDENT && guild.GetPresident() != player_id {
		guild.SetPresident(player_id)
	}

	if guild.GetPresident() == player_id && !guild.Members.HasIndex(player_id) {
		guild.Members.Add(&dbGuildMemberData{
			PlayerId: player_id,
		})
	}

	return guild
}

func _guild_member_num_limit(guild *dbGuildRow) int32 {
	levelup_data := guild_levelup_table_mgr.Get(guild.GetLevel())
	if levelup_data == nil {
		return 0
	}
	return levelup_data.MemberNum
}

func _format_guild_base_info_to_msg(guild *dbGuildRow) (msg *msg_client_message.GuildBaseInfo) {
	msg = &msg_client_message.GuildBaseInfo{
		Id:             guild.GetId(),
		Name:           guild.GetName(),
		Level:          guild.GetLevel(),
		Logo:           guild.GetLogo(),
		MemberNum:      guild.Members.NumAll(),
		MemberNumLimit: _guild_member_num_limit(guild),
	}
	return
}

func _format_guilds_base_info_to_msg(guild_ids []int32) (guilds_msg []*msg_client_message.GuildBaseInfo) {
	for _, gid := range guild_ids {
		guild := guild_manager.GetGuild(gid)
		if guild == nil {
			continue
		}
		guild_msg := _format_guild_base_info_to_msg(guild)
		guilds_msg = append(guilds_msg, guild_msg)
	}
	return
}

func _guild_get_dismiss_remain_seconds(guild *dbGuildRow) (dismiss_remain_seconds int32) {
	if guild.GetExistType() != GUILD_EXIST_TYPE_WILL_DELETE {
		return
	}
	dismiss_time := guild.GetDismissTime()
	dismiss_remain_seconds = GetRemainSeconds(dismiss_time, global_config.GuildDismissWaitingSeconds)
	if dismiss_remain_seconds == 0 {
		// 广播
		member_ids := guild.Members.GetAllIndex()
		if member_ids != nil {
			notify := &msg_client_message.S2CGuildDeleteNotify{
				GuildId: guild.GetId(),
			}
			for _, mid := range member_ids {
				p := player_mgr.GetPlayerById(mid)
				if p == nil {
					continue
				}
				p.db.Guild.SetId(0)
				p.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_DELETE_NOTIFY), notify)
				SendMail(nil, mid, MAIL_TYPE_GUILD, "Guild Dismissed", "Guild Dismissed", nil)
			}
		}
		_remove_guild(guild)
	}
	return
}

func _guild_get_exist_type(guild *dbGuildRow) int32 {
	_guild_get_dismiss_remain_seconds(guild)
	return guild.GetExistType()
}

// 公会日志
func push_new_guild_log(guild *dbGuildRow, log_type, player_id int32) {
	player := player_mgr.GetPlayerById(player_id)
	if player == nil {
		return
	}

	var min_id int32
	ids := guild.Logs.GetAllIndex()
	if ids != nil && len(ids) >= int(GUILD_LOG_MAX_NUM) {
		for _, id := range ids {
			if min_id == 0 || id < min_id {
				min_id = id
			}
		}
		if min_id > 0 {
			guild.Logs.Remove(min_id)
		}
	}

	guild.Logs.Add(&dbGuildLogData{
		Time:     int32(time.Now().Unix()),
		LogType:  log_type,
		PlayerId: player_id,
	})
}

func (this *Player) _format_guild_info_to_msg(guild *dbGuildRow) (msg *msg_client_message.GuildInfo) {
	level := guild.GetLevel()
	if level <= 0 {
		level = 1
		guild.SetLevel(level)
	}
	var dismiss_remain_seconds, sign_remain_seconds, ask_donate_remain_seconds, donate_reset_remain_seconds int32
	dismiss_remain_seconds = _guild_get_dismiss_remain_seconds(guild)
	sign_remain_seconds = utils.GetRemainSeconds2NextDayTime(this.db.Guild.GetSignTime(), global_config.GuildSignRefreshTime)
	ask_donate_remain_seconds = GetRemainSeconds(this.db.Guild.GetLastAskDonateTime(), global_config.GuildAskDonateCDSecs)
	donate_reset_remain_seconds = utils.GetRemainSeconds2NextDayTime(guild.GetLastDonateRefreshTime(), global_config.GuildDonateRefreshTime)
	president_id := guild.GetPresident()
	var president_name string
	president := player_mgr.GetPlayerById(president_id)
	if president != nil {
		president_name = president.db.GetName()
	}
	msg = &msg_client_message.GuildInfo{
		Id:                       guild.GetId(),
		Name:                     guild.GetName(),
		Level:                    level,
		Exp:                      guild.GetExp(),
		Logo:                     guild.GetLogo(),
		Anouncement:              guild.GetAnouncement(),
		DismissRemainSeconds:     dismiss_remain_seconds,
		SignRemainSeconds:        sign_remain_seconds,
		AskDonateRemainSeconds:   ask_donate_remain_seconds,
		DonateResetRemainSeconds: donate_reset_remain_seconds,
		President:                president_id,
		PresidentName:            president_name,
		MemberNum:                guild.Members.NumAll(),
		MemberNumLimit:           _guild_member_num_limit(guild),
	}
	return
}

// 公会基本数据
func (this *Player) send_guild_data() int32 {
	if this.db.Info.GetLvl() < global_config.GuildOpenLevel {
		log.Error("Player[%v] level not enough to open guild", this.Id)
		return -1
	}
	guild_id := this.db.Guild.GetId()
	if guild_id <= 0 {
		log.Error("Player[%v] no guild data", this.Id)
		return -1
	}
	guild := guild_manager.GetGuild(guild_id)
	if guild == nil {
		return -1
	}

	response := &msg_client_message.S2CGuildDataResponse{
		Info: this._format_guild_info_to_msg(guild),
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_DATA_RESPONSE), response)

	log.Debug("Player[%v] guild data %v", this.Id, response)

	return 1
}

// 公会推荐
func (this *Player) guild_recommend() int32 {
	gids := guild_manager.Recommend(this.Id)
	if gids == nil {
		return -1
	}

	guilds_msg := _format_guilds_base_info_to_msg(gids)

	response := &msg_client_message.S2CGuildRecommendResponse{
		InfoList: guilds_msg,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_RECOMMEND_RESPONSE), response)

	log.Debug("Player[%v] recommend guilds %v", this.Id, response)

	return 1
}

// 公会搜索
func (this *Player) guild_search(key string) int32 {
	if this.db.Guild.GetId() > 0 {
		log.Error("Player[%v] already joined one guild, cant search", this.Id)
		return -1
	}

	var guilds_msg []*msg_client_message.GuildBaseInfo
	guild_ids := guild_manager.Search(key)
	if guild_ids != nil {
		guilds_msg = _format_guilds_base_info_to_msg(guild_ids)
	} else {
		guilds_msg = make([]*msg_client_message.GuildBaseInfo, 0)
	}
	response := &msg_client_message.S2CGuildSearchResponse{
		InfoList: guilds_msg,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_SEARCH_RESPONSE), response)

	log.Debug("Player[%v] searched guild %v with key %v", this.Id, response, key)

	return 1
}

// 公会创建
func (this *Player) guild_create(name string, logo int32) int32 {
	if this.db.Info.GetLvl() < global_config.GuildOpenLevel {
		log.Error("Player[%v] cant create guild because level not enough", this.Id)
		return -1
	}

	if this.get_diamond() < global_config.GuildCreateCostGem {
		log.Error("Player[%v] create guild need diamond %v, but only %v", this.Id, global_config.GuildCreateCostGem, this.get_diamond())
		return -1
	}

	guild_id := guild_manager.CreateGuild(this.Id, name, logo)
	if guild_id < 0 {
		log.Error("Player[%v] create guild failed, err %v", this.Id, guild_id)
		return guild_id
	}

	this.add_diamond(-global_config.GuildCreateCostGem)

	guild := guild_manager.GetGuild(guild_id)
	if guild == nil {
		return -1
	}
	guild_msg := this._format_guild_info_to_msg(guild)
	response := &msg_client_message.S2CGuildCreateResponse{
		Info: guild_msg,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_CREATE_RESPONSE), response)

	// 日志
	push_new_guild_log(guild, GUILD_LOG_TYPE_CREATE, this.Id)

	log.Debug("Player[%v] created guild %v", this.Id, response)

	return 1
}

func (this *Player) get_guild() (guild *dbGuildRow) {
	guild_id := this.db.Guild.GetId()
	if guild_id <= 0 {
		return
	}
	return guild_manager.GetGuild(guild_id)
}

// 公会解散
func (this *Player) guild_dismiss() int32 {
	guild := guild_manager._get_guild(this.Id, true)
	if guild == nil {
		log.Error("Player[%v] cant dismiss guild", this.Id)
		return -1
	}
	if guild.GetExistType() != GUILD_EXIST_TYPE_NONE {
		log.Error("Player[%v] cant dismiss guild because guild exist type is %v", this.Id, guild.GetExistType())
		return -1
	}
	guild.SetDismissTime(int32(time.Now().Unix()))
	guild.SetExistType(GUILD_EXIST_TYPE_WILL_DELETE)
	response := &msg_client_message.S2CGuildDismissResponse{
		RealDismissRemainSeconds: global_config.GuildDismissWaitingSeconds,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_DISMISS_RESPONSE), response)

	log.Debug("Player[%v] dismiss guild %v", this.Id, response)

	return 1
}

// 公会取消解散
func (this *Player) guild_cancel_dismiss() int32 {
	guild := guild_manager._get_guild(this.Id, true)
	if guild == nil {
		log.Error("Player[%v] cant cancel dismissing guild", this.Id)
		return -1
	}
	if guild.GetExistType() != GUILD_EXIST_TYPE_WILL_DELETE {
		log.Error("Player[%v] cant cancel dismissing guild because guild exit type is %v", this.Id, guild.GetExistType())
		return -1
	}
	guild.SetDismissTime(0)
	guild.SetExistType(GUILD_EXIST_TYPE_NONE)
	response := &msg_client_message.S2CGuildCancelDismissResponse{}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_CANCEL_DISMISS_RESPONSE), response)

	log.Debug("Player[%v] cancelled dismiss guild", this.Id)

	return 1
}

// 公会修改名称或logo
func (this *Player) guild_info_modify(name string, logo int32) int32 {
	guild := guild_manager._get_guild(this.Id, true)
	if guild == nil {
		log.Error("Player[%v] cant get guild", this.Id)
		return -1
	}

	var cost_diamond int32
	if name != "" {
		if this.get_diamond() < global_config.GuildChangeNameCostGem {
			log.Error("Player[%v] diamond not enough, change name failed", this.Id)
			return -1
		}
		guild.SetName(name)
		cost_diamond = global_config.GuildChangeNameCostGem
		this.add_diamond(-cost_diamond)
	}

	if logo != 0 {
		guild.SetLogo(logo)
	}

	response := &msg_client_message.S2CGuildInfoModifyResponse{
		NewGuildName: name,
		NewGuildLogo: logo,
		CostDiamond:  cost_diamond,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_INFO_MODIFY_RESPONSE), response)

	log.Debug("Player[%v] modified guild info %v", this.Id, response)

	return 1
}

// 公会公告
func (this *Player) guild_anouncement(content string) int32 {
	guild := guild_manager._get_guild(this.Id, true)
	if guild == nil {
		log.Error("Player[%v] cant get guild", this.Id)
		return -1
	}

	guild.SetAnouncement(content)
	response := &msg_client_message.S2CGuildAnouncementResponse{
		Content: content,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_ANOUNCEMENT_RESPONSE), response)

	return 1
}

func (this *Player) _format_guild_member_to_msg() (msg *msg_client_message.GuildMember) {
	var last_online_time int32
	now_time := int32(time.Now().Unix())
	if this.is_logout {
		last_online_time = now_time - this.db.Info.GetLastLogout()
		if last_online_time < 0 {
			last_online_time = 0
		}
	}
	var next_sign_remain_seconds int32
	next_sign_remain_seconds = utils.GetRemainSeconds2NextDayTime(this.db.Guild.GetSignTime(), global_config.GuildSignRefreshTime)
	msg = &msg_client_message.GuildMember{
		Id:                    this.Id,
		Name:                  this.db.GetName(),
		Level:                 this.db.Info.GetLvl(),
		Head:                  this.db.Info.GetHead(),
		Position:              this.db.Guild.GetPosition(),
		LastOnlineTime:        last_online_time,
		NextSignRemainSeconds: next_sign_remain_seconds,
	}
	return
}

// 公会成员列表
func (this *Player) guild_members_list() int32 {
	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		log.Error("Player[%v] no guild", this.Id)
		return -1
	}

	var members_msg []*msg_client_message.GuildMember
	member_ids := guild.Members.GetAllIndex()
	if member_ids != nil {
		for _, mid := range member_ids {
			mem := player_mgr.GetPlayerById(mid)
			if mem == nil {
				guild.Members.Remove(mid)
				continue
			}
			msg := mem._format_guild_member_to_msg()
			members_msg = append(members_msg, msg)
		}
	}

	response := &msg_client_message.S2CGuildMemebersResponse{
		Members: members_msg,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_MEMBERS_RESPONSE), response)

	log.Debug("Player[%v] get guild[%v] members %v", this.Id, guild.GetId(), response)

	return 1
}

// 公会申请加入
func (this *Player) guild_ask_join(guild_id int32) int32 {
	last_ask_time := this.db.Guild.GetQuitTime()
	if last_ask_time > 0 {
		now_time := int32(time.Now().Unix())
		if now_time-last_ask_time < global_config.GuildQuitAskJoinCDSecs {
			log.Error("Player[%v] is already in cool down to last quit", this.Id)
			return -1
		}
	}

	if _player_get_guild_id(this.Id) > 0 {
		log.Error("Player[%v] already joined guild", this.Id)
		return -1
	}

	guild := guild_manager.GetGuild(guild_id)
	if guild == nil {
		log.Error("Player[%v] ask join guild[%v] not found", this.Id, guild_id)
		return -1
	}

	if guild.Members.HasIndex(this.Id) {
		log.Error("Player[%v] already joined guild %v", this.Id, guild_id)
		return -1
	}

	if guild.AskLists.HasIndex(this.Id) {
		log.Warn("Player[%v] already asked join guild %v", this.Id, guild_id)
	} else {
		guild.AskLists.Add(&dbGuildAskListData{
			PlayerId: this.Id,
		})
	}

	response := &msg_client_message.S2CGuildAskJoinResponse{
		GuildId: guild_id,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_ASK_JOIN_RESPONSE), response)

	log.Debug("Player[%v] asked join guild %v", this.Id, guild_id)

	return 1
}

// 公会同意申请加入
func (this *Player) guild_agree_join(player_id int32) int32 {
	player := player_mgr.GetPlayerById(player_id)
	if player == nil {
		log.Error("Player[%v] not found", player_id)
		return int32(msg_client_message.E_ERR_PLAYER_NOT_EXIST)
	}

	// 是否已是其他工会的成员
	if player.db.Guild.GetId() > 0 {
		log.Error("Player[%v] already joined other guild", player_id)
		return -1
	}

	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		log.Error("Player[%v] cant get guild", this.Id)
		return -1
	}

	// 是否已申请
	if !guild.AskLists.HasIndex(player_id) {
		log.Error("Player[%v] not found in guild[%v] ask list", player_id, guild.GetId())
		return -1
	}

	// 是否已是工会成员
	if guild.Members.HasIndex(player_id) {
		log.Error("Player[%v] already joined guild", player_id)
		return -1
	}

	// 升级配置数据
	levelup_data := guild_levelup_table_mgr.Get(guild.GetLevel())
	if levelup_data == nil {
		log.Error("Guild level up table data not found with level %v", guild.GetLevel())
		return -1
	}

	// 人数限制
	if levelup_data.MemberNum <= guild.Members.NumAll() {
		log.Error("Guild %v members num is max, player %v cant agree the player %v join", guild.GetId(), this.Id, player_id)
		return -1
	}

	// 职位
	if this.db.Guild.GetPosition() <= GUILD_POSITION_MEMBER {
		log.Error("Player[%v] no authority to agree new member join", this.Id)
		return -1
	}

	guild.Members.Add(&dbGuildMemberData{
		PlayerId: player_id,
	})
	guild.AskLists.Remove(player_id)

	player.db.Guild.SetId(guild.GetId())
	player.db.Guild.SetJoinTime(int32(time.Now().Unix()))

	response := &msg_client_message.S2CGuildAgreeJoinResponse{
		PlayerId: player_id,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_AGREE_JOIN_RESPONSE), response)

	// 通知加入的成员
	notify := &msg_client_message.S2CGuildAgreeJoinNotify{
		NewMemberId: player_id,
		GuildId:     guild.GetId(),
	}
	player.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_AGREE_JOIN_NOTIFY), notify)

	// 日志
	push_new_guild_log(guild, GUILD_LOG_TYPE_MEMBER_JOIN, player_id)

	log.Debug("Player[%v] agreed player[%v] join guild %v", this.Id, player_id, guild.GetId())

	return 1
}

// 公会申请列表
func (this *Player) guild_ask_list() int32 {
	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		log.Error("Player[%v] cant get guild or guild not found", this.Id)
		return -1
	}

	var info_list []*msg_client_message.PlayerInfo
	ids := guild.AskLists.GetAllIndex()
	if ids != nil {
		info_list = _format_players_info(ids)
	}

	response := &msg_client_message.S2CGuildAskListResponse{
		AskList: info_list,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_ASK_LIST_RESPONSE), response)

	log.Debug("Player[%v] get ask list %v", this.Id, response)

	return 1
}

func (this *Player) clear_guild_data() {
	this.db.Guild.SetId(0)
	this.db.Guild.SetPosition(0)
	this.db.Guild.SetDonateNum(0)
	this.db.Guild.SetSignTime(0)
	this.db.Guild.SetJoinTime(0)
	this.db.Guild.SetLastAskDonateTime(0)
	this.db.Guild.SetQuitTime(int32(time.Now().Unix()))
}

// 退出公会
func (this *Player) guild_quit() int32 {
	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		log.Error("Player[%v] cant get guild or guild not found", this.Id)
		return -1
	}

	if guild.GetPresident() == this.Id {
		log.Error("Player[%v] is president, cant quit guild", this.Id)
		return -1
	}

	guild.Members.Remove(this.Id)
	this.clear_guild_data()

	response := &msg_client_message.S2CGuildQuitResponse{
		RejoinRemainSeconds: global_config.GuildQuitAskJoinCDSecs,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_QUIT_RESPONSE), response)

	// 日志
	push_new_guild_log(guild, GUILD_LOG_TYPE_MEMBER_QUIT, this.Id)

	log.Debug("Player[%v] quit guild %v, rejoin remain seconds %v", this.Id, guild.GetId(), response.GetRejoinRemainSeconds())

	return 1
}

// 公会日志
func (this *Player) guild_logs() int32 {
	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		log.Error("Player[%v] cant get guild or guild not found", this.Id)
		return -1
	}

	var logs []*msg_client_message.GuildLog
	log_ids := guild.Logs.GetAllIndex()
	if log_ids != nil {
		for _, log_id := range log_ids {
			player_id, _ := guild.Logs.GetPlayerId(log_id)
			player := player_mgr.GetPlayerById(player_id)
			if player == nil {
				continue
			}
			log_type, _ := guild.Logs.GetLogType(log_id)
			log_time, _ := guild.Logs.GetTime(log_id)
			log := &msg_client_message.GuildLog{
				Id:         log_id,
				Type:       log_type,
				Time:       log_time,
				PlayerId:   player_id,
				PlayerName: player.db.GetName(),
			}
			logs = append(logs, log)
		}
	}

	response := &msg_client_message.S2CGuildLogsResponse{
		Logs: logs,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_LOGS_RESPONSE), response)

	log.Debug("Player[%v] get guild logs %v", this.Id, response)

	return 1
}

// 公会增加经验
func guild_add_exp(guild *dbGuildRow, add_exp int32) (level, exp int32, is_levelup bool) {
	old_level := guild.GetLevel()
	level = old_level
	if level <= 0 {
		level = 1
		guild.SetLevel(level)
	}
	exp = guild.GetExp() + add_exp

	for {
		level_data := guild_levelup_table_mgr.Get(level)
		if level_data == nil || level_data.Exp <= 0 {
			break
		}
		if level_data.Exp > exp {
			break
		}
		exp -= level_data.Exp
		level += 1
	}

	if level != old_level {
		guild.SetLevel(level)
	}
	guild.SetExp(exp)

	if level > old_level {
		is_levelup = true
	}

	return
}

// 公会签到
func (this *Player) guild_sign_in() int32 {
	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		log.Error("Player[%v] cant get guild", this.Id)
		return -1
	}

	if !utils.CheckDayTimeArrival(this.db.Guild.GetSignTime(), global_config.GuildSignRefreshTime) {
		log.Error("Player[%v] cant sign in guild, time not arrival", this.Id)
		return -1
	}

	now_time := int32(time.Now().Unix())
	this.db.Guild.SetSignTime(now_time)
	// 奖励
	if global_config.GuildSignReward != nil {
		for i := 0; i < len(global_config.GuildSignReward)/2; i++ {
			rid := global_config.GuildSignReward[2*i]
			rnum := global_config.GuildSignReward[2*i+1]
			this.add_resource(rid, rnum)
		}
	}
	// 增加经验
	level, exp, is_levelup := guild_add_exp(guild, global_config.GuildSignAddExp)

	next_remain_seconds := utils.GetRemainSeconds2NextDayTime(now_time, global_config.GuildSignRefreshTime)
	response := &msg_client_message.S2CGuildSignInResponse{
		NextSignInRemainSeconds: next_remain_seconds,
		RewardItems:             global_config.GuildSignReward,
		GuildLevel:              level,
		GuildExp:                exp,
		IsLevelup:               is_levelup,
		MemberNumLimit:          guild_levelup_table_mgr.GetMemberNumLimit(level),
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_SIGN_IN_RESPONSE), response)

	log.Debug("Player[%v] sign in guild[%v]", this.Id, guild.GetId())

	return 1
}

// 公会任免官员
func (this *Player) guild_set_officer(player_ids []int32, set_type int32) int32 {
	// 只有会长有权限
	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		log.Error("Player[%v] cant get guild or guild not exist", this.Id)
		return -1
	}

	var position int32
	if set_type == 1 {
		position = GUILD_POSITION_OFFICER
	} else if set_type == 2 {
		position = GUILD_POSITION_MEMBER
	}

	for i, player_id := range player_ids {
		if player_id == this.Id {
			player_ids[i] = 0
			continue
		}
		if !guild.Members.HasIndex(player_id) {
			player_ids[i] = 0
			log.Error("Player[%v] is not member of guild %v", player_id, guild.GetId())
			continue
		}

		player := player_mgr.GetPlayerById(player_id)
		if player == nil {
			player_ids[i] = 0
			log.Error("Player[%v] not found", player_id)
			continue
		}

		player.db.Guild.SetPosition(position)
	}

	response := &msg_client_message.S2CGuildSetOfficerResponse{
		PlayerIds: player_ids,
		SetType:   set_type,
		Position:  position,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_SET_OFFICER_RESPONSE), response)

	// 通知被任免成员
	notify := &msg_client_message.S2CGuildSetOfficerNotify{
		SetType:     set_type,
		NewPosition: position,
	}
	for _, player_id := range player_ids {
		player := player_mgr.GetPlayerById(player_id)
		if player == nil {
			continue
		}
		notify.MemberId = player_id
		player.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_SET_OFFICER_NOTIFY), notify)

		// 日志
		if set_type == 1 {
			push_new_guild_log(guild, GUILD_LOG_TYPE_MEMBER_APPOINT, player_id)
		} else if set_type == 2 {
			push_new_guild_log(guild, GUILD_LOG_TYPE_OFFICER_DISMISS, player_id)
		}
	}

	log.Debug("Player[%v] set officer %v in guild %v", this.Id, response, guild.GetId())

	return 1
}

// 公会驱逐会员
func (this *Player) guild_kick_member(player_ids []int32) int32 {
	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		log.Error("Player[%v] cant get guild or guild not exist", this.Id)
		return -1
	}

	if this.db.Guild.GetPosition() <= GUILD_POSITION_MEMBER {
		log.Error("Player[%v] position %v no authority to kick member", this.Id, this.db.Guild.GetPosition())
		return -1
	}

	for i, player_id := range player_ids {
		if player_id == this.Id {
			player_ids[i] = 0
			continue
		}
		if !guild.Members.HasIndex(player_id) {
			player_ids[i] = 0
			log.Error("Player[%v] is not member of guild[%v]", player_id, guild.GetId())
			continue
		}
		player := player_mgr.GetPlayerById(player_id)
		if player == nil {
			player_ids[i] = 0
			continue
		}
		if player.db.Guild.GetPosition() != GUILD_POSITION_MEMBER {
			player_ids[i] = 0
			continue
		}
		guild.Members.Remove(player_id)
		player.clear_guild_data()
	}

	response := &msg_client_message.S2CGuildKickMemberResponse{
		PlayerIds: player_ids,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_KICK_MEMBER_RESPONSE), response)

	// 通知被驱逐成员
	notify := &msg_client_message.S2CGuildKickMemberNotify{}
	for _, player_id := range player_ids {
		player := player_mgr.GetPlayerById(player_id)
		if player == nil {
			continue
		}
		notify.MemberId = player_id
		player.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_KICK_MEMBER_NOTIFY), notify)

		// 日志
		push_new_guild_log(guild, GUILD_LOG_TYPE_MEMBER_KICK, player_id)
	}

	log.Debug("Player[%v] kick members %v from guild %v", this.Id, player_ids, guild.GetId())

	return 1
}

// 公会转让会长
func (this *Player) guild_change_president(player_id int32) int32 {
	if player_id == this.Id {
		return -1
	}

	player := player_mgr.GetPlayerById(player_id)
	if player == nil {
		log.Error("Player[%v] not found", player_id)
		return -1
	}

	guild := guild_manager._get_guild(this.Id, true)
	if guild == nil {
		log.Error("Player[%v] cant get guild or guild not exist", this.Id)
		return -1
	}

	if !guild.Members.HasIndex(player_id) {
		log.Error("Guild %v no member %v, cant change president", guild.GetId(), player_id)
		return -1
	}

	guild.SetPresident(player_id)
	this.db.Guild.SetPosition(GUILD_POSITION_MEMBER)
	player.db.Guild.SetPosition(GUILD_POSITION_PRESIDENT)

	response := &msg_client_message.S2CGuildChangePresidentResponse{
		NewPresidentId: player_id,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_CHANGE_PRESIDENT_RESPONSE), response)

	// 通知新会长
	notify := &msg_client_message.S2CGuildChangePresidentNotify{
		OldPresidentId: this.Id,
		NewPresidentId: player_id,
	}
	player.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_CHANGE_PRESIDENT_NOTIFY), notify)

	// 日志
	push_new_guild_log(guild, GUILD_LOG_TYPE_PRESIDENT_CHANGE, player_id)

	log.Debug("Player[%v] change guild %v president to %v", this.Id, guild.GetId(), player_id)

	return 1
}

// 公会招募
func (this *Player) guild_recruit(content []byte) int32 {
	guild_id := this.db.Guild.GetId()
	if guild_id <= 0 {
		log.Error("Player[%v] no join in guild", this.Id)
		return -1
	}

	position := this.db.Guild.GetPosition()
	if position <= GUILD_POSITION_MEMBER {
		log.Error("Player[%v] recruit in guild %v failed, position %v not enough", this.Id, guild_id, position)
		return -1
	}

	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		log.Error("Player[%v] cant get guild", this.Id)
		return -1
	}

	now_time := int32(time.Now().Unix())
	last_recruit_time := guild.GetLastRecruitTime()
	if (now_time - last_recruit_time) < global_config.GuildRecruitIntervalSecs {
		log.Error("Player[%v] recruit too frequently", this.Id)
		return -1
	}

	res := this.chat(CHAT_CHANNEL_RECRUIT, content)
	if res < 0 {
		return res
	}

	guild.SetLastRecruitTime(now_time)

	response := &msg_client_message.S2CGuildRecruitResponse{
		Content: content,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_RECRUIT_RESPONSE), response)

	log.Debug("Player[%v] recruit with content[%v]", this.Id, content)

	return 1
}

func (this *Player) send_guild_donate_list(guild *dbGuildRow) {
	var donate_list []*msg_client_message.GuildAskDonateInfo
	player_ids := guild.AskDonates.GetAllIndex()
	if player_ids != nil {
		for _, player_id := range player_ids {
			player := player_mgr.GetPlayerById(player_id)
			if player == nil {
				continue
			}

			item_id, _ := guild.AskDonates.GetItemId(player_id)
			item_num, _ := guild.AskDonates.GetItemNum(player_id)
			ask_time, _ := guild.AskDonates.GetAskTime(player_id)
			name, level, head := GetPlayerBaseInfo(player_id)
			remain_exist_seconds := GetRemainSeconds(ask_time, global_config.GuildAskDonateCDSecs)
			donate_item := &msg_client_message.GuildAskDonateInfo{
				PlayerId:           player_id,
				PlayerName:         name,
				PlayerHead:         head,
				PlayerLevel:        level,
				ItemId:             item_id,
				ItemNum:            item_num,
				AskTime:            ask_time,
				RemainExistSeconds: remain_exist_seconds,
			}
			donate_list = append(donate_list, donate_item)
		}
	}

	response := &msg_client_message.S2CGuildDonateListResponse{
		InfoList: donate_list,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_DONATE_LIST_RESPONSE), response)

	log.Debug("Player[%v] get donate list %v", this.Id, response)
}

// 检测捐赠列表
func guild_check_donate_list(guild *dbGuildRow) (changed bool) {
	all_ids := guild.AskDonates.GetAllIndex()
	if all_ids == nil {
		return
	}
	for _, player_id := range all_ids {
		ask_time, _ := guild.AskDonates.GetAskTime(player_id)
		// 超时就删除
		if GetRemainSeconds(ask_time, global_config.GuildAskDonateExistSeconds) <= 1 {
			guild.AskDonates.Remove(player_id)
			changed = true
		}
	}
	return
}

// 公会捐赠刷新
func (this *Player) guild_check_donate_refresh() bool {
	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		return false
	}
	last_refresh_time := guild.GetLastDonateRefreshTime()
	if !utils.CheckDayTimeArrival(last_refresh_time, global_config.GuildDonateRefreshTime) {
		return false
	}
	this.db.Guild.SetDonateNum(0)
	this.send_guild_data()
	return true
}

// 公会捐赠列表
func (this *Player) guild_donate_list() int32 {
	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		log.Error("Player[%v] cant get guild or guild not exist", this.Id)
		return -1
	}

	guild_check_donate_list(guild)

	this.send_guild_donate_list(guild)

	return 1
}

// 公会请求捐赠
func (this *Player) guild_ask_donate(item_id int32) int32 {
	item := guild_donate_table_mgr.Get(item_id)
	if item == nil {
		log.Error("Guild Donate item table not found %v", item_id)
		return -1
	}

	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		log.Error("Player[%v] cant get guild or guild not exist", this.Id)
		return -1
	}

	guild_check_donate_list(guild)
	this.guild_check_donate_refresh()

	if guild.AskDonates.HasIndex(this.Id) {
		log.Error("Player[%v] already asked donate", this.Id)
		return -1
	}

	guild.AskDonates.Add(&dbGuildAskDonateData{
		PlayerId: this.Id,
		ItemId:   item_id,
		ItemNum:  item.RequestNum,
		AskTime:  int32(time.Now().Unix()),
	})

	response := &msg_client_message.S2CGuildAskDonateResponse{
		ItemId:  item_id,
		ItemNum: item.RequestNum,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_ASK_DONATE_RESPONSE), response)

	log.Debug("Player[%v] asked donate %v", this.Id, response)

	return 1
}

// 公会捐赠
func (this *Player) guild_donate(player_id int32) int32 {
	if this.Id == player_id {
		return -1
	}

	player := player_mgr.GetPlayerById(player_id)
	if player == nil {
		log.Error("Player[%v] not exist", player_id)
		return int32(msg_client_message.E_ERR_PLAYER_NOT_EXIST)
	}

	guild := guild_manager._get_guild(this.Id, false)
	if guild == nil {
		log.Error("Player[%v] cant get guild or guild not exist", this.Id)
		return -1
	}

	guild_check_donate_list(guild)

	if !guild.AskDonates.HasIndex(player_id) {
		log.Error("Player[%v] no ask donate, player[%v] cant donate", player_id, this.Id)
		return -1
	}

	item_id, _ := guild.AskDonates.GetItemId(player_id)
	item := guild_donate_table_mgr.Get(item_id)
	if item == nil {
		log.Error("Guild Donate item table not found %v", item_id)
		return -1
	}

	// 捐献次数（分数）
	donate_num := this.db.Guild.GetDonateNum()
	if donate_num+item.LimitScore > global_config.GuildDonateLimitDay {
		log.Error("Player[%v] left donate score %v not enough to donate", this.Id, global_config.GuildDonateLimitDay-donate_num)
		return -1
	}

	var donate_over bool
	item_num, _ := guild.AskDonates.GetItemNum(player_id)
	if item_num+1 >= item.RequestNum {
		player.add_resource(item_id, 1)
		guild.AskDonates.Remove(player_id)
		donate_over = true
	} else {
		guild.AskDonates.SetItemNum(player_id, item_num+1)
	}

	// 已捐赠的分数
	this.db.Guild.SetDonateNum(donate_num + item.LimitScore)
	this.db.Guild.SetLastAskDonateTime(int32(time.Now().Unix()))

	// 奖励
	if item.DonateRewardItem != nil {
		this.add_resources(item.DonateRewardItem)
	}

	response := &msg_client_message.S2CGuildDonateResponse{
		PlayerId:  player_id,
		ItemId:    item_id,
		ItemNum:   item_num + 1,
		DonateNum: donate_num + item.LimitScore,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_DONATE_RESPONSE), response)

	// 通知被捐赠者
	notify := &msg_client_message.S2CGuildDonateItemNotify{
		DonatePlayerId: this.Id,
		DonateNum:      response.GetDonateNum(),
		ItemId:         item_id,
		ItemNum:        response.GetItemNum(),
		DonateOver:     donate_over,
	}
	player.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_DONATE_ITEM_NOTIFY), notify)

	log.Debug("Player[%v] donate to player[%v] result %v", this.Id, player_id, response)

	return 1
}

func C2SGuildDataHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildDataRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%s)", err.Error())
		return -1
	}
	return p.send_guild_data()
}

func C2SGuildRecommendHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildRecommendRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_recommend()
}

func C2SGuildSearchHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildSearchRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_search(req.GetKey())
}

func C2SGuildCreateHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildCreateRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}

	return p.guild_create(req.GetGuildName(), req.GetGuildLogo())
}

func C2SGuildDismissHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildDismissRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}

	return p.guild_dismiss()
}

func C2SGuildCancelDismissHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildCancelDismissRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_cancel_dismiss()
}

func C2SGuildInfoModifyHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildInfoModifyRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}

	return p.guild_info_modify(req.GetNewGuildName(), req.GetNewGuildLogo())
}

func C2SGuildMembersHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildMembersRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}

	return p.guild_members_list()
}

func C2SGuildAskJoinHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildAskJoinRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}

	return p.guild_ask_join(req.GetGuildId())
}

func C2SGuildAgreeJoinHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildAgreeJoinRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_agree_join(req.GetPlayerId())
}

func C2SGuildAskListHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildAskListRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_ask_list()
}

func C2SGuildQuitHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildQuitRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_quit()
}

func C2SGuildLogsHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildLogsRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_logs()
}

func C2SGuildSignInHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildSignInRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_sign_in()
}

func C2SGuildSetOfficerHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildSetOfficerRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_set_officer(req.GetPlayerIds(), req.GetSetType())
}

func C2SGuildKickMemberHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildKickMemberRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_kick_member(req.GetPlayerIds())
}

func C2SGuildChangePresidentHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildChangePresidentRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_change_president(req.GetNewPresidentId())
}

func C2SGuildRecruitHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildRecruitRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_recruit(req.GetContent())
}

func C2SGuildDonateListHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildDonateListRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_donate_list()
}

func C2SGuildAskDonateHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildAskDonateRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_ask_donate(req.GetItemId())
}

func C2SGuildDonateHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	var req msg_client_message.C2SGuildDonateRequest
	err := proto.Unmarshal(msg_data, &req)
	if err != nil {
		log.Error("Unmarshal msg failed, err(%v)", err.Error())
		return -1
	}
	return p.guild_donate(req.GetPlayerId())
}
