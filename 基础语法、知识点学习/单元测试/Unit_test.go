package main

/**
TODO Golang中单元测试的文件名必须以_test结尾
	测试用例必须以Test开头，且参数必须为(*testing.T)
*/
import (
	"fmt"
	"testing"
	"time"
)

func TestHello(t *testing.T) {
	s := "abba"
	bytes := []byte(s)
	result := 0
	/**
	  使用两个头尾指针来处理
	  **/
	left := 0
	right := 0
	windows := make(map[byte]int, len(bytes))
	for left < len(s) && right < len(s) {
		value := bytes[right]
		if deleteIndex, ok := windows[value]; ok {
			for left <= deleteIndex {
				delete(windows, bytes[left])
				left++
			}
		} else {
			windows[value] = right
			right = right + 1
			if result <= (right - left) {
				result = (right - left)
			}
		}
	}
	fmt.Println(result)
}

func TestWorld(t *testing.T) {
	fmt.Println("World")
}

func TestUintSub(t *testing.T) {
	//00000001
	var totalPrice uint8 = 1
	//00000010
	var couponPrice uint8 = 2
	//无符号数相减，也要先转为补码，但结果却是作为无符号补码。即255，而非-1（11111111[补] -> 11111110[反] → 10000001[原]）
	//1 + -2 = 00000001 + 10000010[原]= 00000001 + 11111101[反]= 00000001 + 11111110[补]= 11111111[补]
	//2 + -1 = 00000010 + 10000001[原]= 00000010 + 11111110[反]= 00000010 + 11111111[补]= 100000001[补]
	fmt.Println("sum: ", totalPrice-couponPrice)
}

func TestUint8Plus(t *testing.T) {
	var a uint8 = 255
	var b uint8 = 1
	//255 + 1 =  11111111 + 00000001 =
	fmt.Println("sum:", a+b)
}

func TestCacheChannel(t *testing.T) {
	channel := make(chan int, 2)
	channel <- 1
	channel <- 2
	go func() {
		time.Sleep(time.Second * 5)
		<-channel
		fmt.Println("5秒后取出了一个")
	}()
	fmt.Println("准备放3")
	channel <- 3
	fmt.Println("放3了")
	time.Sleep(time.Second * 7)
}
func TestNilChannel(t *testing.T) {
	var channel chan int
	go func() {
		<-channel
		defer fmt.Println("协程的defer")
	}()
	time.Sleep(time.Second * 3)

}
