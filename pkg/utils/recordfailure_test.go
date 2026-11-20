package utils

import (
	"testing"
	"time"
)

// 递归 bug 回归测试：recordFailure 曾自己调用自己，一次工具失败就栈溢出崩进程。
// 这个测试在修复前会 panic（stack overflow），修复后正常返回并把时间点记进窗口。
func TestRecordFailureDoesNotRecurse(t *testing.T) {
	before := (&UtilTools{}).EventFailureRecent(10 * time.Minute)
	recordFailure()
	after := (&UtilTools{}).EventFailureRecent(10 * time.Minute)
	if after != before+1 {
		t.Fatalf("recent failures = %d, want %d (recordFailure 没有记录时间点)", after, before+1)
	}
}
