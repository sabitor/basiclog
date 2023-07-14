package simplelog

import (
	"bufio"
	"fmt"
	// "log"
	"os"
	"time"
)

// simplelog service instance
var s = new(simpleLogService)

// log targets
const (
	STDOUT = iota // write the log record to stdout
	FILE          // write the log record to the log file
	MULTI         // write the log record to stdout and to the log file
)

// log service actions
const (
	start = iota
	stop
	initlog
	switchlog
)

// log service states bitmask
const (
	stopped = 1 << iota // the service is stopped and cannot process log requests
	running             // the service is running
)

// log service attributes
const (
	logbuffer   = iota // defines the buffer size of the logMessage channel
	logfilename        // defines the log file name to be used
	logarchive         // a flag which defines whether the log should be archived
	appendlog          // a flag which defines whether the messages are appended to the existing log
)

// signal to confirm actions across channels
type signal struct{}

// a logMessage represents the log message which will be sent to the log service.
type logMessage struct {
	target int   // the log target bits, e.g. stdout, file, and so on.
	data   []any // the payload of the log message, which will be sent to the log target
}

// a configMessage represents the config message which will be sent to the log service.
type configMessage struct {
	action int    // the configuration action, which is used to trigger certain config tasks by the log service
	data   string // the data, which will be used by the config task
}

// simpleLogService is structure used to handle workflows triggered by the simplelog API.
type simpleLogService struct {
	attribute     map[int]any        // the map which contains the log factory attributes
	logData       chan logMessage    // the channel for sending log messages to the log service; this channel buffered
	serviceConfig chan configMessage // the channel for sending config messages to the log service
	stdoutLogger                     // the stdout logger
	fileLogger                       // the file logger
}

// stdoutLogger is a data collection to support logging to stdout.
type stdoutLogger struct {
	// stdoutLogInstance *log.Logger
	stdoutLogInstance *logger
}

// fileLogger is a data collection to support logging to files.
type fileLogger struct {
	fileWriter *bufio.Writer
	fileDesc   *os.File
	// fileLogInstance *log.Logger
	fileLogInstance *logger
}

// logWriter interface includes definitions of the following method signatures:
//   - instance
type logWriter interface {
	// instance() *log.Logger // create and return a *log.logger instance
	instance() *logger // create and return a *log.logger instance
}

// instance denotes the logWriter interface implementation by the stdoutLog type.
// func (s *stdoutLogger) instance() *log.Logger {
// 	if s.stdoutLogInstance == nil {
// 		s.stdoutLogInstance = log.New(os.Stdout, "", 0)
// 	}
// 	return s.stdoutLogInstance
// }

func (s *stdoutLogger) instance() *logger {
	if s.stdoutLogInstance == nil {
		s.stdoutLogInstance = new2(os.Stdout)
	}
	return s.stdoutLogInstance
}

// instance denotes the logWriter interface implementation by the fileLog type.
// func (f *fileLogger) instance() *log.Logger {
// 	if f.fileLogInstance == nil {
// 		if f.fileDesc == nil {
// 			panic(m001)
// 		}
// 		// f.fileWriter = bufio.NewWriter(s.fileDesc)
// 		f.fileWriter = bufio.NewWriterSize(f.fileDesc, 16384)
// 		f.fileLogInstance = log.New(f.fileWriter, "", log.Ldate|log.Ltime|log.Lmicroseconds)
// 		f.fileWriter.WriteString("\n")
// 	}
// 	return f.fileLogInstance
// }

func (f *fileLogger) instance() *logger {
	if f.fileLogInstance == nil {
		if f.fileDesc == nil {
			panic(m001)
		}
		// f.fileWriter = bufio.NewWriter(s.fileDesc)
		f.fileWriter = bufio.NewWriterSize(f.fileDesc, 16384)
		f.fileLogInstance = new2(f.fileWriter)
		f.fileDesc.WriteString("\n")
	}
	return f.fileLogInstance
}

// getLogWriter returns a log.Logger instance.
// func getLogWriter(lw logWriter) *log.Logger {
func getLogWriter(lw logWriter) *logger {
	return lw.instance()
}

// setupLogFile creates and opens the log file.
func (f *fileLogger) setupLogFile(logName string) {
	var err error
	f.fileDesc, err = os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
}

// releaseFileLogger releases all resources allocated by the fileLogger structure.
func (f *fileLogger) releaseFileLogger(archive bool) {
	if f.fileDesc != nil {
		if f.fileWriter != nil {
			if f.fileWriter.Buffered() >= 0 {
				// only do the flush when the buffer has data to be written
				f.fileWriter.Flush()
			}
		}
		if err := f.fileDesc.Close(); err != nil {
			panic(err)
		}
		if archive {
			s.archiveLogFile(s.fileDesc.Name())
		}
		s.fileWriter = nil
		f.fileDesc = nil
		f.fileLogInstance = nil
	}
}

// archiveLogFile archives the log file.
func (f *fileLogger) archiveLogFile(logFileName string) {
	t := time.Now()
	formatted := fmt.Sprintf("%d%02d%02d%02d%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	logArchiveName := logFileName + "_" + formatted
	if err := os.Rename(logFileName, logArchiveName); err != nil {
		panic(err)
	}
}

// changeLogFile changes the name of the log file.
func (f *fileLogger) changeLogFile(newLogName string) {
	// release old fileLogger resources
	f.releaseFileLogger(false)
	f.setupLogFile(newLogName)
}

// setAttribut sets a log service attribute.
func (s *simpleLogService) setAttribut(key int, value any) {
	if s.attribute == nil {
		s.attribute = make(map[int]any)
	}
	s.attribute[key] = value
}

// run represents the log service.
// This service function runs in a dedicated goroutine and is started as part of the log service startup process.
// It handles client requests by listening on the following channels:
//   - stop
//   - data
//   - config
func (s *simpleLogService) run() {
	var logMsg logMessage
	var cfgMsg configMessage

	c.setState(running)
	defer c.setState(stopped)

	// ticker to periodically trigger a flush of the file buffer
	flushBufferInterval := time.NewTicker(1000 * time.Millisecond)

	// service loop
	for {
		select {
		case <-c.stopService:
			flush()
			return
		case logMsg = <-s.logData:
			writeMessage(logMsg)
		case <-flushBufferInterval.C:
			// only do the flush when the buffer has data to be written
			if s.fileWriter.Buffered() > 0 {
				s.fileWriter.Flush()
			}
		case cfgMsg = <-s.serviceConfig:
			switch cfgMsg.action {
			case initlog:
				s.setupLogFile(cfgMsg.data)
				c.execServiceActionResponse <- signal{}
			case switchlog:
				flush()
				s.changeLogFile(cfgMsg.data)
				c.execServiceActionResponse <- signal{}
			}
		}
	}
}

// writeMessage writes data of log messages to a dedicated target.
func writeMessage(logMsg logMessage) {
	switch logMsg.target {
	case STDOUT:
		// getLogWriter(&s.stdoutLogger).Print(logMsg.data)
		getLogWriter(&s.stdoutLogger).write(logMsg.data)
	case FILE:
		// getLogWriter(&s.fileLogger).Print(logMsg.data)
		getLogWriter(&s.fileLogger).write(logMsg.data)
	case MULTI:
		// getLogWriter(&s.stdoutLogger).Print(logMsg.data)
		// getLogWriter(&s.fileLogger).Print(logMsg.data)
		getLogWriter(&s.stdoutLogger).write(logMsg.data)
		getLogWriter(&s.fileLogger).write(logMsg.data)
	}
}

// flush flushes(writes) messages, which are still buffered in the data channel.
// Buffered channels in Go are always FIFO, so messages are flushed in FIFO approach.
func flush() {
	for len(s.logData) > 0 {
		writeMessage(<-s.logData)
	}
}
