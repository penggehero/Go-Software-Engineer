# net 包

`Package net provides a portable interface for network I/O, including TCP/IP, UDP, domain name resolution, and Unix domain sockets.`

net包为包括TCP/IP、UDP、域名解析和Unix 域套接字网络IO提供了简便的接口

`Although the package provides access to low-level networking primitives, most clients will need only the basic interface provided by the Dial, Listen, and Accept functions and the associated Conn and Listener interfaces. The crypto/tls package uses the same interfaces and similar Dial and Listen functions.`

尽管net包提供了访问底层网络的原语，但大多数客户端只需要 Dial、Listen 和 Accept 函数以及相关的 Conn 和 Listener 接口提供的基本接口。crypto/tls 包使用相同的接口和类似的 Dial 和 Listen 功能。



net包的Go语言是访问网络基础包，下面学习一下几个方法。

## Listen方法

### type Listener

```go
type Listener interface {
	// Accept waits for and returns the next connection to the listener.
	// Accept 等待并将下一个连接返回给侦听器。
	Accept() (Conn, error)

	// Close closes the listener.Close 关闭监听器。
	// Any blocked Accept operations will be unblocked and return errors.  任何被阻塞的 Accept 操作都将被解除阻塞并返回错误。
	Close() error

	// Addr returns the listener's network address.
    // Addr 返回监听器的网络地址。
	Addr() Addr
}
```

`A Listener is a generic network listener for stream-oriented protocols.`

Listener是面向流协议的通用网络侦听器。 

`Multiple goroutines may invoke methods on a Listener simultaneously.`

多个 goroutine 可以同时调用 Listener 上的方法。

一个简单的例子：

```go
package main

import (
	"io"
	"log"
	"net"
)

func main() {
	// Listen on TCP port 2000 on all available unicast and
	// anycast IP addresses of the local system.
	l, err := net.Listen("tcp", ":2000")
	if err != nil {
		log.Fatal(err)
	}
	defer l.Close()
	for {
		// Wait for a connection.
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}
		// Handle the connection in a new goroutine.
		// The loop then returns to accepting, so that
		// multiple connections may be served concurrently.
		go func(c net.Conn) {
			// Echo all incoming data.
			io.Copy(c, c)
			// Shut down the connection.
			c.Close()
		}(conn)
	}
}

```

### func Listen

ListeListen announces on the local network address.

The network must be "tcp", "tcp4", "tcp6", "unix" or "unixpacket".For TCP networks, if the host in the address parameter is empty or a literal unspecified IP address, Listen listens on all available unicast and anycast IP addresses of the local system. To only use IPv4, use network "tcp4". The address can use a host name, but this is not recommended, because it will create a listener for at most one of the host's IP addresses. If the port in the address parameter is empty or "0", as in "127.0.0.1:" or "[::1]:0", a port number is automatically chosen. The Addr method of Listener can be used to discover the chosen port.

```go
func Listen(network, address string) (Listener, error) {
    // 创建一个默认的ListenConfig 配置对象
	var lc ListenConfig
	return lc.Listen(context.Background(), network, address)
}
```



```go
func (lc *ListenConfig) Listen(ctx context.Context, network, address string) (Listener, error) {
    // 解析地址 address
	addrs, err := DefaultResolver.resolveAddrList(ctx, "listen", network, address, nil)
	if err != nil {
		return nil, &OpError{Op: "listen", Net: network, Source: nil, Addr: nil, Err: err}
	}
    // 创建一个sysListener对象，包含了Listen的参数和配置
	sl := &sysListener{
		ListenConfig: *lc,
		network:      network,
		address:      address,
	}
    // 创建一个Listen对象
	var l Listener
    // 返回第一个Ipv4地址
	la := addrs.first(isIPv4)
	switch la := la.(type) {
        // TCP地址
	case *TCPAddr:
		l, err = sl.listenTCP(ctx, la)
        // Unix地址
	case *UnixAddr:
		l, err = sl.listenUnix(ctx, la)
	default:
        // 其他类型直接返回错误
		return nil, &OpError{Op: "listen", Net: sl.network, Source: nil, Addr: la, Err: &AddrError{Err: "unexpected address type", Addr: address}}
	}
	if err != nil {
		return nil, &OpError{Op: "listen", Net: sl.network, Source: nil, Addr: la, Err: err} // l is non-nil interface containing nil pointer
	}
    // 返回Listener 对象
	return l, nil
}
```

```go
func (sl *sysListener) listenTCP(ctx context.Context, laddr *TCPAddr) (*TCPListener, error) {
	// 调用内部的Socket方法，获取fd
    fd, err := internetSocket(ctx, sl.network, laddr, nil, syscall.SOCK_STREAM, 0, "listen", sl.ListenConfig.Control)
	if err != nil {
		return nil, err
	}
    // 创建TCPListener对象
	return &TCPListener{fd: fd, lc: sl.ListenConfig}, nil
}
```

> FD
>
> **文件描述符**（File descriptor）是计算机科学中的一个术语，是一个用于表述指向[文件](https://zh.wikipedia.org/wiki/文件)的引用的抽象化概念。
>
> 文件描述符在形式上是一个非负整数。实际上，它是一个索引值，指向[内核](https://zh.wikipedia.org/wiki/内核)为每一个[进程](https://zh.wikipedia.org/wiki/进程)所维护的该进程打开文件的记录表。当程序打开一个现有文件或者创建一个新文件时，内核向进程返回一个文件描述符。在[程序设计](https://zh.wikipedia.org/wiki/程序设计)中，一些涉及底层的程序编写往往会围绕着文件描述符展开。但是文件描述符这一概念往往只适用于[UNIX](https://zh.wikipedia.org/wiki/UNIX)、[Linux](https://zh.wikipedia.org/wiki/Linux)这样的操作系统。
>

![fD](https://upload.wikimedia.org/wikipedia/commons/f/f8/File_table_and_inode_table.svg)

```go
func internetSocket(ctx context.Context, net string, laddr, raddr sockaddr, sotype, proto int, mode string, ctrlFn func(string, string, syscall.RawConn) error) (fd *netFD, err error) {
   if (runtime.GOOS == "aix" || runtime.GOOS == "windows" || runtime.GOOS == "openbsd") && mode == "dial" && raddr.isWildcard() {
      raddr = raddr.toLocal(net)
   }
   // 获取合适的地址族
   family, ipv6only := favoriteAddrFamily(net, laddr, raddr, mode)
   return socket(ctx, net, family, sotype, proto, ipv6only, laddr, raddr, ctrlFn)
}
```



```go
func socket(ctx context.Context, net string, family, sotype, proto int, ipv6only bool, laddr, raddr sockaddr, ctrlFn func(string, string, syscall.RawConn) error) (fd *netFD, err error) {
   s, err := sysSocket(family, sotype, proto)
   if err != nil {
      return nil, err
   }
   if err = setDefaultSockopts(s, family, sotype, ipv6only); err != nil {
      poll.CloseFunc(s)
      return nil, err
   }
   if fd, err = newFD(s, family, sotype, net); err != nil {
      poll.CloseFunc(s)
      return nil, err
   }

   // This function makes a network file descriptor for the
   // following applications:
   //
   // - An endpoint holder that opens a passive stream
   //   connection, known as a stream listener
   //
   // - An endpoint holder that opens a destination-unspecific
   //   datagram connection, known as a datagram listener
   //
   // - An endpoint holder that opens an active stream or a
   //   destination-specific datagram connection, known as a
   //   dialer
   //
   // - An endpoint holder that opens the other connection, such
   //   as talking to the protocol stack inside the kernel
   //
   // For stream and datagram listeners, they will only require
   // named sockets, so we can assume that it's just a request
   // from stream or datagram listeners when laddr is not nil but
   // raddr is nil. Otherwise we assume it's just for dialers or
   // the other connection holders.

   if laddr != nil && raddr == nil {
      switch sotype {
      case syscall.SOCK_STREAM, syscall.SOCK_SEQPACKET:
         if err := fd.listenStream(laddr, listenerBacklog(), ctrlFn); err != nil {
            fd.Close()
            return nil, err
         }
         return fd, nil
      case syscall.SOCK_DGRAM:
         if err := fd.listenDatagram(laddr, ctrlFn); err != nil {
            fd.Close()
            return nil, err
         }
         return fd, nil
      }
   }
   if err := fd.dial(ctx, laddr, raddr, ctrlFn); err != nil {
      fd.Close()
      return nil, err
   }
   return fd, nil
}
```

 深入 sysSocket

```go
// Wrapper around the socket system call that marks the returned file
// descriptor as nonblocking and close-on-exec.
func sysSocket(family, sotype, proto int) (int, error) {
    // 发起Socket 系统调用
	s, err := socketFunc(family, sotype|syscall.SOCK_NONBLOCK|syscall.SOCK_CLOEXEC, proto)
	// On Linux the SOCK_NONBLOCK and SOCK_CLOEXEC flags were
	// introduced in 2.6.27 kernel and on FreeBSD both flags were
	// introduced in 10 kernel. If we get an EINVAL error on Linux
	// or EPROTONOSUPPORT error on FreeBSD, fall back to using
	// socket without them.
	/* 在 Linux 上，SOCK_NONBLOCK 和 SOCK_CLOEXEC 标志是在 2.6.27 内核中引入的，而在 FreeBSD 上，这两个标志都是在 10 内核中引入的。 如果我们在 Linux 上收到 EINVAL 错误或在 FreeBSD 上收到 EPROTONOSUPPORT 错误，请回退到在没有它们的情况下使用套接字。*/
	switch err {
	case nil:
		return s, nil
	default:
		return -1, os.NewSyscallError("socket", err)
	case syscall.EPROTONOSUPPORT, syscall.EINVAL:
	}

	// See ../syscall/exec_unix.go for description of ForkLock.
	syscall.ForkLock.RLock()
	s, err = socketFunc(family, sotype, proto)
	if err == nil {
		syscall.CloseOnExec(s)
	}
	syscall.ForkLock.RUnlock()
	if err != nil {
		return -1, os.NewSyscallError("socket", err)
	}
	if err = syscall.SetNonblock(s, true); err != nil {
		poll.CloseFunc(s)
		return -1, os.NewSyscallError("setnonblock", err)
	}
	return s, nil
}
```

syscall.Socket 系统调用

```go
var (
   testHookDialChannel  = func() {} // for golang.org/issue/5349
   testHookCanceledDial = func() {} // for golang.org/issue/16523

   // Placeholders for socket system calls.
   socketFunc        func(int, int, int) (int, error)  = syscall.Socket // Socket 调用函数
   connectFunc       func(int, syscall.Sockaddr) error = syscall.Connect 
   listenFunc        func(int, int) error              = syscall.Listen
   getsockoptIntFunc func(int, int, int) (int, error)  = syscall.GetsockoptInt
)
```

```go
func Socket(domain, typ, proto int) (fd int, err error) {
   if domain == AF_INET6 && SocketDisableIPv6 {
      return -1, EAFNOSUPPORT
   }
   fd, err = socket(domain, typ, proto)
   return
}
```



```go
func socket(domain int, typ int, proto int) (fd int, err error) {
   r0, _, e1 := RawSyscall(SYS_SOCKET, uintptr(domain), uintptr(typ), uintptr(proto))
   fd = int(r0)
   if e1 != 0 {
      err = errnoErr(e1)
   }
   return
}
```

```go
func RawSyscall(trap, a1, a2, a3 uintptr) (r1, r2 uintptr, err Errno)
```

