package table_config

import (
	"encoding/json"
	"io/ioutil"
	"libs/log"
	_ "libs/utils"
)

type WeekTime struct {
	WeekDay int32
	Hour    int32
	Minute  int32
	Second  int32
}

type CfgIdNum struct {
	CfgId int32
	Num   int32
}

type DaySignSumReward struct {
	SignNum int32
	ChestId int32
}

type TimeData struct {
	Hour   int32
	Minute int32
	Second int32
}

type GlobalConfig struct {
	InitRoles                            []int32    // 初始角色
	InitItems                            []CfgIdNum // 初始物品
	InitItem_len                         int32      // 初始物品长度
	InitDiamond                          int32      // 初始钻石
	InitCoin                             int32      // 初始金币
	InitEnergy                           int32      // 初始怒气
	MaxEnergy                            int32      // 最大怒气
	EnergyAdd                            int32      // 怒气增量
	HeartbeatInterval                    int32      // 心跳间隔
	MaxRoleCount                         int32      // 最大角色数量
	MailTitleBytes                       int32      // 邮件标题最大字节数
	MailContentBytes                     int32      // 邮件内容最大字节数
	MailMaxCount                         int32      // 邮件最大数量
	MailNormalExistDays                  int32      // 最大无附件邮件保存天数
	MailAttachExistDays                  int32      // 最大附件邮件保存天数
	MailPlayerSendCooldown               int32      // 个人邮件发送间隔(秒)
	PlayerBattleRecordMaxCount           int32      // 玩家战斗录像最大数量
	TowerKeyMax                          int32      // 爬塔钥匙最大值
	TowerKeyGetInterval                  int32      // 爬塔获取钥匙的时间间隔(秒)
	TowerKeyId                           int32      // 爬塔门票
	ItemLeftSlotOpenLevel                int32      // 左槽开启等级
	LeftSlotDropId                       int32      // 左槽掉落ID
	ArenaTicketItemId                    int32      // 竞技场门票ID
	ArenaTicketsDay                      int32      // 竞技场每天的门票
	ArenaTicketRefreshTime               string     // 竞技场门票刷新时间
	ArenaEnterLevel                      int32      // 竞技场进入等级
	ArenaGetTopRankNum                   int32      // 竞技场取最高排名数
	ArenaMatchPlayerNum                  int32      // 竞技场匹配人数
	ArenaRepeatedWinNum                  int32      // 竞技场连胜场数
	ArenaLoseRepeatedNum                 int32      // 竞技场连败场数
	ArenaHighGradeStart                  int32      // 竞技场高段位开始
	ArenaSeasonDays                      int32      // 竞技场赛季天数
	ArenaDayResetTime                    string     // 竞技场每天重置时间
	ArenaSeasonResetTime                 string     // 竞技场赛季重置时间
	ActiveStageRefreshTime               string     // 活动副本数据重置时间
	ActiveStageChallengeNumOfDay         int32      // 活动副本重置挑战次数
	ActiveStageChallengeNumPrice         int32      // 活动副本挑战次数购买价格
	FriendMaxNum                         int32      // 好友最大数量
	FriendRecommendNum                   int32      // 好友推荐数
	FriendPointItemId                    int32      // 友情点ID
	FriendPointsOnceGive                 int32      // 一次赠送的友情点
	FriendStaminaItemId                  int32      // 友情体力ID
	FriendStaminaLimit                   int32      // 友情体力上限
	FriendStartStamina                   int32      // 开始体力
	FriendStaminaResumeOnePointNeedHours int32      // 友情体力恢复一点需要时间(小时)
	FriendBossAttackCostStamina          int32      // 攻击好友BOSS消耗体力
	FriendBossAttackCooldown             int32      // 攻击好友BOSS冷却时间
	FriendRefreshTime                    string     // 好友刷新时间
	FriendSearchBossRefreshHours         int32      // 好友BOSS刷新时间间隔
	FriendAssistPointsGet                int32      // 助战好友获得友情点数
	FriendPointsGetLimitDay              int32      // 每天获得友情点上限
	DailyTaskRefreshTime                 string     // 日常任务刷新时间

	GooglePayUrl       string
	FaceBookPayUrl     string
	ApplePayUrl        string
	ApplePaySandBoxUrl string

	MaxNameLen     int32   // 最大名字长度
	ChgNameCost    []int32 // 改名消耗的钻石
	ChgNameCostLen int32   // 消耗数组的长度

	FirstPayReward int32

	WorldChatMaxMsgNum       int32 // 世界聊天最大消息数
	WorldChatPullMaxMsgNum   int32 // 世界聊天拉取的消息数量最大值
	WorldChatPullMsgCooldown int32 // 世界聊天拉取CD
	WorldChatMsgMaxBytes     int32 // 世界聊天消息最大长度
	WorldChatMsgExistTime    int32 // 世界聊天消息存在时间

	AnouncementMaxNum       int32 // 公告最大数量
	AnouncementSendCooldown int32 // 公告发送间隔冷却时间(分钟)
	AnouncementSendMaxNum   int32 // 公告一次发送最大数量
	AnouncementExistTime    int32 // 公告存在时间
}

func (this *GlobalConfig) Init(conf_path string) bool {
	data, err := ioutil.ReadFile(conf_path)
	if nil != err {
		log.Error("GlobalConfigManager::Init failed to readfile err(%s)!", err.Error())
		return false
	}

	err = json.Unmarshal(data, this)
	if nil != err {
		log.Error("GlobalConfigManager::Init json unmarshal failed err(%s)!", err.Error())
		return false
	}

	return true
}
