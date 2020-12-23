package main

func main() {

}

func lengthOfLongestSubstring(s string) int {
	bytes := []byte(s)
	result := 0
	/**
	  使用两个头尾指针来处理
	  **/
	left := 0
	right := 0
	windows := make(map[byte]int, len(bytes))
	for left < len(s) && right < len(s) {
		value := bytes[right]
		if deleteIndex, ok := windows[value]; ok {
			for left <= deleteIndex {
				delete(windows, bytes[left])
				left++
			}
		} else {
			windows[value] = right
			right = right + 1
			if result <= (right - left) {
				result = (right - left)
			}
		}
	}
	return result

}
