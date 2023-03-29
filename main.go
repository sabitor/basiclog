package simplelog

import (
	"errors"
	"fmt"
	"runtime"
	"time"
)

func WriteToStdout(enabled bool, values ...any) {
	if enabled {
		if simLog.ServiceRunState() {
			logMessage := assembleToString(values)
			ld := data{"", logMessage}
			simLog.StdoutChan() <- ld
		} else {
			fmt.Println("Log service is not running.")
		}
	}
}

func WriteToFile(enabled bool, prefix string, values ...any) {
	if enabled {
		if simLog.ServiceRunState() {
			logMessage := assembleToString(values)
			ld := data{prefix, logMessage}
			simLog.FileChan() <- ld
		} else {
			fmt.Println("Log service is not running.")
		}
	}
}

func WriteToMultiple(enabled bool, prefix string, values ...any) {
	if enabled {
		if simLog.ServiceRunState() {
			logMessage := assembleToString(values)
			ld := data{prefix, logMessage}
			simLog.MultiChan() <- ld
		} else {
			fmt.Println("Log service is not running.")
		}
	}
}

func StartService(logName string) error {
	var err error
	if !simLog.ServiceRunState() {
		simLog.initialize(logName)
		go func() {
			defer close(simLog.StdoutChan())
			defer close(simLog.FileChan())
			defer close(simLog.MultiChan())
			defer close(simLog.ServiceStop())
			defer simLog.fileHandle.Close()

			for {
				select {
				case logToStdout := <-simLog.StdoutChan():
					simLog.StdoutLogger().Print(logToStdout.msg)
				case logToFile := <-simLog.FileChan():
					simLog.FileLogger().SetPrefix(logToFile.prefix)
					simLog.FileLogger().Print(logToFile.msg)
				case logToMulti := <-simLog.MultiChan():
					simLog.StdoutLogger().Print(logToMulti.msg)
					simLog.FileLogger().SetPrefix(logToMulti.prefix)
					simLog.FileLogger().Print(logToMulti.msg)
				case <-simLog.ServiceStop():
					fmt.Println("Service stopped.")
					simLog.SetServiceRunState(false)
					return
				}
			}
		}()
	} else {
		_, filename, line, _ := runtime.Caller(1)
		errMsg := fmt.Sprintf("Log service was already started - %s: %d\n", filename, line)
		err = errors.New(errMsg)
	}

	return err
}

func StopService() {
	time.Sleep(time.Millisecond) // give kicked off log messages a chance to be logged
	if simLog.ServiceRunState() {
		simLog.ServiceStop() <- trigger{}
	}
}
