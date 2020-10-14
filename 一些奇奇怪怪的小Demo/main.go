package main

import (
	"fmt"
	"time"
)

func main() {
	channel := method2()
	fmt.Printf("准备拿结果\n")
	i := <-channel
	fmt.Printf("结果是:%v\n", i)
	time.Sleep(time.Second * 1)
}

/**
在方法内启动协程，方法结束后协程是否会结束？
	TODO 不会,还得看主线程
*/
func method1() {
	go doGo1()
	fmt.Println("方法的事做完了")
}

func doGo1() {
	time.Sleep(time.Second * 6)
	fmt.Println("协程的事才做完")
}

/**
通道作为返回值会如何？
	TODO 如果返回值是<-chan，代表返回的是只读管道，只读管道不用close()
		如果返回值是chan<-，代表返回的是只写管道
		如果返回值是chan，代表返回的是双向管道
		注意，主线程在读取返回的管道时，如果数据还未写入，则会阻塞。
*/
func method2() <-chan int {
	channel := make(chan int, 5)
	go doGo2(channel)
	time.Sleep(time.Second * 10)
	fmt.Printf("method2睡完了\n")
	return channel
}

func doGo2(channel chan int) {
	time.Sleep(time.Second * 5)
	channel <- 1
	//TODO 如果不等待，写入数据到管道后，主线程的读管道阻塞就会结束，直接结束进程，下一句就不会执行
	fmt.Println("写入数据完毕")
}
