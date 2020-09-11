package factories

/**
TODO:Golang中没有构造函数，建议使用工厂模式来完善封装性
*/
type student struct {
	Name  string
	Grade string
}

func (stu *student) String() string {
	return stu.Name
}

/**
TODO:如果Student是student，那么在main包下无法引用这个结构体
	这个时候可以在其他包通过工厂模式生成结构体
*/

func GetStudent(name string, grade string) *student {
	return &student{name, grade}
}
