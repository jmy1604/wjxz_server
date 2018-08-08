package main

import (
	"libs/log"
	_ "libs/utils"
	_ "main/table_config"
	"math/rand"
	"net/http"
	"public_message/gen_go/client_message"
	_ "public_message/gen_go/client_message_id"
	"strconv"
	"sync"
	"time"

	_ "github.com/golang/protobuf/proto"
)

const (
	GUILD_MAX_NUM       = 10000
	GUILD_RECOMMEND_NUM = 5
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
	guild = this.guilds.GetRow(guild_id)
	if guild == nil {
		return
	}
	if guild.GetDeleted() {
		return nil
	}
	if is_president && guild.GetPresident() != player_id {
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

func C2SGuildRecommendHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	return 1
}

func C2SGuildSearchHandler(w http.ResponseWriter, r *http.Request, p *Player, msg_data []byte) int32 {
	return 1
}
