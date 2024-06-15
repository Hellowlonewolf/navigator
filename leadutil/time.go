/**
 * @author zhagnxiaoping
 * @date  2024/6/15 11:58
 */
package leadutil

import (
	"sync"
	"time"
)

// Creator: ChenJingyi
// CreateDate: 2021-03-11
// Description:

// WaitToNextTime 等待到后面某一秒时刻整，支持 秒级 和分钟级, 秒级操作超过 60 秒，将按分钟级来，自动抹零。
// 比如 WaitToNextTime(time.Second * 3) ，当前时间为 20:20:20.333 等待到 20:20:03.000，
// 比如 WaitToNextTime(time.Minute * 3) ，当前时间为 20:20:20.333 等待到 20:23:00.000，
func WaitToNextTime(duration time.Duration) {
	next := GetWaitToNextTime(duration)
	SleepToNext(next)
}

// GetWaitToNextTime 获取下一个时间整点，支持秒级和分钟级, 秒级操作超过 60 秒，将按分钟级来，自动抹零。
// 比如 GetWaitToNextTime(time.Second * 3) ，当前时间为 20:20:20.333 则下一个时间为 20:20:03.000，
// 比如 GetWaitToNextTime(time.Minute * 3) ，当前时间为 20:20:20.333 则下一个时间为 20:23:00.000，
func GetWaitToNextTime(duration time.Duration) time.Time {
	now := time.Now()
	// 计算下一个整点
	next := now.Add(duration)

	sec := next.Second()
	min := next.Minute()

	if duration >= time.Minute {
		sec = 0
	}

	next = time.Date(next.Year(), next.Month(), next.Day(), next.Hour(), min, sec, 0, next.Location())

	return next
}

// WaitToNextClockTime 睡眠等待到下一个时钟时刻，支持秒级和分钟级, 秒级操作超过 60 秒，将按分钟级来，自动抹零。
// numOfClock 表示当前时刻后第 next 时刻的时间
// timeLevel 时间级别，秒级: time.Second,分钟级: time.Minute
// partitions: 时钟分区，分钟级和秒级按标准值60开始分区，如果为12，表示每个分区为 5 秒或 5 分钟。
// 示例 numOfClock:2, timeLevel:time.Minute,partitions:12 ，当前时间为20:20:20.333，则时间计算出来为20:24:00.00
// 60/12=5，每个时刻分区数值为5，5*2=10 下个时刻数值，表示需要的时间为当前时刻的第10个刻度，分钟级的话就是第10分，秒级就是第10秒
// 一般用来特殊的等待时间逻辑中使用。
func WaitToNextClockTime(numOfClock int, timeLevel time.Duration, partitions int) {
	SleepToNext(GetWaitToNextClockTime(numOfClock, timeLevel, partitions))
}

// GetWaitToNextClockTime 计算下一个时钟时刻，支持秒级和分钟级, 秒级操作超过 60 秒，将按分钟级来，自动抹零。
// numOfClock 表示当前时刻后第 next 时刻的时间
// timeLevel 时间级别，秒级: time.Second,分钟级: time.Minute
// partitions: 时钟分区，分钟级和秒级按标准值60开始分区，如果为5，表示每个分区为 12秒或12分钟。
// 示例 numOfClock:2, timeLevel:time.Minute,partitions:5 ，当前时间为20:20:20.333，则时间计算出来为20:24:00.00
// 60/5*2=24，表示需要的时间为当前时刻的第24个点，分钟级的话就是第24分，秒级就是第24秒
// 一般用来特殊的等待时间逻辑中使用。
func GetWaitToNextClockTime(numOfClock int, timeLevel time.Duration, partitions int) time.Time {
	now := time.Now()
	// 计算下一个时刻

	// 计算每个时刻分区的时间跨度
	partitionTimeSize := 60 / partitions

	nextTimeClock := numOfClock * partitionTimeSize

	sec := now.Second()
	minute := now.Minute()

	if timeLevel == time.Second {
		sec = nextTimeClock
	}

	if timeLevel == time.Minute {
		minute = nextTimeClock
		sec = 0
	}

	return time.Date(now.Year(), now.Month(), now.Day(), now.Hour(), minute, sec, 0, now.Location())
}

// SleepToNext 睡眠至某一时刻
func SleepToNext(next time.Time) {
	now := time.Now()
	t1 := GetTimer(next.Sub(now))
	<-t1.C
	defer PutTimer(t1)
}

// GetTimer returns a timer for the given duration d from the pool.
// Return back the timer to the pool with Put.
func GetTimer(d time.Duration) *time.Timer {
	if v := timerPool.Get(); v != nil {
		t := v.(*time.Timer)
		if t.Reset(d) {
			// active timer?
			return time.NewTimer(d)
		}
		return t
	}
	return time.NewTimer(d)
}

// PutTimer returns t to the pool.
// t cannot be accessed after returning to the pool.
func PutTimer(t *time.Timer) {
	if !t.Stop() {
		// Drain t.C if it wasn't obtained by the caller yet.
		select {
		case <-t.C:
		default:
		}
	}
	timerPool.Put(t)
}

var timerPool sync.Pool
