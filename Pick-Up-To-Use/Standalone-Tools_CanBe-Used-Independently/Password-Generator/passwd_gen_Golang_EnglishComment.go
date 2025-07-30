package main

import (
	"bufio" // OPTIMIZATION: Import bufio to optimize file writing
	"flag"
	"fmt"
	"math"
	"os"
	"regexp/syntax"
	"runtime" // OPTIMIZATION: Import runtime to get the number of CPU cores
	"strconv"
	"strings"
	"sync"
	"sync/atomic" // OPTIMIZATION: Import atomic for lock-free counting. Uses atomic.AddInt64 and atomic.LoadUint64 for progress bar printing and rapid updates of the total password count variable.
	"time"
)

// --- Global Constants and Configuration ---
var (
	MAX_TOTAL_LENGTH            int // Limits the maximum password length
	MAX_CONCURRENT_PERMUTATIONS int // Limits the number of goroutines for concurrent permutation tasks in "no-repeat" mode
)

// --- Global Variables for Progress Bar ---
var (
	lastPrintTime sync.Once       // Use sync.Once to ensure it's initialized only once. This is essentially a sync.Once struct.
	lastTime      time.Time       // Stores the actual time
	printInterval = 1 * time.Second // The refresh interval for the progress bar
)

// --- Structs and Functions for Counts Mode ---

// CharSpec defines a specific character type and its required quantity.
type CharSpec struct {
	Chars string // Character set string, e.g., "a-z", "!%"
	Count int    // The required quantity
	Runes []rune // The parsed slice of runes
}

// parseCharSet parses character set strings like "a-z", "!%", "abc"; it returns a slice of runes based on the user-described format. It essentially stores Unicode code points, which can be iterated through by continuous incrementing.
func parseCharSet(s string) ([]rune, error) {
	if strings.Contains(s, "-") && len(s) == 3 && s[1] == '-' {
		start, end := rune(s[0]), rune(s[2]) // Convert characters to Unicode code points
		if start > end {
			return nil, fmt.Errorf("invalid char range: %s (start char '%c' is greater than end char '%c')", s, start, end)
		}
		var runes []rune
		for i := start; i <= end; i++ {
			runes = append(runes, i)  // Incrementally add all corresponding Unicode code points to the slice
		}
		return runes, nil
	}
	return []rune(s), nil
}

// calculateFactorial calculates the factorial
func calculateFactorial(n int) float64 {
	if n < 0 { return 0 }
	res := 1.0
	for i := 2; i <= n; i++ {
		res *= float64(i)
	}
	return res
}

// calculateCombinations calculates the number of combinations C(n, k)
func calculateCombinations(n, k int) float64 {
	if k < 0 || k > n { return 0 }
	if k == 0 || k == n { return 1 }
	if k > n/2 { k = n - k }
	res := 1.0
	for i := 1; i <= k; i++ {
		res = res*float64(n-i+1) / float64(i)
	}
	return res
}

// calculateExpectedCountForCountsModeNoRepeat estimates the total count for countsMode without character repetition.
func calculateExpectedCountForCountsModeNoRepeat(specs []CharSpec, totalLen int) float64 {
	var totalCombinationsOfBags float64 = 1.0
	for _, spec := range specs {  // Calculate the number of combinations (not permutations) for each character set spec subset
		if len(spec.Runes) < spec.Count { return 0 }
		totalCombinationsOfBags *= calculateCombinations(len(spec.Runes), spec.Count)
	} // Multiply the number of non-repeating combinations from different character set spec subsets
	return totalCombinationsOfBags * calculateFactorial(totalLen) // Multiply the number of selected combinations by the number of permutations to get the final total
}

// calculateExpectedCountForCountsModeWithRepetition estimates the total password count for countsMode with character repetition, calculated based on the array of CharSpec structs.
func calculateExpectedCountForCountsModeWithRepetition(specs []CharSpec, totalLen int) float64 {
	positions := calculateFactorial(totalLen) // First, calculate the total number of permutations assuming each character is unique
	for _, spec := range specs {
		positions /= calculateFactorial(spec.Count)
	}  // Then, divide by the number of permutations within the same character type. This abstractly removes internal combinations of the same type, resulting in the number of permutations of different character set types, which can be called the number of templates.
	var contentChoices float64 = 1.0
	for _, spec := range specs {
		contentChoices *= math.Pow(float64(len(spec.Runes)), float64(spec.Count))
	}  // Calculate all possible combinations for filling the template. This involves exponentiation within the same-type subsets of the template, because repetition is allowed.
	return positions * contentChoices  // Finally, multiply the number of template combinations by the filling possibilities
}

// parseCountsPattern parses a string like "a-z:3,!:1,%:1" and processes it into a slice of CharSpec structs for later generation.
func parseCountsPattern(pattern string, allowRepeat bool) ([]CharSpec, int, error) {
	parts := strings.Split(pattern, ",")
	var specs []CharSpec
	totalCount := 0
	for _, part := range parts {
		subParts := strings.Split(part, ":")
		if len(subParts) != 2 {
			return nil, 0, fmt.Errorf("invalid count part format: %s. Expected 'charset:count'", part)
		}
		charSetStr, countStr := subParts[0], subParts[1]
		count, err := strconv.Atoi(countStr) // Convert the character length of the subset, input by the user, to an int format
		if err != nil || count < 0 {
			return nil, 0, fmt.Errorf("invalid count number: %s", countStr)
		}   // Invalid negative length for a character subset
		runes, err := parseCharSet(charSetStr)  // Convert user-specified character subset descriptions like 'a-z' into actual character subsets
		if err != nil {
			return nil, 0, fmt.Errorf("invalid char set in '%s': %v", charSetStr, err)
		}
		if len(runes) == 0 && count > 0 {
			return nil, 0, fmt.Errorf("empty char set '%s' requested with count %d", charSetStr, count)
		}  // If an empty character set is provided, report an error
		if !allowRepeat && len(runes) < count {
			return nil, 0, fmt.Errorf("cannot select %d chars from a set of size %d ('%s') without replacement", count, len(runes), charSetStr)
		}  // If the character set is smaller than the required length and no repetition is allowed, report an error
		specs = append(specs, CharSpec{Chars: charSetStr, Count: count, Runes: runes}) // Continuously produce the array of character set spec structs
		totalCount += count  // Increment the character count for each processed character subset; the final count equals the user's intended password length
	}
	if totalCount > MAX_TOTAL_LENGTH {
		return nil, 0, fmt.Errorf("total password length (%d) exceeds MAX_TOTAL_LENGTH (%d)", totalCount, MAX_TOTAL_LENGTH)
	}
	return specs, totalCount, nil
}

// --- Generator for "no-repeat" mode in countsMode ---

// generateCombinationsNoRepeat recursively generates all combinations of non-repeating characters. It calls itself via the internal 'pick' function.
// out chan<- []rune is a results channel
// currentBag []rune is a slice holding the current combination-in-progress
func generateCombinationsNoRepeat(specIndex int, currentBag []rune, specs []CharSpec, out chan<- []rune) { // No matter how this function calls itself, all combination results are written and sent from the 'out' channel
	if specIndex == len(specs) {  // This means one password combination has been generated. Be aware that this doesn't mean all combinations for the spec have been generated; this is just a result from one branch.
		bagCopy := make([]rune, len(currentBag))
		copy(bagCopy, currentBag) // The currentBag slice points to an underlying array that is used by multiple recursive calls sequentially. A data race could occur, so a copy is used.
		out <- bagCopy // The necessity of using a data copy might be quite obscure. After a password combination result is generated, the downstream consumer channel might not read it in time, which could lead to missing a result, or the generation process ending early, causing the last result to be read repeatedly.
		return
	}
	spec := specs[specIndex]
	var pick func(startIdx int, chosen []rune)
	pick = func(startIdx int, chosen []rune) {
		if len(chosen) == spec.Count {
			generateCombinationsNoRepeat(specIndex+1, append(currentBag, chosen...), specs, out)
			return
		}
		if len(spec.Runes)-startIdx < spec.Count-len(chosen) {
			return  // If the remaining available characters are not enough to form a full-length combination, return.
		}
		for i := startIdx; i < len(spec.Runes); i++ {
			// Starting from startIdx, loop through the remaining characters in spec.Runes to create various branches.
			pick(i+1, append(chosen, spec.Runes[i]))
		}
	}// The real meaning of the startIdx parameter in the 'pick' function is the starting point for the new branch that the current partial password combination is about to create. This starting point is anchored to the spec.Runes character set.
	pick(0, []rune{}) // The startIdx parameter of the 'pick' function, especially this 0, can be misleading. It seems to represent the current length of the partial password combination being generated, but it actually represents which character in spec.Runes to start the loop from to create branches.
}

// generatePermutationsStream generates all permutations for a given slice of runes. The basic mechanism is based on factorial, i.e., X*X-1*X-2*X-3...*1.
func generatePermutationsStream(arr []rune, out chan<- string, semaphore chan struct{}, permWg *sync.WaitGroup) {
	defer permWg.Done()
	semaphore <- struct{}{}
	defer func() { <-semaphore }()
	var permute func(k int)
	permute = func(k int) {
		if k == len(arr) {
			out <- string(arr)
			return
		}
		for i := k; i < len(arr); i++ {
			arr[k], arr[i] = arr[i], arr[k]
			permute(k + 1) // Each increment of k is an abstract representation of the progression from X to X-1, X-2, etc.
			arr[k], arr[i] = arr[i], arr[k]
		}
	}
	permute(0)
}

// --- Generator for "with-repetition" mode (new optimized implementation) ---

// This function now only generates positional "templates", not the final passwords. It sends the templates to a channel. It's used for generating positional templates in countsMode with repetition.
// The specIdx parameter is used to insert the positional classification template.
func generatePositionalTemplates(template []int, specs []CharSpec, specIdx int, out chan<- []int) {
	if specIdx == len(specs) { // The outermost recursive call operates on the basis of a single character set spec.
		templateCopy := make([]int, len(template))
		copy(templateCopy, template)
		out <- templateCopy
		return
	}
	spec := specs[specIdx]
	var place func(startPos, countLeft int)
	place = func(startPos, countLeft int) {
		if countLeft == 0 {
			generatePositionalTemplates(template, specs, specIdx+1, out)
			return
		}
		if len(template)-startPos < countLeft {
			return
		}
		for i := startPos; i < len(template); i++ {
			if template[i] == -1 {
				template[i] = specIdx
				place(i+1, countLeft-1)
				template[i] = -1 // Backtrack to the initial template of the previous step to create a completely different branch from the current starting point.
			}
		}
	}
	place(0, spec.Count)// The innermost loop recursion within a single character set spec, creating branches from different starting points.
}

// fillFromTemplate receives a template defining character types and generates all possible passwords. This function is used for filling password templates in countsMode with repetition.
func fillFromTemplate(template []int, specs []CharSpec, pos int, currentPassword []rune, out chan<- string) {
	if pos == len(template) {
		out <- string(currentPassword)
		return
	}
	specIndex := template[pos]
	spec := specs[specIndex]
	for _, r := range spec.Runes {
		currentPassword[pos] = r
		fillFromTemplate(template, specs, pos+1, currentPassword, out)
	}
}


// calculateExpectedCountForRegex estimates the total number of passwords that a regular expression will generate.
func calculateExpectedCountForRegex(node *syntax.Regexp) float64 {
	switch node.Op {
	case syntax.OpLiteral:
		return 1
	case syntax.OpCharClass:  // This applies to calculating the count for [abc], [acde], [a-z]. The count is calculated by the difference between the Unicode code points of every two anchored characters. Something like [abc] will be pre-parsed into a form similar to "a-a,b-b,c-c", indirectly achieving the effect of summing up individual characters.
		var count float64
		for i := 0; i < len(node.Rune); i += 2 {
			count += float64(node.Rune[i+1] - node.Rune[i] + 1)
		}
		return count
	case syntax.OpConcat:
		total := 1.0
		// node.Sub contains all the concatenated sub-parts
	    // For example, for "[a-z][0-9]", node.Sub would contain the nodes for [a-z] and [0-9]
		// This for loop iterates over each sub-part
		for _, sub := range node.Sub {
			// This is a "recursive" call
			// It calls the function on the current sub-part (e.g., [a-z])
			// to calculate how many combinations this sub-part has on its own
			subCount := calculateExpectedCountForRegex(sub)
			if subCount < 0 {
				// This is a safety check. If any sub-part cannot be calculated (e.g., contains an infinite repetition like *)
				// then the entire concatenation cannot be calculated, so it immediately returns -1 to indicate "cannot compute"
				return -1.0
			}
			total *= subCount
		}
		return total
	case syntax.OpAlternate: // Used to process nodes like [a-c]|[0-9], using a loop to handle them separately.
		var total float64
		for _, sub := range node.Sub {
			subCount := calculateExpectedCountForRegex(sub)
			if subCount < 0 {
				return -1.0
			}
			total += subCount
		}
		return total
	case syntax.OpRepeat:
		if len(node.Sub) == 0 {
			return 1.0
		}
		subCount := calculateExpectedCountForRegex(node.Sub[0]) // First, calculate the total count of the repeated node in its standalone state.
		if subCount < 0 {
			return -1.0
		}
		if node.Max < 0 {  
			// node.Min and node.Max store the minimum and maximum repetition counts for this repeat node (OpRepeat).
			// For example, for the regex "a{2,4}", Min is 2 and Max is 4; for "a*", Min is 0 and Max is -1 (representing infinity).
			return -1.0 // Cannot calculate infinite repetition
		}
		if subCount == 0 {
			return 1.0
		}
		if subCount == 1.0 {
			return float64(node.Max - node.Min + 1)
		}

		var total float64
		for i := node.Min; i <= node.Max; i++ {
			total += math.Pow(subCount, float64(i)) // Based on different repetition counts, accumulate the different total password counts.
		}
		return total
	case syntax.OpBeginText, syntax.OpEndText, syntax.OpEmptyMatch: 
	// Handles tokens like ^ (start of text) and $ (end of text).
	// Since they only represent a position and not an actual character, they don't affect the total password count themselves (can be treated as 1).
		if len(node.Sub) > 0 {
			return calculateExpectedCountForRegex(node.Sub[0])
		}
		return 1.0
	default:
		// For unsupported operators, return -1 to indicate unknown.
		// Simplify() handles OpStar, OpPlus, and OpQuest, so we usually won't see them here.
		return -1.0
	}
}

// --- Regex Mode Functions ---
type GeneratorParams struct {
	Out       chan string
	WaitGroup *sync.WaitGroup
}  // The channel for generating results and its control

func generateCombinationsWithPipelining(node *syntax.Regexp, inPrefixes <-chan string, params GeneratorParams) <-chan string {
	outCombinations := make(chan string, 100)
	params.WaitGroup.Add(1)
	go func() {
		defer params.WaitGroup.Done()
		defer close(outCombinations)

		switch node.Op {
		case syntax.OpBeginText, syntax.OpEndText:
			if len(node.Sub) > 0 {
				subOut := generateCombinationsWithPipelining(node.Sub[0], inPrefixes, params)
				for res := range subOut { outCombinations <- res }
			} else {
				for p := range inPrefixes { outCombinations <- p }
			}
		case syntax.OpLiteral:
			for p := range inPrefixes { outCombinations <- p + string(node.Rune) }
		case syntax.OpCharClass:  // Used to generate password parts like [a-c1-3]
			for p := range inPrefixes {
				// node.Rune will store a slice of Unicode code point ranges like ['1', '3', 'a', 'c']
				for i := 0; i < len(node.Rune); i += 2 {
					for r := node.Rune[i]; r <= node.Rune[i+1]; r++ {
						outCombinations <- p + string(r)
					}
				}
			}
		case syntax.OpConcat: // Generates passwords like [a-b][1-2]
			currentPipe := inPrefixes // Since it's initialized with the original inPrefixes, its capacity at this point is 1.
			for _, subNode := range node.Sub {
				subOutPipe := generateCombinationsWithPipelining(subNode, currentPipe, params)
				currentPipe = subOutPipe // Because the results channel from the previous recursive step has a capacity of 100, currentPipe's capacity is now also 100.
			}
			for finalComb := range currentPipe { outCombinations <- finalComb }
		case syntax.OpAlternate:
			// Buffer all incoming prefixes into a slice, because a channel can only be consumed once.
			var prefixes []string
			for p := range inPrefixes {
				prefixes = append(prefixes, p)
			}

			var altOuts []<-chan string
			// Create a new processing pipeline for each alternate branch (e.g., [a-z]{3}, [A-Z]{2}).
			for _, sub := range node.Sub {
				// Create a new input channel for prefixes for this specific branch.
				branchInput := make(chan string, len(prefixes))

				// Start a dedicated goroutine to "feed" this branch with all the buffered prefixes.
				go func(ps []string, out chan<- string) {
					defer close(out)
					for _, p := range ps {
						out <- p
					}
				}(prefixes, branchInput)

				// Recursively call the generator for this branch.
				// Since this regex has multiple branches, it uses multiple result output channels. Therefore, an array of channels can be used to continuously expand.
				altOuts = append(altOuts, generateCombinationsWithPipelining(sub, branchInput, params))
			}

			// Merge the results from all parallel branches back into the main output channel.
			var mergeWg sync.WaitGroup
			mergeWg.Add(len(altOuts))
			for _, ch := range altOuts {  // Keep in mind that altOuts is an array of channels.
				go func(c <-chan string) {
					defer mergeWg.Done()
					for val := range c {
						outCombinations <- val
					}
				}(ch)
			}
			mergeWg.Wait()
		case syntax.OpRepeat:
			min, max := node.Min, node.Max
			if max < 0 { max = MAX_TOTAL_LENGTH }
			if min > max { return }   // Directly return for abnormal length parameters from user input.
			var repeatFunc func(currentInput <-chan string, n int)
			repeatFunc = func(currentInput <-chan string, n int) {
				if n >= max {  // If the generated password reaches the max parameter
					if n >= min {
						for val := range currentInput { outCombinations <- val }
					}
					return  // If the user's input for min and max is abnormal, return directly here.
				}
				passThrough, loopBack := make(chan string, 100), make(chan string, 100)
				go func() {
					// currentInput, as a channel, will be consumed. To reuse elements, it fans out to two other channels within each loop.
					for val := range currentInput {
						passThrough <- val
						loopBack <- val
					}
					close(passThrough)
					close(loopBack)
				}()
				if n >= min {
					for val := range passThrough { outCombinations <- val }
				}
				nextInput := generateCombinationsWithPipelining(node.Sub[0], loopBack, params)
				repeatFunc(nextInput, n+1)
			} // This repeatFunc function calculates the total password count for the OpRepeat part by recursively calling itself multiple times.
			repeatFunc(inPrefixes, 0)
		default:
			fmt.Fprintf(os.Stderr, "Warning: Unsupported regex operator: %v\n", node.Op)
			for range inPrefixes {
			}
		}
	}()
	return outCombinations
}

// --- Main Function ---

func main() {

	flag.IntVar(&MAX_TOTAL_LENGTH, "max-len", 15, "Set maximum password length")
	flag.IntVar(&MAX_CONCURRENT_PERMUTATIONS, "perm-concurrency", 4, `Set max concurrent permutation goroutines in "no-repeat" mode`)

	regexPattern := flag.String("regex", "", "Regex pattern (e.g., [a-z]{3}[!%]{2})")
	countsPattern := flag.String("counts", "", "Counts pattern (e.g., a-z:3,!:1,%:1)")
	outputFile := flag.String("out", "password_list.txt", "Output file name")
	allowCharRepeat := flag.Bool("allow-char-repeat", false, "Allow character repetition in counts mode (default: false)")
	flag.Parse()

	if (*regexPattern == "" && *countsPattern == "") || (*regexPattern != "" && *countsPattern != "") { // Must specify exactly one of --regex or --counts mode.
		fmt.Fprintln(os.Stderr, "Error: Must specify exactly one of --regex or --counts pattern.")
		flag.Usage()
		os.Exit(1)
	}

	var rootWg sync.WaitGroup
	params := GeneratorParams{
		Out:       make(chan string, 100000),
		WaitGroup: &rootWg,
	}  // This is effectively the final channel that directly connects to the password file writer. Elements read from it go straight to the bufio.Writer.

	var processedCount int64 = 0 // OPTIMIZATION: Changed to int64 for atomic operations
	var expectedTotalCount float64 = -1.0
	//lastTime = time.Now()  // This line is temporarily removed to observe subsequent results.

	// Start file writing goroutine (consumer)
	var writeWg sync.WaitGroup
	writeWg.Add(1)
	go func() {
		defer writeWg.Done()
		file, err := os.Create(*outputFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating file: %v\n", err)
			return
		}
		defer file.Close()

		// OPTIMIZATION: Use bufio.Writer to improve write performance
		writer := bufio.NewWriter(file)
		defer writer.Flush() // Ensure all buffered data is written to the file before the program exits.

		ticker := time.NewTicker(printInterval)
		defer ticker.Stop()

		for {
			select {
			case pwd, ok := <-params.Out:
				if !ok {
					// Manually load one last time before the final print to get the final count.
					finalCount := atomic.LoadInt64(&processedCount)
					printProgressBar(finalCount, expectedTotalCount, true)
					fmt.Println("\nFinished writing", finalCount, "passwords to", *outputFile)
					return
				}
				if strings.TrimSpace(pwd) != "" {
					fmt.Fprintln(writer, pwd)
					// OPTIMIZATION: Use atomic operations instead of a mutex for higher efficiency.
					atomic.AddInt64(&processedCount, 1)
				}
			case <-ticker.C:
				// OPTIMIZATION: Atomically read the count value.
				currentCount := atomic.LoadInt64(&processedCount)
				printProgressBar(currentCount, expectedTotalCount, false)
			}
		}
	}()   // This goroutine function is a separate, real-time task that simultaneously handles writing to the password file and printing the progress bar.

	var finalCombinations <-chan string  // The next stage for this read-only channel is params.Out, which ultimately writes to the file.

	if *regexPattern != "" {
		p := *regexPattern
		re, err := syntax.Parse(p, syntax.Perl)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Invalid regex:", err)
			os.Exit(1)
		}
		re = re.Simplify() // Just a preliminary optimization of the user-inputted regular expression.

		expectedTotalCount = calculateExpectedCountForRegex(re)
		fmt.Printf("Expected total passwords: %.0f\n", expectedTotalCount)

		initialPrefixes := make(chan string, 1)  // Capacity is only 1 to hold an initial prefix. The basic model of this program is to generate multiple branches from a single prefix.
		initialPrefixes <- ""
		close(initialPrefixes)

		finalCombinations = generateCombinationsWithPipelining(re, initialPrefixes, params)
	} else {
		specs, totalLen, err := parseCountsPattern(*countsPattern, *allowCharRepeat)
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error parsing counts pattern:", err)
			os.Exit(1)
		}
		fmt.Printf("Generating passwords with total length %d based on counts: %s\n", totalLen, *countsPattern)

		if *allowCharRepeat {
			// --- New concurrent optimization logic for [Allow Repetition] ---
			expectedTotalCount = calculateExpectedCountForCountsModeWithRepetition(specs, totalLen)
			fmt.Printf("Expected total passwords: %.0f\n", expectedTotalCount)

			// 1. Create a channel to pass template tasks.
			templateChan := make(chan []int, 1000)

			// 2. Start a goroutine to generate a password template initially filled with -1, then close templateChan when done.
			rootWg.Add(1)
			go func() {
				defer rootWg.Done()
				defer close(templateChan)
				template := make([]int, totalLen)
				for i := range template {
					template[i] = -1
				}
				generatePositionalTemplates(template, specs, 0, templateChan)
			}()

			// 3. OPTIMIZATION: Start a Worker Pool to process templates concurrently.
			//    The number of workers is typically set to the number of CPU cores for optimal performance.
			numWorkers := runtime.NumCPU()
			fmt.Printf("Starting %d workers for password generation...\n", numWorkers)
			rootWg.Add(numWorkers)
			for i := 0; i < numWorkers; i++ {
				go func() {
					defer rootWg.Done()
					// Each worker gets tasks from templateChan until the channel is closed.
					for t := range templateChan {
						fillFromTemplate(t, specs, 0, make([]rune, totalLen), params.Out)
					}
				}()
			}

		} else {
			// --- Logic for [CountsMode without repetition] (efficient pipeline model) ---
			expectedTotalCount = calculateExpectedCountForCountsModeNoRepeat(specs, totalLen)
			fmt.Printf("Expected total passwords: %.0f\n", expectedTotalCount)

			bagCombinations := make(chan []rune, 1000)
			permutationOut := make(chan string, 10000)

			// 1. Generate combinations (single-threaded).
			rootWg.Add(1)
			go func() {
				defer rootWg.Done()
				defer close(bagCombinations)
				generateCombinationsNoRepeat(0, []rune{}, specs, bagCombinations)
			}()

			// 2. Permute the combinations (concurrent).
			rootWg.Add(1)
			go func() {
				defer rootWg.Done()
				defer close(permutationOut)
				var permWg sync.WaitGroup
				semaphore := make(chan struct{}, MAX_CONCURRENT_PERMUTATIONS)
				for bag := range bagCombinations {
					permWg.Add(1)
					go generatePermutationsStream(bag, permutationOut, semaphore, &permWg)
				}
				permWg.Wait()
			}()
			finalCombinations = permutationOut
		}   // --- Logic for [CountsMode without repetition] (efficient pipeline model) ---
	}

	// If it's a mode that needs merging into the main output channel.
	if finalCombinations != nil {
		rootWg.Add(1)
		go func() {
			defer rootWg.Done()
			for val := range finalCombinations {
				params.Out <- val
			}
		}()
	}

	// Wait for all generator goroutines to complete.
	rootWg.Wait()
	// All passwords have been generated, close the output channel.
	close(params.Out)

	// Wait for the file writing goroutine to complete.
	writeWg.Wait()
	fmt.Println("Program finished.")
}

// printProgressBar prints a progress bar
func printProgressBar(current int64, total float64, isFinal bool) {
	now := time.Now()
	lastPrintTime.Do(func() { lastTime = now })  // lastPrintTime is actually a sync.Once struct. Its Do method is used here to ensure the contained function is executed only once, effectively serving to initialize a variable.

	if !isFinal && now.Sub(lastTime) < printInterval {
		return
	}
	lastTime = now

	const barLength = 40
	var progress float64
	var percentage, totalStr string

	if total > 0 {
		progress = float64(current) / total
		percentage = fmt.Sprintf("%.2f%%", progress*100)
		totalStr = fmt.Sprintf(" / %.0f", total)  // Note that the printed total password count is preceded by a "/" and a space.
	} else {
		percentage, totalStr = "N/A", " / Unknown"
	}
	filledLength := int(progress * float64(barLength))
	if filledLength > barLength { filledLength = barLength }

	bar := strings.Repeat("█", filledLength) + strings.Repeat(" ", barLength-filledLength)  // Print the progress bar blocks and padding spaces on the same line.

	fmt.Fprintf(os.Stderr, "\rProgress: [%s] %s (%d%s) ", bar, percentage, current, totalStr)
	if isFinal {
		fmt.Fprintln(os.Stderr, "")
	}
}