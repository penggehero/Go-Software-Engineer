# 设计模式

## 责任链模式

顾名思义，责任链模式（Chain of Responsibility Pattern）为请求创建了一个接收者对象的链。这种模式给予请求的类型，对请求的发送者和接收者进行解耦。这种类型的设计模式属于行为型模式。

在这种模式中，通常每个接收者都包含对另一个接收者的引用。如果一个对象不能处理该请求，那么它会把相同的请求传给下一个接收者，依此类推。

### 介绍

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



## 命令模式

命令模式（Command Pattern）是一种数据驱动的设计模式，它属于行为型模式。请求以命令的形式包裹在对象中，并传给调用对象。调用对象寻找可以处理该命令的合适的对象，并把该命令传给相应的对象，该对象执行命令。

### 介绍

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



## 迭代器模式

迭代器模式（Iterator Pattern）是 Java 和 .Net 编程环境中非常常用的设计模式。这种模式用于顺序访问集合对象的元素，不需要知道集合对象的底层表示。

迭代器模式属于行为型模式。

### 介绍

**意图：** 提供一种方法顺序访问一个聚合对象中各个元素, 而又无须暴露该对象的内部表示。

**主要解决：** 不同的方式来遍历整个整合对象。

**何时使用：** 遍历一个聚合对象。

**如何解决：** 把在元素之间游走的责任交给迭代器，而不是聚合对象。

**关键代码：** 定义接口：hasNext, next。

**应用实例：** JAVA 中的 iterator。

**优点：** 
1、它支持以不同的方式遍历一个聚合对象。 
2、迭代器简化了聚合类。 
3、在同一个聚合上可以有多个遍历。 
4、在迭代器模式中，增加新的聚合类和迭代器类都很方便，无须修改原有代码。

**缺点：** 由于迭代器模式将存储数据和遍历数据的职责分离，增加新的聚合类需要对应增加新的迭代器类，类的个数成对增加，这在一定程度上增加了系统的复杂性。

**使用场景：**
1、访问一个聚合对象的内容而无须暴露它的内部表示。 
2、需要为聚合对象提供多种遍历方式。 
3、为遍历不同的聚合结构提供一个统一的接口。

**注意事项：**迭代器模式就是分离了集合对象的遍历行为，抽象出一个迭代器类来负责，这样既可以做到不暴露集合的内部结构，又可让外部代码透明地访问集合内部的数据。

```go
package design_pattern

import (
	"fmt"
	"testing"
)

// 迭代器模式
// 迭代器模式是一种行为设计模式，让你能在不暴露集合底层表现形式 （列表、 栈和树等）的情况下遍历集合中所有的元素。
// 在迭代器的帮助下， 客户端可以用一个迭代器接口以相似的方式遍历不同集合中的元素。
// 这里需要注意的是有两个典型的迭代器接口需要分清楚；
//  一个是集合类实现的可以创建迭代器的工厂方法接口一般命名为Iterable，包含的方法类似CreateIterator；
//  另一个是迭代器本身的接口，命名为Iterator，有Next及hasMore两个主要方法；

// Iterator 迭代器接口
type Iterator interface {
	HasNext() bool
	Next() interface{}
}

// Container 容器接口
type Container interface {
	GetIterator() Iterator
	GetIndex(index int) interface{}
	Len() int
}

// NameIterator 名字迭代器
type NameIterator struct {
	container Container
	index     int
}

// HasNext 是否有下一个
func (n *NameIterator) HasNext() bool {
	if n.index < n.container.Len() {
		return true
	}
	return false
}

// Next 下一个
func (n *NameIterator) Next() interface{} {
	if n.HasNext() {
		n.index++
		return n.container.GetIndex(n.index - 1)
	}
	return nil
}

// NameRepository 名字容器
type NameRepository struct {
	iterator *NameIterator
	names    []string
}

// GetIterator 获取迭代器
func (n *NameRepository) GetIterator() Iterator {
	return &NameIterator{
		container: n,
		index:     0,
	}
}

// GetIndex 获取指定索引的元素
func (n *NameRepository) GetIndex(index int) interface{} {
	return n.names[index]
}

// Len 获取长度
func (n *NameRepository) Len() int {
	return len(n.names)
}

// AddName 添加名字
func (n *NameRepository) AddName(s string) {
	n.names = append(n.names, s)
}

// NewNameRepository 创建名字容器
func NewNameRepository() *NameRepository {
	n := new(NameRepository)
	n.iterator = &NameIterator{n, 0}
	n.names = make([]string, 0)
	return n
}

// TestIterator 迭代器模式测试
func TestIterator(t *testing.T) {
	nameRepository := NewNameRepository()
	nameRepository.AddName("Robert")
	nameRepository.AddName("John")
	nameRepository.AddName("Julie")
	nameRepository.AddName("Lora")
	iterator := nameRepository.GetIterator()
	for iterator.HasNext() {
		name := iterator.Next().(string)
		fmt.Println("Name : " + name)
	}

	iterator2 := nameRepository.GetIterator()
	for iterator2.HasNext() {
		name := iterator2.Next().(string)
		fmt.Println("Name : " + name)
	}

	iterator3 := nameRepository.GetIterator()
	for iterator3.HasNext() {
		name := iterator3.Next().(string)
		fmt.Println("Name : " + name)
	}
}

```



## 中介者模式

是一种行为设计模式，能让你减少对象之间混乱无序的依赖关系。该模式会限制对象之间的直接交互，迫使它们通过一个中介者对象进行合作，将网状依赖变为星状依赖。

中介者能使得程序更易于修改和扩展，而且能更方便地对独立的组件进行复用，因为它们不再依赖于很多其他的类。

中介者模式与观察者模式之间的区别是，中介者模式解决的是同类或者不同类的多个对象之间多对多的依赖关系，观察者模式解决的是多个对象与一个对象之间的多对一的依赖关系

```go
package design_pattern

import (
	"fmt"
	"testing"
)

// 中介者模式
// 中介者模式是一种行为设计模式， 让你可以减少对象之间混乱无序的依赖关系。
// 该模式会限制对象之间的直接交互， 强迫它们通过一个中介者对象进行合作。
// 通过将对象彼此解耦， 也可更方便地对它们进行独立复用。
// 该模式会将系统中的对象分为两组： 具体组件（也就是有用的对象） 和中介者对象（负责协调具体组件之间的交互）。
// 由于组件之间几乎不知道彼此的存在， 所以它们必须通过中介者对象进行间接交流。
// 但是有一点需要注意， 中介者本身并不处理业务逻辑， 而只负责维护组件之间的关系。
//

// Mediator 中介者接口
type Mediator interface {
	// Send 发送消息
	Send(message string, user User)
}

// User 用户
type User struct {
	name     string
	mediator Mediator
}

// NewUser 创建用户
func NewUser(name string, mediator Mediator) *User {
	return &User{name: name, mediator: mediator}
}

// GetName 获取用户名字
func (u *User) GetName() string {
	return u.name
}

// Send 发送消息
func (u *User) Send(message string) {
	u.mediator.Send(message, *u)
}

// ChatRoom 聊天室
type ChatRoom struct {
	users []*User
}

// NewChatRoom 创建聊天室
func NewChatRoom() *ChatRoom {
	return &ChatRoom{users: make([]*User, 0)}
}

// Register 注册用户
func (c *ChatRoom) Register(user *User) {
	c.users = append(c.users, user)
}

// Send 发送消息
func (c *ChatRoom) Send(message string, user User) {
	for _, u := range c.users {
		if u.GetName() != user.GetName() {
			fmt.Printf("%s send message to %s: %s\n", user.GetName(), u.GetName(), message)
		}
	}
}

func TestMediator(t *testing.T) {
	chatRoom := NewChatRoom()
	user1 := NewUser("user1", chatRoom)
	user2 := NewUser("user2", chatRoom)
	user3 := NewUser("user3", chatRoom)
	chatRoom.Register(user1)
	chatRoom.Register(user2)
	chatRoom.Register(user3)
	user1.Send("hello")
	// Output:
	// user1 send message to user2: hello
	// user1 send message to user3: hello
}

```

