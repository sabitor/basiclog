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
<date of local time zone> <time of the local time zone> <prefix> <log message>
```

## How to use simplelog
Using the log framework is pretty easy. After the log service has been started once, any number of log message write calls can be triggered until the log service is  explicitly stopped.

**Hint:** If log messages will only be sent to standard out, there is no need to setup a log file. If, on the other hand, it should also be written to a log file, the log file has to be initialized once before writing log messages to it.

Let's have a look at the following sample application, which uses the simplelog framework:
```go
package main

import "github.com/sabitor/simplelog"

func main() {
    logBuffer := 10 // number of log messages which can be buffered before the log service blocks
    simplelog.StartService(logBuffer)
    defer simplelog.StopService()
    
    simplelog.WriteToStdout("Start application")
    simplelog.InitLogFile("log1.txt")
    simplelog.WriteToFile("[DEV]", "First message to FILE.")
    simplelog.WriteToMulti("[DEV]", "First message to MULTI.")
    simplelog.ChangeLogFile("log2.txt")
    simplelog.WriteToFile("[DEV]", "Second message to File.")
    simplelog.WriteToMulti("[TEST]", "First message to MULTI.")
    simplelog.WriteToStdout("Stop application")
}
```
The following log output was generated:
**STDOUT**
```
Start application
[DEV] First dev-message to MULTI.
[TEST] First test-message to MULTI.
Stop application
```
**Log file log1.txt**
```
2023/04/11 13:38:41.369607 [DEV] First dev-message to FILE.
2023/04/11 13:38:41.369807 [DEV] First dev-message to MULTI.
```
**Log file log2.txt**
```
2023/04/11 13:38:41.370075 [DEV] Second dev-message to File.
2023/04/11 13:38:41.370138 [TEST] First test-message to MULTI.
```


