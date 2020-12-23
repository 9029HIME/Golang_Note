package main

import (
	"encoding/json"
	"fmt"
)

/**
TODO Golang中对Json操作主要通过encoding/json包
*/
func main() {
	//TODO 序列化
	var object Object = Object{"好基友", "22"}
	fmt.Println("一开始的object：", object)
	//TODO json包作为其他包，引用其他包的属性，该属性必须声明为大写，否则该属性不参与json序列化
	bytes, err := json.Marshal(&object)
	fmt.Printf("返回的error是:%v,解析的json结果是:%v\n", err, string(bytes))
	var string2string map[string]string = make(map[string]string)
	string2string["A"] = "1"
	string2string["B"] = "2"
	string2string["C"] = "3"
	mapBytes, mapErr := json.Marshal(string2string)
	fmt.Printf("返回的error是:%v,解析的json结果是:%v\n", mapErr, string(mapBytes))

	//TODO 反序列化
	var toObject *Object = new(Object)
	json.Unmarshal(bytes, toObject)
	fmt.Println("反序列化后的object是：", toObject)
}

type Object struct {
	Name string `json:"更改后的Name"`
	Age  string `json:"age"`
}
