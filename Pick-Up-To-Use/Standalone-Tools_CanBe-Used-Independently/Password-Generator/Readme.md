This folder includes a basic password generator along with its source code. For users who prefer an out-of-the-box solution, you can run the brufjasgen executable directly. For those who are concerned about the security of the source code or wish to customize it, the source code is available for modification. However, it is not licensed for commercial use.

### Tool Philosophy & Limitations

*   **Primary Goal:** This tool is designed to generate large, random password files, intended as a powerful replacement for pre-made wordlists like `rockyou.txt` commonly used in CTF competitions.

*   **Important Distinction:** It is **not** a specialized generator for passwords tailored to social engineering attacks. The underlying mechanism is based on the permutation of individual random characters, not the combination of specific word phrases.

*   **Regarding Regex Mode:** While you can use the tool's `regex` mode to achieve a degree of pattern-specific generation, its rigid foundation cannot handle complex, multi-layered permutations of different wordlists.

*   **For Advanced Users (Social Engineering):** If you have advanced requirements for social engineering attacks, you are encouraged to examine the implementation of `countMode` in the source code. A potential path forward would be to adapt the core logic to handle entire word phrases as base units, instead of just individual characters.

> **Platform Compatibility Note**
>
> The `brufjasgen` executable is compiled for **x64 Linux** systems and has only been fully tested on an x64 Kali Linux machine. Users on **Windows** or **macOS** are advised to compile the application from the source code.

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
