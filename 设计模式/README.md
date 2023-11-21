# 设计模式

## 责任链模式

顾名思义，责任链模式（Chain of Responsibility Pattern）为请求创建了一个接收者对象的链。这种模式给予请求的类型，对请求的发送者和接收者进行解耦。这种类型的设计模式属于行为型模式。

在这种模式中，通常每个接收者都包含对另一个接收者的引用。如果一个对象不能处理该请求，那么它会把相同的请求传给下一个接收者，依此类推。

## 介绍

**意图：**避免请求发送者与接收者耦合在一起，让多个对象都有可能接收请求，将这些对象连接成一条链，并且沿着这条链传递请求，直到有对象处理它为止。

**主要解决：**职责链上的处理者负责处理请求，客户只需要将请求发送到职责链上即可，无须关心请求的处理细节和请求的传递，所以职责链将请求的发送者和请求的处理者解耦了。

**何时使用：**在处理消息的时候以过滤很多道。

**如何解决：**拦截的类都实现统一接口。

**优点：** 

1、降低耦合度。它将请求的发送者和接收者解耦。 

2、简化了对象。使得对象不需要知道链的结构。 

3、增强给对象指派职责的灵活性。通过改变链内的成员或者调动它们的次序，允许动态地新增或者删除责任。 

4、增加新的请求处理类很方便。

**缺点：** 

1、不能保证请求一定被接收。 

2、系统性能将受到一定影响，而且在进行代码调试时不太方便，可能会造成循环调用。 

3、可能不容易观察运行时的特征，有碍于除错。

**使用场景：** 

1、有多个对象可以处理同一个请求，具体哪个对象处理该请求由运行时刻自动确定。 

2、在不明确指定接收者的情况下，向多个对象中的一个提交一个请求。 

3、可动态指定一组对象处理请求。



案例:

```go
package design_pattern

import (
	"fmt"
	"testing"
)

// 责任链模式 Chain of Responsibility

const (
	INFO = iota
	DEBUG
	ERROR
)

type Logger interface {
	Write(message string)
}

type LogHandler interface {
	LogMessage(level int, message string)
	SetNextLogger(nextLogger LogHandler)
}

// baseLogHandler 基础日志处理器
type baseLogHandler struct {
	level      int        // 默认日志级别
	w          Logger     // 当前的日志处理
	nextLogger LogHandler // 下一个日志处理器
}

// SetNextLogger 设置下一个日志处理器
func (c *baseLogHandler) SetNextLogger(nextLogger LogHandler) {
	c.nextLogger = nextLogger
}

// LogMessage 日志处理
func (c *baseLogHandler) LogMessage(level int, message string) {
	if c.level <= level {
		c.w.Write(message)
	}
	if c.nextLogger != nil {
		c.nextLogger.LogMessage(level, message)
	}
}

// ConsoleLogger 控制台日志处理器
type ConsoleLogger struct {
	baseLogHandler
}

// Write 写入控制台日志
func (c *ConsoleLogger) Write(message string) {
	fmt.Println("ConsoleLogger:", message)
}

// NewConsoleLogger 创建控制台日志处理器
func NewConsoleLogger(level int) *ConsoleLogger {
	c := new(ConsoleLogger)
	c.baseLogHandler = baseLogHandler{level, c, nil}
	return c
}

// ErrorLogger 错误日志处理器
type ErrorLogger struct {
	baseLogHandler
}

// Write 写入错误日志
func (c *ErrorLogger) Write(message string) {
	fmt.Println("ErrorLogger:", message)
}

// NewErrorLogger 创建错误日志处理器
func NewErrorLogger(level int) *ErrorLogger {
	c := new(ErrorLogger)
	c.baseLogHandler = baseLogHandler{level, c, nil}
	return c
}

// FileLogger 文件日志处理器
type FileLogger struct {
	baseLogHandler
}

// Write 写入文件日志
func (c *FileLogger) Write(message string) {
	fmt.Println("FileLogger:", message)
}

// NewFileLogger 创建文件日志处理器
func NewFileLogger(level int) *FileLogger {
	c := new(FileLogger)
	c.baseLogHandler = baseLogHandler{level, c, nil}
	return c
}

// GetChainOfLogHandlers 获取日志处理器链
func GetChainOfLogHandlers() LogHandler {
	consoleLogger := NewConsoleLogger(INFO)
	errorLogger := NewErrorLogger(ERROR)
	fileLogger := NewFileLogger(DEBUG)
	errorLogger.SetNextLogger(fileLogger)
	fileLogger.SetNextLogger(consoleLogger)
	return errorLogger
}

// TestChainOfResponsibility 测试责任链模式
// Output:
// ConsoleLogger: This is an information.
// FileLogger: This is a debug level information.
// ConsoleLogger: This is a debug level information.
// ErrorLogger: This is an error information.
// FileLogger: This is an error information.
// ConsoleLogger: This is an error information.
func TestChainOfResponsibility(t *testing.T) {
	loggers := GetChainOfLogHandlers()
	loggers.LogMessage(INFO, "This is an information.")
	loggers.LogMessage(DEBUG, "This is a debug level information.")
	loggers.LogMessage(ERROR, "This is an error information.")
}
```