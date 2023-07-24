// Package simplelog is a logging package with the focus on simplicity,
// ease of use and performance.
// Once started, the simple logger runs as a service and listens for logging
// requests. Logging requests can be send to different logging destinations,
// such as standard out, a log file, or both.
// The simple logger can be used simultaneously from multiple goroutines.
package simplelog

import (
	"os"
)

// message catalog
const (
	m000 = "control is not running"
	m001 = "log file not initialized"
	m002 = "log service was already started"
	m003 = "log service is not running"
	m004 = "log service has not been started"
	m005 = "log file already initialized"
	m006 = "log file already exists"
	m007 = "unknown log destination specified"
	m008 = "only FILE or STDOUT are valid log destinations in this context"
)

// Startup starts the log service.
// The log service runs in its own goroutine.
// The bufferSize specifies the number of log messages which can be buffered before the log service blocks.
func Startup(bufferSize int) {
	if !c.checkState(running) {
		val := convertToString(bufferSize)
		c.startServiceTask <- configMessage{start, map[int]string{logbuffer: val}}
		<-c.startServiceTaskResponse
	} else {
		panic(m002)
	}
}

// Shutdown stops the log service and does some cleanup.
// Before the log service is stopped, all pending log messages are flushed and resources are released.
// The archive flag indicates whether the log file will be archived (true) or not (false).
// Archiving a log file means that it will be renamed and no new messages will be appended on a new run.
// The archived log file is of the following format: <orig file name>_yymmddHHMMSS.
func Shutdown(archive bool) {
	if c.checkState(running) {
		val := convertToString(archive)
		c.startServiceTask <- configMessage{stop, map[int]string{logarchive: val}}
		<-c.startServiceTaskResponse
	} else {
		panic(m003)
	}
}

// InitLog initializes the log file.
// The logName specifies the name of the log file.
// The append flag indicates whether messages are appended to the existing log file (true),
// or on a new run whether the old log is removed and a new log is created in its place (false).
func InitLog(logName string, append bool) {
	if c.checkState(running) {
		var err error
		if s.desc != nil {
			panic(m005)
		}
		if !append {
			// don't append - remove old log
			if _, err = os.Stat(logName); err == nil {
				if err = os.Remove(logName); err != nil {
					panic(err)
				}
			}
		}
		val := convertToString(append)
		s.configService <- configMessage{initlog, map[int]string{appendlog: val, logfilename: logName}}
		if err = <-s.configServiceResponse; err != nil {
			panic(err)
		}
	} else {
		panic(m004)
	}
}

// SwitchLog closes the current log file and a new log file with the specified name is created and used.
// Thereby, the current log file is not deleted, the new log file must not exist and the log service
// doesn't need to be stopped for this task. The new log file must not exist.
// The newLogName specifies the name of the new log to switch to.
func SwitchLog(newLogName string) {
	if c.checkState(running) {
		var err error
		if _, err = os.Stat(newLogName); err == nil {
			panic(m006)
		}
		s.configService <- configMessage{switchlog, map[int]string{logfilename: newLogName}}
		if err = <-s.configServiceResponse; err != nil {
			panic(err)
		}
	} else {
		panic(m004)
	}
}

// SetPrefix sets the prefix for logging lines.
// The destination specifies the name of the log destination where the prefix should be used, e.g. STDOUT or FILE.
// The prefix specifies the prefix for each logging line for a given log destination.
// If the prefix should also contain actual date and time data, the following placeholders can be applied accordingly:
//
//	year: yyyy
//	month: mm
//	day: dd
//	hour: HH
//	minute: MI
//	second: SS
//	millisecond: f[5f]
//
// In addition, to distinguish and parse date and time information, placeholders have to be delimited by
// <DT>...<DT> tags and can be used for example as follows: <DT>yyyy-mm-dd HH:MI:SS.ffffff<DT>.
// Furthermore, not all placeholders have to be used and they can be used in any order.
func SetPrefix(destination int, prefix string) {
	if c.checkState(running) {
		preparedPrefix := preprocessPrefix(prefix)
		switch destination {
		case STDOUT:
			s.configService <- configMessage{setprefix, map[int]string{stdoutlogprefix: preparedPrefix}}
			<-s.configServiceResponse
		case FILE:
			s.configService <- configMessage{setprefix, map[int]string{filelogprefix: preparedPrefix}}
			<-s.configServiceResponse
		default:
			panic(m008)
		}
	} else {
		panic(m004)
	}
}

// var mtx sync.Mutex
// var isUp bool

// Log writes a log message to a specified destination.
// The destination parameter specifies the log destination, where the data will be written to.
// The logValues parameter consists of one or multiple values that are logged.
func Log(destination int, logValues ...any) {
	// if c.checkState(running) {
	if s.isUp {
		switch destination {
		case STDOUT:
			s.logData <- logMessage{STDOUT, logValues}
		case FILE:
			s.logData <- logMessage{FILE, logValues}
		case MULTI:
			s.logData <- logMessage{MULTI, logValues}
		default:
			panic(m007)
		}
		// } else {
		// panic(m004)
		// }
	}
}
