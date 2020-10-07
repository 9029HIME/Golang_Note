package main

import (
	"fmt"
	"reflect"
)

/**
TODO golang中的反射是基于reflect包实现的
*/
func main() {
	student := &Student{"黄俊彦", 18}
	doReflect(student)
	i := 1
	reflectChangeNormalValue(&i)
	fmt.Printf("反射改变后的值%v\n", i)
	reflectChangeStructValue(student)
	fmt.Println(student)
	reflectDoStructMethod(student)
}

type Student struct {
	Name string `fuck:"you"`
	Age  int64
}

//TODO 反射的入口一般是空接口类型
func doReflect(i interface{}) {
	var reflectType = reflect.TypeOf(i)
	fmt.Printf("转换成的ref.type:%v,type.name是%v\n", reflectType, reflectType.Name())
	//将空接口转为value
	var reflectValue = reflect.ValueOf(i)
	fmt.Println("转换成的ref.value:", reflectValue)
	//再将value转为空接口
	var inter = reflectValue.Interface()
	fmt.Printf("转换成的空接口:%v,接口真正的类型是:%T\n", inter, inter)
	//再将空接口转回原来的类型
	student, ok := inter.(*Student)
	if ok {
		fmt.Println("最终又转换回来的结果是：", student)
	}

}

/**
TODO 通过反射更改基本类型值
	如果是指针类型，还需要将reflect.value调用elem()方法，返回"指针指向的值"的reflect.value
	如果不elem()，则是值拷贝，本质只在方法里改变了值而已。外部的变量值仍不变
*/
func reflectChangeNormalValue(i interface{}) {
	reflectValue := reflect.ValueOf(i)
	reflectValue.Elem().SetInt(64)
}

/**
TODO 通过反射改变结构体的值
*/
func reflectChangeStructValue(o interface{}) {
	fmt.Println("============开始通过反射改变结构体值============")
	reflectType := reflect.TypeOf(o)
	reflectValue := reflect.ValueOf(o)
	numField := reflectValue.Elem().NumField()
	for i := 0; i < numField; i++ {
		//TODO 获取第i-1个属性的值
		field := reflectValue.Elem().Field(i)
		//TODO 获取第i-1个属性的名reflectType.Elem().Field(i).Name
		fmt.Printf("属性名是：%v,属性值是：%v\n", reflectType.Elem().Field(i).Name, field)
		//TODO 这一步通过reflectType获取到第i-1个属性的tag（因为o是指针，所以记得Elem()）
		tagValue := reflectType.Elem().Field(i).Tag.Get("fuck")
		if tagValue != "" {
			fmt.Printf("tag值是:%v\n", tagValue)
		} else {
			fmt.Printf("结构体%v的属性%v没有tag值\n", reflectType.Elem().Name(), reflectType.Elem().Field(i).Name)
		}
		if field.Kind() == reflect.String {
			field.SetString("更好后的黄俊艳")
		}
		if field.Kind() == reflect.Int64 {
			field.SetInt(200)
		}
	}
	fmt.Println("============反射改变结构体值结束============")
}

/**
TODO 通过反射调用结构体方法
*/
func reflectDoStructMethod(o interface{}) {
	fmt.Println("============开始通过反射调用结构体方法============")
	reflectType := reflect.TypeOf(o).Elem()
	reflectValue := reflect.ValueOf(o).Elem()
	numMethod := reflectValue.NumMethod()
	/**
	TODO 注意！反射默认只拿大写开头的方法！！！
		如果方法的调用体不是指针，则elem()后才能获得
		如果调用体是指针，则elem()前才能获得
	*/
	fmt.Printf("结构体%v一共有%v个方法\n", reflectType.Name(), numMethod)
	for i := 0; i < numMethod; i++ {
		fmt.Printf("准备调用结构体%v第%v个方法，方法名：%v\n", reflectType.Name(), i+1, reflectType.Method(i).Name)
		//TODO Golang中反射方法数组排序是根据函数名的ASCII码从小到大排序
		reflectValue.Method(i)
		//TODO Golang中反射调用方法传参用Value切片，返回值也是Value切片
	}
	fmt.Println("============反射调用结构体方法结束============")
}

func (student Student) Print(content string) string {
	fmt.Printf("我是学生%v，我要介绍的内容是:%v\n", student.Name, content)
	return "调用完毕"
}
