package main

import (
	"libs/log"
	_ "libs/utils"
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
	if guild.GetExistType() > GUILD_EXIST_TYPE_WILL_DELETE {
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
		Info: &msg_client_message.GuildInfo{
			Id:   guild.GetId(),
			Name: guild.GetName(),
		},
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_GUILD_DATA_RESPONSE), response)
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

	return 1
}

func (this *Player) guild_search(key string) int32 {
	if this.db.Guild.GetId() > 0 {
		log.Error("Player[%v] already joined one guild, cant search", this.Id)
		return -1
	}

	guild_ids := guild_manager.Search(key)
	if guild_ids != nil {

	}
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
	return 1
}
