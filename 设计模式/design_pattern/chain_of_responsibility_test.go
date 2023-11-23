package design_pattern

import (
	"fmt"
	"testing"
)

// 责任链模式 Chain of Responsibility
/*
责任链模式是一种行为设计模式， 允许你将请求沿着处理者链进行发送。收到请求后，每个处理者均可对请求进行处理，或将其传递给链上的下个处理者。

该模式允许多个对象来对请求进行处理，而无需让发送者类与具体接收者类相耦合。链可在运行时由遵循标准处理者接口的任意处理者动态生成。
一般意义上的责任链模式是说，请求在链上流转时任何一个满足条件的节点处理完请求后就会停止流转并返回，不过还可以根据不同的业务情况做一些改进：

1. 请求可以流经处理链的所有节点，不同节点会对请求做不同职责的处理；
2. 可以通过上下文参数保存请求对象及上游节点的处理结果，供下游节点依赖，并进一步处理；
3. 处理链可支持节点的异步处理，通过实现特定接口判断，是否需要异步处理；
4. 责任链对于请求处理节点可以设置停止标志位，不是异常，是一种满足业务流转的中断；
5. 责任链的拼接方式存在两种，一种是节点遍历，一个节点一个节点顺序执行；另一种是节点嵌套，内层节点嵌入在外层节点执行逻辑中，类似递归，或者“回”行结构；
6. 责任链的节点嵌套拼接方式多被称为拦截器链或者过滤器链，更易于实现业务流程的切面，比如监控业务执行时长，日志输出，权限校验等；
*/

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
