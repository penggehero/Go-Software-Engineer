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
