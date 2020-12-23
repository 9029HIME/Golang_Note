package main

import "fmt"

/**
Golang中方法概念与Java类似，即对象中的函数。
但Golang中需要作用在指定的数据类型上（即调用者）
*/
func main() {
	var h *Handler = new(Handler)
	h.name = "aaa"
	/**
	但是方法调用，本身也是值拷贝，除非方法指定的是该结构体指针
	如果方法指向指针，this无论是指针还是对象，都是引用传递
	如果方法指向对象，this无论是指针还是对象，都是值传递
	由方法的指向来决定是引用传递还是值传递
	*/
	h.doHomeWork(5)
	fmt.Println("")
	fmt.Println("改变后的名称", h.name)
}

type Handler struct {
	name string
}

/**
方法的定义
即：Handler这个结构体有一个方法叫doHomeWork
*/
func (h *Handler) doHomeWork(i int) (int, int) {
	fmt.Printf("%s的handler调用了doHomeWork方法%s", h.name, i)
	h.name = "更改后"
	return 1, 2
}
