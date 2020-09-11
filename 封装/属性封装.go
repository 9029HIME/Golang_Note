package main

/**
Golang中可以将属性名小写，防止外界直接访问（类似private）
*/
func main() {
	student := new(Student)
	//TODO:但是本包还是可以直接访问，这是Golang的设计之一，封装性没那么强
	student.grade = "直接访问"
}

type Student struct {
	Name string
	//TODO:我要隐藏班级信息
	grade string
}

//TODO:方法名得大写开头，不然就是private了
func (s *Student) GetGrade() string {
	return s.grade
}
