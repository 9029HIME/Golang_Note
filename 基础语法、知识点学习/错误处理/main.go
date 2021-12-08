package main

import (
	"errors"
	"fmt"
	"log"
)

func main() {

	// 通过异常处理，控制代码流程
	err := func(param int) (err error) {
		fmt.Println("调用下游接口")
		if param%2 == 0 {
			err = errors.New("参数错误")
		}
		return err
	}(2)

	if err != nil {
		log.Fatal(err)
		return
	}

	fmt.Println("下游接口没报错，继续执行")

}
