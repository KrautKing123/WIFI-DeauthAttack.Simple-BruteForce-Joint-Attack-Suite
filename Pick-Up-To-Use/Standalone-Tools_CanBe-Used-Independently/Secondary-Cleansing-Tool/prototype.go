package main

import (
	"bufio"
	"fmt"
	"log"
	"flag"
	"os"
	"runtime"
	"sync"
	"time"
	"unicode"
	"github.com/schollz/progressbar/v3"
	"sort"
     "strconv"
)

// --- 数据结构定义 (保持不变) ---
type IndexInfo struct {
	LineNumber int
	CharIndex  int
}

type Job struct {
	LineNumber int
	LineText   string
}

type ProcessedLine struct {
	OriginalLine  string
	LineNumber    int
	LetterIndices []IndexInfo
	NumberIndices []IndexInfo
	SymbolIndices []IndexInfo
	LowerLetterIndices []IndexInfo
	UpperLetterIndices []IndexInfo
}

type ProcessedLineLettercaseSensitive struct {
	OriginalLine  string
	LineNumber    int
	LowerLetterIndices []IndexInfo
	UpperLetterIndices []IndexInfo
	NumberIndices []IndexInfo
	SymbolIndices []IndexInfo
}

type TaskFunc func([]IndexInfo) bool

var dispatchMap = map[int]TaskFunc{
	0: isCompact,
	1: isCouple,
	2: isEquallySpaced,
	3: isSymmetrical,
}


// --- 新的全局过滤函数 ---

// isCompact 检查单个种类的索引是否是连续的。
func isCompact(indices []IndexInfo) bool {
	if len(indices) <= 1 {
		return true
	}
	firstIndex := indices[0].CharIndex
	lastIndex := indices[len(indices)-1].CharIndex
	return (lastIndex - firstIndex) == (len(indices) - 1)
}

// shouldKeepLine 是我们的主过滤函数。
/**  func shouldKeepLine(letterIndices, numberIndices, symbolIndices []IndexInfo) bool {
	return isCompact(letterIndices) &&
		isCompact(numberIndices) &&
		isCompact(symbolIndices)
}    **/

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


func shouldKeepLine(lineStruct [][]IndexInfo, rulesInSlice [][]int, rulesAvgIntInSlice []float64,  modeIndicator int) bool {
	    if modeIndicator == 0 {
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
		   } else if modeIndicator == 1 {
			         }
     }

// --- 并发工人函数 (已修改) ---
func worker(id int, jobs <-chan Job, results chan<- ProcessedLine, wg *sync.WaitGroup, caseSensitive bool,  rulesInSlice [][]int, rulesAvgIntInSlice []float64, modeIndicator int) {
	defer wg.Done()
	// r := rand.New(...) // 我们不再需要随机数源了

	for job := range jobs {
		// 索引分类逻辑 (保持不变)
		lineLen := len(job.LineText)
		letterIndices := make([]IndexInfo, 0, lineLen)
		LowerLetterIndices := make([]IndexInfo, 0, lineLen)
		UpperLetterIndices := make([]IndexInfo, 0, lineLen)
		numberIndices := make([]IndexInfo, 0, lineLen)
		symbolIndices := make([]IndexInfo, 0, lineLen)

		for charIdx, char := range job.LineText {
			info := IndexInfo{LineNumber: job.LineNumber, CharIndex: charIdx}
			switch {
				case unicode.IsLetter(char):
					 letterIndices = append(letterIndices, info)
		    		 if unicode.IsLower(char) {
				   		LowerLetterIndices = append(LowerLetterIndices, info)
						} else if unicode.IsUpper(char) {
							   UpperLetterIndices = append(UpperLetterIndices, info)
					 }
				case unicode.IsNumber(char):
					 numberIndices = append(numberIndices, info)
				default:
					 symbolIndices = append(symbolIndices, info)
			}
		}

		if caseSensitive == true {
		   charIndices := make([][]IndexInfo, 0, 4)
		   charIndices[0] = LowerLetterIndices
		   charIndices[1] = UpperLetterIndices
		   charIndices[2] = numberIndices
		   charIndices[3] = symbolIndices
		   } else if caseSensitive == false {
				  charIndices := make([][]IndexInfo, 0, 3)
		   		  charIndices[0] = letterIndices
		   		  charIndices[1] = numberIndices
		   		  charIndices[2] = symbolIndices
		}

		// --- 核心修改：使用新的智能过滤函数替换随机判断 ---
		if shouldKeepLine(charIndices, rulesInSlice, rulesAvgIntInSlice, modeIndicator) {
			// 如果该行符合“紧凑”要求，则将其发送到结果通道
			results <- ProcessedLine{
				OriginalLine:  job.LineText,
				LineNumber:    job.LineNumber,
				LetterIndices: letterIndices,
				NumberIndices: numberIndices,
				SymbolIndices: symbolIndices,
			}
		}
		// 如果不符合要求，则像之前一样，什么都不做，该行被静默丢弃。
	}
}

func ContainsUsingBinarySearch(slice []int, target int) bool {
	// 1. 创建一个原始切片的副本，以避免修改它
	sliceCopy := make([]int, len(slice))
	copy(sliceCopy, slice)

	// 2. 对副本进行排序，这是二分查找的前提
	sort.Ints(sliceCopy)

	// 3. 使用 sort.SearchInts 进行二分查找
	// 它返回目标值应该被插入的位置
	index := sort.SearchInts(sliceCopy, target)

	// 4. 进行双重检查并返回结果
	// a) 检查索引是否在切片范围内
	// b) 检查该索引上的值是否确实是我们的目标值
	return index < len(sliceCopy) && sliceCopy[index] == target
}


// --- Main 函数 (无需修改) ---
func main() {
	RulesNumberAvailable := []int{0, 1, 2}
    // ... main 函数的所有内容都和上一个最终版本完全相同 ...
	// ... 它负责文件IO、进度条、启动并发流程 ...
	// ... 它不需要知道过滤逻辑是如何改变的，这就是解耦的好处 ...
	inputFile := flag.String("input-file", "password.txt", "input file name" )
	outputFile := flag.String("output-file", "kept_passwords.txt", "output file name" )
	filterRules := flag.String("filter-rules", "", "")
    avgIntRules := flag.String("avgInt-rules", "", "")
	caseSensitive := flag.Bool("case-sensitive", false, "")
	flag.Parse()


	if (*filterRules == "" && *avgIntRules == "") || (*filterRules != "" && *avgIntRules != "") { 
     // Must specify exactly one of --regex or --counts mode.
		fmt.Fprintln(os.Stderr, "Error: Must specify exactly one of --filter-rules or --avgInt-rules.")
		flag.Usage()
		os.Exit(1)
  	}

	rulesInGroups := strings.Split(*filterRules, ":")
	if len(rulesInGroups) != 4 {
		log.Fatalf("输入必须正好有4个由冒号分割的组，但检测到 %d 个", len(rulesInGroups))
	   }
    
	rulesInSlice := make([][]int, 4)
	// 4. 遍历这四个部分（即使部分是空字符串）
	for i, part := range rulesInGroups {
		// 为当前部分创建一个内层切片
		// 如果 'part' 是空字符串，len(part)为0，循环不执行，innerSlice 保持为空
		innerSlice := make([]int, 0, len(part))

		for _, char := range part {
			if char < '0' || char > '9' {
				// 如果字符不是数字，返回错误
				log.Fatalf("无效字符 '%c' 在组 '%s' 中", char, part)
			}
			// 将字符转换为整数并追加
			innerSlice = append(innerSlice, int(char-'0'))
		}
		// 将（可能为空的）内层切片赋值给结果的相应位置
		rulesInSlice[i] = innerSlice
	}

	for i, innerSlice := range rulesInSlice {
	    for i, number := range innerSlice {
		    if ContainsUsingBinarySearch(RulesNumberAvailable, number) == false {
			   fmt.Printf("规则数字表达式中包含了非法数字%d\n", number)
			   return 
			   }
		}
	}
    rulesAvgIntInGroups := strings.Split(*avgIntRules, ":")
    rulesAvgIntInSlice := make([]float64, 3)
    if len(rulesAvgIntInGroups) == 3 {
       for i, part := range rulesAvgIntInGroups {
            avgIntInFloat, err := strconv.ParseFloat(part, 64)
            if err != nil {
               log.Fatalf("输入的规则数字限制出现格式错误, 程序退出")
               } else {
                      rulesAvgIntInSlice[i] = avgIntInFloat
                 }
            }
       } else {
              log.Fatalf("你所提供的平均间隔数量限制的参数格式不符合要求")
            }
   
     var filterMode int
     if *filterRules != "" && *avgIntRules == "" {
        filterMode = 0
        } else if *avgIntRules != "" && *filterRules == "" {
               filterMode = 1
               }
     
                
          
     

	//const inputFile = "password.txt"
	//const outputFile = "kept_passwords.txt"
	numWorkers := runtime.NumCPU()

	log.Printf("开始智能清洗任务: %s -> %s (使用 %d CPU核心)", *inputFile, *outputFile, numWorkers)

	file, err := os.Open(*inputFile)
	if err != nil {
		log.Fatalf("打开输入文件 '%s' 失败: %v", *inputFile, err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		log.Fatalf("获取文件信息失败: %v", err)
	}
	fileSize := fileInfo.Size()

	outFile, err := os.Create(*outputFile)
	if err != nil {
		log.Fatalf("创建输出文件 '%s' 失败: %v", *outputFile, err)
	}
	defer outFile.Close()
	writer := bufio.NewWriter(outFile)
	defer writer.Flush()

	jobs := make(chan Job, numWorkers*100)
	results := make(chan ProcessedLine, numWorkers*100)
	bar := progressbar.NewOptions64(
		fileSize,
		progressbar.OptionSetDescription("正在处理..."),
		progressbar.OptionSetWriter(os.Stderr),
		progressbar.OptionShowBytes(true),
		progressbar.OptionSetWidth(15),
		progressbar.OptionThrottle(65*time.Millisecond),
		progressbar.OptionShowCount(),
		progressbar.OptionOnCompletion(func() { fmt.Fprint(os.Stderr, "\n") }),
		progressbar.OptionSpinnerType(14),
		progressbar.OptionFullWidth(),
	)

	var workerWg sync.WaitGroup
	for w := 1; w <= numWorkers; w++ {
		workerWg.Add(1)
		go worker(w, jobs, results, &workerWg, caseSensitive, rulesInSlice, rulesAvgIntInSlice, filterMode)
	}

	var collectorWg sync.WaitGroup
	collectorWg.Add(1)
	var keptCount, discardedCount int64
	go func() {
		defer collectorWg.Done()
		for result := range results {
			_, _ = writer.WriteString(result.OriginalLine + "\n")
			keptCount++
		}
	}()
	
	scanner := bufio.NewScanner(file)
	var totalLines int64
	for scanner.Scan() {
		totalLines++
		line := scanner.Text()
		jobs <- Job{LineNumber: int(totalLines), LineText: line}
		_ = bar.Add(len(line) + 1)
	}

	close(jobs)
	workerWg.Wait()
	close(results)
	collectorWg.Wait()

	discardedCount = totalLines - keptCount
	log.Println("----------- 清洗完成 -----------")
	log.Printf("总共处理了 %d 行。", totalLines)
	log.Printf("保留并写入 %d 行到 '%s'。", keptCount, *outputFile)
	log.Printf("丢弃了 %d 行。", discardedCount)
}
