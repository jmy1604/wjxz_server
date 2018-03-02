package table_config

import (
	"encoding/json"
	"io/ioutil"
	"libs/log"
)

type GlobalId struct {
	CancelMakingFormulaReturnMaterial_19       int32 // 取消打造装饰物返还a%材料（百分比）
	GiveFriendPointsOnce_30                    int32 // 单次赠送/收取友情点数
	GiveFriendPointsPlayersCount_31            int32 // 每次赠送友情点好友上限
	GiveFriendPointsRefreshHours_32            int32 // 赠送好友友情点上限刷新时间（小时）
	GetFriendPointsOpenFriendWoodBox_33        int32 // 好友基地打开木质宝箱获得友情点
	GetFriendPointsOpenFriendSilverBox_34      int32 // 好友基地打开银质宝箱获得友情点
	GetFriendPointsOpenFriendGoldBox_35        int32 // 好友基地打开金质宝箱获得友情点
	FriendsMaxCount_38                         int32 // 好友数量上限
	FriendFosterLimit_41                       int32 // 好友寄养上限
	FriendFosterHours_27                       int32 // 好友寄养时间（小时）
	SpiritGrowPointNeedMinute_44               int32 // 体力自动恢复时间（分钟）
	FormulaAddNewSlotCostDiamond_51            int32 // 作坊增加空位消耗钻石
	FormulaSpeedupMakingBuildingCostDiamond_52 int32 // 作坊加速打造建筑消耗钻石 t秒/钻石
	CropSpeedupCostDiamond_53                  int32 // 农作物加速升级钻石价格   t秒/钻石
	CatHouseSpeedupLevelCostDiamond_18         int32 // 猫舍加速升级钻石价格：t秒/钻石
	WorldChannelChatCooldown_40                int32 // 世界频道发送冷却时间：秒
	ChangeNameCostDiamond_58                   int32 // 改名消耗钻石
	ChangeNameFreeNum_59                       int32 // 免费改名次数
}

func (this *GlobalId) Load(file string) bool {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		log.Error("读取配置文件失败 %v", err.Error())
		return false
	}
	err = json.Unmarshal(data, this)
	if err != nil {
		log.Error("解析配置文件失败 %v", err.Error())
		return false
	}
	log.Info("载入全局配置[%v]成功", file)
	return true
}
