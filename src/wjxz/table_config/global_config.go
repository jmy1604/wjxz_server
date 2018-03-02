package table_config

import (
	"encoding/json"
	"io/ioutil"
	"libs/log"
	"libs/utils"
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
	InitItems     []CfgIdNum
	InitItem_len  int32
	InitCats      []CfgIdNum
	InitCats_len  int32
	InitDiamond   int32
	InitCoin      int32
	InitAreas     []int32
	InitAreas_len int32
	InitFormulas  []int32
	InitBuildings []int32

	MaxFriendNum int32

	DaySignSumRewards []DaySignSumReward // 累计签到奖励

	WoodChestUnlockTime     int32
	SilverChestUnlockTime   int32
	GoldenChestUnlockTime   int32
	GiantChestUnlockTime    int32
	MagicChestUnlockTime    int32
	RareChestUnlockTime     int32
	EpicChestUnlockTime     int32
	LegendryChestUnlockTime int32

	GooglePayUrl       string
	FaceBookPayUrl     string
	ApplePayUrl        string
	ApplePaySandBoxUrl string

	MaxNameLen     int32   // 最大名字长度
	ChgNameCost    []int32 // 改名消耗的钻石
	ChgNameCostLen int32   // 消耗数组的长度

	FirstPayReward int32

	ShopStartRefreshTime string
	ShopRefreshTime      string

	ExpeditionTaskCount       int32 // 日常任务的数目
	ExpeditionSPEventSec      int32 // 特殊日常任务间隔
	ExpeditionDayFreeChgCount int32 // 每日免费刷新任务次数
	ExpeditionDayChgAddCost   int32 // 免费结束之后每次增加的钻石
	ExpeditionDayMaxChgCost   int32 // 每日最大刷新花费
	ExpeditionDayStartCount   int32 // 每日探险次数

	MapBlockRefleshSec int32 // 地图障碍刷新时间间隔
	MapChestRefleshSec int32 // 地图宝箱刷新时间间隔

	NormalMailLastSec  int32 // 普通邮件持续秒数
	ReqHelpMailLastSec int32 // 帮助邮件持续秒数
	MaxMailCount       int32 // 最大邮件数目

	ChapterUnlockNeedFriendNum int32 // 解锁章节需要好友同意的数目
	ChapterUnlockSecPerDiamond int32 // 解锁章节每钻石对应的描述
	MaxHelpUnlockNum           int32 // 每天最大帮助别人的次数

	ExpeditionMultiVal []int32 // 参加任务的猫和奖励倍数对应的百分比

	RankingListOnceGetItemsNum int32 // 排行榜一次能取的最大数量

	WorldChatMaxMsgNum       int32 // 世界聊天最大消息数
	WorldChatPullMaxMsgNum   int32 // 世界聊天拉取的消息数量最大值
	WorldChatPullMsgCooldown int32 // 世界聊天拉取CD
	WorldChatMsgMaxBytes     int32 // 世界聊天消息最大长度
	WorldChatMsgExistTime    int32 // 世界聊天消息存在时间

	AnouncementMaxNum       int32 // 公告最大数量
	AnouncementSendCooldown int32 // 公告发送间隔冷却时间(分钟)
	AnouncementSendMaxNum   int32 // 公告一次发送最大数量
	AnouncementExistTime    int32 // 公告存在时间

	FriendGivePointsRefreshTime     TimeData // 赠送友情点刷新时间
	FriendGivePointsPlayerNumOneDay int32    // 好友赠送点数每天最大人数

	MaxDayBuyTiLiCount int32 // 每天最大购买体力的次数
	DayBuyTiliAdd      int32 // 每次购买体力的体力增加值
	DayBuyTiLiCost     int32 // 每次购买体力的消耗的钻石数目
}

type GlobalConfigManager struct {
	global_config     *GlobalConfig
	shop_time_checker *utils.DaysTimeChecker
}

func (this *GlobalConfigManager) Init(conf_path string) bool {
	gc := &GlobalConfig{}
	data, err := ioutil.ReadFile(conf_path)
	if nil != err {
		log.Error("GlobalConfigManager::Init failed to readfile err(%s)!", err.Error())
		return false
	}

	err = json.Unmarshal(data, gc)
	if nil != err {
		log.Error("GlobalConfigManager::Init json unmarshal failed err(%s)!", err.Error())
		return false
	}

	gc.InitItem_len = int32(len(gc.InitItems))
	gc.InitCats_len = int32(len(gc.InitCats))
	gc.InitAreas_len = int32(len(gc.InitAreas))
	gc.ChgNameCostLen = int32(len(gc.ChgNameCost))

	this.global_config = gc

	if gc.InitBuildings != nil && len(gc.InitBuildings)%2 != 0 {
		log.Error("GlobalConfigManager::Init json data InitBuildings invalid length")
		return false
	}

	// 商店起始时间配置
	this.shop_time_checker = &utils.DaysTimeChecker{}
	if !this.shop_time_checker.Init("2006-Jan-02 15:04:05", gc.ShopStartRefreshTime) {
		return false
	}

	return true
}

func (this *GlobalConfigManager) GetGlobalConfig() *GlobalConfig {
	return this.global_config
}

func (this *GlobalConfigManager) GetShopTimeChecker() *utils.DaysTimeChecker {
	return this.shop_time_checker
}
