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

const (
	SETLOGNAME = iota
)

type trigger struct{}

type logMessage struct {
	target int
	prefix string
	record string
}

type cfgMessage struct {
	property int
	data     string
}

type simpleLogger struct {
	// handler
	fileHandle *os.File
	logHandle  map[string]*log.Logger

	// channels
	data           chan logMessage
	config         chan cfgMessage
	stopLogService chan trigger

	// service
	serviceRunState bool
	mtx             sync.Mutex
}

var sLog = &simpleLogger{}
var firstFileLogHandler = false

func (sl *simpleLogger) logger(target int, msgPrefix string) *log.Logger {
	// build key for log handler map
	key := fmt.Sprintf("%d_%s", target, msgPrefix)
	if _, found := sl.logHandle[key]; !found {
		// create a new log handler
		switch target {
		case STDOUT:
			sl.logHandle[key] = log.New(os.Stdout, "", 0)
		case FILE:
			if sl.fileHandle != nil {
				sl.logHandle[key] = log.New(sl.fileHandle, msgPrefix, log.Ldate|log.Ltime|log.Lmicroseconds|log.Lmsgprefix)
				if !firstFileLogHandler {
					// the first file log event always adds an empty line to the log file at the beginning
					sl.fileHandle.WriteString(LineBreak)
					firstFileLogHandler = true
				}
			}
		}
	}

	return sl.logHandle[key]
}

func (sl *simpleLogger) runState() bool {
	sl.mtx.Lock()
	defer sl.mtx.Unlock()
	return sl.serviceRunState
}

func (sl *simpleLogger) setRunState(newState bool) {
	sl.mtx.Lock()
	defer sl.mtx.Unlock()
	sl.serviceRunState = newState
}

func (sl *simpleLogger) stdoutLogger(prefix string) *log.Logger {
	sl.mtx.Lock()
	defer sl.mtx.Unlock()
	return sLog.logger(STDOUT, prefix)
}

func (sl *simpleLogger) fileLogger(prefix string) *log.Logger {
	sl.mtx.Lock()
	defer sl.mtx.Unlock()
	return sLog.logger(FILE, prefix)
}

func (sl *simpleLogger) multiLogger(prefix string) (*log.Logger, *log.Logger) {
	sl.mtx.Lock()
	defer sl.mtx.Unlock()
	return sLog.logger(STDOUT, prefix), sLog.logger(FILE, prefix)
}

func (sl *simpleLogger) initialize(buffer int) {
	// setup log handler
	// The log handler map stores log handler with different properties, e.g. target and/or message prefixes.
	sl.logHandle = make(map[string]*log.Logger)

	// setup channels
	sl.data = make(chan logMessage, buffer)
	sl.config = make(chan cfgMessage)
	sl.stopLogService = make(chan trigger)

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
