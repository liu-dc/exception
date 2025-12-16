# exception 包

一个基于 Go 语言 `panic/recover` 机制实现的增强型异常处理库，提供类 try-catch-finally 语法和丰富的异常管理功能。

## 功能特性

### 核心功能
- **类 try-catch-finally 语法**：提供熟悉的异常处理模式，降低学习成本
- **异常编码系统**：支持自定义异常编码，便于分类和定位问题
- **堆栈追踪**：自动捕获异常堆栈信息，便于问题排查
- **多级处理器**：支持特定异常码、分组、全局、兜底四级处理器
- **异常上下文**：支持在异常中携带键值对形式的上下文信息
- **异常分组**：支持为一组异常码注册共享处理器，简化批量处理
- **异常过滤**：支持动态拦截异常，可基于上下文或其他条件决定是否处理

### 扩展功能
- **异常验证**：提供条件异常抛出函数（True、False、FuncErr 等）
- **类型安全**：泛型支持（ValueErr 函数）
- **性能优化**：动态堆栈缓冲区、反射优化
- **兼容标准库**：实现 error 接口，与标准库无缝集成

## 安装

```bash
go get github.com/liu-dc/exception
```

## 包结构

```
exception/
├── throw/
│   └── throw.go      # 异常定义和抛出函数
├── try/
│   └── try.go        # 异常捕获和处理
└── README.md
```

## 快速开始

### 基本使用

```go
import (
    "fmt"
    "exception/try"
    "exception/throw"
)

func main() {
    try.Run(func() {
        fmt.Println("执行核心业务逻辑...")
        throw.Index(404, "资源不存在") // 抛出带编码的异常
    }).Index(404, func(err throw.Error) {
        fmt.Printf("捕获到404异常: %s\n", err.Message)
    }).Catch(func(err throw.Error) {
        fmt.Printf("全局日志: %d - %s\n", err.Index, err.Message)
    }).Do()
}
```

### 带上下文的异常

```go
try.Run(func() {
    userID := 123
    throw.IndexWithContext(403, "权限不足", "userID", userID) // 附加上下文
}).Index(403, func(err throw.Error) {
    if userID, exists := err.GetContext("userID"); exists {
        fmt.Printf("用户 %v 权限不足\n", userID)
    }
}).Do()
```

### try-finally 模式

```go
try.Run(func() {
    // 可能抛出异常的代码
}).Finally(func() {
    // 无论是否发生异常都会执行的清理代码
})
```

### 异常分组

```go
// 定义HTTP错误分组
httpErrors := try.Group{400, 401, 403, 404}

try.Run(func() {
    throw.Index(404, "资源不存在")
}).Group(httpErrors, func(err throw.Error) {
    fmt.Printf("HTTP错误 %d: %s\n", err.Index, err.Message)
}).Do()
```

### 异常过滤

```go
try.Run(func() {
    throw.IndexWithContext(403, "权限不足", "userID", 123)
}).Filter(func(err throw.Error) bool {
    // 只处理用户123的异常
    if userID, exists := err.GetContext("userID"); exists {
        if id, ok := userID.(int); ok && id == 123 {
            return true
        }
    }
    return false
}).Index(403, func(err throw.Error) {
    fmt.Println("处理用户123的权限不足异常")
}).Do()
```

## API 文档

### try 包

#### Run(func()) *Try
创建一个新的 Try 实例，绑定要执行的核心逻辑。

#### Index(index int, handler ErrorHandler) *Try
注册特定异常码的处理器，仅当抛出的异常编码匹配时执行。

#### Group(group Group, handler ErrorHandler) *Try
注册异常分组处理器，分组内所有异常码都会执行此处理器。

#### Catch(handler ErrorHandler) *Try
注册全局处理器，所有异常都会执行此处理器。

#### Unknown(handler ErrorHandler) *Try
注册兜底处理器，仅当没有匹配的特定异常码处理器时执行。

#### Filter(filter func(throw.Error) bool) *Try
添加异常过滤器，返回true表示允许异常继续处理，false表示拦截异常。

#### Finally(finally func())
执行核心逻辑并捕获异常，最后执行 finally 清理函数。

#### Do()
快捷方法，等同于 Finally(nil)。

#### Group 类型
异常分组类型，用于定义一组相关的异常码。

### throw 包

#### Error 结构体
```go
type Error struct {
    Index   int                    // 异常编码
    Message string                 // 异常描述
    Stack   string                 // 异常堆栈
    Context map[string]interface{} // 异常上下文
}
```

#### 异常抛出函数
- **Index(index int, message string)**：抛出带编码的异常
- **IndexNoStack(index int, message string)**：抛出带编码的异常（不含堆栈）
- **New(message string)**：抛出兜底异常
- **Err(err error)**：将标准 error 转换为异常
- **IndexErr(index int, err error)**：将标准 error 转换为带编码的异常
- **True(value bool, message string)**：条件为真时抛出异常
- **TrueIndex(index int, value bool, message string)**：条件为真时抛出带编码的异常
- **False(value bool, message string)**：条件为假时抛出异常
- **FalseIndex(index int, value bool, message string)**：条件为假时抛出带编码的异常
- **FuncErr(fn func() error)**：函数返回 error 时抛出异常
- **ValueErr[T any](value T, err error) T**：错误时抛出异常，否则返回值

#### 带上下文的异常函数
所有异常函数都有对应的带上下文版本，如：
- **IndexWithContext(index int, message string, key string, value interface{})**
- **IndexNoStackWithContext(index int, message string, key string, value interface{})**
- **NewWithContext(message string, key string, value interface{})**
- **ErrWithContext(err error, key string, value interface{})**
- **IndexErrWithContext(index int, err error, key string, value interface{})**
- **TrueWithContext(value bool, message string, key string, value interface{})**
- **TrueIndexWithContext(index int, value bool, message string, key string, value interface{})**
- **FalseWithContext(value bool, message string, key string, value interface{})**
- **FalseIndexWithContext(index int, value bool, message string, key string, value interface{})**
- **FuncErrWithContext(fn func() error, key string, value interface{})**
- **ValueErrWithContext[T any](value T, err error, key string, value interface{}) T**

#### 上下文管理方法
- **WithContext(key string, value interface{}) Error**：添加单个上下文
- **WithContextMap(ctx map[string]interface{}) Error**：批量添加上下文
- **GetContext(key string) (interface{}, bool)**：获取上下文
- **HasContext(key string) bool**：检查上下文是否存在

## 最佳实践

### 异常编码规范
- 使用正数作为业务异常编码（如 404、500）
- 避免使用 0 和 FallbackErrorIndex（math.MinInt）
- 建立统一的异常编码表，便于维护

### 异常处理器使用
- **Index**：处理特定业务异常，进行业务降级或修复
- **Group**：处理一组相关异常，如HTTP错误、数据库错误等
- **Catch**：进行全局日志记录、监控上报
- **Unknown**：处理未知异常，进行通用降级
- **Filter**：根据上下文或其他条件动态决定是否处理异常

### 性能优化
- 对于性能敏感场景，使用 IndexNoStack 避免堆栈捕获
- 上下文信息按需添加，避免不必要的内存开销

## 性能对比

| 功能 | 性能开销 | 说明 |
|------|---------|------|
| 基本异常抛出 | 低 | 仅创建异常对象 |
| 带堆栈的异常 | 中 | 捕获调用堆栈 |
| 带上下文的异常 | 中低 | 额外的 map 操作 |
| 异常捕获 | 低 | 基于 map 查找处理器 |

## 版本历史

### v1.2.0
- 新增异常分组功能，支持为一组异常码注册共享处理器
- 新增异常过滤功能，支持动态拦截异常
- 优化处理器执行顺序：特定→分组→兜底→全局

### v1.1.0
- 新增异常上下文支持
- 新增带上下文的异常函数
- 性能优化（动态堆栈缓冲区、反射优化）

### v1.0.0
- 初始版本，支持基本的 try-catch-finally 功能
- 异常编码和堆栈追踪

## 许可证

MIT