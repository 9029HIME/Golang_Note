package _4Golang里的Map

/**
值得注意的是，Golang和Map和Java的HashMap结构并不一致，HashMap的桶指向链表/红黑树（1.8）。而Golang的桶指向[]bmap（底层结构是两个数组）
Golang里的map大致可分为两部分：hmap、bmap。由于和Java中HashMap的设计有很大不同，最好先了解结构，再了解过程。
*/

/**
TODO 先了解下hmap结构体，hmap即hashmap
type hmap struct {
	count     int	// 有多少对KV
	flags     uint8
	B         uint8 // 桶的个数为2^B，同时取key的低B位，用来定位key的桶下标
	noverflow uint16
	hash0     uint32 // 哈希计算时的会加入hash0以提高随机性
	buckets    unsafe.Pointer //桶数组，本质是[]bmap的指针
	oldbuckets unsafe.Pointer	// 用来存放扩容前的桶
	nevacuate  uintptr
	extra *mapextra
}
*/

/**
TODO 看一下桶结构
type bmap struct {
	// tophash generally contains the top byte of the hash value
	// for each key in this bucket. If tophash[0] < minTopHash,
	// tophash[0] is a bucket evacuation state instead.
	tophash [bucketCnt]uint8
	// Followed by bucketCnt keys and then bucketCnt elems.
	// NOTE: packing all the keys together and then all the elems together makes the
	// code a bit more complicated than alternating key/elem/key/elem/... but it allows
	// us to eliminate padding which would be needed for, e.g., map[int64]int8.
	// Followed by an overflow pointer.
}
TODO 表面上只有一个tophash属性，但在编译的时候会动态给这个结构体加点属性（目测是用了反射）
type bmap struct {
    topbits  [8]uint8 // key的hash值高8位，用来定位key的位置
    keys     [8]keytype // 用来存放key的数组，长度为8
    values   [8]valuetype // 用来存放value的数组，长度为8
    pad      uintptr
    overflow uintptr
}
TODO 可以看到，Golang中的K和V是存在不同的数组，但取的下标是一致的，如下标为X的K，其V的下标也为X。
*/

/**
TODO 在Java里，Object默认提供了hashCode()和equals()来做哈希运算功能，那Golang呢？其实在类型结构体里提供了，先看看类型结构体
type _type struct {
	size       uintptr
	ptrdata    uintptr // size of memory prefix holding all pointers
	hash       uint32
	tflag      tflag
	align      uint8
	fieldalign uint8
	kind       uint8
	alg        *typeAlg
	gcdata    *byte
	str       nameOff
	ptrToThis typeOff
}
TODO 其中typeAlg结构体提供了两个方法变量，hash用来计算Key的哈希值，equal用来判断两个值的哈希值是否相同？
type typeAlg struct {
	// (ptr to object, seed) -> hash
	hash func(unsafe.Pointer, uintptr) uintptr
	// (ptr to object A, ptr to object B) -> ==?
	equal func(unsafe.Pointer, unsafe.Pointer) bool
}
*/
