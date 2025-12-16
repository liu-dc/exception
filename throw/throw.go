package throw

import (
	"math"
	"runtime"
)

const FallbackErrorIndex = math.MinInt

type Pass struct{}

// Error 异常实体（字段导出，支持外部包读取）
type Error struct {
	Index   int                    // 异常编码（自定义分类）
	Message string                 // 异常描述信息
	Stack   string                 // 异常堆栈（新增：便于问题排查）
	Context map[string]interface{} // 异常上下文（新增：携带额外信息）
}

// WithContext 为异常添加上下文
func (e Error) WithContext(key string, value interface{}) Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	e.Context[key] = value
	return e
}

// WithContextMap 为异常批量添加上下文
func (e Error) WithContextMap(ctx map[string]interface{}) Error {
	if e.Context == nil {
		e.Context = make(map[string]interface{})
	}
	for k, v := range ctx {
		e.Context[k] = v
	}
	return e
}

// GetContext 获取异常上下文
func (e Error) GetContext(key string) (interface{}, bool) {
	if e.Context == nil {
		return nil, false
	}
	value, exists := e.Context[key]
	return value, exists
}

// HasContext 检查异常是否包含指定上下文
func (e Error) HasContext(key string) bool {
	if e.Context == nil {
		return false
	}
	_, exists := e.Context[key]
	return exists
}

// Error 实现error接口，提高兼容性
func (e Error) Error() string {
	return e.Message
}

// GetPanicStack 私有工具函数：获取panic时的调用堆栈信息
func GetPanicStack() string {
	const initialSize = 1024 // 初始1KB，减少内存浪费
	buf := make([]byte, initialSize)
	n := runtime.Stack(buf, false) // false：不包含goroutine信息

	// 如果缓冲区不够，动态扩容
	if n == len(buf) {
		buf = make([]byte, len(buf)*2)
		n = runtime.Stack(buf, false)
	}

	return string(buf[:n])
}

// Index 抛出异常（含堆栈信息，便于排查）
func Index(index int, message string) {
	// 获取当前堆栈信息（跳过Throw函数本身，从调用者开始）
	stack := GetPanicStack()
	panic(Error{
		Index:   index,
		Message: message,
		Stack:   stack,
	})
}

// IndexWithContext 带上下文的Index函数
func IndexWithContext(index int, message string, key string, ctxValue interface{}) {
	stack := GetPanicStack()
	panic(Error{
		Index:   index,
		Message: message,
		Stack:   stack,
		Context: map[string]interface{}{key: ctxValue},
	})
}

// IndexNoStack 抛出异常（不含堆栈信息，性能优先）
func IndexNoStack(index int, message string) {
	panic(Error{
		Index:   index,
		Message: message,
		Stack:   "", // 不包含堆栈
	})
}

// IndexNoStackWithContext 带上下文的IndexNoStack函数
func IndexNoStackWithContext(index int, message string, key string, ctxValue interface{}) {
	panic(Error{
		Index:   index,
		Message: message,
		Stack:   "",
		Context: map[string]interface{}{key: ctxValue},
	})
}

func New(message string) {
	Index(FallbackErrorIndex, message)
}

// NewWithContext 带上下文的New函数
func NewWithContext(message string, key string, ctxValue interface{}) {
	IndexWithContext(FallbackErrorIndex, message, key, ctxValue)
}

func Err(err error) {
	if err != nil {
		Index(FallbackErrorIndex, err.Error()) //使用传入的 index 作为异常码
	}
}

// ErrWithContext 带上下文的Err函数
func ErrWithContext(err error, key string, ctxValue interface{}) {
	if err != nil {
		IndexWithContext(FallbackErrorIndex, err.Error(), key, ctxValue)
	}
}

func True(value bool, message string) {
	if value {
		Index(FallbackErrorIndex, message)
	}
}

// TrueWithContext 带上下文的True函数
func TrueWithContext(value bool, message string, key string, ctxValue interface{}) {
	if value {
		IndexWithContext(FallbackErrorIndex, message, key, ctxValue)
	}
}

func False(value bool, message string) {
	if !value {
		Index(FallbackErrorIndex, message)
	}
}

// FalseWithContext 带上下文的False函数
func FalseWithContext(value bool, message string, key string, ctxValue interface{}) {
	if !value {
		IndexWithContext(FallbackErrorIndex, message, key, ctxValue)
	}
}

func TrueIndex(index int, value bool, message string) {
	if value {
		Index(index, message)
	}
}

// TrueIndexWithContext 带上下文的TrueIndex函数
func TrueIndexWithContext(index int, value bool, message string, key string, ctxValue interface{}) {
	if value {
		IndexWithContext(index, message, key, ctxValue)
	}
}

func FalseIndex(index int, value bool, message string) {
	if !value {
		Index(index, message)
	}
}

// FalseIndexWithContext 带上下文的FalseIndex函数
func FalseIndexWithContext(index int, value bool, message string, key string, ctxValue interface{}) {
	if !value {
		IndexWithContext(index, message, key, ctxValue)
	}
}

func FuncErr(fn func() error) {
	err := fn()
	if err != nil {
		Index(FallbackErrorIndex, err.Error()) //使用传入的 FallbackErrorIndex 作为异常码
	}
}

// FuncErrWithContext 带上下文的FuncErr函数
func FuncErrWithContext(fn func() error, key string, ctxValue interface{}) {
	err := fn()
	if err != nil {
		IndexWithContext(FallbackErrorIndex, err.Error(), key, ctxValue)
	}
}

func ValueErr[T any](value T, err error) T {
	if err != nil {
		Index(FallbackErrorIndex, err.Error()) //使用传入的 index 作为异常码
	}
	return value
}

// ValueErrWithContext 带上下文的ValueErr函数
func ValueErrWithContext[T any](value T, err error, key string, ctxValue interface{}) T {
	if err != nil {
		IndexWithContext(FallbackErrorIndex, err.Error(), key, ctxValue)
	}
	return value
}

func IndexErr(index int, err error) {
	if err != nil {
		Index(index, err.Error()) //使用传入的 index 作为异常码
	}
}

// IndexErrWithContext 带上下文的IndexErr函数
func IndexErrWithContext(index int, err error, key string, ctxValue interface{}) {
	if err != nil {
		IndexWithContext(index, err.Error(), key, ctxValue)
	}
}
