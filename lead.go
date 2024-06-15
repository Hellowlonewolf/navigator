/**
 * @author zhagnxiaoping
 * @date  2024/6/15 12:04
 */
package navigator

import (
	"errors"
	"flag"
	"fmt"
	"github.com/Hellowlonewolf/navigator/boomer"
	"github.com/Hellowlonewolf/navigator/leadutil"
	"log"
	"math/rand"
	"os"
	"runtime/debug"
	"sync"
	"time"
)

var (
	IntervalFlag = flag.String("interval", "1s", "set wait time when run each user task method,using 1s,100ms....")
)

func init() {
	// 初始化随机数种子
	rand.Seed(time.Now().Unix())
}

// Lead 实现重新处理 boomer 任务逻辑，在boomer的基础上实现
// 按对象及函数权重执行任务。
type Lead struct {
	lines  []func() ILine
	option *option
}

// New 创建 Lead 压测任务对象
func New(opts ...Option) *Lead {
	l := &Lead{}
	l.option = defaultOpt()
	l.SetInterval(*IntervalFlag)
	for _, opt := range opts {
		opt(l.option)
	}

	return l
}

func (l *Lead) reset() {
}

// SetInterval 设置函数执行时间,格式为 1ms,1s,1min
func (l *Lead) SetInterval(interval string) {
	i, err := time.ParseDuration(interval)
	if err != nil {
		return
	}

	l.option.interval = i
}

// Run 运行测试任务
func (l *Lead) Run(ls ...func() ILine) {

	l.lines = ls
	newUser := l.lines[0]()
	err := newUser.OnStartInit()
	if err != nil {
		newUser.OnError("OnStartActivity", err)
	}
	// 检查错误
	t := &boomer.Task{
		Weight: 100,
		Fn:     l.forFn,
		Name:   "lead",
	}

	if l.option.boomerClient != nil {
		leadutil.RecordFailure = l.option.boomerClient.RecordFailure
		leadutil.RecordSuccess = l.option.boomerClient.RecordSuccess
		l.option.boomerClient.Run(t)
	}

	boomer.Run(t)
}

// ResetLine 设置压测任务虚拟用户创建函数。
// 注意，不要在压测进行时调用，避免出现异常。
func (l *Lead) ResetLine(ls ...func() ILine) {
	l.lines = ls
}

func (l *Lead) forFn() {
	var user Liner
	var status int
	quitChan := make(chan bool)

	defer l.reset()
	defer func() {
		// don't panic
		err := recover()
		if err != nil {
			stackTrace := debug.Stack()
			errMsg := fmt.Sprintf("%v", err)
			os.Stderr.Write([]byte(errMsg))
			os.Stderr.Write([]byte("\n"))
			os.Stderr.Write(stackTrace)
			if user != nil {
				user.OnError("Run Panic", errors.New(errMsg+"\n"+string(stackTrace)))
			}
		}
		// TODO: 优化流程
		if user != nil {
			user.OnFinish()
		}

		if status == StatusInterrupt {
			<-quitChan
		}

		time.Sleep(l.option.retryInCriticalInterval)

	}()

	// 收到locust 的停止指令后，停止任务
	closeChan := &sync.Once{}

	err := Events.SubscribeOnce(EventStop, func() {
		closeChan.Do(func() {
			close(quitChan)
		})
	})
	if err != nil {
		user.OnError("SubscribeOnce EventStop", err)
	}

	// 检查程序退出
	// 启动循环，根据权重创建虚拟用户以及执行初始化操作
	newUser := l.lines[0]()
	var ok bool
	if user, ok = newUser.(Liner); !ok {
		user = &wrapLine{newUser}
	}

	user.Init()
	err = user.OnStart()
	if err != nil {
		user.OnError("OnStartError", err)
		go func() {
			boomer.WorkerClose <- 1
		}()
		return

	}

	var interval time.Duration
	var nextTime time.Time
	var taskCycle int64

	startTime := time.Now()

	for {
		interval = l.option.interval
		startTime = time.Now()
		if l.option.enableDynamicInterval {
			nextTime = startTime.Add(interval)
		}

		select {
		case <-quitChan:
			return
		default:
			if nextTask := user.Next(); nextTask != nil {
				nextTask.Fn()
				taskCycle++

				if l.option.taskCycle > 0 {
					if taskCycle >= l.option.taskCycle {
						// TODO: 或许需要一些提示

						// 暂时用Log输出
						log.Printf("task interrupt:taskCycle:%d,status:%d\n", taskCycle, user.Status())
						status = StatusInterrupt
						return
					}
				}

				switch user.Status() {
				case StatusInterrupt:
					log.Printf("task interrupt:taskCycle:%d,status:%d\n", taskCycle, user.Status())
					status = StatusInterrupt

					return
				case StatusSkip:
					user.ChangeStatus(StatusNormal)
					continue
				}

			} else {
				user.OnError("NextTask", errors.New("next task not found"))
			}
		}

		if l.option.enableDynamicInterval {
			interval = nextTime.Sub(time.Now())
		}

		if interval > 0 {
			time.Sleep(interval)
		}
	}
}
