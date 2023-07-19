package main

import "fmt"

func change(s *[]int) {
	//值传递还是引用传递
	*s = append(*s, 3)
}

func main() {
	slice := make([]int, 5)
	slice[0] = 1
	slice[1] = 2
	fmt.Println(slice)
	change(&slice)
	fmt.Println(slice)
}
