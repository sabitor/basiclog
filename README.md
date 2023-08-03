# Log framework - simplelog
simplelog is a log framework mainly with a focus on simplicity, ease of use and performance.

Once started, the simple logger runs as a service and listens for logging requests.
The simple logger writes log records to either standard out, a log file, or standard out and a log file simultaneously (multi log).

## simplelog API
In order to use or work with the simplelog package, the following set of functions were exposed to be used as the simplelog API: 

```
// Log writes a log message to a specified destination.
// Possible destinations are STDOUT, FILE or MULTI (a combination of STDOUT and FILE).
func Log(destination int, values ...any)

// SetPrefix sets the prefix for logging lines.
func SetPrefix(destination int, prefix string)

// Shutdown stops the log service including post-processing and cleanup.
func Shutdown(archivelog bool)

// Startup starts the log service.
func Startup(logName string, appendLog bool, bufferSize int)

// SwitchLog closes the current log file and a new log file with the specified name is created and used.
func SwitchLog(newLogName string)
```

## How to use simplelog
Using the log framework is pretty easy. After the log service has been started once, any number of log message write calls can be triggered until the log service is  explicitly stopped.

**Hint:** 
1) The appearance of a log line can be adjusted by specifying prefixes. These prefixes can be defined independently for the standard out logger and the file logger. If the prefix should also contain actual date and time data, the following *placeholders* can be applied for given data:

	 - Year: yyyy
	 - Month: mm
	 - Day: dd
	 - Hour: HH
	 - Minute: MI
	 - Second: SS
	 - Millisecond: FFFFFF

	In addition, to distinguish and parse date and time information, placeholders have to be delimited by __\<DT\>...\<DT\>__ tags and can be used for example as follows: \<DT\>yyyy-mm-dd HH:MI:SS.ffffff\<DT\>. All placeholders are replaced at runtime by the logging service accordingly.

	Note that not all placeholders have to be used, they can be used in any order and even non-datetime characters or strings can be integrated.

2) The log file used by the log service can be changed by calling the *SwitchLog* function. Thereby, the current log is closed (not deleted) and a new log file with the specified name is created (a file with the new name must not already exist). The log service does not have to be stopped for this purpose.
3) Log files can also be archived automatically when the log service is shut down. In such a case, the closed log file is renamed as follows: \<log file name\>_yyyymmddHHMMSS, whereas *yyyymmddHHMMSS* denotes the timestamp when the rename of the log occurred.

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
    simplelog.Log(simplelog.STDOUT, ">>> Start application")
    simplelog.SetPrefix(simplelog.FILE, "<DT>dd/mm/yyyy HH:MI:SS.FFFFFF<DT>")
    simplelog.Log(simplelog.STDOUT, "Log file is", log1)
    simplelog.Log(simplelog.FILE, "[MAIN]", "Write", 1, "to FILE.")
    simplelog.Log(simplelog.MULTI, "[MAIN]", "Write", 1, "to MULTI.")
    
    log2 := "log2.txt"
    simplelog.SwitchLog(log2)
    simplelog.Log(simplelog.STDOUT, "New log file is", log2)

    var wg sync.WaitGroup
    for i := 1; i <= 4; i++ {
        wg.Add(1)
        go func(count int) {
            defer wg.Done()
	    context := "[GOROUTINE " + strconv.Itoa(count) + "]"
	    simplelog.Log(simplelog.FILE, context, "Write", count+1, "to FILE.")
        }(i)
    }
    wg.Wait()

    simplelog.Log(simplelog.MULTI, "[MAIN]", "Write", 2, "to MULTI.")
    simplelog.Log(simplelog.STDOUT, "<<< Stop application")
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
2023/04/14 08:49:02.555266 [MAIN] Write 1 to FILE.
2023/04/14 08:49:02.555332 [MAIN] Write 1 to MULTI.
```
**Log file log2.txt**
```
2023/04/14 08:49:02.555448 [GOROUTINE 4] Write 5 to FILE.
2023/04/14 08:49:02.555456 [GOROUTINE 1] Write 2 to FILE.
2023/04/14 08:49:02.555460 [GOROUTINE 3] Write 4 to FILE.
2023/04/14 08:49:02.555562 [GOROUTINE 2] Write 3 to FILE.
2023/04/14 08:49:02.555604 [MAIN] Write 2 to MULTI.
```


