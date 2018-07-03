package utils

import (
	"libs/log"
	"time"
)

func CheckWeekTimeArrival(last_time_point int32, week_time_string string) bool {
	last_time := time.Unix(int64(last_time_point), 0)
	now_time := time.Now()

	if now_time.Unix() <= last_time.Unix() {
		return false
	}

	if now_time.Unix()-last_time.Unix() >= 7*24*3600 {
		return true
	}

	loc, err := time.LoadLocation("Local")
	if err != nil {
		log.Error("!!!!!!! Load Location Local error[%v]", err.Error())
		return false
	}

	tm, err := time.ParseInLocation("Monday 15:04:05", week_time_string, loc)
	if err != nil {
		log.Error("parse shop refresh time format[%v] failed, err[%v]", week_time_string, err.Error())
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

	if now_time.Unix() >= last_refresh_time && last_refresh_time > int64(last_time_point) {
		return true
	}

	return false
}

func CheckDayTimeArrival(last_time_point int32, day_time_string string) bool {
	last_time := time.Unix(int64(last_time_point), 0)
	now_time := time.Now()

	if now_time.Unix() <= last_time.Unix() {
		return false
	}

	loc, err := time.LoadLocation("Local")
	if err != nil {
		log.Error("!!!!!!! Load Location Local error[%v]", err.Error())
		return false
	}

	tm, err := time.ParseInLocation("15:04:05", day_time_string, loc)
	if err != nil {
		log.Error("parse shop refresh time format[%v] failed, err[%v]", day_time_string, err.Error())
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

func GetRemainSeconds2NextDayTime(last_time_point int32, day_time_string string) int32 {
	last_time := time.Unix(int64(last_time_point), 0)
	now_time := time.Now()

	if now_time.Unix() <= last_time.Unix() {
		return -1
	}

	loc, err := time.LoadLocation("Local")
	if err != nil {
		log.Error("!!!!!!! Load Location Local error[%v]", err.Error())
		return -1
	}

	tm, err := time.ParseInLocation("15:04:05", day_time_string, loc)
	if err != nil {
		log.Error("parse shop refresh time format[%v] failed, err[%v]", day_time_string, err.Error())
		return -1
	}

	today_tm := time.Date(now_time.Year(), now_time.Month(), now_time.Day(), tm.Hour(), tm.Minute(), tm.Second(), tm.Nanosecond(), tm.Location())
	if tm.Unix() >= today_tm.Unix() {
		return -1
	}

	diff_days := (today_tm.Unix() - tm.Unix()) / (24 * 3600)
	y := int(diff_days) % int(1)

	next_refresh_time := int64(0)
	if y == 0 && now_time.Unix() < today_tm.Unix() {
		next_refresh_time = today_tm.Unix()
	} else {
		next_refresh_time = today_tm.Unix() + int64((int(1)-y)*24*3600)
	}
	return int32(next_refresh_time - now_time.Unix())
}

func GetRemainSeconds2NextSeveralDaysTime(last_save int32, day_time_string string, interval_days int32) int32 {
	if last_save <= 0 || interval_days <= 0 {
		return -1
	}
	last_time := time.Unix(int64(last_save), 0)
	now_time := time.Now()
	if last_time.Unix() > now_time.Unix() {
		return -1
	}

	loc, err := time.LoadLocation("Local")
	if err != nil {
		log.Error("!!!!!!! Load Location Local error[%v]", err.Error())
		return -1
	}

	tm, err := time.ParseInLocation("15:04:05", day_time_string, loc)
	if err != nil {
		log.Error("parse shop refresh time format[%v] failed, err[%v]", day_time_string, err.Error())
		return -1
	}

	today_tm := time.Date(now_time.Year(), now_time.Month(), now_time.Day(), tm.Hour(), tm.Minute(), tm.Second(), tm.Nanosecond(), tm.Location())
	if tm.Unix() >= today_tm.Unix() {
		return -1
	}

	diff_days := (today_tm.Unix() - tm.Unix()) / (24 * 3600)
	y := int(diff_days) % int(interval_days)

	next_refresh_time := int64(0)
	if y == 0 && now_time.Unix() < today_tm.Unix() {
		next_refresh_time = today_tm.Unix()
	} else {
		next_refresh_time = today_tm.Unix() + int64((int(interval_days)-y)*24*3600)
	}

	return int32(next_refresh_time - now_time.Unix())
}

type DaysTimeChecker struct {
	time_tm time.Time
}

func (this *DaysTimeChecker) Set(time_layout, time_value string) bool {
	var loc *time.Location
	var err error
	loc, err = time.LoadLocation("Local")
	if err != nil {
		log.Error("!!!!!!! Load Location Local error[%v]", err.Error())
		return false
	}

	this.time_tm, err = time.ParseInLocation(time_layout, time_value, loc)
	if err != nil {
		log.Error("!!!!!!! Parse start time layout[%v] failed, err[%v]", time_layout, err.Error())
		return false
	}

	if this.time_tm.Unix() >= time.Now().Unix() {
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
	st := this.time_tm
	tmp := time.Date(now_time.Year(), now_time.Month(), now_time.Day(), st.Hour(), st.Minute(), st.Second(), st.Nanosecond(), st.Location())
	if this.time_tm.Unix() >= tmp.Unix() {
		return false
	}

	diff_days := (tmp.Unix() - this.time_tm.Unix()) / (24 * 3600)
	y := int(diff_days) % int(interval_days)

	last_refresh_time := int64(0)
	if y == 0 && now_time.Unix() >= tmp.Unix() {
		// 上次的固定刷新时间就是今天
		last_refresh_time = tmp.Unix()
	} else {
		last_refresh_time = tmp.Unix() - int64(y*24*3600)
	}

	if last_refresh_time > int64(last_save) {
		return true
	}

	return false
}

func (this *DaysTimeChecker) RemainSecondsToNextRefresh(last_save int32, interval_days int32) int32 {
	if last_save <= 0 || interval_days <= 0 {
		return -1
	}
	last_time := time.Unix(int64(last_save), 0)
	now_time := time.Now()
	if last_time.Unix() > now_time.Unix() {
		return -1
	}

	st := this.time_tm
	today_tm := time.Date(now_time.Year(), now_time.Month(), now_time.Day(), st.Hour(), st.Minute(), st.Second(), st.Nanosecond(), st.Location())
	if st.Unix() >= today_tm.Unix() {
		return -1
	}

	diff_days := (today_tm.Unix() - st.Unix()) / (24 * 3600)
	y := int(diff_days) % int(interval_days)

	next_refresh_time := int64(0)
	if y == 0 && now_time.Unix() < today_tm.Unix() {
		next_refresh_time = today_tm.Unix()
	} else {
		next_refresh_time = today_tm.Unix() + int64((int(interval_days)-y)*24*3600)
	}

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
