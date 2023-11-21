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
