# Log framework - simplelog
simplelog is a log framework with the focus on simplicity and ease of use. It utilizes the log package from the standard library with some advanced features.
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

Let's have a look at the following sample application, who uses the simplelog framework as an example:
```go
package main

import (
	"sync"

	"github.com/sabitor/simplelog"
)

func main() {
	logBuffer := 10 // number of log messages which can be buffered before the log service blocks
	simplelog.StartService(logBuffer)
	defer simplelog.StopService()

	simplelog.WriteToStdout("Start application")
	simplelog.InitLogFile("log1.txt")
	simplelog.WriteToFile("[MAIN]", "First message to FILE.")
	simplelog.WriteToMulti("[MAIN]", "First message to MULTI.")
	
	simplelog.ChangeLogFile("log2.txt")
	simplelog.WriteToStdout("Changed log file")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
	    defer wg.Done()
	    simplelog.WriteToFile("[GOROUTINE]", "Second message to File.")
	}()
	wg.Wait()

	simplelog.WriteToMulti("[MAIN]", "Second message to MULTI.")
	simplelog.WriteToStdout("Stop application")
}
```

The following log output was generated:

**Standard out**
```
Start application
[MAIN] First message to MULTI.
Changed log file
[MAIN] Second message to MULTI.
Stop application
```
**Log file log1.txt**
```
2023/04/13 10:20:37.164884 [MAIN] First message to FILE.
2023/04/13 10:20:37.165094 [MAIN] First message to MULTI.
```
**Log file log2.txt**
```
2023/04/13 10:20:37.165285 [GOROUTINE] Second message to File.
2023/04/13 10:20:37.165327 [MAIN] Second message to MULTI.
```


