/**
 * @author zhagnxiaoping
 * @date  2024/6/15 12:05
 */
package navigator

import (
	"github.com/Hellowlonewolf/navigator/boomer"
	"time"
)

type option struct {
	// 任务执行间隔时间设置
	interval time.Duration
	// 动态间隔时间
	// 如果第一个任务执行完成时间小于设置的 interval 时间，则仅睡眠剩余时间
	// |--------------- interval time -------------------------|
	// |--- task running time ---|--- dynamic interval time ---|
	enableDynamicInterval bool

	// taskCycle 任务执行周期
	taskCycle int64

	boomerClient *boomer.Boomer

	// retryInCriticalInterval 虚拟用户执行出错时，重试的等待时间，默认2s
	retryInCriticalInterval time.Duration
}

func defaultOpt() *option {
	opt := &option{
		retryInCriticalInterval: time.Second * 2,
	}

	return opt
}

type Option func(opt *option)

// Interval 任务执行间隔时间，
// 参数格式为语义化时间，示例：1ms。其他示例：1s,1min
func Interval(interval string) Option {
	i, err := time.ParseDuration(interval)
	if err != nil {
		i = 0
	}

	return func(opt *option) {
		opt.interval = i
	}
}

// RetryInCriticalInterval 虚拟用户执行出错时，重试的等待时间，默认2s，
// 参数格式为语义化时间，示例：1ms。其他示例：1s,1min
func RetryInCriticalInterval(interval string) Option {
	i, err := time.ParseDuration(interval)
	if err != nil {
		i = 0
	}

	return func(opt *option) {
		opt.retryInCriticalInterval = i
	}
}

// EnableDynamicInterval 动态间隔时间，
// 如果第一个任务执行完成时间小于设置的 interval 时间，则仅睡眠剩余时间
// |--------------- interval time -------------------------|
// |--- task running time ---|--- dynamic interval time ---|
func EnableDynamicInterval() Option {
	return func(opt *option) {
		opt.enableDynamicInterval = true
	}
}

// TaskCycle 设置任务执行周期次数，每次执行 Task 函数则计数一次，达到次数后停止执行 task 。
// onstart不计算
func TaskCycle(times int) Option {
	return func(opt *option) {
		opt.taskCycle = int64(times)
	}
}

// BoomerClient custom boomer
func BoomerClient(boomerClient *boomer.Boomer) Option {
	return func(opt *option) {
		opt.boomerClient = boomerClient
	}
}
