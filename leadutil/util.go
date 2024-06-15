/**
 * @author zhagnxiaoping
 * @date  2024/6/15 11:59
 */
package leadutil

import (
	"github.com/Hellowlonewolf/navigator/boomer"
	"time"
)

func GetElapsedMS(start time.Time) int64 {
	return time.Since(start).Nanoseconds() / int64(time.Millisecond)
}

// RecordSuccess 记录业务处理成功数据。
// requestType: 请求类型，默认就 http
// name: 业务名称
// responseTime: 业务处理时间
// responseLength: 业务处理数据大小
var RecordSuccess = boomer.RecordSuccess

// RecordFailure 记录业务处理失败数据。
// exception 失败原因
var RecordFailure = boomer.RecordFailure
