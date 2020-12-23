package main

import (
	"sync"
)

/**
TODO	但是使用传统互斥锁，主线程无法得知协程什么时候处理完，因此主线程跑完就中断所有协程（其实还有waitgroup放啊）
	因此推荐使用Golang推荐的同步工具：管道Channel
*/

var (
	globalMap = make(map[int]int)
	lock      sync.Mutex
)

//func main() {
//	for i:=1;i<100;i++{
//		go method(i)
//	}
//	//我只能假定10秒，因为我根本不清楚协程什么时候完成
//	time.Sleep(10000)
//}

func method(n int) {
	res := 1
	for i := 1; i < n; i++ {
		res *= i
	}
	/**
	TODO 这个时候会出现对同一个map的并发写问题，可以用传统的互斥锁锁住，类似ReentrantLock
	*/
	lock.Lock()
	globalMap[n] = res
	lock.Unlock()
}
