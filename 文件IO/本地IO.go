package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
)

//TODO Golang中对文件的IO也是基于输入流、输出流。主要是基于os.File结构体操作
func main() {

	//操作文件步骤：打开→读/写→关闭
	/**
	第一步 打开文件流，os包的Open函数,返回File对象指针
	*/
	filee, error := os.Open("C:\\Users\\Administrator\\Desktop\\Golang文件.txt")
	if error != nil {
		fmt.Println("打开文件错误:", error)
	}
	var file *os.File = filee

	///TODO 读文件内字符可以通过缓冲区Reader
	var reader *bufio.Reader = bufio.NewReader(file)
	for {
		//每一次只读到换行那
		bytes, error := reader.ReadString('\n')
		///TODO：表示读到文件末尾了
		if error == io.EOF {
			break
		}
		fmt.Println(bytes)
	}
	fmt.Println(file)

	/**
	最后一步，关闭文件流
	*/
	defer file.Close()
}
