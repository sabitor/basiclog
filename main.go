package simplelog

import (
	"fmt"
	"os"
	"time"
)

func service() {
	defer close(sLog.data)
	defer close(sLog.stopLogService)
	defer sLog.fileHandle.Close()

	for {
		select {
		case logMsg := <-sLog.data:
			switch logMsg.target {
			case STDOUT:
				sLog.Logger(STDOUT, logMsg.prefix).Print(logMsg.record)
			case FILE:
				logger := sLog.Logger(FILE, logMsg.prefix)
				if logger != nil {
					logger.Print(logMsg.record)
				} else {
					panic("log file name not set")
				}
			case MULTI:
				sLog.Logger(STDOUT, logMsg.prefix).Print(logMsg.record)
				sLog.Logger(FILE, logMsg.prefix).Print(logMsg.record)
			}
		case <-sLog.stopLogService:
			sLog.setRunState(false)
			return
		case cfgMsg := <-sLog.config:
			var err error
			sLog.fileHandle, err = os.OpenFile(cfgMsg.data, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				panic(err)
			}
		}
	}
}

func StartService(msgBuffer int) {
	if !sLog.runState() {
		sLog.initialize(msgBuffer)
		go service()
	} else {
		panic("log service was already started")
	}
}

func StopService() {
	if sLog.runState() {
		// wait until all messages have been logged by the service
		for len(sLog.data) > 0 {
			continue
		}
		// all messages are logged - the services can be stopped gracefully
		sLog.stopLogService <- trigger{}
	}
	// TODO: add panic call
}

func SetLogName(logName string) {
	if sLog.runState() {
		time.Sleep(10 * time.Millisecond) // to keep the logical order of write calls and setup calls
		sLog.config <- cfgMessage{SETLOGNAME, logName}
	} else {
		// TODO: add panic call
		fmt.Println("log service has not been started.")
	}
}

func WriteToStdout(prefix string, values ...any) {
	if sLog.runState() {
		logRecord := assembleToString(values)
		sLog.data <- logMessage{STDOUT, prefix, logRecord}
	} else {
		// TODO: add panic call
		fmt.Println("log service has not been started.")
	}
}

func WriteToFile(prefix string, values ...any) {
	if sLog.runState() {
		logRecord := assembleToString(values)
		sLog.data <- logMessage{FILE, prefix, logRecord}
	} else {
		// TODO: add panic call
		fmt.Println("log service has not been started.")
	}
}

func WriteToMultiple(prefix string, values ...any) {
	if sLog.runState() {
		logRecord := assembleToString(values)
		sLog.data <- logMessage{MULTI, prefix, logRecord}
	} else {
		// TODO: add panic call
		fmt.Println("log service has not been started.")
	}
}
