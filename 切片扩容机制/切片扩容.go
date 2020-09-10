package main

import (
	"fmt"
)

/**
TODO:需要注意的是：
	1.如果切片直接切一个容量为cap的数组，切的长度为len。那么该切片的长度为len，容量为cap。只要len<=cap。append()就不会动态扩容
	2.就算一个切片的cap>len,当获取len<index<cap的切片index下标数据时，仍然会抛out of range错误，而非默认值
*/
func main() {
	var arr [3]int = [3]int{1, 2, 3}
	var slice []int = arr[0:3]
	slice[0] = 11111
	fmt.Println("扩容前更改slice第一个：", slice)
	fmt.Println("扩容前更改arr第一个：", arr)
	fmt.Println("扩容前的slice：", slice)
	fmt.Println("开始扩容")
	//TODO:注意！切片本身的地址没变过，扩容后变的只有数组地址
	slice = append(slice, 314)
	fmt.Println("扩容后的arr:", arr)
	fmt.Println("扩容后的slice:", slice)
	slice[2] = 33333
	fmt.Println("扩容后更改slice第3个", slice)
	//TODO:由此可见，动态扩容后切片指向的是新数组
	fmt.Println("扩容后更改arr第3个", arr)
}
