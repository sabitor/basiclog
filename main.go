package basiclog

func WriteToStdout(toLog bool, values ...any) {
	if toLog {
		logMessage := assembleToString(values)
		ld := data{"", logMessage}
		bLog.StdoutChan() <- ld
	}
}

func WriteToFile(toLog bool, prefix string, values ...any) {
	if toLog {
		logMessage := assembleToString(values)
		ld := data{prefix, logMessage}
		bLog.FileChan() <- ld
	}
}

func WriteToMultiple(toLog bool, prefix string, values ...any) {
	if toLog {
		logMessage := assembleToString(values)
		ld := data{prefix, logMessage}
		bLog.MultiChan() <- ld
	}
}

func StartService(logName string) {
	if len(bLog.ServiceStarted()) == 0 {
		bLog.initialize(logName)
		go func() {
			for {
				select {
				case logToStdout := <-bLog.StdoutChan():
					bLog.StdoutLogger().SetPrefix(logToStdout.prefix)
					bLog.StdoutLogger().Print(logToStdout.msg)
				case logToFile := <-bLog.FileChan():
					bLog.FileLogger().SetPrefix(logToFile.prefix)
					bLog.FileLogger().Print(logToFile.msg)
				case logToMulti := <-bLog.MultiChan():
					bLog.StdoutLogger().Print(logToMulti.msg)
					bLog.FileLogger().SetPrefix(logToMulti.prefix)
					bLog.FileLogger().Print(logToMulti.msg)
				case <-bLog.ServiceStop():
					<-bLog.ServiceStarted()
					// signal that the service is closed
					bLog.ServiceStop() <- trigger{}
					return
				}
			}
		}()
	}
}

func StopService() {
	defer bLog.cleanup()
	// close the service
	bLog.ServiceStop() <- trigger{}
	// wait for the service to close
	<-bLog.ServiceStop()
}
