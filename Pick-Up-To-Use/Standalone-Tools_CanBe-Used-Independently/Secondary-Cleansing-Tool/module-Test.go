package main

import (
	"fmt"
)

type TaskFunc func([]IndexInfo) bool

type IndexInfo struct {
	LineNumber int
	CharIndex  int
}

var dispatchMap = map[int]TaskFunc{
	0: isCompact,
	1: isCouple,
	2: isEquallySpaced,
	3: isSymmetrical,
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
	actualLen := countNonEmptyStructs(lineStruct)
	roundsToCount := actualLen - 1
	startIndex := 0
	for i := 0; i < roundsToCount; i += 2 {
		spacedOne := lineStruct[startIndex+1].CharIndex - lineStruct[startIndex].CharIndex
		spacedTwo := lineStruct[startIndex+2].CharIndex - lineStruct[startIndex+1].CharIndex
		if spacedOne != spacedTwo {
			return false
		}
		startIndex += 1
	}
	return true
}

func countNonEmptyStructs(slice []IndexInfo) int {
	var count int = 0
	// 创建一个零值实例用于比较
	var zeroInfo IndexInfo
	// 你也可以直接写 if item == (IndexInfo{})

	for _, item := range slice {
		// 如果当前项不等于零值，我们就计数
		if item != zeroInfo {
			count++
		}
	}
	return count
}

func isSymmetrical(lineStruct []IndexInfo) bool {
	actualLen := countNonEmptyStructs(lineStruct)
	lastIndex := actualLen - 1
	roundsToCount := (actualLen / 2) + (actualLen % 2)
	if roundsToCount == 1 {
		return true
	}
	startIndex := 0
	for i := 0; i < roundsToCount; i += 2 {
		combinationOne := lineStruct[startIndex].CharIndex + lineStruct[lastIndex].CharIndex
		combinationTwo := lineStruct[startIndex+1].CharIndex + lineStruct[lastIndex-1].CharIndex
		if combinationOne != combinationTwo {
			return false
		}
		startIndex += 1
		lastIndex -= 1
	}
	return true
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
		{{1, 4}, {1, 8}, {1, 12}, {1, 16}},
		{{1, 4}, {1, 6}, {1, 8}, {1, 11}},
	}

	rules := [][]int{
		{2},
		{2},
		{2},
	}

	if shouldKeepLine(groups, rules) {
		fmt.Printf("成功\n")
	} else {
		fmt.Printf("失败\n")
	}
}
