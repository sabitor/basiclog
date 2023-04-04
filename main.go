package simplelog

import (
	"os"
	"time"
)

const (
	msg001 = "log file name not set"
	msg002 = "log service was already started"
	msg003 = "log service is not running"
	msg004 = "log service has not been started"
	msg005 = "log file name already set"
)

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
					panic(msg001)
				}
			case multi:
				stdoutLogHandle, fileLogHandle := sLog.multiLog(logMsg.prefix)
				stdoutLogHandle.Print(logMsg.record)
				if fileLogHandle != nil {
					fileLogHandle.Print(logMsg.record)
				} else {
					panic(msg001)
				}
			}
		case cfgMsg := <-sLog.config:
			sLog.mtx.Lock()
			var err error
			sLog.fileHandle, err = os.OpenFile(cfgMsg.data, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
			sLog.mtx.Unlock()
		}
	}
}

func StartService(msgBuffer int) {
	if sLog.serviceState() == stopped {
		sLog.initialize(msgBuffer)
		go service()
	} else {
		panic(msg002)
	}
}

func StopService() {
	defer close(sLog.data)
	defer close(sLog.stopLogService)
	defer sLog.fileHandle.Close()

	if sLog.serviceState() == running {
		// wait until all messages have been logged by the service
		for len(sLog.data) > 0 {
			continue
		}
		// all messages are logged - the services can be stopped gracefully
		sLog.stopLogService <- signal{}
	} else {
		panic(msg003)
	}
}

func SetLogName(logName string) {
	if sLog.serviceState() == running {
		time.Sleep(10 * time.Millisecond) // to keep the logical order of goroutine function calls
		if sLog.fileHandle == nil {
			sLog.config <- cfgMessage{setlogname, logName}
		} else {
			panic(msg005)
		}
	} else {
		panic(msg004)
	}
}

func ChangeLogName(newLogName string) {
	if sLog.serviceState() == running {
		// TODO: implement
	} else {
		panic(msg004)
	}
}

func WriteToStdout(prefix string, values ...any) {
	if sLog.serviceState() == running {
		logRecord := assembleToString(values)
		sLog.data <- logMessage{stdout, prefix, logRecord}
	} else {
		panic(msg004)
	}
}

func WriteToFile(prefix string, values ...any) {
	if sLog.serviceState() == running {
		logRecord := assembleToString(values)
		sLog.data <- logMessage{file, prefix, logRecord}
	} else {
		panic(msg004)
	}
}

func WriteToMultiple(prefix string, values ...any) {
	if sLog.serviceState() == running {
		logRecord := assembleToString(values)
		sLog.data <- logMessage{multi, prefix, logRecord}
	} else {
		panic(msg004)
	}
}
