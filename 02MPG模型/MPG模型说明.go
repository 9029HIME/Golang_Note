package _2MPG模型

/**
好处：减少了CPU对线程的切换消耗，从"CPU切换线程"变为"处理器切换协程"
*/

/**
Golang线程模型组件：
M：内核线程（Go并发编程第一版），也有说是用户线程（aceld），不过在Golang的线程模型中，内核线程与用户线程1：1，且更关心的是协程
因此两个说法都可
P：程序上下文，也可说为调度器
G：协程
M0:
GO:
Local Queue:处理器本地的Goroutine队列，用于存放本处理器（P）的G0创建出来的Goroutine，MAXSIZE=256
Global Queue:全局Goroutine队列，当Local Queue满后，会将Goroutine存放到这里
自旋线程：

*/

/**
MPG模型特性：
1.由于Goroutine是由P调度，在M上运行，因此同一个时间点内只能由P个协程并行工作，P的个数可以由GOMAXPROCS()函数设置
2.Work Stealing：当M并没有可运行的Goroutine时，首先会在Global Queue里拿协程，如果Global Queue没有则会从其他
M绑定的P里的Local Queue里偷一半Goroutine来运行，TODO 如果还是获取不到，M休眠，直到被唤醒
3.Goroutine要等待其他Goroutine主动让出时间片后才能执行
4.当M运行的G进行了阻塞操作：
	if 休眠M队列.size > 0
		在休眠M队列里唤醒一个新的M1，将M的P挂载到M1里(detach函数)
	else
		新建一个新的M1，将M的P挂载到M1里(detach函数)
5.当M的G阻塞完成，M唤醒后，M和G需要找到一个空闲P来继续执行G的代码，如果找不到，M进入休眠，G进入全局
*/
