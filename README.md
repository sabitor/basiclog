# Log framework - simplelog
simplelog is a log framework mainly with a focus on simplicity and usability.

It utilizes the log package from the standard library with some advanced features.
Once started, the simple logger runs as a service and listens for logging requests.
The simple logger writes log records to either standard out, a log file, or standard out and a log file simultaneously (multi log).

A log entry to standard out consists of the following format:
```
<log message>
```

A log entry in a log file consists of the following format:
```
<date of local time zone> <time of the local time zone> <log message>
```

## How to use simplelog
Using the log framework is pretty easy. After the log service has been started once, any number of log message write calls can be triggered until the log service is  explicitly stopped.

**Hint:** 
1) If log messages will only be sent to standard out, there is no need to setup a log file. If, on the other hand, it should also be written to a log file, the log file has to be initialized once by calling the *InitLogFile* function before log messages can be written to the log file.
2) The log file used by the log service can be changed by calling the *ChangeLogFile* function. The log service does not have to be stopped for this purpose.

Let's take a look at the following example application that shows how to use the simplelog framework:
```go
package main

import (
	"strconv"
	"sync"

	"github.com/sabitor/simplelog"
)

func main() {
    logBuffer := 2 // number of log messages which can be buffered before the log service blocks
    simplelog.Startup(logBuffer)
    defer simplelog.Shutdown()

    simplelog.WriteToStdout(">>> Start application")
    log1 := "log1.txt"
    simplelog.InitLogFile(log1, false)
    simplelog.WriteToStdout("Log file is", log1)
    simplelog.WriteToFile("[MAIN]", "Write", 1, "to FILE.")
    simplelog.WriteToMulti("[MAIN]", "Write", 1, "to MULTI.")
    
    log2 := "log2.txt"
    simplelog.NewLogName(log2)
    simplelog.WriteToStdout("New log file is", log2)

    var wg sync.WaitGroup
    for i := 1; i <= 4; i++ {
        wg.Add(1)
        go func(count int) {
            defer wg.Done()
	    prefix := "[GOROUTINE " + strconv.Itoa(count) + "]"
	    simplelog.WriteToFile(prefix, "Write", count+1, "to FILE.")
        }(i)
    }
    wg.Wait()

    simplelog.WriteToMulti("[MAIN]", "Write", 2, "to MULTI.")
    simplelog.WriteToStdout("<<< Stop application")
}
```

The following log output was generated:

**Standard out**
```
>>> Start application
Log file is log1.txt
[MAIN] Write 1 to MULTI.
New log file is log2.txt
[MAIN] Write 2 to MULTI.
<<< Stop application
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


