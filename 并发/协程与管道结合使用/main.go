package main

import (
	"fmt"
	"time"
)

var (
	total    = 10
	flagSize = 1
)

/**
TODO 值得注意的是，协程A往10个容量的带缓冲管道C里写50个数据，协程B在管道C里读数据。当C容量满后，协程A会阻塞，不再写数据。
	直到协程B在C读到数据后,C的容量-1，协程A才会停止阻塞，继续写数据。
	但如果没有协程B读数据，协程A就会一直阻塞，直到抛出死锁异常。
*/
func main() {
	inputChan := make(chan int, total)
	outputChan := make(chan bool, flagSize)
	go writeData(inputChan)
	go readData(inputChan, outputChan)
	//TODO 阻塞主线程
	for {
		flag := <-outputChan
		if flag == true {
			break
		}
	}
}

func writeData(input chan int) {
	for i := 1; i < 50; i++ {
		input <- i
		fmt.Println("写入的数据是：", i)
	}
	//TODO 写完后可以关闭channel，毕竟不影响读
	close(input)
}

func readData(input chan int, output chan bool) {
	for {
		time.Sleep(time.Second * 3)
		value, ok := <-input
		if !ok {
			break
		}
		fmt.Println("每3秒读一次，读到的数据是：", value)
	}
	output <- true
	close(output)
}
