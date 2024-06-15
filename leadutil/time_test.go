/**
 * @author zhagnxiaoping
 * @date  2024/6/15 12:02
 */
package leadutil

import (
	"testing"
	"time"
)

func TestWaitToNextClockTime(t *testing.T) {
	now := time.Now()
	minute := now.Minute()
	// 相当于按一个钟分区，则测试等待到当前时间的下一分钟
	WaitToNextClockTime(minute+1, time.Minute, 60)

	now2 := time.Now()

	if now2.Sub(now).Minutes() > 1 {
		t.Errorf("unexpected WaitToNextClockTime: old:%v,new:%v", now, now2)
	}
}

func TestWaitToNextTime(t *testing.T) {
	now := time.Now()
	WaitToNextTime(time.Second * 5)

	now2 := time.Now()

	if now2.Sub(now).Seconds() < 3 {
		t.Errorf("unexpected WaitToNextTime: old:%v,new:%v", now, now2)
	}
}
