package main

/**
TODO Golang中单元测试的文件名必须以_test结尾
	测试用例必须以Test开头，且参数必须为(*testing.T)
*/
import (
	"fmt"
	"testing"
)

func TestHello(t *testing.T) {
	var resultChan chan int = make(chan int, 10)
	resultChan <- 1
	close(resultChan)
	for v := range resultChan {
		fmt.Println(v)
	}
}

func TestWorld(t *testing.T) {
	fmt.Println("World")
}
