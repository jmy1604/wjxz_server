package main

import (
	"libs/log"
	"main/table_config"
	_ "math"
	"math/rand"
	_ "math/rand"
	"public_message/gen_go/client_message"
	"public_message/gen_go/client_message_id"
	"time"

	_ "github.com/golang/protobuf/proto"
)

// 下一关
func get_next_campaign_id(campaign_id int32) int32 {
	campaign := campaign_table_mgr.Get(campaign_id)
	if campaign == nil {
		return 0
	}
	return campaign.UnlockMap
}

// 获得关卡章节和难度
func get_campaign_chapter_and_difficulty(campaign_id int32) (int32, int32) {
	campaign := campaign_table_mgr.Get(campaign_id)
	if campaign == nil {
		return 0, 0
	}
	return campaign.ChapterMap, campaign.Difficulty
}

// 获取stage_id
func get_stage_by_campaign(campaign_id int32) *table_config.XmlPassItem {
	campaign := campaign_table_mgr.Get(campaign_id)
	if campaign == nil {
		return nil
	}
	return stage_table_mgr.Get(campaign.StageId)
}

// 是否解锁下一章节
func (this *Player) is_unlock_next_chapter(curr_campaign_id int32) (bool, int32) {
	campaign := campaign_table_mgr.Get(curr_campaign_id)
	if campaign == nil {
		return false, 0
	}
	campaigns := campaign_table_mgr.GetChapterCampaign(campaign.ChapterMap)
	if campaigns == nil || len(campaigns) == 0 {
		return false, 0
	}

	for i := 0; i < len(campaigns); i++ {
		if !this.db.Campaigns.HasIndex(campaigns[i]) {
			return false, 0
		}
	}

	if curr_campaign_id != campaigns[len(campaigns)-1] {
		return false, 0
	}
	next_campaign := campaign_table_mgr.Get(campaign.UnlockMap)
	if next_campaign == nil {
		return false, 0
	}

	return true, next_campaign.ChapterMap
}

// 是否解锁下一难度
func (this *Player) is_unlock_next_difficulty(curr_campaign_id int32) (bool, int32) {
	campaign := campaign_table_mgr.Get(curr_campaign_id)
	if campaign == nil {
		return false, 0
	}

	campaign_ids := campaign_table_mgr.GetDifficultyCampaign(campaign.Difficulty)
	if campaign_ids == nil || len(campaign_ids) == 0 {
		return false, 0
	}

	for i := 0; i < len(campaign_ids); i++ {
		if !this.db.Campaigns.HasIndex(campaign_ids[i]) {
			return false, 0
		}
	}

	if curr_campaign_id != campaign_ids[len(campaign_ids)-1] {
		return false, 0
	}

	next_campaign := campaign_table_mgr.Get(campaign.UnlockMap)
	if next_campaign == nil {
		return false, 0
	}

	return true, next_campaign.Difficulty
}

func (this *Player) FightInStage(stage *table_config.XmlPassItem) (is_win bool, my_team, target_team []*msg_client_message.BattleMemberItem, enter_reports []*msg_client_message.BattleReportItem, rounds []*msg_client_message.BattleRoundReports, has_next_wave bool) {
	if this.attack_team == nil {
		this.attack_team = &BattleTeam{}
	}
	if !this.attack_team.Init(this, BATTLE_ATTACK_TEAM, 0) {
		log.Error("Player[%v] init attack team failed", this.Id)
		return
	}

	if this.stage_team == nil {
		this.stage_team = &BattleTeam{}
	}

	if stage.Id != this.stage_id {
		this.stage_wave = 0
	}

	if !this.stage_team.InitWithStage(1, stage.Id, this.stage_wave) {
		log.Error("Player[%v] init stage[%v] wave[%v] team failed", this.Id, stage.Id, this.stage_wave)
		return
	}

	my_team = this.attack_team._format_members_for_msg()
	target_team = this.stage_team._format_members_for_msg()
	is_win, enter_reports, rounds = this.attack_team.Fight(this.stage_team, BATTLE_END_BY_ROUND_OVER, stage.MaxRound)

	this.stage_id = stage.Id
	this.stage_wave += 1
	if this.stage_wave >= stage.MaxWaves {
		this.stage_wave = 0
	} else {
		has_next_wave = true
	}

	return
}

func (this *Player) FightInCampaign(campaign_id int32) int32 {
	stage := get_stage_by_campaign(campaign_id)
	if stage == nil {
		log.Error("Cant found stage by campaign[%v]", campaign_id)
		return int32(msg_client_message.E_ERR_PLAYER_NOT_FOUND_CAMPAIGN_TABLE_DATA)
	}

	if this.db.Campaigns.HasIndex(campaign_id) {
		log.Error("Player[%v] already fight campaign[%v]", this.Id, campaign_id)
		return int32(msg_client_message.E_ERR_PLAYER_ALREADY_FIGHT_CAMPAIGN)
	}

	current_campaign_id := this.db.CampaignCommon.GetCurrentCampaignId()
	if current_campaign_id == 0 {
		if campaign_id != campaign_table_mgr.Array[0].Id {
			log.Error("Player[%v] fight first campaign[%v] invalid", this.Id, campaign_id)
			return -1
		}
		this.db.CampaignCommon.SetCurrentCampaignId(campaign_id)
	}

	is_win, my_team, target_team, enter_reports, rounds, has_next_wave := this.FightInStage(stage)

	if is_win && !has_next_wave {
		this.db.Campaigns.Add(&dbPlayerCampaignData{
			CampaignId: campaign_id,
		})
		this.db.CampaignCommon.SetCurrentCampaignId(campaign_id)
	}

	if enter_reports == nil {
		enter_reports = make([]*msg_client_message.BattleReportItem, 0)
	}
	if rounds == nil {
		rounds = make([]*msg_client_message.BattleRoundReports, 0)
	}
	response := &msg_client_message.S2CBattleResultResponse{
		IsWin:        is_win,
		EnterReports: enter_reports,
		Rounds:       rounds,
		MyTeam:       my_team,
		TargetTeam:   target_team,
		HasNextWave:  has_next_wave,
	}
	this.Send(uint16(msg_client_message_id.MSGID_S2C_BATTLE_RESULT_RESPONSE), response)

	if is_win && !has_next_wave {
		next_campaign_id := get_next_campaign_id(current_campaign_id)
		rewards_msg := &msg_client_message.S2CBattleInCampaignRewardNotify{
			NextCampaignId: next_campaign_id,
		}
		// 奖励
		for i := 0; i < len(stage.RewardList)/2; i++ {
			item_id := stage.RewardList[2*i]
			item_num := stage.RewardList[2*i+1]
			this.add_item(item_id, item_num)
			rewards_msg.Rewards = append(rewards_msg.Rewards, &msg_client_message.ItemInfo{
				ItemCfgId: item_id,
				ItemNum:   item_num,
			})
		}
		this.Send(uint16(msg_client_message_id.MSGID_S2C_BATTLE_IN_CAMPAIGN_REWARD), rewards_msg)
	}

	Output_S2CBattleResult(this, response)
	return 1
}

// 设置挂机战役关卡
func (this *Player) set_hangup_campaign_id(campaign_id int32) bool {
	hangup_id := this.db.CampaignCommon.GetHangupCampaignId()
	if hangup_id == 0 {
		if campaign_id != campaign_table_mgr.Array[0].Id {
			return false
		}
	} else if campaign_id != hangup_id {
		if !this.db.Campaigns.HasIndex(campaign_id) {
			current_campaign_id := this.db.CampaignCommon.GetCurrentCampaignId()
			next_campaign_id := get_next_campaign_id(current_campaign_id)
			if next_campaign_id != campaign_id {
				return false
			}
		}
	}

	// 关卡完成就结算一次挂机收益
	if hangup_id != 0 {
		this.hangup_income_get()
	}

	// 设置挂机开始时间
	now_time := int32(time.Now().Unix())
	if hangup_id == 0 {
		this.db.CampaignCommon.SetHangupLastDropStaticIncomeTime(now_time)
		this.db.CampaignCommon.SetHangupLastDropRandomIncomeTime(now_time)
	}
	this.db.CampaignCommon.SetHangupCampaignId(campaign_id)

	return true
}

func (this *Player) get_campaign_static_income(campaign *table_config.XmlCampaignItem, now_time, static_income_time int32) (incomes []*msg_client_message.ItemInfo, correct_secs int32) {
	st := now_time - static_income_time
	correct_secs = (st % campaign.StaticRewardSec)

	// 固定掉落
	n := st / campaign.StaticRewardSec
	for i := 0; i < len(campaign.StaticRewardItem)/2; i++ {
		item_id := campaign.StaticRewardItem[2*i]
		item_num := n * campaign.StaticRewardItem[2*i+1]
		if this.add_item(item_id, item_num) {
			incomes = append(incomes, &msg_client_message.ItemInfo{
				ItemCfgId: item_id,
				ItemNum:   item_num,
			})
		}
	}

	return
}

func (this *Player) get_campaign_random_income(campaign *table_config.XmlCampaignItem, now_time, random_income_time int32) (incomes []*msg_client_message.ItemInfo, correct_secs int32) {
	rt := now_time - random_income_time
	correct_secs = rt % campaign.RandomDropSec
	// 随机掉落
	rand.Seed(time.Now().Unix())
	out_items := make(map[int32]int32)
	n := rt / campaign.RandomDropSec
	for k := 0; k < int(n); k++ {
		for i := 0; i < len(campaign.RandomDropIDList)/2; i++ {
			group_id := campaign.RandomDropIDList[2*i]
			count := campaign.RandomDropIDList[2*i+1]
			for j := 0; j < int(count); j++ {
				if o, _ := this.drop_item_by_id(group_id, false, out_items); !o {
					continue
				}
			}
		}
	}
	for k, v := range out_items {
		incomes = append(incomes, &msg_client_message.ItemInfo{
			ItemCfgId: k,
			ItemNum:   v,
		})
	}
	return
}

// 关卡挂机收益
func (this *Player) hangup_income_get() {
	hangup_id := this.db.CampaignCommon.GetHangupCampaignId()
	if hangup_id == 0 {
		return
	}

	campaign := campaign_table_mgr.Get(hangup_id)
	if campaign == nil {
		return
	}

	now_time := int32(time.Now().Unix())
	last_logout := this.db.Info.GetLastLogout()
	static_income_time := this.db.CampaignCommon.GetHangupLastDropStaticIncomeTime()
	random_income_time := this.db.CampaignCommon.GetHangupLastDropRandomIncomeTime()

	var static_incomes, random_incomes []*msg_client_message.ItemInfo
	var cs, cr int32
	if last_logout == 0 {
		// 还未下线过
		static_incomes, cs = this.get_campaign_static_income(campaign, now_time, static_income_time)
		random_incomes, cr = this.get_campaign_random_income(campaign, now_time, random_income_time)
	} else {
		if last_logout >= static_income_time {
			if now_time-last_logout >= 8*3600 {
				static_incomes, cs = this.get_campaign_static_income(campaign, last_logout+8*3600, static_income_time)
			} else {
				static_incomes, cs = this.get_campaign_static_income(campaign, now_time, static_income_time)
			}
		} else {
			static_incomes, cs = this.get_campaign_static_income(campaign, now_time, static_income_time)
		}
		if last_logout >= random_income_time {
			if now_time-last_logout >= 8*3600 {
				random_incomes, cr = this.get_campaign_random_income(campaign, last_logout+8*3600, random_income_time)
			} else {
				random_incomes, cr = this.get_campaign_random_income(campaign, now_time, random_income_time)
			}
		} else {
			random_incomes, cr = this.get_campaign_random_income(campaign, now_time, random_income_time)
		}
	}

	this.db.CampaignCommon.SetHangupLastDropStaticIncomeTime(now_time - cs)
	this.db.CampaignCommon.SetHangupLastDropRandomIncomeTime(now_time - cr)

	var incomes []*msg_client_message.ItemInfo
	incomes = append(static_incomes, random_incomes...)
	if incomes != nil {
		var msg msg_client_message.S2CBattleCampaignHangupDropItems
		msg.Items = incomes
		this.Send(uint16(msg_client_message_id.MSGID_S2C_BATTLE_CAMPAIGN_HANGUP_DROP_ITEMS), &msg)
		log.Debug("Player[%v] hangup incomes: %v", this.Id, incomes)
	}
}
