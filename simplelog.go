package simplelog

import (
	"fmt"
	"log"
	"os"
	"strings"
)

const LineBreak = "\n"

type trigger struct{}

type data struct {
	prefix string
	msg    string
}

type simpleLog struct {
	// handler
	fileHandle   *os.File
	stdoutLogger *log.Logger
	fileLogger   *log.Logger

	// data channels
	stdoutChan chan data
	fileChan   chan data
	multiChan  chan data

	// service channels
	serviceStop    chan trigger
	serviceStarted chan trigger
}

var bLog = &simpleLog{}

func (b *simpleLog) StdoutChan() chan data {
	return b.stdoutChan
}

func (b *simpleLog) FileChan() chan data {
	return b.fileChan
}

func (b *simpleLog) MultiChan() chan data {
	return b.multiChan
}

func (b *simpleLog) ServiceStop() chan trigger {
	return b.serviceStop
}

func (b *simpleLog) ServiceStarted() chan trigger {
	return b.serviceStarted
}

func (b *simpleLog) FileHandle() *os.File {
	return b.fileHandle
}

func (b *simpleLog) StdoutLogger() *log.Logger {
	return b.stdoutLogger
}

func (b *simpleLog) FileLogger() *log.Logger {
	return b.fileLogger
}

func (b *simpleLog) initialize(logName string) {
	// setup log file
	var err error
	b.fileHandle, err = os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	b.fileHandle.WriteString(LineBreak)

	// setup log handlers
	b.stdoutLogger = log.New(os.Stdout, "", 0)
	b.fileLogger = log.New(b.fileHandle, "", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix)

	// setup data and service channels
	b.stdoutChan = make(chan data)
	b.fileChan = make(chan data)
	b.multiChan = make(chan data)
	b.serviceStop = make(chan trigger)
	b.serviceStarted = make(chan trigger, 1)

	// all is setup - mark service as started (prevents a second service from being started at the same time)
	bLog.ServiceStarted() <- trigger{}
}

func (b *simpleLog) cleanup() {
	close(b.stdoutChan)
	close(b.fileChan)
	close(b.multiChan)
	close(b.serviceStop)

	b.fileHandle.Close()
}

func assembleToString(values []any) string {
	valueList := make([]string, len(values))
	for i, v := range values {
		if s, ok := v.(string); ok {
			valueList[i] = s
		} else {
			valueList[i] = fmt.Sprint(v)
		}
	}
	msg := strings.Join(valueList, " ")
	return msg
}
