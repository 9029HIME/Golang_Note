package main

import "fmt"

/**
Golang中可以为结构体取别名，此时Golang会认为这是两个结构体，两者之间需要强转
*/
func main() {
	var zhangsan Student
	var lisi dick
	zhangsan = Student(lisi)
	fmt.Println(zhangsan)
}

type Student struct {
	Name  string
	Grade string
}

type dick Student
