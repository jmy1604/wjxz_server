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
	InitRoles    []int32
	InitItems    []CfgIdNum
	InitItem_len int32
	InitDiamond  int32
	InitCoin     int32

	InitEnergy int32
	MaxEnergy  int32
	EnergyAdd  int32

	MaxFriendNum int32

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
	gc.ChgNameCostLen = int32(len(gc.ChgNameCost))

	this.global_config = gc

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
