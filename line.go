/**
 * @author zhagnxiaoping
 * @date  2024/6/15 12:03
 */
package navigator

import (
	"errors"
	"flag"
	"fmt"
	"github.com/Hellowlonewolf/navigator/leadutil"
	"math/rand"
	"os"
	"reflect"
	"runtime"
	"runtime/debug"
	"strings"
)

// Description:

// ILine Line 通用任务接口，我们实现的 locust 任务结构体应该实现这些接口。
type ILine interface {
	Weight() int
	OnStart() error
	OnStartInit() error
	OnFinish()
	OnError(operationName string, err error)
	Init()
	Next() *Task
	GetTask() []*Task
	SetTask(tasks []*Task)
}

// StatusLine line 状态维护
type StatusLine interface {
	// Status Line的运行状态，用于 Lead 检查 Line 状态
	// 0 正常，2 中断继续执行
	Status() int
	// ChangeStatus 修改状态
	ChangeStatus(status int)
}

// LeadLine 引火线接口
type Liner interface {
	ILine
	StatusLine
}

const (
	// 正常状态
	StatusNormal = 0
	// 中断继续执行
	StatusInterrupt = 2
	// StatusSkip 跳过该任务，包括间隔时间也不睡眠，直接下一个任务，随后自动设置回 StatusNormal
	StatusSkip = 3
)

type Line struct {
	// 虚拟用户权重
	// TODO: 后续实现多类型用户测试
	weight      int
	tasks       []*Task
	n           int
	status      int
	orderStatus bool
	HTTPClient  *leadutil.FastHTTPClient
}

func NewLine() *Line {
	line := &Line{}
	line.HTTPClient = leadutil.NewFastHTTPClient()
	return line
}

func (l *Line) SetTask(tasks []*Task) {
	l.tasks = tasks
}

func (l *Line) GetTask() []*Task {
	return l.tasks
}

// OnStart 任务初始化操作
// 应用于用户登录等操作
func (l *Line) OnStart() error {
	return nil
}

// 业务初始化前
func (l *Line) OnStartInit() error {
	return nil
}

// OnError错误处理函数
func (l *Line) OnError(operationName string, err error) {
	fmt.Println(operationName, " catch error:", err)
}

// OnError错误处理函数
func (l *Line) OnFinish() {
}

// Weight 权重
func (l *Line) Weight() int {
	return l.weight
}

// Init 虚拟用户初始化
func (l *Line) Init() {
	l.status = StatusNormal
	l.n = len(l.tasks)
	rand.Shuffle(
		l.n, func(i, j int) {
			l.tasks[i], l.tasks[j] = l.tasks[j], l.tasks[i]
		},
	)

	for _, t := range l.tasks {
		t.EffectiveWeight = t.Weight
	}

	if l.weight == 0 {
		l.weight = 1
	}
}

// Interrupt 中断该虚拟用户的执行
// l.Interrupt()
// return
func (l *Line) Interrupt() {
	l.status = StatusInterrupt
}

// Skip 跳过当前测试任务
// l.Skip()
// return
func (l *Line) Skip() {
	l.status = StatusSkip
}

// Status 用于 Lead 检查 Line 状态
func (l *Line) Status() int {
	return l.status
}

// ChangeStatus 修改 Line 状态
func (l *Line) ChangeStatus(status int) {
	l.status = status
}

// nextWeighted returns next selected weighted object.
func (l *Line) nextWeighted() *Task {
	if l.n == 0 {
		return nil
	}
	if l.n == 1 {
		return l.tasks[0]
	}

	return l.nextSmoothWeighted()
}

// 遍历所有任务权重,执行权重高任务,执行后该任务当前权重会被减去所有任务权重,以此计算任务执行概率
// todo (best *Task) 带指针 直接用指针.属性的方式，会修改原地址的值,所以原本的Task的任务权重会受自此数据变动
func (l *Line) nextSmoothWeighted() (best *Task) {
	total := 0
	if !l.orderStatus {
		// 先执行一次排序
		oldOrder := 9999
		orderIndex := 0
		for i := 0; i < l.n; i++ {
			if l.tasks[i].Order < oldOrder && l.tasks[i].Order != 0 {
				oldOrder = l.tasks[i].Order
				orderIndex = i
			}
		}
		if oldOrder != 9999 {
			best = l.tasks[orderIndex]
			// 重置
			l.tasks[orderIndex].Order = 0
			return best
		} else {
			// 排序执行后,后续不在进入排序执行
			l.orderStatus = true
		}
	}
	for i := 0; i < l.n; i++ {
		w := l.tasks[i]
		if w == nil {
			continue
		}
		w.CurrentWeight += w.EffectiveWeight
		total += w.EffectiveWeight
		// 这段if为无效代码块
		if w.EffectiveWeight < w.Weight {
			w.EffectiveWeight++
		}

		if best == nil || w.CurrentWeight > best.CurrentWeight {
			best = w
		}

	}

	if best == nil {
		return nil
	}
	// 修改原数据的当前任务权重,并返回执行该任务
	best.CurrentWeight -= total
	return best
}

// getWeightSum 获取任务权重
func (l *Line) getWeightSum() (weightSum int) {
	for _, task := range l.tasks {
		weightSum += task.Weight
	}
	return weightSum
}

// AddWeightFunc 添加任务函数，并设置权重
func (l *Line) AddWeightFunc(fn func(), weight int, order ...int) {
	l.AddTask(&Task{
		Weight: weight,
		Order:  l.getOrderFunc(order...),
		Fn:     l.safeFn(fn),
		Name:   runtime.FuncForPC(reflect.ValueOf(fn).Pointer()).Name(),
	})
}

// 获取任务顺序
func (l *Line) getOrderFunc(order ...int) int {
	if len(order) > 0 {
		return order[0]
	}
	return 0
}

func (l *Line) safeFn(fn func()) func() {
	return func() {
		defer func() {
			// don't panic
			err := recover()
			if err != nil {
				stackTrace := debug.Stack()
				errMsg := fmt.Sprintf("%v", err)
				os.Stderr.Write([]byte(errMsg))
				os.Stderr.Write([]byte("\n"))
				os.Stderr.Write(stackTrace)
				l.OnError("Run Panic", errors.New(errMsg+"\n"+string(stackTrace)))
			}
		}()

		fn()
	}
}

// AddTask 添加任务
func (l *Line) AddTask(task ...*Task) {
	l.tasks = append(l.tasks, task...)
}

// Next returns next selected task.
func (l *Line) Next() *Task {
	i := l.nextWeighted()
	if i == nil {
		return nil
	}
	return i
}

// SimpleLine 我们自定义的虚拟用户结构
type SimpleLine struct {
	// 匿名加载 Line 相关接口和功能
	*Line
	// 我们自定义这个用户可以放什么数据，方便各个任务函数执行时，设置上下文
	// 也可以使用 lead 的 Map 对象
	Data    leadutil.Map
	HostUrl string
}

var DebugMode = false
var hosturl = flag.String("host_url", "", "host url,example:http://localhost:1233/abc")
var debug_mode = flag.String("debug_mode", "false", "using debug mode,log debug message or othre.")

// CreateSimpleLine 我们自定义了一个虚拟用户对象，
// 需要提供这样一个创建函数，提供给 Lead进行启动。
// 在这里面，我们创建一个对象，该对象实现了lead.ILine接口
// 同时，注册一下需要执行的任务
func CreateSimpleLine() *SimpleLine {
	flag.Parse()

	if strings.ToUpper(*debug_mode) == "TRUE" {
		DebugMode = true
	}

	mt := &SimpleLine{
		Line:    NewLine(),
		Data:    *leadutil.NewMap(),
		HostUrl: *hosturl,
	}

	return mt
}

// wrapLine 对于不支持 LineStatus 的，进行处理
type wrapLine struct {
	ILine
}

// Status 用于 Lead 检查 Line 状态
func (l *wrapLine) Status() int {
	return StatusNormal
}

// ChangeStatus 修改 Line 状态
func (l *wrapLine) ChangeStatus(status int) {
}
