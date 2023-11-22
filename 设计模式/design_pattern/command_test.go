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
