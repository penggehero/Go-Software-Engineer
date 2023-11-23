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
