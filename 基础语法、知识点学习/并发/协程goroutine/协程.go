package main

import (
	"fmt"
	"runtime"
	"time"
)

/**
TODO Golang中的并发与并行，是基于协程进行，而非线程
	进程：是系统进行资源分配和调度的基本单位，即正在运行的程序
	线程：进程的其中一个执行实例，是程序执行的最小单元，线程是属于进程的。

TODO 并发：多个线程在单个CPU上运行,在操作系统微观角度下，CPU在不停的切换线程，同一个时间片内，只有一个线程在运行。
	 并行：多个线程在多个CPU上运行，在操作系统微观角度下，同一个时间片内，有多个线程在运行
	 传统的编程语言，多线程一般采取并发作业，而Golang可以采取并行作业（所以说Golang充分发挥了计算机性能）

TODO 但是Java具体分配任务到各个内核中去执行的并非JAVA与JVM而是操作系统.也就是说,你所执行的多线程
	可能会被分配到同一个CPU内核中运行.也可能非配到不同的cpu中运行.如果可以控制CPU的分配,那也应该是操作系统的api才能实现的了
	是liunx或是windos ，就有用户线程和内核线程的说法，如果是用户线程，OS是无法感知的，真正的多cpu处理是处理的内核级线程。
	Java使用的是一对一线程模型，所以它的一个用户线程对应于一个内核线程，调度完全交给操作系统来处理（即并行还是并发，交给OS）
	Golang使用的是多用户线程对多内核线程模型，可以由代码层面控制并行还是并发，这也是其高并发的原因，它的线程模型与Java中的
	ForkJoinPool非常类似；更详细的Java，Golang用户态与内核态线程区别，可以看https://juejin.im/post/6844903957664366600
*/
func main() {
	/**
	TODO 而在Golang中，协程可以看作是一个轻量级线程，一个线程可以开启多个协程，每个协程有自己独立的栈空间
		线程是由操作系统调度的，而协程可以代码控制调度（这部分有待完善）
		Golang中协程的基本调度单位是函数或方法，关键字为go
		注意，默认状态下，主线程结束，随之派生的协程也会结束
	*/
	go goroutineMethod()
	for i := 1; i < 10; i++ {
		fmt.Println("主线程方法")
		time.Sleep(time.Second)
	}
	//打印有多少核CPU
	fmt.Println(runtime.NumCPU())
	//设置Golang运行时的CPU数，TODO Golang1.8以后可以不用设置
	runtime.GOMAXPROCS(4)

}

func goroutineMethod() {
	for i := 1; i < 10; i++ {
		fmt.Println("协程方法")
		time.Sleep(time.Second)
	}
}
