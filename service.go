package simplelog

import (
	"log"
	"os"
	"time"
)

// log targets
const (
	stdout = 1 << iota     // write the log record to stdout
	file                   // write the log record to the log file
	multi  = stdout | file // write the log record to stdout and to the log file
)

// configuration categories
const (
	initlog   = iota // initializes a log file
	changelog        // change the log file name
)

// signal to confirm actions across channels
type signal struct{}

// a logMessage represents the log message which will be sent to the log service.
type logMessage struct {
	target int    // the log target bits, e.g. stdout, file, and so on.
	data   string // the payload of the log message, which will be sent to the log target
}

// a configMessage represents the config message which will be sent to the log service.
type configMessage struct {
	category int    // the configuration category bits, which are used to trigger certain config tasks by the log service, e.g. setlogname, changelogname, and so on.
	data     string // the data, which will be processed by a config task
}

// service is the instance to control and handle the way of log workflows.
type service struct {
	logFactory

	config    chan configMessage // the channel for sending config messages to the log service
	confirmed chan signal        // the channel for sending a confirmation signal to the caller
	stop      chan signal        // the channel for sending a stop signal to the log service
}

// logFactory is the base data collection to support logging to multiple targets.
type logFactory struct {
	data     chan logMessage // the channel for sending log messages to the log service; this channel will be a buffered channel
	multiLog                 // the multiLog supports logging to stdout and file
}

// stdoutLogWriter is a data collection to support logging to stdout.
type stdoutLog struct {
	stdoutLogInstance *log.Logger
}

// fileLogWriter is a data collection to support logging to files.
type fileLog struct {
	fileDesc        *os.File
	fileLogInstance *log.Logger
}

// logWriter is the log writer which supports logging to stdout and to files.
type multiLog struct {
	stdoutLog
	fileLog
}

// logWriter interface includes definitions of the following method signatures:
//   - instance
type logWriter interface {
	instance() *log.Logger // create and return a log.logger instance
}

// instance denotes the logWriter interface implementation by the stdoutLog type.
func (slw *stdoutLog) instance() *log.Logger {
	if slw.stdoutLogInstance == nil {
		slw.stdoutLogInstance = log.New(os.Stdout, "", 0)
	}
	return slw.stdoutLogInstance
}

// instance denotes the logWriter interface implementation by the fileLog type.
func (flw *fileLog) instance() *log.Logger {
	if flw.fileLogInstance == nil {
		flw.fileLogInstance = log.New(flw.fileDesc, "", log.Ldate|log.Ltime|log.Lmicroseconds)
		flw.fileDesc.WriteString("\n")
	}
	return flw.fileLogInstance
}

// service instance
var s = new(service)

// getLogWriter returns a log.Logger instance.
func (s *multiLog) getLogWriter(lw logWriter) *log.Logger {
	return lw.instance()
}

// setupLogFile creates and opens the log file.
func (s *multiLog) setupLogFile(logName string) {
	var err error
	s.fileDesc, err = os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
}

// changeLogFileName changes the name of the log file.
func (s *multiLog) changeLogFileName(newLogName string) {
	s.fileDesc.Close()
	s.setupLogFile(newLogName)
}

// run represents the log service.
// This service function runs in a dedicated goroutine and will be started as part of the log service startup process.
// It handles client requests by listening on the following channels:
//   - time.Time
//   - stop
//   - data
//   - config
func (s *service) run() {
	var logMsg logMessage
	var cfgMsg configMessage

	// initial heartbeat to the watchdog
	t := time.Now()
	w.getHeartBeatMonitor() <- t
	heartBeat := time.NewTicker(heartBeatInterval)

	// service loop
	for {
		select {
		case t = <-heartBeat.C:
			w.getHeartBeatMonitor() <- t
		case <-s.stop:
			// write all messages which are still in the data channel and have not been written yet
			s.flushMessages(len(s.data))
			heartBeat.Stop()
			// set the last heartbeat back by one hour so the watchdog assumes the service is no longer running
			t := time.Now()
			t = t.Add((-1) * time.Hour)
			w.getHeartBeatMonitor() <- t
			s.confirmed <- signal{}
			return
		case logMsg = <-s.data:
			s.writeMessage(logMsg)
		case cfgMsg = <-s.config:
			switch cfgMsg.category {
			case initlog:
				s.setupLogFile(cfgMsg.data)
				s.confirmed <- signal{}
			case changelog:
				// write all messages to the old log file, which were already sent to the data channel before the change log name was triggered
				s.flushMessages(len(s.data))
				// change the log file name
				s.changeLogFileName(cfgMsg.data)
				s.confirmed <- signal{}
			}
		}
	}
}

// writeMessage writes data of log messages to a dedicated target.
func (s *service) writeMessage(logMsg logMessage) {
	switch logMsg.target {
	case stdout:
		stdoutLogger := s.getLogWriter(&s.stdoutLog)
		stdoutLogger.Print(logMsg.data)
	case file:
		fileLogger := s.getLogWriter(&s.fileLog)
		fileLogger.Print(logMsg.data)
	case multi:
		stdoutLogger := s.getLogWriter(&s.stdoutLog)
		fileLogger := s.getLogWriter(&s.fileLog)
		stdoutLogger.Print(logMsg.data)
		fileLogger.Print(logMsg.data)
	}
}

// flushMessages flushes(writes) a number of messages to a dedicated target.
// The messages will be read from a buffered channel.
// Buffered channels in Go are always FIFO, so messages are flushed in FIFO approach.
func (s *service) flushMessages(numMessages int) {
	for numMessages > 0 {
		s.writeMessage(<-s.data)
		numMessages--
	}
}
