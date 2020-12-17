package main

import "fmt"

//append后其实返回一个新的切片，只是底层数组的指针一致(不产生扩容的话)
func dualSlice() {
	slice := make([]int, 5, 10)
	newSlice := append(slice, 1)
	/**
	这样打印的是底层数组地址
	*/
	//fmt.Printf("%p,%v\n",slice,len(slice))
	//fmt.Printf("%p,%v\n",newSlice,len(newSlice))
	fmt.Printf("slice的结果是:%v,内存地址是：%p,长度是%v\n", slice, &slice, len(slice))
	fmt.Printf("newSlice的结果是：%v,内存地址是：%p,长度是%v\n", newSlice, &newSlice, len(newSlice))
	/**
	接受一下append返回的切片
	*/
	slice = append(slice, 2)
	fmt.Printf("slice的结果是:%v,内存地址是：%p,长度是%v\n", slice, &slice, len(slice))
}
