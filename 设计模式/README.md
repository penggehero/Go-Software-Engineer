# 设计模式

## 责任链模式

顾名思义，责任链模式（Chain of Responsibility Pattern）为请求创建了一个接收者对象的链。这种模式给予请求的类型，对请求的发送者和接收者进行解耦。这种类型的设计模式属于行为型模式。

在这种模式中，通常每个接收者都包含对另一个接收者的引用。如果一个对象不能处理该请求，那么它会把相同的请求传给下一个接收者，依此类推。

## 介绍

**意图：** 避免请求发送者与接收者耦合在一起，让多个对象都有可能接收请求，将这些对象连接成一条链，并且沿着这条链传递请求，直到有对象处理它为止。

**主要解决：** 职责链上的处理者负责处理请求，客户只需要将请求发送到职责链上即可，无须关心请求的处理细节和请求的传递，所以职责链将请求的发送者和请求的处理者解耦了。

**何时使用：** 在处理消息的时候以过滤很多道。

**如何解决：** 拦截的类都实现统一接口。

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



# 命令模式

命令模式（Command Pattern）是一种数据驱动的设计模式，它属于行为型模式。请求以命令的形式包裹在对象中，并传给调用对象。调用对象寻找可以处理该命令的合适的对象，并把该命令传给相应的对象，该对象执行命令。

## 介绍

**意图：**  将一个请求封装成一个对象，从而使您可以用不同的请求对客户进行参数化。

**主要解决： ** 在软件系统中，行为请求者与行为实现者通常是一种紧耦合的关系，但某些场合，比如需要对行为进行记录、撤销或重做、事务等处理时，这种无法抵御变化的紧耦合的设计就不太合适。

**何时使用：**  在某些场合，比如要对行为进行"记录、撤销/重做、事务"等处理，这种无法抵御变化的紧耦合是不合适的。在这种情况下，如何将"行为请求者"与"行为实现者"解耦？将一组行为抽象为对象，可以实现二者之间的松耦合。

**如何解决：**  通过调用者调用接受者执行命令，顺序：调用者→命令→接受者。

**关键代码：**  定义三个角色：

1、received 真正的命令执行对象 

2、Command 

3、invoker 使用命令对象的入口

**应用实例：**  struts 1 中的 action 核心控制器 ActionServlet 只有一个，相当于 Invoker，而模型层的类会随着不同的应用有不同的模型类，相当于具体的 Command。

**优点：**  1、降低了系统耦合度。 2、新的命令可以很容易添加到系统中去。

**缺点：**  使用命令模式可能会导致某些系统有过多的具体命令类。

**注意事项：** 系统需要支持命令的撤销(Undo)操作和恢复(Redo)操作，也可以考虑使用命令模式，见命令模式的扩展。

```go
package design_pattern

import (
	"fmt"
	"testing"
)

// command 命令模式
// 命令模式是一种行为型设计模式，它允许将请求封装为一个对象，从而使不同请求的调用者能够独立于接收者、请求的内容以及请求的执行方式。
// 在这个示例中，我们将实现一个库存管理系统，用命令模式来实现买入库存和卖出库存的功能。

// Command 命令接口
type Command interface {
	execute()
}

// Stock 库存
type Stock struct {
	quantity int // 库存数量
}

// NewStock 创建库存
func NewStock(quantity int) *Stock {
	return &Stock{quantity: quantity}
}

// buy 买入库存
func (s *Stock) buy() {
	s.quantity++
	fmt.Println("buy stock, quantity:", s.quantity)
}

// sell 卖出库存
func (s *Stock) sell() {
	if s.quantity <= 0 {
		fmt.Println("sell stock failed, quantity is 0")
		return
	}
	s.quantity--
	fmt.Println("sell stock, quantity:", s.quantity)
}

// BuyStock 买入库存命令
type BuyStock struct {
	stock *Stock
}

// NewBuyStock 创建买入库存命令
func NewBuyStock(stock *Stock) *BuyStock {
	return &BuyStock{stock: stock}
}

// execute 执行命令
func (b *BuyStock) execute() {
	b.stock.buy()
}

// SellStock 卖出库存命令
type SellStock struct {
	stock *Stock
}

// NewSellStock 创建卖出库存命令
func NewSellStock(stock *Stock) *SellStock {
	return &SellStock{stock: stock}
}

// execute 执行命令
func (s *SellStock) execute() {
	s.stock.sell()
}

// Broker 命令调用者
type Broker struct {
	orders []Command
}

// NewBroker 创建Broker
func NewBroker() *Broker {
	return &Broker{orders: make([]Command, 0)}
}

// takeOrder 接收命令
func (b *Broker) takeOrder(order Command) {
	b.orders = append(b.orders, order)
}

// placeOrders 执行命令
func (b *Broker) placeOrders() {
	for _, order := range b.orders {
		order.execute()
	}
	// 执行完命令后清空命令列表
	b.orders = b.orders[:0]
}

// TestCommand 命令模式测试
func TestCommand(t *testing.T) {
	stock := NewStock(1)
	broker := NewBroker()

	broker.takeOrder(NewBuyStock(stock))
	broker.takeOrder(NewBuyStock(stock))
	broker.takeOrder(NewSellStock(stock))

	broker.placeOrders()
}

```

