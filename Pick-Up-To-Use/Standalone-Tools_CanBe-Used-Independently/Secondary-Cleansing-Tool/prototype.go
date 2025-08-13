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

type TaskFunc func()

dispatchMap := map[int]TaskFunc{
		0: isCompact,
		1: isCouple,
		2: isEquallySpaced,
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

func shouldKeepLine(lineStruct ProcessedLine, rulesInSlice [][]int) bool {

     }

// --- 并发工人函数 (已修改) ---
func worker(id int, jobs <-chan Job, results chan<- ProcessedLine, wg *sync.WaitGroup) {
	defer wg.Done()
	// r := rand.New(...) // 我们不再需要随机数源了

	for job := range jobs {
		// 索引分类逻辑 (保持不变)
		lineLen := len(job.LineText)
		letterIndices := make([]IndexInfo, 0, lineLen)
		numberIndices := make([]IndexInfo, 0, lineLen)
		symbolIndices := make([]IndexInfo, 0, lineLen)

		for charIdx, char := range job.LineText {
			info := IndexInfo{LineNumber: job.LineNumber, CharIndex: charIdx}
			if unicode.IsLetter(char) {
				letterIndices = append(letterIndices, info)
			} else if unicode.IsNumber(char) {
				numberIndices = append(numberIndices, info)
			} else {
				symbolIndices = append(symbolIndices, info)
			}
		}

		// --- 核心修改：使用新的智能过滤函数替换随机判断 ---
		if shouldKeepLine(letterIndices, numberIndices, symbolIndices) {
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

// --- Main 函数 (无需修改) ---
func main() {
    // ... main 函数的所有内容都和上一个最终版本完全相同 ...
	// ... 它负责文件IO、进度条、启动并发流程 ...
	// ... 它不需要知道过滤逻辑是如何改变的，这就是解耦的好处 ...
	inputFile := flag.String("input-file", "password.txt", "input file name" )
	outputFile := flag.String("output-file", "kept_passwords.txt", "output file name" )
	filterRules := flag.String("filter-rules", "", "")
	caseSensitive := flag.Bool("case-sensitive", false, "")
	flag.Parse()

	rulesInGroups := strings.Split(filterRules, ":")
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
		go worker(w, jobs, results, &workerWg)
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
