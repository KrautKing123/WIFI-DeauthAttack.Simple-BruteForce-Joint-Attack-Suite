This folder includes a basic password generator along with its source code. For users who prefer an out-of-the-box solution, you can run the brufjasgen executable directly. For those who are concerned about the security of the source code or wish to customize it, the source code is available for modification. However, it is not licensed for commercial use.

> **A Note on Social Engineering vs. Brute-Force**
>

> Whereas generic password lists like `rockyou.txt` and `SecLists` are often sufficient for CTF scenarios, `brufjasgen` is engineered to generate massive, randomized password files suited for **real-world attack scenarios** where common wordlists fall short.For instance, in a real-world attack scenario such as brute-forcing a Wi-Fi password—a primary use case for this project—generic wordlists like `rockyou.txt` and `SecLists` are likely to be insufficient. Moreover, these types of general-purpose lists are often **too bulky and lack the specific focus** required for such a targeted attack. However, the `brufjasgen` should not be mistaken for a social engineering password generator, as its core logic is built on character-level permutations, not the complex combination of meaningful words or phrases.
> Although the `regex` mode offers some control, it lacks the flexibility for true social engineering patterns. Users with such needs should consider adapting the `countMode` logic in the source code to work with word-based units instead of character-based ones.

> [!WARNING]
> The `brufjasgen` executable is compiled for **x64 Linux** systems and has only been fully tested on an x64 Kali Linux machine. Users on **Windows** or **macOS** are advised to compile the application from the source code.

> [!WARNING]
> ## Disk Space Consideration
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



## Getting Started: Trying the Regex Mode

```bash
 ./brufjasgen -regex '[a-d]{5}|[e-h1-3]{6}' -out 'password.txt'
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

From this output, you can immediately tell that the `[a-d]{5}` part of the expression was matched. You will also notice that characters from the `a-d` set are repeated within the result `bdacc`. In fact, there is no specific option to control character repetition within the regex mode; it is always allowed by design. Actually, `regex` mode does not allow the `--allow-char-repeat` option to be used at all, 
you would find out when you look at the source code for this project....
## Another Output Sample
Now, please notice the entry `3g2eh1` in the output list. This password was generated because the `[e-h1-3]{6}` branch of the regex pattern was matched.
```text
...
3g2egg
3g2egh
3g2eh1  <<<
3g2eh2
3g2eh3
...
```
Thus, the pattern `[e-h1-3]{6}` provides a way to achieve a **mixed-character effect** to some extent, combining different character classes within the output.
## Regarding Specific Phrases

Consider the following command:
```bash
 ./brufjasgen -regex '[a-c1-3]{4}john|mike[a-c1-3]{4}' -out 'password.txt' 
Expected total passwords: 2592
Progress: [████████████████████████████████████████] 100.00% (2592 / 2592) 

Finished writing 2592 passwords to password.txt
Program finished.
```
You might expect to find results like the following in the output file:

```text
...
ccb2john1cca
...
ccb2mike1cca
...
```
While you might expect a different format, the **actual** results will appear as follows:

```text
...
ccc3john
cccajohn
cccbjohn
ccccjohn
mike1cac
mike1cb1
mike1cb2
mike1cb3
...
```
> [!NOTE]
Although you might hope to use the `|` or `||` symbols to switch between specific phrases, the tool will simply split the entire match into two parts. Attempting to work around this by enclosing the phrases in square brackets `[]` is also ineffective; this will only cause it to pick a single character from the string, and can even lead to the comical result of the `|` symbol itself being treated as a character for selection.

### Regular Expressions for Special Characters
Consider the following command:
```bash
 ./brufjasgen -regex '[!-^]' -out 'password.txt'  
Expected total passwords: 62
Progress: [████████████████████████████████████████] 100.00% (62 / 62) 

Finished writing 62 passwords to password.txt
Program finished.
```
For clarity, let me consolidate the results onto a single `paragraph`:
```text
! " # $ % & ' ( ) * + , - . / 0 1 2 3 4 5 6 7 8 9 : ; < = > ? @
 A B C D E F G H I J K L M N O P Q R S T U V W X Y Z [ \ ] ^
```
### ⚠️ Important: How Regex Handles Character Ranges

You've probably seen that the result contains more than just symbols; it's also mixed with numbers and letters. This is happening because the regex engine doesn't see character *types*. Instead, it processes ranges based on the **Unicode code point value** of each character.

Because the code points for `0-9` and `A-Z` fall between `!` and `^` in the Unicode table, they are included in the generated output. To get *only* the special symbols you need, you must look up their specific code points and create a precise pattern that skips over the letters and numbers.

**Here is a more precise method:**
```bash
 ./brufjasgen -regex '[!-/:-@[-^]' -out 'password.txt' 
Expected total passwords: 26
Progress: [████████████████████████████████████████] 100.00% (26 / 26) 

Finished writing 26 passwords to password.txt
Program finished.
```
Allow me to present the results as a `paragraph`:
```text
! " # $ % & ' ( ) * + , - . /
: ; < = > ? @
[ \ ] ^
```
> [!TIP]
> **A Quick Reminder**
>
> As I've mentioned, when generating special characters, you **must** consult a Unicode code point table and enter a more precise regular expression. This is the only way to ensure you get the exact symbols you want without unintended characters.



# Into the `-counts` Mode...
  When you check this project's source code, you'll see `-counts` mode as the complement to `-regex` mode. With a deep understanding of regular expressions, you know they are highly associated with **pattern and order**. This is by design; they operate just like a **genetic sequence**, where reversing the order results in something completely different. The sequence is absolute.
  
  The `-counts` mode is quite different. For instance, if I needed a mixed password with '3 numbers, 2 lowercase letters, and 1 `!` symbol', `-regex` falls short. `[0-9]{3}[a-z]{2}[!]{1}` is comical because its order is fixed, failing our need for a "mixed" result. `[0-9a-z!]{6}` is even more absurd, as it also fails the "specific quantity" requirement. The `-counts` mode, however, can accomplish this task effortlessly.

Consider the following command:
```bash
 ./brufjasgen -counts '0-9:3,a-z:2,!:1' --allow-char-repeat -out 'password.txt' 
Generating passwords with total length 6 based on counts: 0-9:3,a-z:2,!:1
Expected total passwords: 40560000
Starting 2 workers for password generation...
Progress: [████████████████████████████████████████] 100.00% (40560000 / 40560000) 

Finished writing 40560000 passwords to password.txt
Program finished.

 grep -C 4 '^3!bg44$' password.txt 
3!bg40
3!bg41
3!bg42
3!bg43
3!bg44  <<<
3!bg45
3!bg46
3!bg47
3!bg48
```
Another key feature of the `-counts` mode is its ability to work in a "no-repeat" mode, ensuring unique characters in the output.
```bash
 ./brufjasgen -counts '0-9:3,a-z:2,!:1' --allow-char-repeat=false -out 'password.txt'
Generating passwords with total length 6 based on counts: 0-9:3,a-z:2,!:1
Expected total passwords: 28080000
Progress: [████████████████████████████████████████] 100.00% (28080000 / 28080000) 

Finished writing 28080000 passwords to password.txt
Program finished.

 grep '^3!bg4' password.txt | grep -C 4 '^3!bg45$' 
3!bg40
3!bg41
3!bg42
3!bg45  <<<
3!bg48
3!bg46
3!bg47
3!bg49
```
You will not find passwords like `3!bg43` and `3!bg44` in the results.
## Some other usage of `-counts` mode...
```bash
./brufjasgen -counts '012678:3,abcABC:2,!-%:2' --allow-char-repeat=false -out 'password.txt'
Generating passwords with total length 7 based on counts: 012678:3,abcABC:2,!-%:2
Expected total passwords: 15120000
Progress: [████████████████████████████████████████] 100.00% (15120000 / 15120000) 

Finished writing 15120000 passwords to password.txt
Program finished.

grep -C 4 '^7%B6#0a$' password.txt 
7%B#6a0
7%B#60a
7%B60#a
7%B60a#
7%B6#0a  <<<
7%B6#a0
7%B6a#0
7%B6a0#
7%#aB06
```
What's worth to take a look is that the `!-%`, it is parsed into `[!, ", #, $, %]`, as what we had a brief explantation on `unicode` table. That's the reason that why we see the `#` in the output results.
> [!WARNING]
> ## The parse processing of `-counts` mode
> You should **never** modify the previous command to look like this:
```bash
./brufjasgen -counts '0-26-8:3,a-cA-C:2,!-%:2' --allow-char-repeat=false -out 'password.txt'
Generating passwords with total length 7 based on counts: 0-26-8:3,a-cA-C:2,!-%:2
Expected total passwords: 15120000
Progress: [████████████████████████████████████████] 100.00% (15120000 / 15120000) 

Finished writing 15120000 passwords to password.txt
Program finished.

grep -C 4 '^7%B6#0a$' password.txt

   ## literally nothing ##

grep -C 3 '^!-8A-a%$' password.txt
!-8a%A-
!-8Aa-%
!-8Aa%-
!-8A-a%
!-8A-%a
!-8A%-a  <<<
!-8A%a-
--
!-8a%A-
!-8Aa-%
!-8Aa%-
!-8A-a%
!-8A-%a
!-8A%-a  <<<
!-8A%a-
```
  The program does not parse multiple merged ranges as you would expect. Instead, it parses them as **literal characters**. Furthermore, it includes the hyphen (`-`) character, which should not have been part of the intended character set. So, while you might find the resulting characters `!`, `%`, `8`, `A`, and `a` to be understandable, the remaining `-` characters are clearly the result of an incorrect parsing of the input `'0-26-8:3'`.

  What's worse, two completely identical sets of output are produced. The program itself does not check if a parsed character array contains duplicate characters (please see the `parseCharSet` function); this is not a mission that a tool designed for high efficiency and saving machine resources should undertake. Regarding the two identical `'!-8A-a%'` sets, the program simply mechanically assumes that the two literally identical `-` characters have different underlying anchored indexes, and ultimately generates two results that the program considers to be different but are, in fact, completely identical.


   



