package main

import "fmt"

/**
TODO:Golang支持多继承
*/
func main() {
	var son = new(Son)
	son.Name = "儿子的名字"
	son.Father.Name = "父亲的名字"
	//TODO:编译器先看Son有没有Name，没有的话就找嵌入结构体Father里有没有，再没有就找Father的嵌入结构体............
	println(son.Name)
	println(son.Father.Name)
	//TODO:如果子类访问父母类共有，子类没有的属性，需要明确指定访问的是父的还是母的
	//son.older = true	报错
	son.Father.older = true
	son.Mother.older = false
	son.iam()
}

type Father struct {
	Name  string
	Age   int
	older bool
}

type Mother struct {
	Name  string
	Age   int
	older bool
}

func (father *Father) iam() {
	fmt.Println("我是一个人")
}

/**
TODO:Golang中的继承类似嵌套，直接在结构体内声明父类
	Golang中的继承内容：属性与方法
		完全继承，不缺分大小写（即不区分private）
		同名，按就近原则，先访问自己的，如果要访问父类的，需要子.父.(属性/方法)
		父母同名自己没有，需要明确指定访问是父还是母的(属性/方法)
*/
type Son struct {
	Father
	Mother
	Name  string
	Grade string
}
