package main

/**
Golang的接口与Java类似，但不像Java8那样有变量
TODO：Golang的接口是引用类型
*/
func main() {
	var user = new(InterfaceUser)
	println(useA(user))
	println(useB(user))
	//TODO:这里虽然声明传参是接口B，但接口B和接口A声明的方法一致，所以和println(useA(user))效果一样
	println(useAB(user))
	//TODO:切片排序，详情见“结构体切片排序”
	var list SortedList = make(SortedList, 5)
	list.DoSort()
}

/**
TODO:接口内直接定义方法
*/
type A interface {
	method1(var1 string) string
	method2(var2 string) string
}

type B interface {
	method1(var1 string) string
	method2(var2 string) string
	method3(var2 string) string
}

func useA(interfacee A) string {
	return interfacee.method1("a")
}

func useB(interfacee B) string {
	return interfacee.method3("a")
}

func useAB(interfacee B) string {
	return interfacee.method1("a")
}

type InterfaceUser struct {
	//TODO 还可以这样用，需要传入A接口的实现类
	A
}

/**
TODO：Golang和Java在实现上有很大的不同，Golang不需要显示声明，只需要结构体A定义的一些方法与接口A声明的所有方法一致，
		Golang就认为结构体A实现了接口A
	Q：如果有接口A、接口B，它们定义的方法一样，结构体C实现了接口A，那结构体也实现了接口B吗？
	A：是的，Golang会认为结构体C即实现了接口A又实现了接口B，调用接口B的方法与调用接口A的方法没区别
	Q:如果接口A有方法Method1、Method2。接口B有方法Method1、Method2、Method3。结构体C有什么实现方式？
	A：如果只实现Method1、Method2，Golang认为结构体C实现了接口A，如果实现了Method1、Method2、Method3，Golang会认为结构体C
		实现了接口A和接口B
*/
func (interfaceUser *InterfaceUser) method1(var1 string) string {
	return "方法1被调用了"
}

func (interfaceUser *InterfaceUser) method2(var1 string) string {
	return "方法2被调用了"
}

func (interfaceUser *InterfaceUser) method3(var1 string) string {
	return "方法3被调用了"
}
