package _4Golang里的Map

/**
值得注意的是，Golang和Map和Java的HashMap结构并不一致，HashMap的桶指向链表/红黑树（1.8）。而Golang的桶指向[]bmap（底层结构是两个数组）
Golang里的map大致可分为两部分：hmap、bmap，

*/

/**
type hmap struct {
	count     int
	flags     uint8
	B         uint8
	noverflow uint16
	hash0     uint32
	buckets    unsafe.Pointer
	oldbuckets unsafe.Pointer
	nevacuate  uintptr
	extra *mapextra
}
*/
