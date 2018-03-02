package utils

import (
	"libs/log"
	"time"
)

func CheckWeekTimeArrival(last_time_point int32, week_time_format string) bool {
	log.Debug("last shop refresh time point is %v, week time format is %v", last_time_point, week_time_format)
	last_time := time.Unix(int64(last_time_point), 0)
	now_time := time.Now()

	if now_time.Unix() <= last_time.Unix() {
		return false
	}

	if now_time.Unix()-last_time.Unix() >= 7*24*3600 {
		return true
	}

	tm, err := time.Parse("Monday 08:00:00", week_time_format)
	if err != nil {
		log.Error("parse shop refresh time format[%v] failed, err[%v]", week_time_format, err.Error())
		return false
	}

	hour := tm.Hour()
	minute := tm.Minute()
	second := tm.Second()
	nsecond := tm.Nanosecond()

	tmp := time.Date(now_time.Year(), now_time.Month(), now_time.Day(), hour, minute, second, nsecond, time.Local)

	last_refresh_time := int64(0)
	if now_time.Weekday() < tm.Weekday() {
		last_refresh_time = tmp.Unix() - int64((7+now_time.Weekday()-tm.Weekday())*24*3600)
	} else {
		last_refresh_time = tmp.Unix() - int64((now_time.Weekday()-tm.Weekday())*24*3600)
	}

	log.Debug("now_unix:%v last_refresh_time:%v last_save_time:%v", now_time.Unix(), last_refresh_time, last_time_point)
	if now_time.Unix() >= last_refresh_time && last_refresh_time > int64(last_time_point) {
		return true
	}

	return false
}

func CheckDayTimeArrival(last_time_point int32, day_time_format string) bool {
	last_time := time.Unix(int64(last_time_point), 0)
	now_time := time.Now()

	if now_time.Unix() <= last_time.Unix() {
		return false
	}

	tm, err := time.Parse("08:00:00", day_time_format)
	if err != nil {
		log.Error("parse shop refresh time format[%v] failed, err[%v]", day_time_format, err.Error())
		return false
	}
	hour := tm.Hour()
	minute := tm.Minute()
	second := tm.Second()
	nsecond := tm.Nanosecond()
	tmp := time.Date(now_time.Year(), now_time.Month(), now_time.Day(), hour, minute, second, nsecond, time.Local)

	last_refresh_time := tmp.Unix()
	if now_time.Unix() >= last_refresh_time && last_refresh_time > int64(last_time_point) {
		return true
	}

	return false
}

type DaysTimeChecker struct {
	start_time_tm time.Time
}

func (this *DaysTimeChecker) Init(start_time_layout, start_time_value string) bool {
	var loc *time.Location
	var err error
	loc, err = time.LoadLocation("Local")
	if err != nil {
		log.Error("!!!!!!! Load Location Local error[%v]", err.Error())
		return false
	}

	this.start_time_tm, err = time.ParseInLocation(start_time_layout, start_time_value, loc)
	if err != nil {
		log.Error("!!!!!!! Parse start time layout[%v] failed, err[%v]", start_time_layout, err.Error())
		return false
	}

	if this.start_time_tm.Unix() >= time.Now().Unix() {
		log.Error("!!!!!!! Now time is Early to start time")
		return false
	}

	return true
}

func (this *DaysTimeChecker) IsArrival(last_save int32, interval_days int32) bool {
	if interval_days <= 0 {
		return false
	}

	last_time := time.Unix(int64(last_save), 0)
	now_time := time.Now()

	if last_time.Unix() > now_time.Unix() {
		return false
	}

	// 今天的时间点，与配置相同
	st := this.start_time_tm
	tmp := time.Date(now_time.Year(), now_time.Month(), now_time.Day(), st.Hour(), st.Minute(), st.Second(), st.Nanosecond(), st.Location())
	if this.start_time_tm.Unix() >= tmp.Unix() {
		return false
	}

	diff_days := (tmp.Unix() - this.start_time_tm.Unix()) / (24 * 3600)
	y := int(diff_days) % int(interval_days)

	last_refresh_time := int64(0)
	if y == 0 && now_time.Unix() >= tmp.Unix() {
		// 上次的固定刷新时间就是今天
		last_refresh_time = tmp.Unix()
	} else {
		last_refresh_time = tmp.Unix() - int64(y*24*3600)
	}

	log.Debug("now_unix:%v last_refresh_time:%v last_save_time:%v", now_time.Unix(), last_refresh_time, last_save)
	if last_refresh_time > int64(last_save) {
		return true
	}

	return false
}

func (this *DaysTimeChecker) RemainSecondsToNextRefresh(last_save int32, interval_days int32) int32 {
	log.Debug("@@@@@@@ last_save[%v]  interval_days[%v]", last_save, interval_days)
	if last_save <= 0 || interval_days <= 0 {
		return -1
	}
	last_time := time.Unix(int64(last_save), 0)
	now_time := time.Now()
	if last_time.Unix() > now_time.Unix() {
		return -1
	}

	st := this.start_time_tm
	today_tm := time.Date(now_time.Year(), now_time.Month(), now_time.Day(), st.Hour(), st.Minute(), st.Second(), st.Nanosecond(), st.Location())
	if st.Unix() >= today_tm.Unix() {
		return -1
	}

	diff_days := (today_tm.Unix() - st.Unix()) / (24 * 3600)
	y := int(diff_days) % int(interval_days)
	log.Debug("!!!!!!! today_unix[%v], st_unix[%v], diff_days[%v], y[%v]", today_tm.Unix(), st.Unix(), diff_days, y)

	next_refresh_time := int64(0)
	if y == 0 && now_time.Unix() < today_tm.Unix() {
		next_refresh_time = today_tm.Unix()
	} else {
		next_refresh_time = today_tm.Unix() + int64((int(interval_days)-y)*24*3600)
	}

	log.Debug("now:%v  next_refresh_time:%v  last_save:%v", now_time.Unix(), next_refresh_time, last_save)
	return int32(next_refresh_time - now_time.Unix())
}

func GetRemainSeconds4NextRefresh(config_hour, config_minute, config_second int32, last_save_time int32) (next_refresh_remain_seconds int32 /*, today_will_refresh bool*/) {
	now_time := time.Now()
	if int32(now_time.Unix()) < last_save_time {
		return 0
	}
	today_refresh_time := time.Date(now_time.Year(), now_time.Month(), now_time.Day(), int(config_hour), int(config_minute), int(config_second), 0, time.Local)
	/*if int32(today_refresh_time.Unix()) > last_save_time {
		today_will_refresh = true
	}*/
	if now_time.Unix() < today_refresh_time.Unix() {
		if int32(today_refresh_time.Unix())-24*3600 > last_save_time {
			next_refresh_remain_seconds = 0
		} else {
			next_refresh_remain_seconds = int32(today_refresh_time.Unix() - now_time.Unix())
		}
	} else {
		if int32(today_refresh_time.Unix()) > last_save_time {
			next_refresh_remain_seconds = 0
		} else {
			next_refresh_remain_seconds = 24*3600 - int32(now_time.Unix()-today_refresh_time.Unix())
		}
	}
	return
}

func IsDayTimeRefresh(config_hour, config_minute, config_second int32, last_unix_time int32) bool {
	now_time := time.Now()
	if int32(now_time.Unix()) < last_unix_time {
		return false
	}

	today_refresh_time := time.Date(now_time.Year(), now_time.Month(), now_time.Day(), int(config_hour), int(config_minute), int(config_second), 0, time.Local)
	if int32(today_refresh_time.Unix()) < last_unix_time {
		return false
	}

	return true
}
