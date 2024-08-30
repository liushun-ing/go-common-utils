package main

import (
	"fmt"
	"testing"
)

func TestSlice(t *testing.T) {
	cl := []int{1, 2, 3, 4, 5, 6}
	fmt.Println(&cl)
	index := 2
	dcl := append(cl[:index], cl[index+1:]...)
	fmt.Println(len(cl), cap(cl))
	fmt.Println(len(dcl), cap(dcl))
	fmt.Println(&cl, &dcl)

	cl2 := []int{1, 2, 3, 4, 5, 6}
	fmt.Println(&cl2)
	last := len(cl2) - 1
	cl2[index], cl2[last] = cl2[last], cl2[index]
	dcl2 := cl2[:last]
	fmt.Println(len(cl2), cap(cl2))
	fmt.Println(len(dcl2), cap(dcl2))
	fmt.Println(&cl2, &dcl2)
	// append 方法是以第一个参数的切片为基础去进行追加的
	// 所以第一种方法对切片来说，数据会变化
	// 但是第二种方法，只是顺序变了，但是数据不会改变
	//&[1 2 3 4 5 6]
	//6 6
	//5 6
	//&[1 2 4 5 6 6] &[1 2 4 5 6]
	//&[1 2 3 4 5 6]
	//6 6
	//5 6
	//&[1 2 6 4 5 3] &[1 2 6 4 5]
}
