package main

import (
	"fmt"
)

type TaskFunc func([]IndexInfo) bool

var dispatchMap = map[int]TaskFunc{
	0: isCompact,
	1: isCouple,
	2: isEquallySpaced,
}

type IndexInfo struct {
	LineNumber int
	CharIndex  int
}

func isCompact(lineStruct []IndexInfo) bool {
	var sum int = 0
	for i := 0; i < len(lineStruct); i++ {
		sum += lineStruct[i].CharIndex
	}
	if sum%2 == 0 {
		fmt.Printf("isCompact返回整除2\n")
		return true
	} else {
		fmt.Printf("isCompact返回不能整除2\n")
		return false
	}
}

func isCouple(lineStruct []IndexInfo) bool {
	var sum int = 0
	for i := 0; i < len(lineStruct); i++ {
		sum += lineStruct[i].CharIndex
	}
	if sum%2 == 1 {
		fmt.Printf("isCouple返回除于2余1\n")
		return true
	} else {
		fmt.Printf("isCouple返回整除2\n")
		return false
	}
}

func isEquallySpaced(lineStruct []IndexInfo) bool {
	var sum int = 0
	for i := 0; i < len(lineStruct); i++ {
		sum += lineStruct[i].CharIndex
	}
	if sum%3 == 0 {
		fmt.Printf("isEquallySpaced返回整除3\n")
		return true
	} else {
		fmt.Printf("isEquallySpaced返回不能整除3\n")
		return false
	}
}

func shouldKeepLine(lineStruct [][]IndexInfo, rulesInSlice [][]int) bool {
	for indexOfCharType, innerRulesSlice := range rulesInSlice {
		for i := 0; i <= len(innerRulesSlice); i++ {

			if i == len(innerRulesSlice) {
				fmt.Printf("当前范围内的规则全部不匹配, 强行退出\n")
				return false
			}
			if dispatchMap[innerRulesSlice[i]](lineStruct[indexOfCharType]) == true {
				fmt.Printf("当前范围内的规则已匹配, 退出当前剩余规则循环, 进入下一级循环\n")
				break
			}
		}
	}
	fmt.Printf("所有范围内的规则全部匹配, 运行成功\n")
	return true
}

func main() {
	groups := [][]IndexInfo{
		{{1, 1}, {1, 2}, {1, 3}},
		{{1, 5}, {1, 8}},
		{{1, 4}, {1, 6}, {1, 7}},
	}

	rules := [][]int{
		{0, 2},
		{1, 2},
		{0, 1, 2},
	}

	shouldKeepLine(groups, rules)

}
