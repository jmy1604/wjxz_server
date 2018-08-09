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
	GUILD_MAX_NUM       = 10000
	GUILD_RECOMMEND_NUM = 5
)

const (
	GUILD_EXIST_TYPE_NONE        = iota
	GUILD_EXIST_TYPE_WILL_DELETE = 1
	GUILD_EXIST_TYPE_DELETED     = 2
)

func _player_get_guild_id(player_id int32) int32 {
	player := player_mgr.GetPlayerById(player_id)
	if player == nil {
		return int32(msg_client_message.E_ERR_PLAYER_NOT_EXIST)
	}
	guild_id := player.db.Guild.GetId()
	return guild_id
}

type GuildManager struct {
	guilds           *dbGuildTable
	guild_ids        []int32
	guild_num        int32
	guild_id_map     map[int32]int32
	guild_name_map   map[string]int32
	guild_ids_locker *sync.RWMutex
}

var guild_manager GuildManager

func (this *GuildManager) Init() {
	this.guilds = dbc.Guilds
	this.guild_ids = make([]int32, GUILD_MAX_NUM)
	this.guild_id_map = make(map[int32]int32)
	this.guild_name_map = make(map[string]int32)
	this.guild_ids_locker = &sync.RWMutex{}
	for gid, guild := range this.guilds.m_rows {
		this.guild_ids[this.guild_num] = gid
		this.guild_num += 1
		this.guild_id_map[gid] = gid
		this.guild_name_map[guild.GetName()] = gid
	}
}

func (this *GuildManager) CreateGuild(player_id int32, guild_name string, logo int32) int32 {
	player := player_mgr.GetPlayerById(player_id)
	if player == nil {
		return int32(msg_client_message.E_ERR_PLAYER_NOT_EXIST)
	}
	guild_id := player.db.Guild.GetId()
	if guild_id > 0 {
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
	row.SetLogo(logo)
	row.SetPresident(player_id)
	guild_id = row.GetId()

	player.db.Guild.SetId(guild_id)

	return guild_id
}

func (this *GuildManager) GetGuild(guild_id int32) *dbGuildRow {
	guild := this.guilds.GetRow(guild_id)
	exist_type := _guild_get_exist_type(guild)
	if exist_type == GUILD_EXIST_TYPE_DELETED {
		return nil
	}
	return guild
}

func (this *GuildManager) Recommend(player_id int32) (guild_ids []int32) {
	guild_id := _player_get_guild_id(player_id)
	if guild_id > 0 {
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
	guild_id := _player_get_guild_id(player_id)
	guild = this.GetGuild(guild_id)
	if guild == nil || (is_president && guild.GetPresident() != player_id) {
		return nil
	}
	return guild
}

func (this *GuildManager) SetAnouncement(president_id int32, anouncement string) int32 {
	guild := this._get_guild(president_id, true)
	if guild == nil {
		return -1
	}

	guild.SetAnouncement(anouncement)
	return 1
}

func (this *GuildManager) AskJoin(player_id int32, guild_id int32) int32 {
	guild := this._get_guild(player_id, false)
	if guild == nil {
		return -1
	}

	if guild.GetId() == guild_id || guild.AskLists.HasIndex(player_id) {
		log.Error("Player[%v] already joined guild[%v], no need to ask join", player_id, guild_id)
		return -1
	}

	guild.AskLists.Add(&dbGuildAskListData{
		PlayerId: player_id,
	})

	return 1
}

func (this *GuildManager) AgreeAsk(president_id int32, player_id int32) int32 {
	guild := this._get_guild(president_id, true)
	if guild == nil {
		return -1
	}

	if guild.Members.HasIndex(player_id) {
		log.Error("Player[%v] already joined guild[%v], president[%v] no need to add", player_id, guild.GetId(), president_id)
		return -1
	}

	guild.AskLists.Remove(player_id)

	guild.Members.Add(&dbGuildMemberData{
		player_id,
	})

	return 1
}

func (this *GuildManager) RemoveAsk(president_id int32, player_id int32) int32 {
	guild := this._get_guild(president_id, true)
	if guild == nil {
		return -1
	}
	if !guild.AskLists.HasIndex(player_id) {
		log.Error("Guild[%v] no player[%v] ask, president[%v] remove failed", guild.GetId(), player_id, president_id)
		return -1
	}
	guild.AskLists.Remove(player_id)
	return 1
}

func (this *GuildManager) RemovePlayer(president_id, player_id int32) int32 {
	guild := this._get_guild(president_id, true)
	if guild == nil {
		return -1
	}
	if !guild.Members.HasIndex(player_id) {
		log.Error("Guild[%v] no player[%v], president[%v] remove failed", guild.GetId(), player_id, president_id)
		return -1
	}
	guild.Members.Remove(player_id)
	return 1
}

func _format_guild_base_info_to_msg(guild *dbGuildRow) (msg *msg_client_message.GuildBaseInfo) {
	msg = &msg_client_message.GuildBaseInfo{
		Id:        guild.GetId(),
		Name:      guild.GetName(),
		Level:     guild.GetLevel(),
		Logo:      guild.GetLogo(),
		MemberNum: guild.Members.NumAll(),
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
		guild.SetExistType(GUILD_EXIST_TYPE_DELETED)
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
				p.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_DELETE_NOTIFY), notify)
				SendMail(nil, mid, MAIL_TYPE_GUILD, "Guild Dismissed", "Guild Dismissed", nil)
			}
		}
	}
	return
}

func _guild_get_exist_type(guild *dbGuildRow) int32 {
	_guild_get_dismiss_remain_seconds(guild)
	return guild.GetExistType()
}

func (this *Player) _format_guild_info_to_msg(guild *dbGuildRow) (msg *msg_client_message.GuildInfo) {
	var dismiss_remain_seconds, sign_remain_seconds, ask_donate_remain_seconds, donate_reset_remain_seconds int32
	dismiss_remain_seconds = _guild_get_dismiss_remain_seconds(guild)
	sign_remain_seconds = utils.GetRemainSeconds2NextDayTime(this.db.Guild.GetSignTime(), global_config.GuildSignRefreshTime)
	ask_donate_remain_seconds = GetRemainSeconds(this.db.Guild.GetLastAskDonateTime(), global_config.GuildAskDonateCDSecs)
	msg = &msg_client_message.GuildInfo{
		Id:                       guild.GetId(),
		Name:                     guild.GetName(),
		Level:                    guild.GetLevel(),
		Exp:                      guild.GetExp(),
		Logo:                     guild.GetLogo(),
		Anouncement:              guild.GetAnouncement(),
		DismissRemainSeconds:     dismiss_remain_seconds,
		SignRemainSeconds:        sign_remain_seconds,
		AskDonateRemainSeconds:   ask_donate_remain_seconds,
		DonateResetRemainSeconds: donate_reset_remain_seconds,
	}
	return
}

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

func (this *Player) guild_recommend() int32 {
	if this.db.Guild.GetId() > 0 {
		log.Error("Player[%v] no need to recommend guild", this.Id)
		return -1
	}

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
	this.add_diamond(-global_config.GuildCreateCostGem)

	guild := guild_manager.GetGuild(guild_id)
	guild_msg := this._format_guild_info_to_msg(guild)
	response := &msg_client_message.S2CGuildCreateResponse{
		Info: guild_msg,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_CREATE_RESPONSE), response)

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
	response := &msg_client_message.S2CGuildDismissResponse{
		RealDismissRemainSeconds: global_config.GuildDismissWaitingSeconds,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_DISMISS_RESPONSE), response)

	log.Debug("Player[%v] dismiss guild %v", this.Id, response)

	return 1
}

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
