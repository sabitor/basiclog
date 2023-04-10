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
	"time"
)

const lineBreak = "\n"

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
	setlogname    = iota // triggers to open and set the log file
	changelogname        // triggers to change the log file name
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

// A cfgMessage represents the config message which will be sent to the log service.
type cfgMessage struct {
	category int    // the configuration category bits, which are used to trigger certain config tasks by the log service, e.g. setlogname, changelogname, and so on.
	data     string // the data, which will be processed by a config task
}

// A simpleLog represents an instance of a simple logger.
type simpleLog struct {
	// handler
	fileHandle *os.File                       // the log file handle
	logHandle  map[int]map[string]*log.Logger // a map which stores for every log target bit a map which stores the log prefix and its assigned log handle

	// channels
	data           chan logMessage // the channel for sending log messages to the log service
	config         chan cfgMessage // the channel for sending config messages to the log service
	stopLogService chan signal     // the channel for sending a stop message to the log service

	// service
	state int // to save the current state of the log service repesented by the service bits, e.g. stopped, running, and so on
}

//  global (package) variables
var sLog = &simpleLog{}
var firstFileLogHandler = false

// setLogFile opens a log file.
func (sl *simpleLog) setLogFile(logName string) {
	var err error
	sLog.fileHandle, err = os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
}

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

// multiLog returns a multi log handle.
func (sl *simpleLog) multiLog(prefix string) (*log.Logger, *log.Logger) {
	return sLog.handle(stdout, prefix), sLog.handle(file, prefix)
}

// initialize is invoked during the startup process of the log service.
// It allocates resources for different simpleLog attributes and
// sets the state of the log service to running.
func (sl *simpleLog) initialize(buffer int) {
	// setup log handle map
	sl.logHandle = make(map[int]map[string]*log.Logger)

	// setup channels
	sl.data = make(chan logMessage, buffer)
	sl.config = make(chan cfgMessage)
	sl.stopLogService = make(chan signal)

	// setup service state
	sl.state = running
}

// service is the main component of the log service.
// It listens and handles messages sent to the log service.
// This service runs in a dedicated goroutine and will be started as part of the log service startup process.
// It handles the following messages:
//   * stopLogService - to signal the end of the log service and to stop further processing
//   * data           - log messages, used to write messages to the assigned log target
//   * config         - config messages, used to configure how the log service works
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
			case setlogname:
				sLog.setLogFile(cfgMsg.data)
			case changelogname:
				sLog.changeLogFile(cfgMsg.data)
			}
		}
	}
}

// changeLogFile changes the name of the log file.
func (sl *simpleLog) changeLogFile(newLogName string) {
	// remove all file log handles from the logHandler map which are linked to the old log name
	delete(sLog.logHandle, file)
	sLog.fileHandle.Close()
	firstFileLogHandler = false
	sLog.setLogFile(newLogName)
}

// handle returns a log handle for a given log target and message prefix.
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
			sl.logHandle[stdout][msgPrefix] = log.New(os.Stdout, "", 0)
		case file:
			if sl.fileHandle != nil {
				sl.logHandle[file][msgPrefix] = log.New(sl.fileHandle, msgPrefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix)
				if !firstFileLogHandler {
					// the first file log event always adds an empty line to the log file
					sl.fileHandle.WriteString(lineBreak)
					firstFileLogHandler = true
				}
			}
		}
	}
	return sl.logHandle[target][msgPrefix]
}

// assembleToString joins the variadic function parameters to one result message.
// The different parameters are separated by space characters.
func assembleToString(values []any) string {
	valueList := make([]string, len(values))
	for i, v := range values {
		if s, ok := v.(string); ok {
			valueList[i] = s
		} else {
			valueList[i] = fmt.Sprint(v)
		}
	}
	msg := strings.Join(valueList, " ")
	return msg
}

// StartService starts the log service.
// The log service runs in a dedicated goroutine.
func StartService(bufferSize int) {
	if sLog.serviceState() == stopped {
		sLog.initialize(bufferSize)
		go service()
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

// SetLogName sets the log file name.
func SetLogName(logName string) {
	if sLog.serviceState() == running {
		time.Sleep(10 * time.Millisecond) // CHECK: to keep the logical order of goroutine function calls
		if sLog.fileHandle == nil {
			sLog.config <- cfgMessage{setlogname, logName}
		} else {
			panic(sl005e)
		}
	} else {
		panic(sl004e)
	}
}

// ChangeLogName changes the log file name.
// As part of this task, the current log file is closed (not deleted) and a log file with the new name is created.
// The log service doesn't need to be stopped for this task.
func ChangeLogName(newLogName string) {
	if sLog.serviceState() == running {
		time.Sleep(10 * time.Millisecond) // CHECK: to keep the logical order of goroutine function calls
		sLog.config <- cfgMessage{changelogname, newLogName}
	} else {
		panic(sl004e)
	}
}

// WriteToStdout writes log messages to stdout.
func WriteToStdout(prefix string, values ...any) {
	if sLog.serviceState() == running {
		logRecord := assembleToString(values)
		sLog.data <- logMessage{stdout, prefix, logRecord}
	} else {
		panic(sl004e)
	}
}

// WriteToFile writes log messages to a log file.
func WriteToFile(prefix string, values ...any) {
	if sLog.serviceState() == running {
		logRecord := assembleToString(values)
		sLog.data <- logMessage{file, prefix, logRecord}
	} else {
		panic(sl004e)
	}
}

// WriteToMulti writes log messages to multiple targets.
// Currently supported targets are stdout and a log file.
func WriteToMulti(prefix string, values ...any) {
	if sLog.serviceState() == running {
		logRecord := assembleToString(values)
		sLog.data <- logMessage{multi, prefix, logRecord}
	} else {
		panic(sl004e)
	}
}
