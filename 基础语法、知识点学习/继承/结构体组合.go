package main

import "fmt"

//TODO:结构内引用结构体，叫组合，不叫继承(和Java类内用到其他类的对象一样)
func main() {
	var boss *Boss = new(Boss)

	boss.worker.do()
}

type Boss struct {
	Company string
	worker  Worker
}

type Worker struct {
	Name       string
	Department string
}

func (worker *Worker) do() {
	fmt.Println("打工咯")
}

func (boss *Boss) watch() {
	println("监督你们打工")
}
