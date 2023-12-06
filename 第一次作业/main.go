package main

import (
	"fmt"
)

/*
作业：实现切片的删除操作
实现删除切片特定下标元素的方法。

要求一：能够实现删除操作就可以。
要求二：考虑使用比较高性能的实现。
要求三：改造为泛型方法
要求四：支持缩容，并且设计缩容机制。
*/

//通过创建新的切片，使用append方法来将指定下标两边的元素放在新切片中

func SliceDelete1(idx int, Slice []int) []int {

	if idx < 0 || idx >= len(Slice) {
		return Slice
	}

	newSlice := append(Slice[:idx], Slice[idx+1:]...)

	return newSlice

}

//通过使用循环来遍历指定下标后的元素，将指定下标后的元素都向前移动一位

func SliceDelete2(idx int, Slice []int) []int {

	if idx < 0 || idx >= len(Slice) {
		return Slice
	}

	for i := idx; i < len(Slice)-1; i++ {
		Slice[i] = Slice[i+1]
	}

	return Slice[:len(Slice)-1]

}

//改写为泛型方法

func SliceDelete3[T any](idx int, Slice []T) []T {
	if idx < 0 || idx >= len(Slice) {
		return Slice
	}

	for i := idx; i < len(Slice)-1; i++ {
		Slice[i] = Slice[i+1]
	}

	return Slice[:len(Slice)-1]
}

//缩容机制

func SliceDelete4[T any](idx int, Slice []T) []T {
	if idx < 0 || idx >= len(Slice) {
		return Slice
	}

	for i := idx; i > 0; i-- {
		Slice[i] = Slice[i-1]
	}

	return Slice[1:]
}

func main() {

	s1 := []int{1, 2, 3, 4}
	s2 := []int{1, 2, 3, 4}
	s3 := []string{"lip", "liplip", "lipliplip", "liplipliplip"}
	s4 := []string{"lip", "liplip", "lipliplip", "liplipliplip"}
	newSlice1 := SliceDelete1(0, s1)
	newSlice2 := SliceDelete2(0, s2)
	newSlice3 := SliceDelete3(0, s3)
	newSlice4 := SliceDelete4(0, s4)
	fmt.Printf("%v %d %d \n", newSlice1, len(newSlice1), cap(newSlice1))
	fmt.Printf("%v %d %d \n", newSlice2, len(newSlice2), cap(newSlice2))
	fmt.Printf("%v %d %d \n", newSlice3, len(newSlice3), cap(newSlice3))
	fmt.Printf("%v %d %d \n", newSlice4, len(newSlice4), cap(newSlice4))
}
