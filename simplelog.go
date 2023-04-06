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
	stdout = iota
	file
	multi
)

// configuration categories
const (
	setlogname = iota
	changelogname
)

// service states
const (
	stopped = iota
	running
)

// TODO: add comment
type signal struct{}

// TODO: add comment
type logMessage struct {
	target int
	prefix string
	record string
}

// TODO: add comment
type cfgMessage struct {
	category int
	data     string
}

// TODO: add comment
type simpleLog struct {
	// handler
	fileHandle *os.File
	logHandle  map[int]map[string]*log.Logger

	// channels
	data           chan logMessage
	config         chan cfgMessage
	stopLogService chan signal

	// service
	state int
}

// TODO: add comment
var sLog = &simpleLog{}
var firstFileLogHandler = false

// TODO: add comment
func (sl *simpleLog) setLogFile(logName string) {
	var err error
	sLog.fileHandle, err = os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
}

// TODO: add comment
func (sl *simpleLog) setServiceState(newState int) {
	sl.state = newState
}

// TODO: add comment
func (sl *simpleLog) serviceState() int {
	return sl.state
}

// TODO: add comment
func (sl *simpleLog) stdoutLog(prefix string) *log.Logger {
	return sLog.handler(stdout, prefix)
}

// TODO: add comment
func (sl *simpleLog) fileLog(prefix string) *log.Logger {
	return sLog.handler(file, prefix)
}

// TODO: add comment
func (sl *simpleLog) multiLog(prefix string) (*log.Logger, *log.Logger) {
	return sLog.handler(stdout, prefix), sLog.handler(file, prefix)
}

// TODO: add comment
func (sl *simpleLog) initialize(buffer int) {
	// setup log handler
	// The log handler map stores log handler with different properties - target and message prefixes.
	sl.logHandle = make(map[int]map[string]*log.Logger)

	// setup channels
	sl.data = make(chan logMessage, buffer)
	sl.config = make(chan cfgMessage)
	sl.stopLogService = make(chan signal)

	// setup service state
	sl.state = running
}

// TODO: add comment
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

// TODO: add comment
func (sl *simpleLog) changeLogFile(newLogName string) {
	// remove all file log handles from the logHandler map which are linked to the old log name
	delete(sLog.logHandle, file)
	sLog.fileHandle.Close()
	firstFileLogHandler = false
	sLog.setLogFile(newLogName)
}

// TODO: add comment
func (sl *simpleLog) handler(target int, msgPrefix string) *log.Logger {
	// build key for log handler map
	if _, outer := sl.logHandle[target]; !outer {
		// allocate memory for a new log handler target map
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
					// the first file log event always adds an empty line to the log file at the beginning
					sl.fileHandle.WriteString(lineBreak)
					firstFileLogHandler = true
				}
			}
		}
	}
	return sl.logHandle[target][msgPrefix]
}

// TODO: add comment
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

// TODO: add comment
func StartService(bufferSize int) {
	if sLog.serviceState() == stopped {
		sLog.initialize(bufferSize)
		go service()
	} else {
		panic(sl002e)
	}
}

// TODO: add comment
func StopService() {
	defer close(sLog.data)
	defer close(sLog.stopLogService)

	if sLog.serviceState() == running {
		// wait until all messages have been logged by the service
		for len(sLog.data) > 0 {
			continue
		}
		// all messages are logged - the services can be stopped gracefully
		sLog.stopLogService <- signal{}
		sLog.fileHandle.Close()
	} else {
		panic(sl003e)
	}
}

// TODO: add comment
func SetLogName(logName string) {
	if sLog.serviceState() == running {
		time.Sleep(10 * time.Millisecond) // to keep the logical order of goroutine function calls
		if sLog.fileHandle == nil {
			sLog.config <- cfgMessage{setlogname, logName}
		} else {
			panic(sl005e)
		}
	} else {
		panic(sl004e)
	}
}

// TODO: add comment
func ChangeLogName(newLogName string) {
	if sLog.serviceState() == running {
		time.Sleep(10 * time.Millisecond) // to keep the logical order of goroutine function calls
		sLog.config <- cfgMessage{changelogname, newLogName}
	} else {
		panic(sl004e)
	}
}

// TODO: add comment
func WriteToStdout(prefix string, values ...any) {
	if sLog.serviceState() == running {
		logRecord := assembleToString(values)
		sLog.data <- logMessage{stdout, prefix, logRecord}
	} else {
		panic(sl004e)
	}
}

// TODO: add comment
func WriteToFile(prefix string, values ...any) {
	if sLog.serviceState() == running {
		logRecord := assembleToString(values)
		sLog.data <- logMessage{file, prefix, logRecord}
	} else {
		panic(sl004e)
	}
}

// TODO: add comment
func WriteToMultiple(prefix string, values ...any) {
	if sLog.serviceState() == running {
		logRecord := assembleToString(values)
		sLog.data <- logMessage{multi, prefix, logRecord}
	} else {
		panic(sl004e)
	}
}
