package simplelog

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"time"
)

// simplelog service instance
var s = new(simpleLogService)

// simpleLogService represents an object used to handle workflows triggered by the simplelog exported functions.
type simpleLogService struct {
	logData               chan logMessage    // receiver of log data from the caller; this channel buffered
	configService         chan configMessage // receiver of config service requests from the caller
	configServiceResponse chan error         // sender of an error response to the caller to continue the workflow
	stdoutLogger                             // the stdout logger instance
	fileLogger                               // the file logger instance
	isUp                  bool
}

// instance denotes the logWriter interface implementation by the stdoutLog type.
func (sl *stdoutLogger) instance() *logger {
	if sl.self == nil {
		sl.self = newLogger(os.Stdout)
	}
	return sl.self
}

// instance denotes the logWriter interface implementation by the fileLog type.
func (f *fileLogger) instance() *logger {
	if f.self == nil {
		if f.desc == nil {
			panic(m001)
		}
		// f.fileWriter = bufio.NewWriter(s.fileDesc)
		f.writer = bufio.NewWriterSize(f.desc, 16384)
		f.self = newLogger(f.writer)
		f.desc.WriteString("\n")
	}
	return f.self
}

// simpleLogger returns a logger instance.
func simpleLogger(lw logWriter) *logger {
	return lw.instance()
}

// setupLogFile creates and opens the log file.
func (f *fileLogger) setupLogFile(logName string) error {
	var err error
	f.desc, err = os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	return err
}

// releaseFileLogger releases all resources allocated by the fileLogger structure.
func (f *fileLogger) releaseFileLogger(archive bool) error {
	var err error
	if f.desc != nil {
		if f.writer != nil {
			if f.writer.Buffered() >= 0 {
				// only do the flush when the buffer has data to be written
				f.writer.Flush()
			}
		}
		if err = f.desc.Close(); err != nil {
			return err
		}
		if archive {
			s.archiveLogFile(s.desc.Name())
		}
		f.writer = nil
		f.desc = nil
		f.self = nil
	} else {
		err = errors.New(m001)
	}
	return err
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
func (f *fileLogger) changeLogFile(newLogName string) error {
	var err error
	// release old fileLogger resources
	if err = f.releaseFileLogger(false); err != nil {
		return err
	}
	err = f.setupLogFile(newLogName)
	return err
}

// run represents the log service.
// This service function runs in a dedicated goroutine and is started as part of the log service startup process.
// It handles client requests by listening on the following channels:
//   - stop
//   - data
//   - config
func (s *simpleLogService) run() {
	var logData logMessage
	var cfgData configMessage

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
		case logData = <-s.logData:
			writeMessage(&logData)
		case <-flushBufferInterval.C:
			if s.writer != nil {
				// only do the flush when the buffer has data to be written
				if s.writer.Buffered() > 0 {
					s.writer.Flush()
				}
			}
		case cfgData = <-s.configService:
			switch cfgData.task {
			case initlog:
				logName := cfgData.data[logfilename]
				err := s.setupLogFile(logName)
				s.configServiceResponse <- err
			case switchlog:
				flush()
				newLogName := cfgData.data[logfilename]
				err := s.changeLogFile(newLogName)
				s.configServiceResponse <- err
			case setprefix:
				if logPrefix, ok := cfgData.data[stdoutlogprefix]; ok {
					s.stdoutLogger.prefix = logPrefix
				} else if logPrefix, ok = cfgData.data[filelogprefix]; ok {
					s.fileLogger.prefix = logPrefix
				} else {
					panic(m007)
				}
				s.configServiceResponse <- nil
			}
		}
	}
}

// writeMessage writes data of log messages to a dedicated destination.
func writeMessage(logMsg *logMessage) {
	switch logMsg.destination {
	case STDOUT:
		simpleLogger(&s.stdoutLogger).write(logMsg)
	case FILE:
		simpleLogger(&s.fileLogger).write(logMsg)
	case MULTI:
		logMsg.destination = MULTI & STDOUT
		simpleLogger(&s.stdoutLogger).write(logMsg)
		logMsg.destination = MULTI & FILE
		simpleLogger(&s.fileLogger).write(logMsg)
	}

}

// flush flushes(writes) messages, which are still buffered in the data channel.
// Buffered channels in Go are always FIFO, so messages are flushed in FIFO approach.
func flush() {
	var m logMessage
	for len(s.logData) > 0 {
		m = <-s.logData
		writeMessage(&m)
	}
}
