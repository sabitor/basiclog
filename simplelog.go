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
	"sync"
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
	data   string // the payload of the log message, which will be sent to the log target
}

// A configMessage represents the config message which will be sent to the log service.
type configMessage struct {
	category int    // the configuration category bits, which are used to trigger certain config tasks by the log service, e.g. setlogname, changelogname, and so on.
	data     string // the data, which will be processed by a config task
}

// A simpleLog represents an instance of a simple logger.
type simpleLog struct {
	// handler
	fileHandle *os.File            // the file handle of the log file
	logHandle  map[int]*log.Logger // a map which stores for every log target bit its assigned log handle

	// channels
	data           chan logMessage    // the channel for sending log messages to the log service; this channel will be a buffered channel
	config         chan configMessage // the channel for sending config messages to the log service
	stopLogService chan signal        // the channel for sending a stop message to the log service

	// service
	state int // to save the current state of the log service repesented by the service bits, e.g. stopped, running, and so on
}

// global (package) variables
var (
	firstFileLogHandler = false
	mtx                 sync.Mutex
	sLog                = &simpleLog{}
)

// serviceState returns the state of the log service.
// The returned state bits are stopped, running, and so on.
func (sl *simpleLog) serviceState() int {
	return sl.state
}

// handle returns log handler for a given log target.
func (sl *simpleLog) handle(target int) (*log.Logger, *log.Logger) {
	var logHandle1, logHandle2 *log.Logger
	switch target {
	case stdout:
		logHandle1 = sLog.checkBuildHandle(stdout)
	case file:
		logHandle1 = sLog.checkBuildHandle(file)
	case multi:
		// stdout and file log handler have different properties, thus io.MultiWriter can't be used
		logHandle1 = sLog.checkBuildHandle(stdout)
		logHandle2 = sLog.checkBuildHandle(file)
	}
	return logHandle1, logHandle2
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

// service represents the log service.
// This service function runs in a dedicated goroutine and will be started as part of the log service startup process.
// It handles the following messages:
//   - logMessage
//   - configMessage
//   - signal
func service() {
	var stdoutLogHandle, fileLogHandle *log.Logger
	var logMsg logMessage
	var cfgMsg configMessage

	for {
		select {
		case <-sLog.stopLogService:
			return
		case logMsg = <-sLog.data:
			switch logMsg.target {
			case stdout:
				stdoutLogHandle, _ = sLog.handle(stdout)
				stdoutLogHandle.Print(logMsg.data)
			case file:
				fileLogHandle, _ = sLog.handle(file)
				if fileLogHandle != nil {
					fileLogHandle.Print(logMsg.data)
				} else {
					panic(sl001e)
				}
			case multi:
				stdoutLogHandle, fileLogHandle = sLog.handle(multi)
				stdoutLogHandle.Print(logMsg.data)
				if fileLogHandle != nil {
					fileLogHandle.Print(logMsg.data)
				} else {
					panic(sl001e)
				}
			}
		case cfgMsg = <-sLog.config:
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

// checkBuildHandle checks if a log handle exists for a specific target. If not, it will be created accordingly.
// Each log target is assinged its own log handler.
func (sl *simpleLog) checkBuildHandle(target int) *log.Logger {
	if _, found := sl.logHandle[target]; !found {
		// log handler doesn't exists - create it
		switch target {
		case stdout:
			sl.logHandle[stdout] = log.New(os.Stdout, "", 0)
		case file:
			if sl.fileHandle != nil {
				sl.logHandle[file] = log.New(sl.fileHandle, "", log.Ldate|log.Ltime|log.Lmicroseconds)
				if !firstFileLogHandler {
					// the first file log event always adds an empty line to the log file
					sl.fileHandle.WriteString("\n")
					firstFileLogHandler = true
				}
			}
		}
	}
	return sLog.logHandle[target]
}

// parseValues parses the variadic function parameters, builds a message from them and returns it.
func parseValues(values []any) string {
	valueList := make([]string, len(values))
	for i, v := range values {
		if s, ok := v.(string); ok {
			// the parameter is already a string; no conversion is required
			valueList[i] = s
		} else {
			// convert parameter into a string
			valueList[i] = fmt.Sprint(v)
		}
	}
	return strings.Join(valueList, " ")
}

// StartService starts the log service.
// The bufferSize specifies the number of log messages which can be buffered before the log service blocks.
// The log service runs in its own goroutine.
func StartService(bufferSize int) {
	mtx.Lock()
	defer mtx.Unlock()
	if sLog.serviceState() == stopped {
		// setup log handle map
		sLog.logHandle = make(map[int]*log.Logger)

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
	mtx.Lock()
	defer mtx.Unlock()
	if sLog.serviceState() == running {
		// wait until all log messages have been written
		for len(sLog.data) > 0 {
			continue
		}
		// set service state
		sLog.state = stopped

		// no pending log messages - the services can be stopped gracefully
		sLog.stopLogService <- signal{}

		// cleanup
		sLog.fileHandle.Close()
		close(sLog.data)
		close(sLog.config)
		close(sLog.stopLogService)
		sLog.logHandle = nil
		sLog.fileHandle = nil
	} else {
		panic(sl003e)
	}
}

// InitLogFile initializes the log file.
func InitLogFile(logName string) {
	mtx.Lock()
	defer mtx.Unlock()
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
	mtx.Lock()
	defer mtx.Unlock()
	if sLog.serviceState() == running {
		// wait until all log messages have been written to the old log file
		for len(sLog.data) > 0 {
			continue
		}
		// no pending log messages - log file name can be changed
		sLog.config <- configMessage{changelogname, newLogName}
	} else {
		panic(sl004e)
	}
}

// WriteToStdout writes a log message to stdout.
func WriteToStdout(values ...any) {
	if sLog.serviceState() == running {
		msg := parseValues(values)
		sLog.data <- logMessage{stdout, msg}
	} else {
		panic(sl004e)
	}
}

// WriteToFile writes a log message to a log file.
func WriteToFile(values ...any) {
	if sLog.serviceState() == running {
		msg := parseValues(values)
		sLog.data <- logMessage{file, msg}
	} else {
		panic(sl004e)
	}
}

// WriteToMulti writes a log message to multiple targets.
// Currently supported targets are stdout and file.
func WriteToMulti(values ...any) {
	if sLog.serviceState() == running {
		msg := parseValues(values)
		sLog.data <- logMessage{multi, msg}
	} else {
		panic(sl004e)
	}
}
