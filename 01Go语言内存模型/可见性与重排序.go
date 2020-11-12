package main

import (
	"fmt"
	"sync"
	"time"
)

/**
对应Java代码：
public class random {
	static int j = 0;	TODO 没有volatile
	public static void main(String[] args) throws InterruptedException{
		new Thread(()->{
            while(j == 0){

            }
            System.out.println("子线程跳出了循环");
        }).start();
        Thread.sleep(3000);
        j = 2;
        System.out.println("改变结束");
	}
}
结果：主线程的修改对子线程不可见

而Visible()的协程循环成功退出，Go语言的内存模型规定了一个goroutine可以看到另外一个goroutine修改同一个变量的值的条件
这类似java内存模型中内存可见性问题
--The Go memory model specifies the conditions under which reads of a variable in one goroutine can be
guaranteed to observe values produced by writes to the same variable in a different goroutine.

TODO 但是先要读操作r1对写操作w1可见，必须要求w1发生在r1前，在多协程的情况下，就要保证多个协程的读和写有序
*/
func Visible() {
	i := 1
	wg := sync.WaitGroup{}
	go func() {
		wg.Add(1)
		for i == 1 {

		}
		fmt.Println("协程退出循环")
		wg.Done()
	}()

	time.Sleep(time.Second * 3)
	i++
	wg.Wait()
}

/**
在第一个协程方法中，有可能因为指令重排序而先执行i=2，再执行j=3，在比较苛刻的条件下会输出j is 0
*/
func Resort() {
	i := 1
	j := 0
	go func() {
		j = 3
		i = 2
	}()
	go func() {
		fmt.Println("i is:", i)
		if i == 2 {
			//expect: 3
			fmt.Println("j is:", j)
		}
	}()

	time.Sleep(time.Second * 3)

}
