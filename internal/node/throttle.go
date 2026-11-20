package node

import (
	"fmt"
	"time"

	"github.com/Autumn-27/ScopeSentry-Scan/internal/config"
	"github.com/Autumn-27/ScopeSentry-Scan/internal/pool"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
)

// 容器内存自适应限流：节点被 cgroup 限住时会被压死（心跳停、任务卡住、界面显示离线）。
// 与其等它死，不如自己收手：内存持续高位就把任务并发降到 1，回落后再恢复。
// 2026-11-20 实测：dev-02 容器 1.26GB 上限、两台并发任务能把内存顶到 99.9%。
const (
	memHighWater   = 90.0 // 超过就准备降压
	memLowWater    = 70.0 // 低于就恢复
	memSamplesHigh = 3    // 连续 3 次（心跳 20s，约 60s）才动手，避免抖动
	memSamplesLow  = 3
)

var (
	memHighStreak int
	memLowStreak  int
	throttled     bool
)

// decide 把「采样到多少内存 → 要不要改并发」抽成纯函数，便于单测。
// 返回 (新并发, 是否已throttle, 是否发生了变化)。
func decide(mem float64, base int, throttledIn bool, highStreak, lowStreak int) (bool, int, int, bool) {
	if mem >= memHighWater {
		highStreak++
		lowStreak = 0
	} else if mem <= memLowWater {
		lowStreak++
		highStreak = 0
	} else {
		highStreak = 0
		lowStreak = 0
	}
	changed := false
	nowThrottled := throttledIn
	if !nowThrottled && highStreak >= memSamplesHigh && base > 1 {
		nowThrottled = true
		changed = true
	}
	if nowThrottled && lowStreak >= memSamplesLow {
		nowThrottled = false
		changed = true
	}
	return nowThrottled, highStreak, lowStreak, changed
}

// ObserveMemoryPressure 在每次心跳时调用，返回「当前生效」的任务并发数，
// 由调用方写进 maxTaskNum 上报——界面因此能看到真实的并发上限。
func ObserveMemoryPressure() int {
	base := config.ModulesConfig.MaxGoroutineCount
	if base <= 0 {
		base = 1
	}
	mem := ContainerMemoryPercent()
	if mem < 0 {
		return base // 非容器环境，不做自适应
	}

	nowThrottled, hs, ls, changed := decide(mem, base, throttled, memHighStreak, memLowStreak)
	memHighStreak, memLowStreak = hs, ls
	if changed {
		next := base
		if nowThrottled {
			next = 1
		}
		if err := pool.PoolManage.SetGoroutineCount("task", next); err == nil {
			throttled = nowThrottled
			logger.SlogInfoLocal(fmt.Sprintf("container memory %.1f%%: task concurrency -> %d (base %d, throttled=%v)", mem, next, base, nowThrottled))
		}
	}

	if throttled {
		return 1
	}
	return base
}

var _ = time.Second
