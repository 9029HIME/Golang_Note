package _5当一个G因通道阻塞的时候_他在干什么_

/**
https://www.youtube.com/watch?v=KBZlN0izeiY
https://zhuanlan.zhihu.com/p/27917262
*/
/**
背景：我已经了解了MPG模型与Channel、Goroutine的常用方式。
问题描述：
	假设 chan 是无缓冲通道，假设 G1 往 chan 里扔数据阻塞了。
	此时会触发 Go 的 HandOff 机制，将 P1 detach()到其他 M 上吗？
	如果会，过段时间另一个协程 G2 从 chan 里取数据，G1 的阻塞结束。
	G1 会直接进入某个 P 的协程队列尾(有空闲 P)或全局协程队列尾(无空闲 P)吗？还是说插队回到 P1 的队首？
	channel := make(chan int)
	go func() {
		//TODO 在主线程读通道前，这个Goroutine到底处于什么状态？
		//TODO 还有一个很奇怪的疑问，同一时刻，一个Goroutine会不会阻塞在两个不同的Channel？
		channel <- 1
	}()
	time.Sleep(time.Second*3)
	<-channel
过程：
	1.先看一下channel结构体源码（E:/Enviroment/Go/go1.14.8/src/runtime/chan.go:32）
		type hchan struct {
			qcount   uint           // total data in the queue （缓冲）通道里的数据个数，类似length
			dataqsiz uint           // size of the circular queue （缓冲）通道的长度，目测类似cap
			buf      unsafe.Pointer // points to an array of dataqsiz elements 数据是存储在环形队列，这是队列的头指针。TODO 如果是无缓冲管道，这个属性值未nil（未验证）
			elemsize uint16			// 未知
			closed   uint32			// 通道是否被关闭（关闭后变成只读通道）
			elemtype *_type // element typ	通道内数据的类型
			sendx    uint   // send index	（缓冲）对通道发送数据时，该数据要存到哪个下标
			recvx    uint   // receive index （缓冲）从通道读取数据，要读哪一个下表的数据
			recvq    waitq  // list of recv waiters TODO 等到对通道进行读操作的Goroutine队列
			sendq    waitq  // list of send waiters	TODO 等待对通道进行写操作的Goroutine队列

			// lock protects all fields in hchan, as well as several
			// fields in sudogs blocked on this channel.
			//
			// Do not change another G's status while holding this lock
			// (in particular, do not ready a G), as this can deadlock
			// with stack shrinking.
			lock mutex //channel的锁，再Goroutine对通道进行实际操作时要上锁的
		}
	2.从1.可以初步看出，Goroutine是被打包成sudog结点，挂在channel一个名叫waitq的sudog队列。现在看一下waitq的源码（E:/Enviroment/Go/go1.14.8/src/runtime/chan.go:53）
		type waitq struct {
			first *sudog	队列头指针
			last  *sudog	队列尾指针
		}
	3.为什么Goroutine还要打包成sudog？现在看看sudog源码（E:/Enviroment/Go/go1.14.8/src/runtime/runtime2.go:342）
		type sudog struct {
			// The following fields are protected by the hchan.lock of the
			// channel this sudog is blocking on. shrinkstack depends on
			// this for sudogs involved in channel ops.

			g *g	//Goroutine现实结构体指针

			// isSelect indicates g is participating in a select, so
			// g.selectDone must be CAS'd to win the wake-up race.
			isSelect bool	// 是否处于select的case状态
			next     *sudog	//下一个节点
			prev     *sudog	//上一个节点
			elem     unsafe.Pointer // data element (may point to stack)

			// The following fields are never accessed concurrently.
			// For channels, waitlink is only accessed by g.
			// For semaphores, all fields (including the ones above)
			// are only accessed when holding a semaRoot lock.

			acquiretime int64  // 创建时间
			releasetime int64  // 销毁时间（TODO sudog的销毁并非gc，而是回归到缓存池里）
			ticket      uint32
			parent      *sudog // semaRoot binary tree
			waitlink    *sudog // g.waiting list or semaRoot
			waittail    *sudog // semaRoot
			c           *hchan // channel	挂在哪个通道的waitq
		}
	4.结合3. 再看看获取sudog的方法acquireSudog（E:/Enviroment/Go/go1.14.8/src/runtime/proc.go:320）
		func acquireSudog() *sudog {
			// Delicate dance: the semaphore implementation calls
			// acquireSudog, acquireSudog calls new(sudog),
			// new calls malloc, malloc can call the garbage collector,
			// and the garbage collector calls the semaphore implementation
			// in stopTheWorld.
			// Break the cycle by doing acquirem/releasem around new(sudog).
			// The acquirem/releasem increments m.locks during new(sudog),
			// which keeps the garbage collector from being invoked.
			mp := acquirem()	//获取当前G的M
			pp := mp.p.ptr()	//获取当前G的P
			// 现在P的sudog队列里，寻找有没有可用的sudog TODO 由此可知P存在sudog队列
			if len(pp.sudogcache) == 0 {
				// 如果P的sudog队列为空，则找全局的sudog队列（感觉像找G一样），TODO 对全局sudog队列的操作是有锁的
				lock(&sched.sudoglock)
				// First, try to grab a batch from central cache.
				for len(pp.sudogcache) < cap(pp.sudogcache)/2 && sched.sudogcache != nil {
					s := sched.sudogcache
					sched.sudogcache = s.next
					s.next = nil
					pp.sudogcache = append(pp.sudogcache, s)
				}
				unlock(&sched.sudoglock)
				// If the central cache is empty, allocate a new one.
				// 如果还是没有，则直接在P的sudog队列里new一个新sudog
				if len(pp.sudogcache) == 0 {
					pp.sudogcache = append(pp.sudogcache, new(sudog))
				}
			}
			n := len(pp.sudogcache)
			s := pp.sudogcache[n-1]
			pp.sudogcache[n-1] = nil
			pp.sudogcache = pp.sudogcache[:n-1]
			if s.elem != nil {
				throw("acquireSudog: found s.elem != nil in cache")
			}
			releasem(mp)
			return s
		}
	5. 销毁sudog的方法releaseSudog（E:/Enviroment/Go/go1.14.8/src/runtime/proc.go:358）
		func releaseSudog(s *sudog) {
			if s.elem != nil {
				throw("runtime: sudog with non-nil elem")
			}
			if s.isSelect {
				throw("runtime: sudog with non-false isSelect")
			}
			if s.next != nil {
				throw("runtime: sudog with non-nil next")
			}
			if s.prev != nil {
				throw("runtime: sudog with non-nil prev")
			}
			if s.waitlink != nil {
				throw("runtime: sudog with non-nil waitlink")
			}
			if s.c != nil {
				throw("runtime: sudog with non-nil c")
			}
			gp := getg()
			if gp.param != nil {
				throw("runtime: releaseSudog with non-nil gp.param")
			}
			mp := acquirem() // avoid rescheduling to another P
			pp := mp.p.ptr()
			//如果当前P的sudog队列满了，就扔一半到全局sudog队列(TODO Golang里对一半真是执着）
			if len(pp.sudogcache) == cap(pp.sudogcache) {
				// Transfer half of local cache to the central cache.
				var first, last *sudog
				for len(pp.sudogcache) > cap(pp.sudogcache)/2 {
					n := len(pp.sudogcache)
					p := pp.sudogcache[n-1]
					pp.sudogcache[n-1] = nil
					pp.sudogcache = pp.sudogcache[:n-1]
					if first == nil {
						first = p
					} else {
						last.next = p
					}
					last = p
				}
				lock(&sched.sudoglock)
				last.next = sched.sudogcache
				sched.sudogcache = first
				unlock(&sched.sudoglock)
			}
			//直接回收到P的sudog队列里
			pp.sudogcache = append(pp.sudogcache, s)
			releasem(mp)
		}
	5.为了查看扔数据发生阻塞时会发生什么，先看一下chan <- 1 的源码。TODO ep指的是1的指针
		func chansend(c *hchan, ep unsafe.Pointer, block bool, callerpc uintptr) bool {
			// 如果你往一个nil里扔数据，有两种情况：
				1.如果是非阻塞模式，则直接return false
				2.如果是阻塞模式，则调用gopark()使goroutine停止，此时连goroutine里的defer也不会执行
			if c == nil {
				if !block {
					return false
				}
				gopark(nil, nil, waitReasonChanSendNilChan, traceEvGoStop, 2)
				throw("unreachable")
			}

			if debugChan {
				print("chansend: chan=", c, "\n")
			}

			if raceenabled {
				racereadpc(c.raceaddr(), callerpc, funcPC(chansend))
			}
			//如果 非阻塞模式 && 通道未关闭 && (管道容量为0 || 管道已满)，则直接return false
			if !block && c.closed == 0 && ((c.dataqsiz == 0 && c.recvq.first == nil) ||
				(c.dataqsiz > 0 && c.qcount == c.dataqsiz)) {
				return false
			}

			var t0 int64
			if blockprofilerate > 0 {
				t0 = cputicks()
			}
			//TODO 走到这里，就证明有资格对管道进行操作了，此时就要上一把互斥锁
			lock(&c.lock)
			// 管道关闭后是只读的，因此直接解锁并抛异常
			if c.closed != 0 {
				unlock(&c.lock)
				panic(plainError("send on closed channel"))
			}
			// TODO 写数据进管道前，先看看有没有其他G因为读而阻塞，如果有，唤醒它，并把数据给它，而非通过管道传输。同时对通道解锁
			if sg := c.recvq.dequeue(); sg != nil {
				// Found a waiting receiver. We pass the value we want to send
				// directly to the receiver, bypassing the channel buffer (if any).
				// TODO 直接将数据给要读的Goroutine
				send(c, sg, ep, func() { unlock(&c.lock) }, 3)
				return true
			}
			//TODO 对于有缓冲管道，如果管道数据未满，就将数据放进通道的buffer里，同时解锁
			if c.qcount < c.dataqsiz {
				// Space is available in the channel buffer. Enqueue the element to send.
				//TODO 获取数据的位置，其实值和sendx一样
				qp := chanbuf(c, c.sendx)
				if raceenabled {
					raceacquire(qp)
					racerelease(qp)
				}
				//TODO 将数据拷贝到通道的buffer的qp下标里，注意是拷贝
				typedmemmove(c.elemtype, qp, ep)
				c.sendx++
				//TODO 因为是环形链表，所以判断下标是否越界了，越界了就重置
				if c.sendx == c.dataqsiz {
					c.sendx = 0
				}
				c.qcount++

				unlock(&c.lock)
				return true
			}

			// 如果是缓冲管道，直接解锁，并return false TODO 这里有个问题，如果向已满的缓冲管道扔数据，还是会阻塞啊
			// 上面的说法有误区，这里的意思是：走到这一步，说明(缓冲管道已满||是非缓冲管道），如果不需要阻塞，则直接返回错误。
			// TODO 那么问题来了，什么情况下不需要阻塞（有可能是select噢）？
			if !block {
				unlock(&c.lock)
				return false
			}

			// Block on the channel. Some receiver will complete our operation for us.
			// 获取当前Goroutine，类似Thread.currentThread()；
			gp := getg()
			//TODO 创建sudog并将当前Goroutine的信息传到sudog里，同时指定Goroutine的当前sudog
			mysg := acquireSudog()
			mysg.releasetime = 0
			if t0 != 0 {
				mysg.releasetime = -1
			}
			// No stack splits between assigning elem and enqueuing mysg
			// on gp.waiting where copystack can find it.
			mysg.elem = ep
			mysg.waitlink = nil
			mysg.g = gp
			mysg.isSelect = false
			mysg.c = c
			gp.waiting = mysg
			gp.param = nil
			// TODO 将sudog扔到管道的发送者阻塞队列里
			c.sendq.enqueue(mysg)
			// TODO 这一步很重要！！直接阻塞Goroutine，并解锁
			gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanSend, traceEvGoBlockSend, 2)
			// Ensure the value being sent is kept alive until the
			// receiver copies it out. The sudog has a pointer to the
			// stack object, but sudogs aren't considered as roots of the
			// stack tracer.
			// TODO 这一步目测是等待唤醒
			KeepAlive(ep)
			// TODO 能走到这一步，说明有人把阻塞的Goroutine唤醒了（读通道的协程把我唤醒了，并让我把数据传给它，传完后接下来我要做收尾工作，如释放sudog）
			// someone woke us up.
			if mysg != gp.waiting {
				throw("G waiting list is corrupted")
			}
			gp.waiting = nil
			gp.activeStackChans = false
			if gp.param == nil {
				if c.closed == 0 {
					throw("chansend: spurious wakeup")
				}
				panic(plainError("send on closed channel"))
			}
			gp.param = nil
			if mysg.releasetime > 0 {
				blockevent(mysg.releasetime-t0, 2)
			}
			mysg.c = nil
			releaseSudog(mysg)
			return true
		}
	6.再看看 <- chan的源码
		if debugChan {
			print("chanrecv: chan=", c, "\n")
		}

		if c == nil {
			if !block {
				return
			}
			gopark(nil, nil, waitReasonChanReceiveNilChan, traceEvGoStop, 2)
			throw("unreachable")
		}

		// TODO 如果是非阻塞模式，依旧直接返回了
		if !block && (c.dataqsiz == 0 && c.sendq.first == nil ||
			c.dataqsiz > 0 && atomic.Loaduint(&c.qcount) == 0) &&
			atomic.Load(&c.closed) == 0 {
			return
		}

		var t0 int64
		if blockprofilerate > 0 {
			t0 = cputicks()
		}
		// TODO 加锁
		lock(&c.lock)
		// TODO 是否已关闭且长度为0？(解锁，是否有值接收？返回零值：不做处理)：（继续往下走）
		if c.closed != 0 && c.qcount == 0 {
			if raceenabled {
				raceacquire(c.raceaddr())
			}
			unlock(&c.lock)
			if ep != nil {
				typedmemclr(c.elemtype, ep)
			}
			return true, false
		}
		//TODO 看看是否有发送者在阻塞？
		if sg := c.sendq.dequeue(); sg != nil {
			//TODO 这一步的处理是：
					1.如果为非缓冲通道，则直接将发送者的值拷贝到自己那 //TODO recvDirect()
					2.如果为缓冲通道，本协程取队首值，发送者将值放到队尾。
					3.TODO 最后goready发送者
			recv(c, sg, ep, func() { unlock(&c.lock) }, 3)
			return true, true
		}

		//TODO 如果是缓冲管道，且buf里有数据，则直接获取，并处理好下标
		if c.qcount > 0 {
		// Receive directly from queue
			qp := chanbuf(c, c.recvx)
			if raceenabled {
				raceacquire(qp)
				racerelease(qp)
			}
			if ep != nil {
				typedmemmove(c.elemtype, ep, qp)
			}
			typedmemclr(c.elemtype, qp)
			c.recvx++
			if c.recvx == c.dataqsiz {
				c.recvx = 0
			}
			c.qcount--
			unlock(&c.lock)
			return true, true
		}

		//照旧，如果非阻塞，直接返回
		if !block {
			unlock(&c.lock)
			return false, false
		}

		//TODO 走到这一步，就说明通道内没数据，需要打包成sudog阻塞
		// no sender available: block on this channel.
		gp := getg()
		mysg := acquireSudog()
		mysg.releasetime = 0
		if t0 != 0 {
			mysg.releasetime = -1
		}
		// No stack splits between assigning elem and enqueuing mysg
		// on gp.waiting where copystack can find it.
		mysg.elem = ep
		mysg.waitlink = nil
		gp.waiting = mysg
		mysg.g = gp
		mysg.isSelect = false
		mysg.c = c
		gp.param = nil
		c.recvq.enqueue(mysg)
		gopark(chanparkcommit, unsafe.Pointer(&c.lock), waitReasonChanReceive, traceEvGoBlockRecv, 2)

		// someone woke us up
		if mysg != gp.waiting {
			throw("G waiting list is corrupted")
		}
		gp.waiting = nil
		gp.activeStackChans = false
		if mysg.releasetime > 0 {
			blockevent(mysg.releasetime-t0, 2)
		}
		closed := gp.param == nil
		gp.param = nil
		mysg.c = nil
		releaseSudog(mysg)
		return true, !closed
*/

func main() {
	/**
	源码姑且是解读完了,现在来记录下Goroutine的具体阻塞步骤（先不考虑select的!block）
	如果是发送：
		if 缓冲管道{
			switch{
			case 通道已满（阻塞）：
				1.G从M上摘除，P将下一个G`交给M运行（和handOff基础有点不同，handOff是摘除P）
				2.G被摘除后会打包成sudog，注意这里还有对sudog队列的存取过程，然后挂在chan的sendq队列里(gopark调用scheduler)。此时G处于Waiting状态
					等待另外的G调用goready唤醒
			case 通道为空，有接收者sudog（唤醒）
				1.发送者G直接将值拷贝到接收者G的栈里面（此时不走通道了，TODO Go只有这种情况下才会操作不同Goroutine栈的数据,不过有个问题：网上说这时不会拿通道锁，但源码里写着拷贝过程中是持锁的？
				2.发送者唤醒接收者sudog（goready调用scheduler），使接收者G处于runnable状态，会被原P的runnext所指向（TODO 可以理解为插队？）
			}
		}else{
			switch{
			case 无接收者sudog（阻塞）：
				1.G从M上摘除，P将下一个G`交给M运行（和handOff基础有点不同，handOff是摘除P）
				2.G被摘除后会打包成sudog，注意这里还有对sudog队列的存取过程，然后挂在chan的sendq队列里(gopark调用scheduler)。此时G处于Waiting状态
					等待另外的G调用goready唤醒
			case 有接收者sudog（唤醒）：
				1.发送者G直接将值拷贝到接收者G的栈里面（此时不走通道了，TODO Go只有这种情况下才会操作不同Goroutine栈的数据,不过有个问题：网上说这时不会拿通道锁，但源码里写着拷贝过程中是持锁的？
				2.发送者唤醒接收者sudog（goready调用scheduler），使接收者G处于runnable状态，会被原P的runnext所指向（TODO 可以理解为插队？）
			}
		}


	如果是接收：
		if 缓冲管道{
			switch{
			case 通道已满，有发送者sudog（唤醒）：
				1.接收者拿buf队首的数据
				2.从sendq里拿一个sudog，将其elem拷贝到buf队尾里
				3.接收者调用goready，里面调用scheduler，将发送者G从Waiting状态变为Runnable状态，然后发送者会被原P的runnext所指向
			case 通道为空（阻塞）：
				1.G从M上摘除，P将下一个G`交给M运行
				2.G被摘除后会打包成sudog，注意这里还有对sudog队列的存取过程，然后挂在chan的recvq队列里(G通过调用gopark来调用scheduler)。此时G处于Waiting状态
					等待另外的G调用goready唤醒
			}
		}else{
			switch{
			case 无发送者sudog（阻塞）：
				1.G从M上摘除，P将下一个G`交给M运行（和handOff基础有点不同，handOff是摘除P）
				2.G被摘除后会打包成sudog，注意这里还有对sudog队列的存取过程，然后挂在chan的recvq队列里(G通过调用gopark来调用scheduler)。此时G处于Waiting状态
					等待另外的G调用goready唤醒
			case 有发送者sudog（唤醒）：
				1.发送者G直接将值拷贝到接收者G的栈里面（此时不走通道了，TODO Go只有这种情况下才会操作不同Goroutine栈的数据,不过有个问题：网上说这时不会拿通道锁，但源码里写着拷贝过程中是持锁的？
				2.接收者唤醒发送者sudog（goready调用scheduler），使发送者G处于Runnable状态，会被原P的runnext所指向（TODO 可以理解为插队？）
			}
		}
	*/
}
