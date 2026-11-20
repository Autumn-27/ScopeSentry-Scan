package node

import (
	"fmt"
	"os"
	"runtime/debug"
	"strconv"
	"strings"

	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
)

// ApplyContainerMemoryLimit 让 Go 运行时知道 cgroup 的硬上限。
//
// 默认 Go 只看物理内存、不看 cgroup：容器被限到 1-2GB 时它仍按自己的节奏堆内存，
// 于是 RSS 顶到 cgroup 上限 → 频繁 GC/被限流，表现为「内存长期 99%、心跳变慢、被判离线」
// （2026-11-21 实测三台长期 74-99%）。设一个略低于上限的软目标后，运行时会在接近时
// 主动加大 GC 力度，RSS 稳定在上限之下。
func ApplyContainerMemoryLimit() {
	limit := containerMemoryBytes()
	if limit <= 0 {
		return
	}
	// 80%：留给运行时栈、CGO/子进程（nuclei 等外部工具）和其它开销
	target := int64(float64(limit) * 0.8)
	if target < 256<<20 {
		return
	}
	debug.SetMemoryLimit(target)
	logger.SlogInfoLocal(fmt.Sprintf("memory limit: container=%dMiB soft-limit=%dMiB", limit>>20, target>>20))
}

// SoftLimitMiB 返回当前生效的 Go 软内存目标（MiB），供心跳上报——
// 排查内存问题时「软上限设成了多少」是关键参数（2026-11-21）。
func SoftLimitMiB() int64 {
	limit := containerMemoryBytes()
	if limit <= 0 {
		return 0
	}
	return int64(float64(limit)*0.8) >> 20
}

func containerMemoryBytes() int64 {
	if b, err := os.ReadFile("/sys/fs/cgroup/memory.max"); err == nil {
		s := strings.TrimSpace(string(b))
		if s != "" && s != "max" {
			if v, perr := strconv.ParseInt(s, 10, 64); perr == nil {
				return v
			}
		}
	}
	if b, err := os.ReadFile("/sys/fs/cgroup/memory/memory.limit_in_bytes"); err == nil {
		if v, perr := strconv.ParseInt(strings.TrimSpace(string(b)), 10, 64); perr == nil && v > 0 && v < 1<<60 {
			return v
		}
	}
	return 0
}
