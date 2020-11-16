package _2MPG模型

/**
协程好处：减少了CPU对线程的切换消耗，从"CPU切换线程"变为"处理器（P）切换协程"，CPU对P切换协程是无感的，Golang的优异性在于处理器的优化，使得
Goroutine可以大量开辟，切换成本低（早期的Golang是没有处理器的，纯粹地多个线程在一个全局Goroutine队列里竞争资源）
*/

/**
Golang线程模型组件：
M：内核线程（Go并发编程第一版）
P：程序上下文，也可说为调度器
G：协程，大小在几KB，由M或其他G创建
M0:进程唯一，启动程序的主线程，为第一个M，保存在runtime中。启动G0，之后和其他M一样
GO:线程唯一，每创建一个M都会创建一个G，这个G就是G0。用来调度其他的G（如切换G），G0本身是不能执行函数的
Local Queue:处理器本地的Goroutine队列，用于存放本处理器（P）的G0创建出来的Goroutine，MAXSIZE=256
Global Queue:全局Goroutine队列，当Local Queue满后，会将Goroutine存放到这里,TODO 存取时有锁机制
自旋线程：当M和P执行完G后，发现并没有G可以执行了，此时M就会作为一个自旋线程等待新的G到来（短时间内不会销毁线程）
M Queue：存放休眠的内核线程。如果休眠太久的话，会被GC
P Queue：存放因M阻塞而脱离的P，唤醒、新建、阻塞完成的M会从P Queue里拿P

*/

/**
MPG模型特性：
1.由于Goroutine是由P调度，在M上运行，因此同一个时间点内只能由P个协程"并行"（注意，不是并发）工作，P的个数可以由GOMAXPROCS()函数设置

2.Goroutine要等待其他Goroutine主动让出时间片后才能执行 TODO GO13有优化
每个Goroutine最多占用10ms，当让出时间片后，会放回Local Queue，如果满了就放回Global Queue

3.Work Stealing：当M并没有可运行的Goroutine时（自旋线程），首先会在Global Queue里拿协程，如果Global Queue没有则会从其他
M绑定的P里的Local Queue里偷一半Goroutine来运行，TODO 如果还是获取不到，M休眠进去M Queue，直到被唤醒
从全局队列中获取G的数量是：min(len(Global Queue)/GOMAXPROCS+1 , len(Global Queue)/2),这一步是从全局队列到本地队列的负载均衡
拿到G后就不再是自旋线程了

4.Hand Off：当M运行的G进行了阻塞操作（如read，channel阻塞）：
	if 休眠M队列.size > 0
		在休眠M队列里唤醒一个新的M1，runtime将M的P挂载到M1里(detach函数)
	else
		新建一个新的M1，runtime将M的P挂载到M1里(detach函数)

5.当M的G阻塞完成，M和G需要找到一个空闲P，同时G放入这个空闲P的Local Queue。如果找不到，M进入M Queue，G进入Global Queue

6.	3、4、5是Golang中线程复用（M）的两个机制


*/
