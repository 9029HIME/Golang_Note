package main

import "fmt"

//TODO:Golang与Java的多态不同的是:Golang的多态体现在接口上，而非父子类
func main() {
	var user *User = new(User)
	var maker *Maker = new(Maker)
	useInterface(user)
	useInterface(maker)
	//多态数组
	var interfaceArray [2]Use = [2]Use{user, maker}
	fmt.Println(interfaceArray)
	//useObject(user) 这里就不行了，说明Golang多态只体现在接口上
	//TODO Golang中可以通过类型断言向下转型
	var use Use = user
	toUser, isOk := use.(*User) //将use强转回*User类型,还能多返回一个是否断言成功的布尔值(前提用:=接收)
	fmt.Println("是否断言成功？", isOk)
	fmt.Println(toUser)

}

type Use interface {
	do()
}

type Object struct {
}

type User struct {
	Object
}

type Maker struct {
	Object
}

func (maker *Maker) do() {

}

func (user *User) do() {
	fmt.Println("User类的do")
}

func useObject(object Object) {
	fmt.Println("父类作为传参的方法")
}

func useInterface(use Use) {
	fmt.Println("接口作为传参的方法")
}
