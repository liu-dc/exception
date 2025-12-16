package try

import (
	"fmt"

	"github.com/liu-dc/exception/throw"
)

// ErrorHandler 异常处理函数签名
type ErrorHandler func(throw.Error)

// Try 核心控制结构（管理try逻辑和异常处理器）
type Try struct {
	specificHandlers map[int]ErrorHandler     // 特定异常码处理器
	groupHandlers    map[int][]ErrorHandler   // 异常分组处理器（新增）
	globalHandler    ErrorHandler             // 全局处理器
	fallbackHandler  ErrorHandler             // 兜底处理器
	tryFunc          func()                   // 待执行的核心逻辑
	filters          []func(throw.Error) bool // 异常过滤器（新增）
}

// Group 异常分组类型
type Group []int

// Run 入口函数：创建Try实例并绑定核心逻辑
func Run(tryHandler func()) *Try {
	return &Try{
		specificHandlers: make(map[int]ErrorHandler),
		groupHandlers:    make(map[int][]ErrorHandler),
		globalHandler:    nil,
		fallbackHandler:  nil,
		tryFunc:          tryHandler,
		filters:          make([]func(throw.Error) bool, 0),
	}
}

// Index 注册指定异常码的针对性处理器（支持链式调用）
// index：目标异常编码（禁止使用 0 和 throw.FallbackErrorIndex，避免与全局/兜底处理器冲突）
// handler：异常触发时执行的处理逻辑，适合特定异常的专属修复/降级动作
func (ts *Try) Index(index int, handler ErrorHandler) *Try {
	// 过滤非法异常码，避免覆盖全局/兜底处理器
	if handler != nil && index != 0 && index != throw.FallbackErrorIndex {
		ts.specificHandlers[index] = handler
	}
	return ts
}

// Catch 注册全局处理器（所有异常都会触发，支持链式调用）
// 场景：日志记录、监控上报等通用处理逻辑（无论是否匹配特定异常码，都会执行）
func (ts *Try) Catch(handler ErrorHandler) *Try {
	if handler != nil {
		ts.globalHandler = handler
	}
	return ts
}

// Unknown 注册兜底处理器（仅未匹配到特定异常码时触发，支持链式调用）
// 场景：未知异常的统一降级、告警（全局处理器执行后，若仍未处理则触发）
func (ts *Try) Unknown(handler ErrorHandler) *Try {
	if handler != nil {
		ts.fallbackHandler = handler
	}
	return ts
}

// Group 注册异常分组处理器（支持链式调用）
// group：异常编码列表（如 Group{400, 401, 403, 404}）
// handler：分组内所有异常码共有的处理逻辑
func (ts *Try) Group(group Group, handler ErrorHandler) *Try {
	if handler != nil && len(group) > 0 {
		for _, index := range group {
			if index != 0 && index != throw.FallbackErrorIndex {
				ts.groupHandlers[index] = append(ts.groupHandlers[index], handler)
			}
		}
	}
	return ts
}

// Filter 添加异常过滤器（支持链式调用）
// filter：返回true表示允许异常继续处理，false表示拦截异常
// 场景：根据异常上下文、堆栈等信息进行动态过滤
func (ts *Try) Filter(filter func(throw.Error) bool) *Try {
	if filter != nil {
		ts.filters = append(ts.filters, filter)
	}
	return ts
}

// executeHandlers 私有辅助方法：执行异常处理器（按优先级执行，确保逻辑可预测）
func (ts *Try) executeHandlers(err throw.Error) (handled bool, filtered bool) {
	// 步骤0：应用异常过滤器
	for _, filter := range ts.filters {
		if !filter(err) {
			// 过滤器拦截，不执行任何处理器
			return false, true
		}
	}

	handled = false
	hasSpecificHandler := false

	// 步骤1：执行特定异常码处理器（精准匹配优先）
	if handler, exists := ts.specificHandlers[err.Index]; exists {
		handler(err)
		hasSpecificHandler = true
		handled = true
	}

	// 步骤2：执行异常分组处理器
	if handlers, exists := ts.groupHandlers[err.Index]; exists {
		for _, handler := range handlers {
			handler(err)
			handled = true
		}
	}

	// 步骤3：执行兜底处理器（所有未知异常必执行）
	if !hasSpecificHandler && ts.fallbackHandler != nil {
		ts.fallbackHandler(err)
		handled = true
	}

	// 步骤4：执行全局处理器（所有异常必执行，如日志、监控）
	if ts.globalHandler != nil {
		ts.globalHandler(err)
		handled = true
	}

	return handled, false
}

// normalizeError 私有辅助方法：统一异常格式（确保所有panic都转为throw.Error，含堆栈）
func (ts *Try) normalizeError(recoverObj interface{}) throw.Error {
	// 已是throw.Error类型，直接返回（保留原始异常信息）
	if e, ok := recoverObj.(throw.Error); ok {
		return e
	}

	// 非预期panic：包装为兜底异常，补充堆栈和类型信息（便于排查未知问题）
	stack := throw.GetPanicStack()
	return throw.Error{
		Index:   throw.FallbackErrorIndex,
		Message: fmt.Sprintf("unexpected panic: %v (type: %T)", recoverObj, recoverObj),
		Stack:   stack,
		Context: make(map[string]interface{}), // 初始化空上下文
	}
}

// Finally 执行核心逻辑+捕获异常+执行最终清理逻辑（核心方法）
// 特性：
// 1. finally逻辑确保执行（无论是否发生异常）
// 2. 统一异常格式：非throw.Error类型的panic自动包装，保留完整堆栈
// 3. 处理器执行顺序：特定异常码处理器 → 全局处理器 → 兜底处理器（仅未匹配特定时）
// 4. 无任何处理器时透传异常，避免静默失败
// 5. 异常过滤器：拦截的异常不会触发任何处理器，也不会透传
func (ts *Try) Finally(finally func()) {
	// 优先注册finally延迟函数，确保其最后执行（无论是否panic）
	defer func() {
		if finally != nil {
			finally()
		}
	}()

	// 捕获panic并处理异常
	defer func() {
		if recoverObj := recover(); recoverObj != nil {
			if (recoverObj == throw.Pass{}) {
				//通过
				return
			}
			// 统一异常格式
			err := ts.normalizeError(recoverObj)
			// 执行处理器
			handled, filtered := ts.executeHandlers(err)
			// 无任何处理器时，透传异常（让上层感知未处理的异常）
			// 但如果是被过滤器拦截的异常，则不透传
			if !handled && !filtered {
				panic(err)
			}
		}
	}()

	// 执行核心业务逻辑
	ts.tryFunc()
}

// Do 快捷方法：无finally逻辑时直接执行（简化代码）
func (ts *Try) Do() {
	ts.Finally(nil)
}
