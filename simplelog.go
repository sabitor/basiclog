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
	m001 = "log file not initialized"
	m002 = "log service was already started"
	m003 = "log service is not running"
	m004 = "log service has not been started"
)

// log targets
const (
	stdout = 1 << iota     // write the log record to STDOUT
	file                   // write the log record to the log file
	multi  = stdout | file // write the log record to STDOUT and to the log file
)

// configuration categories
const (
	initlog   = iota // initializes a log file
	changelog        // change the log file name
)

// log service states
const (
	stopped = iota // indicator of a stopped log service
	running        // indicator of a running (active) log service
)

// semaphore to confirm actions across channels
type semaphore struct{}

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

// simpleLog is a representation of a simple logger instance.
type simpleLog struct {
	fileHandle *os.File            // the file handle of the log file
	logHandle  map[int]*log.Logger // a map which stores for every log target its assigned log handle

	data   chan logMessage    // the channel for sending log messages to the log service; this channel will be a buffered channel
	config chan configMessage // the channel for sending config messages to the log service
	stop   chan semaphore     // the channel for sending a stop signal to the log service
	done   chan semaphore     // the channel for sending a done signal to the caller

	state int // represents the state of the log service
	mtx   sync.Mutex
}

// global (package) variables
var (
	sLog = &simpleLog{}
)

// serviceIsActive checks whether the log service is active.
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

// initLogFile creates and opens the log file.
func (sl *simpleLog) initLogFile(logName string) {
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
	sLog.initLogFile(newLogName)
}

// service represents the log service.
// This service function runs in a dedicated goroutine and will be started as part of the log service startup process.
// It listenes on the following channels:
//   - data
//   - config
//   - stop
func (sl *simpleLog) service() {
	var logMsg logMessage
	var cfgMsg configMessage

	for {
		select {
		case <-sLog.stop:
			// Write all messages which are still in the data channel and have not been written yet.
			flushDataChan(len(sLog.data))
			sLog.done <- semaphore{}
			return
		case logMsg = <-sLog.data:
			writeMessage(logMsg)
		case cfgMsg = <-sLog.config:
			switch cfgMsg.category {
			case initlog:
				sLog.initLogFile(cfgMsg.data)
				sLog.done <- semaphore{}
			case changelog:
				// Write all messages to the old log file, which were already sent to the data channel before the change log name was triggered.
				flushDataChan(len(sLog.data))
				// Change the log file name.
				sLog.changeLogFileName(cfgMsg.data)
				sLog.done <- semaphore{}
			}
		}
	}
}

// writeMessage writes data of log messages to a dedicated target.
func writeMessage(logMsg logMessage) {
	var stdoutLogHandle, fileLogHandle *log.Logger

	switch logMsg.target {
	case stdout:
		stdoutLogHandle, _ = sLog.handle(stdout)
		stdoutLogHandle.Print(logMsg.data)
	case file:
		fileLogHandle, _ = sLog.handle(file)
		if fileLogHandle != nil {
			fileLogHandle.Print(logMsg.data)
		} else {
			panic(m001)
		}
	case multi:
		stdoutLogHandle, fileLogHandle = sLog.handle(multi)
		stdoutLogHandle.Print(logMsg.data)
		if fileLogHandle != nil {
			fileLogHandle.Print(logMsg.data)
		} else {
			panic(m001)
		}
	}
}

// flushDataChan writes(flushes) a given number of messages to a dedicated target.
// This is done in a FIFO manner (buffered channels in Go are always FIFO)
func flushDataChan(numMessages int) {
	for numMessages > 0 {
		fmt.Println("Flush")
		writeMessage(<-sLog.data)
		numMessages--
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
				// the first file log event always adds an empty line to the log file
				sl.fileHandle.WriteString("\n")
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
	if sLog.serviceState() == stopped {
		// setup log handle map
		sLog.logHandle = make(map[int]*log.Logger)

		// setup channels
		sLog.data = make(chan logMessage, bufferSize)
		sLog.config = make(chan configMessage)
		sLog.stop = make(chan semaphore)
		sLog.done = make(chan semaphore)

		// start the log service
		go sLog.service()
		sLog.state = running
	} else {
		panic(m002)
	}
}

// StopService stops the log service.
// Before the log service is stopped, all pending log messages are flushed and resources are released.
func StopService() {
	sLog.mtx.Lock()
	defer sLog.mtx.Unlock()
	if sLog.serviceState() == running {
		// stop the log service
		sLog.state = stopped
		sLog.stop <- semaphore{}
		<-sLog.done

		// cleanup
		sLog.fileHandle.Close()
		close(sLog.data)
		close(sLog.config)
		close(sLog.stop)
		close(sLog.done)
		sLog.logHandle = nil
		sLog.fileHandle = nil
	} else {
		panic(m003)
	}
}

// InitLogFile initializes the log file.
func InitLogFile(logName string) {
	sLog.mtx.Lock()
	defer sLog.mtx.Unlock()
	if sLog.serviceState() == running {
		// initialize the log file
		sLog.config <- configMessage{initlog, logName}
		<-sLog.done
	} else {
		panic(m004)
	}
}

// ChangeLogName changes the log file name.
// As part of this task, the current log file is closed (not deleted) and a log file with the new name is created.
// The log service doesn't need to be stopped for this task.
func ChangeLogName(newLogName string) {
	sLog.mtx.Lock()
	defer sLog.mtx.Unlock()
	if sLog.serviceState() == running {
		// change the log name
		sLog.config <- configMessage{changelog, newLogName}
		<-sLog.done
	} else {
		panic(m004)
	}
}

// WriteToStdout writes a log message to stdout.
func WriteToStdout(values ...any) {
	sLog.mtx.Lock()
	defer sLog.mtx.Unlock()
	if sLog.serviceState() == running {
		msg := parseValues(values)
		sLog.data <- logMessage{stdout, msg}
	} else {
		panic(m004)
	}
}

// WriteToFile writes a log message to a log file.
func WriteToFile(values ...any) {
	sLog.mtx.Lock()
	defer sLog.mtx.Unlock()
	if sLog.serviceState() == running {
		msg := parseValues(values)
		sLog.data <- logMessage{file, msg}
	} else {
		panic(m004)
	}
}

// WriteToMulti writes a log message to multiple targets.
// Currently supported targets are stdout and file.
func WriteToMulti(values ...any) {
	sLog.mtx.Lock()
	defer sLog.mtx.Unlock()
	if sLog.serviceState() == running {
		msg := parseValues(values)
		sLog.data <- logMessage{multi, msg}
	} else {
		panic(m004)
	}
}
