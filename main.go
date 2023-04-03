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
		case logMessage := <-sLog.data:
			switch logMessage.target {
			case STDOUT:
				sLog.Logger(STDOUT, logMessage.prefix).Print(logMessage.logData)
			case FILE:
				logger := sLog.Logger(FILE, logMessage.prefix)
				if logger != nil {
					logger.Print(logMessage.logData)
				} else {
					panic("log file name not set")
				}
			case MULTI:
				sLog.Logger(STDOUT, logMessage.prefix).Print(logMessage.logData)
				sLog.Logger(FILE, logMessage.prefix).Print(logMessage.logData)
			}
		case <-sLog.stopLogService:
			sLog.setRunState(false)
			return
		case configMessage := <-sLog.task:
			var err error
			sLog.fileHandle, err = os.OpenFile(configMessage.cfgData, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
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
		task := config{SETLOGNAME, logName}
		sLog.task <- task
	} else {
		// TODO: add panic call
		fmt.Println("log service has not been started.")
	}
}

func WriteToStdout(prefix string, values ...any) {
	if sLog.runState() {
		logMessage := assembleToString(values)
		data := message{STDOUT, prefix, logMessage}
		sLog.data <- data
	} else {
		// TODO: add panic call
		fmt.Println("log service has not been started.")
	}
}

func WriteToFile(prefix string, values ...any) {
	if sLog.runState() {
		logMessage := assembleToString(values)
		data := message{FILE, prefix, logMessage}
		sLog.data <- data
	} else {
		// TODO: add panic call
		fmt.Println("log service has not been started.")
	}
}

func WriteToMultiple(prefix string, values ...any) {
	if sLog.runState() {
		logMessage := assembleToString(values)
		data := message{MULTI, prefix, logMessage}
		sLog.data <- data
	} else {
		// TODO: add panic call
		fmt.Println("log service has not been started.")
	}
}
