# grpc-go 源码client浅读

上文浅读了一下Server端的核心代码，现在浅读client源码，学习一下。

让我们回到helloworld这个例子中。

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

client的核心流程可以分为：

1. 建立连接 grpc.ClientConn
2. 创建客户端Client
3. 发起Rpc调用
4. 关闭连接

## 建立连接 grpc.ClientConn

先认识一下连接参数dialOptions

```go
// dialOptions configure a Dial call. dialOptions are set by the DialOption
// values passed to Dial.
type dialOptions struct {
   unaryInt  UnaryClientInterceptor
   streamInt StreamClientInterceptor

   chainUnaryInts  []UnaryClientInterceptor
   chainStreamInts []StreamClientInterceptor

   cp                          Compressor
   dc                          Decompressor
   bs                          internalbackoff.Strategy
   block                       bool
   returnLastError             bool
   timeout                     time.Duration
   scChan                      <-chan ServiceConfig
   authority                   string
   copts                       transport.ConnectOptions
   callOptions                 []CallOption
   channelzParentID            *channelz.Identifier
   disableServiceConfig        bool
   disableRetry                bool
   disableHealthCheck          bool
   healthCheckFunc             internal.HealthChecker
   minConnectTimeout           func() time.Duration
   defaultServiceConfig        *ServiceConfig // defaultServiceConfig is parsed from defaultServiceConfigRawJSON.
   defaultServiceConfigRawJSON *string
   resolvers                   []resolver.Builder
}
```



下面看建立连接的详细过程：

```go
// Set up a connection to the server.
conn, err := grpc.Dial(*addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
if err != nil {
   log.Fatalf("did not connect: %v", err)
}
```

配置加载在之前的文章已经解读过，这里不赘述。

`grpc.WithTransportCredentials(insecure.NewCredentials())`

创建了一个安全凭证，http2 握手时使用。

```go
// Dial creates a client connection to the given target.
func Dial(target string, opts ...DialOption) (*ClientConn, error) {
   return DialContext(context.Background(), target, opts...)
}
```

DialContext 创建ClientConn客户端连接去连接目标地址。 

默认情况下，它是一个非阻塞Dial（拨号）（该函数不会等待建立连接，并且连接发生在后台）。 要使其成为阻塞Dial，请使用 WithBlock方法 Dial。

在非阻塞情况下，ctx 不会响应连接。 它仅控制设置步骤。

在阻塞情况下，ctx 可用于取消或使得挂起的连接过期。 一旦此函数返回，ctx 的取消函数和过期将无操作。 

DialContext函数返回后，用户应调用 ClientConn.Close 方法以终止所有挂起的操作。

target name 在 https://github.com/grpc/grpc/blob/master/doc/naming.md 中定义。 

例如要使用 dns 解析器，使用“dns:///”前缀。



```go
// DialContext creates a client connection to the given target. By default, it's
// a non-blocking dial (the function won't wait for connections to be
// established, and connecting happens in the background). To make it a blocking
// dial, use WithBlock() dial option.
//
// In the non-blocking case, the ctx does not act against the connection. It
// only controls the setup steps.
//
// In the blocking case, ctx can be used to cancel or expire the pending
// connection. Once this function returns, the cancellation and expiration of
// ctx will be noop. Users should call ClientConn.Close to terminate all the
// pending operations after this function returns.
//
// The target name syntax is defined in
// https://github.com/grpc/grpc/blob/master/doc/naming.md.
// e.g. to use dns resolver, a "dns:///" prefix should be applied to the target.
func DialContext(ctx context.Context, target string, opts ...DialOption) (conn *ClientConn, err error) {
   cc := &ClientConn{
      target:            target,
      csMgr:             &connectivityStateManager{},
      conns:             make(map[*addrConn]struct{}),
      dopts:             defaultDialOptions(),
      blockingpicker:    newPickerWrapper(),
      czData:            new(channelzData),
      firstResolveEvent: grpcsync.NewEvent(),
   }
   cc.retryThrottler.Store((*retryThrottler)(nil))
   cc.safeConfigSelector.UpdateConfigSelector(&defaultConfigSelector{nil})
   cc.ctx, cc.cancel = context.WithCancel(context.Background())

   for _, opt := range extraDialOptions {
      opt.apply(&cc.dopts)
   }

   for _, opt := range opts {
      opt.apply(&cc.dopts)
   }

   chainUnaryClientInterceptors(cc)
   chainStreamClientInterceptors(cc)

   defer func() {
      if err != nil {
         cc.Close()
      }
   }()

   pid := cc.dopts.channelzParentID
   cc.channelzID = channelz.RegisterChannel(&channelzChannel{cc}, pid, target)
   ted := &channelz.TraceEventDesc{
      Desc:     "Channel created",
      Severity: channelz.CtInfo,
   }
   if cc.dopts.channelzParentID != nil {
      ted.Parent = &channelz.TraceEventDesc{
         Desc:     fmt.Sprintf("Nested Channel(id:%d) created", cc.channelzID.Int()),
         Severity: channelz.CtInfo,
      }
   }
   channelz.AddTraceEvent(logger, cc.channelzID, 1, ted)
   cc.csMgr.channelzID = cc.channelzID

   if cc.dopts.copts.TransportCredentials == nil && cc.dopts.copts.CredsBundle == nil {
      return nil, errNoTransportSecurity
   }
   if cc.dopts.copts.TransportCredentials != nil && cc.dopts.copts.CredsBundle != nil {
      return nil, errTransportCredsAndBundle
   }
   if cc.dopts.copts.CredsBundle != nil && cc.dopts.copts.CredsBundle.TransportCredentials() == nil {
      return nil, errNoTransportCredsInBundle
   }
   transportCreds := cc.dopts.copts.TransportCredentials
   if transportCreds == nil {
      transportCreds = cc.dopts.copts.CredsBundle.TransportCredentials()
   }
   if transportCreds.Info().SecurityProtocol == "insecure" {
      for _, cd := range cc.dopts.copts.PerRPCCredentials {
         if cd.RequireTransportSecurity() {
            return nil, errTransportCredentialsMissing
         }
      }
   }

   if cc.dopts.defaultServiceConfigRawJSON != nil {
      scpr := parseServiceConfig(*cc.dopts.defaultServiceConfigRawJSON)
      if scpr.Err != nil {
         return nil, fmt.Errorf("%s: %v", invalidDefaultServiceConfigErrPrefix, scpr.Err)
      }
      cc.dopts.defaultServiceConfig, _ = scpr.Config.(*ServiceConfig)
   }
   cc.mkp = cc.dopts.copts.KeepaliveParams

   if cc.dopts.copts.UserAgent != "" {
      cc.dopts.copts.UserAgent += " " + grpcUA
   } else {
      cc.dopts.copts.UserAgent = grpcUA
   }

   if cc.dopts.timeout > 0 {
      var cancel context.CancelFunc
      ctx, cancel = context.WithTimeout(ctx, cc.dopts.timeout)
      defer cancel()
   }
   defer func() {
      select {
      case <-ctx.Done():
         switch {
         case ctx.Err() == err:
            conn = nil
         case err == nil || !cc.dopts.returnLastError:
            conn, err = nil, ctx.Err()
         default:
            conn, err = nil, fmt.Errorf("%v: %v", ctx.Err(), err)
         }
      default:
      }
   }()

   scSet := false
   if cc.dopts.scChan != nil {
      // Try to get an initial service config.
      select {
      case sc, ok := <-cc.dopts.scChan:
         if ok {
            cc.sc = &sc
            cc.safeConfigSelector.UpdateConfigSelector(&defaultConfigSelector{&sc})
            scSet = true
         }
      default:
      }
   }
   if cc.dopts.bs == nil {
      cc.dopts.bs = backoff.DefaultExponential
   }

   // Determine the resolver to use.
   resolverBuilder, err := cc.parseTargetAndFindResolver()
   if err != nil {
      return nil, err
   }
   cc.authority, err = determineAuthority(cc.parsedTarget.Endpoint, cc.target, cc.dopts)
   if err != nil {
      return nil, err
   }
   channelz.Infof(logger, cc.channelzID, "Channel authority set to %q", cc.authority)

   if cc.dopts.scChan != nil && !scSet {
      // Blocking wait for the initial service config.
      select {
      case sc, ok := <-cc.dopts.scChan:
         if ok {
            cc.sc = &sc
            cc.safeConfigSelector.UpdateConfigSelector(&defaultConfigSelector{&sc})
         }
      case <-ctx.Done():
         return nil, ctx.Err()
      }
   }
   if cc.dopts.scChan != nil {
      go cc.scWatcher()
   }

   var credsClone credentials.TransportCredentials
   if creds := cc.dopts.copts.TransportCredentials; creds != nil {
      credsClone = creds.Clone()
   }
   cc.balancerWrapper = newCCBalancerWrapper(cc, balancer.BuildOptions{
      DialCreds:        credsClone,
      CredsBundle:      cc.dopts.copts.CredsBundle,
      Dialer:           cc.dopts.copts.Dialer,
      Authority:        cc.authority,
      CustomUserAgent:  cc.dopts.copts.UserAgent,
      ChannelzParentID: cc.channelzID,
      Target:           cc.parsedTarget,
   })

   // Build the resolver.
   rWrapper, err := newCCResolverWrapper(cc, resolverBuilder)
   if err != nil {
      return nil, fmt.Errorf("failed to build resolver: %v", err)
   }
   cc.mu.Lock()
   cc.resolverWrapper = rWrapper
   cc.mu.Unlock()

   // A blocking dial blocks until the clientConn is ready.
   if cc.dopts.block {
      for {
         cc.Connect()
         s := cc.GetState()
         if s == connectivity.Ready {
            break
         } else if cc.dopts.copts.FailOnNonTempDialError && s == connectivity.TransientFailure {
            if err = cc.connectionError(); err != nil {
               terr, ok := err.(interface {
                  Temporary() bool
               })
               if ok && !terr.Temporary() {
                  return nil, err
               }
            }
         }
         if !cc.WaitForStateChange(ctx, s) {
            // ctx got timeout or canceled.
            if err = cc.connectionError(); err != nil && cc.dopts.returnLastError {
               return nil, err
            }
            return nil, ctx.Err()
         }
      }
   }

   return cc, nil
}
```