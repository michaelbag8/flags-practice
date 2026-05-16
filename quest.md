## Concept 3 — File I/O: `os.WriteFile` and `os.ReadFile`

### What it is and why it exists

Every useful program eventually needs to persist data — save results, read configuration, log output, process documents. Go's `os` package gives you direct access to the filesystem. `os.ReadFile` and `os.WriteFile` are the simplest and most common tools for this.

Before these functions existed (added in Go 1.16), you had to open a file, read/write it, and close it manually — three steps every time. `ReadFile` and `WriteFile` wrap all three steps into one call.

---

### `os.ReadFile` — reading a file

```go
data, err := os.ReadFile("filename.txt")
```

Breaking down every part:

```go
data          // []byte — the raw bytes of the file content
err           // error — nil if success, non-nil if something went wrong
os            // the package
.ReadFile     // the function
("filename.txt") // the path to the file
```

**What it does internally:**
1. Opens the file
2. Reads all bytes into memory
3. Closes the file
4. Returns the bytes and any error

**Converting bytes to string:**

```go
data, err := os.ReadFile("hello.txt")
if err != nil {
    fmt.Fprintf(os.Stderr, "error: %v\n", err)
    os.Exit(1)
}

content := string(data)  // convert []byte to string
fmt.Println(content)
```

**Why bytes and not string directly:**

Files store raw bytes — not strings. A string in Go is a sequence of bytes with UTF-8 encoding assumed. The conversion `string(data)` interprets the bytes as UTF-8 text. If the file is binary (an image, a PDF), you work with the `[]byte` directly without converting.

---

### `os.WriteFile` — writing a file

```go
err := os.WriteFile("filename.txt", []byte(content), 0644)
```

Three arguments — each one matters:

**Argument 1 — filename:**
```go
"filename.txt"       // relative path — creates in current directory
"output/result.txt"  // relative path with folder — folder must exist
"/home/michael/file.txt"  // absolute path
```

**Argument 2 — data:**
```go
[]byte(content)  // convert string to bytes
// WriteFile needs []byte — not string
// same data, different type
```

**Argument 3 — file permissions (`0644`):**

Unix permissions use a three-digit octal number:

```
0644
^
octal prefix — tells Go this is base-8

6 = owner can read + write  (4+2)
4 = group can read only
4 = others can read only
```

Permission values:
```
4 = read
2 = write
1 = execute

read+write       = 4+2 = 6
read only        = 4
read+write+exec  = 4+2+1 = 7
```

Common permissions:
```
0644 — text files, config files (owner rw, others r)
0755 — executable files (owner rwx, others rx)
0600 — private files (owner rw only, no one else)
0777 — everyone can do everything (avoid this)
```

For any text file your program creates, always use `0644`.

---

### Error handling — every case you need to handle

```go
data, err := os.ReadFile("config.txt")
if err != nil {
    // What can go wrong:
    // 1. file does not exist
    // 2. permission denied
    // 3. path is a directory not a file
    // 4. disk error

    fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
    os.Exit(1)
}
```

The `%v` verb prints the error message from the OS — always include it:

```
error reading file: open config.txt: no such file or directory
error reading file: open config.txt: permission denied
error reading file: read config.txt: is a directory
```

Without `%v` you just print `"error reading file"` — useless for debugging.

---

### Checking if a file exists before reading

```go
// os.Stat returns file info or an error
_, err := os.Stat("config.txt")
if os.IsNotExist(err) {
    fmt.Println("file does not exist")
} else if err != nil {
    fmt.Println("other error:", err)
} else {
    fmt.Println("file exists")
}
```

`os.Stat` is like asking the filesystem "tell me about this file" without reading it. `os.IsNotExist(err)` specifically checks if the error is "file not found" vs other errors.

---

### Writing to a file — what happens to existing content

`os.WriteFile` always **overwrites** the entire file:

```go
// file.txt contains "Hello"
os.WriteFile("file.txt", []byte("World"), 0644)
// file.txt now contains "World" — "Hello" is gone
```

If you want to **append** to an existing file instead:

```go
// open with append flag
file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
if err != nil {
    fmt.Fprintf(os.Stderr, "error: %v\n", err)
    os.Exit(1)
}
defer file.Close()

file.WriteString("new line\n")
```

Three flags combined with `|` (bitwise OR):
- `os.O_APPEND` — write at end of file
- `os.O_CREATE` — create if does not exist
- `os.O_WRONLY` — open for writing only

---

### Reading line by line with `bufio.Scanner`

`os.ReadFile` reads the whole file at once. For large files you read line by line:

```go
file, err := os.Open("data.txt")
if err != nil {
    fmt.Fprintf(os.Stderr, "error: %v\n", err)
    os.Exit(1)
}
defer file.Close()

scanner := bufio.NewScanner(file)
for scanner.Scan() {
    line := scanner.Text()
    fmt.Println(line)
}

if err := scanner.Err(); err != nil {
    fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
}
```

`defer file.Close()` — guarantees the file closes when the function exits, even if something panics. Always defer Close immediately after a successful Open.

---

### Real world examples outside ASCII art

**Example 1 — config file reader:**

```go
func readConfig(filename string) (map[string]string, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("cannot read config: %w", err)
    }

    config := make(map[string]string)
    lines := strings.Split(string(data), "\n")

    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") {
            continue  // skip empty lines and comments
        }

        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }

        key   := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])
        config[key] = value
    }

    return config, nil
}
```

Config file `app.conf`:
```
# database settings
host = localhost
port = 5432
name = myapp

# server settings
server_port = 8080
```

Usage:
```go
config, err := readConfig("app.conf")
fmt.Println(config["host"])        // localhost
fmt.Println(config["server_port"]) // 8080
```

---

**Example 2 — log writer:**

```go
func writeLog(filename string, level string, message string) error {
    // read existing content
    var existing string
    data, err := os.ReadFile(filename)
    if err == nil {
        existing = string(data)
    }
    // if file does not exist, existing stays "" — that is fine

    // build new log entry
    entry := fmt.Sprintf("[%s] %s: %s\n",
        time.Now().Format("2006-01-02 15:04:05"),
        level,
        message,
    )

    // append and write back
    return os.WriteFile(filename, []byte(existing+entry), 0644)
}

// Usage
writeLog("app.log", "INFO",  "server started")
writeLog("app.log", "ERROR", "database connection failed")
writeLog("app.log", "INFO",  "retrying connection")
```

Log file after three calls:
```
[2026-05-15 10:30:01] INFO: server started
[2026-05-15 10:30:02] ERROR: database connection failed
[2026-05-15 10:30:03] INFO: retrying connection
```

---

**Example 3 — word counter:**

```go
func countWords(filename string) (int, int, int, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return 0, 0, 0, fmt.Errorf("error reading %s: %w", filename, err)
    }

    content := string(data)
    chars   := len([]rune(content))
    words   := len(strings.Fields(content))
    lines   := len(strings.Split(content, "\n"))

    return chars, words, lines, nil
}

// Usage
chars, words, lines, err := countWords("essay.txt")
fmt.Printf("characters: %d\n", chars)
fmt.Printf("words:      %d\n", words)
fmt.Printf("lines:      %d\n", lines)
```

---

**Example 4 — file copy:**

```go
func copyFile(src string, dst string) error {
    // read source
    data, err := os.ReadFile(src)
    if err != nil {
        return fmt.Errorf("cannot read source: %w", err)
    }

    // get source permissions
    info, err := os.Stat(src)
    if err != nil {
        return fmt.Errorf("cannot stat source: %w", err)
    }

    // write to destination with same permissions
    err = os.WriteFile(dst, data, info.Mode())
    if err != nil {
        return fmt.Errorf("cannot write destination: %w", err)
    }

    return nil
}

// Usage
err := copyFile("original.txt", "backup.txt")
if err != nil {
    fmt.Fprintf(os.Stderr, "copy failed: %v\n", err)
}
```

---

**Example 5 — JSON config save and load:**

```go
type Config struct {
    Host string
    Port int
    Name string
}

func saveConfig(filename string, cfg Config) error {
    // manual JSON building — real code would use encoding/json
    content := fmt.Sprintf(`{"host":"%s","port":%d,"name":"%s"}`,
        cfg.Host, cfg.Port, cfg.Name)
    return os.WriteFile(filename, []byte(content), 0644)
}

func loadConfig(filename string) (Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return Config{}, err
    }
    // parsing would use encoding/json in real code
    fmt.Println(string(data))
    return Config{}, nil
}
```

---

### Connection to your ASCII art project

Your project uses both functions:

```go
// --output flag — write rendered art to file
err := os.WriteFile(*output, []byte(content), 0644)

// LoadBanner — read the banner file
data, err := os.ReadFile(filename)

// --reverse flag — read rendered art back
data, err := os.ReadFile(filename)
```

The pattern is always the same:
```go
data, err := os.ReadFile(path)
if err != nil { /* handle */ }
content := string(data)
// use content
```

```go
err := os.WriteFile(path, []byte(content), 0644)
if err != nil { /* handle */ }
```

---

### 🏋️ Mini Tasks

**Task 1 — Basic:**
Write a program that:
1. Creates a file called `hello.txt` containing `"Hello, Michael!\nWelcome to Go file I/O\n"`
2. Reads it back
3. Prints the content and the number of lines

```bash
go run main.go
# Hello, Michael!
# Welcome to Go file I/O
# Lines: 2
```

**Task 2 — Medium:**
Write a program that acts as a simple note taker:
- `--add="note text"` — appends a new note to `notes.txt` with a timestamp
- `--list` — reads and prints all notes
- `--clear` — deletes all notes by writing empty content

```bash
go run main.go --add="Buy groceries"
go run main.go --add="Learn Go file I/O"
go run main.go --list
# [2026-05-15 10:30] Buy groceries
# [2026-05-15 10:31] Learn Go file I/O
go run main.go --clear
go run main.go --list
# (empty)
```

**Task 3 — Hard:**
Write a program that:
1. Reads a CSV file called `students.csv`
2. Calculates the average score for each student
3. Writes a new file `results.txt` with each student and their average

`students.csv`:
```
Name,Math,English,Science
Michael,85,90,78
Alice,92,88,95
Bob,70,75,80
```

`results.txt` output:
```
Michael: 84.33
Alice: 91.67
Bob: 75.00
```

Start with Task 1 and paste your solution. We do not move to Task 2 until Task 1 is correct.## Concept 3 — File I/O: `os.WriteFile` and `os.ReadFile`

### What it is and why it exists

Every useful program eventually needs to persist data — save results, read configuration, log output, process documents. Go's `os` package gives you direct access to the filesystem. `os.ReadFile` and `os.WriteFile` are the simplest and most common tools for this.

Before these functions existed (added in Go 1.16), you had to open a file, read/write it, and close it manually — three steps every time. `ReadFile` and `WriteFile` wrap all three steps into one call.

---

### `os.ReadFile` — reading a file

```go
data, err := os.ReadFile("filename.txt")
```

Breaking down every part:

```go
data          // []byte — the raw bytes of the file content
err           // error — nil if success, non-nil if something went wrong
os            // the package
.ReadFile     // the function
("filename.txt") // the path to the file
```

**What it does internally:**
1. Opens the file
2. Reads all bytes into memory
3. Closes the file
4. Returns the bytes and any error

**Converting bytes to string:**

```go
data, err := os.ReadFile("hello.txt")
if err != nil {
    fmt.Fprintf(os.Stderr, "error: %v\n", err)
    os.Exit(1)
}

content := string(data)  // convert []byte to string
fmt.Println(content)
```

**Why bytes and not string directly:**

Files store raw bytes — not strings. A string in Go is a sequence of bytes with UTF-8 encoding assumed. The conversion `string(data)` interprets the bytes as UTF-8 text. If the file is binary (an image, a PDF), you work with the `[]byte` directly without converting.

---

### `os.WriteFile` — writing a file

```go
err := os.WriteFile("filename.txt", []byte(content), 0644)
```

Three arguments — each one matters:

**Argument 1 — filename:**
```go
"filename.txt"       // relative path — creates in current directory
"output/result.txt"  // relative path with folder — folder must exist
"/home/michael/file.txt"  // absolute path
```

**Argument 2 — data:**
```go
[]byte(content)  // convert string to bytes
// WriteFile needs []byte — not string
// same data, different type
```

**Argument 3 — file permissions (`0644`):**

Unix permissions use a three-digit octal number:

```
0644
^
octal prefix — tells Go this is base-8

6 = owner can read + write  (4+2)
4 = group can read only
4 = others can read only
```

Permission values:
```
4 = read
2 = write
1 = execute

read+write       = 4+2 = 6
read only        = 4
read+write+exec  = 4+2+1 = 7
```

Common permissions:
```
0644 — text files, config files (owner rw, others r)
0755 — executable files (owner rwx, others rx)
0600 — private files (owner rw only, no one else)
0777 — everyone can do everything (avoid this)
```

For any text file your program creates, always use `0644`.

---

### Error handling — every case you need to handle

```go
data, err := os.ReadFile("config.txt")
if err != nil {
    // What can go wrong:
    // 1. file does not exist
    // 2. permission denied
    // 3. path is a directory not a file
    // 4. disk error

    fmt.Fprintf(os.Stderr, "error reading file: %v\n", err)
    os.Exit(1)
}
```

The `%v` verb prints the error message from the OS — always include it:

```
error reading file: open config.txt: no such file or directory
error reading file: open config.txt: permission denied
error reading file: read config.txt: is a directory
```

Without `%v` you just print `"error reading file"` — useless for debugging.

---

### Checking if a file exists before reading

```go
// os.Stat returns file info or an error
_, err := os.Stat("config.txt")
if os.IsNotExist(err) {
    fmt.Println("file does not exist")
} else if err != nil {
    fmt.Println("other error:", err)
} else {
    fmt.Println("file exists")
}
```

`os.Stat` is like asking the filesystem "tell me about this file" without reading it. `os.IsNotExist(err)` specifically checks if the error is "file not found" vs other errors.

---

### Writing to a file — what happens to existing content

`os.WriteFile` always **overwrites** the entire file:

```go
// file.txt contains "Hello"
os.WriteFile("file.txt", []byte("World"), 0644)
// file.txt now contains "World" — "Hello" is gone
```

If you want to **append** to an existing file instead:

```go
// open with append flag
file, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
if err != nil {
    fmt.Fprintf(os.Stderr, "error: %v\n", err)
    os.Exit(1)
}
defer file.Close()

file.WriteString("new line\n")
```

Three flags combined with `|` (bitwise OR):
- `os.O_APPEND` — write at end of file
- `os.O_CREATE` — create if does not exist
- `os.O_WRONLY` — open for writing only

---

### Reading line by line with `bufio.Scanner`

`os.ReadFile` reads the whole file at once. For large files you read line by line:

```go
file, err := os.Open("data.txt")
if err != nil {
    fmt.Fprintf(os.Stderr, "error: %v\n", err)
    os.Exit(1)
}
defer file.Close()

scanner := bufio.NewScanner(file)
for scanner.Scan() {
    line := scanner.Text()
    fmt.Println(line)
}

if err := scanner.Err(); err != nil {
    fmt.Fprintf(os.Stderr, "scan error: %v\n", err)
}
```

`defer file.Close()` — guarantees the file closes when the function exits, even if something panics. Always defer Close immediately after a successful Open.

---

### Real world examples outside ASCII art

**Example 1 — config file reader:**

```go
func readConfig(filename string) (map[string]string, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return nil, fmt.Errorf("cannot read config: %w", err)
    }

    config := make(map[string]string)
    lines := strings.Split(string(data), "\n")

    for _, line := range lines {
        line = strings.TrimSpace(line)
        if line == "" || strings.HasPrefix(line, "#") {
            continue  // skip empty lines and comments
        }

        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue
        }

        key   := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])
        config[key] = value
    }

    return config, nil
}
```

Config file `app.conf`:
```
# database settings
host = localhost
port = 5432
name = myapp

# server settings
server_port = 8080
```

Usage:
```go
config, err := readConfig("app.conf")
fmt.Println(config["host"])        // localhost
fmt.Println(config["server_port"]) // 8080
```

---

**Example 2 — log writer:**

```go
func writeLog(filename string, level string, message string) error {
    // read existing content
    var existing string
    data, err := os.ReadFile(filename)
    if err == nil {
        existing = string(data)
    }
    // if file does not exist, existing stays "" — that is fine

    // build new log entry
    entry := fmt.Sprintf("[%s] %s: %s\n",
        time.Now().Format("2006-01-02 15:04:05"),
        level,
        message,
    )

    // append and write back
    return os.WriteFile(filename, []byte(existing+entry), 0644)
}

// Usage
writeLog("app.log", "INFO",  "server started")
writeLog("app.log", "ERROR", "database connection failed")
writeLog("app.log", "INFO",  "retrying connection")
```

Log file after three calls:
```
[2026-05-15 10:30:01] INFO: server started
[2026-05-15 10:30:02] ERROR: database connection failed
[2026-05-15 10:30:03] INFO: retrying connection
```

---

**Example 3 — word counter:**

```go
func countWords(filename string) (int, int, int, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return 0, 0, 0, fmt.Errorf("error reading %s: %w", filename, err)
    }

    content := string(data)
    chars   := len([]rune(content))
    words   := len(strings.Fields(content))
    lines   := len(strings.Split(content, "\n"))

    return chars, words, lines, nil
}

// Usage
chars, words, lines, err := countWords("essay.txt")
fmt.Printf("characters: %d\n", chars)
fmt.Printf("words:      %d\n", words)
fmt.Printf("lines:      %d\n", lines)
```

---

**Example 4 — file copy:**

```go
func copyFile(src string, dst string) error {
    // read source
    data, err := os.ReadFile(src)
    if err != nil {
        return fmt.Errorf("cannot read source: %w", err)
    }

    // get source permissions
    info, err := os.Stat(src)
    if err != nil {
        return fmt.Errorf("cannot stat source: %w", err)
    }

    // write to destination with same permissions
    err = os.WriteFile(dst, data, info.Mode())
    if err != nil {
        return fmt.Errorf("cannot write destination: %w", err)
    }

    return nil
}

// Usage
err := copyFile("original.txt", "backup.txt")
if err != nil {
    fmt.Fprintf(os.Stderr, "copy failed: %v\n", err)
}
```

---

**Example 5 — JSON config save and load:**

```go
type Config struct {
    Host string
    Port int
    Name string
}

func saveConfig(filename string, cfg Config) error {
    // manual JSON building — real code would use encoding/json
    content := fmt.Sprintf(`{"host":"%s","port":%d,"name":"%s"}`,
        cfg.Host, cfg.Port, cfg.Name)
    return os.WriteFile(filename, []byte(content), 0644)
}

func loadConfig(filename string) (Config, error) {
    data, err := os.ReadFile(filename)
    if err != nil {
        return Config{}, err
    }
    // parsing would use encoding/json in real code
    fmt.Println(string(data))
    return Config{}, nil
}
```

---

### Connection to your ASCII art project

Your project uses both functions:

```go
// --output flag — write rendered art to file
err := os.WriteFile(*output, []byte(content), 0644)

// LoadBanner — read the banner file
data, err := os.ReadFile(filename)

// --reverse flag — read rendered art back
data, err := os.ReadFile(filename)
```

The pattern is always the same:
```go
data, err := os.ReadFile(path)
if err != nil { /* handle */ }
content := string(data)
// use content
```

```go
err := os.WriteFile(path, []byte(content), 0644)
if err != nil { /* handle */ }
```

---

### 🏋️ Mini Tasks

**Task 1 — Basic:**
Write a program that:
1. Creates a file called `hello.txt` containing `"Hello, Michael!\nWelcome to Go file I/O\n"`
2. Reads it back
3. Prints the content and the number of lines

```bash
go run main.go
# Hello, Michael!
# Welcome to Go file I/O
# Lines: 2
```

**Task 2 — Medium:**
Write a program that acts as a simple note taker:
- `--add="note text"` — appends a new note to `notes.txt` with a timestamp
- `--list` — reads and prints all notes
- `--clear` — deletes all notes by writing empty content

```bash
go run main.go --add="Buy groceries"
go run main.go --add="Learn Go file I/O"
go run main.go --list
# [2026-05-15 10:30] Buy groceries
# [2026-05-15 10:31] Learn Go file I/O
go run main.go --clear
go run main.go --list
# (empty)
```

**Task 3 — Hard:**
Write a program that:
1. Reads a CSV file called `students.csv`
2. Calculates the average score for each student
3. Writes a new file `results.txt` with each student and their average

`students.csv`:
```
Name,Math,English,Science
Michael,85,90,78
Alice,92,88,95
Bob,70,75,80
```

`results.txt` output:
```
Michael: 84.33
Alice: 91.67
Bob: 75.00
```

Start with Task 1 and paste your solution. We do not move to Task 2 until Task 1 is correct.


## Concept 4 — ANSI Escape Codes

### What they are and why they exist

Your terminal is not just a text display. It is a device that understands a hidden instruction language embedded inside the text stream. These instructions are called **ANSI escape codes** — named after the American National Standards Institute that standardised them in 1979.

When your terminal receives text, it scans for a special sequence starting with the ESC character (`\033`). When it finds one, it does not print those characters — it executes the instruction instead. This is how terminals support colors, cursor movement, bold text, and clearing the screen — all through invisible codes mixed into the text stream.

Every terminal emulator you use today — iTerm2, GNOME Terminal, VS Code's integrated terminal, Windows Terminal — understands ANSI codes. They are universal.

---

### The ESC character — the trigger

Everything starts with ASCII character 27 — the ESC character. In Go you write it three ways:

```go
"\033"   // octal notation  — most common in Go
"\x1b"   // hex notation    — same value, different notation
"\u001b" // unicode notation — same value again
```

All three are identical. `\033` is the most common in Go code. Pick one and be consistent.

---

### The anatomy of an escape sequence

```
\033  [  3  1  m
 |    |  |  |  |
 |    |  |  |  └── "m" — final byte, means "apply this color/style"
 |    |  |  └───── "1" — second digit of the code
 |    |  └──────── "3" — first digit of the code  (31 = red foreground)
 |    └─────────── "[" — opens the Control Sequence Introducer (CSI)
 └──────────────── ESC character — signals "instruction follows"
```

The full sequence `\033[31m` means: "switch foreground color to red."

---

### Foreground colors — text color

```go
const (
    Black   = "\033[30m"
    Red     = "\033[31m"
    Green   = "\033[32m"
    Yellow  = "\033[33m"
    Blue    = "\033[34m"
    Magenta = "\033[35m"
    Cyan    = "\033[36m"
    White   = "\033[37m"
    Reset   = "\033[0m"   // turns off ALL styling
)
```

Usage:
```go
fmt.Println(Red + "This is red" + Reset)
fmt.Println(Green + "This is green" + Reset)
fmt.Println(Blue + "This is blue" + Reset)
fmt.Println("This is normal")
```

---

### Background colors — highlight color

```go
const (
    BgBlack   = "\033[40m"
    BgRed     = "\033[41m"
    BgGreen   = "\033[42m"
    BgYellow  = "\033[43m"
    BgBlue    = "\033[44m"
    BgMagenta = "\033[45m"
    BgCyan    = "\033[46m"
    BgWhite   = "\033[47m"
)
```

Usage:
```go
fmt.Println(BgRed + "highlighted in red" + Reset)
fmt.Println(BgBlue + White + "white text on blue background" + Reset)
```

---

### Text styles

```go
const (
    Bold      = "\033[1m"
    Dim       = "\033[2m"
    Italic    = "\033[3m"
    Underline = "\033[4m"
    Blink     = "\033[5m"
    Reverse   = "\033[7m"  // swaps foreground and background
    Strike    = "\033[9m"  // strikethrough
)
```

Usage:
```go
fmt.Println(Bold + "This is bold" + Reset)
fmt.Println(Underline + "This is underlined" + Reset)
fmt.Println(Bold + Red + "Bold and red" + Reset)
```

---

### Combining styles

You combine styles by concatenating codes:

```go
// bold + red
fmt.Println("\033[1m\033[31m" + "Bold Red" + "\033[0m")

// or using constants
fmt.Println(Bold + Red + "Bold Red" + Reset)

// background + foreground
fmt.Println(BgBlue + White + "White on Blue" + Reset)

// bold + underline + green
fmt.Println(Bold + Underline + Green + "Important!" + Reset)
```

The order does not matter — all codes apply simultaneously until reset.

---

### The reset — why it is not optional

ANSI colors are **sticky**. Once applied, they stay active until explicitly reset:

```go
// WRONG — color bleeds into everything after
fmt.Println("\033[31m" + "Hello")
fmt.Println("This line is also red!")  // still red
fmt.Println("So is this!")             // still red

// CORRECT — always reset after colored text
fmt.Println("\033[31m" + "Hello" + "\033[0m")
fmt.Println("This is normal")   // back to normal
fmt.Println("So is this")       // also normal
```

The reset code `\033[0m` turns off everything — color, bold, underline, all of it.

---

### 256 colors — extended color mode

The basic 8 colors are limited. ANSI also supports 256 colors:

```go
// foreground: \033[38;5;{n}m  where n is 0-255
// background: \033[48;5;{n}m  where n is 0-255

fmt.Println("\033[38;5;208m" + "Orange text" + "\033[0m")
fmt.Println("\033[38;5;93m"  + "Purple text" + "\033[0m")
fmt.Println("\033[48;5;226m" + "Yellow background" + "\033[0m")
```

Color palette:
```
0-7    — standard colors (same as \033[30-37m)
8-15   — bright versions of standard colors
16-231 — 216 colors in a 6x6x6 cube
232-255 — grayscale from dark to light
```

---

### True color — RGB

Modern terminals support full 24-bit RGB color:

```go
// foreground: \033[38;2;{r};{g};{b}m
// background: \033[48;2;{r};{g};{b}m

red   := "\033[38;2;255;0;0m"
green := "\033[38;2;0;255;0m"
blue  := "\033[38;2;0;0;255m"
orange := "\033[38;2;255;165;0m"

fmt.Println(red + "True red" + "\033[0m")
fmt.Println(orange + "True orange" + "\033[0m")
```

A helper function makes this clean:

```go
func fg(r, g, b int) string {
    return fmt.Sprintf("\033[38;2;%d;%d;%dm", r, g, b)
}

func bg(r, g, b int) string {
    return fmt.Sprintf("\033[48;2;%d;%d;%dm", r, g, b)
}

fmt.Println(fg(255, 165, 0) + "Orange text" + "\033[0m")
fmt.Println(bg(0, 0, 128) + "Navy background" + "\033[0m")
```

---

### Cursor movement

ANSI codes also control the cursor position:

```go
const (
    CursorUp    = "\033[1A"  // move cursor up 1 line
    CursorDown  = "\033[1B"  // move cursor down 1 line
    CursorRight = "\033[1C"  // move cursor right 1 column
    CursorLeft  = "\033[1D"  // move cursor left 1 column
    ClearLine   = "\033[2K"  // clear entire current line
    ClearScreen = "\033[2J"  // clear entire screen
    Home        = "\033[H"   // move cursor to top-left
)

// Move cursor to specific position: \033[{row};{col}H
fmt.Printf("\033[5;10H")  // row 5, column 10
```

---

### A progress bar using cursor movement

```go
func progressBar(percent int, width int) string {
    filled := width * percent / 100
    empty  := width - filled

    bar := "\033[32m" +                    // green
        strings.Repeat("█", filled) +
        "\033[0m" +
        strings.Repeat("░", empty)

    return fmt.Sprintf("\r[%s] %d%%", bar, percent)
}

// Animate it
for i := 0; i <= 100; i += 5 {
    fmt.Print(progressBar(i, 20))
    time.Sleep(100 * time.Millisecond)
}
fmt.Println()
```

`\r` is carriage return — moves cursor to start of line without newline. Combined with `ClearLine` you can update a line in place.

---

### Detecting if terminal supports color

Not all outputs support ANSI — if your program's output is piped to a file, ANSI codes appear as garbage:

```bash
go run . --color=red "Hello" > output.txt
cat output.txt
# ^[[31m _    _  ^[[0m...  ← ugly
```

Check if stdout is a terminal before applying color:

```go
import "golang.org/x/term"

func supportsColor() bool {
    return term.IsTerminal(int(os.Stdout.Fd()))
}

colorCode := ""
if supportsColor() {
    colorCode = "\033[31m"
}
```

`term.IsTerminal` returns `true` if stdout is connected to a real terminal, `false` if it is a pipe or file.

---

### Real world examples outside ASCII art

**Example 1 — a colored logger:**

```go
package main

import "fmt"

const (
    Reset  = "\033[0m"
    Red    = "\033[31m"
    Green  = "\033[32m"
    Yellow = "\033[33m"
    Blue   = "\033[34m"
    Bold   = "\033[1m"
)

func logInfo(msg string) {
    fmt.Println(Blue + "[INFO] " + Reset + msg)
}

func logSuccess(msg string) {
    fmt.Println(Green + "[OK]   " + Reset + msg)
}

func logWarning(msg string) {
    fmt.Println(Yellow + "[WARN] " + Reset + msg)
}

func logError(msg string) {
    fmt.Println(Red + Bold + "[ERR]  " + Reset + msg)
}

func main() {
    logInfo("server starting on port 8080")
    logSuccess("database connected")
    logWarning("cache miss rate above 20%")
    logError("failed to reach payment service")
}
```

Output:
```
[INFO] server starting on port 8080     ← blue label
[OK]   database connected               ← green label
[WARN] cache miss rate above 20%        ← yellow label
[ERR]  failed to reach payment service  ← bold red label
```

---

**Example 2 — a colored diff viewer:**

```go
func printDiff(old string, new string) {
    oldLines := strings.Split(old, "\n")
    newLines := strings.Split(new, "\n")

    for _, line := range oldLines {
        fmt.Println("\033[31m- " + line + "\033[0m")
    }
    for _, line := range newLines {
        fmt.Println("\033[32m+ " + line + "\033[0m")
    }
}

printDiff(
    "Hello World\nThis is old",
    "Hello Go\nThis is new",
)
```

Output:
```
- Hello World      ← red
- This is old      ← red
+ Hello Go         ← green
+ This is new      ← green
```

---

**Example 3 — a status dashboard:**

```go
func printStatus(services map[string]bool) {
    fmt.Println("\033[1m" + "Service Status" + "\033[0m")
    fmt.Println(strings.Repeat("─", 30))

    for name, up := range services {
        status := "\033[32m● UP  \033[0m"
        if !up {
            status = "\033[31m● DOWN\033[0m"
        }
        fmt.Printf("%s  %s\n", status, name)
    }
}

services := map[string]bool{
    "API Server":    true,
    "Database":      true,
    "Cache":         false,
    "Email Service": true,
    "Payment":       false,
}
printStatus(services)
```

Output:
```
Service Status
──────────────────────────────
● UP    API Server     ← green dot
● UP    Database       ← green dot
● DOWN  Cache          ← red dot
● UP    Email Service  ← green dot
● DOWN  Payment        ← red dot
```

---

**Example 4 — syntax highlighter:**

```go
func highlight(code string) string {
    keywords := []string{"func", "var", "if", "for", "return", "package", "import"}

    result := code
    for _, kw := range keywords {
        colored := "\033[34m" + kw + "\033[0m"  // blue keywords
        result = strings.ReplaceAll(result, kw, colored)
    }
    return result
}

code := `package main
func main() {
    var x = 10
    if x > 5 {
        return
    }
}`
fmt.Println(highlight(code))
```

---

### Connection to your ASCII art project

```go
// your colors.go
var colorMap = map[string]string{
    "red":     "\033[31m",
    "green":   "\033[32m",
    "yellow":  "\033[33m",
    "blue":    "\033[34m",
    "magenta": "\033[35m",
    "cyan":    "\033[36m",
    "white":   "\033[37m",
}

const resetCode = "\033[0m"

// in renderLines — the color sandwich
if shouldColor {
    rawRow.WriteString(colorCode)       // turn color ON
    rawRow.WriteString(blockMaps[ch][row])
    rawRow.WriteString("\033[0m")       // turn color OFF
}
```

---

### 🏋️ Mini Tasks

**Task 1 — Basic:**
Write a program that prints a menu with colors:

```
===== MAIN MENU =====
1. New Game        ← green
2. Load Game       ← blue
3. Settings        ← yellow
4. Quit            ← red
=====================
```

The title and borders are bold white. Each option has a colored number matching the color of the text.

**Task 2 — Medium:**
Write a function `colorize(text string, fg string, bg string, bold bool) string` that wraps text in the correct ANSI codes. Then use it to print a table where:
- Header row has white text on blue background, bold
- Even rows have normal text
- Odd rows have text on a dark gray background (`\033[48;5;236m`)

```
| Name    | Score | Grade |   ← white on blue, bold
| Michael | 95    | A     |   ← normal
| Alice   | 82    | B     |   ← dark gray background
| Bob     | 91    | A     |   ← normal
| Carol   | 78    | C     |   ← dark gray background
```

**Task 3 — Hard:**
Write an animated loading spinner using cursor movement:

```go
func spinner(message string, duration time.Duration)
```

It should:
- Show a spinning animation: `⠋ ⠙ ⠹ ⠸ ⠼ ⠴ ⠦ ⠧ ⠇ ⠏` cycling through frames
- Print the message next to it in cyan
- Update in place using `\r` — not new lines
- Show green checkmark `✓` when done
- Run for the given duration then stop

```
⠹ Loading data...    ← cyan, updates in place
✓ Loading data...    ← green checkmark when done
```

Start with Task 1 and paste your solution.


## Concept 5 — String Searching: the `strings` package

### What it is and why it exists

Strings are the most common data type in real programs — user input, file content, API responses, config values. Go's `strings` package gives you every tool you need to search, inspect, split, replace, and transform strings.

`strings.ContainsRune` is one function in a large family. Understanding the whole family makes you a much stronger Go developer. We cover all the important ones here.

---

### The difference between byte, rune, and string

Before searching strings you must understand what you are searching through:

```go
s := "Hello, 世界"

// bytes — raw memory
len(s)              // 13 — because 世 and 界 are 3 bytes each in UTF-8

// runes — Unicode characters
len([]rune(s))      // 9  — 7 ASCII chars + 2 Chinese chars

// ranging over a string gives runes
for i, ch := range s {
    fmt.Printf("index %d: %c (rune %d)\n", i, ch, ch)
}
// index 0: H (rune 72)
// index 1: e (rune 101)
// ...
// index 7: 世 (rune 19990)
// index 10: 界 (rune 30028)
```

This matters for searching because byte-based functions can split multi-byte characters. Rune-based functions always treat characters as whole units.

---

### `strings.ContainsRune` — does this character exist in this string

```go
strings.ContainsRune(s string, r rune) bool
```

Answers one question: **does rune `r` appear anywhere in string `s`?**

```go
strings.ContainsRune("Hello", 'H')    // true
strings.ContainsRune("Hello", 'h')    // false — case sensitive
strings.ContainsRune("Hello", 'z')    // false
strings.ContainsRune("He", 'e')       // true
strings.ContainsRune("", 'H')         // false — empty string
```

**Why a string as container and rune as target:**

The container is a `string` because that is what you search through. The target is a `rune` because characters in Go are runes — single Unicode code points. This lets you search for any character including non-ASCII:

```go
strings.ContainsRune("café", 'é')     // true
strings.ContainsRune("日本語", '本')   // true
strings.ContainsRune("Hello", '世')   // false
```

**In your ASCII art project:**

```go
// should this character be colored?
shouldColor := colorCode != "" &&
    (letters == "" || strings.ContainsRune(letters, ch))
// letters is "He" — a string of target characters
// ch is the current character being rendered — a rune
// ContainsRune checks: is ch one of the letters to color?
```

---

### `strings.Contains` — does this substring exist

```go
strings.Contains(s string, substr string) bool
```

Searches for a whole substring, not just one character:

```go
strings.Contains("Hello World", "World")   // true
strings.Contains("Hello World", "world")   // false — case sensitive
strings.Contains("Hello World", "")        // true — empty string always found
strings.Contains("Hello World", "xyz")     // false
```

**Real world uses:**

```go
// check if URL has a scheme
strings.Contains(url, "://")

// check if error message mentions specific issue
strings.Contains(err.Error(), "connection refused")

// check if file is a Go file
strings.Contains(filename, ".go")

// check if HTML has a specific tag
strings.Contains(html, "<script>")
```

---

### `strings.HasPrefix` and `strings.HasSuffix`

```go
strings.HasPrefix(s string, prefix string) bool
strings.HasSuffix(s string, suffix string) bool
```

```go
strings.HasPrefix("Hello World", "Hello")   // true
strings.HasPrefix("Hello World", "World")   // false
strings.HasSuffix("Hello World", "World")   // true
strings.HasSuffix("Hello World", "Hello")   // false

// file type checking
strings.HasSuffix("photo.jpg", ".jpg")      // true
strings.HasSuffix("document.pdf", ".pdf")   // true

// URL checking
strings.HasPrefix("https://example.com", "https://")  // true
strings.HasPrefix("http://example.com",  "https://")  // false

// flag parsing (before flag package existed)
strings.HasPrefix(arg, "--output=")  // true for "--output=file.txt"
```

---

### `strings.Index` and `strings.LastIndex`

```go
strings.Index(s string, substr string) int
```

Returns the **index** of the first occurrence of `substr` in `s`. Returns `-1` if not found:

```go
strings.Index("Hello World", "World")   // 6
strings.Index("Hello World", "o")       // 4  — first 'o'
strings.Index("Hello World", "xyz")     // -1
strings.Index("Hello World", "")        // 0  — empty string at position 0
```

```go
strings.LastIndex("Hello World", "o")   // 7  — last 'o'
strings.LastIndex("Hello World", "l")   // 9  — last 'l'
```

**Real world uses:**

```go
// find where query string starts in URL
url := "https://example.com/search?q=golang"
idx := strings.Index(url, "?")
if idx != -1 {
    path  := url[:idx]          // "https://example.com/search"
    query := url[idx+1:]        // "q=golang"
}

// find file extension
filename := "document.report.pdf"
idx = strings.LastIndex(filename, ".")
if idx != -1 {
    ext  := filename[idx:]   // ".pdf"
    name := filename[:idx]   // "document.report"
}
```

---

### `strings.IndexRune` — find a single character

```go
strings.IndexRune(s string, r rune) int
```

Like `Index` but for a single rune. Returns position or `-1`:

```go
strings.IndexRune("Hello", 'e')    // 1
strings.IndexRune("Hello", 'z')    // -1
strings.IndexRune("café", 'é')     // 3
```

---

### `strings.Count` — how many times does substring appear

```go
strings.Count(s string, substr string) int
```

```go
strings.Count("Hello World", "l")      // 3
strings.Count("Hello World", "o")      // 2
strings.Count("Hello World", "xyz")    // 0
strings.Count("aaa", "aa")             // 1  — non-overlapping

// count words (split on spaces)
strings.Count("one two three", " ") + 1   // 3
```

---

### `strings.Split` and `strings.SplitN`

```go
strings.Split(s string, sep string) []string
```

Splits a string at every occurrence of `sep`:

```go
strings.Split("a,b,c,d", ",")          // ["a", "b", "c", "d"]
strings.Split("Hello World", " ")      // ["Hello", "World"]
strings.Split("Hello", "")             // ["H", "e", "l", "l", "o"]
strings.Split("a,,b", ",")             // ["a", "", "b"]  — empty string preserved
strings.Split("Hello", "xyz")          // ["Hello"]  — not found, returns whole string
```

`strings.SplitN` limits the number of pieces:

```go
strings.SplitN("a,b,c,d", ",", 2)     // ["a", "b,c,d"]  — max 2 pieces
strings.SplitN("a,b,c,d", ",", 3)     // ["a", "b", "c,d"]
```

**Real world uses:**

```go
// parse CSV line
fields := strings.Split("Michael,Lagos,25", ",")
// ["Michael", "Lagos", "25"]

// parse key=value pair
parts := strings.SplitN("name=Michael Obi", "=", 2)
key   := parts[0]  // "name"
value := parts[1]  // "Michael Obi"

// parse path segments
segments := strings.Split("/api/users/123", "/")
// ["", "api", "users", "123"]
```

---

### `strings.Fields` — split on any whitespace

```go
strings.Fields(s string) []string
```

Splits on any whitespace (spaces, tabs, newlines) and ignores leading/trailing whitespace:

```go
strings.Fields("Hello   World")        // ["Hello", "World"]
strings.Fields("  one  two  three  ")  // ["one", "two", "three"]
strings.Fields("Hello\tWorld\nGo")     // ["Hello", "World", "Go"]
strings.Fields("")                     // []  — empty slice
```

Use `Fields` over `Split(s, " ")` when input might have irregular spacing.

---

### `strings.Replace` and `strings.ReplaceAll`

```go
strings.Replace(s, old, new string, n int) string
strings.ReplaceAll(s, old, new string) string
```

```go
strings.Replace("aabbcc", "b", "X", 1)     // "aaXbcc"  — replace first only
strings.Replace("aabbcc", "b", "X", 2)     // "aaXXcc"  — replace first two
strings.Replace("aabbcc", "b", "X", -1)    // "aaXXcc"  — replace all (-1 = unlimited)
strings.ReplaceAll("aabbcc", "b", "X")     // "aaXXcc"  — replace all (cleaner)
```

**Real world uses:**

```go
// sanitise user input
clean := strings.ReplaceAll(input, "<script>", "")

// normalise line endings
normalised := strings.ReplaceAll(content, "\r\n", "\n")

// template substitution
template := "Hello, {name}! You have {count} messages."
result := strings.ReplaceAll(template, "{name}", "Michael")
result  = strings.ReplaceAll(result,   "{count}", "5")
// "Hello, Michael! You have 5 messages."
```

---

### `strings.TrimSpace`, `strings.Trim`, `strings.TrimPrefix`, `strings.TrimSuffix`

```go
strings.TrimSpace("  Hello World  ")    // "Hello World"
strings.Trim("***Hello***", "*")        // "Hello"
strings.TrimLeft("***Hello***", "*")    // "Hello***"
strings.TrimRight("***Hello***", "*")   // "***Hello"
strings.TrimPrefix("Hello World", "Hello ")  // "World"
strings.TrimSuffix("Hello World", " World")  // "Hello"
```

**Key difference between `Trim` and `TrimPrefix`:**

```go
// Trim removes any combination of characters from the set
strings.Trim("aabHelloaab", "ab")  // "Hello" — removes any a or b from edges

// TrimPrefix removes the exact prefix string
strings.TrimPrefix("aabHello", "aab")  // "Hello" — removes "aab" exactly
strings.TrimPrefix("aabHello", "ab")   // "aabHello" — "ab" is not the prefix
```

---

### `strings.ToUpper`, `strings.ToLower`, `strings.Title`

```go
strings.ToUpper("hello")    // "HELLO"
strings.ToLower("HELLO")    // "hello"
strings.Title("hello world") // "Hello World"  — capitalises first letter of each word
```

---

### `strings.Join` — opposite of Split

```go
strings.Join([]string{"a", "b", "c"}, ", ")   // "a, b, c"
strings.Join([]string{"Hello", "World"}, " ")  // "Hello World"
strings.Join([]string{"a", "b", "c"}, "")      // "abc"
```

---

### `strings.Repeat`

```go
strings.Repeat("ab", 3)    // "ababab"
strings.Repeat("-", 20)    // "--------------------"
strings.Repeat(" ", n)     // n spaces — used for padding
```

---

### `strings.EqualFold` — case insensitive comparison

```go
strings.EqualFold("Hello", "hello")   // true
strings.EqualFold("Hello", "HELLO")   // true
strings.EqualFold("Hello", "World")   // false

// better than:
strings.ToLower("Hello") == strings.ToLower("hello")  // works but creates two new strings
```

---

### Real world examples putting it all together

**Example 1 — input validator:**

```go
func validateEmail(email string) error {
    email = strings.TrimSpace(email)

    if email == "" {
        return fmt.Errorf("email cannot be empty")
    }
    if !strings.Contains(email, "@") {
        return fmt.Errorf("email must contain @")
    }
    if !strings.Contains(email, ".") {
        return fmt.Errorf("email must contain a dot")
    }

    parts := strings.Split(email, "@")
    if len(parts) != 2 {
        return fmt.Errorf("email must have exactly one @")
    }
    if parts[0] == "" {
        return fmt.Errorf("email must have a username before @")
    }
    if parts[1] == "" {
        return fmt.Errorf("email must have a domain after @")
    }

    return nil
}

emails := []string{
    "michael@example.com",
    "notanemail",
    "missing@dot",
    "@nodomain.com",
    "  spaces@example.com  ",
}

for _, e := range emails {
    if err := validateEmail(e); err != nil {
        fmt.Printf("INVALID: %s — %v\n", e, err)
    } else {
        fmt.Printf("VALID:   %s\n", strings.TrimSpace(e))
    }
}
```

---

**Example 2 — URL parser:**

```go
func parseURL(url string) map[string]string {
    result := make(map[string]string)

    // scheme
    if idx := strings.Index(url, "://"); idx != -1 {
        result["scheme"] = url[:idx]
        url = url[idx+3:]
    }

    // query string
    if idx := strings.Index(url, "?"); idx != -1 {
        result["query"] = url[idx+1:]
        url = url[:idx]
    }

    // path
    if idx := strings.Index(url, "/"); idx != -1 {
        result["path"] = url[idx:]
        url = url[:idx]
    }

    result["host"] = url
    return result
}

parsed := parseURL("https://example.com/api/users?page=1&limit=10")
fmt.Println(parsed["scheme"])  // https
fmt.Println(parsed["host"])    // example.com
fmt.Println(parsed["path"])    // /api/users
fmt.Println(parsed["query"])   // page=1&limit=10
```

---

**Example 3 — word frequency counter:**

```go
func wordFrequency(text string) map[string]int {
    freq := make(map[string]int)
    words := strings.Fields(strings.ToLower(text))

    for _, word := range words {
        // remove punctuation
        word = strings.Trim(word, ".,!?;:\"'")
        if word != "" {
            freq[word]++
        }
    }

    return freq
}

text := "Go is great. Go is fast. Go is simple and Go is fun!"
freq := wordFrequency(text)
for word, count := range freq {
    fmt.Printf("%s: %d\n", word, count)
}
// go: 4
// is: 4
// great: 1
// fast: 1
// simple: 1
// and: 1
// fun: 1
```

---

**Example 4 — template engine:**

```go
func render(template string, data map[string]string) string {
    result := template
    for key, value := range data {
        placeholder := "{{" + key + "}}"
        result = strings.ReplaceAll(result, placeholder, value)
    }
    return result
}

template := `Dear {{name}},

Your order {{order_id}} has been shipped to {{address}}.
Expected delivery: {{date}}.

Thank you for shopping with us!`

data := map[string]string{
    "name":     "Michael",
    "order_id": "ORD-2026-001",
    "address":  "Lagos, Nigeria",
    "date":     "May 20, 2026",
}

fmt.Println(render(template, data))
```

---

### Connection to your ASCII art project

```go
// SplitInput — splits on literal \n
func SplitInput(str string) []string {
    return strings.Split(str, "\\n")
}

// LoadBanner — normalise line endings
content = strings.ReplaceAll(string(data), "\r\n", "\n")
content = strings.TrimPrefix(content, "\n")
lines   = strings.Split(content, "\n")

// Generate — validate characters
if _, ok := banner[ch]; !ok {
    fmt.Fprintf(os.Stderr, "unknown character %q\n", ch)
}

// renderLines — target specific characters
shouldColor := colorCode != "" &&
    (letters == "" || strings.ContainsRune(letters, ch))
```

---

### 🏋️ Mini Tasks

**Task 1 — Basic:**
Write a function `isPalindrome(s string) bool` that returns true if a string reads the same forwards and backwards. Ignore case and spaces.

```go
isPalindrome("racecar")         // true
isPalindrome("Race Car")        // true  — ignore case and spaces
isPalindrome("hello")           // false
isPalindrome("A man a plan a canal Panama")  // true
```

Hint — use `strings.ToLower`, `strings.ReplaceAll`, then compare the string to its reverse.

**Task 2 — Medium:**
Write a function `parseQueryString(query string) map[string]string` that parses a URL query string into a map:

```go
parseQueryString("name=Michael&city=Lagos&age=25")
// map[string]string{
//     "name": "Michael",
//     "city": "Lagos",
//     "age":  "25",
// }

parseQueryString("search=hello+world&page=2")
// map[string]string{
//     "search": "hello world",   ← + becomes space
//     "page":   "2",
// }
```

Use `strings.Split` to split on `&`, then `strings.SplitN` to split each pair on `=`.

**Task 3 — Hard:**
Write a function `highlight(code string, keywords []string, color string) string` that wraps every occurrence of every keyword in the given ANSI color:

```go
keywords := []string{"func", "var", "return", "if", "for"}
result   := highlight(goCode, keywords, "\033[34m")  // blue
```

Rules:
- Only highlight whole words — `"format"` should not highlight the `"for"` inside it
- Case sensitive
- Must work for any list of keywords
- Reset color after each keyword

Hint — use `strings.Fields` to walk word by word and `strings.ContainsRune` or direct comparison to check each word.

Start with Task 1 and paste your solution.