package _5当一个G因通道阻塞的时候_他在干什么_

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
		channel <- 1
	}()
	time.Sleep(time.Second*3)
	<-channel
过程：
	1.先看一下channel结构体源码（E:/Enviroment/Go/go1.14.8/src/runtime/chan.go:32）
		type hchan struct {
			qcount   uint           // total data in the queue 通道里的数据个数，类似length
			dataqsiz uint           // size of the circular queue 通道的长度，目测类似cap
			buf      unsafe.Pointer // points to an array of dataqsiz elements 数据是存储在环形队列，这是队列的头指针
			elemsize uint16			// 未知
			closed   uint32			// 通道是否被关闭（关闭后变成只读通道）
			elemtype *_type // element typ	通道内数据的类型
			sendx    uint   // send index	通道发送操作对应的buf下标
			recvx    uint   // receive index 通道接收操作对应的buf下标
			recvq    waitq  // list of recv waiters TODO 等到对通道进行读操作的Goroutine队列
			sendq    waitq  // list of send waiters	TODO 等待对通道进行写操作的Goroutine队列

			// lock protects all fields in hchan, as well as several
			// fields in sudogs blocked on this channel.
			//
			// Do not change another G's status while holding this lock
			// (in particular, do not ready a G), as this can deadlock
			// with stack shrinking.
			lock mutex //channel的锁
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
*/

func main() {

}
