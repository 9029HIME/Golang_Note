 

# 前提

​	本笔记基于Golang v1.14

​	Golang的网络IO，底层是基于epoll实现的（Linux2.6以上），语言层面通过netpoll管理goroutine操作fd时的阻塞与唤醒。另外，netpoll还自定义了一个pd结构体（type pollDesc struct）用来直接保存goroutine与操作的fd信息，**用来实现内核态与用户态信息共享**。

​	runtime作为协程调度的模块，当G对fd进行操作发生阻塞时，将G挂起。当epoll()发现该fd的事件就绪后，可以通过特定手段查找到该fd的G并激活，使其重新回到P的local queue里。在用户态（G）是阻塞的，在内核态（epoll_wait）是在不断监听的，这样就能实现epoll与Goroutine的调度。（待润色）

# 先看一下netpoll有哪些关键结构体

## netFD

```go
type netFD struct {
   pfd poll.FD

   // immutable until Close
   family      int
   sotype      int
   isConnected bool // handshake completed or use of association with peer
   net         string
   laddr       Addr
   raddr       Addr
}
```

​	不同的操作系统有不同的netFD与poll.FD，在netFD里最关键的结构体是poll.FD。**这里都以Unix为例**

## FD

```go
type FD struct {
   // 多个协程对FD进行操作的互斥锁
   fdmu fdMutex

   // 系统返回的fd，即fd本身
   Sysfd int

   // 最重要的结构体，里面记录了fd与goroutine之间的关系
   pd pollDesc

   // Writev cache.
   iovecs *[]syscall.Iovec

   // Semaphore signaled when file is closed.
   csema uint32

   // Non-zero if this file has been set to blocking mode.
   isBlocking uint32

   // Whether this is a streaming descriptor, as opposed to a
   // packet-based descriptor like a UDP socket. Immutable.
   IsStream bool

   // Whether a zero byte read indicates EOF. This is false for a
   // message based socket connection.
   ZeroReadIsEOF bool

   // Whether this is a file rather than a network socket.
   isFile bool
}
```

​	在Golang中，每一个连接都会被抽象成一个FD。

## pollDesc

```go
type pollDesc struct {
   link *pollDesc // in pollcache, protected by pollcache.lock

   // The lock protects pollOpen, pollSetDeadline, pollUnblock and deadlineimpl operations.
   // This fully covers seq, rt and wt variables. fd is constant throughout the PollDesc lifetime.
   // pollReset, pollWait, pollWaitCanceled and runtime·netpollready (IO readiness notification)
   // proceed w/o taking the lock. So closing, everr, rg, rd, wg and wd are manipulated
   // in a lock-free way by all operations.
   // NOTE(dvyukov): the following code uses uintptr to store *g (rg/wg),
   // that will blow up when GC starts moving objects.
   lock    mutex // 多个协程对pd操作时的互斥锁
   fd      uintptr	// 指向fd的指针
   closing bool
   everr   bool    // marks event scanning error happened
   user    uint32  // user settable cookie
   rseq    uintptr // protects from stale read timers
   rg      uintptr // 对pd进行读操作并阻塞的协程地址，若epoll_wait发现rdllist有就绪ep_item后，会通过rg找到阻塞的协程并唤醒
   rt      timer   // 防止读超时的定时器
   rd      int64   // read deadline
   wseq    uintptr // protects from stale write timers
   wg      uintptr // 对pd进行写操作的协程地址，和wg功能相似
   wt      timer   // 防止写超时的定时器
   wd      int64   // write deadline
}
```

​	**pd会与内核态里负责epoll()的线程进行通信，当epoll()事件就绪时就可以通知goroutine唤醒**。wg和rg默认值为0，且有两个const：pdReady，pdWait。前者表示pd已经就绪（事件已触发），后者表示pd还在阻塞中（这里有个疑问，如何同时表示pdWait和对应的goroutine？）。



# netpoll流程

​	用一个很普通的listen代码来说明

```go
func main() {
	lister, err := net.Listen("tcp", "localhost:8080") //这里不会阻塞
	if err != nil {
		fmt.Println("监听失败", err)
		return
	}
	for {
		conn, err := lister.Accept() 	//这里会阻塞
		if err != nil {
			fmt.Println("本次连接失败", err)
			continue
		}
        
		go func() {
			defer conn.Close()
			for {
				buf := make([]byte, 128)
				n, err := conn.Read(buf)	//这里会阻塞
				if err != nil {
					fmt.Println("读出错", err)
					return
				}
				fmt.Println("读取到的数据：", string(buf[:n]))
			}
		}()
	}
}
```
## Listen()过程

​	Listen()的调用过程是：ListenTCP --> listenTCP --> internetSocket --> socket --> listenStream，前面的都是建立socket等过程，主要是对监听的设置做初始化操作，这里直接看listenStream里的初始化操作。

```go
/**
	注意，在调用listenStream前，已经通过newFD()新建了一个netFD对象
**/
func (fd *netFD) listenStream(laddr sockaddr, backlog int, ctrlFn func(string, string, syscall.RawConn) error) error {
   var err error
   if err = setDefaultListenerSockopts(fd.pfd.Sysfd); err != nil {
      return err
   }
   var lsa syscall.Sockaddr
    // 这里是返回一个syscall.Sockaddr接口的实现对象，用lsa接收，有以下类型：syscall.SockaddrInet4、syscall.SockaddrInet4、syscall.SockaddrUnix
   if lsa, err = laddr.sockaddr(fd.family); err != nil {
      return err
   }
    // ctrlFn()是用来设置socket选项，但从listenTCP()里传了nil，调用到这里也是nil。
   if ctrlFn != nil {
      c, err := newRawConn(fd)
      if err != nil {
         return err
      }
      if err := ctrlFn(fd.ctrlNetwork(), laddr.String(), c); err != nil {
         return err
      }
   }
    // 为socket绑定地址，可以理解为将Sysfd(系统返回的fd)与lsa绑定起来，应该是建立了双向关系
   if err = syscall.Bind(fd.pfd.Sysfd, lsa); err != nil {
      return os.NewSyscallError("bind", err)
   }
    // 监听socket，这里其实调用了系统函数	SYS_LISTEN,第二个参数指的是监听队列大小
   if err = listenFunc(fd.pfd.Sysfd, backlog); err != nil {
      return os.NewSyscallError("listen", err)
   }
    // init()是对netFD对象进行初始化操作，包含epoll的初始化(epoll_create)与epoll的注册事件(epoll_ctl)与 TODO 等待(epoll_wait)
   if err = fd.init(); err != nil {
      return err
   }
    // 获取socket信息，这里其实调用了系统函数getsockname，socket信息会添加到fd的laddr里。
   lsa, _ = syscall.Getsockname(fd.pfd.Sysfd)
   fd.setAddr(fd.addrFunc()(lsa), nil)
   return nil
}
```

​	listenStream()方法里最核心的部分是fd.init()，该方法会调用

```go
func (fd *netFD) init() error { //第一个参数目测是listen传入的"tcp"，第二个参数置为true，表示使用poll机制
    //netFD的pdf指的是FD
    return fd.pfd.Init(fd.net, true)
}
```

​		↓

```go
func (fd *FD) Init(net string, pollable bool) error {
    // We don't actually care about the various network types.
    if net == "file" {
        fd.isFile = true
    }    //如果不使用poll机制，则fd置为blocking mode并返回。当net=="file"时不会使用poll
    if !pollable {
        fd.isBlocking = 1
        return nil
    }    //FD的pd指的是pollDesc，在上层方法socket()中调用newFD()初始化netFD时，应该也初始化的netFD.pfd.pd，所以pd才不为nil
    err := fd.pd.init(fd)
    if err != nil {
        // If we could not initialize the runtime poller,
        // assume we are using blocking mode.
        fd.isBlocking = 1
    }
    return err
}
```

​	顺便说一下，Golang里定义了一些内置函数，用来直接调用epoll的相关操作。

```go
 func runtime_pollServerInit()
 func runtime_pollOpen(fd uintptr) (uintptr, int)
 func runtime_pollClose(ctx uintptr)
 func runtime_pollWait(ctx uintptr, mode int) int
 func runtime_pollWaitCanceled(ctx uintptr, mode int) int
 func runtime_pollReset(ctx uintptr, mode int) int
 func runtime_pollSetDeadline(ctx uintptr, d int64, mode int)
 func runtime_pollUnblock(ctx uintptr)
 func runtime_isPollServerDescriptor(fd uintptr) bool
```

​		↓

```go
func (pd *pollDesc) init(fd *FD) error {
    // runtime_pollServerInit()的作用是创建ep_fd，相当于调用epoll_create()，serverInit.Do()用来保证只调用一个epoll_create()
   serverInit.Do(runtime_pollServerInit)
    /* runtime_pollOpen的作用是将Sysfd注册到ep_fd里，相当于调用epoll_ctl()，同时返回的ctx是fd的pollDesc对象。
    TODO 这里有两个疑问：
    1.是如何注册到上面创建的ep_fd里的
    2.是不是默认注册可读事件了
    答：1.从func netpollopen(fd uintptr, pd *pollDesc) int32可以得知，最终是向netpoll_epoll.go里的一个全局变量epfd注册事件，runtime_pollServerInit应该是修改了这个全局变量的值
    	2.注册的是_EPOLLIN | _EPOLLOUT | _EPOLLRDHUP | _EPOLLET 事件
   */
   ctx, errno := runtime_pollOpen(uintptr(fd.Sysfd))
    /*
     如果注册失败了，会做以下操作
     	解锁阻塞的goroutine，TODO 但是listen什么时候阻塞了goroutine？答：这里在listen里确实不会阻塞
     	回收pollDesc对象
     	关闭fd，在层层返回error后，最后会在socket()里关闭
    */
   if errno != 0 {
      if ctx != 0 {
         runtime_pollUnblock(ctx)
         runtime_pollClose(ctx)
      }
      return errnoErr(syscall.Errno(errno))
   }
    // TODO 这个runtimeCtx哪来的？
   pd.runtimeCtx = ctx
   return nil
}
```

​	先看一下epoll_create()，调用runtime_pollServerInit最终会调用internal/poll.runtime_pollServerInit函数

```go
func poll_runtime_pollServerInit() {
    // 调用netpollinit()创建ep_fd
    netpollinit()
    // 将netpollInited设为1，表示已初始化ep_fd
    atomic.Store(&netpollInited, 1)
}
```

​	netpollinit()的实际内容如下

```go
func netpollinit() {    
    // 一个Go进程通过listen创建的ep_fd是一个全局变量
    epfd = epollcreate1(_EPOLL_CLOEXEC)
    if epfd >= 0 {
        return
    }
    epfd = epollcreate(1024)
    if epfd >= 0 {  
        closeonexec(epfd)
        return
    }
    println("runtime: epollcreate failed with", -epfd)
    throw("runtime: netpollinit failed")
}
```

​	先调用epoll_create1，对应OS的int epoll_create1(int flag)来创建ep_fd，如果失败（epfd<0）就调用epollcreate，对应OS的int epoll_create(int size)来创建ep_fd。closeonexec貌似是防止ep_fd文件泄漏的。

​	再看一下epoll_ctl()部分

```go
func poll_runtime_pollOpen(fd uintptr) (*pollDesc, int) {
    // 通过netpoll.go的全局变量pollcache的alloc()方法返回一个pollDesc对象
    pd := pollcache.alloc()
    lock(&pd.lock)    //上锁
    // 这里的rg,wg正常来说是要等于0的
    if pd.wg != 0 && pd.wg != pdReady {
        throw("runtime: blocked write on free polldesc")
    }
    if pd.rg != 0 && pd.rg != pdReady {
        throw("runtime: blocked read on free polldesc")
    }    
    //注意这里的fd是要监听的sysfd（实际fd）
    pd.fd = fd   
    pd.closing = false
    pd.rseq++
    pd.rg = 0
    pd.rd = 0
    pd.wseq++
    pd.wg = 0
    pd.wd = 0
    unlock(&pd.lock)

    var errno int32    
    //实际将fd注册到ep_fd里
    errno = netpollopen(fd, pd)
    return pd, int(errno)
}
```

​	epoll_ctl()部分有两个关键的方法与函数：alloc()与netpollopen()，前者获取pollDesc对象，后者将fd实际注册到ep_fd里，注册过程需要pd的参与。这里先看一下alloc()的过程：

```go
/**
	pollCache的底层结构是单链表，pollCache含有一个first指针，用来指向准备使用的pollDesc节点(可以理解为Java-Iterator里的next指针)
*/
func (c *pollCache) alloc() *pollDesc {
    //上锁
    lock(&c.lock)
    // 相当于 iterator.hasNext() == false，即链表为空，此时会创建一个特定长度的pollDesc链表
    if c.first == nil {
        /*
        	pollBlockSize：netpoll.go里的常量，值为4*1024
        	unsafe.Sizeof()：返回结构体内数据类型占用的字节大小，如Golang中string结构体包含两个属性：Data uintptr、Len int。unsafe.Sizeof("a")返回的是16（在64位系统中，uintptr和int都占8个字节）
        */
        const pdSize = unsafe.Sizeof(pollDesc{})
        // n为 (4*1024 / 一个pd对象占用的内存大小)，即pollCache内要创建n个pd
        n := pollBlockSize / pdSize
        if n == 0 {
            n = 1
        }
        // Must be in non-GC memory because can be referenced
        // only from epoll/kqueue internals.        
        //通过malloc.go的函数申请内存，这部分的内存不会被GC回收，且只能被epoll和kqueue引用
        mem := persistentalloc(n*pdSize, 0, &memstats.other_sys)
        
        //通过指针移动来获取节点并将节点串成链表。
        for i := uintptr(0); i < n; i++ {
            //Golang中*unsafe.Pointer可以直接用()强转成其他类型
            pd := (*pollDesc)(add(mem, i*pdSize))
            pd.link = c.first
            c.first = pd
        }
    }
    // 当删除注册的event事件时，会回收该节点.
    pd := c.first
    c.first = pd.link
    unlock(&c.lock)
    return pd
}
```

​	netpollopen()是实际将alloc()返回的pd对象保存到epoll_event的data里，然后将epoll_event注册到ep_fd中

```go
func netpollopen(fd uintptr, pd *pollDesc) int32 {
    var ev epollevent    
    //对应描述符上有可读数据|对应描述符上有可写数据|描述符被挂起|设置为边缘触发模式
    ev.events = _EPOLLIN | _EPOLLOUT | _EPOLLRDHUP | _EPOLLET   
    //构造epoll的用户数据
    *(**pollDesc)(unsafe.Pointer(&ev.data)) = pd 
    //这里的函数应该对应了OS的int epoll_ctl(int epfd, int op, int fd, struct epoll_event *event);走到这里则说明事件已被注册到ep_fd里了
    return -epollctl(epfd, _EPOLL_CTL_ADD, int32(fd), &ev)
}
```

​	回到注册失败后的情况，runtime_pollOpen(uintptr(fd.Sysfd))返回的error != nil。此时会做两个关键的处理runtime_pollUnblock(ctx)与runtime_pollClose(ctx)

​	先看一下runtime_pollUnblock()

```go
func poll_runtime_pollUnblock(pd *pollDesc) {
    lock(&pd.lock)
    if pd.closing {
        throw("runtime: unblock on closing polldesc")
    }    
    // TODO 为什么要设为true？
    pd.closing = true
    pd.rseq++
    pd.wseq++
    var rg, wg *g
    atomic.StorepNoWB(noescape(unsafe.Pointer(&rg)), nil)   
    /*
    netpollunblock的第三个函数ioready用来说明本次解锁是否因为epoll事件触发了，很明显这里的是因为注册事件失败了，所以为false。如果有对应阻塞的goroutine则直接返回，如果没有则返回nil。
    */
    rg = netpollunblock(pd, 'r', false)
    wg = netpollunblock(pd, 'w', false)    
    if pd.rt.f != nil {
        deltimer(&pd.rt)
        pd.rt.f = nil
    }
    if pd.wt.f != nil {
        deltimer(&pd.wt)
        pd.wt.f = nil
    }
    unlock(&pd.lock)  
    /* 获取完读和写的goroutine后，调用netpollgoready()唤醒这个goroutine
    注意！runtime_pollUnblock，runtime_pollUnblock是因为epoll_ctl()有异常才调用，在Go1.14里epoll_wait()后唤醒goroutine是基于runtime.netpoll()
    */
    if rg != nil {
        netpollgoready(rg, 3)
    }
    if wg != nil {
        netpollgoready(wg, 3)
    }
}
```

​	再看一下runtime_pollUnblock()获取阻塞goroutine的关键部分

```go
func netpollunblock(pd *pollDesc, mode int32, ioready bool) *g {
    // 先netpoll的话：gpp = 0
    gpp := &pd.rg
    if mode == 'w' {
        gpp = &pd.wg
    }
    //使用for循环用于执行atomic.Casuintptr原子操作。一个pd对于一条连接，而一条连接可能被多个goroutine操作。
    for {
        // 先netpoll的话：old = 0
        old := *gpp        
        //如果gpp为pdReady，则对应的goroutine已经被unblock状态，直接返回nil就好了
        if old == pdReady {
            return nil
        }        
        //当Listen未Accept场景，并没有goroutine阻塞，因此直接返回
        if old == 0 && !ioready {
            // Only set READY for ioready. runtime_pollWait
            // will check for timeout/cancel before waiting.
            return nil
        }
        var new uintptr
        if ioready {
            // 先netpoll的话：new = 0
            new = pdReady
        }        
       	// atomic.Casuintptr(gpp, old, new)本质上是一把自旋锁
        // 先netpoll的话 gpp = old = new = 0，结果大家还是0
        if atomic.Casuintptr(gpp, old, new) {
            //old == pdReady的判断应该不会被执行，如果old为pdReady，上面代码已经直接返回了
            if old == pdReady || old == pdWait {
                old = 0
            }            
            //返回阻塞在pd.rg/pd.wg上的goroutine地址
            // 先netpoll的话 就return了一个unsafe.Pointer(uintptr(0)) 
            return (*g)(unsafe.Pointer(old))
        }
    }
}
```

​	然后是注册失败的第二个处理：runtime_pollClose(ctx)，主要是用来从ep_fd删除事件，并回收与之关联的pollDesc对象

```go
func poll_runtime_pollClose(pd *pollDesc) {
    if !pd.closing {
        throw("runtime: close polldesc w/o unblock")
    }    
    if pd.wg != 0 && pd.wg != pdReady {
        throw("runtime: blocked write on closing polldesc")
    }
    if pd.rg != 0 && pd.rg != pdReady {
        throw("runtime: blocked read on closing polldesc")
    }    
    //调用_EPOLL_CTL_DEL删除注册的epoll事件
    netpollclose(pd.fd)    
    //回收fd节点
    pollcache.free(pd)
}
```

### Listen总结

​	总的来看，netpoll里调用listen监听端口后，只是做了epoll_create()与epoll_ctl()部分。剩下的epoll_wait会在Accept()，Read()方法里调用，调用链通过poll_runtime_pollWait → netpollblock → gopark 来阻塞当前goroutine。

## Accept()过程

​	在demo代码中，通过net.Listen("tcp", "localhost:8080")返回了一个TCPListener对象，listener调用Aceept()方法监听连接到来（此时会epoll_wait），连接建立后Accept()会返回一个TCPConn对象。

​	先看看Accept源码。

```go
func (l *TCPListener) Accept() (Conn, error) {    
    if !l.ok() {
        return nil, syscall.EINVAL
    }
    c, err := l.accept()
    if err != nil {
        return nil, &OpError{Op: "accept", Net: l.fd.net, Source: nil, Addr: l.fd.laddr, Err: err}
    }
    return c, nil
}
```

​	可以看到Accept()内主要调用了accpet()方法。

```go
func (ln *TCPListener) accept() (*TCPConn, error) {
    fd, err := ln.fd.accept()
    if err != nil {
        return nil, err
    }    //设置TCP_NODELAY选项来禁止Nagle算法（避免粘包？）
    return newTCPConn(fd), nil
}
```

​	这里又是一层封装，可以看到实际是调用了listener的netFD的accept()方法，并将返回的fd封装到TCPConn内。

```go
func (fd *netFD) accept() (netfd *netFD, err error) {
   // 阻塞会发生在这里，阻塞完成说明连接已建立，此时返回该连接对应的socket fd。
   d, rsa, errcall, err := fd.pfd.Accept()
   if err != nil {
      if errcall != "" {
         err = wrapSyscallError(errcall, err)
      }
      return nil, err
   }

    /*
    能走到这里说明连接已建立完成了,这时会基于listener的fd、连接的fd来构造一个netFD，该netFD对应这个刚建立的连接。
    */
   if netfd, err = newFD(d, fd.family, fd.sotype, fd.net); err != nil {
      poll.CloseFunc(d)
      return nil, err
   }
    /*
    这里会用上面新建的netFD初始化一个pollDesc，注册到ep_fd里，与listen不同的是本次注册读就绪与写就绪事件。但还未阻塞，除非调用Read()或Write()。
   	*/
   if err = netfd.init(); err != nil {
      netfd.Close()
      return nil, err
   }
    // 最终将连接对应的netFD返回
   lsa, _ := syscall.Getsockname(netfd.pfd.Sysfd)
   netfd.setAddr(netfd.addrFunc()(lsa), netfd.addrFunc()(rsa))
   return netfd, nil
}
```

​	这里又封装了一层，实际上返回的fd是[netFD]的[pfd(FD)]的[Accpet方法]的返回值，具体看一下实现：

```go
func (fd *FD) Accept() (int, syscall.Sockaddr, string, error) {   
    if err := fd.readLock(); err != nil {
        return -1, nil, "", err
    }    
    defer fd.readUnlock()
    
    if err := fd.pd.prepareRead(fd.isFile); err != nil {
        return -1, nil, "", err
    }
    for {
        /*
        	最关键的方法，调用OS的accpet函数返回一个fd
        	参数传入的fd其实是listener的netFD的FD，在创建listener的时候已经设置fd为非阻塞了，也就是说调用了accept()后会立即返回。如果有连接到来，返回err为nil，并直接返回。如果连接还没到，会返回syscall.EAGAIN这个err。
        */
        s, rsa, errcall, err := accept(fd.Sysfd)
        if err == nil {
            return s, rsa, "", err
        }
        switch err {
        /*
        	syscall.EAGAIN代表连接还未到，此时先用pd.pollable()判断是否使用epoll机制， 如果是的话则调用waitRead()使当前goroutine阻塞在这里，waitRead()底层调用了runtime_pollWaitd。此时goroutine就在这里等待唤醒了。
        */
        case syscall.EAGAIN:                     
        	if fd.pd.pollable() {
                if err = fd.pd.waitRead(fd.isFile); err == nil {
                    continue
                }
            }      
        case syscall.ECONNABORTED:
            // This means that a socket on the listen
            // queue was closed before we Accept()ed it;
            // it's a silly error, so try again.
            continue
        }
        return -1, nil, errcall, err
    }
}
```

​	runtime_pollWaitd实际是调用了runtime.poll_runtime_pollWait() → netpollblock() → gopark() 来完成goroutine阻塞,**注意！！！这里的gopark和操作channel时的gopark一致。**

```go
func poll_runtime_pollWait(pd *pollDesc, mode int) int {
    // 是否超时或关闭？
    err := netpollcheckerr(pd, int32(mode))
    if err != 0 {
        return err
    }
    // As for now only Solaris and AIX use level-triggered IO.
    if GOOS == "solaris" || GOOS == "aix" {
        netpollarm(pd, mode)
    }
    for !netpollblock(pd, int32(mode), false) {        
        err = netpollcheckerr(pd, int32(mode))
        if err != 0 {
            return err
        }
        // Can happen if timeout has fired and unblocked us,
        // but before we had a chance to run, timeout has been reset.
        // Pretend it has not happened and retry.
    }
    return 0
}
```

```go
func netpollblock(pd *pollDesc, mode int32, waitio bool) bool {
    gpp := &pd.rg
    if mode == 'w' {
        gpp = &pd.wg
    }

    for {
        old := *gpp
        // 等于pdReady说明有事件就绪了，直接返回
        if old == pdReady {
            *gpp = 0
            return true
        }
        // 如果既不等于pdReady，又不等于0，说明已经有一个goroutine在阻塞中，此时抛出异常不允许两个goroutine同时阻塞。
        if old != 0 {
            throw("runtime: double wait")
        }
        if atomic.Casuintptr(gpp, 0, pdWait) {
            break
        }
    }

    // need to recheck error states after setting gpp to WAIT
    // this is necessary because runtime_pollUnblock/runtime_pollSetDeadline/deadlineimpl
    // do the opposite: store to closing/rd/wd, membarrier, load of rg/wg    //waitio为true可用于等待ioReady
    if waitio || netpollcheckerr(pd, mode) == 0 {        
        //直接阻塞goroutine，等待唤醒。注意！！！这个netpollblockcommit()函数可以将当前goroutine放到pd的wg或rg里，里面用到了二级指针。
        gopark(netpollblockcommit, unsafe.Pointer(gpp), waitReasonIOWait, traceEvGoBlockNet, 5)
    }
    // be careful to not lose concurrent READY notification
    old := atomic.Xchguintptr(gpp, 0)
    if old > pdWait {
        throw("runtime: corrupted polldesc")
    }
    return old == pdReady
}
```

### Accept总结

​	总的来说，Accept()比较关键的方法在func (fd *netFD) accept() (netfd *netFD, err error)。listener调用Accept()后，主要是先查看listener所监听的socket fd是否有连接事件到来了？有的话直接建立连接，没有的话阻塞调用Accept()的goroutine，直到连接到来后才唤醒，并建立连接。**在建立连接后会返回该连接的socket fd，同时为该socket fd注册读写监听事件，但不会阻塞。最后将该socket fd封装成一个TCPConn指针**。

​	值得注意的是，在goroutine阻塞部分，即调用netpollblock()时，只会调用gopark()把goroutine给阻塞了而已，本质上还没调用epoll_wait()，**实际的epoll_wait是在runtime.epollwait里调用，runtime.epollwait实际在runtime.netpoll里调用，runtime.netpoll会在单独的线程运行**，这个问题在下面会解答。



## Read()过程

```go
func (c *conn) Read(b []byte) (int, error) {
    if !c.ok() {
        return 0, syscall.EINVAL
    }
    n, err := c.fd.Read(b)
    if err != nil && err != io.EOF {
        err = &OpError{Op: "read", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
    }    
    return n, err
}
```

```go
func (fd *netFD) Read(p []byte) (n int, err error) {
    n, err = fd.pfd.Read(p)    
    runtime.KeepAlive(fd)
    return n, wrapSyscallError("read", err)
}
```

​	最核心的代码是fd.pfd.Read(p)

```go
func (fd *FD) Read(p []byte) (int, error) {   
    if err := fd.readLock(); err != nil {
        return 0, err
    }   
    defer fd.readUnlock()
    if len(p) == 0 {
        // If the caller wanted a zero byte read, return immediately
        // without trying (but after acquiring the readLock).
        // Otherwise syscall.Read returns 0, nil which looks like
        // io.EOF.
        // TODO(bradfitz): make it wait for readability? (Issue 15735)
        return 0, nil
    }
    if err := fd.pd.prepareRead(fd.isFile); err != nil {
        return 0, err
    }
    if fd.IsStream && len(p) > maxRW {
        p = p[:maxRW]
    }
    for {       
        /*
        在accept一个连接的时候，已经把这个连接设为非阻塞了，因此用这个连接调用Read()时，会立即返回。如果有数据，err为nil，如果没数据，err不为nil。
        */
        n, err := syscall.Read(fd.Sysfd, p)
        if err != nil {
            n = 0            
            // EAGAIN表示期待的IO事件未发生，在这里代表无数据可读，并且还判断了是否为epoll模式
            if err == syscall.EAGAIN && fd.pd.pollable() {
                //阻塞goroutine，等待唤醒
                if err = fd.pd.waitRead(fd.isFile); err == nil {
                    continue
                }
            }

            // On MacOS we can see EINTR here if the user
            // pressed ^Z.  See issue #22838.
            if runtime.GOOS == "darwin" && err == syscall.EINTR {
                continue
            }
        }
        err = fd.eofError(n, err)
        return n, err
    }
}
```

### Read总结

​	总得来说Read没有Listen()和Accpet()复杂，如果就绪了直接读，没就绪直接阻塞。



## Write()过程

```go
func (c *conn) Write(b []byte) (int, error) {
    if !c.ok() {
        return 0, syscall.EINVAL
    }
    n, err := c.fd.Write(b)
    if err != nil {
        err = &OpError{Op: "write", Net: c.fd.net, Source: c.fd.laddr, Addr: c.fd.raddr, Err: err}
    }
    return n, err
}
```

```go
func (fd *netFD) Write(p []byte) (nn int, err error) {
    nn, err = fd.pfd.Write(p)
    runtime.KeepAlive(fd)
    return nn, wrapSyscallError("write", err)
}
```

```go
func (fd *FD) Write(p []byte) (int, error) {
    if err := fd.writeLock(); err != nil {
        return 0, err
    }
    defer fd.writeUnlock()
    if err := fd.pd.prepareWrite(fd.isFile); err != nil {
        return 0, err
    }
    var nn int    
    //循环写数据。
    for {
        max := len(p)        
        if fd.IsStream && max-nn > maxRW {
            max = nn + maxRW
        }        
        //调用syscall.Write函数，向缓冲区写数据，返回发送成功的字节数。
        n, err := syscall.Write(fd.Sysfd, p[nn:max])
        if n > 0 {
            nn += n
        }        
        if nn == len(p) {
            return nn, err
        }       
        //如果syscall.Write()返回了err，且为EAGAIN时，代表事件未就绪，即不可写，同时判断是epoll模式的话，进入阻塞。
        if err == syscall.EAGAIN && fd.pd.pollable() {
            if err = fd.pd.waitWrite(fd.isFile); err == nil {
                continue
            }
        }
        if err != nil {
            return nn, err
        }
        if n == 0 {
            return nn, io.ErrUnexpectedEOF
        }
    }
}
```

### Write总结

​	可以看到Accept和Read都是调用waitRead，唯独写操作调用的是waitWrite，初次以外和读操作没特别大的差距。

## runtime.netpoll总过程

​	通过Listen()方法返回一个listener，调用listener的Accpet()方法等待连接（阻塞），直到有连接后返回一个Conn对象，并且Accpet()会将Conn对象对应的socket fd注册到ep_fd里监听读与写事件。一个Conn可以开一个goroutine来执行Read()/Write()方法来等待可读/可写（也可以不开，但这样后果很严重）。

​	当Accpet()、Read()、Write()需要阻塞时，goroutine会调用gopark()挂起自己。此时P会调度下一个G给M使用。那什么时候唤醒呢？会在循环被调度的**runtime.schedule()**与**sysmon 监控线程**中调用**runtime.netpoll**，netpoll()底层封装了epoll_wait来返回事件触发的goroutine，并通过injectglist这个结构将事件触发的goroutine扔到P的local queue或global queue里等待调度。

​	所以说epoll_wait的核心在于runtime.netpoll，那runtime.netpoll是如何调度呢？先看下runtime.netpoll源码

```go
// delay指的是超时时间
func netpoll(delay int64) gList {
   if epfd == -1 {
      return gList{}
   }
   var waitms int32
    // 根据delay转换成有效的超时时间waitms
   if delay < 0 {
      waitms = -1
   } else if delay == 0 {
      waitms = 0
   } else if delay < 1e6 {
      waitms = 1
   } else if delay < 1e15 {
      waitms = int32(delay / 1e6)
   } else {
      // An arbitrary cap on how long to wait for a timer.
      // 1e9 ms == ~11.5 days.
      waitms = 1e9
   }
   var events [128]epollevent
retry:
    //调用epoll_wait，如果不指定waitms会阻塞
   n := epollwait(epfd, &events[0], int32(len(events)), waitms)
   if n < 0 {
      if n != -_EINTR {
         println("runtime: epollwait on fd", epfd, "failed with", -n)
         throw("runtime: netpoll failed")
      }
      // If a timed sleep was interrupted, just return to
      // recalculate how long we should sleep now.
      if waitms > 0 {
         return gList{}
      }
       // 如果epoll_wait的n小于0，且等待时间也小于0，则直接goto继续epoll_wait
      goto retry
   }
    // 就绪的goroutine队列
   var toRun gList
    // 遍历epoll_wait里返回的就绪event,从event里获取到data（netpollopen在注册fd的时候会将pollDesc保存到data里），即pd，从pd
   for i := int32(0); i < n; i++ {
      ev := &events[i]
      if ev.events == 0 {
         continue
      }

      if *(**uintptr)(unsafe.Pointer(&ev.data)) == &netpollBreakRd {
         if ev.events != _EPOLLIN {
            println("runtime: netpoll: break fd ready for", ev.events)
            throw("runtime: netpoll: break fd ready for something unexpected")
         }
         if delay != 0 {
            // netpollBreak could be picked up by a
            // nonblocking poll. Only read the byte
            // if blocking.
            var tmp [16]byte
            read(int32(netpollBreakRd), noescape(unsafe.Pointer(&tmp[0])), int32(len(tmp)))
         }
         continue
      }
       // 判断这个event是就绪事件是什么
      var mode int32
      if ev.events&(_EPOLLIN|_EPOLLRDHUP|_EPOLLHUP|_EPOLLERR) != 0 {
         mode += 'r'
      }
      if ev.events&(_EPOLLOUT|_EPOLLHUP|_EPOLLERR) != 0 {
         mode += 'w'
      }
      if mode != 0 {
          //用万能指针将event.data强转为pollDesc
         pd := *(**pollDesc)(unsafe.Pointer(&ev.data))
         pd.everr = false
         if ev.events == _EPOLLERR {
            pd.everr = true
         }
          // 将goroutine扔到toRun链表里
         netpollready(&toRun, pd, mode)
      }
   }
   return toRun
}
```

​	可以看到netpoll()里有两个关键的函数，一个是**epollwait(epfd, &events[0], int32(len(events)), waitms)**，调用后会阻塞等待事件发生（直到超时），如果有事件触发则遍历事件，拿出事件里**封装的data(pollDesc)**以及**事件类型**，调用**netpollready()**将pollDesc里的goroutine放入toRun队列里。

```go
func netpollready(toRun *gList, pd *pollDesc, mode int32) {
   var rg, wg *g
   if mode == 'r' || mode == 'r'+'w' {
       // 唤醒等待读、等待读或写的goroutine
      rg = netpollunblock(pd, 'r', true)
   }
   if mode == 'w' || mode == 'r'+'w' {
       // 唤醒等待写、等待读或写的goroutine
      wg = netpollunblock(pd, 'w', true)
   }
    // 将goroutine添加到toRun队列里
   if rg != nil {
      toRun.push(rg)
   }
   if wg != nil {
      toRun.push(wg)
   }
}
```

​	此时又回到了netpollunblock()函数，最后回到netpoll返回的glist那，**注意！此时netpollunblock()返回的goroutine的状态还是_Gwaiting**。那现在最重要的问题来了，netpoll()这个函数到底是谁在调用？

​	答1：runtime.netpoll()函数是在runtime.findrunable()函数里调用的，runtime.findrunable()又是在runtime.schedule()里调用的。我们先看下源码

```go
func schedule() {
    // runtime.schedule()的函数很长，这里只拿了关键部分
    if gp == nil {
       gp, inheritTime = findrunnable() // blocks until work is available
    }
}
```

```go
func findrunnable() (gp *g, inheritTime bool) {
    // runtime.findrunnable()的函数也很长，这里只拿了关键部分
    if netpollinited() && (atomic.Load(&netpollWaiters) > 0 || pollUntil != 0) && atomic.Xchg64(&sched.lastpoll, 0) != 0 {
		atomic.Store64(&sched.pollUntil, uint64(pollUntil))
		if _g_.m.p != 0 {
			throw("findrunnable: netpoll with p")
		}
		if _g_.m.spinning {
			throw("findrunnable: netpoll with spinning")
		}
		if faketime != 0 {
			// When using fake time, just poll.
			delta = 0
		}
        // 核心部分，调用netpoll获取就绪的goroutine链表
		list := netpoll(delta) 
		atomic.Store64(&sched.pollUntil, 0)
		atomic.Store64(&sched.lastpoll, uint64(nanotime()))
		if faketime != 0 && list.empty() {
			// Using fake time and nothing is ready; stop M.
			// When all M's stop, checkdead will call timejump.
			stopm()
			goto top
		}
		lock(&sched.lock)
		_p_ = pidleget()
		unlock(&sched.lock)
		if _p_ == nil {
			injectglist(&list)
		} else {
			acquirep(_p_)
             // 如果返回的goroutine链表不为空
			if !list.empty() {
                // 先取出一个goroutine
				gp := list.pop()
                // 将整个链表传给injectglist()函数，TODO 那上面的gp有啥用？
				injectglist(&list)
				casgstatus(gp, _Gwaiting, _Grunnable)
				if trace.enabled {
					traceGoUnpark(gp, 0)
				}
				return gp, false
			}
			if wasSpinning {
				_g_.m.spinning = true
				atomic.Xadd(&sched.nmspinning, 1)
			}
			goto top
		}
}
```

​	runtime.findrunable()里比较关键的函数是runtime.injectglist()，它会将gList里所有goroutine从阻塞变为就绪，并扔到P的local queue或global queue里。

```go
func injectglist(glist *gList) {
   if glist.empty() {
      return
   }
   if trace.enabled {
      for gp := glist.head.ptr(); gp != nil; gp = gp.schedlink.ptr() {
         traceGoUnpark(gp, 0)
      }
   }
   lock(&sched.lock)
   var n int
   for n = 0; !glist.empty(); n++ {
       // 循环glist，每一个goroutine
      gp := glist.pop()
       // 将goroutine的状态从_Gwaiting变为_Grunnable
      casgstatus(gp, _Gwaiting, _Grunnable)
      globrunqput(gp)
   }
   unlock(&sched.lock)
   for ; n != 0 && sched.npidle != 0; n-- {
      startm(nil, false)
   }
   *glist = gList{}
}
```

​	答2：runtime.netpoll()还会在sysmon线程的sysmon()函数里调用，先看一下sysmon()的部分源码

```go
func sysmon() {
    // 查看上一次runtime.netpoll()的时间，如果距离现在已超过10ms，则再netpoll一次
    lastpoll := int64(atomic.Load64(&sched.lastpoll))
    if netpollinited() && lastpoll != 0 && lastpoll+10*1000*1000 < now {
       atomic.Cas64(&sched.lastpoll, uint64(lastpoll), uint64(now))
        // 再一次netpoll()
       list := netpoll(0) 
       if !list.empty() {
          // Need to decrement number of idle locked M's
          // (pretending that one more is running) before injectglist.
          // Otherwise it can lead to the following situation:
          // injectglist grabs all P's but before it starts M's to run the P's,
          // another M returns from syscall, finishes running its G,
          // observes that there is no work to do and no other running M's
          // and reports deadlock.
          incidlelocked(-1)
           //再一次injectglist()
          injectglist(&list)
          incidlelocked(1)
       }
    }
}
```

​	sysmon线程是Go runtime启动后新建的一个线程，他是一个独立的M，不需要依赖P就可执行。sysmon函数每20μs - 10ms执行一次。**TODO 这个线程以后得新开一篇笔记来记录。**



# netpoll的总结与意义

## 总得来说，netpoll是基于Non Blocking IO + IO多路复用 + MPG模型形成的一个同步网络IO模型，大致过程如下：

​	1.通过Listen()新建listener，此时listener已内置了一个需要监听的fd，**即已经历了epoll_create()、epoll_ctl()**。

​	2.listener.Accpet()，会先利用Non Blocking IO调用accept()函数，由于是非阻塞，所以立刻返回。如果返回的结果是未有连接到来，那么accept()还会返回一个err，程序根据这个err判断需要阻塞等待连接，于是最终gopark()阻塞住调用listener.Accept()的goroutine，**此时会将goroutine放到对应的pollDesc里**。**注意！这一部分体现了Non Blocking IO + MPG模型。**

​	3.当runtime.netpoll()被调度到，使用epoll_wait发现2.的fd来了个连接，此时会返回epoll_events给runtime，runtime通过遍历epoll_events里event的data找到pollDesc（epoll_ctl的时候会将pollDesc放到event的data里），从而唤醒pollDesc里的goroutine。	**注意！这一部分体现了IO多路复用+MPG模型**。

​	此时会有一个问题，如果我在2.调用Accpet()之前就已经调用了runtime.netpoll()，会发生什么？此时从event里拿到的pollDesc里找到rg/wg为0，在netpollunblock()里会因为rg/wg为0而返回nil，最后因为返回的goroutine为nil而不会push到toRun队列里。（具体可以看netpollunblock的笔记）。	

​	4.当2.里的goroutine被3.唤醒后，会再一次调用accept()，此时因为连接就绪，不会返回err，还会返回一个socket fd，这个socket fd代表accept后建立的连接描述符，然后对这个socket fd注册到ep_fd里（epoll_ctl）。最后经过层层处理socket fd，返回一个Conn对象。

​	5.在4.返回的Conn已经对应了一个fd，且4.又会对这个fd注册读写监听。当Conn调用Read()或Write()时，会先基于Non Blocking IO调用read或write，如果未就绪则返回err，如果err!=nil则最终gopark()阻塞住调用Read()或Write()的goroutine。这一点和2.很像，剩下的就是3.步骤等待唤醒了。goroutine被唤醒后又会再调用一遍read()/write()，此时不会阻塞，流程继续走下去。

​	netpoll的核心操作是：最开始会为fd注册epoll事件，然后对fd进行**非阻塞IO调用**，查看是否可行。如果事件未就绪导致不可行，则gopark()住goroutine。此时runtime.netpoll()会不断执行等待就绪的事件，当发现fd的事件就绪后，会通过fd → pd → goroutine找到要唤醒的goroutine，最后唤醒它，扔到P的local queue或global queue里。

## netpoll的意义

​	为什么netpoll会先调用非阻塞IO呢？回到MPG模型的笔记说过，G如果进行系统调用而阻塞，如sys_read();会触发MPG模型里的HandOff机制，将P从MPG中摘除，让G与M一直阻塞在那。

​	然而HandOff机制的效率很低，为什么？因为HandOff机制使goroutine跟随M进入内核态，**此时G的调度权就交给了OS，而不是Go scheduler了。**这也是为什么goroutine操作channel时，会被打包成sudog并阻塞，而不是触发HandOff机制（具体看 05当一个G因通道阻塞的时候，他在干什么？）。所以在netpoll的设计中，会先调用非阻塞IO执行系统函数，即使失败也会立即返回而不是阻塞，并声明一个err，netpoll会根据这个err从用户层面gopark()住这个goroutine，且epoll_wait的实际调用也是基于独立线程sysmon，当事件就绪后netpoll再唤醒被阻塞住的goroutine。netpoll这种基于Non Blocking IO + IO多路复用 + MPG的设计，保证goroutine的阻塞发生在用户态而非内核态。



# netpoll的缺陷与接下来的计划

​	netpoll与Java BIO有一些相似的地方，虽然netpoll底层是基于epoll()来调度唤醒阻塞的goroutine，但netpoll的设计理念是goroutine per connection，Java BIO是thread per connection。虽然是1对1，但goroutine占用的资源比Java Thread轻的多，默认2KB并按需动态扩容的goroutine在一定程度上能浪费得起用来与连接1对1（相比之下Java就不行了，thread与连接一对一太浪费了）。但1对1终究是1对1，如果在海量业务超高数量的长连接下，会建立起同等数量的goroutine，首先goroutine的创建和销毁就是一个问题，虽然底层提供了gouroutine缓存链表，可以理解为池，但还是不能满足超高数量的连接。

​	当goroutine的数量到达一个峰值，goroutines本身侵占的资源总量会给go scheduler调度goroutine造成极大压力，从而导致性能下降。因此避免netpoll带来的缺陷，还得设计出一款更高效的网络IO框架，这个框架的整体思想应该与Netty类似，即主从Reactor多线程（协程），我感觉比起Netty的多线程，Golang的多协程能更大程度上提高请求的吞吐量。因此接下来的计划是寻找类似的网络框架进行研究，目前找到最完美的是gnet（提前埋个学习的坑）。