package main

import (
	"Study/factory_no_chinese/factories"
	"fmt"
)

func main() {
	//不过用工厂模式，就不能在变量名后面强调类型了
	var stu = factories.GetStudent("黄俊严", "16网络工程2班")
	fmt.Println(stu)
}
