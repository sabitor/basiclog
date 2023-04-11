# Log framework - simplelog
simplelog is a log framework with the focus on simplicity and ease of use. It utilizes the log package from the standard library with some advanced features.
Once started, the simple logger runs as a service and listens for logging requests.
The simple logger writes log records to either standard out, a log file, or standard out and a log file simultaneously.

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

**Hint:** If log messages will only be sent to standard out, there is no need to setup a log file. If, on the other hand, it should also be written to a log file, the log service has to be configured once before writing log messages to the log file.
