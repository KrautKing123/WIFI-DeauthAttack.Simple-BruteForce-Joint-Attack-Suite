This folder includes a basic password generator along with its source code. For users who prefer an out-of-the-box solution, you can run the brufjasgen executable directly. For those who are concerned about the security of the source code or wish to customize it, the source code is available for modification. However, it is not licensed for commercial use.

To view the usage instructions, execute the following command:
```bash
./brufjasgen -h
```
```bash
$ ./brufjasgen -h

Usage of ./brufjasgen:
  -allow-char-repeat
    	Allow character repetition in counts mode (default: false)
  -counts string
    	Counts pattern (e.g., a-z:3,!:1,%:1)
  -max-len int
    	Set maximum password length (default 15)
  -out string
    	Output file name (default "password_list.txt")
  -perm-concurrency int
    	Set max concurrent permutation goroutines in "no-repeat" mode (default 4)
  -regex string
    	Regex pattern (e.g., [a-z]{3}[!%]{2})
