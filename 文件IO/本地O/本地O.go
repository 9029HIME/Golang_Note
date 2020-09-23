package main

import (
	"bufio"
	"os"
)

/**
TODO 打开操作主要使用os.OpenFile()函数
*/
func main() {
	/**
	TODO:三个参数详解：
		name:文件路径
		flag:打开方式，int常量，可以用或运算符|添加多个
		FileMode:权限控制（rwx,只有Linux与Unix有用）,在windows里可以随便填
	*/
	filePath := "C:\\Users\\Administrator\\Desktop\\Golang文件.txt"
	file, err := os.OpenFile(filePath, os.O_APPEND, 666)
	if err == nil {

		/**
		TODO 写操作实际主要使用带缓冲的bufio生成的Writer
		*/
		writer := bufio.NewWriter(file)
		writer.WriteString("新增加内容")
		/**
		TODO Writer是带缓存的，必须要flush()后才能将内容添加到文件内
		*/
		writer.Flush()
		/**
		TODO 手动打开的记得关闭
		*/
		file.Close()

	}
}
