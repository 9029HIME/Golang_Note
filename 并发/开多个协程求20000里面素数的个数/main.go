package main

import (
	"fmt"
	"runtime"
	"time"
)

var (
	cpu  = runtime.NumCPU()
	nums = 100000
)

func main() {
	var inputChan chan int = make(chan int, nums)
	var resultChan chan int = make(chan int, nums)
	var exitChan chan bool = make(chan bool, 4)
	t1 := time.Now()
	go writeResult(inputChan)
	for i := 1; i <= cpu; i++ {
		go judge(inputChan, resultChan, exitChan)
	}
	for i := 1; i <= cpu; i++ {
		//TODO 如果拿不到会阻塞,达到主线程阻塞的效果
		<-exitChan
	}
	close(resultChan)
	close(exitChan)

	durGoroutine := time.Since(t1)
	t2 := time.Now()
	//TODO 用单线程算
	all := 0
	for i := 1; i <= nums; i++ {
		if isPrimeNum(i) {
			all++
		}
	}
	durMain := time.Since(t2)
	fmt.Println("用四个协程算出1到20000素数的个数是:", len(resultChan))
	fmt.Println("用四个协程算出1到20000素数的时间是:", durGoroutine)
	fmt.Println("一个线程算出1到20000素数的个数是:", all)
	fmt.Println("一个线程算出1到20000素数的时间是:", durMain)

}

func writeResult(input chan int) {
	for i := 1; i <= nums; i++ {
		input <- i
	}
	close(input)
}

func judge(input chan int, result chan int, exit chan bool) {
	for {
		//TODO 很重要的一点，如果ok是false，那么拿到的值是默认值
		num, ok := <-input
		//TODO 如果取不出了，直接跳出循环
		if !ok {
			break
		}
		if isPrimeNum(num) {
			//fmt.Println("准备放入的素数是：",num)
			result <- num
		}
	}
	//TODO 因为有四个协程共同工作，不能因为其中一个协程停止了就关闭管道
	exit <- true
}

func isPrimeNum(num int) bool {
	if num == 1 {
		return false
	}
	for i := 2; i < num; i++ {
		if num%i == 0 {
			return false
		}
	}
	return true
}
