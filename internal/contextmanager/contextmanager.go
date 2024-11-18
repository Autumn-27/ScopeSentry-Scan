// contextmanager-------------------------------------
// @file      : contextmanager.go
// @author    : Autumn
// @contact   : rainy-autumn@outlook.com
// @time      : 2024/11/7 21:29
// -------------------------------------------

package contextmanager

import (
	"context"
	"fmt"
	"github.com/Autumn-27/ScopeSentry-Scan/pkg/logger"
	"sync"
)

// ContextManager 管理多个上下文的结构体
type ContextManager struct {
	mu         sync.Mutex
	contexts   map[string]context.Context    // 存储上下文
	cancels    map[string]context.CancelFunc // 存储取消函数
	waitGroups map[string]*sync.WaitGroup    // 存储每个任务的 WaitGroup
}

// Global map to store all ContextManagers by their IDs
var GlobalContextManagers *ContextManager

// NewContextManager 创建一个新的上下文管理器
func NewContextManager() {
	GlobalContextManagers = &ContextManager{
		contexts:   make(map[string]context.Context),
		cancels:    make(map[string]context.CancelFunc),
		waitGroups: make(map[string]*sync.WaitGroup),
	}
}

// AddContext 创建并添加一个新的上下文
func (cm *ContextManager) AddContext(taskID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 如果 taskID 已经存在，跳过添加
	if _, exists := cm.contexts[taskID]; exists {
		// 这里可以选择直接返回，或者执行其他处理逻辑
		return
	}
	// 创建新的上下文及其取消函数
	ctx, cancel := context.WithCancel(context.Background())

	// 添加到管理器
	cm.contexts[taskID] = ctx
	cm.cancels[taskID] = cancel
	cm.waitGroups[taskID] = &sync.WaitGroup{}
}

// CancelContext 取消指定任务的上下文
func (cm *ContextManager) CancelContext(taskID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cancel, ok := cm.cancels[taskID]; ok {
		cancel()
		logger.SlogInfo(fmt.Sprintf("stop task success: %v", taskID))
	}
}

// CancelAllContexts 取消所有上下文
func (cm *ContextManager) CancelAllContexts() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	for _, cancel := range cm.cancels {
		cancel()
	}
}

// WaitForAll 等待所有上下文相关的任务完成
func (cm *ContextManager) WaitForAll() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	var wg sync.WaitGroup
	for _, wgItem := range cm.waitGroups {
		wg.Add(1)
		go func(wgItem *sync.WaitGroup) {
			defer wg.Done()
			wgItem.Wait()
		}(wgItem)
	}
	wg.Wait()
}

// DeleteContext 删除指定任务的上下文
func (cm *ContextManager) DeleteContext(taskID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 删除上下文、取消函数和 WaitGroup
	if _, ok := cm.contexts[taskID]; ok {
		delete(cm.contexts, taskID)
		delete(cm.cancels, taskID)
		delete(cm.waitGroups, taskID)
		logger.SlogInfoLocal(fmt.Sprintf("Context %s deleted\n", taskID))
	} else {
		logger.SlogInfoLocal(fmt.Sprintf("Context %s not found\n", taskID))
	}
}

// GetContext 获取指定任务的上下文
func (cm *ContextManager) GetContext(taskID string) context.Context {
	// 获取锁保护上下文读取操作
	cm.mu.Lock()
	ctx, exists := cm.contexts[taskID]
	cm.mu.Unlock()

	// 如果上下文不存在，则需要创建
	if !exists {
		// 创建并添加上下文时，不要在锁住的状态下调用 AddContext
		cm.AddContext(taskID)
		// 再次获取锁来获取新创建的上下文
		cm.mu.Lock()
		ctx = cm.contexts[taskID]
		cm.mu.Unlock()
	}
	return ctx
}
