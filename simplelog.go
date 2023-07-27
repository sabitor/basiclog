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
	m000 = "log service is not running"
	m001 = "log file not initialized"
	m002 = "log service was already started"
	m003 = "log service is not running"
	m004 = "log service has not been started"
	m005 = "log file already initialized"
	m006 = "unknown log destination specified"
)

// Log writes a log message to a specified destination.
// The destination parameter specifies the log destination, where the data will be written to.
// The logValues parameter consists of one or multiple values that are logged.
func Log(destination int, logValues ...any) {
	if s.isActive() {
		switch destination {
		case STDOUT:
			dataQueue <- logMessage{STDOUT, logValues}
		case FILE:
			dataQueue <- logMessage{FILE, logValues}
		case MULTI:
			dataQueue <- logMessage{MULTI, logValues}
		default:
			panic(m006)
		}
	} else {
		panic(m004)
	}
}

// InitLog initializes the log file.
// The logName specifies the name of the log file.
// The append flag indicates whether messages are appended to the existing log file (true),
// or on a new run whether the old log is truncated (false).
func InitLog(logName string, append bool) {
	if s.isActive() {
		if s.desc != nil {
			panic(m005)
		}
		var err error
		var flag int
		if append {
			flag = os.O_APPEND | os.O_CREATE | os.O_WRONLY
		} else {
			flag = os.O_TRUNC | os.O_CREATE | os.O_WRONLY
		}
		configService <- configMessage{initlog, map[int]any{logflag: flag, logfilename: logName}}
		if err = <-configServiceResponse; err != nil {
			panic(err)
		}
	} else {
		panic(m004)
	}
}

// SetPrefix sets the prefix for logging lines.
// If the prefix should also contain actual date and time data, the following placeholders
// can be applied for given data:
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
// Note that not all placeholders have to be used and they can be used in any order.
//
// The destination specifies the name of the log destination where the prefix should be used, e.g. STDOUT or FILE.
// The prefix specifies the prefix for each logging line for a given log destination.
func SetPrefix(destination int, prefix string) {
	if s.isActive() {
		preparedPrefix := preprocessPrefix(prefix)
		switch destination {
		case STDOUT:
			configService <- configMessage{setprefix, map[int]any{stdoutlogprefix: preparedPrefix}}
			<-configServiceResponse
		case FILE:
			configService <- configMessage{setprefix, map[int]any{filelogprefix: preparedPrefix}}
			<-configServiceResponse
		default:
			panic(m006)
		}
	} else {
		panic(m004)
	}
}

// Shutdown stops the log service and does some cleanup.
// Before the log service is stopped, all pending log messages are flushed and resources are released.
// Archiving a log file means that it will be renamed and no new messages will be appended on a new run.
// The archived log file is of the following format: <orig file name>_yyyymmddHHMMSS.
// The archivelog flag indicates whether the log file will be archived (true) or not (false).
func Shutdown(archivelog bool) {
	if s.isActive() {
		s.stop(archivelog)
		s.setActive(false)
	} else {
		panic(m003)
	}
}

// Startup starts the log service.
// The log service runs in its own goroutine.
// The bufferSize specifies the number of log messages which can be buffered before the log service blocks.
func Startup(bufferSize int) {
	if !s.isActive() {
		dataQueue = make(chan logMessage, bufferSize)
		configService = make(chan configMessage)
		configServiceResponse = make(chan error)
		stopService = make(chan bool)
		stopServiceResponse = make(chan struct{})
		serviceRunning := make(chan bool)
		go s.run(serviceRunning)
		if !<-serviceRunning {
			panic(m000)
		} else {
			s.setActive(true)
		}
	} else {
		panic(m002)
	}
}

// SwitchLog closes the current log file and a new log file with the specified name is created and used.
// Thereby, the current log file is not deleted, the new log file must not exist and the log service
// doesn't need to be stopped for this task. The new log file must not exist.
// The newLogName specifies the name of the new log to switch to.
func SwitchLog(newLogName string) {
	if s.isActive() {
		var err error
		flag := os.O_EXCL | os.O_CREATE | os.O_WRONLY
		configService <- configMessage{switchlog, map[int]any{logflag: flag, logfilename: newLogName}}
		if err = <-configServiceResponse; err != nil {
			panic(err)
		}
	} else {
		panic(m004)
	}
}
