// Package simplelog is a logging package with the focus on simplicity and
// ease of use. It utilizes the log package from the standard library with
// some advanced features.
// Once started, the simple logger runs as a service and listens for logging
// requests through the functions WriteTo[Stdout|File|Multiple].
// As the name of the WriteTo functions suggests, the simple logger writes
// to either standard out, a log file, or multiple targets.
package simplelog

import (
	"fmt"
	"log"
	"os"
	"strings"
)

// message catalog
const (
	sl001e = "log file name not set"
	sl002e = "log service was already started"
	sl003e = "log service is not running"
	sl004e = "log service has not been started"
	sl005e = "log file name already set"
)

// log targets
const (
	stdout = 1 << iota     // write the log record to STDOUT
	file                   // write the log record to the log file
	multi  = stdout | file // write the log record to STDOUT and to the log file
)

// configuration categories
const (
	openlog       = iota // open a log file
	changelogname        // change the log file name
)

// service states
const (
	stopped = iota // indicator of a stopped log service
	running        // indicator of a running (active) log service
)

// A signal will be used as a trigger to handle certain tasks by the log service, whereby no payload data has to be sent the recipient.
type signal struct{}

// A logMessage represents the log message which will be sent to the log service.
type logMessage struct {
	target int    // the log target bits, e.g. stdout, file, and so on.
	prefix string // the log prefix, which will be written in front of the log record
	record string // the payload of the log message, which will be sent to the log target
}

// A configMessage represents the config message which will be sent to the log service.
type configMessage struct {
	category int    // the configuration category bits, which are used to trigger certain config tasks by the log service, e.g. setlogname, changelogname, and so on.
	data     string // the data, which will be processed by a config task
}

// A simpleLog represents an instance of a simple logger.
type simpleLog struct {
	// handler
	fileHandle *os.File                       // the log file handle
	logHandle  map[int]map[string]*log.Logger // a map which stores for every log target bit a map which stores the log prefix and its assigned log handle

	// channels
	data           chan logMessage    // the channel for sending log messages to the log service; this channel will be a buffered channel
	config         chan configMessage // the channel for sending config messages to the log service
	stopLogService chan signal        // the channel for sending a stop message to the log service

	// service
	state int // to save the current state of the log service repesented by the service bits, e.g. stopped, running, and so on
}

// global (package) variables
var sLog = &simpleLog{}
var firstFileLogHandler = false

// setServiceState sets the state of the log service.
// The state bits are stopped, running, and so on.
func (sl *simpleLog) setServiceState(newState int) {
	sl.state = newState
}

// serviceState returns the state of the log service.
// The returned state bits are stopped, running, and so on.
func (sl *simpleLog) serviceState() int {
	return sl.state
}

// stdoutLog returns a stdout log handle.
func (sl *simpleLog) stdoutLog(prefix string) *log.Logger {
	return sLog.handle(stdout, prefix)
}

// fileLog returns a file log handle.
func (sl *simpleLog) fileLog(prefix string) *log.Logger {
	return sLog.handle(file, prefix)
}

// multiLog returns a stdout log handle and a file log handle.
func (sl *simpleLog) multiLog(prefix string) (*log.Logger, *log.Logger) {
	return sLog.handle(stdout, prefix), sLog.handle(file, prefix)
}

// openLogFile opens a log file.
func (sl *simpleLog) openLogFile(logName string) {
	var err error
	sLog.fileHandle, err = os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
}

// changeLogFileName changes the name of the log file.
func (sl *simpleLog) changeLogFileName(newLogName string) {
	// remove all file log handles from the logHandler map which are linked to the old log name
	delete(sLog.logHandle, file)
	sLog.fileHandle.Close()
	firstFileLogHandler = false
	sLog.openLogFile(newLogName)
}

// service receives and handles messages.
// This service function runs in a dedicated goroutine and will be started as part of the log service startup process.
// It handles the following messages:
//   - logMessage
//   - configMessage
//   - signal
func service() {
	for {
		select {
		case <-sLog.stopLogService:
			sLog.setServiceState(stopped)
			return
		case logMsg := <-sLog.data:
			switch logMsg.target {
			case stdout:
				stdoutLogHandle := sLog.stdoutLog(logMsg.prefix)
				stdoutLogHandle.Print(logMsg.record)
			case file:
				fileLogHandle := sLog.fileLog(logMsg.prefix)
				if fileLogHandle != nil {
					fileLogHandle.Print(logMsg.record)
				} else {
					panic(sl001e)
				}
			case multi:
				stdoutLogHandle, fileLogHandle := sLog.multiLog(logMsg.prefix)
				stdoutLogHandle.Print(logMsg.record)
				if fileLogHandle != nil {
					fileLogHandle.Print(logMsg.record)
				} else {
					panic(sl001e)
				}
			}
		case cfgMsg := <-sLog.config:
			switch cfgMsg.category {
			case openlog:
				if sLog.fileHandle == nil {
					sLog.openLogFile(cfgMsg.data)
				} else {
					panic(sl005e)
				}
			case changelogname:
				sLog.changeLogFileName(cfgMsg.data)
			}
		}
	}
}

// handle returns a log handle for a given combination of log target and message prefix.
// Each combination of log target and message prefix is assinged its own log handler.
func (sl *simpleLog) handle(target int, msgPrefix string) *log.Logger {
	// build key for log handler map
	if _, outer := sl.logHandle[target]; !outer {
		// allocate resources for a new log handler target map
		sl.logHandle[target] = make(map[string]*log.Logger)
	}
	if _, inner := sl.logHandle[target][msgPrefix]; !inner {
		// create a new log handler
		switch target {
		case stdout:
			sl.logHandle[stdout][msgPrefix] = log.New(os.Stdout, msgPrefix, 0)
		case file:
			if sl.fileHandle != nil {
				sl.logHandle[file][msgPrefix] = log.New(sl.fileHandle, msgPrefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix)
				if !firstFileLogHandler {
					// the first file log event always adds an empty line to the log file
					sl.fileHandle.WriteString("\n")
					firstFileLogHandler = true
				}
			}
		}
	}
	return sl.logHandle[target][msgPrefix]
}

// parseValues parses the variadic function parameters and returns a message prefix and a log record.
// Thereby, the following scenarios will be considered:
//   - Only one parameter is specified. In this case, this parameter is used as log record and the prefix is set to an emtpy character.
//   - Two or more parameter are specified. In this case, the first parameter is the message prefix and the remaining parameters are joined into one log record.
func parseValues(values []any) (string, string) {
	var prefix, msg string
	if len(values) > 1 {
		valueList := make([]string, len(values)-1)
		prefix = fmt.Sprint(values[0]) + " "
		for i, v := range values[1:] {
			if s, ok := v.(string); ok {
				// the parameter is already a string; no conversion is required
				valueList[i] = s
			} else {
				// convert parameter into a string
				valueList[i] = fmt.Sprint(v)
			}
		}
		msg = strings.Join(valueList, " ")
	} else {
		msg = fmt.Sprint(values[0])
	}

	return prefix, msg
}

// StartService starts the log service.
// The bufferSize specifies the number of log messages which can be buffered before the log service blocks.
// The log service runs in a dedicated goroutine.
func StartService(bufferSize int) {
	if sLog.serviceState() == stopped {
		// setup log handle map
		sLog.logHandle = make(map[int]map[string]*log.Logger)

		// setup channels
		sLog.data = make(chan logMessage, bufferSize)
		sLog.config = make(chan configMessage)
		sLog.stopLogService = make(chan signal)

		// start the log service
		go service()

		// set service state
		sLog.state = running
	} else {
		panic(sl002e)
	}
}

// StopService stops the log service.
// Before the log service is stopped, all pending log messages are flushed and resources are released.
func StopService() {
	defer close(sLog.data)
	defer close(sLog.stopLogService)

	if sLog.serviceState() == running {
		// wait until all log messages have been handled by the service
		for len(sLog.data) > 0 {
			continue
		}
		// all log messages are logged - the services can be stopped gracefully
		sLog.stopLogService <- signal{}
		sLog.fileHandle.Close()
	} else {
		panic(sl003e)
	}
}

// InitLogFile initializes the log file.
func InitLogFile(logName string) {
	if sLog.serviceState() == running {
		sLog.config <- configMessage{openlog, logName}
	} else {
		panic(sl004e)
	}
}

// ChangeLogFile changes the log file name.
// As part of this task, the current log file is closed (not deleted) and a log file with the new name is created.
// The log service doesn't need to be stopped for this task.
func ChangeLogFile(newLogName string) {
	if sLog.serviceState() == running {
		// wait until all log messages have been written to the old log file
		for len(sLog.data) > 0 {
			continue
		}
		sLog.config <- configMessage{changelogname, newLogName}
	} else {
		panic(sl004e)
	}
}

// WriteToStdout writes a log message to stdout.
func WriteToStdout(values ...any) {
	if sLog.serviceState() == running {
		prefix, logRecord := parseValues(values)
		sLog.data <- logMessage{stdout, prefix, logRecord}
	} else {
		panic(sl004e)
	}
}

// WriteToFile writes a log message to a log file.
func WriteToFile(values ...any) {
	if sLog.serviceState() == running {
		prefix, logRecord := parseValues(values)
		sLog.data <- logMessage{file, prefix, logRecord}
	} else {
		panic(sl004e)
	}
}

// WriteToMulti writes a log message to multiple targets.
// Currently supported targets are stdout and a log file.
func WriteToMulti(values ...any) {
	if sLog.serviceState() == running {
		prefix, logRecord := parseValues(values)
		sLog.data <- logMessage{multi, prefix, logRecord}
	} else {
		panic(sl004e)
	}
}
