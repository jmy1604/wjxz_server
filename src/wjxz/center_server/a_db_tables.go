package main

import (
	"sort"
	"3p/code.google.com.protobuf/proto"
	_ "3p/mysql"
	"database/sql"
	"errors"
	"fmt"
	"libs/log"
	"math/rand"
	"os"
	"public_message/gen_go/db_center"
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

type dbHeroDesireParamData struct{
	TargetValue int32
	Value1 int32
	Value2 int32
}
func (this* dbHeroDesireParamData)from_pb(pb *db.HeroDesireParam){
	if pb == nil {
		return
	}
	this.TargetValue = pb.GetTargetValue()
	this.Value1 = pb.GetValue1()
	this.Value2 = pb.GetValue2()
	return
}
func (this* dbHeroDesireParamData)to_pb()(pb *db.HeroDesireParam){
	pb = &db.HeroDesireParam{}
	pb.TargetValue = proto.Int32(this.TargetValue)
	pb.Value1 = proto.Int32(this.Value1)
	pb.Value2 = proto.Int32(this.Value2)
	return
}
func (this* dbHeroDesireParamData)clone_to(d *dbHeroDesireParamData){
	d.TargetValue = this.TargetValue
	d.Value1 = this.Value1
	d.Value2 = this.Value2
	return
}
type dbSmallRankRecordData struct{
	Rank int32
	Id int32
	Val int32
	Name string
	Lvl int32
	Icon int32
	CustomIcon string
}
func (this* dbSmallRankRecordData)from_pb(pb *db.SmallRankRecord){
	if pb == nil {
		return
	}
	this.Rank = pb.GetRank()
	this.Id = pb.GetId()
	this.Val = pb.GetVal()
	this.Name = pb.GetName()
	this.Lvl = pb.GetLvl()
	this.Icon = pb.GetIcon()
	this.CustomIcon = pb.GetCustomIcon()
	return
}
func (this* dbSmallRankRecordData)to_pb()(pb *db.SmallRankRecord){
	pb = &db.SmallRankRecord{}
	pb.Rank = proto.Int32(this.Rank)
	pb.Id = proto.Int32(this.Id)
	pb.Val = proto.Int32(this.Val)
	pb.Name = proto.String(this.Name)
	pb.Lvl = proto.Int32(this.Lvl)
	pb.Icon = proto.Int32(this.Icon)
	pb.CustomIcon = proto.String(this.CustomIcon)
	return
}
func (this* dbSmallRankRecordData)clone_to(d *dbSmallRankRecordData){
	d.Rank = this.Rank
	d.Id = this.Id
	d.Val = this.Val
	d.Name = this.Name
	d.Lvl = this.Lvl
	d.Icon = this.Icon
	d.CustomIcon = this.CustomIcon
	return
}
type dbCampFightRecordData struct{
	FightIdx int32
	XScore int32
	TScore int32
}
func (this* dbCampFightRecordData)from_pb(pb *db.CampFightRecord){
	if pb == nil {
		return
	}
	this.FightIdx = pb.GetFightIdx()
	this.XScore = pb.GetXScore()
	this.TScore = pb.GetTScore()
	return
}
func (this* dbCampFightRecordData)to_pb()(pb *db.CampFightRecord){
	pb = &db.CampFightRecord{}
	pb.FightIdx = proto.Int32(this.FightIdx)
	pb.XScore = proto.Int32(this.XScore)
	pb.TScore = proto.Int32(this.TScore)
	return
}
func (this* dbCampFightRecordData)clone_to(d *dbCampFightRecordData){
	d.FightIdx = this.FightIdx
	d.XScore = this.XScore
	d.TScore = this.TScore
	return
}
type dbIdNumData struct{
	Id int32
	Num int32
}
func (this* dbIdNumData)from_pb(pb *db.IdNum){
	if pb == nil {
		return
	}
	this.Id = pb.GetId()
	this.Num = pb.GetNum()
	return
}
func (this* dbIdNumData)to_pb()(pb *db.IdNum){
	pb = &db.IdNum{}
	pb.Id = proto.Int32(this.Id)
	pb.Num = proto.Int32(this.Num)
	return
}
func (this* dbIdNumData)clone_to(d *dbIdNumData){
	d.Id = this.Id
	d.Num = this.Num
	return
}
type dbStageScoreRankRankRecordsData struct{
	Records []dbSmallRankRecordData
}
func (this* dbStageScoreRankRankRecordsData)from_pb(pb *db.StageScoreRankRankRecords){
	if pb == nil {
		this.Records = make([]dbSmallRankRecordData,0)
		return
	}
	this.Records = make([]dbSmallRankRecordData,len(pb.GetRecords()))
	for i, v := range pb.GetRecords() {
		this.Records[i].from_pb(v)
	}
	return
}
func (this* dbStageScoreRankRankRecordsData)to_pb()(pb *db.StageScoreRankRankRecords){
	pb = &db.StageScoreRankRankRecords{}
	pb.Records = make([]*db.SmallRankRecord, len(this.Records))
	for i, v := range this.Records {
		pb.Records[i]=v.to_pb()
	}
	return
}
func (this* dbStageScoreRankRankRecordsData)clone_to(d *dbStageScoreRankRankRecordsData){
	d.Records = make([]dbSmallRankRecordData, len(this.Records))
	for _ii, _vv := range this.Records {
		_vv.clone_to(&d.Records[_ii])
	}
	return
}
type dbServerRewardRewardInfoData struct{
	RewardId int32
	Items []dbIdNumData
	Channel string
	EndUnix int32
	Content string
}
func (this* dbServerRewardRewardInfoData)from_pb(pb *db.ServerRewardRewardInfo){
	if pb == nil {
		this.Items = make([]dbIdNumData,0)
		return
	}
	this.RewardId = pb.GetRewardId()
	this.Items = make([]dbIdNumData,len(pb.GetItems()))
	for i, v := range pb.GetItems() {
		this.Items[i].from_pb(v)
	}
	this.Channel = pb.GetChannel()
	this.EndUnix = pb.GetEndUnix()
	this.Content = pb.GetContent()
	return
}
func (this* dbServerRewardRewardInfoData)to_pb()(pb *db.ServerRewardRewardInfo){
	pb = &db.ServerRewardRewardInfo{}
	pb.RewardId = proto.Int32(this.RewardId)
	pb.Items = make([]*db.IdNum, len(this.Items))
	for i, v := range this.Items {
		pb.Items[i]=v.to_pb()
	}
	pb.Channel = proto.String(this.Channel)
	pb.EndUnix = proto.Int32(this.EndUnix)
	pb.Content = proto.String(this.Content)
	return
}
func (this* dbServerRewardRewardInfoData)clone_to(d *dbServerRewardRewardInfoData){
	d.RewardId = this.RewardId
	d.Items = make([]dbIdNumData, len(this.Items))
	for _ii, _vv := range this.Items {
		_vv.clone_to(&d.Items[_ii])
	}
	d.Channel = this.Channel
	d.EndUnix = this.EndUnix
	d.Content = this.Content
	return
}

type dbStageScoreRankRankRecordsColumn struct{
	m_row *dbStageScoreRankRow
	m_data *dbStageScoreRankRankRecordsData
	m_changed bool
}
func (this *dbStageScoreRankRankRecordsColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_data = &dbStageScoreRankRankRecordsData{}
		this.m_changed = false
		return nil
	}
	pb := &db.StageScoreRankRankRecords{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetStageId())
		return
	}
	this.m_data = &dbStageScoreRankRankRecordsData{}
	this.m_data.from_pb(pb)
	this.m_changed = false
	return
}
func (this *dbStageScoreRankRankRecordsColumn)save( )(data []byte,err error){
	pb:=this.m_data.to_pb()
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetStageId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbStageScoreRankRankRecordsColumn)Get( )(v *dbStageScoreRankRankRecordsData ){
	this.m_row.m_lock.UnSafeRLock("dbStageScoreRankRankRecordsColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v=&dbStageScoreRankRankRecordsData{}
	this.m_data.clone_to(v)
	return
}
func (this *dbStageScoreRankRankRecordsColumn)Set(v dbStageScoreRankRankRecordsData ){
	this.m_row.m_lock.UnSafeLock("dbStageScoreRankRankRecordsColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=&dbStageScoreRankRankRecordsData{}
	v.clone_to(this.m_data)
	this.m_changed=true
	return
}
func (this *dbStageScoreRankRankRecordsColumn)GetRecords( )(v []dbSmallRankRecordData ){
	this.m_row.m_lock.UnSafeRLock("dbStageScoreRankRankRecordsColumn.GetRecords")
	defer this.m_row.m_lock.UnSafeRUnlock()
	v = make([]dbSmallRankRecordData, len(this.m_data.Records))
	for _ii, _vv := range this.m_data.Records {
		_vv.clone_to(&v[_ii])
	}
	return
}
func (this *dbStageScoreRankRankRecordsColumn)SetRecords(v []dbSmallRankRecordData){
	this.m_row.m_lock.UnSafeLock("dbStageScoreRankRankRecordsColumn.SetRecords")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data.Records = make([]dbSmallRankRecordData, len(v))
	for _ii, _vv := range v {
		_vv.clone_to(&this.m_data.Records[_ii])
	}
	this.m_changed = true
	return
}
type dbStageScoreRankRow struct {
	m_table *dbStageScoreRankTable
	m_lock       *RWMutex
	m_loaded  bool
	m_new     bool
	m_remove  bool
	m_touch      int32
	m_releasable bool
	m_valid   bool
	m_StageId        int32
	RankRecords dbStageScoreRankRankRecordsColumn
}
func new_dbStageScoreRankRow(table *dbStageScoreRankTable, StageId int32) (r *dbStageScoreRankRow) {
	this := &dbStageScoreRankRow{}
	this.m_table = table
	this.m_StageId = StageId
	this.m_lock = NewRWMutex()
	this.RankRecords.m_row=this
	this.RankRecords.m_data=&dbStageScoreRankRankRecordsData{}
	return this
}
func (this *dbStageScoreRankRow) GetStageId() (r int32) {
	return this.m_StageId
}
func (this *dbStageScoreRankRow) save_data(release bool) (err error, released bool, state int32, update_string string, args []interface{}) {
	this.m_lock.UnSafeLock("dbStageScoreRankRow.save_data")
	defer this.m_lock.UnSafeUnlock()
	if this.m_new {
		db_args:=new_db_args(2)
		db_args.Push(this.m_StageId)
		dRankRecords,db_err:=this.RankRecords.save()
		if db_err!=nil{
			log.Error("insert save RankRecords failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dRankRecords)
		args=db_args.GetArgs()
		state = 1
	} else {
		if this.RankRecords.m_changed{
			update_string = "UPDATE StageScoreRanks SET "
			db_args:=new_db_args(2)
			if this.RankRecords.m_changed{
				update_string+="RankRecords=?,"
				dRankRecords,err:=this.RankRecords.save()
				if err!=nil{
					log.Error("update save RankRecords failed")
					return err,false,0,"",nil
				}
				db_args.Push(dRankRecords)
			}
			update_string = strings.TrimRight(update_string, ", ")
			update_string+=" WHERE StageId=?"
			db_args.Push(this.m_StageId)
			args=db_args.GetArgs()
			state = 2
		}
	}
	this.m_new = false
	this.RankRecords.m_changed = false
	if release && this.m_loaded {
		atomic.AddInt32(&this.m_table.m_gc_n, -1)
		this.m_loaded = false
		released = true
	}
	return nil,released,state,update_string,args
}
func (this *dbStageScoreRankRow) Save(release bool) (err error, d bool, released bool) {
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
			log.Error("INSERT StageScoreRanks exec failed %v ", this.m_StageId)
			return err, false, released
		}
		d = true
	} else if state == 2 {
		_, err = this.m_table.m_dbc.Exec(update_string, args...)
		if err != nil {
			log.Error("UPDATE StageScoreRanks exec failed %v", this.m_StageId)
			return err, false, released
		}
		d = true
	}
	return nil, d, released
}
func (this *dbStageScoreRankRow) Touch(releasable bool) {
	this.m_touch = int32(time.Now().Unix())
	this.m_releasable = releasable
}
type dbStageScoreRankRowSort struct {
	rows []*dbStageScoreRankRow
}
func (this *dbStageScoreRankRowSort) Len() (length int) {
	return len(this.rows)
}
func (this *dbStageScoreRankRowSort) Less(i int, j int) (less bool) {
	return this.rows[i].m_touch < this.rows[j].m_touch
}
func (this *dbStageScoreRankRowSort) Swap(i int, j int) {
	temp := this.rows[i]
	this.rows[i] = this.rows[j]
	this.rows[j] = temp
}
type dbStageScoreRankTable struct{
	m_dbc *DBC
	m_lock *RWMutex
	m_rows map[int32]*dbStageScoreRankRow
	m_new_rows map[int32]*dbStageScoreRankRow
	m_removed_rows map[int32]*dbStageScoreRankRow
	m_gc_n int32
	m_gcing int32
	m_pool_size int32
	m_preload_select_stmt *sql.Stmt
	m_preload_max_id int32
	m_save_insert_stmt *sql.Stmt
	m_delete_stmt *sql.Stmt
}
func new_dbStageScoreRankTable(dbc *DBC) (this *dbStageScoreRankTable) {
	this = &dbStageScoreRankTable{}
	this.m_dbc = dbc
	this.m_lock = NewRWMutex()
	this.m_rows = make(map[int32]*dbStageScoreRankRow)
	this.m_new_rows = make(map[int32]*dbStageScoreRankRow)
	this.m_removed_rows = make(map[int32]*dbStageScoreRankRow)
	return this
}
func (this *dbStageScoreRankTable) check_create_table() (err error) {
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS StageScoreRanks(StageId int(11),PRIMARY KEY (StageId))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS StageScoreRanks failed")
		return
	}
	rows, err := this.m_dbc.Query("SELECT COLUMN_NAME,ORDINAL_POSITION FROM information_schema.`COLUMNS` WHERE TABLE_SCHEMA=? AND TABLE_NAME='StageScoreRanks'", this.m_dbc.m_db_name)
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
	_, hasRankRecords := columns["RankRecords"]
	if !hasRankRecords {
		_, err = this.m_dbc.Exec("ALTER TABLE StageScoreRanks ADD COLUMN RankRecords LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN RankRecords failed")
			return
		}
	}
	return
}
func (this *dbStageScoreRankTable) prepare_preload_select_stmt() (err error) {
	this.m_preload_select_stmt,err=this.m_dbc.StmtPrepare("SELECT StageId,RankRecords FROM StageScoreRanks")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbStageScoreRankTable) prepare_save_insert_stmt()(err error){
	this.m_save_insert_stmt,err=this.m_dbc.StmtPrepare("INSERT INTO StageScoreRanks (StageId,RankRecords) VALUES (?,?)")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbStageScoreRankTable) prepare_delete_stmt() (err error) {
	this.m_delete_stmt,err=this.m_dbc.StmtPrepare("DELETE FROM StageScoreRanks WHERE StageId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbStageScoreRankTable) Init() (err error) {
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
func (this *dbStageScoreRankTable) Preload() (err error) {
	r, err := this.m_dbc.StmtQuery(this.m_preload_select_stmt)
	if err != nil {
		log.Error("SELECT")
		return
	}
	var StageId int32
	var dRankRecords []byte
		this.m_preload_max_id = 0
	for r.Next() {
		err = r.Scan(&StageId,&dRankRecords)
		if err != nil {
			log.Error("Scan")
			return
		}
		if StageId>this.m_preload_max_id{
			this.m_preload_max_id =StageId
		}
		row := new_dbStageScoreRankRow(this,StageId)
		err = row.RankRecords.load(dRankRecords)
		if err != nil {
			log.Error("RankRecords %v", StageId)
			return
		}
		row.m_valid = true
		this.m_rows[StageId]=row
	}
	return
}
func (this *dbStageScoreRankTable) GetPreloadedMaxId() (max_id int32) {
	return this.m_preload_max_id
}
func (this *dbStageScoreRankTable) fetch_rows(rows map[int32]*dbStageScoreRankRow) (r map[int32]*dbStageScoreRankRow) {
	this.m_lock.UnSafeLock("dbStageScoreRankTable.fetch_rows")
	defer this.m_lock.UnSafeUnlock()
	r = make(map[int32]*dbStageScoreRankRow)
	for i, v := range rows {
		r[i] = v
	}
	return r
}
func (this *dbStageScoreRankTable) fetch_new_rows() (new_rows map[int32]*dbStageScoreRankRow) {
	this.m_lock.UnSafeLock("dbStageScoreRankTable.fetch_new_rows")
	defer this.m_lock.UnSafeUnlock()
	new_rows = make(map[int32]*dbStageScoreRankRow)
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
func (this *dbStageScoreRankTable) save_rows(rows map[int32]*dbStageScoreRankRow, quick bool) {
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
func (this *dbStageScoreRankTable) Save(quick bool) (err error){
	removed_rows := this.fetch_rows(this.m_removed_rows)
	for _, v := range removed_rows {
		_, err := this.m_dbc.StmtExec(this.m_delete_stmt, v.GetStageId())
		if err != nil {
			log.Error("exec delete stmt failed %v", err)
		}
		v.m_valid = false
		if !quick {
			time.Sleep(time.Millisecond * 5)
		}
	}
	this.m_removed_rows = make(map[int32]*dbStageScoreRankRow)
	rows := this.fetch_rows(this.m_rows)
	this.save_rows(rows, quick)
	new_rows := this.fetch_new_rows()
	this.save_rows(new_rows, quick)
	return
}
func (this *dbStageScoreRankTable) AddRow(StageId int32) (row *dbStageScoreRankRow) {
	this.m_lock.UnSafeLock("dbStageScoreRankTable.AddRow")
	defer this.m_lock.UnSafeUnlock()
	row = new_dbStageScoreRankRow(this,StageId)
	row.m_new = true
	row.m_loaded = true
	row.m_valid = true
	_, has := this.m_new_rows[StageId]
	if has{
		log.Error("已经存在 %v", StageId)
		return nil
	}
	this.m_new_rows[StageId] = row
	atomic.AddInt32(&this.m_gc_n,1)
	return row
}
func (this *dbStageScoreRankTable) RemoveRow(StageId int32) {
	this.m_lock.UnSafeLock("dbStageScoreRankTable.RemoveRow")
	defer this.m_lock.UnSafeUnlock()
	row := this.m_rows[StageId]
	if row != nil {
		row.m_remove = true
		delete(this.m_rows, StageId)
		rm_row := this.m_removed_rows[StageId]
		if rm_row != nil {
			log.Error("rows and removed rows both has %v", StageId)
		}
		this.m_removed_rows[StageId] = row
		_, has_new := this.m_new_rows[StageId]
		if has_new {
			delete(this.m_new_rows, StageId)
			log.Error("rows and new_rows both has %v", StageId)
		}
	} else {
		row = this.m_removed_rows[StageId]
		if row == nil {
			_, has_new := this.m_new_rows[StageId]
			if has_new {
				delete(this.m_new_rows, StageId)
			} else {
				log.Error("row not exist %v", StageId)
			}
		} else {
			log.Error("already removed %v", StageId)
			_, has_new := this.m_new_rows[StageId]
			if has_new {
				delete(this.m_new_rows, StageId)
				log.Error("removed rows and new_rows both has %v", StageId)
			}
		}
	}
}
func (this *dbStageScoreRankTable) GetRow(StageId int32) (row *dbStageScoreRankRow) {
	this.m_lock.UnSafeRLock("dbStageScoreRankTable.GetRow")
	defer this.m_lock.UnSafeRUnlock()
	row = this.m_rows[StageId]
	if row == nil {
		row = this.m_new_rows[StageId]
	}
	return row
}
func (this *dbForbidTalkRow)GetPlayerId( )(r int32 ){
	this.m_lock.UnSafeRLock("dbForbidTalkRow.GetdbForbidTalkPlayerIdColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_PlayerId)
}
func (this *dbForbidTalkRow)SetPlayerId(v int32){
	this.m_lock.UnSafeLock("dbForbidTalkRow.SetdbForbidTalkPlayerIdColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_PlayerId=int32(v)
	this.m_PlayerId_changed=true
	return
}
func (this *dbForbidTalkRow)GetEndUnix( )(r int32 ){
	this.m_lock.UnSafeRLock("dbForbidTalkRow.GetdbForbidTalkEndUnixColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_EndUnix)
}
func (this *dbForbidTalkRow)SetEndUnix(v int32){
	this.m_lock.UnSafeLock("dbForbidTalkRow.SetdbForbidTalkEndUnixColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_EndUnix=int32(v)
	this.m_EndUnix_changed=true
	return
}
func (this *dbForbidTalkRow)GetReason( )(r string ){
	this.m_lock.UnSafeRLock("dbForbidTalkRow.GetdbForbidTalkReasonColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_Reason)
}
func (this *dbForbidTalkRow)SetReason(v string){
	this.m_lock.UnSafeLock("dbForbidTalkRow.SetdbForbidTalkReasonColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_Reason=string(v)
	this.m_Reason_changed=true
	return
}
type dbForbidTalkRow struct {
	m_table *dbForbidTalkTable
	m_lock       *RWMutex
	m_loaded  bool
	m_new     bool
	m_remove  bool
	m_touch      int32
	m_releasable bool
	m_valid   bool
	m_KeyId        int32
	m_PlayerId_changed bool
	m_PlayerId int32
	m_EndUnix_changed bool
	m_EndUnix int32
	m_Reason_changed bool
	m_Reason string
}
func new_dbForbidTalkRow(table *dbForbidTalkTable, KeyId int32) (r *dbForbidTalkRow) {
	this := &dbForbidTalkRow{}
	this.m_table = table
	this.m_KeyId = KeyId
	this.m_lock = NewRWMutex()
	this.m_PlayerId_changed=true
	this.m_EndUnix_changed=true
	this.m_Reason_changed=true
	return this
}
func (this *dbForbidTalkRow) GetKeyId() (r int32) {
	return this.m_KeyId
}
func (this *dbForbidTalkRow) Load() (err error) {
	this.m_table.GC()
	this.m_lock.UnSafeLock("dbForbidTalkRow.Load")
	defer this.m_lock.UnSafeUnlock()
	if this.m_loaded {
		return
	}
	var dEndUnix int32
	var dReason string
	r := this.m_table.m_dbc.StmtQueryRow(this.m_table.m_load_select_stmt, this.m_KeyId)
	err = r.Scan(&dEndUnix,&dReason)
	if err != nil {
		log.Error("scan")
		return
	}
		this.m_EndUnix=dEndUnix
		this.m_Reason=dReason
	this.m_loaded=true
	this.m_EndUnix_changed=false
	this.m_Reason_changed=false
	this.Touch(false)
	atomic.AddInt32(&this.m_table.m_gc_n,1)
	return
}
func (this *dbForbidTalkRow) save_data(release bool) (err error, released bool, state int32, update_string string, args []interface{}) {
	this.m_lock.UnSafeLock("dbForbidTalkRow.save_data")
	defer this.m_lock.UnSafeUnlock()
	if this.m_new {
		db_args:=new_db_args(4)
		db_args.Push(this.m_KeyId)
		db_args.Push(this.m_PlayerId)
		db_args.Push(this.m_EndUnix)
		db_args.Push(this.m_Reason)
		args=db_args.GetArgs()
		state = 1
	} else {
		if this.m_PlayerId_changed||this.m_EndUnix_changed||this.m_Reason_changed{
			update_string = "UPDATE ForbidTalks SET "
			db_args:=new_db_args(4)
			if this.m_PlayerId_changed{
				update_string+="PlayerId=?,"
				db_args.Push(this.m_PlayerId)
			}
			if this.m_EndUnix_changed{
				update_string+="EndUnix=?,"
				db_args.Push(this.m_EndUnix)
			}
			if this.m_Reason_changed{
				update_string+="Reason=?,"
				db_args.Push(this.m_Reason)
			}
			update_string = strings.TrimRight(update_string, ", ")
			update_string+=" WHERE KeyId=?"
			db_args.Push(this.m_KeyId)
			args=db_args.GetArgs()
			state = 2
		}
	}
	this.m_new = false
	this.m_PlayerId_changed = false
	this.m_EndUnix_changed = false
	this.m_Reason_changed = false
	if release && this.m_loaded {
		atomic.AddInt32(&this.m_table.m_gc_n, -1)
		this.m_loaded = false
		released = true
	}
	return nil,released,state,update_string,args
}
func (this *dbForbidTalkRow) Save(release bool) (err error, d bool, released bool) {
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
			log.Error("INSERT ForbidTalks exec failed %v ", this.m_KeyId)
			return err, false, released
		}
		d = true
	} else if state == 2 {
		_, err = this.m_table.m_dbc.Exec(update_string, args...)
		if err != nil {
			log.Error("UPDATE ForbidTalks exec failed %v", this.m_KeyId)
			return err, false, released
		}
		d = true
	}
	return nil, d, released
}
func (this *dbForbidTalkRow) Touch(releasable bool) {
	this.m_touch = int32(time.Now().Unix())
	this.m_releasable = releasable
}
type dbForbidTalkRowSort struct {
	rows []*dbForbidTalkRow
}
func (this *dbForbidTalkRowSort) Len() (length int) {
	return len(this.rows)
}
func (this *dbForbidTalkRowSort) Less(i int, j int) (less bool) {
	return this.rows[i].m_touch < this.rows[j].m_touch
}
func (this *dbForbidTalkRowSort) Swap(i int, j int) {
	temp := this.rows[i]
	this.rows[i] = this.rows[j]
	this.rows[j] = temp
}
type dbForbidTalkTable struct{
	m_dbc *DBC
	m_lock *RWMutex
	m_rows map[int32]*dbForbidTalkRow
	m_new_rows map[int32]*dbForbidTalkRow
	m_removed_rows map[int32]*dbForbidTalkRow
	m_gc_n int32
	m_gcing int32
	m_pool_size int32
	m_preload_select_stmt *sql.Stmt
	m_preload_max_id int32
	m_load_select_stmt *sql.Stmt
	m_save_insert_stmt *sql.Stmt
	m_delete_stmt *sql.Stmt
}
func new_dbForbidTalkTable(dbc *DBC) (this *dbForbidTalkTable) {
	this = &dbForbidTalkTable{}
	this.m_dbc = dbc
	this.m_lock = NewRWMutex()
	this.m_rows = make(map[int32]*dbForbidTalkRow)
	this.m_new_rows = make(map[int32]*dbForbidTalkRow)
	this.m_removed_rows = make(map[int32]*dbForbidTalkRow)
	return this
}
func (this *dbForbidTalkTable) check_create_table() (err error) {
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS ForbidTalks(KeyId int(11),PRIMARY KEY (KeyId))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS ForbidTalks failed")
		return
	}
	rows, err := this.m_dbc.Query("SELECT COLUMN_NAME,ORDINAL_POSITION FROM information_schema.`COLUMNS` WHERE TABLE_SCHEMA=? AND TABLE_NAME='ForbidTalks'", this.m_dbc.m_db_name)
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
	_, hasPlayerId := columns["PlayerId"]
	if !hasPlayerId {
		_, err = this.m_dbc.Exec("ALTER TABLE ForbidTalks ADD COLUMN PlayerId int(11)")
		if err != nil {
			log.Error("ADD COLUMN PlayerId failed")
			return
		}
	}
	_, hasEndUnix := columns["EndUnix"]
	if !hasEndUnix {
		_, err = this.m_dbc.Exec("ALTER TABLE ForbidTalks ADD COLUMN EndUnix int(11)")
		if err != nil {
			log.Error("ADD COLUMN EndUnix failed")
			return
		}
	}
	_, hasReason := columns["Reason"]
	if !hasReason {
		_, err = this.m_dbc.Exec("ALTER TABLE ForbidTalks ADD COLUMN Reason varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN Reason failed")
			return
		}
	}
	return
}
func (this *dbForbidTalkTable) prepare_preload_select_stmt() (err error) {
	this.m_preload_select_stmt,err=this.m_dbc.StmtPrepare("SELECT KeyId,PlayerId FROM ForbidTalks")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbForbidTalkTable) prepare_load_select_stmt() (err error) {
	this.m_load_select_stmt,err=this.m_dbc.StmtPrepare("SELECT EndUnix,Reason FROM ForbidTalks WHERE KeyId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbForbidTalkTable) prepare_save_insert_stmt()(err error){
	this.m_save_insert_stmt,err=this.m_dbc.StmtPrepare("INSERT INTO ForbidTalks (KeyId,PlayerId,EndUnix,Reason) VALUES (?,?,?,?)")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbForbidTalkTable) prepare_delete_stmt() (err error) {
	this.m_delete_stmt,err=this.m_dbc.StmtPrepare("DELETE FROM ForbidTalks WHERE KeyId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbForbidTalkTable) Init() (err error) {
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
func (this *dbForbidTalkTable) Preload() (err error) {
	r, err := this.m_dbc.StmtQuery(this.m_preload_select_stmt)
	if err != nil {
		log.Error("SELECT")
		return
	}
	var KeyId int32
	var dPlayerId int32
		this.m_preload_max_id = 0
	for r.Next() {
		err = r.Scan(&KeyId,&dPlayerId)
		if err != nil {
			log.Error("Scan")
			return
		}
		if KeyId>this.m_preload_max_id{
			this.m_preload_max_id =KeyId
		}
		row := new_dbForbidTalkRow(this,KeyId)
		row.m_PlayerId=dPlayerId
		row.m_PlayerId_changed=false
		row.m_valid = true
		this.m_rows[KeyId]=row
	}
	return
}
func (this *dbForbidTalkTable) GetPreloadedMaxId() (max_id int32) {
	return this.m_preload_max_id
}
func (this *dbForbidTalkTable) fetch_rows(rows map[int32]*dbForbidTalkRow) (r map[int32]*dbForbidTalkRow) {
	this.m_lock.UnSafeLock("dbForbidTalkTable.fetch_rows")
	defer this.m_lock.UnSafeUnlock()
	r = make(map[int32]*dbForbidTalkRow)
	for i, v := range rows {
		r[i] = v
	}
	return r
}
func (this *dbForbidTalkTable) fetch_new_rows() (new_rows map[int32]*dbForbidTalkRow) {
	this.m_lock.UnSafeLock("dbForbidTalkTable.fetch_new_rows")
	defer this.m_lock.UnSafeUnlock()
	new_rows = make(map[int32]*dbForbidTalkRow)
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
func (this *dbForbidTalkTable) save_rows(rows map[int32]*dbForbidTalkRow, quick bool) {
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
func (this *dbForbidTalkTable) Save(quick bool) (err error){
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
	this.m_removed_rows = make(map[int32]*dbForbidTalkRow)
	rows := this.fetch_rows(this.m_rows)
	this.save_rows(rows, quick)
	new_rows := this.fetch_new_rows()
	this.save_rows(new_rows, quick)
	return
}
func (this *dbForbidTalkTable) AddRow(KeyId int32) (row *dbForbidTalkRow) {
	this.GC()
	this.m_lock.UnSafeLock("dbForbidTalkTable.AddRow")
	defer this.m_lock.UnSafeUnlock()
	row = new_dbForbidTalkRow(this,KeyId)
	row.m_new = true
	row.m_loaded = true
	row.m_valid = true
	_, has := this.m_new_rows[KeyId]
	if has{
		log.Error("已经存在 %v", KeyId)
		return nil
	}
	this.m_new_rows[KeyId] = row
	atomic.AddInt32(&this.m_gc_n,1)
	return row
}
func (this *dbForbidTalkTable) RemoveRow(KeyId int32) {
	this.m_lock.UnSafeLock("dbForbidTalkTable.RemoveRow")
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
func (this *dbForbidTalkTable) GetRow(KeyId int32) (row *dbForbidTalkRow) {
	this.m_lock.UnSafeRLock("dbForbidTalkTable.GetRow")
	defer this.m_lock.UnSafeRUnlock()
	row = this.m_rows[KeyId]
	if row == nil {
		row = this.m_new_rows[KeyId]
	}
	return row
}
func (this *dbForbidTalkTable) SetPoolSize(n int32) {
	this.m_pool_size = n
}
func (this *dbForbidTalkTable) GC() {
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
	arr := dbForbidTalkRowSort{}
	rows := this.fetch_rows(this.m_rows)
	arr.rows = make([]*dbForbidTalkRow, len(rows))
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
func (this *dbForbidLoginRow)GetPlayerId( )(r int32 ){
	this.m_lock.UnSafeRLock("dbForbidLoginRow.GetdbForbidLoginPlayerIdColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_PlayerId)
}
func (this *dbForbidLoginRow)SetPlayerId(v int32){
	this.m_lock.UnSafeLock("dbForbidLoginRow.SetdbForbidLoginPlayerIdColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_PlayerId=int32(v)
	this.m_PlayerId_changed=true
	return
}
func (this *dbForbidLoginRow)GetEndUnix( )(r int32 ){
	this.m_lock.UnSafeRLock("dbForbidLoginRow.GetdbForbidLoginEndUnixColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_EndUnix)
}
func (this *dbForbidLoginRow)SetEndUnix(v int32){
	this.m_lock.UnSafeLock("dbForbidLoginRow.SetdbForbidLoginEndUnixColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_EndUnix=int32(v)
	this.m_EndUnix_changed=true
	return
}
func (this *dbForbidLoginRow)GetReason( )(r string ){
	this.m_lock.UnSafeRLock("dbForbidLoginRow.GetdbForbidLoginReasonColumn")
	defer this.m_lock.UnSafeRUnlock()
	return string(this.m_Reason)
}
func (this *dbForbidLoginRow)SetReason(v string){
	this.m_lock.UnSafeLock("dbForbidLoginRow.SetdbForbidLoginReasonColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_Reason=string(v)
	this.m_Reason_changed=true
	return
}
type dbForbidLoginRow struct {
	m_table *dbForbidLoginTable
	m_lock       *RWMutex
	m_loaded  bool
	m_new     bool
	m_remove  bool
	m_touch      int32
	m_releasable bool
	m_valid   bool
	m_KeyId        int32
	m_PlayerId_changed bool
	m_PlayerId int32
	m_EndUnix_changed bool
	m_EndUnix int32
	m_Reason_changed bool
	m_Reason string
}
func new_dbForbidLoginRow(table *dbForbidLoginTable, KeyId int32) (r *dbForbidLoginRow) {
	this := &dbForbidLoginRow{}
	this.m_table = table
	this.m_KeyId = KeyId
	this.m_lock = NewRWMutex()
	this.m_PlayerId_changed=true
	this.m_EndUnix_changed=true
	this.m_Reason_changed=true
	return this
}
func (this *dbForbidLoginRow) GetKeyId() (r int32) {
	return this.m_KeyId
}
func (this *dbForbidLoginRow) Load() (err error) {
	this.m_table.GC()
	this.m_lock.UnSafeLock("dbForbidLoginRow.Load")
	defer this.m_lock.UnSafeUnlock()
	if this.m_loaded {
		return
	}
	var dEndUnix int32
	var dReason string
	r := this.m_table.m_dbc.StmtQueryRow(this.m_table.m_load_select_stmt, this.m_KeyId)
	err = r.Scan(&dEndUnix,&dReason)
	if err != nil {
		log.Error("scan")
		return
	}
		this.m_EndUnix=dEndUnix
		this.m_Reason=dReason
	this.m_loaded=true
	this.m_EndUnix_changed=false
	this.m_Reason_changed=false
	this.Touch(false)
	atomic.AddInt32(&this.m_table.m_gc_n,1)
	return
}
func (this *dbForbidLoginRow) save_data(release bool) (err error, released bool, state int32, update_string string, args []interface{}) {
	this.m_lock.UnSafeLock("dbForbidLoginRow.save_data")
	defer this.m_lock.UnSafeUnlock()
	if this.m_new {
		db_args:=new_db_args(4)
		db_args.Push(this.m_KeyId)
		db_args.Push(this.m_PlayerId)
		db_args.Push(this.m_EndUnix)
		db_args.Push(this.m_Reason)
		args=db_args.GetArgs()
		state = 1
	} else {
		if this.m_PlayerId_changed||this.m_EndUnix_changed||this.m_Reason_changed{
			update_string = "UPDATE ForbidLogins SET "
			db_args:=new_db_args(4)
			if this.m_PlayerId_changed{
				update_string+="PlayerId=?,"
				db_args.Push(this.m_PlayerId)
			}
			if this.m_EndUnix_changed{
				update_string+="EndUnix=?,"
				db_args.Push(this.m_EndUnix)
			}
			if this.m_Reason_changed{
				update_string+="Reason=?,"
				db_args.Push(this.m_Reason)
			}
			update_string = strings.TrimRight(update_string, ", ")
			update_string+=" WHERE KeyId=?"
			db_args.Push(this.m_KeyId)
			args=db_args.GetArgs()
			state = 2
		}
	}
	this.m_new = false
	this.m_PlayerId_changed = false
	this.m_EndUnix_changed = false
	this.m_Reason_changed = false
	if release && this.m_loaded {
		atomic.AddInt32(&this.m_table.m_gc_n, -1)
		this.m_loaded = false
		released = true
	}
	return nil,released,state,update_string,args
}
func (this *dbForbidLoginRow) Save(release bool) (err error, d bool, released bool) {
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
			log.Error("INSERT ForbidLogins exec failed %v ", this.m_KeyId)
			return err, false, released
		}
		d = true
	} else if state == 2 {
		_, err = this.m_table.m_dbc.Exec(update_string, args...)
		if err != nil {
			log.Error("UPDATE ForbidLogins exec failed %v", this.m_KeyId)
			return err, false, released
		}
		d = true
	}
	return nil, d, released
}
func (this *dbForbidLoginRow) Touch(releasable bool) {
	this.m_touch = int32(time.Now().Unix())
	this.m_releasable = releasable
}
type dbForbidLoginRowSort struct {
	rows []*dbForbidLoginRow
}
func (this *dbForbidLoginRowSort) Len() (length int) {
	return len(this.rows)
}
func (this *dbForbidLoginRowSort) Less(i int, j int) (less bool) {
	return this.rows[i].m_touch < this.rows[j].m_touch
}
func (this *dbForbidLoginRowSort) Swap(i int, j int) {
	temp := this.rows[i]
	this.rows[i] = this.rows[j]
	this.rows[j] = temp
}
type dbForbidLoginTable struct{
	m_dbc *DBC
	m_lock *RWMutex
	m_rows map[int32]*dbForbidLoginRow
	m_new_rows map[int32]*dbForbidLoginRow
	m_removed_rows map[int32]*dbForbidLoginRow
	m_gc_n int32
	m_gcing int32
	m_pool_size int32
	m_preload_select_stmt *sql.Stmt
	m_preload_max_id int32
	m_load_select_stmt *sql.Stmt
	m_save_insert_stmt *sql.Stmt
	m_delete_stmt *sql.Stmt
}
func new_dbForbidLoginTable(dbc *DBC) (this *dbForbidLoginTable) {
	this = &dbForbidLoginTable{}
	this.m_dbc = dbc
	this.m_lock = NewRWMutex()
	this.m_rows = make(map[int32]*dbForbidLoginRow)
	this.m_new_rows = make(map[int32]*dbForbidLoginRow)
	this.m_removed_rows = make(map[int32]*dbForbidLoginRow)
	return this
}
func (this *dbForbidLoginTable) check_create_table() (err error) {
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS ForbidLogins(KeyId int(11),PRIMARY KEY (KeyId))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS ForbidLogins failed")
		return
	}
	rows, err := this.m_dbc.Query("SELECT COLUMN_NAME,ORDINAL_POSITION FROM information_schema.`COLUMNS` WHERE TABLE_SCHEMA=? AND TABLE_NAME='ForbidLogins'", this.m_dbc.m_db_name)
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
	_, hasPlayerId := columns["PlayerId"]
	if !hasPlayerId {
		_, err = this.m_dbc.Exec("ALTER TABLE ForbidLogins ADD COLUMN PlayerId int(11)")
		if err != nil {
			log.Error("ADD COLUMN PlayerId failed")
			return
		}
	}
	_, hasEndUnix := columns["EndUnix"]
	if !hasEndUnix {
		_, err = this.m_dbc.Exec("ALTER TABLE ForbidLogins ADD COLUMN EndUnix int(11)")
		if err != nil {
			log.Error("ADD COLUMN EndUnix failed")
			return
		}
	}
	_, hasReason := columns["Reason"]
	if !hasReason {
		_, err = this.m_dbc.Exec("ALTER TABLE ForbidLogins ADD COLUMN Reason varchar(45)")
		if err != nil {
			log.Error("ADD COLUMN Reason failed")
			return
		}
	}
	return
}
func (this *dbForbidLoginTable) prepare_preload_select_stmt() (err error) {
	this.m_preload_select_stmt,err=this.m_dbc.StmtPrepare("SELECT KeyId,PlayerId FROM ForbidLogins")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbForbidLoginTable) prepare_load_select_stmt() (err error) {
	this.m_load_select_stmt,err=this.m_dbc.StmtPrepare("SELECT EndUnix,Reason FROM ForbidLogins WHERE KeyId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbForbidLoginTable) prepare_save_insert_stmt()(err error){
	this.m_save_insert_stmt,err=this.m_dbc.StmtPrepare("INSERT INTO ForbidLogins (KeyId,PlayerId,EndUnix,Reason) VALUES (?,?,?,?)")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbForbidLoginTable) prepare_delete_stmt() (err error) {
	this.m_delete_stmt,err=this.m_dbc.StmtPrepare("DELETE FROM ForbidLogins WHERE KeyId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbForbidLoginTable) Init() (err error) {
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
func (this *dbForbidLoginTable) Preload() (err error) {
	r, err := this.m_dbc.StmtQuery(this.m_preload_select_stmt)
	if err != nil {
		log.Error("SELECT")
		return
	}
	var KeyId int32
	var dPlayerId int32
		this.m_preload_max_id = 0
	for r.Next() {
		err = r.Scan(&KeyId,&dPlayerId)
		if err != nil {
			log.Error("Scan")
			return
		}
		if KeyId>this.m_preload_max_id{
			this.m_preload_max_id =KeyId
		}
		row := new_dbForbidLoginRow(this,KeyId)
		row.m_PlayerId=dPlayerId
		row.m_PlayerId_changed=false
		row.m_valid = true
		this.m_rows[KeyId]=row
	}
	return
}
func (this *dbForbidLoginTable) GetPreloadedMaxId() (max_id int32) {
	return this.m_preload_max_id
}
func (this *dbForbidLoginTable) fetch_rows(rows map[int32]*dbForbidLoginRow) (r map[int32]*dbForbidLoginRow) {
	this.m_lock.UnSafeLock("dbForbidLoginTable.fetch_rows")
	defer this.m_lock.UnSafeUnlock()
	r = make(map[int32]*dbForbidLoginRow)
	for i, v := range rows {
		r[i] = v
	}
	return r
}
func (this *dbForbidLoginTable) fetch_new_rows() (new_rows map[int32]*dbForbidLoginRow) {
	this.m_lock.UnSafeLock("dbForbidLoginTable.fetch_new_rows")
	defer this.m_lock.UnSafeUnlock()
	new_rows = make(map[int32]*dbForbidLoginRow)
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
func (this *dbForbidLoginTable) save_rows(rows map[int32]*dbForbidLoginRow, quick bool) {
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
func (this *dbForbidLoginTable) Save(quick bool) (err error){
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
	this.m_removed_rows = make(map[int32]*dbForbidLoginRow)
	rows := this.fetch_rows(this.m_rows)
	this.save_rows(rows, quick)
	new_rows := this.fetch_new_rows()
	this.save_rows(new_rows, quick)
	return
}
func (this *dbForbidLoginTable) AddRow(KeyId int32) (row *dbForbidLoginRow) {
	this.GC()
	this.m_lock.UnSafeLock("dbForbidLoginTable.AddRow")
	defer this.m_lock.UnSafeUnlock()
	row = new_dbForbidLoginRow(this,KeyId)
	row.m_new = true
	row.m_loaded = true
	row.m_valid = true
	_, has := this.m_new_rows[KeyId]
	if has{
		log.Error("已经存在 %v", KeyId)
		return nil
	}
	this.m_new_rows[KeyId] = row
	atomic.AddInt32(&this.m_gc_n,1)
	return row
}
func (this *dbForbidLoginTable) RemoveRow(KeyId int32) {
	this.m_lock.UnSafeLock("dbForbidLoginTable.RemoveRow")
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
func (this *dbForbidLoginTable) GetRow(KeyId int32) (row *dbForbidLoginRow) {
	this.m_lock.UnSafeRLock("dbForbidLoginTable.GetRow")
	defer this.m_lock.UnSafeRUnlock()
	row = this.m_rows[KeyId]
	if row == nil {
		row = this.m_new_rows[KeyId]
	}
	return row
}
func (this *dbForbidLoginTable) SetPoolSize(n int32) {
	this.m_pool_size = n
}
func (this *dbForbidLoginTable) GC() {
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
	arr := dbForbidLoginRowSort{}
	rows := this.fetch_rows(this.m_rows)
	arr.rows = make([]*dbForbidLoginRow, len(rows))
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
func (this *dbServerRewardRow)GetNextRewardId( )(r int32 ){
	this.m_lock.UnSafeRLock("dbServerRewardRow.GetdbServerRewardNextRewardIdColumn")
	defer this.m_lock.UnSafeRUnlock()
	return int32(this.m_NextRewardId)
}
func (this *dbServerRewardRow)SetNextRewardId(v int32){
	this.m_lock.UnSafeLock("dbServerRewardRow.SetdbServerRewardNextRewardIdColumn")
	defer this.m_lock.UnSafeUnlock()
	this.m_NextRewardId=int32(v)
	this.m_NextRewardId_changed=true
	return
}
type dbServerRewardRewardInfoColumn struct{
	m_row *dbServerRewardRow
	m_data map[int32]*dbServerRewardRewardInfoData
	m_changed bool
}
func (this *dbServerRewardRewardInfoColumn)load(data []byte)(err error){
	if data == nil || len(data) == 0 {
		this.m_changed = false
		return nil
	}
	pb := &db.ServerRewardRewardInfoList{}
	err = proto.Unmarshal(data, pb)
	if err != nil {
		log.Error("Unmarshal %v", this.m_row.GetKeyId())
		return
	}
	for _, v := range pb.List {
		d := &dbServerRewardRewardInfoData{}
		d.from_pb(v)
		this.m_data[int32(d.RewardId)] = d
	}
	this.m_changed = false
	return
}
func (this *dbServerRewardRewardInfoColumn)save( )(data []byte,err error){
	pb := &db.ServerRewardRewardInfoList{}
	pb.List=make([]*db.ServerRewardRewardInfo,len(this.m_data))
	i:=0
	for _, v := range this.m_data {
		pb.List[i] = v.to_pb()
		i++
	}
	data, err = proto.Marshal(pb)
	if err != nil {
		log.Error("Marshal %v", this.m_row.GetKeyId())
		return
	}
	this.m_changed = false
	return
}
func (this *dbServerRewardRewardInfoColumn)HasIndex(id int32)(has bool){
	this.m_row.m_lock.UnSafeRLock("dbServerRewardRewardInfoColumn.HasIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	_, has = this.m_data[id]
	return
}
func (this *dbServerRewardRewardInfoColumn)GetAllIndex()(list []int32){
	this.m_row.m_lock.UnSafeRLock("dbServerRewardRewardInfoColumn.GetAllIndex")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]int32, len(this.m_data))
	i := 0
	for k, _ := range this.m_data {
		list[i] = k
		i++
	}
	return
}
func (this *dbServerRewardRewardInfoColumn)GetAll()(list []dbServerRewardRewardInfoData){
	this.m_row.m_lock.UnSafeRLock("dbServerRewardRewardInfoColumn.GetAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	list = make([]dbServerRewardRewardInfoData, len(this.m_data))
	i := 0
	for _, v := range this.m_data {
		v.clone_to(&list[i])
		i++
	}
	return
}
func (this *dbServerRewardRewardInfoColumn)Get(id int32)(v *dbServerRewardRewardInfoData){
	this.m_row.m_lock.UnSafeRLock("dbServerRewardRewardInfoColumn.Get")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return nil
	}
	v=&dbServerRewardRewardInfoData{}
	d.clone_to(v)
	return
}
func (this *dbServerRewardRewardInfoColumn)Set(v dbServerRewardRewardInfoData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbServerRewardRewardInfoColumn.Set")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[int32(v.RewardId)]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetKeyId(), v.RewardId)
		return false
	}
	v.clone_to(d)
	this.m_changed = true
	return true
}
func (this *dbServerRewardRewardInfoColumn)Add(v *dbServerRewardRewardInfoData)(ok bool){
	this.m_row.m_lock.UnSafeLock("dbServerRewardRewardInfoColumn.Add")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[int32(v.RewardId)]
	if has {
		log.Error("already added %v %v",this.m_row.GetKeyId(), v.RewardId)
		return false
	}
	d:=&dbServerRewardRewardInfoData{}
	v.clone_to(d)
	this.m_data[int32(v.RewardId)]=d
	this.m_changed = true
	return true
}
func (this *dbServerRewardRewardInfoColumn)Remove(id int32){
	this.m_row.m_lock.UnSafeLock("dbServerRewardRewardInfoColumn.Remove")
	defer this.m_row.m_lock.UnSafeUnlock()
	_, has := this.m_data[id]
	if has {
		delete(this.m_data,id)
	}
	this.m_changed = true
	return
}
func (this *dbServerRewardRewardInfoColumn)Clear(){
	this.m_row.m_lock.UnSafeLock("dbServerRewardRewardInfoColumn.Clear")
	defer this.m_row.m_lock.UnSafeUnlock()
	this.m_data=make(map[int32]*dbServerRewardRewardInfoData)
	this.m_changed = true
	return
}
func (this *dbServerRewardRewardInfoColumn)NumAll()(n int32){
	this.m_row.m_lock.UnSafeRLock("dbServerRewardRewardInfoColumn.NumAll")
	defer this.m_row.m_lock.UnSafeRUnlock()
	return int32(len(this.m_data))
}
func (this *dbServerRewardRewardInfoColumn)GetItems(id int32)(v []dbIdNumData,has bool ){
	this.m_row.m_lock.UnSafeRLock("dbServerRewardRewardInfoColumn.GetItems")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = make([]dbIdNumData, len(d.Items))
	for _ii, _vv := range d.Items {
		_vv.clone_to(&v[_ii])
	}
	return v,true
}
func (this *dbServerRewardRewardInfoColumn)SetItems(id int32,v []dbIdNumData)(has bool){
	this.m_row.m_lock.UnSafeLock("dbServerRewardRewardInfoColumn.SetItems")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetKeyId(), id)
		return
	}
	d.Items = make([]dbIdNumData, len(v))
	for _ii, _vv := range v {
		_vv.clone_to(&d.Items[_ii])
	}
	this.m_changed = true
	return true
}
func (this *dbServerRewardRewardInfoColumn)GetChannel(id int32)(v string ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbServerRewardRewardInfoColumn.GetChannel")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.Channel
	return v,true
}
func (this *dbServerRewardRewardInfoColumn)SetChannel(id int32,v string)(has bool){
	this.m_row.m_lock.UnSafeLock("dbServerRewardRewardInfoColumn.SetChannel")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetKeyId(), id)
		return
	}
	d.Channel = v
	this.m_changed = true
	return true
}
func (this *dbServerRewardRewardInfoColumn)GetEndUnix(id int32)(v int32 ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbServerRewardRewardInfoColumn.GetEndUnix")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.EndUnix
	return v,true
}
func (this *dbServerRewardRewardInfoColumn)SetEndUnix(id int32,v int32)(has bool){
	this.m_row.m_lock.UnSafeLock("dbServerRewardRewardInfoColumn.SetEndUnix")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetKeyId(), id)
		return
	}
	d.EndUnix = v
	this.m_changed = true
	return true
}
func (this *dbServerRewardRewardInfoColumn)GetContent(id int32)(v string ,has bool){
	this.m_row.m_lock.UnSafeRLock("dbServerRewardRewardInfoColumn.GetContent")
	defer this.m_row.m_lock.UnSafeRUnlock()
	d := this.m_data[id]
	if d==nil{
		return
	}
	v = d.Content
	return v,true
}
func (this *dbServerRewardRewardInfoColumn)SetContent(id int32,v string)(has bool){
	this.m_row.m_lock.UnSafeLock("dbServerRewardRewardInfoColumn.SetContent")
	defer this.m_row.m_lock.UnSafeUnlock()
	d := this.m_data[id]
	if d==nil{
		log.Error("not exist %v %v",this.m_row.GetKeyId(), id)
		return
	}
	d.Content = v
	this.m_changed = true
	return true
}
type dbServerRewardRow struct {
	m_table *dbServerRewardTable
	m_lock       *RWMutex
	m_loaded  bool
	m_new     bool
	m_remove  bool
	m_touch      int32
	m_releasable bool
	m_valid   bool
	m_KeyId        int32
	m_NextRewardId_changed bool
	m_NextRewardId int32
	RewardInfos dbServerRewardRewardInfoColumn
}
func new_dbServerRewardRow(table *dbServerRewardTable, KeyId int32) (r *dbServerRewardRow) {
	this := &dbServerRewardRow{}
	this.m_table = table
	this.m_KeyId = KeyId
	this.m_lock = NewRWMutex()
	this.m_NextRewardId_changed=true
	this.RewardInfos.m_row=this
	this.RewardInfos.m_data=make(map[int32]*dbServerRewardRewardInfoData)
	return this
}
func (this *dbServerRewardRow) GetKeyId() (r int32) {
	return this.m_KeyId
}
func (this *dbServerRewardRow) save_data(release bool) (err error, released bool, state int32, update_string string, args []interface{}) {
	this.m_lock.UnSafeLock("dbServerRewardRow.save_data")
	defer this.m_lock.UnSafeUnlock()
	if this.m_new {
		db_args:=new_db_args(3)
		db_args.Push(this.m_KeyId)
		db_args.Push(this.m_NextRewardId)
		dRewardInfos,db_err:=this.RewardInfos.save()
		if db_err!=nil{
			log.Error("insert save RewardInfo failed")
			return db_err,false,0,"",nil
		}
		db_args.Push(dRewardInfos)
		args=db_args.GetArgs()
		state = 1
	} else {
		if this.m_NextRewardId_changed||this.RewardInfos.m_changed{
			update_string = "UPDATE ServerRewards SET "
			db_args:=new_db_args(3)
			if this.m_NextRewardId_changed{
				update_string+="NextRewardId=?,"
				db_args.Push(this.m_NextRewardId)
			}
			if this.RewardInfos.m_changed{
				update_string+="RewardInfos=?,"
				dRewardInfos,err:=this.RewardInfos.save()
				if err!=nil{
					log.Error("insert save RewardInfo failed")
					return err,false,0,"",nil
				}
				db_args.Push(dRewardInfos)
			}
			update_string = strings.TrimRight(update_string, ", ")
			update_string+=" WHERE KeyId=?"
			db_args.Push(this.m_KeyId)
			args=db_args.GetArgs()
			state = 2
		}
	}
	this.m_new = false
	this.m_NextRewardId_changed = false
	this.RewardInfos.m_changed = false
	if release && this.m_loaded {
		atomic.AddInt32(&this.m_table.m_gc_n, -1)
		this.m_loaded = false
		released = true
	}
	return nil,released,state,update_string,args
}
func (this *dbServerRewardRow) Save(release bool) (err error, d bool, released bool) {
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
			log.Error("INSERT ServerRewards exec failed %v ", this.m_KeyId)
			return err, false, released
		}
		d = true
	} else if state == 2 {
		_, err = this.m_table.m_dbc.Exec(update_string, args...)
		if err != nil {
			log.Error("UPDATE ServerRewards exec failed %v", this.m_KeyId)
			return err, false, released
		}
		d = true
	}
	return nil, d, released
}
func (this *dbServerRewardRow) Touch(releasable bool) {
	this.m_touch = int32(time.Now().Unix())
	this.m_releasable = releasable
}
type dbServerRewardRowSort struct {
	rows []*dbServerRewardRow
}
func (this *dbServerRewardRowSort) Len() (length int) {
	return len(this.rows)
}
func (this *dbServerRewardRowSort) Less(i int, j int) (less bool) {
	return this.rows[i].m_touch < this.rows[j].m_touch
}
func (this *dbServerRewardRowSort) Swap(i int, j int) {
	temp := this.rows[i]
	this.rows[i] = this.rows[j]
	this.rows[j] = temp
}
type dbServerRewardTable struct{
	m_dbc *DBC
	m_lock *RWMutex
	m_rows map[int32]*dbServerRewardRow
	m_new_rows map[int32]*dbServerRewardRow
	m_removed_rows map[int32]*dbServerRewardRow
	m_gc_n int32
	m_gcing int32
	m_pool_size int32
	m_preload_select_stmt *sql.Stmt
	m_preload_max_id int32
	m_save_insert_stmt *sql.Stmt
	m_delete_stmt *sql.Stmt
}
func new_dbServerRewardTable(dbc *DBC) (this *dbServerRewardTable) {
	this = &dbServerRewardTable{}
	this.m_dbc = dbc
	this.m_lock = NewRWMutex()
	this.m_rows = make(map[int32]*dbServerRewardRow)
	this.m_new_rows = make(map[int32]*dbServerRewardRow)
	this.m_removed_rows = make(map[int32]*dbServerRewardRow)
	return this
}
func (this *dbServerRewardTable) check_create_table() (err error) {
	_, err = this.m_dbc.Exec("CREATE TABLE IF NOT EXISTS ServerRewards(KeyId int(11),PRIMARY KEY (KeyId))ENGINE=InnoDB ROW_FORMAT=DYNAMIC")
	if err != nil {
		log.Error("CREATE TABLE IF NOT EXISTS ServerRewards failed")
		return
	}
	rows, err := this.m_dbc.Query("SELECT COLUMN_NAME,ORDINAL_POSITION FROM information_schema.`COLUMNS` WHERE TABLE_SCHEMA=? AND TABLE_NAME='ServerRewards'", this.m_dbc.m_db_name)
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
	_, hasNextRewardId := columns["NextRewardId"]
	if !hasNextRewardId {
		_, err = this.m_dbc.Exec("ALTER TABLE ServerRewards ADD COLUMN NextRewardId int(11)")
		if err != nil {
			log.Error("ADD COLUMN NextRewardId failed")
			return
		}
	}
	_, hasRewardInfo := columns["RewardInfos"]
	if !hasRewardInfo {
		_, err = this.m_dbc.Exec("ALTER TABLE ServerRewards ADD COLUMN RewardInfos LONGBLOB")
		if err != nil {
			log.Error("ADD COLUMN RewardInfos failed")
			return
		}
	}
	return
}
func (this *dbServerRewardTable) prepare_preload_select_stmt() (err error) {
	this.m_preload_select_stmt,err=this.m_dbc.StmtPrepare("SELECT KeyId,NextRewardId,RewardInfos FROM ServerRewards")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbServerRewardTable) prepare_save_insert_stmt()(err error){
	this.m_save_insert_stmt,err=this.m_dbc.StmtPrepare("INSERT INTO ServerRewards (KeyId,NextRewardId,RewardInfos) VALUES (?,?,?)")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbServerRewardTable) prepare_delete_stmt() (err error) {
	this.m_delete_stmt,err=this.m_dbc.StmtPrepare("DELETE FROM ServerRewards WHERE KeyId=?")
	if err!=nil{
		log.Error("prepare failed")
		return
	}
	return
}
func (this *dbServerRewardTable) Init() (err error) {
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
func (this *dbServerRewardTable) Preload() (err error) {
	r, err := this.m_dbc.StmtQuery(this.m_preload_select_stmt)
	if err != nil {
		log.Error("SELECT")
		return
	}
	var KeyId int32
	var dNextRewardId int32
	var dRewardInfos []byte
		this.m_preload_max_id = 0
	for r.Next() {
		err = r.Scan(&KeyId,&dNextRewardId,&dRewardInfos)
		if err != nil {
			log.Error("Scan")
			return
		}
		if KeyId>this.m_preload_max_id{
			this.m_preload_max_id =KeyId
		}
		row := new_dbServerRewardRow(this,KeyId)
		row.m_NextRewardId=dNextRewardId
		err = row.RewardInfos.load(dRewardInfos)
		if err != nil {
			log.Error("RewardInfos %v", KeyId)
			return
		}
		row.m_NextRewardId_changed=false
		row.m_valid = true
		this.m_rows[KeyId]=row
	}
	return
}
func (this *dbServerRewardTable) GetPreloadedMaxId() (max_id int32) {
	return this.m_preload_max_id
}
func (this *dbServerRewardTable) fetch_rows(rows map[int32]*dbServerRewardRow) (r map[int32]*dbServerRewardRow) {
	this.m_lock.UnSafeLock("dbServerRewardTable.fetch_rows")
	defer this.m_lock.UnSafeUnlock()
	r = make(map[int32]*dbServerRewardRow)
	for i, v := range rows {
		r[i] = v
	}
	return r
}
func (this *dbServerRewardTable) fetch_new_rows() (new_rows map[int32]*dbServerRewardRow) {
	this.m_lock.UnSafeLock("dbServerRewardTable.fetch_new_rows")
	defer this.m_lock.UnSafeUnlock()
	new_rows = make(map[int32]*dbServerRewardRow)
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
func (this *dbServerRewardTable) save_rows(rows map[int32]*dbServerRewardRow, quick bool) {
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
func (this *dbServerRewardTable) Save(quick bool) (err error){
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
	this.m_removed_rows = make(map[int32]*dbServerRewardRow)
	rows := this.fetch_rows(this.m_rows)
	this.save_rows(rows, quick)
	new_rows := this.fetch_new_rows()
	this.save_rows(new_rows, quick)
	return
}
func (this *dbServerRewardTable) AddRow(KeyId int32) (row *dbServerRewardRow) {
	this.m_lock.UnSafeLock("dbServerRewardTable.AddRow")
	defer this.m_lock.UnSafeUnlock()
	row = new_dbServerRewardRow(this,KeyId)
	row.m_new = true
	row.m_loaded = true
	row.m_valid = true
	_, has := this.m_new_rows[KeyId]
	if has{
		log.Error("已经存在 %v", KeyId)
		return nil
	}
	this.m_new_rows[KeyId] = row
	atomic.AddInt32(&this.m_gc_n,1)
	return row
}
func (this *dbServerRewardTable) RemoveRow(KeyId int32) {
	this.m_lock.UnSafeLock("dbServerRewardTable.RemoveRow")
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
func (this *dbServerRewardTable) GetRow(KeyId int32) (row *dbServerRewardRow) {
	this.m_lock.UnSafeRLock("dbServerRewardTable.GetRow")
	defer this.m_lock.UnSafeRUnlock()
	row = this.m_rows[KeyId]
	if row == nil {
		row = this.m_new_rows[KeyId]
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
	StageScoreRanks *dbStageScoreRankTable
	ForbidTalks *dbForbidTalkTable
	ForbidLogins *dbForbidLoginTable
	ServerRewards *dbServerRewardTable
}
func (this *DBC)init_tables()(err error){
	this.StageScoreRanks = new_dbStageScoreRankTable(this)
	err = this.StageScoreRanks.Init()
	if err != nil {
		log.Error("init StageScoreRanks table failed")
		return
	}
	this.ForbidTalks = new_dbForbidTalkTable(this)
	err = this.ForbidTalks.Init()
	if err != nil {
		log.Error("init ForbidTalks table failed")
		return
	}
	this.ForbidLogins = new_dbForbidLoginTable(this)
	err = this.ForbidLogins.Init()
	if err != nil {
		log.Error("init ForbidLogins table failed")
		return
	}
	this.ServerRewards = new_dbServerRewardTable(this)
	err = this.ServerRewards.Init()
	if err != nil {
		log.Error("init ServerRewards table failed")
		return
	}
	return
}
func (this *DBC)Preload()(err error){
	err = this.StageScoreRanks.Preload()
	if err != nil {
		log.Error("preload StageScoreRanks table failed")
		return
	}else{
		log.Info("preload StageScoreRanks table succeed !")
	}
	err = this.ForbidTalks.Preload()
	if err != nil {
		log.Error("preload ForbidTalks table failed")
		return
	}else{
		log.Info("preload ForbidTalks table succeed !")
	}
	err = this.ForbidLogins.Preload()
	if err != nil {
		log.Error("preload ForbidLogins table failed")
		return
	}else{
		log.Info("preload ForbidLogins table succeed !")
	}
	err = this.ServerRewards.Preload()
	if err != nil {
		log.Error("preload ServerRewards table failed")
		return
	}else{
		log.Info("preload ServerRewards table succeed !")
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
	err = this.StageScoreRanks.Save(quick)
	if err != nil {
		log.Error("save StageScoreRanks table failed")
		return
	}
	err = this.ForbidTalks.Save(quick)
	if err != nil {
		log.Error("save ForbidTalks table failed")
		return
	}
	err = this.ForbidLogins.Save(quick)
	if err != nil {
		log.Error("save ForbidLogins table failed")
		return
	}
	err = this.ServerRewards.Save(quick)
	if err != nil {
		log.Error("save ServerRewards table failed")
		return
	}
	return
}
