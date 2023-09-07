# Simplelog
Simplelog is a log framework mainly with a focus on simplicity, ease of use and performance.

Once started, the simple logger runs as a service and listens for logging requests.
The simple logger writes log records to either standard out, a log file, or standard out and a log file simultaneously (multi log).

## simplelog API
In order to use or work with the simplelog package, the following set of functions were exposed to be used as the simplelog API: 

```
// SetPrefix sets the prefix for log records.
func SetPrefix(destination int, prefix ...string)

// Shutdown stops the log service including post-processing and cleanup.
func Shutdown(archivelog bool)

// Startup starts the log service.
func Startup(logName string, appendLog bool, bufferSize int)

// SwitchLog closes the current log file and a new log file with the specified name is created and used.
func SwitchLog(newLogName string)

// Write writes a log message to a specified destination.
// Possible destinations are STDOUT, FILE or MULTI (a combination of STDOUT and FILE).
func Write(destination int, values ...any)
```

## How to use simplelog
Using the simplelog framework is pretty easy. After the log service has been started by calling the *Startup* function, any number of *Write* function calls can be triggered until the log service has been explicitly stopped by calling the *Shutdown* function, for example:

	Startup(...)
 	Write(...)
	...
 	Write(...)
  	Shutdown(...)

**Hint:** 
1) The appearance of a log line can be adjusted by specifying prefixes. These prefixes can be defined independently for the standard out logger and the file logger by calling the *SetPrefix* function. If the prefix should also contain actual date and time data, the Golang *reference time placeholders* can be applied for given data:

	| Time Item | Placeholder |
	| -------- | ------- |
	| Year | 2006 |
	| Month | 01 |
	| Day | 02 |
	| Hour | 15 |
	| Minute | 04 |
	| Second | 05 |
	| Millisecond | 000000 |

	In addition, to distinguish and parse date and time information, the reference time string has to be delimited by the prefix and suffix tag #, for example: #2006-01-02 15:04:05.000000#. Then, all placeholders are replaced at runtime by the logging service accordingly.

	Note that not all placeholders have to be used and they can be used in any order.

3) The log file used by the log service can be changed by calling the *SwitchLog* function. Thereby, the current log is closed (not deleted) and a new log file with the specified name is created (a file with the new name must not already exist). The log service does not have to be stopped for this purpose.
4) Log files can also be archived automatically when the log service is shut down. In such a case, the closed log file is renamed as follows: \<log file name\>_yyyymmddHHMMSS, whereas *yyyymmddHHMMSS* denotes the timestamp when the rename of the log occurred.

**Example:** 
```go
package main

import (
	"strconv"
	"sync"

	"github.com/sabitor/simplelog"
)

func main() {
    log1 := "log1.txt"
    logBuffer := 2 // number of log messages which can be buffered before the log service blocks
    simplelog.Startup(log1, false, logBuffer)
    defer simplelog.Shutdown(false)

    simplelog.SetPrefix(simplelog.STDOUT, "STDOUT$")
    simplelog.Write(simplelog.STDOUT, ">>> Start application")
    simplelog.SetPrefix(simplelog.FILE, "#02/01/2023 15:04:05.000000#", "-")
    simplelog.Write(simplelog.STDOUT, "Log file is", log1)
    simplelog.Write(simplelog.FILE, "[MAIN]", "Write", 1, "to FILE.")
    simplelog.Write(simplelog.MULTI, "[MAIN]", "Write", 1, "to MULTI.")
    
    log2 := "log2.txt"
    simplelog.SwitchLog(log2)
    simplelog.Write(simplelog.STDOUT, "New log file is", log2)

    var wg sync.WaitGroup
    for i := 1; i <= 4; i++ {
        wg.Add(1)
        go func(count int) {
            defer wg.Done()
	    context := "[GOROUTINE " + strconv.Itoa(count) + "]"
	    simplelog.Write(simplelog.FILE, context, "Write", count+1, "to FILE.")
        }(i)
    }
    wg.Wait()

    simplelog.Write(simplelog.MULTI, "[MAIN]", "Write", 2, "to MULTI.")
    simplelog.Write(simplelog.STDOUT, "<<< Stop application")
}
```

The following log output was generated:

**Standard out**
```
STDOUT$ >>> Start application
STDOUT$ Log file is log1.txt
STDOUT$ [MAIN] Write 1 to MULTI.
STDOUT$ New log file is log2.txt
STDOUT$ [MAIN] Write 2 to MULTI.
STDOUT$ <<< Stop application
```
**Log file log1.txt**
```
2023/04/14 08:49:02.555266 - [MAIN] Write 1 to FILE.
2023/04/14 08:49:02.555332 - [MAIN] Write 1 to MULTI.
```
**Log file log2.txt**
```
2023/04/14 08:49:02.555448 - [GOROUTINE 4] Write 5 to FILE.
2023/04/14 08:49:02.555456 - [GOROUTINE 1] Write 2 to FILE.
2023/04/14 08:49:02.555460 - [GOROUTINE 3] Write 4 to FILE.
2023/04/14 08:49:02.555562 - [GOROUTINE 2] Write 3 to FILE.
2023/04/14 08:49:02.555604 - [MAIN] Write 2 to MULTI.
```


