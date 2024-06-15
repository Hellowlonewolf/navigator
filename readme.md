# navigator引火线
Boomer的特殊启动器。
用于模拟原版Locust 的Task对象。

Task对象实例对应了一个虚拟用户，有自己的数据。



## 安装

`go get https://github.com/Hellowlonewolf/navigatorr`

Option

- Interval 任务执行间隔时间， 参数格式为语义化时间，示例：1ms。其他示例：1s,1min。

- EnableDynamicInterval 动态间隔时间。

```go
// 如果第一个任务执行完成时间小于设置的 interval 时间，则仅睡眠剩余时间
// |--------------- interval time -------------------------|
// |--- task running time ---|--- dynamic interval time ---|
```

- TaskCycle 设置任务执行周期次数，每次执行 Task 函数则计数一次，达到次数后停止执行 task 。 OnStart等不计算。

## 使用

1. 构建测试任务逻辑

需要构建一个`Task`结构体，需要实现`navigator.ILine`接口，无其他需求的情况下，可以直接匿名组合`*navigator.Line`。提供了默认实现。
然后在此结构体上实现各类测试方法以及重写自定义`navigator.ILine`接口。

实例：
```go
// MyTask 任务对象，其实例化后，每个实例对应一个虚拟用户，对象内数据独立维护，不会串。
type MyTask struct {
	// 匿名加载 Line 相关接口和功能
	*navigator.Line
	// 我们自定义这个用户可以放什么数据，方便各个任务函数执行时，设置上下文
	// 也可以使用 navigator 的 Map 对象
	Data map[string]string
	// 同上
	helloCount int
	// 同上
	wordCount int
}


// OnStart 复写OnStart逻辑，每个虚拟用户启动后，都会执行一次
func (m *MyTask) OnStart() error {
	time.Sleep(time.Second)
	// 记录操作状态，用于locust统计接口数据
	leadutil.RecordSuccess("http", "OnStart", 111, int64(10))
	return nil
}

// OnFinish 复写 OnFinish 逻辑，在虚拟用户关闭时，执行操作。
func (m *MyTask) OnFinish() {

}

// SayHello 测试业务逻辑
func (m *MyTask) SayHello() {
	startTime := time.Now()
	m.helloCount++
	
	if m.helloCount >= 3{
		// 中断用户执行操作
		m.Interrupt()
	}
	
	// 记录请求成功，用于locust统计接口数据
	// TODO: 一般都是封装到对应的 HTTP 请求中或其他流程，不用再手动调用
	leadutil.RecordSuccess("http", "Hello", leadutil.GetElapsedMS(startTime), int64(10))
}
```

2. 写好`Task`后，需要提供一个构造方法。
   构造方法用于`navigator`在收到locust分配的用户数量的时候，进行创建`Task`实例。

同时，这个构造方法也是设置各测试逻辑执行权重的地方。

```go

// CreateMyTask 我们自定义了一个虚拟用户对象，
// 需要提供这样一个创建函数，提供给 navigator进行启动。
// 在这里面，我们创建一个对象，该对象实现了navigator.ILine接口
// 同时，注册一下需要执行的任务
func CreateMyTask() navigator.ILine {
	mt := &MyTask{
		Line: navigator.NewLine(),
		Data: map[string]string{},
	}

	// TODO: 顺序执行任务
	// 添加要处理的方法和权重
	// mt.AddWeightFunc(mt.SayHello, 2)
	// mt.AddWeightFunc(mt.WriteWord, 3)
	mt.AddWeightFunc(mt.Sleep500, 3)

	return mt
}
```

3. 整理完`Task`后，再提供一个启动逻辑，就完事了。

目前`navigator`的`Run`仅支持一个`Task`构造方法的参数。

```go
func main() {
	l := navigator.New(navigator.Interval("1s"),navigator.EnableDynamicInterval(),navigator.TaskCycle(3))
	l.Run(gotest.CreateMyTask)
}

```

## FAQ

### zmq版本问题
遇到如下:
```shell
 Version 3.0 received does match expected version 3.1
```

表示当前 master 使用的zmq是3.0版本，但是压测客户端使用的是3.1。
需要调整下述模块版本(go.mod)
```shell
github.com/zeromq/goczmq v4.1.0+incompatible // indirect
github.com/zeromq/gomq v0.0.0-20181008000130-95dc37dee5c4 // indirect
```

如果是下面这种提示：
```shell
 Version 3.1 received does match expected version 3.0
```

表示当前 master 使用的zmq版本是3.1，但是压测客户端用的是3.0的。
需要调整下述模块版本(go.mod)

```shell
github.com/zeromq/gomq v0.0.0-20201031135124-cef4e507bb8e
```



