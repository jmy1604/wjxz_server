package main

import (
	_ "3p/code.google.com.protobuf/proto"
	"github.com/golang/protobuf/proto"
	_ "3p/mysql"
	"database/sql"
	"errors"
	"fmt"
	"libs/log"
	"math/rand"
	"os"
	"public_message/gen_go/db_hall"
	"sort"
	"strings"
	"sync/atomic"
	"time"
)

type dbArgs struct {
	args  []interface{}
	count int32
}

func new_db_args(count int32) (this *dbArgs) {
	this = &dbArgs{}
	this.args = make([]interface{}, count)
	this.count = 0
	return this
}
func (this *dbArgs) Push(arg interface{}) {
	this.args[this.count] = arg
	this.count++
}
func (this *dbArgs) GetArgs() (args []interface{}) {
	return this.args[0:this.count]
}
func (this *DBC) StmtPrepare(s string) (r *sql.Stmt, e error) {
	this.m_db_lock.Lock("DBC.StmtPrepare")
	defer this.m_db_lock.Unlock()
	return this.m_db.Prepare(s)
}
func (this *DBC) StmtExec(stmt *sql.Stmt, args ...interface{}) (r sql.Result, err error) {
	this.m_db_lock.Lock("DBC.StmtExec")
	defer this.m_db_lock.Unlock()
	return stmt.Exec(args...)
}
func (this *DBC) StmtQuery(stmt *sql.Stmt, args ...interface{}) (r *sql.Rows, err error) {
	this.m_db_lock.Lock("DBC.StmtQuery")
	defer this.m_db_lock.Unlock()
	return stmt.Query(args...)
}
func (this *DBC) StmtQueryRow(stmt *sql.Stmt, args ...interface{}) (r *sql.Row) {
	this.m_db_lock.Lock("DBC.StmtQueryRow")
	defer this.m_db_lock.Unlock()
	return stmt.QueryRow(args...)
}
func (this *DBC) Query(s string, args ...interface{}) (r *sql.Rows, e error) {
	this.m_db_lock.Lock("DBC.Query")
	defer this.m_db_lock.Unlock()
	return this.m_db.Query(s, args...)
}
func (this *DBC) QueryRow(s string, args ...interface{}) (r *sql.Row) {
	this.m_db_lock.Lock("DBC.QueryRow")
	defer this.m_db_lock.Unlock()
	return this.m_db.QueryRow(s, args...)
}
func (this *DBC) Exec(s string, args ...interface{}) (r sql.Result, e error) {
	this.m_db_lock.Lock("DBC.Exec")
	defer this.m_db_lock.Unlock()
	return this.m_db.Exec(s, args...)
}
func (this *DBC) Conn(name string, addr string, acc string, pwd string, db_copy_path string) (err error) {
	log.Trace("%v %v %v %v", name, addr, acc, pwd)
	this.m_db_name = name
	source := acc + ":" + pwd + "@tcp(" + addr + ")/" + name + "?charset=utf8"
	this.m_db, err = sql.Open("mysql", source)
	if err != nil {
		log.Error("open db failed %v", err)
		return
	}

	this.m_db_lock = NewMutex()
	this.m_shutdown_lock = NewMutex()

	if config.DBCST_MAX-config.DBCST_MIN <= 1 {
		return errors.New("DBCST_MAX sub DBCST_MIN should greater than 1s")
	}

	err = this.init_tables()
	if err != nil {
		log.Error("init tables failed")
		return
	}

	if os.MkdirAll(db_copy_path, os.ModePerm) == nil {
		os.Chmod(db_copy_path, os.ModePerm)
	}
	
	this.m_db_last_copy_time = int32(time.Now().Hour())
	this.m_db_copy_path = db_copy_path
	addr_list := strings.Split(addr, ":")
	this.m_db_addr = addr_list[0]
	this.m_db_account = acc
	this.m_db_password = pwd
	this.m_initialized = true

	return
}
func (this *DBC) check_files_exist() (file_name string) {
	f_name := fmt.Sprintf("%v/%v_%v", this.m_db_copy_path, this.m_db_name, time.Now().Format("20060102-15"))
	num := int32(0)
	for {
		if num == 0 {
			file_name = f_name
		} else {
			file_name = f_name + fmt.Sprintf("_%v", num)
		}
		_, err := os.Lstat(file_name)
		if err != nil {
			break
		}
		num++
	}
	return file_name
}
func (this *DBC) Loop() {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}

		log.Trace("数据库主循环退出")
		this.m_shutdown_completed = true
	}()

	for {
		t := config.DBCST_MIN + rand.Intn(config.DBCST_MAX-config.DBCST_MIN)
		if t <= 0 {
			t = 600
		}

		for i := 0; i < t; i++ {
			time.Sleep(time.Second)
			if this.m_quit {
				break
			}
		}

		if this.m_quit {
			break
		}

		begin := time.Now()
		err := this.Save(false)
		if err != nil {
			log.Error("save db failed %v", err)
		}
		log.Trace("db存数据花费时长: %v", time.Now().Sub(begin).Nanoseconds())
		
		/*
			now_time_hour := int32(time.Now().Hour())
			if now_time_hour != this.m_db_last_copy_time {
				args := []string {
					fmt.Sprintf("-h%v", this.m_db_addr),
					fmt.Sprintf("-u%v", this.m_db_account),
					fmt.Sprintf("-p%v", this.m_db_password),
					this.m_db_name,
				}
				cmd := exec.Command("mysqldump", args...)
				var out bytes.Buffer
				cmd.Stdout = &out
				cmd_err := cmd.Run()
				if cmd_err == nil {
					file_name := this.check_files_exist()
					file, file_err := os.OpenFile(file_name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0660)
					defer file.Close()
					if file_err == nil {
						_, write_err := file.Write(out.Bytes())
						if write_err == nil {
							log.Trace("数据库备份成功！备份文件名:%v", file_name)
						} else {
							log.Error("数据库备份文件写入失败！备份文件名%v", file_name)
						}
					} else {
						log.Error("数据库备份文件打开失败！备份文件名%v", file_name)
					}
					file.Close()
				} else {
					log.Error("数据库备份失败！")
				}
				this.m_db_last_copy_time = now_time_hour
			}
		*/
		
		if this.m_quit {
			break
		}
	}

	log.Trace("数据库缓存主循环退出，保存所有数据")

	err := this.Save(true)
	if err != nil {
		log.Error("shutdwon save db failed %v", err)
		return
	}

	err = this.m_db.Close()
	if err != nil {
		log.Error("close db failed %v", err)
		return
	}
}
func (this *DBC) Shutdown() {
	if !this.m_initialized {
		return
	}

	this.m_shutdown_lock.UnSafeLock("DBC.Shutdown")
	defer this.m_shutdown_lock.UnSafeUnlock()

	if this.m_quit {
		return
	}
	this.m_quit = true

	log.Trace("关闭数据库缓存")

	begin := time.Now()

	for {
		if this.m_shutdown_completed {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	log.Trace("关闭数据库缓存耗时 %v 秒", time.Now().Sub(begin).Seconds())
}


const DBC_VERSION = 1
const DBC_SUB_VERSION = 0

type dbSmallRankRecordData struct{
	Rank int32
	Id int32
	Val int32
	Name string
}
func (this* dbSmallRankRecordData)from_pb(pb *db.SmallRankRecord){
	if pb == nil {
		return
	}
	this.Rank = pb.GetRank()
	this.Id = pb.GetId()
	this.Val = pb.GetVal()
	this.Name = pb.GetName()
	return
}
func (this* dbSmallRankRecordData)to_pb()(pb *db.SmallRankRecord){
	pb = &db.SmallRankRecord{}
	pb.Rank = proto.Int32(this.Rank)
	pb.Id = proto.Int32(this.Id)
	pb.Val = proto.Int32(this.Val)
	pb.Name = proto.String(this.Name)
	return
}
func (this* dbSmallRankRecordData)clone_to(d *dbSmallRankRecordData){
	d.Rank = this.Rank
	d.Id = this.Id
	d.Val = this.Val
	d.Name = this.Name
	return
}
type dbLimitShopItemData struct{
	CommodityId int32
	LeftNum int32
}
func (this* dbLimitShopItemData)from_pb(pb *db.LimitShopItem){
	if pb == nil {
		return
	}
	this.CommodityId = pb.GetCommodityId()
	this.LeftNum = pb.GetLeftNum()
	return
}
func (this* dbLimitShopItemData)to_pb()(pb *db.LimitShopItem){
	pb = &db.LimitShopItem{}
	pb.CommodityId = proto.Int32(this.CommodityId)
	pb.LeftNum = proto.Int32(this.LeftNum)
	return
}
func (this* dbLimitShopItemData)clone_to(d *dbLimitShopItemData){
	d.CommodityId = this.CommodityId
	d.LeftNum = this.LeftNum
	return
}
type dbPlayerInfoData struct{
	Lvl int32
	Exp int32
	CreateUnix int32
	Gold int32
	Diamond int32
	LastDialyTaskUpUinx int32
	Icon string
	CustomIcon string
	LastLogout int32
	LastLogin int32
	Spirit int32
	FriendPoints int32
	LastRefreshShopTime int32
	VipLvl int32
	FriendMessageUnreadCurrId int32
}
func (this* dbPlayerInfoData)from_pb(pb *db.PlayerInfo){
	if pb == nil {
		return
	}
	this.Lvl = pb.GetLvl()
	this.Exp = pb.GetExp()
	this.CreateUnix = pb.GetCreateUnix()
	this.Gold = pb.GetGold()
	this.Diamond = pb.GetDiamond()
	this.LastDialyTaskUpUinx = pb.GetLastDialyTaskUpUinx()
	this.Icon = pb.GetIcon()
	this.CustomIcon = pb.GetCustomIcon()
	this.LastLogout = pb.GetLastLogout()
	this.LastLogin = pb.GetLastLogin()
	this.Spirit = pb.GetSpirit()
	this.FriendPoints = pb.GetFriendPoints()
	this.LastRefreshShopTime = pb.GetLastRefreshShopTime()
	this.VipLvl = pb.GetVipLvl()
	this.FriendMessageUnreadCurrId = pb.GetFriendMessageUnreadCurrId()
	return
}
func (this* dbPlayerInfoData)to_pb()(pb *db.PlayerInfo){
	pb = &db.PlayerInfo{}
	pb.Lvl = proto.Int32(this.Lvl)
	pb.Exp = proto.Int32(this.Exp)
	pb.CreateUnix = proto.Int32(this.CreateUnix)
	pb.Gold = proto.Int32(this.Gold)
	pb.Diamond = proto.Int32(this.Diamond)
	pb.LastDialyTaskUpUinx = proto.Int32(this.LastDialyTaskUpUinx)
	pb.Icon = proto.String(this.Icon)
	pb.CustomIcon = proto.String(this.CustomIcon)
	pb.LastLogout = proto.Int32(this.LastLogout)
	pb.LastLogin = proto.Int32(this.LastLogin)
	pb.Spirit = proto.Int32(this.Spirit)
	pb.FriendPoints = proto.Int32(this.FriendPoints)
	pb.LastRefreshShopTime = proto.Int32(this.LastRefreshShopTime)
	pb.VipLvl = proto.Int32(this.VipLvl)
	pb.FriendMessageUnreadCurrId = proto.Int32(this.FriendMessageUnreadCurrId)
	return
}
func (this* dbPlayerInfoData)clone_to(d *dbPlayerInfoData){
	d.Lvl = this.Lvl
	d.Exp = this.Exp
	d.CreateUnix = this.CreateUnix
	d.Gold = this.Gold
	d.Diamond = this.Diamond
	d.LastDialyTaskUpUinx = this.LastDialyTaskUpUinx
	d.Icon = this.Icon
	d.CustomIcon = this.CustomIcon
	d.LastLogout = this.LastLogout
	d.LastLogin = this.LastLogin
	d.Spirit = this.Spirit
	d.FriendPoints = this.FriendPoints
	d.LastRefreshShopTime = this.LastRefreshShopTime
	d.VipLvl = this.VipLvl
	d.FriendMessageUnreadCurrId = this.FriendMessageUnreadCurrId
	return
}
type dbPlayerGlobalData struct{
	CurrentRoleId int32
}
func (this* dbPlayerGlobalData)from_pb(pb *db.PlayerGlobal){
	if pb == nil {
		return
	}
	this.CurrentRoleId = pb.GetCurrentRoleId()
	return
}
func (this* dbPlayerGlobalData)to_pb()(pb *db.PlayerGlobal){
	pb = &db.PlayerGlobal{}
	pb.CurrentRoleId = proto.Int32(this.CurrentRoleId)
	return
}
func (this* dbPlayerGlobalData)clone_to(d *dbPlayerGlobalData){
	d.CurrentRoleId = this.CurrentRoleId
	return
}
type dbPlayerItemData struct{
	Id int32
	Count int32
}
func (this* dbPlayerItemData)from_pb(pb *db.PlayerItem){
	if pb == nil {
		return
	}
	this.Id = pb.GetId()
	this.Count = pb.GetCount()
	return
}
func (this* dbPlayerItemData)to_pb()(pb *db.PlayerItem){
	pb = &db.PlayerItem{}
	pb.Id = proto.Int32(this.Id)
	pb.Count = proto.Int32(this.Count)
	return
}
func (this* dbPlayerItemData)clone_to(d *dbPlayerItemData){
	d.Id = this.Id
	d.Count = this.Count
	return
}
type dbPlayerRoleData struct{
	Id int32
	TableId int32
	Rank int32
	Level int32
	Attr []int32
	Equip []int32
}
func (this* dbPlayerRoleData)from_pb(pb *db.PlayerRole){
	if pb == nil {
		this.Attr = make([]int32,0)
		this.Equip = make([]int32,0)
		return
	}
	this.Id = pb.GetId()
	this.TableId = pb.GetTableId()
	this.Rank = pb.GetRank()
	this.Level = pb.GetLevel()
	this.Attr = make([]int32,len(pb.GetAttr()))
	for i, v := range pb.GetAttr() {
		this.Attr[i] = v
	}
	this.Equip = make([]int32,len(pb.GetEquip()))
	for i, v := range pb.GetEquip() {
		this.Equip[i] = v
	}
	return
}
func (this* dbPlayerRoleData)to_pb()(pb *db.PlayerRole){
	pb = &db.PlayerRole{}
	pb.Id = proto.Int32(this.Id)
	pb.TableId = proto.Int32(this.TableId)
	pb.Rank = proto.Int32(this.Rank)
	pb.Level = proto.Int32(this.Level)
	pb.Attr = make([]int32, len(this.Attr))
	for i, v := range this.Attr {
		pb.Attr[i]=v
	}
	pb.Equip = make([]int32, len(this.Equip))
	for i, v := range this.Equip {
		pb.Equip[i]=v
	}
	return
}
func (this* dbPlayerRoleData)clone_to(d *dbPlayerRoleData){
	d.Id = this.Id
	d.TableId = this.TableId
	d.Rank = this.Rank
	d.Level = this.Level
	d.Attr = make([]int32, len(this.Attr))
	for _ii, _vv := range this.Attr {
		d.Attr[_ii]=_vv
	}
	d.Equip = make([]int32, len(this.Equip))
	for _ii, _vv := range this.Equip {
		d.Equip[_ii]=_vv
	}
	return
}
type dbPlayerBattleTeamData struct{
	AttackMembers []int32
	DefenseMembers []int32
}
func (this* dbPlayerBattleTeamData)from_pb(pb *db.PlayerBattleTeam){
	if pb == nil {
		this.AttackMembers = make([]int32,0)
		this.DefenseMembers = make([]int32,0)
		return
	}
	this.AttackMembers = make([]int32,len(pb.GetAttackMembers()))
	for i, v := range pb.GetAttackMembers() {
		this.AttackMembers[i] = v
	}
	this.DefenseMembers = make([]int32,len(pb.GetDefenseMembers()))
	for i, v := range pb.GetDefenseMembers() {
		this.DefenseMembers[i] = v
	}
	return
}
func (this* dbPlayerBattleTeamData)to_pb()(pb *db.PlayerBattleTeam){
	pb = &db.PlayerBattleTeam{}
	pb.AttackMembers = make([]int32, len(this.AttackMembers))
	for i, v := range this.AttackMembers {
		pb.AttackMembers[i]=v
	}
	pb.DefenseMembers = make([]int32, len(this.DefenseMembers))
	for i, v := range this.DefenseMembers {
		pb.DefenseMembers[i]=v
	}
	return
}
func (this* dbPlayerBattleTeamData)clone_to(d *dbPlayerBattleTeamData){
	d.AttackMembers = make([]int32, len(this.AttackMembers))
	for _ii, _vv := range this.AttackMembers {
		d.AttackMembers[_ii]=_vv
	}
	d.DefenseMembers = make([]int32, len(this.DefenseMembers))
	for _ii, _vv := range this.DefenseMembers {
		d.DefenseMembers[_ii]=_vv
	}
	return
}
type dbPlayerCampaignCommonData struct{
	CurrentCampaignId int32
	HangupLastDropStaticIncomeTime int32
	HangupLastDropRandomIncomeTime int32
	HangupCampaignId int32
}
func (this* dbPlayerCampaignCommonData)from_pb(pb *db.PlayerCampaignCommon){
	if pb == nil {
		return
	}
	this.CurrentCampaignId = pb.GetCurrentCampaignId()
	this.HangupLastDropStaticIncomeTime = pb.GetHangupLastDropStaticIncomeTime()
	this.HangupLastDropRandomIncomeTime = pb.GetHangupLastDropRandomIncomeTime()
	this.HangupCampaignId = pb.GetHangupCampaignId()
	return
}
func (this* dbPlayerCampaignCommonData)to_pb()(pb *db.PlayerCampaignCommon){
	pb = &db.PlayerCampaignCommon{}
	pb.CurrentCampaignId = proto.Int32(this.CurrentCampaignId)
	pb.HangupLastDropStaticIncomeTime = proto.Int32(this.HangupLastDropStaticIncomeTime)
	pb.HangupLastDropRandomIncomeTime = proto.Int32(this.HangupLastDropRandomIncomeTime)
	pb.HangupCampaignId = proto.Int32(this.HangupCampaignId)
	return
}
func (this* dbPlayerCampaignCommonData)clone_to(d *dbPlayerCampaignCommonData){
	d.CurrentCampaignId = this.CurrentCampaignId
	d.HangupLastDropStaticIncomeTime = this.HangupLastDropStaticIncomeTime
	d.HangupLastDropRandomIncomeTime = this.HangupLastDropRandomIncomeTime
	d.HangupCampaignId = this.HangupCampaignId
	return
}
type dbPlayerCampaignData struct{
	CampaignId int32
	Stars int32
	LastFinishedTime int32
	TopScore int32
	PlayedCount int32
	PassCount int32
}
func (this* dbPlayerCampaignData)from_pb(pb *db.PlayerCampaign){
	if pb == nil {
		return
	}
	this.CampaignId = pb.GetCampaignId()
	this.Stars = pb.GetStars()
	this.LastFinishedTime = pb.GetLastFinishedTime()
	this.TopScore = pb.GetTopScore()
	this.PlayedCount = pb.GetPlayedCount()
	this.PassCount = pb.GetPassCount()
	return
}
func (this* dbPlayerCampaignData)to_pb()(pb *db.PlayerCampaign){
	pb = &db.PlayerCampaign{}
	pb.CampaignId = proto.Int32(this.CampaignId)
	pb.Stars = proto.Int32(this.Stars)
	pb.LastFinishedTime = proto.Int32(this.LastFinishedTime)
	pb.TopScore = proto.Int32(this.TopScore)
	pb.PlayedCount = proto.Int32(this.PlayedCount)
	pb.PassCount = proto.Int32(this.PassCount)
	return
}
func (this* dbPlayerCampaignData)clone_to(d *dbPlayerCampaignData){
	d.CampaignId = this.CampaignId
	d.Stars = this.Stars
	d.LastFinishedTime = this.LastFinishedTime
	d.TopScore = this.TopScore
	d.PlayedCount = this.PlayedCount
	d.PassCount = this.PassCount
	return
}
type dbPlayerChapterUnLockData struct{
	ChapterId int32
	PlayerIds []int32
	CurHelpIds []int32
	StartUnix int32
}
func (this* dbPlayerChapterUnLockData)from_pb(pb *db.PlayerChapterUnLock){
	if pb == nil {
		this.PlayerIds = make([]int32,0)
		this.CurHelpIds = make([]int32,0)
		return
	}
	this.ChapterId = pb.GetChapterId()
	this.PlayerIds = make([]int32,len(pb.GetPlayerIds()))
	for i, v := range pb.GetPlayerIds() {
		this.PlayerIds[i] = v
	}
	this.CurHelpIds = make([]int32,len(pb.GetCurHelpIds()))
	for i, v := range pb.GetCurHelpIds() {
		this.CurHelpIds[i] = v
	}
	this.StartUnix = pb.GetStartUnix()
	return
}
func (this* dbPlayerChapterUnLockData)to_pb()(pb *db.PlayerChapterUnLock){
	pb = &db.PlayerChapterUnLock{}
	pb.ChapterId = proto.Int32(this.ChapterId)
	pb.PlayerIds = make([]int32, len(this.PlayerIds))
	for i, v := range this.PlayerIds {
		pb.PlayerIds[i]=v
	}
	pb.CurHelpIds = make([]int32, len(this.CurHelpIds))
	for i, v := range this.CurHelpIds {
		pb.CurHelpIds[i]=v
	}
	pb.StartUnix = proto.Int32(this.StartUnix)
	return
}
func (this* dbPlayerChapterUnLockData)clone_to(d *dbPlayerChapterUnLockData){
	d.ChapterId = this.ChapterId
	d.PlayerIds = make([]int32, len(this.PlayerIds))
	for _ii, _vv := range this.PlayerIds {
		d.PlayerIds[_ii]=_vv
	}
	d.CurHelpIds = make([]int32, len(this.CurHelpIds))
	for _ii, _vv := range this.CurHelpIds {
		d.CurHelpIds[_ii]=_vv
	}
	d.StartUnix = this.StartUnix
	return
}
type dbPlayerShopItemData struct{
	Id int32
	LeftNum int32
}
func (this* dbPlayerShopItemData)from_pb(pb *db.PlayerShopItem){
	if pb == nil {
		return
	}
	this.Id = pb.GetId()
	this.LeftNum = pb.GetLeftNum()
	return
}
func (this* dbPlayerShopItemData)to_pb()(pb *db.PlayerShopItem){
	pb = &db.PlayerShopItem{}
	pb.Id = proto.Int32(this.Id)
	pb.LeftNum = proto.Int32(this.LeftNum)
	return
}
func (this* dbPlayerShopItemData)clone_to(d *dbPlayerShopItemData){
	d.Id = this.Id
	d.LeftNum = this.LeftNum
	return
}
type dbPlayerShopLimitedInfoData struct{
	LimitedDays int32
	LastSaveTime int32
}
func (this* dbPlayerShopLimitedInfoData)from_pb(pb *db.PlayerShopLimitedInfo){
	if pb == nil {
		return
	}
	this.LimitedDays = pb.GetLimitedDays()
	this.LastSaveTime = pb.GetLastSaveTime()
	return
}
func (this* dbPlayerShopLimitedInfoData)to_pb()(pb *db.PlayerShopLimitedInfo){
	pb = &db.PlayerShopLimitedInfo{}
	pb.LimitedDays = proto.Int32(this.LimitedDays)
	pb.LastSaveTime = proto.Int32(this.LastSaveTime)
	return
}
func (this* dbPlayerShopLimitedInfoData)clone_to(d *dbPlayerShopLimitedInfoData){
	d.LimitedDays = this.LimitedDays
	d.LastSaveTime = this.LastSaveTime
	return
}
type dbPlayerChestData struct{
	Pos int32
	ChestId int32
	OpenSec int32
}
func (this* dbPlayerChestData)from_pb(pb *db.PlayerChest){
	if pb == nil {
		return
	}
	this.Pos = pb.GetPos()
	this.ChestId = pb.GetChestId()
	this.OpenSec = pb.GetOpenSec()
	return
}
func (this* dbPlayerChestData)to_pb()(pb *db.PlayerChest){
	pb = &db.PlayerChest{}
	pb.Pos = proto.Int32(this.Pos)
	pb.ChestId = proto.Int32(this.ChestId)
	pb.OpenSec = proto.Int32(this.OpenSec)
	return
}
func (this* dbPlayerChestData)clone_to(d *dbPlayerChestData){
	d.Pos = this.Pos
	d.ChestId = this.ChestId
	d.OpenSec = this.OpenSec
	return
}
type dbPlayerMailData struct{
	MailId int32
	MailType int8
	MailTitle string
	SenderId int32
	SenderName string
	Content string
	SendUnix int32
	OverUnix int32
	ObjIds []int32
	ObjNums []int32
	State int8
	ExtraDatas []int32
}
func (this* dbPlayerMailData)from_pb(pb *db.PlayerMail){
	if pb == nil {
		this.ObjIds = make([]int32,0)
		this.ObjNums = make([]int32,0)
		this.ExtraDatas = make([]int32,0)
		return
	}
	this.MailId = pb.GetMailId()
	this.MailType = int8(pb.GetMailType())
	this.MailTitle = pb.GetMailTitle()
	this.SenderId = pb.GetSenderId()
	this.SenderName = pb.GetSenderName()
	this.Content = pb.GetContent()
	this.SendUnix = pb.GetSendUnix()
	this.OverUnix = pb.GetOverUnix()
	this.ObjIds = make([]int32,len(pb.GetObjIds()))
	for i, v := range pb.GetObjIds() {
		this.ObjIds[i] = v
	}
	this.ObjNums = make([]int32,len(pb.GetObjNums()))
	for i, v := range pb.GetObjNums() {
		this.ObjNums[i] = v
	}
	this.State = int8(pb.GetState())
	this.ExtraDatas = make([]int32,len(pb.GetExtraDatas()))
	for i, v := range pb.GetExtraDatas() {
		this.ExtraDatas[i] = v
	}
	return
}
func (this* dbPlayerMailData)to_pb()(pb *db.PlayerMail){
	pb = &db.PlayerMail{}
	pb.MailId = proto.Int32(this.MailId)
	temp_MailType:=int32(this.MailType)
	pb.MailType = proto.Int32(temp_MailType)
	pb.MailTitle = proto.String(this.MailTitle)
	pb.SenderId = proto.Int32(this.SenderId)
	pb.SenderName = proto.String(this.SenderName)
	pb.Content = proto.String(this.Content)
	pb.SendUnix = proto.Int32(this.SendUnix)
	pb.OverUnix = proto.Int32(this.OverUnix)
	pb.ObjIds = make([]int32, len(this.ObjIds))
	for i, v := range this.ObjIds {
		pb.ObjIds[i]=v
	}
	pb.ObjNums = make([]int32, len(this.ObjNums))
	for i, v := range this.ObjNums {
		pb.ObjNums[i]=v
	}
	temp_State:=int32(this.State)
	pb.State = proto.Int32(temp_State)
	pb.ExtraDatas = make([]int32, len(this.ExtraDatas))
	for i, v := range this.ExtraDatas {
		pb.ExtraDatas[i]=v
	}
	return
}
func (this* dbPlayerMailData)clone_to(d *dbPlayerMailData){
	d.MailId = this.MailId
	d.MailType = int8(this.MailType)
	d.MailTitle = this.MailTitle
	d.SenderId = this.SenderId
	d.SenderName = this.SenderName
	d.Content = this.Content
	d.SendUnix = this.SendUnix
	d.OverUnix = this.OverUnix
	d.ObjIds = make([]int32, len(this.ObjIds))
	for _ii, _vv := range this.ObjIds {
		d.ObjIds[_ii]=_vv
	}
	d.ObjNums = make([]int32, len(this.ObjNums))
	for _ii, _vv := range this.ObjNums {
		d.ObjNums[_ii]=_vv
	}
	d.State = int8(this.State)
	d.ExtraDatas = make([]int32, len(this.ExtraDatas))
	for _ii, _vv := range this.ExtraDatas {
		d.ExtraDatas[_ii]=_vv
	}
	return
}
type dbPlayerDialyTaskData struct{
	TaskId int32
	Value int32
	RewardUnix int32
}
func (this* dbPlayerDialyTaskData)from_pb(pb *db.PlayerDialyTask){
	if pb == nil {
		return
	}
	this.TaskId = pb.GetTaskId()
	this.Value = pb.GetValue()
	this.RewardUnix = pb.GetRewardUnix()
	return
}
func (this* dbPlayerDialyTaskData)to_pb()(pb *db.PlayerDialyTask){
	pb = &db.PlayerDialyTask{}
	pb.TaskId = proto.Int32(this.TaskId)
	pb.Value = proto.Int32(this.Value)
	pb.RewardUnix = proto.Int32(this.RewardUnix)
	return
}
func (this* dbPlayerDialyTaskData)clone_to(d *dbPlayerDialyTaskData){
	d.TaskId = this.TaskId
	d.Value = this.Value
	d.RewardUnix = this.RewardUnix
	return
}
type dbPlayerAchieveData struct{
	AchieveId int32
	Value int32
	RewardUnix int32
}
func (this* dbPlayerAchieveData)from_pb(pb *db.PlayerAchieve){
	if pb == nil {
		return
	}
	this.AchieveId = pb.GetAchieveId()
	this.Value = pb.GetValue()
	this.RewardUnix = pb.GetRewardUnix()
	return
}
func (this* dbPlayerAchieveData)to_pb()(pb *db.PlayerAchieve){
	pb = &db.PlayerAchieve{}
	pb.AchieveId = proto.Int32(this.AchieveId)
	pb.Value = proto.Int32(this.Value)
	pb.RewardUnix = proto.Int32(this.RewardUnix)
	return
}
func (this* dbPlayerAchieveData)clone_to(d *dbPlayerAchieveData){
	d.AchieveId = this.AchieveId
	d.Value = this.Value
	d.RewardUnix = this.RewardUnix
	return
}
type dbPlayerFinishedAchieveData struct{
	AchieveId int32
}
func (this* dbPlayerFinishedAchieveData)from_pb(pb *db.PlayerFinishedAchieve){
	if pb == nil {
		return
	}
	this.AchieveId = pb.GetAchieveId()
	return
}
func (this* dbPlayerFinishedAchieveData)to_pb()(pb *db.PlayerFinishedAchieve){
	pb = &db.PlayerFinishedAchieve{}
	pb.AchieveId = proto.Int32(this.AchieveId)
	return
}
func (this* dbPlayerFinishedAchieveData)clone_to(d *dbPlayerFinishedAchieveData){
	d.AchieveId = this.AchieveId
	return
}
type dbPlayerDailyTaskWholeDailyData struct{
	CompleteTaskId int32
}
func (this* dbPlayerDailyTaskWholeDailyData)from_pb(pb *db.PlayerDailyTaskWholeDaily){
	if pb == nil {
		return
	}
	this.CompleteTaskId = pb.GetCompleteTaskId()
	return
}
func (this* dbPlayerDailyTaskWholeDailyData)to_pb()(pb *db.PlayerDailyTaskWholeDaily){
	pb = &db.PlayerDailyTaskWholeDaily{}
	pb.CompleteTaskId = proto.Int32(this.CompleteTaskId)
	return
}
func (this* dbPlayerDailyTaskWholeDailyData)clone_to(d *dbPlayerDailyTaskWholeDailyData){
	d.CompleteTaskId = this.CompleteTaskId
	return
}
type dbPlayerSevenActivityData struct{
	ActivityId int32
	Value int32
	RewardUnix int32
}
func (this* dbPlayerSevenActivityData)from_pb(pb *db.PlayerSevenActivity){
	if pb == nil {
		return
	}
	this.ActivityId = pb.GetActivityId()
	this.Value = pb.GetValue()
	this.RewardUnix = pb.GetRewardUnix()
	return
}
func (this* dbPlayerSevenActivityData)to_pb()(pb *db.PlayerSevenActivity){
	pb = &db.PlayerSevenActivity{}
	pb.ActivityId = proto.Int32(this.ActivityId)
	pb.Value = proto.Int32(this.Value)
	pb.RewardUnix = proto.Int32(this.RewardUnix)
	return
}
func (this* dbPlayerSevenActivityData)clone_to(d *dbPlayerSevenActivityData){
	d.ActivityId = this.ActivityId
	d.Value = this.Value
	d.RewardUnix = this.RewardUnix
	return
}
type dbPlayerGuidesData struct{
	GuideId int32
	SetUnix int32
}
func (this* dbPlayerGuidesData)from_pb(pb *db.PlayerGuides){
	if pb == nil {
		return
	}
	this.GuideId = pb.GetGuideId()
	this.SetUnix = pb.GetSetUnix()
	return
}
func (this* dbPlayerGuidesData)to_pb()(pb *db.PlayerGuides){
	pb = &db.PlayerGuides{}
	pb.GuideId = proto.Int32(this.GuideId)
	pb.SetUnix = proto.Int32(this.SetUnix)
	return
}
func (this* dbPlayerGuidesData)clone_to(d *dbPlayerGuidesData){
	d.GuideId = this.GuideId
	d.SetUnix = this.SetUnix
	return
}
type dbPlayerFriendRelativeData struct{
	LastGiveFriendPointsTime int32
	GiveNumToday int32
	LastRefreshTime int32
}
func (this* dbPlayerFriendRelativeData)from_pb(pb *db.PlayerFriendRelative){
	if pb == nil {
		return
	}
	this.LastGiveFriendPointsTime = pb.GetLastGiveFriendPointsTime()
	this.GiveNumToday = pb.GetGiveNumToday()
	this.LastRefreshTime = pb.GetLastRefreshTime()
	return
}
func (this* dbPlayerFriendRelativeData)to_pb()(pb *db.PlayerFriendRelative){
	pb = &db.PlayerFriendRelative{}
	pb.LastGiveFriendPointsTime = proto.Int32(this.LastGiveFriendPointsTime)
	pb.GiveNumToday = proto.Int32(this.GiveNumToday)
	pb.LastRefreshTime = proto.Int32(this.LastRefreshTime)
	return
}
func (this* dbPlayerFriendRelativeData)clone_to(d *dbPlayerFriendRelativeData){
	d.LastGiveFriendPointsTime = this.LastGiveFriendPointsTime
	d.GiveNumToday = this.GiveNumToday
	d.LastRefreshTime = this.LastRefreshTime
	return
}
type dbPlayerFriendData struct{
	FriendId int32
	FriendName string
	Head string
	Level int32
	VipLevel int32
	LastLogin int32
	LastGivePointsTime int32
}
func (this* dbPlayerFriendData)from_pb(pb *db.PlayerFriend){
	if pb == nil {
		return
	}
	this.FriendId = pb.GetFriendId()
	this.FriendName = pb.GetFriendName()
	this.Head = pb.GetHead()
	this.Level = pb.GetLevel()
	this.VipLevel = pb.GetVipLevel()
	this.LastLogin = pb.GetLastLogin()
	this.LastGivePointsTime = pb.GetLastGivePointsTime()
	return
}
func (this* dbPlayerFriendData)to_pb()(pb *db.PlayerFriend){
	pb = &db.PlayerFriend{}
	pb.FriendId = proto.Int32(this.FriendId)
	pb.FriendName = proto.String(this.FriendName)
	pb.Head = proto.String(this.Head)
	pb.Level = proto.Int32(this.Level)
	pb.VipLevel = proto.Int32(this.VipLevel)
	pb.LastLogin = proto.Int32(this.LastLogin)
	pb.LastGivePointsTime = proto.Int32(this.LastGivePointsTime)
	return
}
func (this* dbPlayerFriendData)clone_to(d *dbPlayerFriendData){
	d.FriendId = this.FriendId
	d.FriendName = this.FriendName
	d.Head = this.Head
	d.Level = this.Level
	d.VipLevel = this.VipLevel
	d.LastLogin = this.LastLogin
	d.LastGivePointsTime = this.LastGivePointsTime
	return
}
type dbPlayerFriendReqData struct{
	PlayerId int32
	PlayerName string
	ReqUnix int32
}
func (this* dbPlayerFriendReqData)from_pb(pb *db.PlayerFriendReq){
	if pb == nil {
		return
	}
	this.PlayerId = pb.GetPlayerId()
	this.PlayerName = pb.GetPlayerName()
	this.ReqUnix = pb.GetReqUnix()
	return
}
func (this* dbPlayerFriendReqData)to_pb()(pb *db.PlayerFriendReq){
	pb = &db.PlayerFriendReq{}
	pb.PlayerId = proto.Int32(this.PlayerId)
	pb.PlayerName = proto.String(this.PlayerName)
	pb.ReqUnix = proto.Int32(this.ReqUnix)
	return
}
func (this* dbPlayerFriendReqData)clone_to(d *dbPlayerFriendReqData){
	d.PlayerId = this.PlayerId
	d.PlayerName = this.PlayerName
	d.ReqUnix = this.ReqUnix
	return
}
type dbPlayerFriendPointData struct{
	FromPlayerId int32
	GivePoints int32
	LastGiveTime int32
	IsTodayGive int32
}
func (this* dbPlayerFriendPointData)from_pb(pb *db.PlayerFriendPoint){
	if pb == nil {
		return
	}
	this.FromPlayerId = pb.GetFromPlayerId()
	this.GivePoints = pb.GetGivePoints()
	this.LastGiveTime = pb.GetLastGiveTime()
	this.IsTodayGive = pb.GetIsTodayGive()
	return
}
func (this* dbPlayerFriendPointData)to_pb()(pb *db.PlayerFriendPoint){
	pb = &db.PlayerFriendPoint{}
	pb.FromPlayerId = proto.Int32(this.FromPlayerId)
	pb.GivePoints = proto.Int32(this.GivePoints)
	pb.LastGiveTime = proto.Int32(this.LastGiveTime)
	pb.IsTodayGive = proto.Int32(this.IsTodayGive)
	return
}
func (this* dbPlayerFriendPointData)clone_to(d *dbPlayerFriendPointData){
	d.FromPlayerId = this.FromPlayerId
	d.GivePoints = this.GivePoints
	d.LastGiveTime = this.LastGiveTime
	d.IsTodayGive = this.IsTodayGive
	return
}
type dbPlayerFriendChatUnreadIdData struct{
	FriendId int32
	MessageIds []int32
	CurrMessageId int32
}
func (this* dbPlayerFriendChatUnreadIdData)from_pb(pb *db.PlayerFriendChatUnreadId){
	if pb == nil {
		this.MessageIds = make([]int32,0)
		return
	}
	this.FriendId = pb.GetFriendId()
	this.MessageIds = make([]int32,len(pb.GetMessageIds()))
	for i, v := range pb.GetMessageIds() {
		this.MessageIds[i] = v
	}
	this.CurrMessageId = pb.GetCurrMessageId()
	return
}
func (this* dbPlayerFriendChatUnreadIdData)to_pb()(pb *db.PlayerFriendChatUnreadId){
	pb = &db.PlayerFriendChatUnreadId{}
	pb.FriendId = proto.Int32(this.FriendId)
	pb.MessageIds = make([]int32, len(this.MessageIds))
	for i, v := range this.MessageIds {
		pb.MessageIds[i]=v
	}
	pb.CurrMessageId = proto.Int32(this.CurrMessageId)
	return
}
func (this* dbPlayerFriendChatUnreadIdData)clone_to(d *dbPlayerFriendChatUnreadIdData){
	d.FriendId = this.FriendId
	d.MessageIds = make([]int32, len(this.MessageIds))
	for _ii, _vv := range this.MessageIds {
		d.MessageIds[_ii]=_vv
	}
	d.CurrMessageId = this.CurrMessageId
	return
}
type dbPlayerFriendChatUnreadMessageData struct{
	PlayerMessageId int64
	Message []byte
	SendTime int32
	IsRead int32
}
func (this* dbPlayerFriendChatUnreadMessageData)from_pb(pb *db.PlayerFriendChatUnreadMessage){
	if pb == nil {
		return
	}
	this.PlayerMessageId = pb.GetPlayerMessageId()
	this.Message = pb.GetMessage()
	this.SendTime = pb.GetSendTime()
	this.IsRead = pb.GetIsRead()
	return
}
func (this* dbPlayerFriendChatUnreadMessageData)to_pb()(pb *db.PlayerFriendChatUnreadMessage){
	pb = &db.PlayerFriendChatUnreadMessage{}
	pb.PlayerMessageId = proto.Int64(this.PlayerMessageId)
	pb.Message = this.Message
	pb.SendTime = proto.Int32(this.SendTime)
	pb.IsRead = proto.Int32(this.IsRead)
	return
}
func (this* dbPlayerFriendChatUnreadMessageData)clone_to(d *dbPlayerFriendChatUnreadMessageData){
	d.PlayerMessageId = this.PlayerMessageId
	d.Message = make([]byte, len(this.Message))
	for _ii, _vv := range this.Message {
		d.Message[_ii]=_vv
	}
	d.SendTime = this.SendTime
	d.IsRead = this.IsRead
	return
}
type dbPlayerChaterOpenRequestData struct{
	CustomData []byte
}
func (this* dbPlayerChaterOpenRequestData)from_pb(pb *db.PlayerChaterOpenRequest){
	if pb == nil {
		return
	}
	this.CustomData = pb.GetCustomData()
	return
}
func (this* dbPlayerChaterOpenRequestData)to_pb()(pb *db.PlayerChaterOpenRequest){
	pb = &db.PlayerChaterOpenRequest{}
	pb.CustomData = this.CustomData
	return
}
func (this* dbPlayerChaterOpenRequestData)clone_to(d *dbPlayerChaterOpenRequestData){
	d.CustomData = make([]byte, len(this.CustomData))
	for _ii, _vv := range this.CustomData {
		d.CustomData[_ii]=_vv
	}
	return
}
type dbPlayerHandbookItemData struct{
	Id int32
}
func (this* dbPlayerHandbookItemData)from_pb(pb *db.PlayerHandbookItem){
	if pb == nil {
		return
	}
	this.Id = pb.GetId()
	return
}
func (this* dbPlayerHandbookItemData)to_pb()(pb *db.PlayerHandbookItem){
	pb = &db.PlayerHandbookItem{}
	pb.Id = proto.Int32(this.Id)
	return
}
func (this* dbPlayerHandbookItemData)clone_to(d *dbPlayerHandbookItemData){
	d.Id = this.Id
	return
}
type dbPlayerHeadItemData struct{
	Id int32
}
func (this* dbPlayerHeadItemData)from_pb(pb *db.PlayerHeadItem){
	if pb == nil {
		return
	}
	this.Id = pb.GetId()
	return
}
func (this* dbPlayerHeadItemData)to_pb()(pb *db.PlayerHeadItem){
	pb = &db.PlayerHeadItem{}
	pb.Id = proto.Int32(this.Id)
	return
}
func (this* dbPlayerHeadItemData)clone_to(d *dbPlayerHeadItemData){
	d.Id = this.Id
	return
}
type dbPlayerSuitAwardData struct{
	Id int32
	AwardTime int32
}
func (this* dbPlayerSuitAwardData)from_pb(pb *db.PlayerSuitAward){
	if pb == nil {
		return
	}
	this.Id = pb.GetId()
	this.AwardTime = pb.GetAwardTime()
	return
}
func (this* dbPlayerSuitAwardData)to_pb()(pb *db.PlayerSuitAward){
	pb = &db.PlayerSuitAward{}
	pb.Id = proto.Int32(this.Id)
	pb.AwardTime = proto.Int32(this.AwardTime)
	return
}
func (this* dbPlayerSuitAwardData)clone_to(d *dbPlayerSuitAwardData){
	d.Id = this.Id
	d.AwardTime = this.AwardTime
	return
}
type dbPlayerWorldChatData struct{
	LastChatTime int32
	LastPullTime int32
	LastMsgIndex int32
}
func (this* dbPlayerWorldChatData)from_pb(pb *db.PlayerWorldChat){
	if pb == nil {
		return
	}
	this.LastChatTime = pb.GetLastChatTime()
	this.LastPullTime = pb.GetLastPullTime()
	this.LastMsgIndex = pb.GetLastMsgIndex()
	return
}
func (this* dbPlayerWorldChatData)to_pb()(pb *db.PlayerWorldChat){
	pb = &db.PlayerWorldChat{}
	pb.LastChatTime = proto.Int32(this.LastChatTime)
	pb.LastPullTime = proto.Int32(this.LastPullTime)
	pb.LastMsgIndex = proto.Int32(this.LastMsgIndex)
	return
}
func (this* dbPlayerWorldChatData)clone_to(d *dbPlayerWorldChatData){
	d.LastChatTime = this.LastChatTime
	d.LastPullTime = this.LastPullTime
	d.LastMsgIndex = this.LastMsgIndex
	return
}
type dbPlayerAnouncementData struct{
	LastSendTime int32
}
func (this* dbPlayerAnouncementData)from_pb(pb *db.PlayerAnouncement){
	if pb == nil {
		return
	}
	this.LastSendTime = pb.GetLastSendTime()
	return
}
func (this* dbPlayerAnouncementData)to_pb()(pb *db.PlayerAnouncement){
	pb = &db.PlayerAnouncement{}
	pb.LastSendTime = proto.Int32(this.LastSendTime)
	return
}
func (this* dbPlayerAnouncementData)clone_to(d *dbPlayerAnouncementData){
	d.LastSendTime = this.LastSendTime
	return
}
type dbPlayerFirstDrawCardData struct{
	Id int32
	Drawed int32
}
func (this* dbPlayerFirstDrawCardData)from_pb(pb *db.PlayerFirstDrawCard){
	if pb == nil {
		return
	}
	this.Id = pb.GetId()
	this.Drawed = pb.GetDrawed()
	return
}
func (this* dbPlayerFirstDrawCardData)to_pb()(pb *db.PlayerFirstDrawCard){
	pb = &db.PlayerFirstDrawCard{}
	pb.Id = proto.Int32(this.Id)
	pb.Drawed = proto.Int32(this.Drawed)
	return
}
func (this* dbPlayerFirstDrawCardData)clone_to(d *dbPlayerFirstDrawCardData){
	d.Id = this.Id
	d.Drawed = this.Drawed
	return
}

func (this *dbGlobalRow)GetCurrentPlayerId( )(r int32 ){
	this.m_lock.UnSafeRLock("dbGlobalRow.GetdbGlobalCurrentPlayerIdColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_CurrentPlayerId)
}
func (this *dbGlobalRow)SetCurrentPlayerId(v int32){
	this.m_lock.UnSafeLock("dbGlobalRow.SetdbGlobalCurrentPlayerIdColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_CurrentPlayerId=int32(v)
	this.m_CurrentPlayerId_changed=true
	return
}
type dbGlobalRow struct {
	m_table *dbGlobalTable
	m_lock       *RWMutex
	m_loaded  bool
	m_new     bool
	m_remove  bool
	m_touch      int32
	m_releasable bool
	m_valid   bool
	m_Id        int32
	m_CurrentPlayerId_changed bool
	m_CurrentPlayerId int32
}
func new_dbGlobalRow(table *dbGlobalTable, Id int32) (r *dbGlobalRow) {
	this := &dbGlobalRow{}
	this.m_table = table
	this.m_Id = Id
	this.m_lock = NewRWMutex()
	this.m_CurrentPlayerId_changed=true
	return this
}
func (this *dbGlobalRow) save_data(release bool) (err error, released bool, state int32, update_string string, args []interface{}) {
	this.m_lock.UnSafeLock("dbGlobalRow.save_data")
	defer this.m_lock.UnSafeUnlock()
	if this.m_new {
		db_args:=new_db_args(2)
		db_args.Push(this.m_Id)
		db_args.Push(this.m_CurrentPlayerId)
		args=db_args.GetArgs()
		state = 1
	} else {
		if this.m_CurrentPlayerId_changed{
			update_string = "UPDATE Global SET "
			db_args:=new_db_args(2)
			if this.m_CurrentPlayerId_changed{
				update_string+="CurrentPlayerId=?,"
				db_args.Push(this.m_CurrentPlayerId)
			}
			update_string = strings.TrimRight(update_string, ", ")
			update_string+=" WHERE Id=?"
			db_args.Push(this.m_Id)
			args=db_args.GetArgs()
			state = 2
		}
	}
	this.m_new = false
	this.m_CurrentPlayerId_changed = false
	if release && this.m_loaded {
		this.m_loaded = false
		released = true
	}
	return nil,released,state,update_string,args
}
func (this *dbGlobalRow) Save(release bool) (err error, d bool, released bool) {
	err,released, state, update_string, args := this.save_data(release)
	if err != nil {
		log.Error("save data failed")
		return err, false, false
	}
	if state == 0 {
		d = false
	} else if state == 1 {
		_, err = this.m_table.m_dbc.StmtExec(this.m_table.m_save_insert_stmt, args...)
		if err != nil {
			log.Error("INSERT Global exec failed %v ", this.m_Id)
			return err, false, released
		}
		d = true
	} else if state == 2 {
		_, err = this.m_table.m_dbc.Exec(update_string, args...)
		if err != nil {
			log.Error("UPDATE Global exec failed %v", this.m_Id)
			return err, false, released
		}
		d = true
	}
	return nil, d, released
}
type dbGlobalTable struct{
	m_dbc *DBC
	m_lock *RWMutex
	m_row *dbGlobalRow
	m_preload_select_stmt *sql.Stmt
	m_save_insert_stmt *sql.Stmt
}
func new_dbGlobalTable(dbc *DBC) (this *dbGlobalTable) {
	this = &dbGlobalTable{}
	this.m_dbc = dbc
	this.m_lock = NewRWMutex()
	return this
}
func (this *dbGlobalTable) check_create_table() (err error) {
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS Global(Id int(11),PRIMARY KEY (Id))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS Global failed")
		return
	}
	rows, err := this.m_dbc.Query("SELECT COLUMN_NAME,ORDINAL_POSITION FROM information_schema.`COLUMNS` WHERE TABLE_SCHEMA=? AND TABLE_NAME='Global'", this.m_dbc.m_db_name)
	if err != nil {
		log.Error("SELECT information_schema failed")
		return
	}
	columns := make(map[string]int32)
	for rows.Next() {
		var column_name string
		var ordinal_position int32
		err = rows.Scan(&column_name, &ordinal_position)
		if err != nil {
			log.Error("scan information_schema row failed")
			return
		}
		if ordinal_position < 1 {
			log.Error("col ordinal out of range")
			continue
		}
		columns[column_name] = ordinal_position
	}
	_, hasCurrentPlayerId := columns["CurrentPlayerId"]
	if !hasCurrentPlayerId {
		_, err = this.m_dbc.Exec("ALTER TABLE Global ADD COLUMN CurrentPlayerId int(11)")
		if err != nil {
			log.Error("ADD COLUMN CurrentPlayerId failed")
			return
		}
	}
	return
}
func (this *dbGlobalTable) prepare_preload_select_stmt() (err error) {
	this.m_preload_select_stmt,err=this.m_dbc.StmtPrepare("SELECT CurrentPlayerId FROM Global WHERE Id=0")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbGlobalTable) prepare_save_insert_stmt()(err error){
	this.m_save_insert_stmt,err=this.m_dbc.StmtPrepare("INSERT INTO Global (Id,CurrentPlayerId) VALUES (?,?)")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbGlobalTable) Init() (err error) {
	err=this.check_create_table()
	if err!=nil{
		log.Error("check_create_table failed")
		return
	}
	err=this.prepare_preload_select_stmt()
	if err!=nil{
		log.Error("prepare_preload_select_stmt failed")
		return
	}
	err=this.prepare_save_insert_stmt()
	if err!=nil{
		log.Error("prepare_save_insert_stmt failed")
		return
	}
	return
}
func (this *dbGlobalTable) Preload() (err error) {
	r := this.m_dbc.StmtQueryRow(this.m_preload_select_stmt)
	var dCurrentPlayerId int32
	err = r.Scan(&dCurrentPlayerId)
	if err!=nil{
		if err!=sql.ErrNoRows{
			log.Error("Scan failed")
			return
		}
	}else{
		row := new_dbGlobalRow(this,0)
		row.m_CurrentPlayerId=dCurrentPlayerId
		row.m_CurrentPlayerId_changed=false
		row.m_valid = true
		row.m_loaded=true
		this.m_row=row
	}
	if this.m_row == nil {
		this.m_row = new_dbGlobalRow(this, 0)
		this.m_row.m_new = true
		this.m_row.m_valid = true
		err = this.Save(false)
		if err != nil {
			log.Error("save failed")
			return
		}
		this.m_row.m_loaded = true
	}
	return
}
func (this *dbGlobalTable) Save(quick bool) (err error) {
	if this.m_row==nil{
		return errors.New("row nil")
	}
	err, _, _ = this.m_row.Save(false)
	return err
}
func (this *dbGlobalTable) GetRow( ) (row *dbGlobalRow) {
	return this.m_row
}
func (this *dbPlayerRow)GetAccount( )(r string ){
	this.m_lock.UnSafeRLock("dbPlayerRow.GetdbPlayerAccountColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_Account)
}
func (this *dbPlayerRow)SetAccount(v string){
	this.m_lock.UnSafeLock("dbPlayerRow.SetdbPlayerAccountColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_Account=string(v)
	this.m_Account_changed=true
	return
}
func (this *dbPlayerRow)GetName( )(r string ){
	this.m_lock.UnSafeRLock("dbPlayerRow.GetdbPlayerNameColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_Name)
}
func (this *dbPlayerRow)SetName(v string){
	this.m_lock.UnSafeLock("dbPlayerRow.SetdbPlayerNameColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_Name=string(v)
	this.m_Name_changed=true
	return
}
type dbPlayerInfoColumn struct{
	m_row *dbPlayerRow
	m_data *dbPlayerInfoData
	m_changed bool
}
func (this *dbPlayerInfoColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_data = &dbPlayerInfoData{}
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerInfo{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_data = &dbPlayerInfoData{}
	this.m_data.from_pb(pb)
	this.m_changed = false
	return
}
func (this *dbPlayerInfoColumn)save( )(data []byte,err error){
	pb:=this.m_data.to_pb()
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerInfoColumn)Get( )(v *dbPlayerInfoData ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v=&dbPlayerInfoData{}
	this.m_data.clone_to(v)
	return
}
func (this *dbPlayerInfoColumn)Set(v dbPlayerInfoData ){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=&dbPlayerInfoData{}
	v.clone_to(this.m_data)
	this.m_changed=true
	return
}
func (this *dbPlayerInfoColumn)GetLvl( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetLvl")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.Lvl
	return
}
func (this *dbPlayerInfoColumn)SetLvl(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetLvl")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.Lvl = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)GetExp( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetExp")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.Exp
	return
}
func (this *dbPlayerInfoColumn)SetExp(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetExp")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.Exp = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)GetCreateUnix( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetCreateUnix")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.CreateUnix
	return
}
func (this *dbPlayerInfoColumn)SetCreateUnix(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetCreateUnix")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.CreateUnix = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)GetGold( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetGold")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.Gold
	return
}
func (this *dbPlayerInfoColumn)SetGold(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetGold")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.Gold = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)GetDiamond( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetDiamond")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.Diamond
	return
}
func (this *dbPlayerInfoColumn)SetDiamond(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetDiamond")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.Diamond = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)GetLastDialyTaskUpUinx( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetLastDialyTaskUpUinx")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.LastDialyTaskUpUinx
	return
}
func (this *dbPlayerInfoColumn)SetLastDialyTaskUpUinx(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetLastDialyTaskUpUinx")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.LastDialyTaskUpUinx = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)GetIcon( )(v string ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetIcon")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.Icon
	return
}
func (this *dbPlayerInfoColumn)SetIcon(v string){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetIcon")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.Icon = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)GetCustomIcon( )(v string ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetCustomIcon")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.CustomIcon
	return
}
func (this *dbPlayerInfoColumn)SetCustomIcon(v string){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetCustomIcon")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.CustomIcon = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)GetLastLogout( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetLastLogout")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.LastLogout
	return
}
func (this *dbPlayerInfoColumn)SetLastLogout(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetLastLogout")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.LastLogout = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)GetLastLogin( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetLastLogin")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.LastLogin
	return
}
func (this *dbPlayerInfoColumn)SetLastLogin(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetLastLogin")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.LastLogin = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)GetSpirit( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetSpirit")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.Spirit
	return
}
func (this *dbPlayerInfoColumn)SetSpirit(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetSpirit")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.Spirit = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)IncbySpirit(v int32)(r int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.IncbySpirit")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.Spirit += v
	this.m_changed = true
	return this.m_data.Spirit
}
func (this *dbPlayerInfoColumn)GetFriendPoints( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetFriendPoints")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.FriendPoints
	return
}
func (this *dbPlayerInfoColumn)SetFriendPoints(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetFriendPoints")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.FriendPoints = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)IncbyFriendPoints(v int32)(r int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.IncbyFriendPoints")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.FriendPoints += v
	this.m_changed = true
	return this.m_data.FriendPoints
}
func (this *dbPlayerInfoColumn)GetLastRefreshShopTime( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetLastRefreshShopTime")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.LastRefreshShopTime
	return
}
func (this *dbPlayerInfoColumn)SetLastRefreshShopTime(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetLastRefreshShopTime")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.LastRefreshShopTime = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)GetVipLvl( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetVipLvl")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.VipLvl
	return
}
func (this *dbPlayerInfoColumn)SetVipLvl(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetVipLvl")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.VipLvl = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)GetFriendMessageUnreadCurrId( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerInfoColumn.GetFriendMessageUnreadCurrId")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.FriendMessageUnreadCurrId
	return
}
func (this *dbPlayerInfoColumn)SetFriendMessageUnreadCurrId(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.SetFriendMessageUnreadCurrId")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.FriendMessageUnreadCurrId = v
	this.m_changed = true
	return
}
func (this *dbPlayerInfoColumn)IncbyFriendMessageUnreadCurrId(v int32)(r int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerInfoColumn.IncbyFriendMessageUnreadCurrId")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.FriendMessageUnreadCurrId += v
	this.m_changed = true
	return this.m_data.FriendMessageUnreadCurrId
}
type dbPlayerGlobalColumn struct{
	m_row *dbPlayerRow
	m_data *dbPlayerGlobalData
	m_changed bool
}
func (this *dbPlayerGlobalColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_data = &dbPlayerGlobalData{}
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerGlobal{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_data = &dbPlayerGlobalData{}
	this.m_data.from_pb(pb)
	this.m_changed = false
	return
}
func (this *dbPlayerGlobalColumn)save( )(data []byte,err error){
	pb:=this.m_data.to_pb()
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerGlobalColumn)Get( )(v *dbPlayerGlobalData ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerGlobalColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v=&dbPlayerGlobalData{}
	this.m_data.clone_to(v)
	return
}
func (this *dbPlayerGlobalColumn)Set(v dbPlayerGlobalData ){
	this.m_row.m_lock.UnSafeLock("dbPlayerGlobalColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=&dbPlayerGlobalData{}
	v.clone_to(this.m_data)
	this.m_changed=true
	return
}
func (this *dbPlayerGlobalColumn)GetCurrentRoleId( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerGlobalColumn.GetCurrentRoleId")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.CurrentRoleId
	return
}
func (this *dbPlayerGlobalColumn)SetCurrentRoleId(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerGlobalColumn.SetCurrentRoleId")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.CurrentRoleId = v
	this.m_changed = true
	return
}
func (this *dbPlayerGlobalColumn)IncbyCurrentRoleId(v int32)(r int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerGlobalColumn.IncbyCurrentRoleId")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.CurrentRoleId += v
	this.m_changed = true
	return this.m_data.CurrentRoleId
}
type dbPlayerItemColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerItemData
	m_changed bool
}
func (this *dbPlayerItemColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerItemList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerItemData{}
		d.from_pb(v)
		this.m_data[int32(d.Id)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerItemColumn)save( )(data []byte,err error){
	pb := &db.PlayerItemList{}
	pb.List=make([]*db.PlayerItem,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerItemColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerItemColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerItemColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerItemColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerItemColumn)GetAll()(list []dbPlayerItemData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerItemColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerItemData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerItemColumn)Get(id int32)(v *dbPlayerItemData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerItemColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerItemData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerItemColumn)Set(v dbPlayerItemData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerItemColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.Id)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.Id)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerItemColumn)Add(v *dbPlayerItemData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerItemColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.Id)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.Id)
		return false
	}
	d:=&dbPlayerItemData{}
	v.clone_to(d)
	this.m_data[int32(v.Id)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerItemColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerItemColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerItemColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerItemColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerItemData)
	this.m_changed = true
	return
}
func (this *dbPlayerItemColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerItemColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerItemColumn)GetCount(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerItemColumn.GetCount")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.Count
	return v,true
}
func (this *dbPlayerItemColumn)SetCount(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerItemColumn.SetCount")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.Count = v
	this.m_changed = true
	return true
}
func (this *dbPlayerItemColumn)IncbyCount(id int32,v int32)(r int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerItemColumn.IncbyCount")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		d = &dbPlayerItemData{}
		this.m_data[id] = d
	}
	d.Count +=  v
	this.m_changed = true
	return d.Count
}
type dbPlayerRoleColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerRoleData
	m_changed bool
}
func (this *dbPlayerRoleColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerRoleList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerRoleData{}
		d.from_pb(v)
		this.m_data[int32(d.Id)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerRoleColumn)save( )(data []byte,err error){
	pb := &db.PlayerRoleList{}
	pb.List=make([]*db.PlayerRole,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerRoleColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerRoleColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerRoleColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerRoleColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerRoleColumn)GetAll()(list []dbPlayerRoleData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerRoleColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerRoleData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerRoleColumn)Get(id int32)(v *dbPlayerRoleData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerRoleColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerRoleData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerRoleColumn)Set(v dbPlayerRoleData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerRoleColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.Id)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.Id)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerRoleColumn)Add(v *dbPlayerRoleData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerRoleColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.Id)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.Id)
		return false
	}
	d:=&dbPlayerRoleData{}
	v.clone_to(d)
	this.m_data[int32(v.Id)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerRoleColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerRoleColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerRoleColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerRoleColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerRoleData)
	this.m_changed = true
	return
}
func (this *dbPlayerRoleColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerRoleColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerRoleColumn)GetTableId(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerRoleColumn.GetTableId")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.TableId
	return v,true
}
func (this *dbPlayerRoleColumn)SetTableId(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerRoleColumn.SetTableId")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.TableId = v
	this.m_changed = true
	return true
}
func (this *dbPlayerRoleColumn)GetRank(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerRoleColumn.GetRank")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.Rank
	return v,true
}
func (this *dbPlayerRoleColumn)SetRank(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerRoleColumn.SetRank")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.Rank = v
	this.m_changed = true
	return true
}
func (this *dbPlayerRoleColumn)GetLevel(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerRoleColumn.GetLevel")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.Level
	return v,true
}
func (this *dbPlayerRoleColumn)SetLevel(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerRoleColumn.SetLevel")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.Level = v
	this.m_changed = true
	return true
}
func (this *dbPlayerRoleColumn)GetAttr(id int32)(v []int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerRoleColumn.GetAttr")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = make([]int32, len(d.Attr))
	for _ii, _vv := range d.Attr {
		v[_ii]=_vv
	}
	return v,true
}
func (this *dbPlayerRoleColumn)SetAttr(id int32,v []int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerRoleColumn.SetAttr")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.Attr = make([]int32, len(v))
	for _ii, _vv := range v {
		d.Attr[_ii]=_vv
	}
	this.m_changed = true
	return true
}
func (this *dbPlayerRoleColumn)GetEquip(id int32)(v []int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerRoleColumn.GetEquip")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = make([]int32, len(d.Equip))
	for _ii, _vv := range d.Equip {
		v[_ii]=_vv
	}
	return v,true
}
func (this *dbPlayerRoleColumn)SetEquip(id int32,v []int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerRoleColumn.SetEquip")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.Equip = make([]int32, len(v))
	for _ii, _vv := range v {
		d.Equip[_ii]=_vv
	}
	this.m_changed = true
	return true
}
type dbPlayerBattleTeamColumn struct{
	m_row *dbPlayerRow
	m_data *dbPlayerBattleTeamData
	m_changed bool
}
func (this *dbPlayerBattleTeamColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_data = &dbPlayerBattleTeamData{}
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerBattleTeam{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_data = &dbPlayerBattleTeamData{}
	this.m_data.from_pb(pb)
	this.m_changed = false
	return
}
func (this *dbPlayerBattleTeamColumn)save( )(data []byte,err error){
	pb:=this.m_data.to_pb()
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerBattleTeamColumn)Get( )(v *dbPlayerBattleTeamData ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerBattleTeamColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v=&dbPlayerBattleTeamData{}
	this.m_data.clone_to(v)
	return
}
func (this *dbPlayerBattleTeamColumn)Set(v dbPlayerBattleTeamData ){
	this.m_row.m_lock.UnSafeLock("dbPlayerBattleTeamColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=&dbPlayerBattleTeamData{}
	v.clone_to(this.m_data)
	this.m_changed=true
	return
}
func (this *dbPlayerBattleTeamColumn)GetAttackMembers( )(v []int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerBattleTeamColumn.GetAttackMembers")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = make([]int32, len(this.m_data.AttackMembers))
	for _ii, _vv := range this.m_data.AttackMembers {
		v[_ii]=_vv
	}
	return
}
func (this *dbPlayerBattleTeamColumn)SetAttackMembers(v []int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerBattleTeamColumn.SetAttackMembers")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.AttackMembers = make([]int32, len(v))
	for _ii, _vv := range v {
		this.m_data.AttackMembers[_ii]=_vv
	}
	this.m_changed = true
	return
}
func (this *dbPlayerBattleTeamColumn)GetDefenseMembers( )(v []int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerBattleTeamColumn.GetDefenseMembers")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = make([]int32, len(this.m_data.DefenseMembers))
	for _ii, _vv := range this.m_data.DefenseMembers {
		v[_ii]=_vv
	}
	return
}
func (this *dbPlayerBattleTeamColumn)SetDefenseMembers(v []int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerBattleTeamColumn.SetDefenseMembers")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.DefenseMembers = make([]int32, len(v))
	for _ii, _vv := range v {
		this.m_data.DefenseMembers[_ii]=_vv
	}
	this.m_changed = true
	return
}
type dbPlayerCampaignCommonColumn struct{
	m_row *dbPlayerRow
	m_data *dbPlayerCampaignCommonData
	m_changed bool
}
func (this *dbPlayerCampaignCommonColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_data = &dbPlayerCampaignCommonData{}
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerCampaignCommon{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_data = &dbPlayerCampaignCommonData{}
	this.m_data.from_pb(pb)
	this.m_changed = false
	return
}
func (this *dbPlayerCampaignCommonColumn)save( )(data []byte,err error){
	pb:=this.m_data.to_pb()
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerCampaignCommonColumn)Get( )(v *dbPlayerCampaignCommonData ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignCommonColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v=&dbPlayerCampaignCommonData{}
	this.m_data.clone_to(v)
	return
}
func (this *dbPlayerCampaignCommonColumn)Set(v dbPlayerCampaignCommonData ){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignCommonColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=&dbPlayerCampaignCommonData{}
	v.clone_to(this.m_data)
	this.m_changed=true
	return
}
func (this *dbPlayerCampaignCommonColumn)GetCurrentCampaignId( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignCommonColumn.GetCurrentCampaignId")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.CurrentCampaignId
	return
}
func (this *dbPlayerCampaignCommonColumn)SetCurrentCampaignId(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignCommonColumn.SetCurrentCampaignId")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.CurrentCampaignId = v
	this.m_changed = true
	return
}
func (this *dbPlayerCampaignCommonColumn)GetHangupLastDropStaticIncomeTime( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignCommonColumn.GetHangupLastDropStaticIncomeTime")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.HangupLastDropStaticIncomeTime
	return
}
func (this *dbPlayerCampaignCommonColumn)SetHangupLastDropStaticIncomeTime(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignCommonColumn.SetHangupLastDropStaticIncomeTime")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.HangupLastDropStaticIncomeTime = v
	this.m_changed = true
	return
}
func (this *dbPlayerCampaignCommonColumn)GetHangupLastDropRandomIncomeTime( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignCommonColumn.GetHangupLastDropRandomIncomeTime")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.HangupLastDropRandomIncomeTime
	return
}
func (this *dbPlayerCampaignCommonColumn)SetHangupLastDropRandomIncomeTime(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignCommonColumn.SetHangupLastDropRandomIncomeTime")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.HangupLastDropRandomIncomeTime = v
	this.m_changed = true
	return
}
func (this *dbPlayerCampaignCommonColumn)GetHangupCampaignId( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignCommonColumn.GetHangupCampaignId")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.HangupCampaignId
	return
}
func (this *dbPlayerCampaignCommonColumn)SetHangupCampaignId(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignCommonColumn.SetHangupCampaignId")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.HangupCampaignId = v
	this.m_changed = true
	return
}
type dbPlayerCampaignColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerCampaignData
	m_changed bool
}
func (this *dbPlayerCampaignColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerCampaignList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerCampaignData{}
		d.from_pb(v)
		this.m_data[int32(d.CampaignId)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerCampaignColumn)save( )(data []byte,err error){
	pb := &db.PlayerCampaignList{}
	pb.List=make([]*db.PlayerCampaign,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerCampaignColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerCampaignColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerCampaignColumn)GetAll()(list []dbPlayerCampaignData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerCampaignData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerCampaignColumn)Get(id int32)(v *dbPlayerCampaignData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerCampaignData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerCampaignColumn)Set(v dbPlayerCampaignData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.CampaignId)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.CampaignId)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerCampaignColumn)Add(v *dbPlayerCampaignData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.CampaignId)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.CampaignId)
		return false
	}
	d:=&dbPlayerCampaignData{}
	v.clone_to(d)
	this.m_data[int32(v.CampaignId)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerCampaignColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerCampaignColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerCampaignData)
	this.m_changed = true
	return
}
func (this *dbPlayerCampaignColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerCampaignColumn)GetStars(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignColumn.GetStars")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.Stars
	return v,true
}
func (this *dbPlayerCampaignColumn)SetStars(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignColumn.SetStars")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.Stars = v
	this.m_changed = true
	return true
}
func (this *dbPlayerCampaignColumn)GetLastFinishedTime(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignColumn.GetLastFinishedTime")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.LastFinishedTime
	return v,true
}
func (this *dbPlayerCampaignColumn)SetLastFinishedTime(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignColumn.SetLastFinishedTime")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.LastFinishedTime = v
	this.m_changed = true
	return true
}
func (this *dbPlayerCampaignColumn)GetTopScore(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignColumn.GetTopScore")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.TopScore
	return v,true
}
func (this *dbPlayerCampaignColumn)SetTopScore(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignColumn.SetTopScore")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.TopScore = v
	this.m_changed = true
	return true
}
func (this *dbPlayerCampaignColumn)GetPlayedCount(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignColumn.GetPlayedCount")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.PlayedCount
	return v,true
}
func (this *dbPlayerCampaignColumn)SetPlayedCount(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignColumn.SetPlayedCount")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.PlayedCount = v
	this.m_changed = true
	return true
}
func (this *dbPlayerCampaignColumn)IncbyPlayedCount(id int32,v int32)(r int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignColumn.IncbyPlayedCount")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		d = &dbPlayerCampaignData{}
		this.m_data[id] = d
	}
	d.PlayedCount +=  v
	this.m_changed = true
	return d.PlayedCount
}
func (this *dbPlayerCampaignColumn)GetPassCount(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerCampaignColumn.GetPassCount")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.PassCount
	return v,true
}
func (this *dbPlayerCampaignColumn)SetPassCount(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignColumn.SetPassCount")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.PassCount = v
	this.m_changed = true
	return true
}
func (this *dbPlayerCampaignColumn)IncbyPassCount(id int32,v int32)(r int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerCampaignColumn.IncbyPassCount")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		d = &dbPlayerCampaignData{}
		this.m_data[id] = d
	}
	d.PassCount +=  v
	this.m_changed = true
	return d.PassCount
}
type dbPlayerChapterUnLockColumn struct{
	m_row *dbPlayerRow
	m_data *dbPlayerChapterUnLockData
	m_changed bool
}
func (this *dbPlayerChapterUnLockColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_data = &dbPlayerChapterUnLockData{}
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerChapterUnLock{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_data = &dbPlayerChapterUnLockData{}
	this.m_data.from_pb(pb)
	this.m_changed = false
	return
}
func (this *dbPlayerChapterUnLockColumn)save( )(data []byte,err error){
	pb:=this.m_data.to_pb()
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerChapterUnLockColumn)Get( )(v *dbPlayerChapterUnLockData ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerChapterUnLockColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v=&dbPlayerChapterUnLockData{}
	this.m_data.clone_to(v)
	return
}
func (this *dbPlayerChapterUnLockColumn)Set(v dbPlayerChapterUnLockData ){
	this.m_row.m_lock.UnSafeLock("dbPlayerChapterUnLockColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=&dbPlayerChapterUnLockData{}
	v.clone_to(this.m_data)
	this.m_changed=true
	return
}
func (this *dbPlayerChapterUnLockColumn)GetChapterId( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerChapterUnLockColumn.GetChapterId")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.ChapterId
	return
}
func (this *dbPlayerChapterUnLockColumn)SetChapterId(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerChapterUnLockColumn.SetChapterId")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.ChapterId = v
	this.m_changed = true
	return
}
func (this *dbPlayerChapterUnLockColumn)GetPlayerIds( )(v []int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerChapterUnLockColumn.GetPlayerIds")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = make([]int32, len(this.m_data.PlayerIds))
	for _ii, _vv := range this.m_data.PlayerIds {
		v[_ii]=_vv
	}
	return
}
func (this *dbPlayerChapterUnLockColumn)SetPlayerIds(v []int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerChapterUnLockColumn.SetPlayerIds")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.PlayerIds = make([]int32, len(v))
	for _ii, _vv := range v {
		this.m_data.PlayerIds[_ii]=_vv
	}
	this.m_changed = true
	return
}
func (this *dbPlayerChapterUnLockColumn)GetCurHelpIds( )(v []int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerChapterUnLockColumn.GetCurHelpIds")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = make([]int32, len(this.m_data.CurHelpIds))
	for _ii, _vv := range this.m_data.CurHelpIds {
		v[_ii]=_vv
	}
	return
}
func (this *dbPlayerChapterUnLockColumn)SetCurHelpIds(v []int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerChapterUnLockColumn.SetCurHelpIds")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.CurHelpIds = make([]int32, len(v))
	for _ii, _vv := range v {
		this.m_data.CurHelpIds[_ii]=_vv
	}
	this.m_changed = true
	return
}
func (this *dbPlayerChapterUnLockColumn)GetStartUnix( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerChapterUnLockColumn.GetStartUnix")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.StartUnix
	return
}
func (this *dbPlayerChapterUnLockColumn)SetStartUnix(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerChapterUnLockColumn.SetStartUnix")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.StartUnix = v
	this.m_changed = true
	return
}
type dbPlayerShopItemColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerShopItemData
	m_changed bool
}
func (this *dbPlayerShopItemColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerShopItemList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerShopItemData{}
		d.from_pb(v)
		this.m_data[int32(d.Id)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerShopItemColumn)save( )(data []byte,err error){
	pb := &db.PlayerShopItemList{}
	pb.List=make([]*db.PlayerShopItem,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerShopItemColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerShopItemColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerShopItemColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerShopItemColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerShopItemColumn)GetAll()(list []dbPlayerShopItemData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerShopItemColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerShopItemData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerShopItemColumn)Get(id int32)(v *dbPlayerShopItemData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerShopItemColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerShopItemData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerShopItemColumn)Set(v dbPlayerShopItemData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerShopItemColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.Id)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.Id)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerShopItemColumn)Add(v *dbPlayerShopItemData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerShopItemColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.Id)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.Id)
		return false
	}
	d:=&dbPlayerShopItemData{}
	v.clone_to(d)
	this.m_data[int32(v.Id)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerShopItemColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerShopItemColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerShopItemColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerShopItemColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerShopItemData)
	this.m_changed = true
	return
}
func (this *dbPlayerShopItemColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerShopItemColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerShopItemColumn)GetLeftNum(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerShopItemColumn.GetLeftNum")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.LeftNum
	return v,true
}
func (this *dbPlayerShopItemColumn)SetLeftNum(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerShopItemColumn.SetLeftNum")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.LeftNum = v
	this.m_changed = true
	return true
}
func (this *dbPlayerShopItemColumn)IncbyLeftNum(id int32,v int32)(r int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerShopItemColumn.IncbyLeftNum")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		d = &dbPlayerShopItemData{}
		this.m_data[id] = d
	}
	d.LeftNum +=  v
	this.m_changed = true
	return d.LeftNum
}
type dbPlayerShopLimitedInfoColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerShopLimitedInfoData
	m_changed bool
}
func (this *dbPlayerShopLimitedInfoColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerShopLimitedInfoList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerShopLimitedInfoData{}
		d.from_pb(v)
		this.m_data[int32(d.LimitedDays)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerShopLimitedInfoColumn)save( )(data []byte,err error){
	pb := &db.PlayerShopLimitedInfoList{}
	pb.List=make([]*db.PlayerShopLimitedInfo,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerShopLimitedInfoColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerShopLimitedInfoColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerShopLimitedInfoColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerShopLimitedInfoColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerShopLimitedInfoColumn)GetAll()(list []dbPlayerShopLimitedInfoData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerShopLimitedInfoColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerShopLimitedInfoData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerShopLimitedInfoColumn)Get(id int32)(v *dbPlayerShopLimitedInfoData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerShopLimitedInfoColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerShopLimitedInfoData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerShopLimitedInfoColumn)Set(v dbPlayerShopLimitedInfoData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerShopLimitedInfoColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.LimitedDays)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.LimitedDays)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerShopLimitedInfoColumn)Add(v *dbPlayerShopLimitedInfoData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerShopLimitedInfoColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.LimitedDays)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.LimitedDays)
		return false
	}
	d:=&dbPlayerShopLimitedInfoData{}
	v.clone_to(d)
	this.m_data[int32(v.LimitedDays)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerShopLimitedInfoColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerShopLimitedInfoColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerShopLimitedInfoColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerShopLimitedInfoColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerShopLimitedInfoData)
	this.m_changed = true
	return
}
func (this *dbPlayerShopLimitedInfoColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerShopLimitedInfoColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerShopLimitedInfoColumn)GetLastSaveTime(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerShopLimitedInfoColumn.GetLastSaveTime")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.LastSaveTime
	return v,true
}
func (this *dbPlayerShopLimitedInfoColumn)SetLastSaveTime(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerShopLimitedInfoColumn.SetLastSaveTime")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.LastSaveTime = v
	this.m_changed = true
	return true
}
type dbPlayerChestColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerChestData
	m_changed bool
}
func (this *dbPlayerChestColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerChestList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerChestData{}
		d.from_pb(v)
		this.m_data[int32(d.Pos)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerChestColumn)save( )(data []byte,err error){
	pb := &db.PlayerChestList{}
	pb.List=make([]*db.PlayerChest,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerChestColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerChestColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerChestColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerChestColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerChestColumn)GetAll()(list []dbPlayerChestData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerChestColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerChestData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerChestColumn)Get(id int32)(v *dbPlayerChestData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerChestColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerChestData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerChestColumn)Set(v dbPlayerChestData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerChestColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.Pos)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.Pos)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerChestColumn)Add(v *dbPlayerChestData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerChestColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.Pos)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.Pos)
		return false
	}
	d:=&dbPlayerChestData{}
	v.clone_to(d)
	this.m_data[int32(v.Pos)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerChestColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerChestColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerChestColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerChestColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerChestData)
	this.m_changed = true
	return
}
func (this *dbPlayerChestColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerChestColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerChestColumn)GetChestId(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerChestColumn.GetChestId")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.ChestId
	return v,true
}
func (this *dbPlayerChestColumn)SetChestId(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerChestColumn.SetChestId")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.ChestId = v
	this.m_changed = true
	return true
}
func (this *dbPlayerChestColumn)GetOpenSec(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerChestColumn.GetOpenSec")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.OpenSec
	return v,true
}
func (this *dbPlayerChestColumn)SetOpenSec(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerChestColumn.SetOpenSec")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.OpenSec = v
	this.m_changed = true
	return true
}
type dbPlayerMailColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerMailData
	m_changed bool
}
func (this *dbPlayerMailColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerMailList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerMailData{}
		d.from_pb(v)
		this.m_data[int32(d.MailId)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerMailColumn)save( )(data []byte,err error){
	pb := &db.PlayerMailList{}
	pb.List=make([]*db.PlayerMail,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerMailColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerMailColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerMailColumn)GetAll()(list []dbPlayerMailData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerMailData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerMailColumn)Get(id int32)(v *dbPlayerMailData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerMailData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerMailColumn)Set(v dbPlayerMailData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.MailId)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.MailId)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerMailColumn)Add(v *dbPlayerMailData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.MailId)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.MailId)
		return false
	}
	d:=&dbPlayerMailData{}
	v.clone_to(d)
	this.m_data[int32(v.MailId)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerMailColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerMailColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerMailData)
	this.m_changed = true
	return
}
func (this *dbPlayerMailColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerMailColumn)GetMailType(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.GetMailType")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = int32(d.MailType)
	return v,true
}
func (this *dbPlayerMailColumn)SetMailType(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.SetMailType")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.MailType = int8(v)
	this.m_changed = true
	return true
}
func (this *dbPlayerMailColumn)GetMailTitle(id int32)(v string ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.GetMailTitle")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.MailTitle
	return v,true
}
func (this *dbPlayerMailColumn)SetMailTitle(id int32,v string)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.SetMailTitle")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.MailTitle = v
	this.m_changed = true
	return true
}
func (this *dbPlayerMailColumn)GetSenderId(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.GetSenderId")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.SenderId
	return v,true
}
func (this *dbPlayerMailColumn)SetSenderId(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.SetSenderId")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.SenderId = v
	this.m_changed = true
	return true
}
func (this *dbPlayerMailColumn)GetSenderName(id int32)(v string ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.GetSenderName")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.SenderName
	return v,true
}
func (this *dbPlayerMailColumn)SetSenderName(id int32,v string)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.SetSenderName")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.SenderName = v
	this.m_changed = true
	return true
}
func (this *dbPlayerMailColumn)GetContent(id int32)(v string ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.GetContent")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.Content
	return v,true
}
func (this *dbPlayerMailColumn)SetContent(id int32,v string)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.SetContent")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.Content = v
	this.m_changed = true
	return true
}
func (this *dbPlayerMailColumn)GetSendUnix(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.GetSendUnix")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.SendUnix
	return v,true
}
func (this *dbPlayerMailColumn)SetSendUnix(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.SetSendUnix")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.SendUnix = v
	this.m_changed = true
	return true
}
func (this *dbPlayerMailColumn)GetOverUnix(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.GetOverUnix")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.OverUnix
	return v,true
}
func (this *dbPlayerMailColumn)SetOverUnix(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.SetOverUnix")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.OverUnix = v
	this.m_changed = true
	return true
}
func (this *dbPlayerMailColumn)GetObjIds(id int32)(v []int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.GetObjIds")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = make([]int32, len(d.ObjIds))
	for _ii, _vv := range d.ObjIds {
		v[_ii]=_vv
	}
	return v,true
}
func (this *dbPlayerMailColumn)SetObjIds(id int32,v []int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.SetObjIds")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.ObjIds = make([]int32, len(v))
	for _ii, _vv := range v {
		d.ObjIds[_ii]=_vv
	}
	this.m_changed = true
	return true
}
func (this *dbPlayerMailColumn)GetObjNums(id int32)(v []int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.GetObjNums")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = make([]int32, len(d.ObjNums))
	for _ii, _vv := range d.ObjNums {
		v[_ii]=_vv
	}
	return v,true
}
func (this *dbPlayerMailColumn)SetObjNums(id int32,v []int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.SetObjNums")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.ObjNums = make([]int32, len(v))
	for _ii, _vv := range v {
		d.ObjNums[_ii]=_vv
	}
	this.m_changed = true
	return true
}
func (this *dbPlayerMailColumn)GetState(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.GetState")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = int32(d.State)
	return v,true
}
func (this *dbPlayerMailColumn)SetState(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.SetState")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.State = int8(v)
	this.m_changed = true
	return true
}
func (this *dbPlayerMailColumn)GetExtraDatas(id int32)(v []int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerMailColumn.GetExtraDatas")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = make([]int32, len(d.ExtraDatas))
	for _ii, _vv := range d.ExtraDatas {
		v[_ii]=_vv
	}
	return v,true
}
func (this *dbPlayerMailColumn)SetExtraDatas(id int32,v []int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerMailColumn.SetExtraDatas")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.ExtraDatas = make([]int32, len(v))
	for _ii, _vv := range v {
		d.ExtraDatas[_ii]=_vv
	}
	this.m_changed = true
	return true
}
type dbPlayerDialyTaskColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerDialyTaskData
	m_changed bool
}
func (this *dbPlayerDialyTaskColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerDialyTaskList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerDialyTaskData{}
		d.from_pb(v)
		this.m_data[int32(d.TaskId)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerDialyTaskColumn)save( )(data []byte,err error){
	pb := &db.PlayerDialyTaskList{}
	pb.List=make([]*db.PlayerDialyTask,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerDialyTaskColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerDialyTaskColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerDialyTaskColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerDialyTaskColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerDialyTaskColumn)GetAll()(list []dbPlayerDialyTaskData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerDialyTaskColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerDialyTaskData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerDialyTaskColumn)Get(id int32)(v *dbPlayerDialyTaskData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerDialyTaskColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerDialyTaskData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerDialyTaskColumn)Set(v dbPlayerDialyTaskData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerDialyTaskColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.TaskId)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.TaskId)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerDialyTaskColumn)Add(v *dbPlayerDialyTaskData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerDialyTaskColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.TaskId)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.TaskId)
		return false
	}
	d:=&dbPlayerDialyTaskData{}
	v.clone_to(d)
	this.m_data[int32(v.TaskId)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerDialyTaskColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerDialyTaskColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerDialyTaskColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerDialyTaskColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerDialyTaskData)
	this.m_changed = true
	return
}
func (this *dbPlayerDialyTaskColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerDialyTaskColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerDialyTaskColumn)GetValue(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerDialyTaskColumn.GetValue")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.Value
	return v,true
}
func (this *dbPlayerDialyTaskColumn)SetValue(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerDialyTaskColumn.SetValue")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.Value = v
	this.m_changed = true
	return true
}
func (this *dbPlayerDialyTaskColumn)IncbyValue(id int32,v int32)(r int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerDialyTaskColumn.IncbyValue")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		d = &dbPlayerDialyTaskData{}
		this.m_data[id] = d
	}
	d.Value +=  v
	this.m_changed = true
	return d.Value
}
func (this *dbPlayerDialyTaskColumn)GetRewardUnix(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerDialyTaskColumn.GetRewardUnix")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.RewardUnix
	return v,true
}
func (this *dbPlayerDialyTaskColumn)SetRewardUnix(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerDialyTaskColumn.SetRewardUnix")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.RewardUnix = v
	this.m_changed = true
	return true
}
type dbPlayerAchieveColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerAchieveData
	m_changed bool
}
func (this *dbPlayerAchieveColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerAchieveList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerAchieveData{}
		d.from_pb(v)
		this.m_data[int32(d.AchieveId)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerAchieveColumn)save( )(data []byte,err error){
	pb := &db.PlayerAchieveList{}
	pb.List=make([]*db.PlayerAchieve,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerAchieveColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerAchieveColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerAchieveColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerAchieveColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerAchieveColumn)GetAll()(list []dbPlayerAchieveData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerAchieveColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerAchieveData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerAchieveColumn)Get(id int32)(v *dbPlayerAchieveData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerAchieveColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerAchieveData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerAchieveColumn)Set(v dbPlayerAchieveData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerAchieveColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.AchieveId)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.AchieveId)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerAchieveColumn)Add(v *dbPlayerAchieveData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerAchieveColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.AchieveId)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.AchieveId)
		return false
	}
	d:=&dbPlayerAchieveData{}
	v.clone_to(d)
	this.m_data[int32(v.AchieveId)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerAchieveColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerAchieveColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerAchieveColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerAchieveColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerAchieveData)
	this.m_changed = true
	return
}
func (this *dbPlayerAchieveColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerAchieveColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerAchieveColumn)GetValue(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerAchieveColumn.GetValue")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.Value
	return v,true
}
func (this *dbPlayerAchieveColumn)SetValue(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerAchieveColumn.SetValue")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.Value = v
	this.m_changed = true
	return true
}
func (this *dbPlayerAchieveColumn)IncbyValue(id int32,v int32)(r int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerAchieveColumn.IncbyValue")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		d = &dbPlayerAchieveData{}
		this.m_data[id] = d
	}
	d.Value +=  v
	this.m_changed = true
	return d.Value
}
func (this *dbPlayerAchieveColumn)GetRewardUnix(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerAchieveColumn.GetRewardUnix")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.RewardUnix
	return v,true
}
func (this *dbPlayerAchieveColumn)SetRewardUnix(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerAchieveColumn.SetRewardUnix")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.RewardUnix = v
	this.m_changed = true
	return true
}
type dbPlayerFinishedAchieveColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerFinishedAchieveData
	m_changed bool
}
func (this *dbPlayerFinishedAchieveColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerFinishedAchieveList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerFinishedAchieveData{}
		d.from_pb(v)
		this.m_data[int32(d.AchieveId)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFinishedAchieveColumn)save( )(data []byte,err error){
	pb := &db.PlayerFinishedAchieveList{}
	pb.List=make([]*db.PlayerFinishedAchieve,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFinishedAchieveColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFinishedAchieveColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerFinishedAchieveColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFinishedAchieveColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerFinishedAchieveColumn)GetAll()(list []dbPlayerFinishedAchieveData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFinishedAchieveColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerFinishedAchieveData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerFinishedAchieveColumn)Get(id int32)(v *dbPlayerFinishedAchieveData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFinishedAchieveColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerFinishedAchieveData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerFinishedAchieveColumn)Set(v dbPlayerFinishedAchieveData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFinishedAchieveColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.AchieveId)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.AchieveId)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerFinishedAchieveColumn)Add(v *dbPlayerFinishedAchieveData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFinishedAchieveColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.AchieveId)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.AchieveId)
		return false
	}
	d:=&dbPlayerFinishedAchieveData{}
	v.clone_to(d)
	this.m_data[int32(v.AchieveId)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerFinishedAchieveColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerFinishedAchieveColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerFinishedAchieveColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerFinishedAchieveColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerFinishedAchieveData)
	this.m_changed = true
	return
}
func (this *dbPlayerFinishedAchieveColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFinishedAchieveColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
type dbPlayerDailyTaskWholeDailyColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerDailyTaskWholeDailyData
	m_changed bool
}
func (this *dbPlayerDailyTaskWholeDailyColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerDailyTaskWholeDailyList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerDailyTaskWholeDailyData{}
		d.from_pb(v)
		this.m_data[int32(d.CompleteTaskId)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerDailyTaskWholeDailyColumn)save( )(data []byte,err error){
	pb := &db.PlayerDailyTaskWholeDailyList{}
	pb.List=make([]*db.PlayerDailyTaskWholeDaily,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerDailyTaskWholeDailyColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerDailyTaskWholeDailyColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerDailyTaskWholeDailyColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerDailyTaskWholeDailyColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerDailyTaskWholeDailyColumn)GetAll()(list []dbPlayerDailyTaskWholeDailyData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerDailyTaskWholeDailyColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerDailyTaskWholeDailyData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerDailyTaskWholeDailyColumn)Get(id int32)(v *dbPlayerDailyTaskWholeDailyData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerDailyTaskWholeDailyColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerDailyTaskWholeDailyData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerDailyTaskWholeDailyColumn)Set(v dbPlayerDailyTaskWholeDailyData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerDailyTaskWholeDailyColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.CompleteTaskId)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.CompleteTaskId)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerDailyTaskWholeDailyColumn)Add(v *dbPlayerDailyTaskWholeDailyData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerDailyTaskWholeDailyColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.CompleteTaskId)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.CompleteTaskId)
		return false
	}
	d:=&dbPlayerDailyTaskWholeDailyData{}
	v.clone_to(d)
	this.m_data[int32(v.CompleteTaskId)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerDailyTaskWholeDailyColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerDailyTaskWholeDailyColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerDailyTaskWholeDailyColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerDailyTaskWholeDailyColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerDailyTaskWholeDailyData)
	this.m_changed = true
	return
}
func (this *dbPlayerDailyTaskWholeDailyColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerDailyTaskWholeDailyColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
type dbPlayerSevenActivityColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerSevenActivityData
	m_changed bool
}
func (this *dbPlayerSevenActivityColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerSevenActivityList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerSevenActivityData{}
		d.from_pb(v)
		this.m_data[int32(d.ActivityId)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerSevenActivityColumn)save( )(data []byte,err error){
	pb := &db.PlayerSevenActivityList{}
	pb.List=make([]*db.PlayerSevenActivity,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerSevenActivityColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerSevenActivityColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerSevenActivityColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerSevenActivityColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerSevenActivityColumn)GetAll()(list []dbPlayerSevenActivityData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerSevenActivityColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerSevenActivityData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerSevenActivityColumn)Get(id int32)(v *dbPlayerSevenActivityData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerSevenActivityColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerSevenActivityData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerSevenActivityColumn)Set(v dbPlayerSevenActivityData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerSevenActivityColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.ActivityId)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.ActivityId)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerSevenActivityColumn)Add(v *dbPlayerSevenActivityData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerSevenActivityColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.ActivityId)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.ActivityId)
		return false
	}
	d:=&dbPlayerSevenActivityData{}
	v.clone_to(d)
	this.m_data[int32(v.ActivityId)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerSevenActivityColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerSevenActivityColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerSevenActivityColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerSevenActivityColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerSevenActivityData)
	this.m_changed = true
	return
}
func (this *dbPlayerSevenActivityColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerSevenActivityColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerSevenActivityColumn)GetValue(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerSevenActivityColumn.GetValue")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.Value
	return v,true
}
func (this *dbPlayerSevenActivityColumn)SetValue(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerSevenActivityColumn.SetValue")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.Value = v
	this.m_changed = true
	return true
}
func (this *dbPlayerSevenActivityColumn)IncbyValue(id int32,v int32)(r int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerSevenActivityColumn.IncbyValue")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		d = &dbPlayerSevenActivityData{}
		this.m_data[id] = d
	}
	d.Value +=  v
	this.m_changed = true
	return d.Value
}
func (this *dbPlayerSevenActivityColumn)GetRewardUnix(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerSevenActivityColumn.GetRewardUnix")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.RewardUnix
	return v,true
}
func (this *dbPlayerSevenActivityColumn)SetRewardUnix(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerSevenActivityColumn.SetRewardUnix")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.RewardUnix = v
	this.m_changed = true
	return true
}
type dbPlayerGuidesColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerGuidesData
	m_changed bool
}
func (this *dbPlayerGuidesColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerGuidesList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerGuidesData{}
		d.from_pb(v)
		this.m_data[int32(d.GuideId)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerGuidesColumn)save( )(data []byte,err error){
	pb := &db.PlayerGuidesList{}
	pb.List=make([]*db.PlayerGuides,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerGuidesColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerGuidesColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerGuidesColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerGuidesColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerGuidesColumn)GetAll()(list []dbPlayerGuidesData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerGuidesColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerGuidesData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerGuidesColumn)Get(id int32)(v *dbPlayerGuidesData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerGuidesColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerGuidesData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerGuidesColumn)Set(v dbPlayerGuidesData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerGuidesColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.GuideId)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.GuideId)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerGuidesColumn)Add(v *dbPlayerGuidesData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerGuidesColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.GuideId)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.GuideId)
		return false
	}
	d:=&dbPlayerGuidesData{}
	v.clone_to(d)
	this.m_data[int32(v.GuideId)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerGuidesColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerGuidesColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerGuidesColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerGuidesColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerGuidesData)
	this.m_changed = true
	return
}
func (this *dbPlayerGuidesColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerGuidesColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerGuidesColumn)GetSetUnix(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerGuidesColumn.GetSetUnix")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.SetUnix
	return v,true
}
func (this *dbPlayerGuidesColumn)SetSetUnix(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerGuidesColumn.SetSetUnix")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.SetUnix = v
	this.m_changed = true
	return true
}
type dbPlayerFriendRelativeColumn struct{
	m_row *dbPlayerRow
	m_data *dbPlayerFriendRelativeData
	m_changed bool
}
func (this *dbPlayerFriendRelativeColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_data = &dbPlayerFriendRelativeData{}
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerFriendRelative{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_data = &dbPlayerFriendRelativeData{}
	this.m_data.from_pb(pb)
	this.m_changed = false
	return
}
func (this *dbPlayerFriendRelativeColumn)save( )(data []byte,err error){
	pb:=this.m_data.to_pb()
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFriendRelativeColumn)Get( )(v *dbPlayerFriendRelativeData ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendRelativeColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v=&dbPlayerFriendRelativeData{}
	this.m_data.clone_to(v)
	return
}
func (this *dbPlayerFriendRelativeColumn)Set(v dbPlayerFriendRelativeData ){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendRelativeColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=&dbPlayerFriendRelativeData{}
	v.clone_to(this.m_data)
	this.m_changed=true
	return
}
func (this *dbPlayerFriendRelativeColumn)GetLastGiveFriendPointsTime( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendRelativeColumn.GetLastGiveFriendPointsTime")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.LastGiveFriendPointsTime
	return
}
func (this *dbPlayerFriendRelativeColumn)SetLastGiveFriendPointsTime(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendRelativeColumn.SetLastGiveFriendPointsTime")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.LastGiveFriendPointsTime = v
	this.m_changed = true
	return
}
func (this *dbPlayerFriendRelativeColumn)GetGiveNumToday( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendRelativeColumn.GetGiveNumToday")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.GiveNumToday
	return
}
func (this *dbPlayerFriendRelativeColumn)SetGiveNumToday(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendRelativeColumn.SetGiveNumToday")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.GiveNumToday = v
	this.m_changed = true
	return
}
func (this *dbPlayerFriendRelativeColumn)IncbyGiveNumToday(v int32)(r int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendRelativeColumn.IncbyGiveNumToday")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.GiveNumToday += v
	this.m_changed = true
	return this.m_data.GiveNumToday
}
func (this *dbPlayerFriendRelativeColumn)GetLastRefreshTime( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendRelativeColumn.GetLastRefreshTime")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.LastRefreshTime
	return
}
func (this *dbPlayerFriendRelativeColumn)SetLastRefreshTime(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendRelativeColumn.SetLastRefreshTime")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.LastRefreshTime = v
	this.m_changed = true
	return
}
type dbPlayerFriendColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerFriendData
	m_changed bool
}
func (this *dbPlayerFriendColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerFriendList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerFriendData{}
		d.from_pb(v)
		this.m_data[int32(d.FriendId)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFriendColumn)save( )(data []byte,err error){
	pb := &db.PlayerFriendList{}
	pb.List=make([]*db.PlayerFriend,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFriendColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerFriendColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerFriendColumn)GetAll()(list []dbPlayerFriendData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerFriendData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerFriendColumn)Get(id int32)(v *dbPlayerFriendData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerFriendData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerFriendColumn)Set(v dbPlayerFriendData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.FriendId)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.FriendId)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendColumn)Add(v *dbPlayerFriendData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.FriendId)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.FriendId)
		return false
	}
	d:=&dbPlayerFriendData{}
	v.clone_to(d)
	this.m_data[int32(v.FriendId)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerFriendColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerFriendData)
	this.m_changed = true
	return
}
func (this *dbPlayerFriendColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerFriendColumn)GetFriendName(id int32)(v string ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.GetFriendName")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.FriendName
	return v,true
}
func (this *dbPlayerFriendColumn)SetFriendName(id int32,v string)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendColumn.SetFriendName")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.FriendName = v
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendColumn)GetHead(id int32)(v string ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.GetHead")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.Head
	return v,true
}
func (this *dbPlayerFriendColumn)SetHead(id int32,v string)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendColumn.SetHead")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.Head = v
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendColumn)GetLevel(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.GetLevel")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.Level
	return v,true
}
func (this *dbPlayerFriendColumn)SetLevel(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendColumn.SetLevel")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.Level = v
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendColumn)GetVipLevel(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.GetVipLevel")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.VipLevel
	return v,true
}
func (this *dbPlayerFriendColumn)SetVipLevel(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendColumn.SetVipLevel")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.VipLevel = v
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendColumn)GetLastLogin(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.GetLastLogin")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.LastLogin
	return v,true
}
func (this *dbPlayerFriendColumn)SetLastLogin(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendColumn.SetLastLogin")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.LastLogin = v
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendColumn)GetLastGivePointsTime(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendColumn.GetLastGivePointsTime")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.LastGivePointsTime
	return v,true
}
func (this *dbPlayerFriendColumn)SetLastGivePointsTime(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendColumn.SetLastGivePointsTime")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.LastGivePointsTime = v
	this.m_changed = true
	return true
}
type dbPlayerFriendReqColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerFriendReqData
	m_changed bool
}
func (this *dbPlayerFriendReqColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerFriendReqList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerFriendReqData{}
		d.from_pb(v)
		this.m_data[int32(d.PlayerId)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFriendReqColumn)save( )(data []byte,err error){
	pb := &db.PlayerFriendReqList{}
	pb.List=make([]*db.PlayerFriendReq,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFriendReqColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendReqColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerFriendReqColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendReqColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerFriendReqColumn)GetAll()(list []dbPlayerFriendReqData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendReqColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerFriendReqData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerFriendReqColumn)Get(id int32)(v *dbPlayerFriendReqData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendReqColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerFriendReqData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerFriendReqColumn)Set(v dbPlayerFriendReqData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendReqColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.PlayerId)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.PlayerId)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendReqColumn)Add(v *dbPlayerFriendReqData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendReqColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.PlayerId)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.PlayerId)
		return false
	}
	d:=&dbPlayerFriendReqData{}
	v.clone_to(d)
	this.m_data[int32(v.PlayerId)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendReqColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendReqColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerFriendReqColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendReqColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerFriendReqData)
	this.m_changed = true
	return
}
func (this *dbPlayerFriendReqColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendReqColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerFriendReqColumn)GetPlayerName(id int32)(v string ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendReqColumn.GetPlayerName")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.PlayerName
	return v,true
}
func (this *dbPlayerFriendReqColumn)SetPlayerName(id int32,v string)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendReqColumn.SetPlayerName")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.PlayerName = v
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendReqColumn)GetReqUnix(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendReqColumn.GetReqUnix")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.ReqUnix
	return v,true
}
func (this *dbPlayerFriendReqColumn)SetReqUnix(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendReqColumn.SetReqUnix")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.ReqUnix = v
	this.m_changed = true
	return true
}
type dbPlayerFriendPointColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerFriendPointData
	m_changed bool
}
func (this *dbPlayerFriendPointColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerFriendPointList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerFriendPointData{}
		d.from_pb(v)
		this.m_data[int32(d.FromPlayerId)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFriendPointColumn)save( )(data []byte,err error){
	pb := &db.PlayerFriendPointList{}
	pb.List=make([]*db.PlayerFriendPoint,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFriendPointColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendPointColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerFriendPointColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendPointColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerFriendPointColumn)GetAll()(list []dbPlayerFriendPointData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendPointColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerFriendPointData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerFriendPointColumn)Get(id int32)(v *dbPlayerFriendPointData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendPointColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerFriendPointData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerFriendPointColumn)Set(v dbPlayerFriendPointData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendPointColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.FromPlayerId)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.FromPlayerId)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendPointColumn)Add(v *dbPlayerFriendPointData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendPointColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.FromPlayerId)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.FromPlayerId)
		return false
	}
	d:=&dbPlayerFriendPointData{}
	v.clone_to(d)
	this.m_data[int32(v.FromPlayerId)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendPointColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendPointColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerFriendPointColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendPointColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerFriendPointData)
	this.m_changed = true
	return
}
func (this *dbPlayerFriendPointColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendPointColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerFriendPointColumn)GetGivePoints(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendPointColumn.GetGivePoints")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.GivePoints
	return v,true
}
func (this *dbPlayerFriendPointColumn)SetGivePoints(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendPointColumn.SetGivePoints")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.GivePoints = v
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendPointColumn)GetLastGiveTime(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendPointColumn.GetLastGiveTime")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.LastGiveTime
	return v,true
}
func (this *dbPlayerFriendPointColumn)SetLastGiveTime(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendPointColumn.SetLastGiveTime")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.LastGiveTime = v
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendPointColumn)GetIsTodayGive(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendPointColumn.GetIsTodayGive")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.IsTodayGive
	return v,true
}
func (this *dbPlayerFriendPointColumn)SetIsTodayGive(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendPointColumn.SetIsTodayGive")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.IsTodayGive = v
	this.m_changed = true
	return true
}
type dbPlayerFriendChatUnreadIdColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerFriendChatUnreadIdData
	m_changed bool
}
func (this *dbPlayerFriendChatUnreadIdColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerFriendChatUnreadIdList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerFriendChatUnreadIdData{}
		d.from_pb(v)
		this.m_data[int32(d.FriendId)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFriendChatUnreadIdColumn)save( )(data []byte,err error){
	pb := &db.PlayerFriendChatUnreadIdList{}
	pb.List=make([]*db.PlayerFriendChatUnreadId,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFriendChatUnreadIdColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadIdColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerFriendChatUnreadIdColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadIdColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerFriendChatUnreadIdColumn)GetAll()(list []dbPlayerFriendChatUnreadIdData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadIdColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerFriendChatUnreadIdData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerFriendChatUnreadIdColumn)Get(id int32)(v *dbPlayerFriendChatUnreadIdData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadIdColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerFriendChatUnreadIdData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerFriendChatUnreadIdColumn)Set(v dbPlayerFriendChatUnreadIdData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadIdColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.FriendId)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.FriendId)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendChatUnreadIdColumn)Add(v *dbPlayerFriendChatUnreadIdData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadIdColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.FriendId)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.FriendId)
		return false
	}
	d:=&dbPlayerFriendChatUnreadIdData{}
	v.clone_to(d)
	this.m_data[int32(v.FriendId)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendChatUnreadIdColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadIdColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerFriendChatUnreadIdColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadIdColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerFriendChatUnreadIdData)
	this.m_changed = true
	return
}
func (this *dbPlayerFriendChatUnreadIdColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadIdColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerFriendChatUnreadIdColumn)GetMessageIds(id int32)(v []int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadIdColumn.GetMessageIds")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = make([]int32, len(d.MessageIds))
	for _ii, _vv := range d.MessageIds {
		v[_ii]=_vv
	}
	return v,true
}
func (this *dbPlayerFriendChatUnreadIdColumn)SetMessageIds(id int32,v []int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadIdColumn.SetMessageIds")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.MessageIds = make([]int32, len(v))
	for _ii, _vv := range v {
		d.MessageIds[_ii]=_vv
	}
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendChatUnreadIdColumn)GetCurrMessageId(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadIdColumn.GetCurrMessageId")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.CurrMessageId
	return v,true
}
func (this *dbPlayerFriendChatUnreadIdColumn)SetCurrMessageId(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadIdColumn.SetCurrMessageId")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.CurrMessageId = v
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendChatUnreadIdColumn)IncbyCurrMessageId(id int32,v int32)(r int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadIdColumn.IncbyCurrMessageId")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		d = &dbPlayerFriendChatUnreadIdData{}
		this.m_data[id] = d
	}
	d.CurrMessageId +=  v
	this.m_changed = true
	return d.CurrMessageId
}
type dbPlayerFriendChatUnreadMessageColumn struct{
	m_row *dbPlayerRow
	m_data map[int64]*dbPlayerFriendChatUnreadMessageData
	m_changed bool
}
func (this *dbPlayerFriendChatUnreadMessageColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerFriendChatUnreadMessageList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerFriendChatUnreadMessageData{}
		d.from_pb(v)
		this.m_data[int64(d.PlayerMessageId)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFriendChatUnreadMessageColumn)save( )(data []byte,err error){
	pb := &db.PlayerFriendChatUnreadMessageList{}
	pb.List=make([]*db.PlayerFriendChatUnreadMessage,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFriendChatUnreadMessageColumn)HasIndex(id int64)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadMessageColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerFriendChatUnreadMessageColumn)GetAllIndex()(list []int64){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadMessageColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int64, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerFriendChatUnreadMessageColumn)GetAll()(list []dbPlayerFriendChatUnreadMessageData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadMessageColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerFriendChatUnreadMessageData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerFriendChatUnreadMessageColumn)Get(id int64)(v *dbPlayerFriendChatUnreadMessageData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadMessageColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerFriendChatUnreadMessageData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerFriendChatUnreadMessageColumn)Set(v dbPlayerFriendChatUnreadMessageData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadMessageColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int64(v.PlayerMessageId)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.PlayerMessageId)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendChatUnreadMessageColumn)Add(v *dbPlayerFriendChatUnreadMessageData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadMessageColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int64(v.PlayerMessageId)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.PlayerMessageId)
		return false
	}
	d:=&dbPlayerFriendChatUnreadMessageData{}
	v.clone_to(d)
	this.m_data[int64(v.PlayerMessageId)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendChatUnreadMessageColumn)Remove(id int64){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadMessageColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerFriendChatUnreadMessageColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadMessageColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int64]*dbPlayerFriendChatUnreadMessageData)
	this.m_changed = true
	return
}
func (this *dbPlayerFriendChatUnreadMessageColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadMessageColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerFriendChatUnreadMessageColumn)GetMessage(id int64)(v []byte,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadMessageColumn.GetMessage")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = make([]byte, len(d.Message))
	for _ii, _vv := range d.Message {
		v[_ii]=_vv
	}
	return v,true
}
func (this *dbPlayerFriendChatUnreadMessageColumn)SetMessage(id int64,v []byte)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadMessageColumn.SetMessage")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.Message = make([]byte, len(v))
	for _ii, _vv := range v {
		d.Message[_ii]=_vv
	}
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendChatUnreadMessageColumn)GetSendTime(id int64)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadMessageColumn.GetSendTime")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.SendTime
	return v,true
}
func (this *dbPlayerFriendChatUnreadMessageColumn)SetSendTime(id int64,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadMessageColumn.SetSendTime")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.SendTime = v
	this.m_changed = true
	return true
}
func (this *dbPlayerFriendChatUnreadMessageColumn)GetIsRead(id int64)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFriendChatUnreadMessageColumn.GetIsRead")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.IsRead
	return v,true
}
func (this *dbPlayerFriendChatUnreadMessageColumn)SetIsRead(id int64,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFriendChatUnreadMessageColumn.SetIsRead")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.IsRead = v
	this.m_changed = true
	return true
}
type dbPlayerChaterOpenRequestColumn struct{
	m_row *dbPlayerRow
	m_data *dbPlayerChaterOpenRequestData
	m_changed bool
}
func (this *dbPlayerChaterOpenRequestColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_data = &dbPlayerChaterOpenRequestData{}
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerChaterOpenRequest{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_data = &dbPlayerChaterOpenRequestData{}
	this.m_data.from_pb(pb)
	this.m_changed = false
	return
}
func (this *dbPlayerChaterOpenRequestColumn)save( )(data []byte,err error){
	pb:=this.m_data.to_pb()
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerChaterOpenRequestColumn)Get( )(v *dbPlayerChaterOpenRequestData ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerChaterOpenRequestColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v=&dbPlayerChaterOpenRequestData{}
	this.m_data.clone_to(v)
	return
}
func (this *dbPlayerChaterOpenRequestColumn)Set(v dbPlayerChaterOpenRequestData ){
	this.m_row.m_lock.UnSafeLock("dbPlayerChaterOpenRequestColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=&dbPlayerChaterOpenRequestData{}
	v.clone_to(this.m_data)
	this.m_changed=true
	return
}
func (this *dbPlayerChaterOpenRequestColumn)GetCustomData( )(v []byte){
	this.m_row.m_lock.UnSafeRLock("dbPlayerChaterOpenRequestColumn.GetCustomData")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = make([]byte, len(this.m_data.CustomData))
	for _ii, _vv := range this.m_data.CustomData {
		v[_ii]=_vv
	}
	return
}
func (this *dbPlayerChaterOpenRequestColumn)SetCustomData(v []byte){
	this.m_row.m_lock.UnSafeLock("dbPlayerChaterOpenRequestColumn.SetCustomData")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.CustomData = make([]byte, len(v))
	for _ii, _vv := range v {
		this.m_data.CustomData[_ii]=_vv
	}
	this.m_changed = true
	return
}
type dbPlayerHandbookItemColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerHandbookItemData
	m_changed bool
}
func (this *dbPlayerHandbookItemColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerHandbookItemList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerHandbookItemData{}
		d.from_pb(v)
		this.m_data[int32(d.Id)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerHandbookItemColumn)save( )(data []byte,err error){
	pb := &db.PlayerHandbookItemList{}
	pb.List=make([]*db.PlayerHandbookItem,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerHandbookItemColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerHandbookItemColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerHandbookItemColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerHandbookItemColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerHandbookItemColumn)GetAll()(list []dbPlayerHandbookItemData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerHandbookItemColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerHandbookItemData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerHandbookItemColumn)Get(id int32)(v *dbPlayerHandbookItemData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerHandbookItemColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerHandbookItemData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerHandbookItemColumn)Set(v dbPlayerHandbookItemData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerHandbookItemColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.Id)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.Id)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerHandbookItemColumn)Add(v *dbPlayerHandbookItemData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerHandbookItemColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.Id)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.Id)
		return false
	}
	d:=&dbPlayerHandbookItemData{}
	v.clone_to(d)
	this.m_data[int32(v.Id)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerHandbookItemColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerHandbookItemColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerHandbookItemColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerHandbookItemColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerHandbookItemData)
	this.m_changed = true
	return
}
func (this *dbPlayerHandbookItemColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerHandbookItemColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
type dbPlayerHeadItemColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerHeadItemData
	m_changed bool
}
func (this *dbPlayerHeadItemColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerHeadItemList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerHeadItemData{}
		d.from_pb(v)
		this.m_data[int32(d.Id)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerHeadItemColumn)save( )(data []byte,err error){
	pb := &db.PlayerHeadItemList{}
	pb.List=make([]*db.PlayerHeadItem,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerHeadItemColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerHeadItemColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerHeadItemColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerHeadItemColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerHeadItemColumn)GetAll()(list []dbPlayerHeadItemData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerHeadItemColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerHeadItemData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerHeadItemColumn)Get(id int32)(v *dbPlayerHeadItemData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerHeadItemColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerHeadItemData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerHeadItemColumn)Set(v dbPlayerHeadItemData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerHeadItemColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.Id)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.Id)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerHeadItemColumn)Add(v *dbPlayerHeadItemData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerHeadItemColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.Id)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.Id)
		return false
	}
	d:=&dbPlayerHeadItemData{}
	v.clone_to(d)
	this.m_data[int32(v.Id)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerHeadItemColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerHeadItemColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerHeadItemColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerHeadItemColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerHeadItemData)
	this.m_changed = true
	return
}
func (this *dbPlayerHeadItemColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerHeadItemColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
type dbPlayerSuitAwardColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerSuitAwardData
	m_changed bool
}
func (this *dbPlayerSuitAwardColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerSuitAwardList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerSuitAwardData{}
		d.from_pb(v)
		this.m_data[int32(d.Id)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerSuitAwardColumn)save( )(data []byte,err error){
	pb := &db.PlayerSuitAwardList{}
	pb.List=make([]*db.PlayerSuitAward,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerSuitAwardColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerSuitAwardColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerSuitAwardColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerSuitAwardColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerSuitAwardColumn)GetAll()(list []dbPlayerSuitAwardData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerSuitAwardColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerSuitAwardData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerSuitAwardColumn)Get(id int32)(v *dbPlayerSuitAwardData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerSuitAwardColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerSuitAwardData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerSuitAwardColumn)Set(v dbPlayerSuitAwardData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerSuitAwardColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.Id)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.Id)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerSuitAwardColumn)Add(v *dbPlayerSuitAwardData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerSuitAwardColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.Id)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.Id)
		return false
	}
	d:=&dbPlayerSuitAwardData{}
	v.clone_to(d)
	this.m_data[int32(v.Id)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerSuitAwardColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerSuitAwardColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerSuitAwardColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerSuitAwardColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerSuitAwardData)
	this.m_changed = true
	return
}
func (this *dbPlayerSuitAwardColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerSuitAwardColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerSuitAwardColumn)GetAwardTime(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerSuitAwardColumn.GetAwardTime")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.AwardTime
	return v,true
}
func (this *dbPlayerSuitAwardColumn)SetAwardTime(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerSuitAwardColumn.SetAwardTime")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.AwardTime = v
	this.m_changed = true
	return true
}
type dbPlayerWorldChatColumn struct{
	m_row *dbPlayerRow
	m_data *dbPlayerWorldChatData
	m_changed bool
}
func (this *dbPlayerWorldChatColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_data = &dbPlayerWorldChatData{}
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerWorldChat{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_data = &dbPlayerWorldChatData{}
	this.m_data.from_pb(pb)
	this.m_changed = false
	return
}
func (this *dbPlayerWorldChatColumn)save( )(data []byte,err error){
	pb:=this.m_data.to_pb()
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerWorldChatColumn)Get( )(v *dbPlayerWorldChatData ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerWorldChatColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v=&dbPlayerWorldChatData{}
	this.m_data.clone_to(v)
	return
}
func (this *dbPlayerWorldChatColumn)Set(v dbPlayerWorldChatData ){
	this.m_row.m_lock.UnSafeLock("dbPlayerWorldChatColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=&dbPlayerWorldChatData{}
	v.clone_to(this.m_data)
	this.m_changed=true
	return
}
func (this *dbPlayerWorldChatColumn)GetLastChatTime( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerWorldChatColumn.GetLastChatTime")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.LastChatTime
	return
}
func (this *dbPlayerWorldChatColumn)SetLastChatTime(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerWorldChatColumn.SetLastChatTime")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.LastChatTime = v
	this.m_changed = true
	return
}
func (this *dbPlayerWorldChatColumn)GetLastPullTime( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerWorldChatColumn.GetLastPullTime")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.LastPullTime
	return
}
func (this *dbPlayerWorldChatColumn)SetLastPullTime(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerWorldChatColumn.SetLastPullTime")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.LastPullTime = v
	this.m_changed = true
	return
}
func (this *dbPlayerWorldChatColumn)GetLastMsgIndex( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerWorldChatColumn.GetLastMsgIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.LastMsgIndex
	return
}
func (this *dbPlayerWorldChatColumn)SetLastMsgIndex(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerWorldChatColumn.SetLastMsgIndex")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.LastMsgIndex = v
	this.m_changed = true
	return
}
type dbPlayerAnouncementColumn struct{
	m_row *dbPlayerRow
	m_data *dbPlayerAnouncementData
	m_changed bool
}
func (this *dbPlayerAnouncementColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_data = &dbPlayerAnouncementData{}
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerAnouncement{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_data = &dbPlayerAnouncementData{}
	this.m_data.from_pb(pb)
	this.m_changed = false
	return
}
func (this *dbPlayerAnouncementColumn)save( )(data []byte,err error){
	pb:=this.m_data.to_pb()
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerAnouncementColumn)Get( )(v *dbPlayerAnouncementData ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerAnouncementColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v=&dbPlayerAnouncementData{}
	this.m_data.clone_to(v)
	return
}
func (this *dbPlayerAnouncementColumn)Set(v dbPlayerAnouncementData ){
	this.m_row.m_lock.UnSafeLock("dbPlayerAnouncementColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=&dbPlayerAnouncementData{}
	v.clone_to(this.m_data)
	this.m_changed=true
	return
}
func (this *dbPlayerAnouncementColumn)GetLastSendTime( )(v int32 ){
	this.m_row.m_lock.UnSafeRLock("dbPlayerAnouncementColumn.GetLastSendTime")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = this.m_data.LastSendTime
	return
}
func (this *dbPlayerAnouncementColumn)SetLastSendTime(v int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerAnouncementColumn.SetLastSendTime")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.LastSendTime = v
	this.m_changed = true
	return
}
type dbPlayerFirstDrawCardColumn struct{
	m_row *dbPlayerRow
	m_data map[int32]*dbPlayerFirstDrawCardData
	m_changed bool
}
func (this *dbPlayerFirstDrawCardColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.PlayerFirstDrawCardList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetPlayerId())
		return
	}
	for _, v := range pb.List {
		d := &dbPlayerFirstDrawCardData{}
		d.from_pb(v)
		this.m_data[int32(d.Id)] = d
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFirstDrawCardColumn)save( )(data []byte,err error){
	pb := &db.PlayerFirstDrawCardList{}
	pb.List=make([]*db.PlayerFirstDrawCard,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetPlayerId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbPlayerFirstDrawCardColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFirstDrawCardColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbPlayerFirstDrawCardColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFirstDrawCardColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbPlayerFirstDrawCardColumn)GetAll()(list []dbPlayerFirstDrawCardData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFirstDrawCardColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbPlayerFirstDrawCardData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbPlayerFirstDrawCardColumn)Get(id int32)(v *dbPlayerFirstDrawCardData){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFirstDrawCardColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbPlayerFirstDrawCardData{}
	d.clone_to(v)
	return
}
func (this *dbPlayerFirstDrawCardColumn)Set(v dbPlayerFirstDrawCardData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFirstDrawCardColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.Id)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), v.Id)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbPlayerFirstDrawCardColumn)Add(v *dbPlayerFirstDrawCardData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFirstDrawCardColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.Id)]
	if has {
		log.Error("already added %v %v",this.m_row.GetPlayerId(), v.Id)
		return false
	}
	d:=&dbPlayerFirstDrawCardData{}
	v.clone_to(d)
	this.m_data[int32(v.Id)]=d
	this.m_changed = true
	return true
}
func (this *dbPlayerFirstDrawCardColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbPlayerFirstDrawCardColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbPlayerFirstDrawCardColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbPlayerFirstDrawCardColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbPlayerFirstDrawCardData)
	this.m_changed = true
	return
}
func (this *dbPlayerFirstDrawCardColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFirstDrawCardColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbPlayerFirstDrawCardColumn)GetDrawed(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbPlayerFirstDrawCardColumn.GetDrawed")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.Drawed
	return v,true
}
func (this *dbPlayerFirstDrawCardColumn)SetDrawed(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbPlayerFirstDrawCardColumn.SetDrawed")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetPlayerId(), id)
		return
	}
	d.Drawed = v
	this.m_changed = true
	return true
}
type dbPlayerRow struct {
	m_table *dbPlayerTable
	m_lock       *RWMutex
	m_loaded  bool
	m_new     bool
	m_remove  bool
	m_touch      int32
	m_releasable bool
	m_valid   bool
	m_PlayerId        int32
	m_Account_changed bool
	m_Account string
	m_Name_changed bool
	m_Name string
	Info dbPlayerInfoColumn
	Global dbPlayerGlobalColumn
	Items dbPlayerItemColumn
	Roles dbPlayerRoleColumn
	BattleTeam dbPlayerBattleTeamColumn
	CampaignCommon dbPlayerCampaignCommonColumn
	Campaigns dbPlayerCampaignColumn
	ChapterUnLock dbPlayerChapterUnLockColumn
	ShopItems dbPlayerShopItemColumn
	ShopLimitedInfos dbPlayerShopLimitedInfoColumn
	Chests dbPlayerChestColumn
	Mails dbPlayerMailColumn
	DialyTasks dbPlayerDialyTaskColumn
	Achieves dbPlayerAchieveColumn
	FinishedAchieves dbPlayerFinishedAchieveColumn
	DailyTaskWholeDailys dbPlayerDailyTaskWholeDailyColumn
	SevenActivitys dbPlayerSevenActivityColumn
	Guidess dbPlayerGuidesColumn
	FriendRelative dbPlayerFriendRelativeColumn
	Friends dbPlayerFriendColumn
	FriendReqs dbPlayerFriendReqColumn
	FriendPoints dbPlayerFriendPointColumn
	FriendChatUnreadIds dbPlayerFriendChatUnreadIdColumn
	FriendChatUnreadMessages dbPlayerFriendChatUnreadMessageColumn
	ChaterOpenRequest dbPlayerChaterOpenRequestColumn
	HandbookItems dbPlayerHandbookItemColumn
	HeadItems dbPlayerHeadItemColumn
	SuitAwards dbPlayerSuitAwardColumn
	WorldChat dbPlayerWorldChatColumn
	Anouncement dbPlayerAnouncementColumn
	FirstDrawCards dbPlayerFirstDrawCardColumn
}
func new_dbPlayerRow(table *dbPlayerTable, PlayerId int32) (r *dbPlayerRow) {
	this := &dbPlayerRow{}
	this.m_table = table
	this.m_PlayerId = PlayerId
	this.m_lock = NewRWMutex()
	this.m_Account_changed=true
	this.m_Name_changed=true
	this.Info.m_row=this
	this.Info.m_data=&dbPlayerInfoData{}
	this.Global.m_row=this
	this.Global.m_data=&dbPlayerGlobalData{}
	this.Items.m_row=this
	this.Items.m_data=make(map[int32]*dbPlayerItemData)
	this.Roles.m_row=this
	this.Roles.m_data=make(map[int32]*dbPlayerRoleData)
	this.BattleTeam.m_row=this
	this.BattleTeam.m_data=&dbPlayerBattleTeamData{}
	this.CampaignCommon.m_row=this
	this.CampaignCommon.m_data=&dbPlayerCampaignCommonData{}
	this.Campaigns.m_row=this
	this.Campaigns.m_data=make(map[int32]*dbPlayerCampaignData)
	this.ChapterUnLock.m_row=this
	this.ChapterUnLock.m_data=&dbPlayerChapterUnLockData{}
	this.ShopItems.m_row=this
	this.ShopItems.m_data=make(map[int32]*dbPlayerShopItemData)
	this.ShopLimitedInfos.m_row=this
	this.ShopLimitedInfos.m_data=make(map[int32]*dbPlayerShopLimitedInfoData)
	this.Chests.m_row=this
	this.Chests.m_data=make(map[int32]*dbPlayerChestData)
	this.Mails.m_row=this
	this.Mails.m_data=make(map[int32]*dbPlayerMailData)
	this.DialyTasks.m_row=this
	this.DialyTasks.m_data=make(map[int32]*dbPlayerDialyTaskData)
	this.Achieves.m_row=this
	this.Achieves.m_data=make(map[int32]*dbPlayerAchieveData)
	this.FinishedAchieves.m_row=this
	this.FinishedAchieves.m_data=make(map[int32]*dbPlayerFinishedAchieveData)
	this.DailyTaskWholeDailys.m_row=this
	this.DailyTaskWholeDailys.m_data=make(map[int32]*dbPlayerDailyTaskWholeDailyData)
	this.SevenActivitys.m_row=this
	this.SevenActivitys.m_data=make(map[int32]*dbPlayerSevenActivityData)
	this.Guidess.m_row=this
	this.Guidess.m_data=make(map[int32]*dbPlayerGuidesData)
	this.FriendRelative.m_row=this
	this.FriendRelative.m_data=&dbPlayerFriendRelativeData{}
	this.Friends.m_row=this
	this.Friends.m_data=make(map[int32]*dbPlayerFriendData)
	this.FriendReqs.m_row=this
	this.FriendReqs.m_data=make(map[int32]*dbPlayerFriendReqData)
	this.FriendPoints.m_row=this
	this.FriendPoints.m_data=make(map[int32]*dbPlayerFriendPointData)
	this.FriendChatUnreadIds.m_row=this
	this.FriendChatUnreadIds.m_data=make(map[int32]*dbPlayerFriendChatUnreadIdData)
	this.FriendChatUnreadMessages.m_row=this
	this.FriendChatUnreadMessages.m_data=make(map[int64]*dbPlayerFriendChatUnreadMessageData)
	this.ChaterOpenRequest.m_row=this
	this.ChaterOpenRequest.m_data=&dbPlayerChaterOpenRequestData{}
	this.HandbookItems.m_row=this
	this.HandbookItems.m_data=make(map[int32]*dbPlayerHandbookItemData)
	this.HeadItems.m_row=this
	this.HeadItems.m_data=make(map[int32]*dbPlayerHeadItemData)
	this.SuitAwards.m_row=this
	this.SuitAwards.m_data=make(map[int32]*dbPlayerSuitAwardData)
	this.WorldChat.m_row=this
	this.WorldChat.m_data=&dbPlayerWorldChatData{}
	this.Anouncement.m_row=this
	this.Anouncement.m_data=&dbPlayerAnouncementData{}
	this.FirstDrawCards.m_row=this
	this.FirstDrawCards.m_data=make(map[int32]*dbPlayerFirstDrawCardData)
	return this
}
func (this *dbPlayerRow) GetPlayerId() (r int32) {
	return this.m_PlayerId
}
func (this *dbPlayerRow) save_data(release bool) (err error, released bool, state int32, update_string string, args []interface{}) {
	this.m_lock.UnSafeLock("dbPlayerRow.save_data")
	defer this.m_lock.UnSafeUnlock()
	if this.m_new {
		db_args:=new_db_args(34)
		db_args.Push(this.m_PlayerId)
		db_args.Push(this.m_Account)
		db_args.Push(this.m_Name)
		dInfo,db_err:=this.Info.save()
		if db_err!=nil{
			log.Error("insert save Info failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dInfo)
		dGlobal,db_err:=this.Global.save()
		if db_err!=nil{
			log.Error("insert save Global failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dGlobal)
		dItems,db_err:=this.Items.save()
		if db_err!=nil{
			log.Error("insert save Item failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dItems)
		dRoles,db_err:=this.Roles.save()
		if db_err!=nil{
			log.Error("insert save Role failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dRoles)
		dBattleTeam,db_err:=this.BattleTeam.save()
		if db_err!=nil{
			log.Error("insert save BattleTeam failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dBattleTeam)
		dCampaignCommon,db_err:=this.CampaignCommon.save()
		if db_err!=nil{
			log.Error("insert save CampaignCommon failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dCampaignCommon)
		dCampaigns,db_err:=this.Campaigns.save()
		if db_err!=nil{
			log.Error("insert save Campaign failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dCampaigns)
		dChapterUnLock,db_err:=this.ChapterUnLock.save()
		if db_err!=nil{
			log.Error("insert save ChapterUnLock failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dChapterUnLock)
		dShopItems,db_err:=this.ShopItems.save()
		if db_err!=nil{
			log.Error("insert save ShopItem failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dShopItems)
		dShopLimitedInfos,db_err:=this.ShopLimitedInfos.save()
		if db_err!=nil{
			log.Error("insert save ShopLimitedInfo failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dShopLimitedInfos)
		dChests,db_err:=this.Chests.save()
		if db_err!=nil{
			log.Error("insert save Chest failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dChests)
		dMails,db_err:=this.Mails.save()
		if db_err!=nil{
			log.Error("insert save Mail failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dMails)
		dDialyTasks,db_err:=this.DialyTasks.save()
		if db_err!=nil{
			log.Error("insert save DialyTask failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dDialyTasks)
		dAchieves,db_err:=this.Achieves.save()
		if db_err!=nil{
			log.Error("insert save Achieve failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dAchieves)
		dFinishedAchieves,db_err:=this.FinishedAchieves.save()
		if db_err!=nil{
			log.Error("insert save FinishedAchieve failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dFinishedAchieves)
		dDailyTaskWholeDailys,db_err:=this.DailyTaskWholeDailys.save()
		if db_err!=nil{
			log.Error("insert save DailyTaskWholeDaily failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dDailyTaskWholeDailys)
		dSevenActivitys,db_err:=this.SevenActivitys.save()
		if db_err!=nil{
			log.Error("insert save SevenActivity failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dSevenActivitys)
		dGuidess,db_err:=this.Guidess.save()
		if db_err!=nil{
			log.Error("insert save Guides failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dGuidess)
		dFriendRelative,db_err:=this.FriendRelative.save()
		if db_err!=nil{
			log.Error("insert save FriendRelative failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dFriendRelative)
		dFriends,db_err:=this.Friends.save()
		if db_err!=nil{
			log.Error("insert save Friend failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dFriends)
		dFriendReqs,db_err:=this.FriendReqs.save()
		if db_err!=nil{
			log.Error("insert save FriendReq failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dFriendReqs)
		dFriendPoints,db_err:=this.FriendPoints.save()
		if db_err!=nil{
			log.Error("insert save FriendPoint failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dFriendPoints)
		dFriendChatUnreadIds,db_err:=this.FriendChatUnreadIds.save()
		if db_err!=nil{
			log.Error("insert save FriendChatUnreadId failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dFriendChatUnreadIds)
		dFriendChatUnreadMessages,db_err:=this.FriendChatUnreadMessages.save()
		if db_err!=nil{
			log.Error("insert save FriendChatUnreadMessage failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dFriendChatUnreadMessages)
		dChaterOpenRequest,db_err:=this.ChaterOpenRequest.save()
		if db_err!=nil{
			log.Error("insert save ChaterOpenRequest failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dChaterOpenRequest)
		dHandbookItems,db_err:=this.HandbookItems.save()
		if db_err!=nil{
			log.Error("insert save HandbookItem failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dHandbookItems)
		dHeadItems,db_err:=this.HeadItems.save()
		if db_err!=nil{
			log.Error("insert save HeadItem failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dHeadItems)
		dSuitAwards,db_err:=this.SuitAwards.save()
		if db_err!=nil{
			log.Error("insert save SuitAward failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dSuitAwards)
		dWorldChat,db_err:=this.WorldChat.save()
		if db_err!=nil{
			log.Error("insert save WorldChat failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dWorldChat)
		dAnouncement,db_err:=this.Anouncement.save()
		if db_err!=nil{
			log.Error("insert save Anouncement failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dAnouncement)
		dFirstDrawCards,db_err:=this.FirstDrawCards.save()
		if db_err!=nil{
			log.Error("insert save FirstDrawCard failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dFirstDrawCards)
		args=db_args.GetArgs()
		state = 1
	} else {
		if this.m_Account_changed||this.m_Name_changed||this.Info.m_changed||this.Global.m_changed||this.Items.m_changed||this.Roles.m_changed||this.BattleTeam.m_changed||this.CampaignCommon.m_changed||this.Campaigns.m_changed||this.ChapterUnLock.m_changed||this.ShopItems.m_changed||this.ShopLimitedInfos.m_changed||this.Chests.m_changed||this.Mails.m_changed||this.DialyTasks.m_changed||this.Achieves.m_changed||this.FinishedAchieves.m_changed||this.DailyTaskWholeDailys.m_changed||this.SevenActivitys.m_changed||this.Guidess.m_changed||this.FriendRelative.m_changed||this.Friends.m_changed||this.FriendReqs.m_changed||this.FriendPoints.m_changed||this.FriendChatUnreadIds.m_changed||this.FriendChatUnreadMessages.m_changed||this.ChaterOpenRequest.m_changed||this.HandbookItems.m_changed||this.HeadItems.m_changed||this.SuitAwards.m_changed||this.WorldChat.m_changed||this.Anouncement.m_changed||this.FirstDrawCards.m_changed{
			update_string = "UPDATE Players SET "
			db_args:=new_db_args(34)
			if this.m_Account_changed{
				update_string+="Account=?,"
				db_args.Push(this.m_Account)
			}
			if this.m_Name_changed{
				update_string+="Name=?,"
				db_args.Push(this.m_Name)
			}
			if this.Info.m_changed{
				update_string+="Info=?,"
				dInfo,err:=this.Info.save()
				if err!=nil{
					log.Error("update save Info failed")
					return err,false,0,"",nil
				}
				db_args.Push(dInfo)
			}
			if this.Global.m_changed{
				update_string+="Global=?,"
				dGlobal,err:=this.Global.save()
				if err!=nil{
					log.Error("update save Global failed")
					return err,false,0,"",nil
				}
				db_args.Push(dGlobal)
			}
			if this.Items.m_changed{
				update_string+="Items=?,"
				dItems,err:=this.Items.save()
				if err!=nil{
					log.Error("insert save Item failed")
					return err,false,0,"",nil
				}
				db_args.Push(dItems)
			}
			if this.Roles.m_changed{
				update_string+="Roles=?,"
				dRoles,err:=this.Roles.save()
				if err!=nil{
					log.Error("insert save Role failed")
					return err,false,0,"",nil
				}
				db_args.Push(dRoles)
			}
			if this.BattleTeam.m_changed{
				update_string+="BattleTeam=?,"
				dBattleTeam,err:=this.BattleTeam.save()
				if err!=nil{
					log.Error("update save BattleTeam failed")
					return err,false,0,"",nil
				}
				db_args.Push(dBattleTeam)
			}
			if this.CampaignCommon.m_changed{
				update_string+="CampaignCommon=?,"
				dCampaignCommon,err:=this.CampaignCommon.save()
				if err!=nil{
					log.Error("update save CampaignCommon failed")
					return err,false,0,"",nil
				}
				db_args.Push(dCampaignCommon)
			}
			if this.Campaigns.m_changed{
				update_string+="Campaigns=?,"
				dCampaigns,err:=this.Campaigns.save()
				if err!=nil{
					log.Error("insert save Campaign failed")
					return err,false,0,"",nil
				}
				db_args.Push(dCampaigns)
			}
			if this.ChapterUnLock.m_changed{
				update_string+="ChapterUnLock=?,"
				dChapterUnLock,err:=this.ChapterUnLock.save()
				if err!=nil{
					log.Error("update save ChapterUnLock failed")
					return err,false,0,"",nil
				}
				db_args.Push(dChapterUnLock)
			}
			if this.ShopItems.m_changed{
				update_string+="ShopItems=?,"
				dShopItems,err:=this.ShopItems.save()
				if err!=nil{
					log.Error("insert save ShopItem failed")
					return err,false,0,"",nil
				}
				db_args.Push(dShopItems)
			}
			if this.ShopLimitedInfos.m_changed{
				update_string+="ShopLimitedInfos=?,"
				dShopLimitedInfos,err:=this.ShopLimitedInfos.save()
				if err!=nil{
					log.Error("insert save ShopLimitedInfo failed")
					return err,false,0,"",nil
				}
				db_args.Push(dShopLimitedInfos)
			}
			if this.Chests.m_changed{
				update_string+="Chests=?,"
				dChests,err:=this.Chests.save()
				if err!=nil{
					log.Error("insert save Chest failed")
					return err,false,0,"",nil
				}
				db_args.Push(dChests)
			}
			if this.Mails.m_changed{
				update_string+="Mails=?,"
				dMails,err:=this.Mails.save()
				if err!=nil{
					log.Error("insert save Mail failed")
					return err,false,0,"",nil
				}
				db_args.Push(dMails)
			}
			if this.DialyTasks.m_changed{
				update_string+="DialyTasks=?,"
				dDialyTasks,err:=this.DialyTasks.save()
				if err!=nil{
					log.Error("insert save DialyTask failed")
					return err,false,0,"",nil
				}
				db_args.Push(dDialyTasks)
			}
			if this.Achieves.m_changed{
				update_string+="Achieves=?,"
				dAchieves,err:=this.Achieves.save()
				if err!=nil{
					log.Error("insert save Achieve failed")
					return err,false,0,"",nil
				}
				db_args.Push(dAchieves)
			}
			if this.FinishedAchieves.m_changed{
				update_string+="FinishedAchieves=?,"
				dFinishedAchieves,err:=this.FinishedAchieves.save()
				if err!=nil{
					log.Error("insert save FinishedAchieve failed")
					return err,false,0,"",nil
				}
				db_args.Push(dFinishedAchieves)
			}
			if this.DailyTaskWholeDailys.m_changed{
				update_string+="DailyTaskWholeDailys=?,"
				dDailyTaskWholeDailys,err:=this.DailyTaskWholeDailys.save()
				if err!=nil{
					log.Error("insert save DailyTaskWholeDaily failed")
					return err,false,0,"",nil
				}
				db_args.Push(dDailyTaskWholeDailys)
			}
			if this.SevenActivitys.m_changed{
				update_string+="SevenActivitys=?,"
				dSevenActivitys,err:=this.SevenActivitys.save()
				if err!=nil{
					log.Error("insert save SevenActivity failed")
					return err,false,0,"",nil
				}
				db_args.Push(dSevenActivitys)
			}
			if this.Guidess.m_changed{
				update_string+="Guidess=?,"
				dGuidess,err:=this.Guidess.save()
				if err!=nil{
					log.Error("insert save Guides failed")
					return err,false,0,"",nil
				}
				db_args.Push(dGuidess)
			}
			if this.FriendRelative.m_changed{
				update_string+="FriendRelative=?,"
				dFriendRelative,err:=this.FriendRelative.save()
				if err!=nil{
					log.Error("update save FriendRelative failed")
					return err,false,0,"",nil
				}
				db_args.Push(dFriendRelative)
			}
			if this.Friends.m_changed{
				update_string+="Friends=?,"
				dFriends,err:=this.Friends.save()
				if err!=nil{
					log.Error("insert save Friend failed")
					return err,false,0,"",nil
				}
				db_args.Push(dFriends)
			}
			if this.FriendReqs.m_changed{
				update_string+="FriendReqs=?,"
				dFriendReqs,err:=this.FriendReqs.save()
				if err!=nil{
					log.Error("insert save FriendReq failed")
					return err,false,0,"",nil
				}
				db_args.Push(dFriendReqs)
			}
			if this.FriendPoints.m_changed{
				update_string+="FriendPoints=?,"
				dFriendPoints,err:=this.FriendPoints.save()
				if err!=nil{
					log.Error("insert save FriendPoint failed")
					return err,false,0,"",nil
				}
				db_args.Push(dFriendPoints)
			}
			if this.FriendChatUnreadIds.m_changed{
				update_string+="FriendChatUnreadIds=?,"
				dFriendChatUnreadIds,err:=this.FriendChatUnreadIds.save()
				if err!=nil{
					log.Error("insert save FriendChatUnreadId failed")
					return err,false,0,"",nil
				}
				db_args.Push(dFriendChatUnreadIds)
			}
			if this.FriendChatUnreadMessages.m_changed{
				update_string+="FriendChatUnreadMessages=?,"
				dFriendChatUnreadMessages,err:=this.FriendChatUnreadMessages.save()
				if err!=nil{
					log.Error("insert save FriendChatUnreadMessage failed")
					return err,false,0,"",nil
				}
				db_args.Push(dFriendChatUnreadMessages)
			}
			if this.ChaterOpenRequest.m_changed{
				update_string+="ChaterOpenRequest=?,"
				dChaterOpenRequest,err:=this.ChaterOpenRequest.save()
				if err!=nil{
					log.Error("update save ChaterOpenRequest failed")
					return err,false,0,"",nil
				}
				db_args.Push(dChaterOpenRequest)
			}
			if this.HandbookItems.m_changed{
				update_string+="HandbookItems=?,"
				dHandbookItems,err:=this.HandbookItems.save()
				if err!=nil{
					log.Error("insert save HandbookItem failed")
					return err,false,0,"",nil
				}
				db_args.Push(dHandbookItems)
			}
			if this.HeadItems.m_changed{
				update_string+="HeadItems=?,"
				dHeadItems,err:=this.HeadItems.save()
				if err!=nil{
					log.Error("insert save HeadItem failed")
					return err,false,0,"",nil
				}
				db_args.Push(dHeadItems)
			}
			if this.SuitAwards.m_changed{
				update_string+="SuitAwards=?,"
				dSuitAwards,err:=this.SuitAwards.save()
				if err!=nil{
					log.Error("insert save SuitAward failed")
					return err,false,0,"",nil
				}
				db_args.Push(dSuitAwards)
			}
			if this.WorldChat.m_changed{
				update_string+="WorldChat=?,"
				dWorldChat,err:=this.WorldChat.save()
				if err!=nil{
					log.Error("update save WorldChat failed")
					return err,false,0,"",nil
				}
				db_args.Push(dWorldChat)
			}
			if this.Anouncement.m_changed{
				update_string+="Anouncement=?,"
				dAnouncement,err:=this.Anouncement.save()
				if err!=nil{
					log.Error("update save Anouncement failed")
					return err,false,0,"",nil
				}
				db_args.Push(dAnouncement)
			}
			if this.FirstDrawCards.m_changed{
				update_string+="FirstDrawCards=?,"
				dFirstDrawCards,err:=this.FirstDrawCards.save()
				if err!=nil{
					log.Error("insert save FirstDrawCard failed")
					return err,false,0,"",nil
				}
				db_args.Push(dFirstDrawCards)
			}
			update_string = strings.TrimRight(update_string, ", ")
			update_string+=" WHERE PlayerId=?"
			db_args.Push(this.m_PlayerId)
			args=db_args.GetArgs()
			state = 2
		}
	}
	this.m_new = false
	this.m_Account_changed = false
	this.m_Name_changed = false
	this.Info.m_changed = false
	this.Global.m_changed = false
	this.Items.m_changed = false
	this.Roles.m_changed = false
	this.BattleTeam.m_changed = false
	this.CampaignCommon.m_changed = false
	this.Campaigns.m_changed = false
	this.ChapterUnLock.m_changed = false
	this.ShopItems.m_changed = false
	this.ShopLimitedInfos.m_changed = false
	this.Chests.m_changed = false
	this.Mails.m_changed = false
	this.DialyTasks.m_changed = false
	this.Achieves.m_changed = false
	this.FinishedAchieves.m_changed = false
	this.DailyTaskWholeDailys.m_changed = false
	this.SevenActivitys.m_changed = false
	this.Guidess.m_changed = false
	this.FriendRelative.m_changed = false
	this.Friends.m_changed = false
	this.FriendReqs.m_changed = false
	this.FriendPoints.m_changed = false
	this.FriendChatUnreadIds.m_changed = false
	this.FriendChatUnreadMessages.m_changed = false
	this.ChaterOpenRequest.m_changed = false
	this.HandbookItems.m_changed = false
	this.HeadItems.m_changed = false
	this.SuitAwards.m_changed = false
	this.WorldChat.m_changed = false
	this.Anouncement.m_changed = false
	this.FirstDrawCards.m_changed = false
	if release && this.m_loaded {
		atomic.AddInt32(&this.m_table.m_gc_n, -1)
		this.m_loaded = false
		released = true
	}
	return nil,released,state,update_string,args
}
func (this *dbPlayerRow) Save(release bool) (err error, d bool, released bool) {
	err,released, state, update_string, args := this.save_data(release)
	if err != nil {
		log.Error("save data failed")
		return err, false, false
	}
	if state == 0 {
		d = false
	} else if state == 1 {
		_, err = this.m_table.m_dbc.StmtExec(this.m_table.m_save_insert_stmt, args...)
		if err != nil {
			log.Error("INSERT Players exec failed %v ", this.m_PlayerId)
			return err, false, released
		}
		d = true
	} else if state == 2 {
		_, err = this.m_table.m_dbc.Exec(update_string, args...)
		if err != nil {
			log.Error("UPDATE Players exec failed %v", this.m_PlayerId)
			return err, false, released
		}
		d = true
	}
	return nil, d, released
}
func (this *dbPlayerRow) Touch(releasable bool) {
	this.m_touch = int32(time.Now().Unix())
	this.m_releasable = releasable
}
type dbPlayerRowSort struct {
	rows []*dbPlayerRow
}
func (this *dbPlayerRowSort) Len() (length int) {
	return len(this.rows)
}
func (this *dbPlayerRowSort) Less(i int, j int) (less bool) {
	return this.rows[i].m_touch < this.rows[j].m_touch
}
func (this *dbPlayerRowSort) Swap(i int, j int) {
	temp := this.rows[i]
	this.rows[i] = this.rows[j]
	this.rows[j] = temp
}
type dbPlayerTable struct{
	m_dbc *DBC
	m_lock *RWMutex
	m_rows map[int32]*dbPlayerRow
	m_new_rows map[int32]*dbPlayerRow
	m_removed_rows map[int32]*dbPlayerRow
	m_gc_n int32
	m_gcing int32
	m_pool_size int32
	m_preload_select_stmt *sql.Stmt
	m_preload_max_id int32
	m_save_insert_stmt *sql.Stmt
	m_delete_stmt *sql.Stmt
}
func new_dbPlayerTable(dbc *DBC) (this *dbPlayerTable) {
	this = &dbPlayerTable{}
	this.m_dbc = dbc
	this.m_lock = NewRWMutex()
	this.m_rows = make(map[int32]*dbPlayerRow)
	this.m_new_rows = make(map[int32]*dbPlayerRow)
	this.m_removed_rows = make(map[int32]*dbPlayerRow)
	return this
}
func (this *dbPlayerTable) check_create_table() (err error) {
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS Players(PlayerId int(11),PRIMARY KEY (PlayerId))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS Players failed")
		return
	}
	rows, err := this.m_dbc.Query("SELECT COLUMN_NAME,ORDINAL_POSITION FROM information_schema.`COLUMNS` WHERE TABLE_SCHEMA=? AND TABLE_NAME='Players'", this.m_dbc.m_db_name)
	if err != nil {
		log.Error("SELECT information_schema failed")
		return
	}
	columns := make(map[string]int32)
	for rows.Next() {
		var column_name string
		var ordinal_position int32
		err = rows.Scan(&column_name, &ordinal_position)
		if err != nil {
			log.Error("scan information_schema row failed")
			return
		}
		if ordinal_position < 1 {
			log.Error("col ordinal out of range")
			continue
		}
		columns[column_name] = ordinal_position
	}
	_, hasAccount := columns["Account"]
	if !hasAccount {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN Account varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN Account failed")
			return
		}
	}
	_, hasName := columns["Name"]
	if !hasName {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN Name varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN Name failed")
			return
		}
	}
	_, hasInfo := columns["Info"]
	if !hasInfo {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN Info LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN Info failed")
			return
		}
	}
	_, hasGlobal := columns["Global"]
	if !hasGlobal {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN Global LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN Global failed")
			return
		}
	}
	_, hasItem := columns["Items"]
	if !hasItem {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN Items LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN Items failed")
			return
		}
	}
	_, hasRole := columns["Roles"]
	if !hasRole {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN Roles LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN Roles failed")
			return
		}
	}
	_, hasBattleTeam := columns["BattleTeam"]
	if !hasBattleTeam {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN BattleTeam LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN BattleTeam failed")
			return
		}
	}
	_, hasCampaignCommon := columns["CampaignCommon"]
	if !hasCampaignCommon {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN CampaignCommon LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN CampaignCommon failed")
			return
		}
	}
	_, hasCampaign := columns["Campaigns"]
	if !hasCampaign {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN Campaigns LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN Campaigns failed")
			return
		}
	}
	_, hasChapterUnLock := columns["ChapterUnLock"]
	if !hasChapterUnLock {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN ChapterUnLock LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN ChapterUnLock failed")
			return
		}
	}
	_, hasShopItem := columns["ShopItems"]
	if !hasShopItem {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN ShopItems LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN ShopItems failed")
			return
		}
	}
	_, hasShopLimitedInfo := columns["ShopLimitedInfos"]
	if !hasShopLimitedInfo {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN ShopLimitedInfos LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN ShopLimitedInfos failed")
			return
		}
	}
	_, hasChest := columns["Chests"]
	if !hasChest {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN Chests LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN Chests failed")
			return
		}
	}
	_, hasMail := columns["Mails"]
	if !hasMail {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN Mails LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN Mails failed")
			return
		}
	}
	_, hasDialyTask := columns["DialyTasks"]
	if !hasDialyTask {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN DialyTasks LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN DialyTasks failed")
			return
		}
	}
	_, hasAchieve := columns["Achieves"]
	if !hasAchieve {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN Achieves LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN Achieves failed")
			return
		}
	}
	_, hasFinishedAchieve := columns["FinishedAchieves"]
	if !hasFinishedAchieve {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN FinishedAchieves LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN FinishedAchieves failed")
			return
		}
	}
	_, hasDailyTaskWholeDaily := columns["DailyTaskWholeDailys"]
	if !hasDailyTaskWholeDaily {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN DailyTaskWholeDailys LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN DailyTaskWholeDailys failed")
			return
		}
	}
	_, hasSevenActivity := columns["SevenActivitys"]
	if !hasSevenActivity {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN SevenActivitys LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN SevenActivitys failed")
			return
		}
	}
	_, hasGuides := columns["Guidess"]
	if !hasGuides {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN Guidess LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN Guidess failed")
			return
		}
	}
	_, hasFriendRelative := columns["FriendRelative"]
	if !hasFriendRelative {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN FriendRelative LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN FriendRelative failed")
			return
		}
	}
	_, hasFriend := columns["Friends"]
	if !hasFriend {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN Friends LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN Friends failed")
			return
		}
	}
	_, hasFriendReq := columns["FriendReqs"]
	if !hasFriendReq {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN FriendReqs LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN FriendReqs failed")
			return
		}
	}
	_, hasFriendPoint := columns["FriendPoints"]
	if !hasFriendPoint {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN FriendPoints LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN FriendPoints failed")
			return
		}
	}
	_, hasFriendChatUnreadId := columns["FriendChatUnreadIds"]
	if !hasFriendChatUnreadId {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN FriendChatUnreadIds LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN FriendChatUnreadIds failed")
			return
		}
	}
	_, hasFriendChatUnreadMessage := columns["FriendChatUnreadMessages"]
	if !hasFriendChatUnreadMessage {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN FriendChatUnreadMessages LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN FriendChatUnreadMessages failed")
			return
		}
	}
	_, hasChaterOpenRequest := columns["ChaterOpenRequest"]
	if !hasChaterOpenRequest {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN ChaterOpenRequest LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN ChaterOpenRequest failed")
			return
		}
	}
	_, hasHandbookItem := columns["HandbookItems"]
	if !hasHandbookItem {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN HandbookItems LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN HandbookItems failed")
			return
		}
	}
	_, hasHeadItem := columns["HeadItems"]
	if !hasHeadItem {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN HeadItems LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN HeadItems failed")
			return
		}
	}
	_, hasSuitAward := columns["SuitAwards"]
	if !hasSuitAward {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN SuitAwards LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN SuitAwards failed")
			return
		}
	}
	_, hasWorldChat := columns["WorldChat"]
	if !hasWorldChat {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN WorldChat LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN WorldChat failed")
			return
		}
	}
	_, hasAnouncement := columns["Anouncement"]
	if !hasAnouncement {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN Anouncement LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN Anouncement failed")
			return
		}
	}
	_, hasFirstDrawCard := columns["FirstDrawCards"]
	if !hasFirstDrawCard {
		_, err = this.m_dbc.Exec("ALTER TABLE Players ADD COLUMN FirstDrawCards LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN FirstDrawCards failed")
			return
		}
	}
	return
}
func (this *dbPlayerTable) prepare_preload_select_stmt() (err error) {
	this.m_preload_select_stmt,err=this.m_dbc.StmtPrepare("SELECT PlayerId,Account,Name,Info,Global,Items,Roles,BattleTeam,CampaignCommon,Campaigns,ChapterUnLock,ShopItems,ShopLimitedInfos,Chests,Mails,DialyTasks,Achieves,FinishedAchieves,DailyTaskWholeDailys,SevenActivitys,Guidess,FriendRelative,Friends,FriendReqs,FriendPoints,FriendChatUnreadIds,FriendChatUnreadMessages,ChaterOpenRequest,HandbookItems,HeadItems,SuitAwards,WorldChat,Anouncement,FirstDrawCards FROM Players")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbPlayerTable) prepare_save_insert_stmt()(err error){
	this.m_save_insert_stmt,err=this.m_dbc.StmtPrepare("INSERT INTO Players (PlayerId,Account,Name,Info,Global,Items,Roles,BattleTeam,CampaignCommon,Campaigns,ChapterUnLock,ShopItems,ShopLimitedInfos,Chests,Mails,DialyTasks,Achieves,FinishedAchieves,DailyTaskWholeDailys,SevenActivitys,Guidess,FriendRelative,Friends,FriendReqs,FriendPoints,FriendChatUnreadIds,FriendChatUnreadMessages,ChaterOpenRequest,HandbookItems,HeadItems,SuitAwards,WorldChat,Anouncement,FirstDrawCards) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbPlayerTable) prepare_delete_stmt() (err error) {
	this.m_delete_stmt,err=this.m_dbc.StmtPrepare("DELETE FROM Players WHERE PlayerId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbPlayerTable) Init() (err error) {
	err=this.check_create_table()
	if err!=nil{
		log.Error("check_create_table failed")
		return
	}
	err=this.prepare_preload_select_stmt()
	if err!=nil{
		log.Error("prepare_preload_select_stmt failed")
		return
	}
	err=this.prepare_save_insert_stmt()
	if err!=nil{
		log.Error("prepare_save_insert_stmt failed")
		return
	}
	err=this.prepare_delete_stmt()
	if err!=nil{
		log.Error("prepare_save_insert_stmt failed")
		return
	}
	return
}
func (this *dbPlayerTable) Preload() (err error) {
	r, err := this.m_dbc.StmtQuery(this.m_preload_select_stmt)
	if err != nil {
		log.Error("SELECT")
		return
	}
	var PlayerId int32
	var dAccount string
	var dName string
	var dInfo []byte
	var dGlobal []byte
	var dItems []byte
	var dRoles []byte
	var dBattleTeam []byte
	var dCampaignCommon []byte
	var dCampaigns []byte
	var dChapterUnLock []byte
	var dShopItems []byte
	var dShopLimitedInfos []byte
	var dChests []byte
	var dMails []byte
	var dDialyTasks []byte
	var dAchieves []byte
	var dFinishedAchieves []byte
	var dDailyTaskWholeDailys []byte
	var dSevenActivitys []byte
	var dGuidess []byte
	var dFriendRelative []byte
	var dFriends []byte
	var dFriendReqs []byte
	var dFriendPoints []byte
	var dFriendChatUnreadIds []byte
	var dFriendChatUnreadMessages []byte
	var dChaterOpenRequest []byte
	var dHandbookItems []byte
	var dHeadItems []byte
	var dSuitAwards []byte
	var dWorldChat []byte
	var dAnouncement []byte
	var dFirstDrawCards []byte
		this.m_preload_max_id = 0
	for r.Next() {
		err = r.Scan(&PlayerId,&dAccount,&dName,&dInfo,&dGlobal,&dItems,&dRoles,&dBattleTeam,&dCampaignCommon,&dCampaigns,&dChapterUnLock,&dShopItems,&dShopLimitedInfos,&dChests,&dMails,&dDialyTasks,&dAchieves,&dFinishedAchieves,&dDailyTaskWholeDailys,&dSevenActivitys,&dGuidess,&dFriendRelative,&dFriends,&dFriendReqs,&dFriendPoints,&dFriendChatUnreadIds,&dFriendChatUnreadMessages,&dChaterOpenRequest,&dHandbookItems,&dHeadItems,&dSuitAwards,&dWorldChat,&dAnouncement,&dFirstDrawCards)
		if err != nil {
			log.Error("Scan")
			return
		}
		if PlayerId>this.m_preload_max_id{
			this.m_preload_max_id =PlayerId
		}
		row := new_dbPlayerRow(this,PlayerId)
		row.m_Account=dAccount
		row.m_Name=dName
		err = row.Info.load(dInfo)
		if err != nil {
			log.Error("Info %v", PlayerId)
			return
		}
		err = row.Global.load(dGlobal)
		if err != nil {
			log.Error("Global %v", PlayerId)
			return
		}
		err = row.Items.load(dItems)
		if err != nil {
			log.Error("Items %v", PlayerId)
			return
		}
		err = row.Roles.load(dRoles)
		if err != nil {
			log.Error("Roles %v", PlayerId)
			return
		}
		err = row.BattleTeam.load(dBattleTeam)
		if err != nil {
			log.Error("BattleTeam %v", PlayerId)
			return
		}
		err = row.CampaignCommon.load(dCampaignCommon)
		if err != nil {
			log.Error("CampaignCommon %v", PlayerId)
			return
		}
		err = row.Campaigns.load(dCampaigns)
		if err != nil {
			log.Error("Campaigns %v", PlayerId)
			return
		}
		err = row.ChapterUnLock.load(dChapterUnLock)
		if err != nil {
			log.Error("ChapterUnLock %v", PlayerId)
			return
		}
		err = row.ShopItems.load(dShopItems)
		if err != nil {
			log.Error("ShopItems %v", PlayerId)
			return
		}
		err = row.ShopLimitedInfos.load(dShopLimitedInfos)
		if err != nil {
			log.Error("ShopLimitedInfos %v", PlayerId)
			return
		}
		err = row.Chests.load(dChests)
		if err != nil {
			log.Error("Chests %v", PlayerId)
			return
		}
		err = row.Mails.load(dMails)
		if err != nil {
			log.Error("Mails %v", PlayerId)
			return
		}
		err = row.DialyTasks.load(dDialyTasks)
		if err != nil {
			log.Error("DialyTasks %v", PlayerId)
			return
		}
		err = row.Achieves.load(dAchieves)
		if err != nil {
			log.Error("Achieves %v", PlayerId)
			return
		}
		err = row.FinishedAchieves.load(dFinishedAchieves)
		if err != nil {
			log.Error("FinishedAchieves %v", PlayerId)
			return
		}
		err = row.DailyTaskWholeDailys.load(dDailyTaskWholeDailys)
		if err != nil {
			log.Error("DailyTaskWholeDailys %v", PlayerId)
			return
		}
		err = row.SevenActivitys.load(dSevenActivitys)
		if err != nil {
			log.Error("SevenActivitys %v", PlayerId)
			return
		}
		err = row.Guidess.load(dGuidess)
		if err != nil {
			log.Error("Guidess %v", PlayerId)
			return
		}
		err = row.FriendRelative.load(dFriendRelative)
		if err != nil {
			log.Error("FriendRelative %v", PlayerId)
			return
		}
		err = row.Friends.load(dFriends)
		if err != nil {
			log.Error("Friends %v", PlayerId)
			return
		}
		err = row.FriendReqs.load(dFriendReqs)
		if err != nil {
			log.Error("FriendReqs %v", PlayerId)
			return
		}
		err = row.FriendPoints.load(dFriendPoints)
		if err != nil {
			log.Error("FriendPoints %v", PlayerId)
			return
		}
		err = row.FriendChatUnreadIds.load(dFriendChatUnreadIds)
		if err != nil {
			log.Error("FriendChatUnreadIds %v", PlayerId)
			return
		}
		err = row.FriendChatUnreadMessages.load(dFriendChatUnreadMessages)
		if err != nil {
			log.Error("FriendChatUnreadMessages %v", PlayerId)
			return
		}
		err = row.ChaterOpenRequest.load(dChaterOpenRequest)
		if err != nil {
			log.Error("ChaterOpenRequest %v", PlayerId)
			return
		}
		err = row.HandbookItems.load(dHandbookItems)
		if err != nil {
			log.Error("HandbookItems %v", PlayerId)
			return
		}
		err = row.HeadItems.load(dHeadItems)
		if err != nil {
			log.Error("HeadItems %v", PlayerId)
			return
		}
		err = row.SuitAwards.load(dSuitAwards)
		if err != nil {
			log.Error("SuitAwards %v", PlayerId)
			return
		}
		err = row.WorldChat.load(dWorldChat)
		if err != nil {
			log.Error("WorldChat %v", PlayerId)
			return
		}
		err = row.Anouncement.load(dAnouncement)
		if err != nil {
			log.Error("Anouncement %v", PlayerId)
			return
		}
		err = row.FirstDrawCards.load(dFirstDrawCards)
		if err != nil {
			log.Error("FirstDrawCards %v", PlayerId)
			return
		}
		row.m_Account_changed=false
		row.m_Name_changed=false
		row.m_valid = true
		this.m_rows[PlayerId]=row
	}
	return
}
func (this *dbPlayerTable) GetPreloadedMaxId() (max_id int32) {
	return this.m_preload_max_id
}
func (this *dbPlayerTable) fetch_rows(rows map[int32]*dbPlayerRow) (r map[int32]*dbPlayerRow) {
	this.m_lock.UnSafeLock("dbPlayerTable.fetch_rows")
	defer this.m_lock.UnSafeUnlock()
	r = make(map[int32]*dbPlayerRow)
	for i, v := range rows {
		r[i] = v
	}
	return r
}
func (this *dbPlayerTable) fetch_new_rows() (new_rows map[int32]*dbPlayerRow) {
	this.m_lock.UnSafeLock("dbPlayerTable.fetch_new_rows")
	defer this.m_lock.UnSafeUnlock()
	new_rows = make(map[int32]*dbPlayerRow)
	for i, v := range this.m_new_rows {
		_, has := this.m_rows[i]
		if has {
			log.Error("rows already has new rows %v", i)
			continue
		}
		this.m_rows[i] = v
		new_rows[i] = v
	}
	for i, _ := range new_rows {
		delete(this.m_new_rows, i)
	}
	return
}
func (this *dbPlayerTable) save_rows(rows map[int32]*dbPlayerRow, quick bool) {
	for _, v := range rows {
		if this.m_dbc.m_quit && !quick {
			return
		}
		err, delay, _ := v.Save(false)
		if err != nil {
			log.Error("save failed %v", err)
		}
		if this.m_dbc.m_quit && !quick {
			return
		}
		if delay&&!quick {
			time.Sleep(time.Millisecond * 5)
		}
	}
}
func (this *dbPlayerTable) Save(quick bool) (err error){
	removed_rows := this.fetch_rows(this.m_removed_rows)
	for _, v := range removed_rows {
		_, err := this.m_dbc.StmtExec(this.m_delete_stmt, v.GetPlayerId())
		if err != nil {
			log.Error("exec delete stmt failed %v", err)
		}
		v.m_valid = false
		if !quick {
			time.Sleep(time.Millisecond * 5)
		}
	}
	this.m_removed_rows = make(map[int32]*dbPlayerRow)
	rows := this.fetch_rows(this.m_rows)
	this.save_rows(rows, quick)
	new_rows := this.fetch_new_rows()
	this.save_rows(new_rows, quick)
	return
}
func (this *dbPlayerTable) AddRow(PlayerId int32) (row *dbPlayerRow) {
	this.m_lock.UnSafeLock("dbPlayerTable.AddRow")
	defer this.m_lock.UnSafeUnlock()
	row = new_dbPlayerRow(this,PlayerId)
	row.m_new = true
	row.m_loaded = true
	row.m_valid = true
	_, has := this.m_new_rows[PlayerId]
	if has{
		log.Error("已经存在 %v", PlayerId)
		return nil
	}
	this.m_new_rows[PlayerId] = row
	atomic.AddInt32(&this.m_gc_n,1)
	return row
}
func (this *dbPlayerTable) RemoveRow(PlayerId int32) {
	this.m_lock.UnSafeLock("dbPlayerTable.RemoveRow")
	defer this.m_lock.UnSafeUnlock()
	row := this.m_rows[PlayerId]
	if row != nil {
		row.m_remove = true
		delete(this.m_rows, PlayerId)
		rm_row := this.m_removed_rows[PlayerId]
		if rm_row != nil {
			log.Error("rows and removed rows both has %v", PlayerId)
		}
		this.m_removed_rows[PlayerId] = row
		_, has_new := this.m_new_rows[PlayerId]
		if has_new {
			delete(this.m_new_rows, PlayerId)
			log.Error("rows and new_rows both has %v", PlayerId)
		}
	} else {
		row = this.m_removed_rows[PlayerId]
		if row == nil {
			_, has_new := this.m_new_rows[PlayerId]
			if has_new {
				delete(this.m_new_rows, PlayerId)
			} else {
				log.Error("row not exist %v", PlayerId)
			}
		} else {
			log.Error("already removed %v", PlayerId)
			_, has_new := this.m_new_rows[PlayerId]
			if has_new {
				delete(this.m_new_rows, PlayerId)
				log.Error("removed rows and new_rows both has %v", PlayerId)
			}
		}
	}
}
func (this *dbPlayerTable) GetRow(PlayerId int32) (row *dbPlayerRow) {
	this.m_lock.UnSafeRLock("dbPlayerTable.GetRow")
	defer this.m_lock.UnSafeRUnlock()
	row = this.m_rows[PlayerId]
	if row == nil {
		row = this.m_new_rows[PlayerId]
	}
	return row
}
func (this *dbGooglePayRecordRow)GetSn( )(r string ){
	this.m_lock.UnSafeRLock("dbGooglePayRecordRow.GetdbGooglePayRecordSnColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_Sn)
}
func (this *dbGooglePayRecordRow)SetSn(v string){
	this.m_lock.UnSafeLock("dbGooglePayRecordRow.SetdbGooglePayRecordSnColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_Sn=string(v)
	this.m_Sn_changed=true
	return
}
func (this *dbGooglePayRecordRow)GetBid( )(r string ){
	this.m_lock.UnSafeRLock("dbGooglePayRecordRow.GetdbGooglePayRecordBidColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_Bid)
}
func (this *dbGooglePayRecordRow)SetBid(v string){
	this.m_lock.UnSafeLock("dbGooglePayRecordRow.SetdbGooglePayRecordBidColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_Bid=string(v)
	this.m_Bid_changed=true
	return
}
func (this *dbGooglePayRecordRow)GetPlayerId( )(r int32 ){
	this.m_lock.UnSafeRLock("dbGooglePayRecordRow.GetdbGooglePayRecordPlayerIdColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_PlayerId)
}
func (this *dbGooglePayRecordRow)SetPlayerId(v int32){
	this.m_lock.UnSafeLock("dbGooglePayRecordRow.SetdbGooglePayRecordPlayerIdColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_PlayerId=int32(v)
	this.m_PlayerId_changed=true
	return
}
func (this *dbGooglePayRecordRow)GetPayTime( )(r int32 ){
	this.m_lock.UnSafeRLock("dbGooglePayRecordRow.GetdbGooglePayRecordPayTimeColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_PayTime)
}
func (this *dbGooglePayRecordRow)SetPayTime(v int32){
	this.m_lock.UnSafeLock("dbGooglePayRecordRow.SetdbGooglePayRecordPayTimeColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_PayTime=int32(v)
	this.m_PayTime_changed=true
	return
}
type dbGooglePayRecordRow struct {
	m_table *dbGooglePayRecordTable
	m_lock       *RWMutex
	m_loaded  bool
	m_new     bool
	m_remove  bool
	m_touch      int32
	m_releasable bool
	m_valid   bool
	m_KeyId        int32
	m_Sn_changed bool
	m_Sn string
	m_Bid_changed bool
	m_Bid string
	m_PlayerId_changed bool
	m_PlayerId int32
	m_PayTime_changed bool
	m_PayTime int32
}
func new_dbGooglePayRecordRow(table *dbGooglePayRecordTable, KeyId int32) (r *dbGooglePayRecordRow) {
	this := &dbGooglePayRecordRow{}
	this.m_table = table
	this.m_KeyId = KeyId
	this.m_lock = NewRWMutex()
	this.m_Sn_changed=true
	this.m_Bid_changed=true
	this.m_PlayerId_changed=true
	this.m_PayTime_changed=true
	return this
}
func (this *dbGooglePayRecordRow) GetKeyId() (r int32) {
	return this.m_KeyId
}
func (this *dbGooglePayRecordRow) Load() (err error) {
	this.m_table.GC()
	this.m_lock.UnSafeLock("dbGooglePayRecordRow.Load")
	defer this.m_lock.UnSafeUnlock()
	if this.m_loaded {
		return
	}
	var dBid string
	var dPlayerId int32
	var dPayTime int32
	r := this.m_table.m_dbc.StmtQueryRow(this.m_table.m_load_select_stmt, this.m_KeyId)
	err = r.Scan(&dBid,&dPlayerId,&dPayTime)
	if err != nil {
		log.Error("scan")
		return
	}
		this.m_Bid=dBid
		this.m_PlayerId=dPlayerId
		this.m_PayTime=dPayTime
	this.m_loaded=true
	this.m_Bid_changed=false
	this.m_PlayerId_changed=false
	this.m_PayTime_changed=false
	this.Touch(false)
	atomic.AddInt32(&this.m_table.m_gc_n,1)
	return
}
func (this *dbGooglePayRecordRow) save_data(release bool) (err error, released bool, state int32, update_string string, args []interface{}) {
	this.m_lock.UnSafeLock("dbGooglePayRecordRow.save_data")
	defer this.m_lock.UnSafeUnlock()
	if this.m_new {
		db_args:=new_db_args(5)
		db_args.Push(this.m_KeyId)
		db_args.Push(this.m_Sn)
		db_args.Push(this.m_Bid)
		db_args.Push(this.m_PlayerId)
		db_args.Push(this.m_PayTime)
		args=db_args.GetArgs()
		state = 1
	} else {
		if this.m_Sn_changed||this.m_Bid_changed||this.m_PlayerId_changed||this.m_PayTime_changed{
			update_string = "UPDATE GooglePayRecords SET "
			db_args:=new_db_args(5)
			if this.m_Sn_changed{
				update_string+="Sn=?,"
				db_args.Push(this.m_Sn)
			}
			if this.m_Bid_changed{
				update_string+="Bid=?,"
				db_args.Push(this.m_Bid)
			}
			if this.m_PlayerId_changed{
				update_string+="PlayerId=?,"
				db_args.Push(this.m_PlayerId)
			}
			if this.m_PayTime_changed{
				update_string+="PayTime=?,"
				db_args.Push(this.m_PayTime)
			}
			update_string = strings.TrimRight(update_string, ", ")
			update_string+=" WHERE KeyId=?"
			db_args.Push(this.m_KeyId)
			args=db_args.GetArgs()
			state = 2
		}
	}
	this.m_new = false
	this.m_Sn_changed = false
	this.m_Bid_changed = false
	this.m_PlayerId_changed = false
	this.m_PayTime_changed = false
	if release && this.m_loaded {
		atomic.AddInt32(&this.m_table.m_gc_n, -1)
		this.m_loaded = false
		released = true
	}
	return nil,released,state,update_string,args
}
func (this *dbGooglePayRecordRow) Save(release bool) (err error, d bool, released bool) {
	err,released, state, update_string, args := this.save_data(release)
	if err != nil {
		log.Error("save data failed")
		return err, false, false
	}
	if state == 0 {
		d = false
	} else if state == 1 {
		_, err = this.m_table.m_dbc.StmtExec(this.m_table.m_save_insert_stmt, args...)
		if err != nil {
			log.Error("INSERT GooglePayRecords exec failed %v ", this.m_KeyId)
			return err, false, released
		}
		d = true
	} else if state == 2 {
		_, err = this.m_table.m_dbc.Exec(update_string, args...)
		if err != nil {
			log.Error("UPDATE GooglePayRecords exec failed %v", this.m_KeyId)
			return err, false, released
		}
		d = true
	}
	return nil, d, released
}
func (this *dbGooglePayRecordRow) Touch(releasable bool) {
	this.m_touch = int32(time.Now().Unix())
	this.m_releasable = releasable
}
type dbGooglePayRecordRowSort struct {
	rows []*dbGooglePayRecordRow
}
func (this *dbGooglePayRecordRowSort) Len() (length int) {
	return len(this.rows)
}
func (this *dbGooglePayRecordRowSort) Less(i int, j int) (less bool) {
	return this.rows[i].m_touch < this.rows[j].m_touch
}
func (this *dbGooglePayRecordRowSort) Swap(i int, j int) {
	temp := this.rows[i]
	this.rows[i] = this.rows[j]
	this.rows[j] = temp
}
type dbGooglePayRecordTable struct{
	m_dbc *DBC
	m_lock *RWMutex
	m_rows map[int32]*dbGooglePayRecordRow
	m_new_rows map[int32]*dbGooglePayRecordRow
	m_removed_rows map[int32]*dbGooglePayRecordRow
	m_gc_n int32
	m_gcing int32
	m_pool_size int32
	m_preload_select_stmt *sql.Stmt
	m_preload_max_id int32
	m_load_select_stmt *sql.Stmt
	m_save_insert_stmt *sql.Stmt
	m_delete_stmt *sql.Stmt
	m_max_id int32
	m_max_id_changed bool
}
func new_dbGooglePayRecordTable(dbc *DBC) (this *dbGooglePayRecordTable) {
	this = &dbGooglePayRecordTable{}
	this.m_dbc = dbc
	this.m_lock = NewRWMutex()
	this.m_rows = make(map[int32]*dbGooglePayRecordRow)
	this.m_new_rows = make(map[int32]*dbGooglePayRecordRow)
	this.m_removed_rows = make(map[int32]*dbGooglePayRecordRow)
	return this
}
func (this *dbGooglePayRecordTable) check_create_table() (err error) {
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS GooglePayRecordsMaxId(PlaceHolder int(11),MaxKeyId int(11),PRIMARY KEY (PlaceHolder))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS GooglePayRecordsMaxId failed")
		return
	}
	r := this.m_dbc.QueryRow("SELECT Count(*) FROM GooglePayRecordsMaxId WHERE PlaceHolder=0")
	if r != nil {
		var count int32
		err = r.Scan(&count)
		if err != nil {
			log.Error("scan count failed")
			return
		}
		if count == 0 {
		_, err = this.m_dbc.Exec("INSERT INTO GooglePayRecordsMaxId (PlaceHolder,MaxKeyId) VALUES (0,0)")
			if err != nil {
				log.Error("INSERTGooglePayRecordsMaxId failed")
				return
			}
		}
	}
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS GooglePayRecords(KeyId int(11),PRIMARY KEY (KeyId))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS GooglePayRecords failed")
		return
	}
	rows, err := this.m_dbc.Query("SELECT COLUMN_NAME,ORDINAL_POSITION FROM information_schema.`COLUMNS` WHERE TABLE_SCHEMA=? AND TABLE_NAME='GooglePayRecords'", this.m_dbc.m_db_name)
	if err != nil {
		log.Error("SELECT information_schema failed")
		return
	}
	columns := make(map[string]int32)
	for rows.Next() {
		var column_name string
		var ordinal_position int32
		err = rows.Scan(&column_name, &ordinal_position)
		if err != nil {
			log.Error("scan information_schema row failed")
			return
		}
		if ordinal_position < 1 {
			log.Error("col ordinal out of range")
			continue
		}
		columns[column_name] = ordinal_position
	}
	_, hasSn := columns["Sn"]
	if !hasSn {
		_, err = this.m_dbc.Exec("ALTER TABLE GooglePayRecords ADD COLUMN Sn varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN Sn failed")
			return
		}
	}
	_, hasBid := columns["Bid"]
	if !hasBid {
		_, err = this.m_dbc.Exec("ALTER TABLE GooglePayRecords ADD COLUMN Bid varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN Bid failed")
			return
		}
	}
	_, hasPlayerId := columns["PlayerId"]
	if !hasPlayerId {
		_, err = this.m_dbc.Exec("ALTER TABLE GooglePayRecords ADD COLUMN PlayerId int(11)")
		if err != nil {
			log.Error("ADD COLUMN PlayerId failed")
			return
		}
	}
	_, hasPayTime := columns["PayTime"]
	if !hasPayTime {
		_, err = this.m_dbc.Exec("ALTER TABLE GooglePayRecords ADD COLUMN PayTime int(11)")
		if err != nil {
			log.Error("ADD COLUMN PayTime failed")
			return
		}
	}
	return
}
func (this *dbGooglePayRecordTable) prepare_preload_select_stmt() (err error) {
	this.m_preload_select_stmt,err=this.m_dbc.StmtPrepare("SELECT KeyId,Sn FROM GooglePayRecords")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbGooglePayRecordTable) prepare_load_select_stmt() (err error) {
	this.m_load_select_stmt,err=this.m_dbc.StmtPrepare("SELECT Bid,PlayerId,PayTime FROM GooglePayRecords WHERE KeyId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbGooglePayRecordTable) prepare_save_insert_stmt()(err error){
	this.m_save_insert_stmt,err=this.m_dbc.StmtPrepare("INSERT INTO GooglePayRecords (KeyId,Sn,Bid,PlayerId,PayTime) VALUES (?,?,?,?,?)")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbGooglePayRecordTable) prepare_delete_stmt() (err error) {
	this.m_delete_stmt,err=this.m_dbc.StmtPrepare("DELETE FROM GooglePayRecords WHERE KeyId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbGooglePayRecordTable) Init() (err error) {
	err=this.check_create_table()
	if err!=nil{
		log.Error("check_create_table failed")
		return
	}
	err=this.prepare_preload_select_stmt()
	if err!=nil{
		log.Error("prepare_preload_select_stmt failed")
		return
	}
	err=this.prepare_load_select_stmt()
	if err!=nil{
		log.Error("prepare_load_select_stmt failed")
		return
	}
	err=this.prepare_save_insert_stmt()
	if err!=nil{
		log.Error("prepare_save_insert_stmt failed")
		return
	}
	err=this.prepare_delete_stmt()
	if err!=nil{
		log.Error("prepare_save_insert_stmt failed")
		return
	}
	return
}
func (this *dbGooglePayRecordTable) Preload() (err error) {
	r_max_id := this.m_dbc.QueryRow("SELECT MaxKeyId FROM GooglePayRecordsMaxId WHERE PLACEHOLDER=0")
	if r_max_id != nil {
		err = r_max_id.Scan(&this.m_max_id)
		if err != nil {
			log.Error("scan max id failed")
			return
		}
	}
	r, err := this.m_dbc.StmtQuery(this.m_preload_select_stmt)
	if err != nil {
		log.Error("SELECT")
		return
	}
	var KeyId int32
	var dSn string
	for r.Next() {
		err = r.Scan(&KeyId,&dSn)
		if err != nil {
			log.Error("Scan")
			return
		}
		if KeyId>this.m_max_id{
			log.Error("max id ext")
			this.m_max_id = KeyId
			this.m_max_id_changed = true
		}
		row := new_dbGooglePayRecordRow(this,KeyId)
		row.m_Sn=dSn
		row.m_Sn_changed=false
		row.m_valid = true
		this.m_rows[KeyId]=row
	}
	return
}
func (this *dbGooglePayRecordTable) GetPreloadedMaxId() (max_id int32) {
	return this.m_preload_max_id
}
func (this *dbGooglePayRecordTable) fetch_rows(rows map[int32]*dbGooglePayRecordRow) (r map[int32]*dbGooglePayRecordRow) {
	this.m_lock.UnSafeLock("dbGooglePayRecordTable.fetch_rows")
	defer this.m_lock.UnSafeUnlock()
	r = make(map[int32]*dbGooglePayRecordRow)
	for i, v := range rows {
		r[i] = v
	}
	return r
}
func (this *dbGooglePayRecordTable) fetch_new_rows() (new_rows map[int32]*dbGooglePayRecordRow) {
	this.m_lock.UnSafeLock("dbGooglePayRecordTable.fetch_new_rows")
	defer this.m_lock.UnSafeUnlock()
	new_rows = make(map[int32]*dbGooglePayRecordRow)
	for i, v := range this.m_new_rows {
		_, has := this.m_rows[i]
		if has {
			log.Error("rows already has new rows %v", i)
			continue
		}
		this.m_rows[i] = v
		new_rows[i] = v
	}
	for i, _ := range new_rows {
		delete(this.m_new_rows, i)
	}
	return
}
func (this *dbGooglePayRecordTable) save_rows(rows map[int32]*dbGooglePayRecordRow, quick bool) {
	for _, v := range rows {
		if this.m_dbc.m_quit && !quick {
			return
		}
		err, delay, _ := v.Save(false)
		if err != nil {
			log.Error("save failed %v", err)
		}
		if this.m_dbc.m_quit && !quick {
			return
		}
		if delay&&!quick {
			time.Sleep(time.Millisecond * 5)
		}
	}
}
func (this *dbGooglePayRecordTable) Save(quick bool) (err error){
	if this.m_max_id_changed {
		max_id := atomic.LoadInt32(&this.m_max_id)
		_, err := this.m_dbc.Exec("UPDATE GooglePayRecordsMaxId SET MaxKeyId=?", max_id)
		if err != nil {
			log.Error("save max id failed %v", err)
		}
	}
	removed_rows := this.fetch_rows(this.m_removed_rows)
	for _, v := range removed_rows {
		_, err := this.m_dbc.StmtExec(this.m_delete_stmt, v.GetKeyId())
		if err != nil {
			log.Error("exec delete stmt failed %v", err)
		}
		v.m_valid = false
		if !quick {
			time.Sleep(time.Millisecond * 5)
		}
	}
	this.m_removed_rows = make(map[int32]*dbGooglePayRecordRow)
	rows := this.fetch_rows(this.m_rows)
	this.save_rows(rows, quick)
	new_rows := this.fetch_new_rows()
	this.save_rows(new_rows, quick)
	return
}
func (this *dbGooglePayRecordTable) AddRow() (row *dbGooglePayRecordRow) {
	this.GC()
	this.m_lock.UnSafeLock("dbGooglePayRecordTable.AddRow")
	defer this.m_lock.UnSafeUnlock()
	KeyId := atomic.AddInt32(&this.m_max_id, 1)
	this.m_max_id_changed = true
	row = new_dbGooglePayRecordRow(this,KeyId)
	row.m_new = true
	row.m_loaded = true
	row.m_valid = true
	this.m_new_rows[KeyId] = row
	atomic.AddInt32(&this.m_gc_n,1)
	return row
}
func (this *dbGooglePayRecordTable) RemoveRow(KeyId int32) {
	this.m_lock.UnSafeLock("dbGooglePayRecordTable.RemoveRow")
	defer this.m_lock.UnSafeUnlock()
	row := this.m_rows[KeyId]
	if row != nil {
		row.m_remove = true
		delete(this.m_rows, KeyId)
		rm_row := this.m_removed_rows[KeyId]
		if rm_row != nil {
			log.Error("rows and removed rows both has %v", KeyId)
		}
		this.m_removed_rows[KeyId] = row
		_, has_new := this.m_new_rows[KeyId]
		if has_new {
			delete(this.m_new_rows, KeyId)
			log.Error("rows and new_rows both has %v", KeyId)
		}
	} else {
		row = this.m_removed_rows[KeyId]
		if row == nil {
			_, has_new := this.m_new_rows[KeyId]
			if has_new {
				delete(this.m_new_rows, KeyId)
			} else {
				log.Error("row not exist %v", KeyId)
			}
		} else {
			log.Error("already removed %v", KeyId)
			_, has_new := this.m_new_rows[KeyId]
			if has_new {
				delete(this.m_new_rows, KeyId)
				log.Error("removed rows and new_rows both has %v", KeyId)
			}
		}
	}
}
func (this *dbGooglePayRecordTable) GetRow(KeyId int32) (row *dbGooglePayRecordRow) {
	this.m_lock.UnSafeRLock("dbGooglePayRecordTable.GetRow")
	defer this.m_lock.UnSafeRUnlock()
	row = this.m_rows[KeyId]
	if row == nil {
		row = this.m_new_rows[KeyId]
	}
	return row
}
func (this *dbGooglePayRecordTable) SetPoolSize(n int32) {
	this.m_pool_size = n
}
func (this *dbGooglePayRecordTable) GC() {
	if this.m_pool_size<=0{
		return
	}
	if !atomic.CompareAndSwapInt32(&this.m_gcing, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&this.m_gcing, 0)
	n := atomic.LoadInt32(&this.m_gc_n)
	if float32(n) < float32(this.m_pool_size)*1.2 {
		return
	}
	max := (n - this.m_pool_size) / 2
	arr := dbGooglePayRecordRowSort{}
	rows := this.fetch_rows(this.m_rows)
	arr.rows = make([]*dbGooglePayRecordRow, len(rows))
	index := 0
	for _, v := range rows {
		arr.rows[index] = v
		index++
	}
	sort.Sort(&arr)
	count := int32(0)
	for _, v := range arr.rows {
		err, _, released := v.Save(true)
		if err != nil {
			log.Error("release failed %v", err)
			continue
		}
		if released {
			count++
			if count > max {
				return
			}
		}
	}
	return
}
func (this *dbApplePayRecordRow)GetSn( )(r string ){
	this.m_lock.UnSafeRLock("dbApplePayRecordRow.GetdbApplePayRecordSnColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_Sn)
}
func (this *dbApplePayRecordRow)SetSn(v string){
	this.m_lock.UnSafeLock("dbApplePayRecordRow.SetdbApplePayRecordSnColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_Sn=string(v)
	this.m_Sn_changed=true
	return
}
func (this *dbApplePayRecordRow)GetBid( )(r string ){
	this.m_lock.UnSafeRLock("dbApplePayRecordRow.GetdbApplePayRecordBidColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_Bid)
}
func (this *dbApplePayRecordRow)SetBid(v string){
	this.m_lock.UnSafeLock("dbApplePayRecordRow.SetdbApplePayRecordBidColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_Bid=string(v)
	this.m_Bid_changed=true
	return
}
func (this *dbApplePayRecordRow)GetPlayerId( )(r int32 ){
	this.m_lock.UnSafeRLock("dbApplePayRecordRow.GetdbApplePayRecordPlayerIdColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_PlayerId)
}
func (this *dbApplePayRecordRow)SetPlayerId(v int32){
	this.m_lock.UnSafeLock("dbApplePayRecordRow.SetdbApplePayRecordPlayerIdColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_PlayerId=int32(v)
	this.m_PlayerId_changed=true
	return
}
func (this *dbApplePayRecordRow)GetPayTime( )(r int32 ){
	this.m_lock.UnSafeRLock("dbApplePayRecordRow.GetdbApplePayRecordPayTimeColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_PayTime)
}
func (this *dbApplePayRecordRow)SetPayTime(v int32){
	this.m_lock.UnSafeLock("dbApplePayRecordRow.SetdbApplePayRecordPayTimeColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_PayTime=int32(v)
	this.m_PayTime_changed=true
	return
}
type dbApplePayRecordRow struct {
	m_table *dbApplePayRecordTable
	m_lock       *RWMutex
	m_loaded  bool
	m_new     bool
	m_remove  bool
	m_touch      int32
	m_releasable bool
	m_valid   bool
	m_KeyId        int32
	m_Sn_changed bool
	m_Sn string
	m_Bid_changed bool
	m_Bid string
	m_PlayerId_changed bool
	m_PlayerId int32
	m_PayTime_changed bool
	m_PayTime int32
}
func new_dbApplePayRecordRow(table *dbApplePayRecordTable, KeyId int32) (r *dbApplePayRecordRow) {
	this := &dbApplePayRecordRow{}
	this.m_table = table
	this.m_KeyId = KeyId
	this.m_lock = NewRWMutex()
	this.m_Sn_changed=true
	this.m_Bid_changed=true
	this.m_PlayerId_changed=true
	this.m_PayTime_changed=true
	return this
}
func (this *dbApplePayRecordRow) GetKeyId() (r int32) {
	return this.m_KeyId
}
func (this *dbApplePayRecordRow) Load() (err error) {
	this.m_table.GC()
	this.m_lock.UnSafeLock("dbApplePayRecordRow.Load")
	defer this.m_lock.UnSafeUnlock()
	if this.m_loaded {
		return
	}
	var dBid string
	var dPlayerId int32
	var dPayTime int32
	r := this.m_table.m_dbc.StmtQueryRow(this.m_table.m_load_select_stmt, this.m_KeyId)
	err = r.Scan(&dBid,&dPlayerId,&dPayTime)
	if err != nil {
		log.Error("scan")
		return
	}
		this.m_Bid=dBid
		this.m_PlayerId=dPlayerId
		this.m_PayTime=dPayTime
	this.m_loaded=true
	this.m_Bid_changed=false
	this.m_PlayerId_changed=false
	this.m_PayTime_changed=false
	this.Touch(false)
	atomic.AddInt32(&this.m_table.m_gc_n,1)
	return
}
func (this *dbApplePayRecordRow) save_data(release bool) (err error, released bool, state int32, update_string string, args []interface{}) {
	this.m_lock.UnSafeLock("dbApplePayRecordRow.save_data")
	defer this.m_lock.UnSafeUnlock()
	if this.m_new {
		db_args:=new_db_args(5)
		db_args.Push(this.m_KeyId)
		db_args.Push(this.m_Sn)
		db_args.Push(this.m_Bid)
		db_args.Push(this.m_PlayerId)
		db_args.Push(this.m_PayTime)
		args=db_args.GetArgs()
		state = 1
	} else {
		if this.m_Sn_changed||this.m_Bid_changed||this.m_PlayerId_changed||this.m_PayTime_changed{
			update_string = "UPDATE ApplePayRecords SET "
			db_args:=new_db_args(5)
			if this.m_Sn_changed{
				update_string+="Sn=?,"
				db_args.Push(this.m_Sn)
			}
			if this.m_Bid_changed{
				update_string+="Bid=?,"
				db_args.Push(this.m_Bid)
			}
			if this.m_PlayerId_changed{
				update_string+="PlayerId=?,"
				db_args.Push(this.m_PlayerId)
			}
			if this.m_PayTime_changed{
				update_string+="PayTime=?,"
				db_args.Push(this.m_PayTime)
			}
			update_string = strings.TrimRight(update_string, ", ")
			update_string+=" WHERE KeyId=?"
			db_args.Push(this.m_KeyId)
			args=db_args.GetArgs()
			state = 2
		}
	}
	this.m_new = false
	this.m_Sn_changed = false
	this.m_Bid_changed = false
	this.m_PlayerId_changed = false
	this.m_PayTime_changed = false
	if release && this.m_loaded {
		atomic.AddInt32(&this.m_table.m_gc_n, -1)
		this.m_loaded = false
		released = true
	}
	return nil,released,state,update_string,args
}
func (this *dbApplePayRecordRow) Save(release bool) (err error, d bool, released bool) {
	err,released, state, update_string, args := this.save_data(release)
	if err != nil {
		log.Error("save data failed")
		return err, false, false
	}
	if state == 0 {
		d = false
	} else if state == 1 {
		_, err = this.m_table.m_dbc.StmtExec(this.m_table.m_save_insert_stmt, args...)
		if err != nil {
			log.Error("INSERT ApplePayRecords exec failed %v ", this.m_KeyId)
			return err, false, released
		}
		d = true
	} else if state == 2 {
		_, err = this.m_table.m_dbc.Exec(update_string, args...)
		if err != nil {
			log.Error("UPDATE ApplePayRecords exec failed %v", this.m_KeyId)
			return err, false, released
		}
		d = true
	}
	return nil, d, released
}
func (this *dbApplePayRecordRow) Touch(releasable bool) {
	this.m_touch = int32(time.Now().Unix())
	this.m_releasable = releasable
}
type dbApplePayRecordRowSort struct {
	rows []*dbApplePayRecordRow
}
func (this *dbApplePayRecordRowSort) Len() (length int) {
	return len(this.rows)
}
func (this *dbApplePayRecordRowSort) Less(i int, j int) (less bool) {
	return this.rows[i].m_touch < this.rows[j].m_touch
}
func (this *dbApplePayRecordRowSort) Swap(i int, j int) {
	temp := this.rows[i]
	this.rows[i] = this.rows[j]
	this.rows[j] = temp
}
type dbApplePayRecordTable struct{
	m_dbc *DBC
	m_lock *RWMutex
	m_rows map[int32]*dbApplePayRecordRow
	m_new_rows map[int32]*dbApplePayRecordRow
	m_removed_rows map[int32]*dbApplePayRecordRow
	m_gc_n int32
	m_gcing int32
	m_pool_size int32
	m_preload_select_stmt *sql.Stmt
	m_preload_max_id int32
	m_load_select_stmt *sql.Stmt
	m_save_insert_stmt *sql.Stmt
	m_delete_stmt *sql.Stmt
	m_max_id int32
	m_max_id_changed bool
}
func new_dbApplePayRecordTable(dbc *DBC) (this *dbApplePayRecordTable) {
	this = &dbApplePayRecordTable{}
	this.m_dbc = dbc
	this.m_lock = NewRWMutex()
	this.m_rows = make(map[int32]*dbApplePayRecordRow)
	this.m_new_rows = make(map[int32]*dbApplePayRecordRow)
	this.m_removed_rows = make(map[int32]*dbApplePayRecordRow)
	return this
}
func (this *dbApplePayRecordTable) check_create_table() (err error) {
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS ApplePayRecordsMaxId(PlaceHolder int(11),MaxKeyId int(11),PRIMARY KEY (PlaceHolder))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS ApplePayRecordsMaxId failed")
		return
	}
	r := this.m_dbc.QueryRow("SELECT Count(*) FROM ApplePayRecordsMaxId WHERE PlaceHolder=0")
	if r != nil {
		var count int32
		err = r.Scan(&count)
		if err != nil {
			log.Error("scan count failed")
			return
		}
		if count == 0 {
		_, err = this.m_dbc.Exec("INSERT INTO ApplePayRecordsMaxId (PlaceHolder,MaxKeyId) VALUES (0,0)")
			if err != nil {
				log.Error("INSERTApplePayRecordsMaxId failed")
				return
			}
		}
	}
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS ApplePayRecords(KeyId int(11),PRIMARY KEY (KeyId))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS ApplePayRecords failed")
		return
	}
	rows, err := this.m_dbc.Query("SELECT COLUMN_NAME,ORDINAL_POSITION FROM information_schema.`COLUMNS` WHERE TABLE_SCHEMA=? AND TABLE_NAME='ApplePayRecords'", this.m_dbc.m_db_name)
	if err != nil {
		log.Error("SELECT information_schema failed")
		return
	}
	columns := make(map[string]int32)
	for rows.Next() {
		var column_name string
		var ordinal_position int32
		err = rows.Scan(&column_name, &ordinal_position)
		if err != nil {
			log.Error("scan information_schema row failed")
			return
		}
		if ordinal_position < 1 {
			log.Error("col ordinal out of range")
			continue
		}
		columns[column_name] = ordinal_position
	}
	_, hasSn := columns["Sn"]
	if !hasSn {
		_, err = this.m_dbc.Exec("ALTER TABLE ApplePayRecords ADD COLUMN Sn varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN Sn failed")
			return
		}
	}
	_, hasBid := columns["Bid"]
	if !hasBid {
		_, err = this.m_dbc.Exec("ALTER TABLE ApplePayRecords ADD COLUMN Bid varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN Bid failed")
			return
		}
	}
	_, hasPlayerId := columns["PlayerId"]
	if !hasPlayerId {
		_, err = this.m_dbc.Exec("ALTER TABLE ApplePayRecords ADD COLUMN PlayerId int(11)")
		if err != nil {
			log.Error("ADD COLUMN PlayerId failed")
			return
		}
	}
	_, hasPayTime := columns["PayTime"]
	if !hasPayTime {
		_, err = this.m_dbc.Exec("ALTER TABLE ApplePayRecords ADD COLUMN PayTime int(11)")
		if err != nil {
			log.Error("ADD COLUMN PayTime failed")
			return
		}
	}
	return
}
func (this *dbApplePayRecordTable) prepare_preload_select_stmt() (err error) {
	this.m_preload_select_stmt,err=this.m_dbc.StmtPrepare("SELECT KeyId,Sn FROM ApplePayRecords")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbApplePayRecordTable) prepare_load_select_stmt() (err error) {
	this.m_load_select_stmt,err=this.m_dbc.StmtPrepare("SELECT Bid,PlayerId,PayTime FROM ApplePayRecords WHERE KeyId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbApplePayRecordTable) prepare_save_insert_stmt()(err error){
	this.m_save_insert_stmt,err=this.m_dbc.StmtPrepare("INSERT INTO ApplePayRecords (KeyId,Sn,Bid,PlayerId,PayTime) VALUES (?,?,?,?,?)")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbApplePayRecordTable) prepare_delete_stmt() (err error) {
	this.m_delete_stmt,err=this.m_dbc.StmtPrepare("DELETE FROM ApplePayRecords WHERE KeyId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbApplePayRecordTable) Init() (err error) {
	err=this.check_create_table()
	if err!=nil{
		log.Error("check_create_table failed")
		return
	}
	err=this.prepare_preload_select_stmt()
	if err!=nil{
		log.Error("prepare_preload_select_stmt failed")
		return
	}
	err=this.prepare_load_select_stmt()
	if err!=nil{
		log.Error("prepare_load_select_stmt failed")
		return
	}
	err=this.prepare_save_insert_stmt()
	if err!=nil{
		log.Error("prepare_save_insert_stmt failed")
		return
	}
	err=this.prepare_delete_stmt()
	if err!=nil{
		log.Error("prepare_save_insert_stmt failed")
		return
	}
	return
}
func (this *dbApplePayRecordTable) Preload() (err error) {
	r_max_id := this.m_dbc.QueryRow("SELECT MaxKeyId FROM ApplePayRecordsMaxId WHERE PLACEHOLDER=0")
	if r_max_id != nil {
		err = r_max_id.Scan(&this.m_max_id)
		if err != nil {
			log.Error("scan max id failed")
			return
		}
	}
	r, err := this.m_dbc.StmtQuery(this.m_preload_select_stmt)
	if err != nil {
		log.Error("SELECT")
		return
	}
	var KeyId int32
	var dSn string
	for r.Next() {
		err = r.Scan(&KeyId,&dSn)
		if err != nil {
			log.Error("Scan")
			return
		}
		if KeyId>this.m_max_id{
			log.Error("max id ext")
			this.m_max_id = KeyId
			this.m_max_id_changed = true
		}
		row := new_dbApplePayRecordRow(this,KeyId)
		row.m_Sn=dSn
		row.m_Sn_changed=false
		row.m_valid = true
		this.m_rows[KeyId]=row
	}
	return
}
func (this *dbApplePayRecordTable) GetPreloadedMaxId() (max_id int32) {
	return this.m_preload_max_id
}
func (this *dbApplePayRecordTable) fetch_rows(rows map[int32]*dbApplePayRecordRow) (r map[int32]*dbApplePayRecordRow) {
	this.m_lock.UnSafeLock("dbApplePayRecordTable.fetch_rows")
	defer this.m_lock.UnSafeUnlock()
	r = make(map[int32]*dbApplePayRecordRow)
	for i, v := range rows {
		r[i] = v
	}
	return r
}
func (this *dbApplePayRecordTable) fetch_new_rows() (new_rows map[int32]*dbApplePayRecordRow) {
	this.m_lock.UnSafeLock("dbApplePayRecordTable.fetch_new_rows")
	defer this.m_lock.UnSafeUnlock()
	new_rows = make(map[int32]*dbApplePayRecordRow)
	for i, v := range this.m_new_rows {
		_, has := this.m_rows[i]
		if has {
			log.Error("rows already has new rows %v", i)
			continue
		}
		this.m_rows[i] = v
		new_rows[i] = v
	}
	for i, _ := range new_rows {
		delete(this.m_new_rows, i)
	}
	return
}
func (this *dbApplePayRecordTable) save_rows(rows map[int32]*dbApplePayRecordRow, quick bool) {
	for _, v := range rows {
		if this.m_dbc.m_quit && !quick {
			return
		}
		err, delay, _ := v.Save(false)
		if err != nil {
			log.Error("save failed %v", err)
		}
		if this.m_dbc.m_quit && !quick {
			return
		}
		if delay&&!quick {
			time.Sleep(time.Millisecond * 5)
		}
	}
}
func (this *dbApplePayRecordTable) Save(quick bool) (err error){
	if this.m_max_id_changed {
		max_id := atomic.LoadInt32(&this.m_max_id)
		_, err := this.m_dbc.Exec("UPDATE ApplePayRecordsMaxId SET MaxKeyId=?", max_id)
		if err != nil {
			log.Error("save max id failed %v", err)
		}
	}
	removed_rows := this.fetch_rows(this.m_removed_rows)
	for _, v := range removed_rows {
		_, err := this.m_dbc.StmtExec(this.m_delete_stmt, v.GetKeyId())
		if err != nil {
			log.Error("exec delete stmt failed %v", err)
		}
		v.m_valid = false
		if !quick {
			time.Sleep(time.Millisecond * 5)
		}
	}
	this.m_removed_rows = make(map[int32]*dbApplePayRecordRow)
	rows := this.fetch_rows(this.m_rows)
	this.save_rows(rows, quick)
	new_rows := this.fetch_new_rows()
	this.save_rows(new_rows, quick)
	return
}
func (this *dbApplePayRecordTable) AddRow() (row *dbApplePayRecordRow) {
	this.GC()
	this.m_lock.UnSafeLock("dbApplePayRecordTable.AddRow")
	defer this.m_lock.UnSafeUnlock()
	KeyId := atomic.AddInt32(&this.m_max_id, 1)
	this.m_max_id_changed = true
	row = new_dbApplePayRecordRow(this,KeyId)
	row.m_new = true
	row.m_loaded = true
	row.m_valid = true
	this.m_new_rows[KeyId] = row
	atomic.AddInt32(&this.m_gc_n,1)
	return row
}
func (this *dbApplePayRecordTable) RemoveRow(KeyId int32) {
	this.m_lock.UnSafeLock("dbApplePayRecordTable.RemoveRow")
	defer this.m_lock.UnSafeUnlock()
	row := this.m_rows[KeyId]
	if row != nil {
		row.m_remove = true
		delete(this.m_rows, KeyId)
		rm_row := this.m_removed_rows[KeyId]
		if rm_row != nil {
			log.Error("rows and removed rows both has %v", KeyId)
		}
		this.m_removed_rows[KeyId] = row
		_, has_new := this.m_new_rows[KeyId]
		if has_new {
			delete(this.m_new_rows, KeyId)
			log.Error("rows and new_rows both has %v", KeyId)
		}
	} else {
		row = this.m_removed_rows[KeyId]
		if row == nil {
			_, has_new := this.m_new_rows[KeyId]
			if has_new {
				delete(this.m_new_rows, KeyId)
			} else {
				log.Error("row not exist %v", KeyId)
			}
		} else {
			log.Error("already removed %v", KeyId)
			_, has_new := this.m_new_rows[KeyId]
			if has_new {
				delete(this.m_new_rows, KeyId)
				log.Error("removed rows and new_rows both has %v", KeyId)
			}
		}
	}
}
func (this *dbApplePayRecordTable) GetRow(KeyId int32) (row *dbApplePayRecordRow) {
	this.m_lock.UnSafeRLock("dbApplePayRecordTable.GetRow")
	defer this.m_lock.UnSafeRUnlock()
	row = this.m_rows[KeyId]
	if row == nil {
		row = this.m_new_rows[KeyId]
	}
	return row
}
func (this *dbApplePayRecordTable) SetPoolSize(n int32) {
	this.m_pool_size = n
}
func (this *dbApplePayRecordTable) GC() {
	if this.m_pool_size<=0{
		return
	}
	if !atomic.CompareAndSwapInt32(&this.m_gcing, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&this.m_gcing, 0)
	n := atomic.LoadInt32(&this.m_gc_n)
	if float32(n) < float32(this.m_pool_size)*1.2 {
		return
	}
	max := (n - this.m_pool_size) / 2
	arr := dbApplePayRecordRowSort{}
	rows := this.fetch_rows(this.m_rows)
	arr.rows = make([]*dbApplePayRecordRow, len(rows))
	index := 0
	for _, v := range rows {
		arr.rows[index] = v
		index++
	}
	sort.Sort(&arr)
	count := int32(0)
	for _, v := range arr.rows {
		err, _, released := v.Save(true)
		if err != nil {
			log.Error("release failed %v", err)
			continue
		}
		if released {
			count++
			if count > max {
				return
			}
		}
	}
	return
}
func (this *dbFaceBPayRecordRow)GetSn( )(r string ){
	this.m_lock.UnSafeRLock("dbFaceBPayRecordRow.GetdbFaceBPayRecordSnColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_Sn)
}
func (this *dbFaceBPayRecordRow)SetSn(v string){
	this.m_lock.UnSafeLock("dbFaceBPayRecordRow.SetdbFaceBPayRecordSnColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_Sn=string(v)
	this.m_Sn_changed=true
	return
}
func (this *dbFaceBPayRecordRow)GetBid( )(r string ){
	this.m_lock.UnSafeRLock("dbFaceBPayRecordRow.GetdbFaceBPayRecordBidColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_Bid)
}
func (this *dbFaceBPayRecordRow)SetBid(v string){
	this.m_lock.UnSafeLock("dbFaceBPayRecordRow.SetdbFaceBPayRecordBidColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_Bid=string(v)
	this.m_Bid_changed=true
	return
}
func (this *dbFaceBPayRecordRow)GetPlayerId( )(r int32 ){
	this.m_lock.UnSafeRLock("dbFaceBPayRecordRow.GetdbFaceBPayRecordPlayerIdColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_PlayerId)
}
func (this *dbFaceBPayRecordRow)SetPlayerId(v int32){
	this.m_lock.UnSafeLock("dbFaceBPayRecordRow.SetdbFaceBPayRecordPlayerIdColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_PlayerId=int32(v)
	this.m_PlayerId_changed=true
	return
}
func (this *dbFaceBPayRecordRow)GetPayTime( )(r int32 ){
	this.m_lock.UnSafeRLock("dbFaceBPayRecordRow.GetdbFaceBPayRecordPayTimeColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_PayTime)
}
func (this *dbFaceBPayRecordRow)SetPayTime(v int32){
	this.m_lock.UnSafeLock("dbFaceBPayRecordRow.SetdbFaceBPayRecordPayTimeColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_PayTime=int32(v)
	this.m_PayTime_changed=true
	return
}
type dbFaceBPayRecordRow struct {
	m_table *dbFaceBPayRecordTable
	m_lock       *RWMutex
	m_loaded  bool
	m_new     bool
	m_remove  bool
	m_touch      int32
	m_releasable bool
	m_valid   bool
	m_KeyId        int32
	m_Sn_changed bool
	m_Sn string
	m_Bid_changed bool
	m_Bid string
	m_PlayerId_changed bool
	m_PlayerId int32
	m_PayTime_changed bool
	m_PayTime int32
}
func new_dbFaceBPayRecordRow(table *dbFaceBPayRecordTable, KeyId int32) (r *dbFaceBPayRecordRow) {
	this := &dbFaceBPayRecordRow{}
	this.m_table = table
	this.m_KeyId = KeyId
	this.m_lock = NewRWMutex()
	this.m_Sn_changed=true
	this.m_Bid_changed=true
	this.m_PlayerId_changed=true
	this.m_PayTime_changed=true
	return this
}
func (this *dbFaceBPayRecordRow) GetKeyId() (r int32) {
	return this.m_KeyId
}
func (this *dbFaceBPayRecordRow) Load() (err error) {
	this.m_table.GC()
	this.m_lock.UnSafeLock("dbFaceBPayRecordRow.Load")
	defer this.m_lock.UnSafeUnlock()
	if this.m_loaded {
		return
	}
	var dBid string
	var dPlayerId int32
	var dPayTime int32
	r := this.m_table.m_dbc.StmtQueryRow(this.m_table.m_load_select_stmt, this.m_KeyId)
	err = r.Scan(&dBid,&dPlayerId,&dPayTime)
	if err != nil {
		log.Error("scan")
		return
	}
		this.m_Bid=dBid
		this.m_PlayerId=dPlayerId
		this.m_PayTime=dPayTime
	this.m_loaded=true
	this.m_Bid_changed=false
	this.m_PlayerId_changed=false
	this.m_PayTime_changed=false
	this.Touch(false)
	atomic.AddInt32(&this.m_table.m_gc_n,1)
	return
}
func (this *dbFaceBPayRecordRow) save_data(release bool) (err error, released bool, state int32, update_string string, args []interface{}) {
	this.m_lock.UnSafeLock("dbFaceBPayRecordRow.save_data")
	defer this.m_lock.UnSafeUnlock()
	if this.m_new {
		db_args:=new_db_args(5)
		db_args.Push(this.m_KeyId)
		db_args.Push(this.m_Sn)
		db_args.Push(this.m_Bid)
		db_args.Push(this.m_PlayerId)
		db_args.Push(this.m_PayTime)
		args=db_args.GetArgs()
		state = 1
	} else {
		if this.m_Sn_changed||this.m_Bid_changed||this.m_PlayerId_changed||this.m_PayTime_changed{
			update_string = "UPDATE FaceBPayRecords SET "
			db_args:=new_db_args(5)
			if this.m_Sn_changed{
				update_string+="Sn=?,"
				db_args.Push(this.m_Sn)
			}
			if this.m_Bid_changed{
				update_string+="Bid=?,"
				db_args.Push(this.m_Bid)
			}
			if this.m_PlayerId_changed{
				update_string+="PlayerId=?,"
				db_args.Push(this.m_PlayerId)
			}
			if this.m_PayTime_changed{
				update_string+="PayTime=?,"
				db_args.Push(this.m_PayTime)
			}
			update_string = strings.TrimRight(update_string, ", ")
			update_string+=" WHERE KeyId=?"
			db_args.Push(this.m_KeyId)
			args=db_args.GetArgs()
			state = 2
		}
	}
	this.m_new = false
	this.m_Sn_changed = false
	this.m_Bid_changed = false
	this.m_PlayerId_changed = false
	this.m_PayTime_changed = false
	if release && this.m_loaded {
		atomic.AddInt32(&this.m_table.m_gc_n, -1)
		this.m_loaded = false
		released = true
	}
	return nil,released,state,update_string,args
}
func (this *dbFaceBPayRecordRow) Save(release bool) (err error, d bool, released bool) {
	err,released, state, update_string, args := this.save_data(release)
	if err != nil {
		log.Error("save data failed")
		return err, false, false
	}
	if state == 0 {
		d = false
	} else if state == 1 {
		_, err = this.m_table.m_dbc.StmtExec(this.m_table.m_save_insert_stmt, args...)
		if err != nil {
			log.Error("INSERT FaceBPayRecords exec failed %v ", this.m_KeyId)
			return err, false, released
		}
		d = true
	} else if state == 2 {
		_, err = this.m_table.m_dbc.Exec(update_string, args...)
		if err != nil {
			log.Error("UPDATE FaceBPayRecords exec failed %v", this.m_KeyId)
			return err, false, released
		}
		d = true
	}
	return nil, d, released
}
func (this *dbFaceBPayRecordRow) Touch(releasable bool) {
	this.m_touch = int32(time.Now().Unix())
	this.m_releasable = releasable
}
type dbFaceBPayRecordRowSort struct {
	rows []*dbFaceBPayRecordRow
}
func (this *dbFaceBPayRecordRowSort) Len() (length int) {
	return len(this.rows)
}
func (this *dbFaceBPayRecordRowSort) Less(i int, j int) (less bool) {
	return this.rows[i].m_touch < this.rows[j].m_touch
}
func (this *dbFaceBPayRecordRowSort) Swap(i int, j int) {
	temp := this.rows[i]
	this.rows[i] = this.rows[j]
	this.rows[j] = temp
}
type dbFaceBPayRecordTable struct{
	m_dbc *DBC
	m_lock *RWMutex
	m_rows map[int32]*dbFaceBPayRecordRow
	m_new_rows map[int32]*dbFaceBPayRecordRow
	m_removed_rows map[int32]*dbFaceBPayRecordRow
	m_gc_n int32
	m_gcing int32
	m_pool_size int32
	m_preload_select_stmt *sql.Stmt
	m_preload_max_id int32
	m_load_select_stmt *sql.Stmt
	m_save_insert_stmt *sql.Stmt
	m_delete_stmt *sql.Stmt
	m_max_id int32
	m_max_id_changed bool
}
func new_dbFaceBPayRecordTable(dbc *DBC) (this *dbFaceBPayRecordTable) {
	this = &dbFaceBPayRecordTable{}
	this.m_dbc = dbc
	this.m_lock = NewRWMutex()
	this.m_rows = make(map[int32]*dbFaceBPayRecordRow)
	this.m_new_rows = make(map[int32]*dbFaceBPayRecordRow)
	this.m_removed_rows = make(map[int32]*dbFaceBPayRecordRow)
	return this
}
func (this *dbFaceBPayRecordTable) check_create_table() (err error) {
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS FaceBPayRecordsMaxId(PlaceHolder int(11),MaxKeyId int(11),PRIMARY KEY (PlaceHolder))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS FaceBPayRecordsMaxId failed")
		return
	}
	r := this.m_dbc.QueryRow("SELECT Count(*) FROM FaceBPayRecordsMaxId WHERE PlaceHolder=0")
	if r != nil {
		var count int32
		err = r.Scan(&count)
		if err != nil {
			log.Error("scan count failed")
			return
		}
		if count == 0 {
		_, err = this.m_dbc.Exec("INSERT INTO FaceBPayRecordsMaxId (PlaceHolder,MaxKeyId) VALUES (0,0)")
			if err != nil {
				log.Error("INSERTFaceBPayRecordsMaxId failed")
				return
			}
		}
	}
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS FaceBPayRecords(KeyId int(11),PRIMARY KEY (KeyId))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS FaceBPayRecords failed")
		return
	}
	rows, err := this.m_dbc.Query("SELECT COLUMN_NAME,ORDINAL_POSITION FROM information_schema.`COLUMNS` WHERE TABLE_SCHEMA=? AND TABLE_NAME='FaceBPayRecords'", this.m_dbc.m_db_name)
	if err != nil {
		log.Error("SELECT information_schema failed")
		return
	}
	columns := make(map[string]int32)
	for rows.Next() {
		var column_name string
		var ordinal_position int32
		err = rows.Scan(&column_name, &ordinal_position)
		if err != nil {
			log.Error("scan information_schema row failed")
			return
		}
		if ordinal_position < 1 {
			log.Error("col ordinal out of range")
			continue
		}
		columns[column_name] = ordinal_position
	}
	_, hasSn := columns["Sn"]
	if !hasSn {
		_, err = this.m_dbc.Exec("ALTER TABLE FaceBPayRecords ADD COLUMN Sn varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN Sn failed")
			return
		}
	}
	_, hasBid := columns["Bid"]
	if !hasBid {
		_, err = this.m_dbc.Exec("ALTER TABLE FaceBPayRecords ADD COLUMN Bid varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN Bid failed")
			return
		}
	}
	_, hasPlayerId := columns["PlayerId"]
	if !hasPlayerId {
		_, err = this.m_dbc.Exec("ALTER TABLE FaceBPayRecords ADD COLUMN PlayerId int(11)")
		if err != nil {
			log.Error("ADD COLUMN PlayerId failed")
			return
		}
	}
	_, hasPayTime := columns["PayTime"]
	if !hasPayTime {
		_, err = this.m_dbc.Exec("ALTER TABLE FaceBPayRecords ADD COLUMN PayTime int(11)")
		if err != nil {
			log.Error("ADD COLUMN PayTime failed")
			return
		}
	}
	return
}
func (this *dbFaceBPayRecordTable) prepare_preload_select_stmt() (err error) {
	this.m_preload_select_stmt,err=this.m_dbc.StmtPrepare("SELECT KeyId,Sn FROM FaceBPayRecords")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbFaceBPayRecordTable) prepare_load_select_stmt() (err error) {
	this.m_load_select_stmt,err=this.m_dbc.StmtPrepare("SELECT Bid,PlayerId,PayTime FROM FaceBPayRecords WHERE KeyId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbFaceBPayRecordTable) prepare_save_insert_stmt()(err error){
	this.m_save_insert_stmt,err=this.m_dbc.StmtPrepare("INSERT INTO FaceBPayRecords (KeyId,Sn,Bid,PlayerId,PayTime) VALUES (?,?,?,?,?)")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbFaceBPayRecordTable) prepare_delete_stmt() (err error) {
	this.m_delete_stmt,err=this.m_dbc.StmtPrepare("DELETE FROM FaceBPayRecords WHERE KeyId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbFaceBPayRecordTable) Init() (err error) {
	err=this.check_create_table()
	if err!=nil{
		log.Error("check_create_table failed")
		return
	}
	err=this.prepare_preload_select_stmt()
	if err!=nil{
		log.Error("prepare_preload_select_stmt failed")
		return
	}
	err=this.prepare_load_select_stmt()
	if err!=nil{
		log.Error("prepare_load_select_stmt failed")
		return
	}
	err=this.prepare_save_insert_stmt()
	if err!=nil{
		log.Error("prepare_save_insert_stmt failed")
		return
	}
	err=this.prepare_delete_stmt()
	if err!=nil{
		log.Error("prepare_save_insert_stmt failed")
		return
	}
	return
}
func (this *dbFaceBPayRecordTable) Preload() (err error) {
	r_max_id := this.m_dbc.QueryRow("SELECT MaxKeyId FROM FaceBPayRecordsMaxId WHERE PLACEHOLDER=0")
	if r_max_id != nil {
		err = r_max_id.Scan(&this.m_max_id)
		if err != nil {
			log.Error("scan max id failed")
			return
		}
	}
	r, err := this.m_dbc.StmtQuery(this.m_preload_select_stmt)
	if err != nil {
		log.Error("SELECT")
		return
	}
	var KeyId int32
	var dSn string
	for r.Next() {
		err = r.Scan(&KeyId,&dSn)
		if err != nil {
			log.Error("Scan")
			return
		}
		if KeyId>this.m_max_id{
			log.Error("max id ext")
			this.m_max_id = KeyId
			this.m_max_id_changed = true
		}
		row := new_dbFaceBPayRecordRow(this,KeyId)
		row.m_Sn=dSn
		row.m_Sn_changed=false
		row.m_valid = true
		this.m_rows[KeyId]=row
	}
	return
}
func (this *dbFaceBPayRecordTable) GetPreloadedMaxId() (max_id int32) {
	return this.m_preload_max_id
}
func (this *dbFaceBPayRecordTable) fetch_rows(rows map[int32]*dbFaceBPayRecordRow) (r map[int32]*dbFaceBPayRecordRow) {
	this.m_lock.UnSafeLock("dbFaceBPayRecordTable.fetch_rows")
	defer this.m_lock.UnSafeUnlock()
	r = make(map[int32]*dbFaceBPayRecordRow)
	for i, v := range rows {
		r[i] = v
	}
	return r
}
func (this *dbFaceBPayRecordTable) fetch_new_rows() (new_rows map[int32]*dbFaceBPayRecordRow) {
	this.m_lock.UnSafeLock("dbFaceBPayRecordTable.fetch_new_rows")
	defer this.m_lock.UnSafeUnlock()
	new_rows = make(map[int32]*dbFaceBPayRecordRow)
	for i, v := range this.m_new_rows {
		_, has := this.m_rows[i]
		if has {
			log.Error("rows already has new rows %v", i)
			continue
		}
		this.m_rows[i] = v
		new_rows[i] = v
	}
	for i, _ := range new_rows {
		delete(this.m_new_rows, i)
	}
	return
}
func (this *dbFaceBPayRecordTable) save_rows(rows map[int32]*dbFaceBPayRecordRow, quick bool) {
	for _, v := range rows {
		if this.m_dbc.m_quit && !quick {
			return
		}
		err, delay, _ := v.Save(false)
		if err != nil {
			log.Error("save failed %v", err)
		}
		if this.m_dbc.m_quit && !quick {
			return
		}
		if delay&&!quick {
			time.Sleep(time.Millisecond * 5)
		}
	}
}
func (this *dbFaceBPayRecordTable) Save(quick bool) (err error){
	if this.m_max_id_changed {
		max_id := atomic.LoadInt32(&this.m_max_id)
		_, err := this.m_dbc.Exec("UPDATE FaceBPayRecordsMaxId SET MaxKeyId=?", max_id)
		if err != nil {
			log.Error("save max id failed %v", err)
		}
	}
	removed_rows := this.fetch_rows(this.m_removed_rows)
	for _, v := range removed_rows {
		_, err := this.m_dbc.StmtExec(this.m_delete_stmt, v.GetKeyId())
		if err != nil {
			log.Error("exec delete stmt failed %v", err)
		}
		v.m_valid = false
		if !quick {
			time.Sleep(time.Millisecond * 5)
		}
	}
	this.m_removed_rows = make(map[int32]*dbFaceBPayRecordRow)
	rows := this.fetch_rows(this.m_rows)
	this.save_rows(rows, quick)
	new_rows := this.fetch_new_rows()
	this.save_rows(new_rows, quick)
	return
}
func (this *dbFaceBPayRecordTable) AddRow() (row *dbFaceBPayRecordRow) {
	this.GC()
	this.m_lock.UnSafeLock("dbFaceBPayRecordTable.AddRow")
	defer this.m_lock.UnSafeUnlock()
	KeyId := atomic.AddInt32(&this.m_max_id, 1)
	this.m_max_id_changed = true
	row = new_dbFaceBPayRecordRow(this,KeyId)
	row.m_new = true
	row.m_loaded = true
	row.m_valid = true
	this.m_new_rows[KeyId] = row
	atomic.AddInt32(&this.m_gc_n,1)
	return row
}
func (this *dbFaceBPayRecordTable) RemoveRow(KeyId int32) {
	this.m_lock.UnSafeLock("dbFaceBPayRecordTable.RemoveRow")
	defer this.m_lock.UnSafeUnlock()
	row := this.m_rows[KeyId]
	if row != nil {
		row.m_remove = true
		delete(this.m_rows, KeyId)
		rm_row := this.m_removed_rows[KeyId]
		if rm_row != nil {
			log.Error("rows and removed rows both has %v", KeyId)
		}
		this.m_removed_rows[KeyId] = row
		_, has_new := this.m_new_rows[KeyId]
		if has_new {
			delete(this.m_new_rows, KeyId)
			log.Error("rows and new_rows both has %v", KeyId)
		}
	} else {
		row = this.m_removed_rows[KeyId]
		if row == nil {
			_, has_new := this.m_new_rows[KeyId]
			if has_new {
				delete(this.m_new_rows, KeyId)
			} else {
				log.Error("row not exist %v", KeyId)
			}
		} else {
			log.Error("already removed %v", KeyId)
			_, has_new := this.m_new_rows[KeyId]
			if has_new {
				delete(this.m_new_rows, KeyId)
				log.Error("removed rows and new_rows both has %v", KeyId)
			}
		}
	}
}
func (this *dbFaceBPayRecordTable) GetRow(KeyId int32) (row *dbFaceBPayRecordRow) {
	this.m_lock.UnSafeRLock("dbFaceBPayRecordTable.GetRow")
	defer this.m_lock.UnSafeRUnlock()
	row = this.m_rows[KeyId]
	if row == nil {
		row = this.m_new_rows[KeyId]
	}
	return row
}
func (this *dbFaceBPayRecordTable) SetPoolSize(n int32) {
	this.m_pool_size = n
}
func (this *dbFaceBPayRecordTable) GC() {
	if this.m_pool_size<=0{
		return
	}
	if !atomic.CompareAndSwapInt32(&this.m_gcing, 0, 1) {
		return
	}
	defer atomic.StoreInt32(&this.m_gcing, 0)
	n := atomic.LoadInt32(&this.m_gc_n)
	if float32(n) < float32(this.m_pool_size)*1.2 {
		return
	}
	max := (n - this.m_pool_size) / 2
	arr := dbFaceBPayRecordRowSort{}
	rows := this.fetch_rows(this.m_rows)
	arr.rows = make([]*dbFaceBPayRecordRow, len(rows))
	index := 0
	for _, v := range rows {
		arr.rows[index] = v
		index++
	}
	sort.Sort(&arr)
	count := int32(0)
	for _, v := range arr.rows {
		err, _, released := v.Save(true)
		if err != nil {
			log.Error("release failed %v", err)
			continue
		}
		if released {
			count++
			if count > max {
				return
			}
		}
	}
	return
}
func (this *dbServerInfoRow)GetCreateUnix( )(r int32 ){
	this.m_lock.UnSafeRLock("dbServerInfoRow.GetdbServerInfoCreateUnixColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_CreateUnix)
}
func (this *dbServerInfoRow)SetCreateUnix(v int32){
	this.m_lock.UnSafeLock("dbServerInfoRow.SetdbServerInfoCreateUnixColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_CreateUnix=int32(v)
	this.m_CreateUnix_changed=true
	return
}
func (this *dbServerInfoRow)GetCurStartUnix( )(r int32 ){
	this.m_lock.UnSafeRLock("dbServerInfoRow.GetdbServerInfoCurStartUnixColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_CurStartUnix)
}
func (this *dbServerInfoRow)SetCurStartUnix(v int32){
	this.m_lock.UnSafeLock("dbServerInfoRow.SetdbServerInfoCurStartUnixColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_CurStartUnix=int32(v)
	this.m_CurStartUnix_changed=true
	return
}
type dbServerInfoRow struct {
	m_table *dbServerInfoTable
	m_lock       *RWMutex
	m_loaded  bool
	m_new     bool
	m_remove  bool
	m_touch      int32
	m_releasable bool
	m_valid   bool
	m_KeyId        int32
	m_CreateUnix_changed bool
	m_CreateUnix int32
	m_CurStartUnix_changed bool
	m_CurStartUnix int32
}
func new_dbServerInfoRow(table *dbServerInfoTable, KeyId int32) (r *dbServerInfoRow) {
	this := &dbServerInfoRow{}
	this.m_table = table
	this.m_KeyId = KeyId
	this.m_lock = NewRWMutex()
	this.m_CreateUnix_changed=true
	this.m_CurStartUnix_changed=true
	return this
}
func (this *dbServerInfoRow) save_data(release bool) (err error, released bool, state int32, update_string string, args []interface{}) {
	this.m_lock.UnSafeLock("dbServerInfoRow.save_data")
	defer this.m_lock.UnSafeUnlock()
	if this.m_new {
		db_args:=new_db_args(3)
		db_args.Push(this.m_KeyId)
		db_args.Push(this.m_CreateUnix)
		db_args.Push(this.m_CurStartUnix)
		args=db_args.GetArgs()
		state = 1
	} else {
		if this.m_CreateUnix_changed||this.m_CurStartUnix_changed{
			update_string = "UPDATE ServerInfo SET "
			db_args:=new_db_args(3)
			if this.m_CreateUnix_changed{
				update_string+="CreateUnix=?,"
				db_args.Push(this.m_CreateUnix)
			}
			if this.m_CurStartUnix_changed{
				update_string+="CurStartUnix=?,"
				db_args.Push(this.m_CurStartUnix)
			}
			update_string = strings.TrimRight(update_string, ", ")
			update_string+=" WHERE KeyId=?"
			db_args.Push(this.m_KeyId)
			args=db_args.GetArgs()
			state = 2
		}
	}
	this.m_new = false
	this.m_CreateUnix_changed = false
	this.m_CurStartUnix_changed = false
	if release && this.m_loaded {
		this.m_loaded = false
		released = true
	}
	return nil,released,state,update_string,args
}
func (this *dbServerInfoRow) Save(release bool) (err error, d bool, released bool) {
	err,released, state, update_string, args := this.save_data(release)
	if err != nil {
		log.Error("save data failed")
		return err, false, false
	}
	if state == 0 {
		d = false
	} else if state == 1 {
		_, err = this.m_table.m_dbc.StmtExec(this.m_table.m_save_insert_stmt, args...)
		if err != nil {
			log.Error("INSERT ServerInfo exec failed %v ", this.m_KeyId)
			return err, false, released
		}
		d = true
	} else if state == 2 {
		_, err = this.m_table.m_dbc.Exec(update_string, args...)
		if err != nil {
			log.Error("UPDATE ServerInfo exec failed %v", this.m_KeyId)
			return err, false, released
		}
		d = true
	}
	return nil, d, released
}
type dbServerInfoTable struct{
	m_dbc *DBC
	m_lock *RWMutex
	m_row *dbServerInfoRow
	m_preload_select_stmt *sql.Stmt
	m_save_insert_stmt *sql.Stmt
}
func new_dbServerInfoTable(dbc *DBC) (this *dbServerInfoTable) {
	this = &dbServerInfoTable{}
	this.m_dbc = dbc
	this.m_lock = NewRWMutex()
	return this
}
func (this *dbServerInfoTable) check_create_table() (err error) {
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS ServerInfo(KeyId int(11),PRIMARY KEY (KeyId))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS ServerInfo failed")
		return
	}
	rows, err := this.m_dbc.Query("SELECT COLUMN_NAME,ORDINAL_POSITION FROM information_schema.`COLUMNS` WHERE TABLE_SCHEMA=? AND TABLE_NAME='ServerInfo'", this.m_dbc.m_db_name)
	if err != nil {
		log.Error("SELECT information_schema failed")
		return
	}
	columns := make(map[string]int32)
	for rows.Next() {
		var column_name string
		var ordinal_position int32
		err = rows.Scan(&column_name, &ordinal_position)
		if err != nil {
			log.Error("scan information_schema row failed")
			return
		}
		if ordinal_position < 1 {
			log.Error("col ordinal out of range")
			continue
		}
		columns[column_name] = ordinal_position
	}
	_, hasCreateUnix := columns["CreateUnix"]
	if !hasCreateUnix {
		_, err = this.m_dbc.Exec("ALTER TABLE ServerInfo ADD COLUMN CreateUnix int(11)")
		if err != nil {
			log.Error("ADD COLUMN CreateUnix failed")
			return
		}
	}
	_, hasCurStartUnix := columns["CurStartUnix"]
	if !hasCurStartUnix {
		_, err = this.m_dbc.Exec("ALTER TABLE ServerInfo ADD COLUMN CurStartUnix int(11)")
		if err != nil {
			log.Error("ADD COLUMN CurStartUnix failed")
			return
		}
	}
	return
}
func (this *dbServerInfoTable) prepare_preload_select_stmt() (err error) {
	this.m_preload_select_stmt,err=this.m_dbc.StmtPrepare("SELECT CreateUnix,CurStartUnix FROM ServerInfo WHERE KeyId=0")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbServerInfoTable) prepare_save_insert_stmt()(err error){
	this.m_save_insert_stmt,err=this.m_dbc.StmtPrepare("INSERT INTO ServerInfo (KeyId,CreateUnix,CurStartUnix) VALUES (?,?,?)")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbServerInfoTable) Init() (err error) {
	err=this.check_create_table()
	if err!=nil{
		log.Error("check_create_table failed")
		return
	}
	err=this.prepare_preload_select_stmt()
	if err!=nil{
		log.Error("prepare_preload_select_stmt failed")
		return
	}
	err=this.prepare_save_insert_stmt()
	if err!=nil{
		log.Error("prepare_save_insert_stmt failed")
		return
	}
	return
}
func (this *dbServerInfoTable) Preload() (err error) {
	r := this.m_dbc.StmtQueryRow(this.m_preload_select_stmt)
	var dCreateUnix int32
	var dCurStartUnix int32
	err = r.Scan(&dCreateUnix,&dCurStartUnix)
	if err!=nil{
		if err!=sql.ErrNoRows{
			log.Error("Scan failed")
			return
		}
	}else{
		row := new_dbServerInfoRow(this,0)
		row.m_CreateUnix=dCreateUnix
		row.m_CurStartUnix=dCurStartUnix
		row.m_CreateUnix_changed=false
		row.m_CurStartUnix_changed=false
		row.m_valid = true
		row.m_loaded=true
		this.m_row=row
	}
	if this.m_row == nil {
		this.m_row = new_dbServerInfoRow(this, 0)
		this.m_row.m_new = true
		this.m_row.m_valid = true
		err = this.Save(false)
		if err != nil {
			log.Error("save failed")
			return
		}
		this.m_row.m_loaded = true
	}
	return
}
func (this *dbServerInfoTable) Save(quick bool) (err error) {
	if this.m_row==nil{
		return errors.New("row nil")
	}
	err, _, _ = this.m_row.Save(false)
	return err
}
func (this *dbServerInfoTable) GetRow( ) (row *dbServerInfoRow) {
	return this.m_row
}
func (this *dbPlayerLoginRow)GetPlayerAccount( )(r string ){
	this.m_lock.UnSafeRLock("dbPlayerLoginRow.GetdbPlayerLoginPlayerAccountColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_PlayerAccount)
}
func (this *dbPlayerLoginRow)SetPlayerAccount(v string){
	this.m_lock.UnSafeLock("dbPlayerLoginRow.SetdbPlayerLoginPlayerAccountColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_PlayerAccount=string(v)
	this.m_PlayerAccount_changed=true
	return
}
func (this *dbPlayerLoginRow)GetPlayerId( )(r int32 ){
	this.m_lock.UnSafeRLock("dbPlayerLoginRow.GetdbPlayerLoginPlayerIdColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_PlayerId)
}
func (this *dbPlayerLoginRow)SetPlayerId(v int32){
	this.m_lock.UnSafeLock("dbPlayerLoginRow.SetdbPlayerLoginPlayerIdColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_PlayerId=int32(v)
	this.m_PlayerId_changed=true
	return
}
func (this *dbPlayerLoginRow)GetPlayerName( )(r string ){
	this.m_lock.UnSafeRLock("dbPlayerLoginRow.GetdbPlayerLoginPlayerNameColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_PlayerName)
}
func (this *dbPlayerLoginRow)SetPlayerName(v string){
	this.m_lock.UnSafeLock("dbPlayerLoginRow.SetdbPlayerLoginPlayerNameColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_PlayerName=string(v)
	this.m_PlayerName_changed=true
	return
}
func (this *dbPlayerLoginRow)GetLoginTime( )(r int32 ){
	this.m_lock.UnSafeRLock("dbPlayerLoginRow.GetdbPlayerLoginLoginTimeColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_LoginTime)
}
func (this *dbPlayerLoginRow)SetLoginTime(v int32){
	this.m_lock.UnSafeLock("dbPlayerLoginRow.SetdbPlayerLoginLoginTimeColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_LoginTime=int32(v)
	this.m_LoginTime_changed=true
	return
}
type dbPlayerLoginRow struct {
	m_table *dbPlayerLoginTable
	m_lock       *RWMutex
	m_loaded  bool
	m_new     bool
	m_remove  bool
	m_touch      int32
	m_releasable bool
	m_valid   bool
	m_KeyId        int32
	m_PlayerAccount_changed bool
	m_PlayerAccount string
	m_PlayerId_changed bool
	m_PlayerId int32
	m_PlayerName_changed bool
	m_PlayerName string
	m_LoginTime_changed bool
	m_LoginTime int32
}
func new_dbPlayerLoginRow(table *dbPlayerLoginTable, KeyId int32) (r *dbPlayerLoginRow) {
	this := &dbPlayerLoginRow{}
	this.m_table = table
	this.m_KeyId = KeyId
	this.m_lock = NewRWMutex()
	this.m_PlayerAccount_changed=true
	this.m_PlayerId_changed=true
	this.m_PlayerName_changed=true
	this.m_LoginTime_changed=true
	return this
}
func (this *dbPlayerLoginRow) GetKeyId() (r int32) {
	return this.m_KeyId
}
func (this *dbPlayerLoginRow) save_data(release bool) (err error, released bool, state int32, update_string string, args []interface{}) {
	this.m_lock.UnSafeLock("dbPlayerLoginRow.save_data")
	defer this.m_lock.UnSafeUnlock()
	if this.m_new {
		db_args:=new_db_args(5)
		db_args.Push(this.m_KeyId)
		db_args.Push(this.m_PlayerAccount)
		db_args.Push(this.m_PlayerId)
		db_args.Push(this.m_PlayerName)
		db_args.Push(this.m_LoginTime)
		args=db_args.GetArgs()
		state = 1
	} else {
		if this.m_PlayerAccount_changed||this.m_PlayerId_changed||this.m_PlayerName_changed||this.m_LoginTime_changed{
			update_string = "UPDATE PlayerLogins SET "
			db_args:=new_db_args(5)
			if this.m_PlayerAccount_changed{
				update_string+="PlayerAccount=?,"
				db_args.Push(this.m_PlayerAccount)
			}
			if this.m_PlayerId_changed{
				update_string+="PlayerId=?,"
				db_args.Push(this.m_PlayerId)
			}
			if this.m_PlayerName_changed{
				update_string+="PlayerName=?,"
				db_args.Push(this.m_PlayerName)
			}
			if this.m_LoginTime_changed{
				update_string+="LoginTime=?,"
				db_args.Push(this.m_LoginTime)
			}
			update_string = strings.TrimRight(update_string, ", ")
			update_string+=" WHERE KeyId=?"
			db_args.Push(this.m_KeyId)
			args=db_args.GetArgs()
			state = 2
		}
	}
	this.m_new = false
	this.m_PlayerAccount_changed = false
	this.m_PlayerId_changed = false
	this.m_PlayerName_changed = false
	this.m_LoginTime_changed = false
	if release && this.m_loaded {
		atomic.AddInt32(&this.m_table.m_gc_n, -1)
		this.m_loaded = false
		released = true
	}
	return nil,released,state,update_string,args
}
func (this *dbPlayerLoginRow) Save(release bool) (err error, d bool, released bool) {
	err,released, state, update_string, args := this.save_data(release)
	if err != nil {
		log.Error("save data failed")
		return err, false, false
	}
	if state == 0 {
		d = false
	} else if state == 1 {
		_, err = this.m_table.m_dbc.StmtExec(this.m_table.m_save_insert_stmt, args...)
		if err != nil {
			log.Error("INSERT PlayerLogins exec failed %v ", this.m_KeyId)
			return err, false, released
		}
		d = true
	} else if state == 2 {
		_, err = this.m_table.m_dbc.Exec(update_string, args...)
		if err != nil {
			log.Error("UPDATE PlayerLogins exec failed %v", this.m_KeyId)
			return err, false, released
		}
		d = true
	}
	return nil, d, released
}
func (this *dbPlayerLoginRow) Touch(releasable bool) {
	this.m_touch = int32(time.Now().Unix())
	this.m_releasable = releasable
}
type dbPlayerLoginRowSort struct {
	rows []*dbPlayerLoginRow
}
func (this *dbPlayerLoginRowSort) Len() (length int) {
	return len(this.rows)
}
func (this *dbPlayerLoginRowSort) Less(i int, j int) (less bool) {
	return this.rows[i].m_touch < this.rows[j].m_touch
}
func (this *dbPlayerLoginRowSort) Swap(i int, j int) {
	temp := this.rows[i]
	this.rows[i] = this.rows[j]
	this.rows[j] = temp
}
type dbPlayerLoginTable struct{
	m_dbc *DBC
	m_lock *RWMutex
	m_rows map[int32]*dbPlayerLoginRow
	m_new_rows map[int32]*dbPlayerLoginRow
	m_removed_rows map[int32]*dbPlayerLoginRow
	m_gc_n int32
	m_gcing int32
	m_pool_size int32
	m_preload_select_stmt *sql.Stmt
	m_preload_max_id int32
	m_save_insert_stmt *sql.Stmt
	m_delete_stmt *sql.Stmt
	m_max_id int32
	m_max_id_changed bool
}
func new_dbPlayerLoginTable(dbc *DBC) (this *dbPlayerLoginTable) {
	this = &dbPlayerLoginTable{}
	this.m_dbc = dbc
	this.m_lock = NewRWMutex()
	this.m_rows = make(map[int32]*dbPlayerLoginRow)
	this.m_new_rows = make(map[int32]*dbPlayerLoginRow)
	this.m_removed_rows = make(map[int32]*dbPlayerLoginRow)
	return this
}
func (this *dbPlayerLoginTable) check_create_table() (err error) {
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS PlayerLoginsMaxId(PlaceHolder int(11),MaxKeyId int(11),PRIMARY KEY (PlaceHolder))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS PlayerLoginsMaxId failed")
		return
	}
	r := this.m_dbc.QueryRow("SELECT Count(*) FROM PlayerLoginsMaxId WHERE PlaceHolder=0")
	if r != nil {
		var count int32
		err = r.Scan(&count)
		if err != nil {
			log.Error("scan count failed")
			return
		}
		if count == 0 {
		_, err = this.m_dbc.Exec("INSERT INTO PlayerLoginsMaxId (PlaceHolder,MaxKeyId) VALUES (0,0)")
			if err != nil {
				log.Error("INSERTPlayerLoginsMaxId failed")
				return
			}
		}
	}
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS PlayerLogins(KeyId int(11),PRIMARY KEY (KeyId))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS PlayerLogins failed")
		return
	}
	rows, err := this.m_dbc.Query("SELECT COLUMN_NAME,ORDINAL_POSITION FROM information_schema.`COLUMNS` WHERE TABLE_SCHEMA=? AND TABLE_NAME='PlayerLogins'", this.m_dbc.m_db_name)
	if err != nil {
		log.Error("SELECT information_schema failed")
		return
	}
	columns := make(map[string]int32)
	for rows.Next() {
		var column_name string
		var ordinal_position int32
		err = rows.Scan(&column_name, &ordinal_position)
		if err != nil {
			log.Error("scan information_schema row failed")
			return
		}
		if ordinal_position < 1 {
			log.Error("col ordinal out of range")
			continue
		}
		columns[column_name] = ordinal_position
	}
	_, hasPlayerAccount := columns["PlayerAccount"]
	if !hasPlayerAccount {
		_, err = this.m_dbc.Exec("ALTER TABLE PlayerLogins ADD COLUMN PlayerAccount varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN PlayerAccount failed")
			return
		}
	}
	_, hasPlayerId := columns["PlayerId"]
	if !hasPlayerId {
		_, err = this.m_dbc.Exec("ALTER TABLE PlayerLogins ADD COLUMN PlayerId int(11)")
		if err != nil {
			log.Error("ADD COLUMN PlayerId failed")
			return
		}
	}
	_, hasPlayerName := columns["PlayerName"]
	if !hasPlayerName {
		_, err = this.m_dbc.Exec("ALTER TABLE PlayerLogins ADD COLUMN PlayerName varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN PlayerName failed")
			return
		}
	}
	_, hasLoginTime := columns["LoginTime"]
	if !hasLoginTime {
		_, err = this.m_dbc.Exec("ALTER TABLE PlayerLogins ADD COLUMN LoginTime int(11)")
		if err != nil {
			log.Error("ADD COLUMN LoginTime failed")
			return
		}
	}
	return
}
func (this *dbPlayerLoginTable) prepare_preload_select_stmt() (err error) {
	this.m_preload_select_stmt,err=this.m_dbc.StmtPrepare("SELECT KeyId,PlayerAccount,PlayerId,PlayerName,LoginTime FROM PlayerLogins")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbPlayerLoginTable) prepare_save_insert_stmt()(err error){
	this.m_save_insert_stmt,err=this.m_dbc.StmtPrepare("INSERT INTO PlayerLogins (KeyId,PlayerAccount,PlayerId,PlayerName,LoginTime) VALUES (?,?,?,?,?)")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbPlayerLoginTable) prepare_delete_stmt() (err error) {
	this.m_delete_stmt,err=this.m_dbc.StmtPrepare("DELETE FROM PlayerLogins WHERE KeyId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbPlayerLoginTable) Init() (err error) {
	err=this.check_create_table()
	if err!=nil{
		log.Error("check_create_table failed")
		return
	}
	err=this.prepare_preload_select_stmt()
	if err!=nil{
		log.Error("prepare_preload_select_stmt failed")
		return
	}
	err=this.prepare_save_insert_stmt()
	if err!=nil{
		log.Error("prepare_save_insert_stmt failed")
		return
	}
	err=this.prepare_delete_stmt()
	if err!=nil{
		log.Error("prepare_save_insert_stmt failed")
		return
	}
	return
}
func (this *dbPlayerLoginTable) Preload() (err error) {
	r_max_id := this.m_dbc.QueryRow("SELECT MaxKeyId FROM PlayerLoginsMaxId WHERE PLACEHOLDER=0")
	if r_max_id != nil {
		err = r_max_id.Scan(&this.m_max_id)
		if err != nil {
			log.Error("scan max id failed")
			return
		}
	}
	r, err := this.m_dbc.StmtQuery(this.m_preload_select_stmt)
	if err != nil {
		log.Error("SELECT")
		return
	}
	var KeyId int32
	var dPlayerAccount string
	var dPlayerId int32
	var dPlayerName string
	var dLoginTime int32
	for r.Next() {
		err = r.Scan(&KeyId,&dPlayerAccount,&dPlayerId,&dPlayerName,&dLoginTime)
		if err != nil {
			log.Error("Scan")
			return
		}
		if KeyId>this.m_max_id{
			log.Error("max id ext")
			this.m_max_id = KeyId
			this.m_max_id_changed = true
		}
		row := new_dbPlayerLoginRow(this,KeyId)
		row.m_PlayerAccount=dPlayerAccount
		row.m_PlayerId=dPlayerId
		row.m_PlayerName=dPlayerName
		row.m_LoginTime=dLoginTime
		row.m_PlayerAccount_changed=false
		row.m_PlayerId_changed=false
		row.m_PlayerName_changed=false
		row.m_LoginTime_changed=false
		row.m_valid = true
		this.m_rows[KeyId]=row
	}
	return
}
func (this *dbPlayerLoginTable) GetPreloadedMaxId() (max_id int32) {
	return this.m_preload_max_id
}
func (this *dbPlayerLoginTable) fetch_rows(rows map[int32]*dbPlayerLoginRow) (r map[int32]*dbPlayerLoginRow) {
	this.m_lock.UnSafeLock("dbPlayerLoginTable.fetch_rows")
	defer this.m_lock.UnSafeUnlock()
	r = make(map[int32]*dbPlayerLoginRow)
	for i, v := range rows {
		r[i] = v
	}
	return r
}
func (this *dbPlayerLoginTable) fetch_new_rows() (new_rows map[int32]*dbPlayerLoginRow) {
	this.m_lock.UnSafeLock("dbPlayerLoginTable.fetch_new_rows")
	defer this.m_lock.UnSafeUnlock()
	new_rows = make(map[int32]*dbPlayerLoginRow)
	for i, v := range this.m_new_rows {
		_, has := this.m_rows[i]
		if has {
			log.Error("rows already has new rows %v", i)
			continue
		}
		this.m_rows[i] = v
		new_rows[i] = v
	}
	for i, _ := range new_rows {
		delete(this.m_new_rows, i)
	}
	return
}
func (this *dbPlayerLoginTable) save_rows(rows map[int32]*dbPlayerLoginRow, quick bool) {
	for _, v := range rows {
		if this.m_dbc.m_quit && !quick {
			return
		}
		err, delay, _ := v.Save(false)
		if err != nil {
			log.Error("save failed %v", err)
		}
		if this.m_dbc.m_quit && !quick {
			return
		}
		if delay&&!quick {
			time.Sleep(time.Millisecond * 5)
		}
	}
}
func (this *dbPlayerLoginTable) Save(quick bool) (err error){
	if this.m_max_id_changed {
		max_id := atomic.LoadInt32(&this.m_max_id)
		_, err := this.m_dbc.Exec("UPDATE PlayerLoginsMaxId SET MaxKeyId=?", max_id)
		if err != nil {
			log.Error("save max id failed %v", err)
		}
	}
	removed_rows := this.fetch_rows(this.m_removed_rows)
	for _, v := range removed_rows {
		_, err := this.m_dbc.StmtExec(this.m_delete_stmt, v.GetKeyId())
		if err != nil {
			log.Error("exec delete stmt failed %v", err)
		}
		v.m_valid = false
		if !quick {
			time.Sleep(time.Millisecond * 5)
		}
	}
	this.m_removed_rows = make(map[int32]*dbPlayerLoginRow)
	rows := this.fetch_rows(this.m_rows)
	this.save_rows(rows, quick)
	new_rows := this.fetch_new_rows()
	this.save_rows(new_rows, quick)
	return
}
func (this *dbPlayerLoginTable) AddRow() (row *dbPlayerLoginRow) {
	this.m_lock.UnSafeLock("dbPlayerLoginTable.AddRow")
	defer this.m_lock.UnSafeUnlock()
	KeyId := atomic.AddInt32(&this.m_max_id, 1)
	this.m_max_id_changed = true
	row = new_dbPlayerLoginRow(this,KeyId)
	row.m_new = true
	row.m_loaded = true
	row.m_valid = true
	this.m_new_rows[KeyId] = row
	atomic.AddInt32(&this.m_gc_n,1)
	return row
}
func (this *dbPlayerLoginTable) RemoveRow(KeyId int32) {
	this.m_lock.UnSafeLock("dbPlayerLoginTable.RemoveRow")
	defer this.m_lock.UnSafeUnlock()
	row := this.m_rows[KeyId]
	if row != nil {
		row.m_remove = true
		delete(this.m_rows, KeyId)
		rm_row := this.m_removed_rows[KeyId]
		if rm_row != nil {
			log.Error("rows and removed rows both has %v", KeyId)
		}
		this.m_removed_rows[KeyId] = row
		_, has_new := this.m_new_rows[KeyId]
		if has_new {
			delete(this.m_new_rows, KeyId)
			log.Error("rows and new_rows both has %v", KeyId)
		}
	} else {
		row = this.m_removed_rows[KeyId]
		if row == nil {
			_, has_new := this.m_new_rows[KeyId]
			if has_new {
				delete(this.m_new_rows, KeyId)
			} else {
				log.Error("row not exist %v", KeyId)
			}
		} else {
			log.Error("already removed %v", KeyId)
			_, has_new := this.m_new_rows[KeyId]
			if has_new {
				delete(this.m_new_rows, KeyId)
				log.Error("removed rows and new_rows both has %v", KeyId)
			}
		}
	}
}
func (this *dbPlayerLoginTable) GetRow(KeyId int32) (row *dbPlayerLoginRow) {
	this.m_lock.UnSafeRLock("dbPlayerLoginTable.GetRow")
	defer this.m_lock.UnSafeRUnlock()
	row = this.m_rows[KeyId]
	if row == nil {
		row = this.m_new_rows[KeyId]
	}
	return row
}
func (this *dbOtherServerPlayerRow)GetAccount( )(r string ){
	this.m_lock.UnSafeRLock("dbOtherServerPlayerRow.GetdbOtherServerPlayerAccountColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_Account)
}
func (this *dbOtherServerPlayerRow)SetAccount(v string){
	this.m_lock.UnSafeLock("dbOtherServerPlayerRow.SetdbOtherServerPlayerAccountColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_Account=string(v)
	this.m_Account_changed=true
	return
}
func (this *dbOtherServerPlayerRow)GetName( )(r string ){
	this.m_lock.UnSafeRLock("dbOtherServerPlayerRow.GetdbOtherServerPlayerNameColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_Name)
}
func (this *dbOtherServerPlayerRow)SetName(v string){
	this.m_lock.UnSafeLock("dbOtherServerPlayerRow.SetdbOtherServerPlayerNameColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_Name=string(v)
	this.m_Name_changed=true
	return
}
func (this *dbOtherServerPlayerRow)GetLevel( )(r int32 ){
	this.m_lock.UnSafeRLock("dbOtherServerPlayerRow.GetdbOtherServerPlayerLevelColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_Level)
}
func (this *dbOtherServerPlayerRow)SetLevel(v int32){
	this.m_lock.UnSafeLock("dbOtherServerPlayerRow.SetdbOtherServerPlayerLevelColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_Level=int32(v)
	this.m_Level_changed=true
	return
}
func (this *dbOtherServerPlayerRow)GetHead( )(r string ){
	this.m_lock.UnSafeRLock("dbOtherServerPlayerRow.GetdbOtherServerPlayerHeadColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_Head)
}
func (this *dbOtherServerPlayerRow)SetHead(v string){
	this.m_lock.UnSafeLock("dbOtherServerPlayerRow.SetdbOtherServerPlayerHeadColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_Head=string(v)
	this.m_Head_changed=true
	return
}
type dbOtherServerPlayerRow struct {
	m_table *dbOtherServerPlayerTable
	m_lock       *RWMutex
	m_loaded  bool
	m_new     bool
	m_remove  bool
	m_touch      int32
	m_releasable bool
	m_valid   bool
	m_PlayerId        int32
	m_Account_changed bool
	m_Account string
	m_Name_changed bool
	m_Name string
	m_Level_changed bool
	m_Level int32
	m_Head_changed bool
	m_Head string
}
func new_dbOtherServerPlayerRow(table *dbOtherServerPlayerTable, PlayerId int32) (r *dbOtherServerPlayerRow) {
	this := &dbOtherServerPlayerRow{}
	this.m_table = table
	this.m_PlayerId = PlayerId
	this.m_lock = NewRWMutex()
	this.m_Account_changed=true
	this.m_Name_changed=true
	this.m_Level_changed=true
	this.m_Head_changed=true
	return this
}
func (this *dbOtherServerPlayerRow) GetPlayerId() (r int32) {
	return this.m_PlayerId
}
func (this *dbOtherServerPlayerRow) save_data(release bool) (err error, released bool, state int32, update_string string, args []interface{}) {
	this.m_lock.UnSafeLock("dbOtherServerPlayerRow.save_data")
	defer this.m_lock.UnSafeUnlock()
	if this.m_new {
		db_args:=new_db_args(5)
		db_args.Push(this.m_PlayerId)
		db_args.Push(this.m_Account)
		db_args.Push(this.m_Name)
		db_args.Push(this.m_Level)
		db_args.Push(this.m_Head)
		args=db_args.GetArgs()
		state = 1
	} else {
		if this.m_Account_changed||this.m_Name_changed||this.m_Level_changed||this.m_Head_changed{
			update_string = "UPDATE OtherServerPlayers SET "
			db_args:=new_db_args(5)
			if this.m_Account_changed{
				update_string+="Account=?,"
				db_args.Push(this.m_Account)
			}
			if this.m_Name_changed{
				update_string+="Name=?,"
				db_args.Push(this.m_Name)
			}
			if this.m_Level_changed{
				update_string+="Level=?,"
				db_args.Push(this.m_Level)
			}
			if this.m_Head_changed{
				update_string+="Head=?,"
				db_args.Push(this.m_Head)
			}
			update_string = strings.TrimRight(update_string, ", ")
			update_string+=" WHERE PlayerId=?"
			db_args.Push(this.m_PlayerId)
			args=db_args.GetArgs()
			state = 2
		}
	}
	this.m_new = false
	this.m_Account_changed = false
	this.m_Name_changed = false
	this.m_Level_changed = false
	this.m_Head_changed = false
	if release && this.m_loaded {
		atomic.AddInt32(&this.m_table.m_gc_n, -1)
		this.m_loaded = false
		released = true
	}
	return nil,released,state,update_string,args
}
func (this *dbOtherServerPlayerRow) Save(release bool) (err error, d bool, released bool) {
	err,released, state, update_string, args := this.save_data(release)
	if err != nil {
		log.Error("save data failed")
		return err, false, false
	}
	if state == 0 {
		d = false
	} else if state == 1 {
		_, err = this.m_table.m_dbc.StmtExec(this.m_table.m_save_insert_stmt, args...)
		if err != nil {
			log.Error("INSERT OtherServerPlayers exec failed %v ", this.m_PlayerId)
			return err, false, released
		}
		d = true
	} else if state == 2 {
		_, err = this.m_table.m_dbc.Exec(update_string, args...)
		if err != nil {
			log.Error("UPDATE OtherServerPlayers exec failed %v", this.m_PlayerId)
			return err, false, released
		}
		d = true
	}
	return nil, d, released
}
func (this *dbOtherServerPlayerRow) Touch(releasable bool) {
	this.m_touch = int32(time.Now().Unix())
	this.m_releasable = releasable
}
type dbOtherServerPlayerRowSort struct {
	rows []*dbOtherServerPlayerRow
}
func (this *dbOtherServerPlayerRowSort) Len() (length int) {
	return len(this.rows)
}
func (this *dbOtherServerPlayerRowSort) Less(i int, j int) (less bool) {
	return this.rows[i].m_touch < this.rows[j].m_touch
}
func (this *dbOtherServerPlayerRowSort) Swap(i int, j int) {
	temp := this.rows[i]
	this.rows[i] = this.rows[j]
	this.rows[j] = temp
}
type dbOtherServerPlayerTable struct{
	m_dbc *DBC
	m_lock *RWMutex
	m_rows map[int32]*dbOtherServerPlayerRow
	m_new_rows map[int32]*dbOtherServerPlayerRow
	m_removed_rows map[int32]*dbOtherServerPlayerRow
	m_gc_n int32
	m_gcing int32
	m_pool_size int32
	m_preload_select_stmt *sql.Stmt
	m_preload_max_id int32
	m_save_insert_stmt *sql.Stmt
	m_delete_stmt *sql.Stmt
}
func new_dbOtherServerPlayerTable(dbc *DBC) (this *dbOtherServerPlayerTable) {
	this = &dbOtherServerPlayerTable{}
	this.m_dbc = dbc
	this.m_lock = NewRWMutex()
	this.m_rows = make(map[int32]*dbOtherServerPlayerRow)
	this.m_new_rows = make(map[int32]*dbOtherServerPlayerRow)
	this.m_removed_rows = make(map[int32]*dbOtherServerPlayerRow)
	return this
}
func (this *dbOtherServerPlayerTable) check_create_table() (err error) {
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS OtherServerPlayers(PlayerId int(11),PRIMARY KEY (PlayerId))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS OtherServerPlayers failed")
		return
	}
	rows, err := this.m_dbc.Query("SELECT COLUMN_NAME,ORDINAL_POSITION FROM information_schema.`COLUMNS` WHERE TABLE_SCHEMA=? AND TABLE_NAME='OtherServerPlayers'", this.m_dbc.m_db_name)
	if err != nil {
		log.Error("SELECT information_schema failed")
		return
	}
	columns := make(map[string]int32)
	for rows.Next() {
		var column_name string
		var ordinal_position int32
		err = rows.Scan(&column_name, &ordinal_position)
		if err != nil {
			log.Error("scan information_schema row failed")
			return
		}
		if ordinal_position < 1 {
			log.Error("col ordinal out of range")
			continue
		}
		columns[column_name] = ordinal_position
	}
	_, hasAccount := columns["Account"]
	if !hasAccount {
		_, err = this.m_dbc.Exec("ALTER TABLE OtherServerPlayers ADD COLUMN Account varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN Account failed")
			return
		}
	}
	_, hasName := columns["Name"]
	if !hasName {
		_, err = this.m_dbc.Exec("ALTER TABLE OtherServerPlayers ADD COLUMN Name varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN Name failed")
			return
		}
	}
	_, hasLevel := columns["Level"]
	if !hasLevel {
		_, err = this.m_dbc.Exec("ALTER TABLE OtherServerPlayers ADD COLUMN Level int(11)")
		if err != nil {
			log.Error("ADD COLUMN Level failed")
			return
		}
	}
	_, hasHead := columns["Head"]
	if !hasHead {
		_, err = this.m_dbc.Exec("ALTER TABLE OtherServerPlayers ADD COLUMN Head varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN Head failed")
			return
		}
	}
	return
}
func (this *dbOtherServerPlayerTable) prepare_preload_select_stmt() (err error) {
	this.m_preload_select_stmt,err=this.m_dbc.StmtPrepare("SELECT PlayerId,Account,Name,Level,Head FROM OtherServerPlayers")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbOtherServerPlayerTable) prepare_save_insert_stmt()(err error){
	this.m_save_insert_stmt,err=this.m_dbc.StmtPrepare("INSERT INTO OtherServerPlayers (PlayerId,Account,Name,Level,Head) VALUES (?,?,?,?,?)")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbOtherServerPlayerTable) prepare_delete_stmt() (err error) {
	this.m_delete_stmt,err=this.m_dbc.StmtPrepare("DELETE FROM OtherServerPlayers WHERE PlayerId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbOtherServerPlayerTable) Init() (err error) {
	err=this.check_create_table()
	if err!=nil{
		log.Error("check_create_table failed")
		return
	}
	err=this.prepare_preload_select_stmt()
	if err!=nil{
		log.Error("prepare_preload_select_stmt failed")
		return
	}
	err=this.prepare_save_insert_stmt()
	if err!=nil{
		log.Error("prepare_save_insert_stmt failed")
		return
	}
	err=this.prepare_delete_stmt()
	if err!=nil{
		log.Error("prepare_save_insert_stmt failed")
		return
	}
	return
}
func (this *dbOtherServerPlayerTable) Preload() (err error) {
	r, err := this.m_dbc.StmtQuery(this.m_preload_select_stmt)
	if err != nil {
		log.Error("SELECT")
		return
	}
	var PlayerId int32
	var dAccount string
	var dName string
	var dLevel int32
	var dHead string
		this.m_preload_max_id = 0
	for r.Next() {
		err = r.Scan(&PlayerId,&dAccount,&dName,&dLevel,&dHead)
		if err != nil {
			log.Error("Scan")
			return
		}
		if PlayerId>this.m_preload_max_id{
			this.m_preload_max_id =PlayerId
		}
		row := new_dbOtherServerPlayerRow(this,PlayerId)
		row.m_Account=dAccount
		row.m_Name=dName
		row.m_Level=dLevel
		row.m_Head=dHead
		row.m_Account_changed=false
		row.m_Name_changed=false
		row.m_Level_changed=false
		row.m_Head_changed=false
		row.m_valid = true
		this.m_rows[PlayerId]=row
	}
	return
}
func (this *dbOtherServerPlayerTable) GetPreloadedMaxId() (max_id int32) {
	return this.m_preload_max_id
}
func (this *dbOtherServerPlayerTable) fetch_rows(rows map[int32]*dbOtherServerPlayerRow) (r map[int32]*dbOtherServerPlayerRow) {
	this.m_lock.UnSafeLock("dbOtherServerPlayerTable.fetch_rows")
	defer this.m_lock.UnSafeUnlock()
	r = make(map[int32]*dbOtherServerPlayerRow)
	for i, v := range rows {
		r[i] = v
	}
	return r
}
func (this *dbOtherServerPlayerTable) fetch_new_rows() (new_rows map[int32]*dbOtherServerPlayerRow) {
	this.m_lock.UnSafeLock("dbOtherServerPlayerTable.fetch_new_rows")
	defer this.m_lock.UnSafeUnlock()
	new_rows = make(map[int32]*dbOtherServerPlayerRow)
	for i, v := range this.m_new_rows {
		_, has := this.m_rows[i]
		if has {
			log.Error("rows already has new rows %v", i)
			continue
		}
		this.m_rows[i] = v
		new_rows[i] = v
	}
	for i, _ := range new_rows {
		delete(this.m_new_rows, i)
	}
	return
}
func (this *dbOtherServerPlayerTable) save_rows(rows map[int32]*dbOtherServerPlayerRow, quick bool) {
	for _, v := range rows {
		if this.m_dbc.m_quit && !quick {
			return
		}
		err, delay, _ := v.Save(false)
		if err != nil {
			log.Error("save failed %v", err)
		}
		if this.m_dbc.m_quit && !quick {
			return
		}
		if delay&&!quick {
			time.Sleep(time.Millisecond * 5)
		}
	}
}
func (this *dbOtherServerPlayerTable) Save(quick bool) (err error){
	removed_rows := this.fetch_rows(this.m_removed_rows)
	for _, v := range removed_rows {
		_, err := this.m_dbc.StmtExec(this.m_delete_stmt, v.GetPlayerId())
		if err != nil {
			log.Error("exec delete stmt failed %v", err)
		}
		v.m_valid = false
		if !quick {
			time.Sleep(time.Millisecond * 5)
		}
	}
	this.m_removed_rows = make(map[int32]*dbOtherServerPlayerRow)
	rows := this.fetch_rows(this.m_rows)
	this.save_rows(rows, quick)
	new_rows := this.fetch_new_rows()
	this.save_rows(new_rows, quick)
	return
}
func (this *dbOtherServerPlayerTable) AddRow(PlayerId int32) (row *dbOtherServerPlayerRow) {
	this.m_lock.UnSafeLock("dbOtherServerPlayerTable.AddRow")
	defer this.m_lock.UnSafeUnlock()
	row = new_dbOtherServerPlayerRow(this,PlayerId)
	row.m_new = true
	row.m_loaded = true
	row.m_valid = true
	_, has := this.m_new_rows[PlayerId]
	if has{
		log.Error("已经存在 %v", PlayerId)
		return nil
	}
	this.m_new_rows[PlayerId] = row
	atomic.AddInt32(&this.m_gc_n,1)
	return row
}
func (this *dbOtherServerPlayerTable) RemoveRow(PlayerId int32) {
	this.m_lock.UnSafeLock("dbOtherServerPlayerTable.RemoveRow")
	defer this.m_lock.UnSafeUnlock()
	row := this.m_rows[PlayerId]
	if row != nil {
		row.m_remove = true
		delete(this.m_rows, PlayerId)
		rm_row := this.m_removed_rows[PlayerId]
		if rm_row != nil {
			log.Error("rows and removed rows both has %v", PlayerId)
		}
		this.m_removed_rows[PlayerId] = row
		_, has_new := this.m_new_rows[PlayerId]
		if has_new {
			delete(this.m_new_rows, PlayerId)
			log.Error("rows and new_rows both has %v", PlayerId)
		}
	} else {
		row = this.m_removed_rows[PlayerId]
		if row == nil {
			_, has_new := this.m_new_rows[PlayerId]
			if has_new {
				delete(this.m_new_rows, PlayerId)
			} else {
				log.Error("row not exist %v", PlayerId)
			}
		} else {
			log.Error("already removed %v", PlayerId)
			_, has_new := this.m_new_rows[PlayerId]
			if has_new {
				delete(this.m_new_rows, PlayerId)
				log.Error("removed rows and new_rows both has %v", PlayerId)
			}
		}
	}
}
func (this *dbOtherServerPlayerTable) GetRow(PlayerId int32) (row *dbOtherServerPlayerRow) {
	this.m_lock.UnSafeRLock("dbOtherServerPlayerTable.GetRow")
	defer this.m_lock.UnSafeRUnlock()
	row = this.m_rows[PlayerId]
	if row == nil {
		row = this.m_new_rows[PlayerId]
	}
	return row
}

type DBC struct {
	m_db_name            string
	m_db                 *sql.DB
	m_db_lock            *Mutex
	m_initialized        bool
	m_quit               bool
	m_shutdown_completed bool
	m_shutdown_lock      *Mutex
	m_db_last_copy_time	int32
	m_db_copy_path		string
	m_db_addr			string
	m_db_account			string
	m_db_password		string
	Global *dbGlobalTable
	Players *dbPlayerTable
	GooglePayRecords *dbGooglePayRecordTable
	ApplePayRecords *dbApplePayRecordTable
	FaceBPayRecords *dbFaceBPayRecordTable
	ServerInfo *dbServerInfoTable
	PlayerLogins *dbPlayerLoginTable
	OtherServerPlayers *dbOtherServerPlayerTable
}
func (this *DBC)init_tables()(err error){
	this.Global = new_dbGlobalTable(this)
	err = this.Global.Init()
	if err != nil {
		log.Error("init Global table failed")
		return
	}
	this.Players = new_dbPlayerTable(this)
	err = this.Players.Init()
	if err != nil {
		log.Error("init Players table failed")
		return
	}
	this.GooglePayRecords = new_dbGooglePayRecordTable(this)
	err = this.GooglePayRecords.Init()
	if err != nil {
		log.Error("init GooglePayRecords table failed")
		return
	}
	this.ApplePayRecords = new_dbApplePayRecordTable(this)
	err = this.ApplePayRecords.Init()
	if err != nil {
		log.Error("init ApplePayRecords table failed")
		return
	}
	this.FaceBPayRecords = new_dbFaceBPayRecordTable(this)
	err = this.FaceBPayRecords.Init()
	if err != nil {
		log.Error("init FaceBPayRecords table failed")
		return
	}
	this.ServerInfo = new_dbServerInfoTable(this)
	err = this.ServerInfo.Init()
	if err != nil {
		log.Error("init ServerInfo table failed")
		return
	}
	this.PlayerLogins = new_dbPlayerLoginTable(this)
	err = this.PlayerLogins.Init()
	if err != nil {
		log.Error("init PlayerLogins table failed")
		return
	}
	this.OtherServerPlayers = new_dbOtherServerPlayerTable(this)
	err = this.OtherServerPlayers.Init()
	if err != nil {
		log.Error("init OtherServerPlayers table failed")
		return
	}
	return
}
func (this *DBC)Preload()(err error){
	err = this.Global.Preload()
	if err != nil {
		log.Error("preload Global table failed")
		return
	}else{
		log.Info("preload Global table succeed !")
	}
	err = this.Players.Preload()
	if err != nil {
		log.Error("preload Players table failed")
		return
	}else{
		log.Info("preload Players table succeed !")
	}
	err = this.GooglePayRecords.Preload()
	if err != nil {
		log.Error("preload GooglePayRecords table failed")
		return
	}else{
		log.Info("preload GooglePayRecords table succeed !")
	}
	err = this.ApplePayRecords.Preload()
	if err != nil {
		log.Error("preload ApplePayRecords table failed")
		return
	}else{
		log.Info("preload ApplePayRecords table succeed !")
	}
	err = this.FaceBPayRecords.Preload()
	if err != nil {
		log.Error("preload FaceBPayRecords table failed")
		return
	}else{
		log.Info("preload FaceBPayRecords table succeed !")
	}
	err = this.ServerInfo.Preload()
	if err != nil {
		log.Error("preload ServerInfo table failed")
		return
	}else{
		log.Info("preload ServerInfo table succeed !")
	}
	err = this.PlayerLogins.Preload()
	if err != nil {
		log.Error("preload PlayerLogins table failed")
		return
	}else{
		log.Info("preload PlayerLogins table succeed !")
	}
	err = this.OtherServerPlayers.Preload()
	if err != nil {
		log.Error("preload OtherServerPlayers table failed")
		return
	}else{
		log.Info("preload OtherServerPlayers table succeed !")
	}
	err = this.on_preload()
	if err != nil {
		log.Error("on_preload failed")
		return
	}
	err = this.Save(true)
	if err != nil {
		log.Error("save on preload failed")
		return
	}
	return
}
func (this *DBC)Save(quick bool)(err error){
	err = this.Global.Save(quick)
	if err != nil {
		log.Error("save Global table failed")
		return
	}
	err = this.Players.Save(quick)
	if err != nil {
		log.Error("save Players table failed")
		return
	}
	err = this.GooglePayRecords.Save(quick)
	if err != nil {
		log.Error("save GooglePayRecords table failed")
		return
	}
	err = this.ApplePayRecords.Save(quick)
	if err != nil {
		log.Error("save ApplePayRecords table failed")
		return
	}
	err = this.FaceBPayRecords.Save(quick)
	if err != nil {
		log.Error("save FaceBPayRecords table failed")
		return
	}
	err = this.ServerInfo.Save(quick)
	if err != nil {
		log.Error("save ServerInfo table failed")
		return
	}
	err = this.PlayerLogins.Save(quick)
	if err != nil {
		log.Error("save PlayerLogins table failed")
		return
	}
	err = this.OtherServerPlayers.Save(quick)
	if err != nil {
		log.Error("save OtherServerPlayers table failed")
		return
	}
	return
}
