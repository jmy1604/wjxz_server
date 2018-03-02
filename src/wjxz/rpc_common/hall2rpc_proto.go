package rpc_common

// 转发消息
type H2R_Transfer struct {
	Method          string
	Args            interface{}
	ReceivePlayerId int32
}
type H2R_TransferResult struct {
	Result interface{}
}

// ping RPC服务
type H2R_Ping struct {
}

type H2R_Pong struct {
}

// 大厅通知RPC监听端口
type H2R_ListenIPNoitfy struct {
	ListenIP string
	ServerId int32
}
type H2R_ListenIPResult struct {
}

// 增加昵称和ID，用于查找玩家
type H2R_AddIdNick struct {
	Nick string
	Id   int32
}
type H2R_AddIdNickResult struct {
}

// 修改昵称
type H2R_RenameNick struct {
	FromPlayerId int32
	OldNick      string
	NewNick      string
}
type H2R_RenameNickResult struct {
	Error int32
}

// 昵称查找好友
type H2R_SearchFriendByNick struct {
	FromPlayerId int32
	Nick         string
}

// ID查找好友
type H2R_SearchFriendById struct {
	FromPlayerId int32
	Id           int32
}

// 关键字查找好友
type H2R_SearchFriendByKey struct {
	FromPlayerId int32
	Key          string
}

// 玩家搜索好友数据
type H2R_SearchPlayerInfo struct {
	Id        int32
	Nick      string
	Level     int32
	VipLevel  int32
	Head      string
	LastLogin int32
}

// 搜索好友结果
type H2R_SearchFriendResult struct {
	Players []*H2R_SearchPlayerInfo
}

// 好友申请
type H2R_AddFriendById struct {
	PlayerId    int32
	PlayerName  string
	AddPlayerId int32
}
type H2R_AddFriendByName struct {
	PlayerId      int32
	PlayerName    string
	AddPlayerName string
}
type H2R_AddFriendResult struct {
	PlayerId    int32
	AddPlayerId int32
	Error       int32
}

// 同意或拒绝好友申请
type H2R_AgreeAddFriend struct {
	IsAgree       bool
	PlayerId      int32
	PlayerName    string
	AgreePlayerId int32
}
type H2R_AgreeAddFriendResult struct {
	IsAgree              bool
	PlayerId             int32
	AgreePlayerId        int32
	AgreePlayerName      string
	AgreePlayerLevel     int32
	AgreePlayerVipLevel  int32
	AgreePlayerHead      string
	AgreePlayerLastLogin int32
}

// 删除好友
type H2R_RemoveFriend struct {
	PlayerId       int32
	RemovePlayerId int32
}
type H2R_RemoveFriendResult struct {
}

// 获取好友数据
type H2R_GetFriendInfo struct {
	PlayerId int32
}
type H2R_GetFriendInfoResult struct {
	PlayerId   int32
	PlayerName string
	Head       string
	Level      int32
	VipLevel   int32
	LastLogin  int32
}

// 限定商品数量
type H2R_ShopLimitedItem struct {
	ItemId int32 // 商品ID
}
type H2R_ShopLimitedItemResult struct {
	ItemId   int32 // 商品ID
	Num      int32 // 数量
	SaveTime int32 // 保存时间
	ErrCode  int32 // 1 无此商品  2 数量不足
}

// 全服限定商品购买
type H2R_BuyLimitedShopItem struct {
	ItemId int32 // 商品ID
	Num    int32 // 商品数量
}
type H2R_BuyLimitedShopItemResult struct {
	ItemId  int32 // 商品ID
	Num     int32 // 买到的数量
	LeftNum int32 // 剩余数量
	ErrCode int32 // 错误码  1 无此商品  2 商品数量不足
}

// 刷新全服限时商品
type H2R_RefreshLimitedShopItem struct {
}
type H2R_RefreshLimitedShopItemResult struct {
}

// 刷新部分全服限时商品
type H2R_RefreshSomeShopLimitedItem struct {
	ItemId []int32
}
type H2R_RefreshSomeShopLimitedItemResult struct {
}

// 检测刷新商店
type H2R_CheckRefreshShop struct {
	Days int32
}
type H2R_CheckRefreshShopResult struct {
	Result int32 // -1 错误  1 成功  0 没有刷新
}

// 自己的数据结算并获取排名同时获取好友关卡数据排名
type H2R_FriendsStagePassDataRequest struct {
	PlayerId  int32
	StageId   int32
	FriendIds []int32
}
type H2R_PlayerStageInfo struct {
	PlayerId int32
	Level    int32
	Name     string
	TopScore int32
	Head     string
}
type H2R_FriendsStagePassDataResult struct {
	StageInfos []*H2R_PlayerStageInfo // 不需要在rpc服务器上排名
}

type H2R_RankNode interface {
	New() H2R_RankNode
}

// 更新玩家关卡最高总分
type H2R_RankStageTotalScore struct {
	PlayerId    int32
	PlayerLevel int32
	TotalScore  int32
}

func (this *H2R_RankStageTotalScore) New() H2R_RankNode {
	return &H2R_RankStageTotalScore{}
}

type H2R_RankStageTotalScoreResult struct {
}

// 获取关卡最高总分排行范围
type H2R_RanklistGetStageTotalScore struct {
	PlayerId  int32
	RankStart int32
	RankNum   int32
}
type H2R_RanklistGetStageTotalScoreResult struct {
	RankItems      []*H2R_RankStageTotalScore
	SelfRank       int32
	SelfTotalScore int32
}

// 同步多个玩家关卡最高总分
type H2R_RankStageTotalScoreList struct {
	Items []*H2R_RankStageTotalScore
}
type H2R_RankStageTotalScoreListResult struct {
}

// 更新玩家关卡最高积分
type H2R_RankStageScore struct {
	PlayerId    int32
	PlayerLevel int32
	StageId     int32
	StageScore  int32
}

func (this *H2R_RankStageScore) New() H2R_RankNode {
	return &H2R_RankStageScore{}
}

type H2R_RankStageScoreResult struct {
}

// 获取玩家关卡最高积分排行
type H2R_RanklistGetStageScore struct {
	PlayerId  int32
	StageId   int32
	RankStart int32
	RankNum   int32
}
type H2R_RanklistGetStageScoreResult struct {
	StageId   int32
	RankItems []*H2R_RankStageScore
	SelfRank  int32
	SelfScore int32
}

// 同步多个玩家关卡最高积分
type H2R_RankStageScoreList struct {
	Items []*H2R_RankStageScore
}
type H2R_RankStageScoreListResult struct {
}

// 更新玩家魅力值
type H2R_RankCharm struct {
	PlayerId    int32
	PlayerLevel int32
	Charm       int32
}

func (this *H2R_RankCharm) New() H2R_RankNode {
	return &H2R_RankCharm{}
}

type H2R_RankCharmResult struct {
}

// 获取玩家魅力值排行
type H2R_RanklistGetCharm struct {
	PlayerId  int32
	RankStart int32
	RankNum   int32
}
type H2R_RanklistGetCharmResult struct {
	RankItems []*H2R_RankCharm
	SelfRank  int32
	SelfCharm int32
}

// 同步多个玩家魅力值
type H2R_RankCharmList struct {
	Items []*H2R_RankCharm
}
type H2R_RankCharmListResult struct {
}

// 更新玩家欧气值
type H2R_RankCatOuqi struct {
	PlayerId    int32
	PlayerLevel int32
	CatId       int32
	CatTableId  int32
	CatNick     string
	CatLevel    int32
	CatStar     int32
	CatOuqi     int32
	IsMaxOuqi   bool
}

func (this *H2R_RankCatOuqi) New() H2R_RankNode {
	return &H2R_RankCatOuqi{}
}

type H2R_RankCatOuqiResult struct {
}

// 获取玩家欧气值排行
type H2R_RanklistGetCatOuqi struct {
	PlayerId     int32
	MaxOuqiCatId int32
	RankStart    int32
	RankNum      int32
}
type H2R_RanklistGetCatOuqiResult struct {
	RankItems   []*H2R_RankCatOuqi
	SelfRank    int32
	SelfCatId   int32
	SelfCatOuqi int32
}

// 更新玩家被赞
type H2R_RankZaned struct {
	PlayerId    int32
	PlayerLevel int32
	Zaned       int32
}

func (this *H2R_RankZaned) New() H2R_RankNode {
	return &H2R_RankZaned{}
}

type H2R_RankZanedResult struct {
}

// 获取玩家被赞数排行
type H2R_RanklistGetZaned struct {
	PlayerId  int32
	RankStart int32
	RankNum   int32
}
type H2R_RanklistGetZanedResult struct {
	RankItems []*H2R_RankZaned
	SelfRank  int32
	SendZaned int32
}

// 删除玩家排名
type H2R_RankDelete struct {
	PlayerId int32
	RankType int32
	Param    int32
}
type H2R_RankDeleteResult struct {
	PlayerId int32
	RankType int32
	Param    int32
}

// ----------------------------------------------------------------------------
// 创建个人空间
type H2R_PersonalSpaceCreate struct {
	FromPlayerId int32
}
type H2R_PersonalSpaceCreateResult struct {
	Error int32
}

// 个人空间获取
type H2R_GetPersonalSpace struct {
	FromPlayerId int32
	ToPlayerId   int32
}

// 个人空间照片
type H2R_PSPictureData struct {
	Id          int32  // 照片ID
	Url         string // 照片地址
	UploadTime  int32  // 上传时间
	Zaned       int32  // 被赞数
	LeaveMsgNum int32  // 留言数
}

// 个人空间返回结果
type H2R_GetPersonalSpaceResult struct {
	FromPlayerId   int32
	ToPlayerId     int32
	Signature      string
	Pictures       []*H2R_PSPictureData
	LeaveMsgs      []*H2R_PSLeaveMessageData
	LeaveMsgIsMore bool
	Error          int32
}

// 个人空间点赞
type H2R_PSZan struct {
	FromPlayerId int32
	ToPlayerId   int32
	PictureId    int32
}

type H2R_PSZanResult struct {
	FromPlayerId int32
	ToPlayerId   int32
	PictureId    int32
	Zaned        int32
	Error        int32
}

// 个人空间留言评论
type H2R_PSLeaveMessageCommentData struct {
	Id           int32
	Content      []byte
	SendTime     int32
	SendPlayerId int32
}

// 个人空间留言
type H2R_PSLeaveMessageData struct {
	Id           int32
	SendPlayerId int32
	Content      []byte
	SendTime     int32
	Comments     []*H2R_PSLeaveMessageCommentData
}

// 拉取留言
type H2R_PSPullLeaveMessages struct {
	FromPlayerId int32
	ToPlayerId   int32
	PictureId    int32
	StartIndex   int32
	MessageNum   int32
}
type H2R_PSPullLeaveMessagesResult struct {
	FromPlayerId int32
	ToPlayerId   int32
	PictureId    int32
	LeaveMsgs    []*H2R_PSLeaveMessageData
	IsMoreMsg    bool
	Error        int32
}

// 发送留言
type H2R_PSSendLeaveMsg struct {
	FromPlayerId int32
	ToPlayerId   int32
	PictureId    int32
	Content      []byte
}
type H2R_PSSendLeaveMsgResult struct {
	MsgId int32
	Error int32
}

// 删除留言
type H2R_PSDeleteLeaveMsg struct {
	FromPlayerId   int32
	ToPlayerId     int32
	PictureId      int32 // 0表示空间留言
	LeaveMessageId int32
}
type H2R_PSDeleteLeaveMsgResult struct {
	Error int32
}

// 发送评论
type H2R_PSSendLeaveMsgComment struct {
	FromPlayerId  int32
	ToPlayerId    int32
	PictureId     int32 // 0表示空间评论
	LeaveMsgId    int32
	ReplyPlayerId int32
	Content       []byte
}
type H2R_PSSendLeaveMsgCommentResult struct {
	CommentId int32
	Error     int32
}

// 删除评论
type H2R_PSDeleteLeaveMsgComment struct {
	FromPlayerId   int32
	ToPlayerId     int32
	PictureId      int32 // 0表示空间留言
	LeaveMessageId int32
	CommentId      int32
}
type H2R_PSDeleteLeaveMsgCommentResult struct {
	Error int32
}

// 拉取评论
type H2R_PSPullLeaveMessageComments struct {
	FromPlayerId int32
	ToPlayerId   int32
	PictureId    int32
	MessageId    int32
	StartIndex   int32
	CommentNum   int32
}
type H2R_PSPullLeaveMessageCommentsResult struct {
	FromPlayerId int32
	ToPlayerId   int32
	PictureId    int32
	MessageId    int32
	Comments     []*H2R_PSLeaveMessageCommentData
	Error        int32
}

// 删除照片
type H2R_PSDeletePicture struct {
	FromPlayerId int32
	PictureId    int32
}
type H2R_PSDeletePictureResult struct {
	Error int32
}

// 修改空间签名
type H2R_PSModifySignature struct {
	FromPlayerId int32
	Signature    string
}
type H2R_PSModifySignatureResult struct {
	Error int32
}
