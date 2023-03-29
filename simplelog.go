package simplelog

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
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
	serviceStop chan trigger
	// serviceStarted chan trigger

	// service parts
	serviceRunState bool
	mtx             sync.Mutex
}

var simLog = &simpleLog{}

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

func (b *simpleLog) FileHandle() *os.File {
	return b.fileHandle
}

func (b *simpleLog) StdoutLogger() *log.Logger {
	return b.stdoutLogger
}

func (b *simpleLog) FileLogger() *log.Logger {
	return b.fileLogger
}

func (b *simpleLog) ServiceRunState() bool {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	return b.serviceRunState
}

func (b *simpleLog) SetServiceRunState(newState bool) {
	b.mtx.Lock()
	defer b.mtx.Unlock()
	b.serviceRunState = newState
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
	b.stdoutChan = make(chan data, 10)
	b.fileChan = make(chan data, 10)
	b.multiChan = make(chan data, 10)
	b.serviceStop = make(chan trigger)

	simLog.serviceRunState = true
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
