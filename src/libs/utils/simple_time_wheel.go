package utils

import (
	"libs/log"
	"sync/atomic"
	"time"
)

const (
	SIMPLE_TIMER_CHAN_LENGTH        = 4096
	DEFAULT_TIMER_INTERVAL_MSECONDS = 10
	DEFAULT_TIME_PERIOD_SECONDS     = 24 * 3600
)

type SimpleTimerFunc func(interface{}) int32

type SimpleTimerOpData struct {
	op         int32
	timer_id   int32
	timer_func SimpleTimerFunc
	param      interface{}
}

type SimpleTimer struct {
	timer_id    int32
	timer_func  SimpleTimerFunc
	param       interface{}
	next        *SimpleTimer
	prev        *SimpleTimer
	parent_list *SimpleTimerList
}

type SimpleTimerList struct {
	head *SimpleTimer
	tail *SimpleTimer
}

func (this *SimpleTimerList) add(timer_id int32, timer_func SimpleTimerFunc, param interface{}) *SimpleTimer {
	node := &SimpleTimer{
		timer_id:    timer_id,
		timer_func:  timer_func,
		param:       param,
		parent_list: this,
	}
	if this.head == nil {
		this.head = node
		this.tail = node
	} else {
		node.prev = this.tail
		this.tail.next = node
		this.tail = node
	}
	return node
}

func (this *SimpleTimerList) remove(timer *SimpleTimer) {
	if timer.prev != nil {
		timer.prev.next = timer.next
	}
	if timer.next != nil {
		timer.next.prev = timer.prev
	}
	if timer == this.head {
		this.head = timer.next
	}
	if timer == this.tail {
		this.tail = timer.prev
	}
}

type SimpleTimeWheel struct {
	timer_lists             []*SimpleTimerList
	curr_timer_index        int32
	last_check_time         int64
	timer_interval_mseconds int32
	op_chan                 chan *SimpleTimerOpData
	curr_timer_id           int32
	id2timer                map[int32]*SimpleTimer
}

func (this *SimpleTimeWheel) Init(timer_interval_mseconds, time_period_seconds int32) bool {
	if timer_interval_mseconds == 0 {
		timer_interval_mseconds = DEFAULT_TIMER_INTERVAL_MSECONDS
	}
	if time_period_seconds == 0 {
		time_period_seconds = DEFAULT_TIME_PERIOD_SECONDS
	}
	if (time_period_seconds*1000)%timer_interval_mseconds != 0 {
		return false
	}
	this.timer_lists = make([]*SimpleTimerList, time_period_seconds*1000/timer_interval_mseconds)
	this.curr_timer_index = -1
	this.timer_interval_mseconds = timer_interval_mseconds
	this.op_chan = make(chan *SimpleTimerOpData, SIMPLE_TIMER_CHAN_LENGTH)
	this.id2timer = make(map[int32]*SimpleTimer)
	return true
}

func (this *SimpleTimeWheel) Insert(timer_func SimpleTimerFunc, param interface{}) int32 {
	new_timer_id := atomic.AddInt32(&this.curr_timer_id, 1)
	this.op_chan <- &SimpleTimerOpData{
		op:         1,
		timer_id:   new_timer_id,
		timer_func: timer_func,
		param:      param,
	}
	return new_timer_id
}

func (this *SimpleTimeWheel) Remove(timer_id int32) {
	this.op_chan <- &SimpleTimerOpData{
		op:       2,
		timer_id: timer_id,
	}
}

func (this *SimpleTimeWheel) insert(timer_id int32, timer_func SimpleTimerFunc, param interface{}) bool {
	lists_len := int32(len(this.timer_lists))
	insert_list_index := (this.curr_timer_index + int32(len(this.timer_lists))) % lists_len
	list := this.timer_lists[insert_list_index]
	if list == nil {
		list = &SimpleTimerList{}
		this.timer_lists[insert_list_index] = list
	}
	timer := list.add(timer_id, timer_func, param)
	tmp_timer := this.id2timer[timer_id]
	if tmp_timer != nil {
		this.remove(tmp_timer)
		log.Warn("SimpleTimeWheel already exists timer[%v], remove it", timer_id)
	}
	this.id2timer[timer_id] = timer
	//log.Debug("Player[%v] conn insert in index[%v] list", player_id, insert_list_index)
	return true
}

func (this *SimpleTimeWheel) remove(timer *SimpleTimer) bool {
	timer.parent_list.remove(timer)
	delete(this.id2timer, timer.timer_id)
	return true
}

func (this *SimpleTimeWheel) remove_by_id(timer_id int32) bool {
	timer := this.id2timer[timer_id]
	if timer == nil {
		return false
	}
	this.remove(timer)
	return true
}

func (this *SimpleTimeWheel) Run() {
	defer func() {
		if err := recover(); err != nil {
			log.Stack(err)
		}
	}()

	for {
		// 处理操作队列
		is_break := false
		for !is_break {
			select {
			case d, ok := <-this.op_chan:
				{
					if !ok {
						log.Error("conn timer wheel op chan receive invalid !!!!!")
						return
					}

					if d.op == 1 {
						this.insert(d.timer_id, d.timer_func, d.param)
					} else if d.op == 2 {
						this.remove_by_id(d.timer_id)
					}
				}
			default:
				{
					is_break = true
				}
			}
		}

		now_time := int64(time.Now().Unix()*1000 + time.Now().UnixNano()/1000000)
		if this.last_check_time == 0 {
			this.last_check_time = now_time
		}
		// 跟上一次相差毫秒数
		diff_msecs := int32(now_time - this.last_check_time)
		y := diff_msecs / this.timer_interval_mseconds
		if y > 0 {
			var idx int32
			lists_len := int32(len(this.timer_lists))
			if y >= lists_len {
				if this.curr_timer_index > 0 {
					idx = this.curr_timer_index - 1
				} else {
					idx = lists_len - 1
				}
			} else {
				idx = (this.curr_timer_index + y) % lists_len
			}

			i := (this.curr_timer_index + 1) % lists_len
			for {
				list := this.timer_lists[i]
				if list != nil {
					t := list.head
					for t != nil {
						// execute timer function
						t.timer_func(t.param)
						this.remove(t)
						t = t.next
					}
					this.timer_lists[i] = nil
				}
				if i == idx {
					break
				}
				i = (i + 1) % lists_len
			}
			this.curr_timer_index = idx
			this.last_check_time = now_time
		}

		time.Sleep(time.Millisecond * 2)
	}
}
