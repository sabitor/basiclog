# Log framework - simplelog
simplelog is a log framework mainly with a focus on simplicity, ease of use and performance.

Once started, the simple logger runs as a service and listens for logging requests.
The simple logger writes log records to either standard out, a log file, or standard out and a log file simultaneously (multi log).

## simplelog API
In order to use or work with the simplelog package, the following functions were exposed to be used as the simplelog API: 

```
func Log(destination int, logValues ...any)
```
Log writes a log message to a specified destination.
The destination parameter specifies the log destination, where the data will be written to.
The logValues parameter consists of one or multiple values that are logged.

```
func SetPrefix(destination int, prefix string)
```
SetPrefix sets the prefix for logging lines.
The destination specifies the name of the log destination where the prefix should be used, e.g. STDOUT or FILE.
The prefix specifies the prefix for each logging line for a given log destination.

```
func Shutdown(archivelog bool)
```
Shutdown stops the log service and does some cleanup.
Before the log service is stopped, all pending log messages are flushed and resources are released.
If enabled, the closed log can also be archived. In this context archiving a log file means that it will be renamed and no new messages will be appended on a new run. The archived log file is of the following format: \<orig log name\>_yyyymmddHHMMSS.
The archivelog flag indicates whether the log file will be archived (true) or not (false).


```
func Startup(bufferSize int)
```
Startup starts the log service.
The log service runs in its own goroutine.
The bufferSize specifies the number of log messages which can be buffered before the log service blocks.

```
func SwitchLog(newLogName string)
```
SwitchLog closes the current log file and a new log file with the specified name is created and used.
Thereby, the current log file is not deleted, the new log file must not exist and the log service doesn't need to be stopped for this task. The new log file must not exist.
The newLogName specifies the name of the new log to switch to.

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
	 - Millisecond: f[5f]

	In addition, to distinguish and parse date and time information, placeholders have to be delimited by __\<DT\>...\<DT\>__ tags and can be used for example as follows: \<DT\>yyyy-mm-dd HH:MI:SS.ffffff\<DT\>. All placeholders are replaced at runtime by the logging service accordingly.

	Note that not all placeholders have to be used, they can be used in any order and even non-datetime characters or strings can be integrated.

3) If log messages will only be sent to standard out, there is no need to setup a log file. If, on the other hand, it should also be written to a log file, the log file has to be initialized once by calling the *InitLog* function before log messages can be written to the log file.
4) The log file used by the log service can be changed by calling the *SwitchLog* function. Thereby, the current log is closed (not deleted) and a new log file with the specified name is created (a file with the new name must not already exist). The log service does not have to be stopped for this purpose.
5) Log files can also be archived automatically when the log service is shut down. In such a case, the closed log file is renamed as follows: \<log file name\>_yyyymmddHHMMSS, whereas *yyyymmddHHMMSS* denotes the timestamp when the rename of the log occurred.

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


