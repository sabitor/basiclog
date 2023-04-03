package simplelog

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

const LineBreak = "\n"

const (
	STDOUT = iota
	FILE
	MULTI
)

type trigger struct{}

type message struct {
	target int
	prefix string
	data   string
}

type simpleLogger struct {
	// handler
	fileHandle *os.File
	logHandle  map[string]*log.Logger

	// channels
	data        chan message
	stopService chan trigger

	// service
	serviceRunState bool
	mtx             sync.Mutex
}

var sLog = &simpleLogger{}
var firstFileLogHandler = false

func (sl *simpleLogger) Logger(target int, msgPrefix string) *log.Logger {
	// build key for log handler map
	key := fmt.Sprintf("%d_%s", target, msgPrefix)
	if _, found := sl.logHandle[key]; !found {
		// create a new log handler
		switch target {
		case STDOUT:
			sl.logHandle[key] = log.New(os.Stdout, "", 0)
		case FILE:
			sl.logHandle[key] = log.New(sl.fileHandle, msgPrefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix)
			if !firstFileLogHandler {
				// the first file log event always adds an empty line to the log file at the beginning
				sl.fileHandle.WriteString(LineBreak)
				firstFileLogHandler = true
			}
		}
	}

	return sl.logHandle[key]
}

func (sl *simpleLogger) ServiceRunState() bool {
	sl.mtx.Lock()
	defer sl.mtx.Unlock()
	return sl.serviceRunState
}

func (sl *simpleLogger) SetServiceRunState(newState bool) {
	sl.mtx.Lock()
	defer sl.mtx.Unlock()
	sl.serviceRunState = newState
}

func (sl *simpleLogger) SetLogName(logName string) {
	var err error
	sl.fileHandle, err = os.OpenFile(logName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
}

func (sl *simpleLogger) initialize(buffer int) {
	// setup log file using an initial log name
	// initialLogName, err := os.Executable()
	// if err != nil {
	// 	panic(err)
	// }
	// initialLogName += ".log"
	// sl.fileHandle, err = os.OpenFile(initialLogName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	// if err != nil {
	// 	panic(err)
	// }

	// setup log handler
	// This map stored log handler with different properties, e.g. target and/or message prefixes.
	sl.logHandle = make(map[string]*log.Logger)

	// setup channels
	sl.data = make(chan message, buffer)
	sl.stopService = make(chan trigger)

	// setup state
	sl.serviceRunState = true
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
