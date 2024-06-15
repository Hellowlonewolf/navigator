/**
 * @author zhagnxiaoping
 * @date  2024/6/15 12:05
 */
package navigator

// Task 任务结构
type Task struct {
	// The weight is used to distribute goroutines over multiple tasks.
	Weight int
	// 1-9999  Execution from small to large
	Order int
	// Fn is called by the goroutines allocated to this task, in a loop.
	Fn   func()
	Name string

	SmoothWeight
}

type SmoothWeight struct {
	CurrentWeight   int
	EffectiveWeight int
}
