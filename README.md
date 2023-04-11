# simplelog
simplelog is a logging package with the focus on simplicity and ease of use. It utilizes the log package from the standard library with some advanced features.
Once started, the simple logger runs as a service and listens for logging requests.
The simple logger writes log records to either standard out, a log file, or standard out and a log file simultaneously.

A log record to standard out consists of the following format:
<log message>

A log record to a log file consists of the following format:
<date of local time zone> <time of the local time zone> <prefix> <log message>
