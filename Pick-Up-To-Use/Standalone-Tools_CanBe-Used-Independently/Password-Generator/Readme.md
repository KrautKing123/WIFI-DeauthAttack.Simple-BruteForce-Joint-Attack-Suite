This folder includes a basic password generator along with its source code. For users who prefer an out-of-the-box solution, you can run the brufjasgen executable directly. For those who are concerned about the security of the source code or wish to customize it, the source code is available for modification. However, it is not licensed for commercial use.

> **A Note on Social Engineering vs. Brute-Force**
>
> While `brufjasgen` aims to replace generic password lists (e.g., `rockyou.txt` and `SecLists`) with massive, randomized password files for CTF scenarios, it should not be mistaken for a social engineering password generator. Its core logic is built on character-level permutations, not the complex combination of meaningful words or phrases.
>
> Although the `regex` mode offers some control, it lacks the flexibility for true social engineering patterns. Users with such needs should consider adapting the `countMode` logic in the source code to work with word-based units instead of character-based ones.

> [!WARNING]
> The `brufjasgen` executable is compiled for **x64 Linux** systems and has only been fully tested on an x64 Kali Linux machine. Users on **Windows** or **macOS** are advised to compile the application from the source code.

> [!WARNING]
> **Disk Space Consideration**
> A password file with hundreds of millions of entries can easily reach several gigabytes (GB) in size. Before running a large-scale task, please check your available disk space. It is also recommended to perform a small test run to estimate the potential output file size.

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
```

The `-perm-concurrency` parameter controls goroutine concurrency. The default value, `4`, is optimized for lower-spec machines. Feel free to modify this number to best fit your system's resources.



### Getting Started: Trying the Regex Mode

```bash
# ./brufjasgen -regex '[a-d]{5}|[e-h1-3]{6}' -out 'password.txt'
Expected total passwords: 118673
Progress: [████████████████████████████████████████] 100.00% (118673 / 118673)
Finished writing 118673 passwords to password.txt
Program finished.
```

### Example Output Preview

The following is a snippet from a generated password file. Please have a attention on the term `bdacc`.

```text
...
bcddd
bdaaa
bdaab
bdaac
bdaad
bdaba
bdabb
bdabc
bdabd
bdaca
bdacb
bdacc <<<
bdacd
bdada
bdadb
...

```

From this output, you can immediately tell that the `[a-d]{5}` part of the expression was matched. You will also notice that characters from the `a-d` set are repeated within the result `bdacc`. In fact, there is no specific option to control character repetition within the regex mode; it is always allowed by design. Actually, `regex` mode does not allow the `--allow-char-repeat` option to be used at all.
