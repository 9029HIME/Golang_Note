package main

import "fmt"

/**
TODO 管道是用于各协程间的通信，Golang和其他编程语言如Java很大的不同是，Java是基于多个线程内存共享完成同步，
	而Golang是基于管道通信完成同步。Golang的主旨是：不要以共享内存的方式来通信，相反，要通过通信来共享内存。
	管道是线程安全的
*/
func main() {
	//创建一个可以存放5个string的管道 	TODO 容量无法动态扩容
	var stringChannel chan string = make(chan string, 5)
	//%v引用类型指向的地址，%p引用类型本身的地址
	fmt.Printf("%v,%p\n", stringChannel, &stringChannel)
	//TODO 管道写入数据
	stringChannel <- "数据1"
	fmt.Printf("容量:%v,长度%v\n", cap(stringChannel), len(stringChannel))
	//TODO 管道读取数据
	var result = <-stringChannel
	fmt.Printf("去除的值:%v,容量:%v,长度%v\n", result, cap(stringChannel), len(stringChannel))

}
