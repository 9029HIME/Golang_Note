package main

import (
	"fmt"
	"math/rand"
	"sort"
)

type Sorted struct {
	Name string
	Age  int
}

type SortedList []Sorted

//TODO 实现sort.Interface的三个方法Len() int、Less(i,j int) bool、Swap(i,j int)
func (list SortedList) Len() int {
	return len(list)
}

//年龄从小到大排序
func (list SortedList) Less(i, j int) bool {
	return list[i].Age < list[j].Age
}

func (list SortedList) Swap(i, j int) {
	//temp := list[i]
	//list[i] = list[j]
	//list[j] = temp
	list[i], list[j] = list[j], list[j]
}

func (list SortedList) DoSort() {
	for i := 0; i < 10; i++ {
		var sorted Sorted = Sorted{
			Name: fmt.Sprintf("元素%d", rand.Intn(10)),
			Age:  rand.Intn(100),
		}
		list = append(list, sorted)
	}
	fmt.Println("排序前：", list)
	sort.Sort(list)
	fmt.Println("排序后：", list)
}
