#   grpc-go 源码Sever浅读



grpc平时工作在使用，可是一直没好好深入过。通过对源码的阅读，学习谷歌工程师们优秀的设计和思想，从而提高自己的视野和水平。

## helloworld

https://github.com/grpc/grpc-go/tree/master/examples/helloworld

Grpc官网提供了简单入门案例，从最近的案例出发，一步一步深入Grpc内部。

简单了解一下Protocol buffers。

> Protocol buffers provide a language-neutral, platform-neutral, extensible mechanism for serializing structured data in a forward-compatible and backward-compatible way. It’s like JSON, except it's smaller and faster, and it generates native language bindings.
>
> Protocol buffers提供了一种语言中立、平台中立、可扩展的机制，用于以向前兼容和向后兼容的方式序列化结构化数据。它类似于 JSON，只是它更小更快，并且生成本地语言绑定。

proto

```protobuf
syntax = "proto3";

option go_package = "google.golang.org/grpc/examples/helloworld/helloworld";
option java_multiple_files = true;
option java_package = "io.grpc.examples.helloworld";
option java_outer_classname = "HelloWorldProto";

package helloworld;

// The greeting service definition.
service Greeter {
  // Sends a greeting
  rpc SayHello (HelloRequest) returns (HelloReply) {}
}

// The request message containing the user's name.
message HelloRequest {
  string name = 1;
}

// The response message containing the greetings
message HelloReply {
  string message = 1;
}
```

greeter_server

```go
/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a server for Greeter service.
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"

	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

var (
	port = flag.Int("port", 50051, "The server port")
)

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

func main() {
	flag.Parse()
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	pb.RegisterGreeterServer(s, &server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

```

greeter_client

```go
/*
 *
 * Copyright 2015 gRPC authors.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

// Package main implements a client for Greeter service.
package main

import (
   "context"
   "flag"
   "log"
   "time"

   "google.golang.org/grpc"
   "google.golang.org/grpc/credentials/insecure"
   pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

const (
   defaultName = "world"
)

var (
   addr = flag.String("addr", "localhost:50051", "the address to connect to")
   name = flag.String("name", defaultName, "Name to greet")
)

func main() {
   flag.Parse()
   // Set up a connection to the server.
   conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
   if err != nil {
      log.Fatalf("did not connect: %v", err)
   }
   defer conn.Close()
   c := pb.NewGreeterClient(conn)

   // Contact the server and print out its response.
   ctx, cancel := context.WithTimeout(context.Background(), time.Second)
   defer cancel()
   r, err := c.SayHello(ctx, &pb.HelloRequest{Name: *name})
   if err != nil {
      log.Fatalf("could not greet: %v", err)
   }
   log.Printf("Greeting: %s", r.GetMessage())
}
```

上面的示例很简单，包含一个server和client，Sever只提供SayHello的方法。

##  Server核心代码

简单的来说，Server核心有4个步骤：

1. 创建Listener监听器。
2. 创建Server 服务。
3. 服务注册，注册实现了helloworld.GreeterServer接口的结构体。
4. 向Sever中注册Listener监听器，并启动服务。

### 创建Listener

```go
lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
```

这里调用net的包的Listen方法，返回Listener对象。

net包中的Listener是个接口，很方便自己实现 Listener。

```go
type Listener interface {
   // Accept waits for and returns the next connection to the listener.
   Accept() (Conn, error)

   // Close closes the listener.
   // Any blocked Accept operations will be unblocked and return errors.
   Close() error

   // Addr returns the listener's network address.
   Addr() Addr
}
```

### 创建Server 服务

```go
func NewServer(opt ...ServerOption) *Server {
   // 加载配置
   opts := defaultServerOptions
   for _, o := range extraServerOptions {
      o.apply(&opts)
   }
   for _, o := range opt {
      o.apply(&opts)
   }
   // 创建Server对象
   s := &Server{
      lis:      make(map[net.Listener]bool),
      opts:     opts,
      conns:    make(map[string]map[transport.ServerTransport]bool),
      services: make(map[string]*serviceInfo),
      quit:     grpcsync.NewEvent(),
      done:     grpcsync.NewEvent(),
      czData:   new(channelzData),
   }
   // 加载一元拦截器和流式拦截器
   chainUnaryServerInterceptors(s)
   chainStreamServerInterceptors(s)
   // GracefullStop 使用 
   s.cv = sync.NewCond(&s.mu)
    
   // 开启trace 
   if EnableTracing {
      _, file, line, _ := runtime.Caller(1)
      s.events = trace.NewEventLog("grpc.Server", fmt.Sprintf("%s:%d", file, line))
   }

   // 开启 ServerWorker 
   if s.opts.numServerWorkers > 0 {
      s.initServerWorkers()
   }

   // 注册 channelzID 
   s.channelzID = channelz.RegisterServer(&channelzServer{s}, "")
   channelz.Info(logger, s.channelzID, "Server created")
   return s
}
```

#### grpc的配置参数加载（Option模式）

grpc的配置参数设计的很巧妙，下面以服务端为例子。

serverOptions是服务端的所有参数的结构体。

```go
type serverOptions struct {
	creds                 credentials.TransportCredentials
	codec                 baseCodec
	cp                    Compressor
	dc                    Decompressor
	unaryInt              UnaryServerInterceptor
	streamInt             StreamServerInterceptor
	chainUnaryInts        []UnaryServerInterceptor
	chainStreamInts       []StreamServerInterceptor
	inTapHandle           tap.ServerInHandle
	statsHandlers         []stats.Handler
	maxConcurrentStreams  uint32
	maxReceiveMessageSize int
	maxSendMessageSize    int
	unknownStreamDesc     *StreamDesc
	keepaliveParams       keepalive.ServerParameters
	keepalivePolicy       keepalive.EnforcementPolicy
	initialWindowSize     int32
	initialConnWindowSize int32
	writeBufferSize       int
	readBufferSize        int
	connectionTimeout     time.Duration
	maxHeaderListSize     *uint32
	headerTableSize       *uint32
	numServerWorkers      uint32
}
```

ServerOption 服务端参数是一个接口，需要实现apply方法，apply传递的是serverOptions的指针。

```go
// A ServerOption sets options such as credentials, codec and keepalive parameters, etc.
type ServerOption interface {
	apply(*serverOptions)
}
```

funcServerOption 包装了一个将 serverOptions 修改为 ServerOption 接口的实现结构体函数。

```go
// funcServerOption wraps a function that modifies serverOptions into an
// implementation of the ServerOption interface.
type funcServerOption struct {
	f func(*serverOptions)
}

func (fdo *funcServerOption) apply(do *serverOptions) {
	fdo.f(do)
}

func newFuncServerOption(f func(*serverOptions)) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}
```

funcServerOption内部包含处理serverOptions的方法f，并且实现了ServerOption接口。

KeepaliveParams方法的例子


```go
// KeepaliveParams returns a ServerOption that sets keepalive and max-age parameters for the server.
func KeepaliveParams(kp keepalive.ServerParameters) ServerOption {
   if kp.Time > 0 && kp.Time < time.Second {
      logger.Warning("Adjusting keepalive ping interval to minimum period of 1s")
      kp.Time = time.Second
   }

   return newFuncServerOption(func(o *serverOptions) {
      o.keepaliveParams = kp
   })
}
```

通过newFuncServerOption方法把keepaliveParams的方法传递到funcServerOption对象中，此时KeepaliveParams方法被包装成了apply方法，funcServerOption对象也变成了ServerOption对象。

返回看主逻辑。

```go

for _, o := range opt {
   o.apply(&opts)
}
```

opts对象就是默认的serverOptions结构体，opt就是ServerOption对象，外部传的方法都被包装成了apply方法，执行apply对象后，参数被成功加载。



那么为什么要设计的怎么麻烦的？

Option函数的方式，方便扩展参数，每次添加只需要中新的方法即可，不必修改原来的代码。

此外，对参数进行内部校验等操作，校验的工作可以放内部方法中做，和其他逻辑解耦。



#### Interceptor 加载拦截器

以Server unary（一元方式，一问一答的方式拦截器为例子，server有2类一元拦截器

```go
	unaryInt              UnaryServerInterceptor
	chainUnaryInts        []UnaryServerInterceptor
```

- unaryInt ：默认拦截器，最终server执行的拦截器
- chainUnaryInts ：链式拦截器，外部传的拦截器

chainUnaryServerInterceptors和chainUnaryInterceptors 方法将 unaryInt和chainUnaryInts组装成一个拦截器.

```go
func chainUnaryServerInterceptors(s *Server) {
	// Prepend opts.unaryInt to the chaining interceptors if it exists, since unaryInt will
	// be executed before any other chained interceptors.
    // 获取chainUnaryInts
	interceptors := s.opts.chainUnaryInts
    // 将unaryInt和chainUnaryInts合并为interceptors
	if s.opts.unaryInt != nil {
		interceptors = append([]UnaryServerInterceptor{s.opts.unaryInt}, s.opts.chainUnaryInts...)
	}

	var chainedInt UnaryServerInterceptor
	if len(interceptors) == 0 {
		chainedInt = nil
	} else if len(interceptors) == 1 {
		chainedInt = interceptors[0]
	} else {
        // 将interceptors 合并一个Interceptor
		chainedInt = chainUnaryInterceptors(interceptors)
	}

	s.opts.unaryInt = chainedInt
}

```


```go
func chainUnaryInterceptors(interceptors []UnaryServerInterceptor) UnaryServerInterceptor {
   return func(ctx context.Context, req interface{}, info *UnaryServerInfo, handler UnaryHandler) (interface{}, error) {
      // the struct ensures the variables are allocated together, rather than separately, since we
      // know they should be garbage collected together. This saves 1 allocation and decreases
      // time/call by about 10% on the microbenchmark.
      var state struct {
         i    int          // 计时器
         next UnaryHandler // handler处理器
      }
      state.next = func(ctx context.Context, req interface{}) (interface{}, error) {
         // 执行最后一个拦截器，直接传递handler
         if state.i == len(interceptors)-1 {
            return interceptors[state.i](ctx, req, info, handler)
         }
         // 游标+1 
         state.i++
         // 执行当前拦截器,传递 state.next，state.next被包装成了handler，当拦截器执行handler方法时，其实执行的state.next方法。 
         return interceptors[state.i-1](ctx, req, info, state.next)
      }
      return state.next(ctx, req)
   }
}
```

chainUnaryInterceptors 方法 将interceptors 合并一个Interceptor，这里巧妙的将多个拦截器变为一个链式执行的拦截器。

state 作为闭包函数的自由变量，记录着 游标 和 handler处理器。

当执行拦截器完自己本身的逻辑之后，变会执行handler方法。例如：

```go
func MyInterceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo,
   handler grpc.UnaryHandler) (resp interface{}, err error) {
   fmt.Println("执行拦截逻辑...")
   return handler(ctx, req)
}
```

此时handler方法被包装成了state.next方法，在其内部通过state的游标找到下一个拦截器并执行之。

### 服务注册

 ServiceRegistrar方法 用于向服务执行注册。

```go
// ServiceRegistrar wraps a single method that supports service registration. It
// enables users to pass concrete types other than grpc.Server to the service
// registration methods exported by the IDL generated code.
type ServiceRegistrar interface {
   // RegisterService registers a service and its implementation to the
   // concrete type implementing this interface.  It may not be called
   // once the server has started serving.
   // desc describes the service and its methods and handlers. impl is the
   // service implementation which is passed to the method handlers.
   RegisterService(desc *ServiceDesc, impl interface{})
}
```

核心参数

- ServiceDesc：rpc 服务的描述信息。
- impl: rpc 服务的具体实现。

在helloworld这个例子中。

```go
// GreeterServer is the server API for Greeter service.
// All implementations must embed UnimplementedGreeterServer
// for forward compatibility
type GreeterServer interface {
   // Sends a greeting
   SayHello(context.Context, *HelloRequest) (*HelloReply, error)
   mustEmbedUnimplementedGreeterServer()
}

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}
```

注册的具体逻辑：

```go
func RegisterGreeterServer(s grpc.ServiceRegistrar, srv GreeterServer) {
   s.RegisterService(&Greeter_ServiceDesc, srv)
}
```

Greeter 服务的 ServiceDesc

```go
var Greeter_ServiceDesc = grpc.ServiceDesc{
   ServiceName: "helloworld.Greeter",
   HandlerType: (*GreeterServer)(nil),
   Methods: []grpc.MethodDesc{
      {
         MethodName: "SayHello",
         Handler:    _Greeter_SayHello_Handler,
      },
   },
   Streams:  []grpc.StreamDesc{},
   Metadata: "examples/helloworld/helloworld/helloworld.proto",
}
```

关键参数：

- ServiceName：服务名
- HandlerType：处理器类型
- Methods：方法
- Streams：流（流式传输使用）
- Metadata：元数据信息

这里暂时不作深入探究，不过可以发现，ServiceDesc包含了Rpc调用的所有关键信息，比如服务名、方法名、服务的具体实现等等。

ServiceRegistrar的具体实现

```go
func (s *Server) RegisterService(sd *ServiceDesc, ss interface{}) {
   // 校验实现类
   if ss != nil {
      ht := reflect.TypeOf(sd.HandlerType).Elem()
      st := reflect.TypeOf(ss)
      if !st.Implements(ht) {
         logger.Fatalf("grpc: Server.RegisterService found the handler of type %v that does not satisfy %v", st, ht)
      }
   }
   s.register(sd, ss)
}
```

这里还可以注册一个空实现，哈哈哈哈。

```go
func (s *Server) register(sd *ServiceDesc, ss interface{}) {
   s.mu.Lock()
   defer s.mu.Unlock()
   s.printf("RegisterService(%q)", sd.ServiceName)
   if s.serve {
      logger.Fatalf("grpc: Server.RegisterService after Server.Serve for %q", sd.ServiceName)
   }
   if _, ok := s.services[sd.ServiceName]; ok {
      logger.Fatalf("grpc: Server.RegisterService found duplicate service registration for %q", sd.ServiceName)
   }
   info := &serviceInfo{
      serviceImpl: ss,
      methods:     make(map[string]*MethodDesc),
      streams:     make(map[string]*StreamDesc),
      mdata:       sd.Metadata,
   }
   for i := range sd.Methods {
      d := &sd.Methods[i]
      info.methods[d.MethodName] = d
   }
   for i := range sd.Streams {
      d := &sd.Streams[i]
      info.streams[d.StreamName] = d
   }
   s.services[sd.ServiceName] = info
}
```

Server对象的services 存储服务的信息，提供 服务名 和 服务信息（包含实现结构体） 的映射。 

`services map[string]*serviceInfo // service name -> service info`

### 服务启动

在彻底理解Serve要明白几个基本的概念。

#### Event 事件

- Event：表示将来可能发生的一次性事件。
- Fire：标识完成某个事件的结束。
- Done：返回，等待Fire方法。
- HasFired：事情是否结束。

```go
// Package grpcsync implements additional synchronization primitives built upon
// the sync package.
package grpcsync

import (
	"sync"
	"sync/atomic"
)

// Event represents a one-time event that may occur in the future.
type Event struct {
	fired int32
	c     chan struct{}
	o     sync.Once
}

// Fire causes e to complete.  It is safe to call multiple times, and
// concurrently.  It returns true iff this call to Fire caused the signaling
// channel returned by Done to close.
func (e *Event) Fire() bool {
	ret := false
	e.o.Do(func() {
		atomic.StoreInt32(&e.fired, 1)
		close(e.c)
		ret = true
	})
	return ret
}

// Done returns a channel that will be closed when Fire is called.
func (e *Event) Done() <-chan struct{} {
	return e.c
}

// HasFired returns true if Fire has been called.
func (e *Event) HasFired() bool {
	return atomic.LoadInt32(&e.fired) == 1
}

// NewEvent returns a new, ready-to-use Event.
func NewEvent() *Event {
	return &Event{c: make(chan struct{})}
}

```

Sever中2个主要的Event 事件

```go
quit               *grpcsync.Event
done               *grpcsync.Event
```

- quit：退出事件
- done：完成事件，quit事件结束后，触发done事件（Stop and GracefulStop 使用）

#### ServerTransport

ServerTransport是所有gRPC服务器端处理传输的公共接口。

ServerTransport的方法可以从多个 goroutine 并发调用，但是给定 Stream 的 Write 方法只能被串行调用。

````go
type ServerTransport interface {
	// HandleStreams receives incoming streams using the given handler.
	HandleStreams(func(*Stream), func(context.Context, string) context.Context)

	// WriteHeader sends the header metadata for the given stream.
	// WriteHeader may not be called on all streams.
	WriteHeader(s *Stream, md metadata.MD) error

	// Write sends the data for the given stream.
	// Write may not be called on all streams.
	Write(s *Stream, hdr []byte, data []byte, opts *Options) error

	// WriteStatus sends the status of a stream to the client.  WriteStatus is
	// the final call made on a stream and always occurs.
	WriteStatus(s *Stream, st *status.Status) error

	// Close tears down the transport. Once it is called, the transport
	// should not be accessed any more. All the pending streams and their
	// handlers will be terminated asynchronously.
	Close()

	// RemoteAddr returns the remote network address.
	RemoteAddr() net.Addr

	// Drain notifies the client this ServerTransport stops accepting new RPCs.
	Drain()

	// IncrMsgSent increments the number of message sent through this transport.
	IncrMsgSent()

	// IncrMsgRecv increments the number of message received through this transport.
	IncrMsgRecv()
}
````

ServerTransport 具体实现主要是http2Server，Grpc 使用标准的HTTP 2.0。



#### Serve方法

Serve 接受侦听器进来的连接，为每个连接创建一个新的 ServerTransport 和 服务 goroutine。 

服务 goroutine 读取 gRPC 请求，然后调用注册的handlers来相应请求。 

```go
// Serve accepts incoming connections on the listener lis, creating a new
// ServerTransport and service goroutine for each. The service goroutines
// read gRPC requests and then call the registered handlers to reply to them.
// Serve returns when lis.Accept fails with fatal errors.  lis will be closed when
// this method returns.
// Serve will return a non-nil error unless Stop or GracefulStop is called.
func (s *Server) Serve(lis net.Listener) error {
	// 上锁
	s.mu.Lock()
	s.printf("serving")
	// 设置serve状态，在serve状态不允许注册服务哦
	s.serve = true
	if s.lis == nil {
		// Serve called after Stop or GracefulStop.
		s.mu.Unlock()
		lis.Close()
		return ErrServerStopped
	}

	s.serveWG.Add(1)
	defer func() {
		s.serveWG.Done()
		// 退出事件执行完毕
    
		if s.quit.HasFired() {
			// Stop or GracefulStop called; block until done and return nil.
			// 等待done执行完毕，完成Stop or GracefulStop
			<-s.done.Done()
		}
	}()

	// 存储Listener信息
	ls := &listenSocket{Listener: lis}
	s.lis[ls] = true

	defer func() {
		// Serve方法退出时，关闭Listener并删除Listener信息
		s.mu.Lock()
		if s.lis != nil && s.lis[ls] {
			ls.Close()
			delete(s.lis, ls)
		}
		s.mu.Unlock()
	}()

	// 不太明白，这里向一个db进行注册
	var err error
	ls.channelzID, err = channelz.RegisterListenSocket(ls, s.channelzID, lis.Addr().String())
	if err != nil {
		s.mu.Unlock()
		return err
	}
    // 释放锁
	s.mu.Unlock()
	channelz.Info(logger, ls.channelzID, "ListenSocket created")

	// 睡多久重试失败
	var tempDelay time.Duration // how long to sleep on accept failure
	for {
		// 处理连接
		rawConn, err := lis.Accept()
		if err != nil {
			// 如果是临时性的错误
			if ne, ok := err.(interface {
				Temporary() bool
			}); ok && ne.Temporary() {
				// 刚刚开始睡5ms
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					// 再次翻倍
					tempDelay *= 2
				}
				// 最大只睡1s
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				s.mu.Lock()
				s.printf("Accept error: %v; retrying in %v", err, tempDelay)
				s.mu.Unlock()
				timer := time.NewTimer(tempDelay)
				// 进入阻塞，等待退出事件或者时间到期
				select {
				case <-timer.C:
				case <-s.quit.Done():
					timer.Stop()
					return nil
				}
				// 退出,等待下次连接
				continue
			}
			s.mu.Lock()
			s.printf("done serving; Accept = %v", err)
			s.mu.Unlock()

			// 此时刚好发生退出事件，直接返回nil
			if s.quit.HasFired() {
				return nil
			}
			// 返回错误
			return err
		}
		// tempDelay置0
		tempDelay = 0
		// Start a new goroutine to deal with rawConn so we don't stall this Accept
		// loop goroutine.
		//
		// Make sure we account for the goroutine so GracefulStop doesn't nil out
		// s.conns before this conn can be added.
        // 开一个goroutine去处理连接
		s.serveWG.Add(1)
		go func() {
			s.handleRawConn(lis.Addr().String(), rawConn)
			s.serveWG.Done()
		}()
	}
}
```

服务启动的过程，可大致分下面几步：

- 退出时，defer开启 退出事件， 完成Stop or GracefulStop
- 退出时，defer关闭Listener并删除Listener信息
- 完成channelz.RegisterListenSocket 注册
- 开启  Accept loop 监听循环，等待连接Conn到达
  - Conn到达 正常到达，开一个goroutine，调用handleRawConn 方法处理连接。
  - 出现error:
    - error 为暂时性的error：比如DeadlineExceededError，则阻塞一段时间后，进入下次循环
    - error 不是暂时性的，若此时刚好发生退出事件，直接返回nil，否则返回error

handleRawConn方法如下。

```go
// handleRawConn forks a goroutine to handle a just-accepted connection that
// has not had any I/O performed on it yet.
func (s *Server) handleRawConn(lisAddr string, rawConn net.Conn) {
   // 发生退出事件，关闭连接直接返回。
   if s.quit.HasFired() {
      rawConn.Close()
      return
   }
   // 设置连接的超时时间
   rawConn.SetDeadline(time.Now().Add(s.opts.connectionTimeout))

   // Finish handshaking (HTTP2)
   // HTTP2 握手，
   // net.Conn被转换为transport.ServerTransport
   st := s.newHTTP2Transport(rawConn)
   rawConn.SetDeadline(time.Time{})
   if st == nil {
      return
   }

   // 添加ServerTransport
   if !s.addConn(lisAddr, st) {
      return
   }
   go func() {
      // 处理Streams
      s.serveStreams(st)
      // 移除ServerTransport
      s.removeConn(lisAddr, st)
   }()
}
```



#### Stop  方法

Stop  提供Sever的退出。

下面看下Stop方法，Stop方法还是比较简单容易看懂，就是关闭一些资源。

```go
// Stop stops the gRPC server. It immediately closes all open
// connections and listeners.
// It cancels all active RPCs on the server side and the corresponding
// pending RPCs on the client side will get notified by connection
// errors.
func (s *Server) Stop() {
	// 开启退出事件
	s.quit.Fire()

	defer func() {
		// 等待其他serve事件结束
		s.serveWG.Wait()
		// 完成done事件
		s.done.Fire()
	}()

	//移除 channelz 调试信息
	s.channelzRemoveOnce.Do(func() { channelz.RemoveEntry(s.channelzID) })

	// 上锁
	s.mu.Lock()
	// s.lis 和s.conns 置空
	listeners := s.lis
	s.lis = nil
	conns := s.conns
	s.conns = nil
	// interrupt GracefulStop if Stop and GracefulStop are called concurrently.
	// 如果同时调用 Stop 和 GracefulStop，则中断 GracefulStop。
	s.cv.Broadcast()
	// 释放锁
	s.mu.Unlock()

	// 关闭所有 listeners 和 conns
	for lis := range listeners {
		lis.Close()
	}
	for _, cs := range conns {
		for st := range cs {
			st.Close()
		}
	}
	// 关闭 Server Worker
	if s.opts.numServerWorkers > 0 {
		s.stopServerWorkers()
	}

	
	s.mu.Lock()
    // 完成所有事件
	if s.events != nil {
		s.events.Finish()
		s.events = nil
	}
	s.mu.Unlock()
}
```

#### GracefulStop 优雅退出

要搞明白Server 如何实现的优雅退出需要把几个地方联系起来。

1. ServerTransport的Drain方法

   Drain方法 通知客户端 此 ServerTransport 停止接受新的 RPC请求。
   
   ```go
   	// Drain notifies the client this ServerTransport stops accepting new RPCs.
   	Drain()
   ```
   
   具体看下http2Server的实现：
   
   ```go
   func (t *http2Server) Drain() {
      t.mu.Lock()
      defer t.mu.Unlock()
      if t.drainChan != nil {
         return
      }
      t.drainChan = make(chan struct{})
      t.controlBuf.put(&goAway{code: http2.ErrCodeNo, debugData: []byte{}, headsUp: true})
   }
   ```
   
   server会发送`GOAWAY`帧，停止接受新的流。
   
   > `GOAWAY`帧（类型= 0x7）用于**启动连接关闭或发出严重错误状态信号**。 GOAWAY允许端点正常停止接受新的流，同时仍然完成对先前建立的流的处理。这可以实现管理操作，例如服务器维护。
   >
   > 
   >
   > 引用地址：https://skyao.io/learning-http2/
   
   总上：**ServerTransport调用Drain 方法后，连接不会处理新的请求。**
   
1. handleRawConn方法中 调用serveStreams 方法 处理 ServerTransport后，会将ServerTransport从conns 中删除掉并唤醒 cv。

   ```go
   s.serveStreams(st)
   s.removeConn(lisAddr, st)
   ```
   
   ```go
   // conns contains all active server transports. It is a map keyed on a
   // listener address with the value being the set of active transports
   // belonging to that listener.
   // 
   conns    map[string]map[transport.ServerTransport]bool
   cv       *sync.Cond              // signaled when connections close for GracefulStop
   ```
   
   conns 包含所有活跃的ServerTransport。 它是key 为 listener address 的map，value是属于该listener的一组ServerTransport。
   
   GracefulStop 中通过 cv 等待，连接处理完毕。
   
   ```go
   func (s *Server) removeConn(addr string, st transport.ServerTransport) {
   	s.mu.Lock()
   	defer s.mu.Unlock()
   
   	conns := s.conns[addr]
   	if conns != nil {
   		delete(conns, st)
   		if len(conns) == 0 {
   			// If the last connection for this address is being removed, also
   			// remove the map entry corresponding to the address. This is used
   			// in GracefulStop() when waiting for all connections to be closed.
   			// 如果这个address的最后一个connection正在被移除，也移除该地址对应map。
   			// 这个map在 GracefulStop 等待所有连接关闭时时使用。
   			delete(s.conns, addr)
   		}
   		// 唤醒GracefulStop
   		s.cv.Broadcast()
   	}
   }
   ```
   
   当处理完请求后，ServerTransport会被从conns中删除，并且去唤醒所有等待cv 的goroutine
   
   **一句话，当所有请求处理完以后，conns将不会有如何ServerTransport，GracefulStop的goroutine会被唤醒**

明白上面2点就，就可以理解GracefulStop的实现了。

```go
// GracefulStop stops the gRPC server gracefully. It stops the server from
// accepting new connections and RPCs and blocks until all the pending RPCs are
// finished.
func (s *Server) GracefulStop() {
   // 开启退出事件（标识退出开始）
   s.quit.Fire()
   // 完成done事件
   defer s.done.Fire()

   // 移除 channelz 调试信息
   s.channelzRemoveOnce.Do(func() { channelz.RemoveEntry(s.channelzID) })
   // 上锁
   s.mu.Lock()
   if s.conns == nil {
      s.mu.Unlock()
      return
   }

   // 关闭所有监听器
   for lis := range s.lis {
      lis.Close()
   }
   s.lis = nil
   // s.drain是特殊标识，没有活跃ServerTransport为true
   if !s.drain {
       // 调用 Drain 停止接受所有的请求
      for _, conns := range s.conns {
         for st := range conns {
            st.Drain()
         }
      }
      s.drain = true
   }

   // Wait for serving threads to be ready to exit.  Only then can we be sure no
   // new conns will be created.
   // 等待服务线端程准备好退出。 只有这样，我们才能确定不会创建新的conns。
   s.mu.Unlock()
   // 等待Serve goroutines 结束
   s.serveWG.Wait()
   s.mu.Lock()

   // 等待连接关闭
   for len(s.conns) != 0 {
      s.cv.Wait()
   }
   // 关闭所有事件
   s.conns = nil
   if s.events != nil {
      s.events.Finish()
      s.events = nil
   }
   s.mu.Unlock()
}
```

核心步骤：

1. 关闭所有lis 监听器。

2. 调用 drain 停止接受所有的请求

3. 等待服务goroutines退出

4. 阻塞等待，conns 连接处理完所有请求（等待Broadcast 唤醒）

   

#### channelz包

channelz 的介绍：https://github.com/grpc/proposal/blob/master/A14-channelz.md

gRPC 以 RPC 服务的形式提供调试统计信息， [channelz](https://github.com/grpc/proposal/blob/master/A14-channelz.md) 是一个提供通道级调试信息的服务。

如何使用：

```go
import "google.golang.org/grpc/channelz/service"

s := grpc.NewServer()
service.RegisterChannelzServiceToServer(s)
```

这里也是注册一份服务：

```go
func RegisterChannelzServiceToServer(s grpc.ServiceRegistrar) {
   channelzgrpc.RegisterChannelzServer(s, newCZServer())
}
```

实现了下列方法：

```go
type ChannelzServer interface {
   // Gets all root channels (i.e. channels the application has directly
   // created). This does not include subchannels nor non-top level channels.
   GetTopChannels(context.Context, *GetTopChannelsRequest) (*GetTopChannelsResponse, error)
   // Gets all servers that exist in the process.
   GetServers(context.Context, *GetServersRequest) (*GetServersResponse, error)
   // Returns a single Server, or else a NOT_FOUND code.
   GetServer(context.Context, *GetServerRequest) (*GetServerResponse, error)
   // Gets all server sockets that exist in the process.
   GetServerSockets(context.Context, *GetServerSocketsRequest) (*GetServerSocketsResponse, error)
   // Returns a single Channel, or else a NOT_FOUND code.
   GetChannel(context.Context, *GetChannelRequest) (*GetChannelResponse, error)
   // Returns a single Subchannel, or else a NOT_FOUND code.
   GetSubchannel(context.Context, *GetSubchannelRequest) (*GetSubchannelResponse, error)
   // Returns a single Socket or else a NOT_FOUND code.
   GetSocket(context.Context, *GetSocketRequest) (*GetSocketResponse, error)
}
```

从方法名就可以看出，这里提供了Server的各种信息。

gdebug工具：

https://github.com/grpc/grpc-experiments/tree/master/gdebug

这个 repo 包含一个连接到远程 gRPC服务并使用本地Web 服务器`channelz`将数据显示为网页的工具。`golang`目标是提供一个可以显示所有 gRPC 调试页面的 CLI 工具。

深入Server源码中：

```go
ls.channelzID, err = channelz.RegisterListenSocket(ls, s.channelzID, lis.Addr().String())
```

```go
func RegisterListenSocket(s Socket, pid *Identifier, ref string) (*Identifier, error) {
   if pid == nil {
      return nil, errors.New("a ListenSocket's parent id cannot be 0")
   }
   id := idGen.genID()
   if !IsOn() {
      return newIdentifer(RefListenSocket, id, pid), nil
   }

   ls := &listenSocket{refName: ref, s: s, id: id, pid: pid.Int()}
   db.get().addListenSocket(id, ls, pid.Int())
   return newIdentifer(RefListenSocket, id, pid), nil
}
```

```go
func (c *channelMap) addListenSocket(id int64, ls *listenSocket, pid int64) {
   c.mu.Lock()
   ls.cm = c
   c.listenSockets[id] = ls
   c.findEntry(pid).addChild(id, ls)
   c.mu.Unlock()
}
```

```go
// channelMap is the storage data structure for channelz.
// Methods of channelMap can be divided in two two categories with respect to locking.
// 1. Methods acquire the global lock.
// 2. Methods that can only be called when global lock is held.
// A second type of method need always to be called inside a first type of method.
type channelMap struct {
   mu               sync.RWMutex
   topLevelChannels map[int64]struct{}
   servers          map[int64]*server
   channels         map[int64]*channel
   subChannels      map[int64]*subChannel
   listenSockets    map[int64]*listenSocket
   normalSockets    map[int64]*normalSocket
}
```

```go
db    dbWrapper
```

grpc-go 内部有个本地变量dbWrapper，内部持有 channelMap 存储着 channelz 所需要的数据机构。

更加深入的内容，可以阅读下面的文章：

https://github.com/grpc/proposal/blob/master/A14-channelz.md



#### ServerWorkers  工作协程池

ServerWorkers  server的工作协程池，可以并行处理请求，提高处理效率。

initServerWorkers 方法

```go
// 创建工作 goroutine 和channel来处理连接，减少花费在 runtime.morestack 上的时间。
// initServerWorkers creates worker goroutines and channels to process incoming
// connections to reduce the time spent overall on runtime.morestack.
func (s *Server) initServerWorkers() {
   s.serverWorkerChannels = make([]chan *serverWorkerData, s.opts.numServerWorkers)
   for i := uint32(0); i < s.opts.numServerWorkers; i++ {
      s.serverWorkerChannels[i] = make(chan *serverWorkerData)
      go s.serverWorker(s.serverWorkerChannels[i])
   }
}
```

serverWorker 方法

```go
// N requests, by spawning a new goroutine in its place, a worker can reset its
// stack so that large stacks don't live in memory forever. 2^16 should allow
// each goroutine stack to live for at least a few seconds in a typical
// workload (assuming a QPS of a few thousand requests/sec).
const serverWorkerResetThreshold = 1 << 16
```

serverWorkerResetThreshold 定义了必须重置堆栈的频率。

 每 N 个请求，通过在其位置生成一个新的 goroutine，worker 可以重置其堆栈**，以便大堆栈不会永远存在于内存中。** 2^16 应该允许每个 goroutine 堆栈在典型的情况下至少存活几秒钟工作负载（假设 QPS 为几千个）。

`serverWorkers blocks on a *transport.Stream channel forever and waits for data to be fed by serveStreams. This allows different requests to be processed by the same goroutine, removing the need for expensive stack re-allocations (see the runtime.morestack problem [1]). https://github.com/golang/go/issues/18138 `

serverWorkers 永远阻塞在 *transport.Stream 通道上，直到 serveStreams 提供数据。 这允许由同一个 goroutine 处理不同的请求，从而无需宝贵的堆栈重新分配（请参阅 runtime.morestack 问题 [1]）。

[1] https://github.com/golang/go/issues/18138

```go
// serverWorkers blocks on a *transport.Stream channel forever and waits for
// data to be fed by serveStreams. This allows different requests to be
// processed by the same goroutine, removing the need for expensive stack
// re-allocations (see the runtime.morestack problem [1]).
//
// [1] https://github.com/golang/go/issues/18138
func (s *Server) serverWorker(ch chan *serverWorkerData) {
   // To make sure all server workers don't reset at the same time, choose a
   // random number of iterations before resetting.
   // 为确保所有server workers 不会同时重置，在重置前选择随机的迭代次数。 
   threshold := serverWorkerResetThreshold + grpcrand.Intn(serverWorkerResetThreshold)
   for completed := 0; completed < threshold; completed++ {
      // 等待 transport.Stream 通道的数据 
      data, ok := <-ch
      if !ok {
         return
      }
      s.handleStream(data.st, data.stream, s.traceInfo(data.st, data.stream))
      data.wg.Done()
   }
   // 重启一下 server worker ，防止堆栈过大。
   go s.serverWorker(ch)
}
```

serveStreams 方法

```go
func (s *Server) serveStreams(st transport.ServerTransport) {
   defer st.Close()
   var wg sync.WaitGroup

   var roundRobinCounter uint32
   st.HandleStreams(func(stream *transport.Stream) {
      wg.Add(1)
       
      // 开启 ServerWorkers
      if s.opts.numServerWorkers > 0 {
         // 组装数据  
         data := &serverWorkerData{st: st, wg: &wg, stream: stream}
         select {
             // 轮训的方式，提交数据给 server worker处理
         case s.serverWorkerChannels[atomic.AddUint32(&roundRobinCounter, 1)%s.opts.numServerWorkers] <- data:
         default:
            // If all stream workers are busy, fallback to the default code path.
            // worker 繁忙（channel 队列慢），降级到 单独处理 
            go func() {
               s.handleStream(st, stream, s.traceInfo(st, stream))
               wg.Done()
            }()
         }
      } else {
         go func() {
            defer wg.Done()
            s.handleStream(st, stream, s.traceInfo(st, stream))
         }()
      }
   }, func(ctx context.Context, method string) context.Context {
      if !EnableTracing {
         return ctx
      }
      tr := trace.New("grpc.Recv."+methodFamily(method), method)
      return trace.NewContext(ctx, tr)
   })
   wg.Wait()
}
```



#### serveStreams 方法

```go
func (s *Server) serveStreams(st transport.ServerTransport) {
   defer st.Close()
   var wg sync.WaitGroup

   var roundRobinCounter uint32 // 轮训计数器 
   st.HandleStreams(func(stream *transport.Stream) {
      wg.Add(1)
       
      // 开启 ServerWorkers
      if s.opts.numServerWorkers > 0 {
         // 组装数据  
         data := &serverWorkerData{st: st, wg: &wg, stream: stream}
         select {
             // 轮训的方式，提交数据给 server worker处理
         case s.serverWorkerChannels[atomic.AddUint32(&roundRobinCounter, 1)%s.opts.numServerWorkers] <- data:
         default:
            // If all stream workers are busy, fallback to the default code path.
            // worker 繁忙（channel 队列已满），降级到单独处理 
            go func() {
               s.handleStream(st, stream, s.traceInfo(st, stream))
               wg.Done()
            }()
         }
      } else {
          // 沒有开启Server Worker，单独处理 
         go func() {
            defer wg.Done()
            s.handleStream(st, stream, s.traceInfo(st, stream))
         }()
      }
   }, func(ctx context.Context, method string) context.Context {
      if !EnableTracing {
         return ctx
      }
      tr := trace.New("grpc.Recv."+methodFamily(method), method)
      return trace.NewContext(ctx, tr)
   })
   wg.Wait()
}
```

handleStream方法

```go
func (s *Server) handleStream(t transport.ServerTransport, stream *transport.Stream, trInfo *traceInfo) {
	// 获取Method，合法Method格式为 server_name/method_name 例如：helloworld.Greeter/SayHello
	sm := stream.Method()
	if sm != "" && sm[0] == '/' {
		sm = sm[1:]
	}
	pos := strings.LastIndex(sm, "/")
	// Method格式有误，没有/
	if pos == -1 {
		if trInfo != nil {
			trInfo.tr.LazyLog(&fmtStringer{"Malformed method name %q", []interface{}{sm}}, true)
			trInfo.tr.SetError()
		}
		errDesc := fmt.Sprintf("malformed method name: %q", stream.Method())
		if err := t.WriteStatus(stream, status.New(codes.Unimplemented, errDesc)); err != nil {
			if trInfo != nil {
				trInfo.tr.LazyLog(&fmtStringer{"%v", []interface{}{err}}, true)
				trInfo.tr.SetError()
			}
			channelz.Warningf(logger, s.channelzID, "grpc: Server.handleStream failed to write status: %v", err)
		}
		if trInfo != nil {
			trInfo.tr.Finish()
		}
		return
	}
	service := sm[:pos]  // 获取服务
	method := sm[pos+1:] // 获取服务名

	srv, knownService := s.services[service]
	// 如果存对应的服务和方法
	if knownService {
		// 处理一元服务请求
		if md, ok := srv.methods[method]; ok {
			s.processUnaryRPC(t, stream, srv, md, trInfo)
			return
		}
		// 处理流式服务请求
		if sd, ok := srv.streams[method]; ok {
			s.processStreamingRPC(t, stream, srv, sd, trInfo)
			return
		}
	}
	// Unknown service, or known server unknown method.
	// 以下都是未知服务或者方法的错误处理
	
	// opts.unknownStreamDesc 若配置就使用这个错误处理
	if unknownDesc := s.opts.unknownStreamDesc; unknownDesc != nil {
		s.processStreamingRPC(t, stream, nil, unknownDesc, trInfo)
		return
	}
	var errDesc string
	if !knownService {
		// 未知服务
		errDesc = fmt.Sprintf("unknown service %v", service)
	} else {
		// 未知方法
		errDesc = fmt.Sprintf("unknown method %v for service %v", method, service)
	}
	// traceInfo 记录错误
	if trInfo != nil {
		trInfo.tr.LazyPrintf("%s", errDesc)
		trInfo.tr.SetError()
	}
	if err := t.WriteStatus(stream, status.New(codes.Unimplemented, errDesc)); err != nil {
		if trInfo != nil {
			trInfo.tr.LazyLog(&fmtStringer{"%v", []interface{}{err}}, true)
			trInfo.tr.SetError()
		}
		channelz.Warningf(logger, s.channelzID, "grpc: Server.handleStream failed to write status: %v", err)
	}
	if trInfo != nil {
		trInfo.tr.Finish()
	}
}
```

