package simplelog

import (
	"bufio"
	"fmt"
	"os"
	"time"
)

var (
	s = new(simpleLogService) // create instance of a simplelog service
)

// simpleLogService represents an object used to handle workflows triggered by the simplelog exported functions.
type simpleLogService struct {
	active                bool               // flag to indicate whether the log service is up and running
	stdoutLogger                             // the stdout logger instance
	fileLogger                               // the file logger instance
	dataQueue             chan logMessage    // to receive log data from the caller; this channel is buffered
	configService         chan configMessage // to receive config service requests from the caller
	configServiceResponse chan error         // to send an error response to the caller to continue the workflow
	stopService           chan bool          // to receive a stop service request from the caller
	stopServiceResponse   chan struct{}      // to send a signal to the caller to continue the workflow
}

// isActive returns true, if the log service is up and running, false otherwise.
func (s *simpleLogService) isActive() bool {
	return s.active
}

// setActive sets the active flag of the log service.
func (s *simpleLogService) setActive(state bool) {
	s.active = state
}

// instance denotes the logWriter interface implementation by the stdoutLogger type.
func (sl *stdoutLogger) instance() *logger {
	if sl.self == nil {
		sl.self = newLogger(os.Stdout)
	}
	return sl.self
}

// instance denotes the logWriter interface implementation by the fileLogger type.
func (f *fileLogger) instance() *logger {
	if f.self == nil {
		if f.desc == nil {
			panic(sg004)
		}
		f.writer = bufio.NewWriter(f.desc)
		// f.writer = bufio.NewWriterSize(f.desc, 10000000)
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
func (f *fileLogger) setupLogFile(flag int, logName string) error {
	var err error
	f.desc, err = os.OpenFile(logName, flag, 0644)
	return err
}

// releaseFileLogger releases all fileLogger resources.
func (f *fileLogger) releaseFileLogger(archive bool) error {
	var err error
	if f.self != nil {
		if f.writer.Buffered() >= 0 {
			// only do the flush when the buffer has data to be written
			f.writer.Flush()
		}
	}
	if err = f.desc.Close(); err != nil {
		return err
	}
	if archive {
		if err = s.archiveLogFile(s.desc.Name()); err != nil {
			return err
		}
	}
	f.writer = nil
	f.desc = nil
	f.self = nil
	return err
}

// archiveLogFile archives the log file.
func (f *fileLogger) archiveLogFile(logFileName string) error {
	var err error
	t := time.Now()
	formatted := fmt.Sprintf("%d%02d%02d%02d%02d%02d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
	logArchiveName := logFileName + "_" + formatted
	err = os.Rename(logFileName, logArchiveName)
	return err
}

// changeLogFile changes the name of the log file.
func (f *fileLogger) changeLogFile(flag int, newLogName string) error {
	var err error
	// release old fileLogger resources
	if err = f.releaseFileLogger(false); err != nil {
		return err
	}
	err = f.setupLogFile(flag, newLogName)
	return err
}

// stop stops the log service.
// A part of this step the underlying goroutine is also stopped.
func (s *simpleLogService) stop(archivelog bool) {
	s.stopService <- archivelog
	<-s.stopServiceResponse
}

// run represents the log service.
// This function is kicked off in a dedicated goroutine.
// It handles client requests by listening on the following channels:
//   - stopService
//   - dataQueue
//   - configService
func (s *simpleLogService) run(serviceRunning chan<- bool) {
	var logData logMessage
	var cfgData configMessage

	defer close(s.stopServiceResponse)

	// ticker to periodically trigger a flush of the log file buffer
	flushBufferInterval := time.NewTicker(1000 * time.Millisecond)

	// service loop
	for {
		select {
		case serviceRunning <- true:
		case archivelog := <-s.stopService:
			flush()
			s.releaseFileLogger(archivelog)
			return
		case logData = <-s.dataQueue:
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
				flag := cfgData.data[logflag].(int)
				logName := cfgData.data[logfilename].(string)
				err := s.setupLogFile(flag, logName)
				s.configServiceResponse <- err
			case switchlog:
				flush()
				flag := cfgData.data[logflag].(int)
				newLogName := cfgData.data[logfilename].(string)
				err := s.changeLogFile(flag, newLogName)
				s.configServiceResponse <- err
			case setprefix:
				if logPrefix, ok := cfgData.data[stdoutlogprefix]; ok {
					s.stdoutLogger.prefix = logPrefix.([]string)
				} else if logPrefix, ok = cfgData.data[filelogprefix]; ok {
					s.fileLogger.prefix = logPrefix.([]string)
				} else {
					panic(sg003)
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

// flush flushes(writes) messages, which are still buffered in the data channel
// and not yet wrtitten do disc.
func flush() {
	var m logMessage
	for len(s.dataQueue) > 0 {
		m = <-s.dataQueue
		writeMessage(&m)
	}
}
