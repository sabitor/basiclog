package simplelog

import (
	"fmt"
	"log"
	"os"
	"strings"
	"sync"
)

const LineBreak = "\n"

// log targets
const (
	STDOUT = iota
	FILE
	MULTI
)

// configuration properties
const (
	SETLOGNAME = iota
	CHANGELOGNAME
)

// service states
const (
	STOPPED = iota
	RUNNING
)

type signal struct{}

type logMessage struct {
	target int
	prefix string
	record string
}

type cfgMessage struct {
	property int
	data     string
}

type simpleLog struct {
	// handler
	fileHandle *os.File
	logHandle  map[string]*log.Logger

	// channels
	data           chan logMessage
	config         chan cfgMessage
	stopLogService chan signal

	// service
	state int
	mtx   sync.Mutex
}

var sLog = &simpleLog{}
var firstFileLogHandler = false

func (sl *simpleLog) logger(target int, msgPrefix string) *log.Logger {
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

func (sl *simpleLog) serviceState() int {
	sl.mtx.Lock()
	defer sl.mtx.Unlock()
	return sl.state
}

func (sl *simpleLog) setServiceState(newState int) {
	sl.mtx.Lock()
	defer sl.mtx.Unlock()
	sl.state = newState
}

func (sl *simpleLog) stdoutLogger(prefix string) *log.Logger {
	sl.mtx.Lock()
	defer sl.mtx.Unlock()
	return sLog.logger(STDOUT, prefix)
}

func (sl *simpleLog) fileLogger(prefix string) *log.Logger {
	sl.mtx.Lock()
	defer sl.mtx.Unlock()
	return sLog.logger(FILE, prefix)
}

func (sl *simpleLog) multiLogger(prefix string) (*log.Logger, *log.Logger) {
	sl.mtx.Lock()
	defer sl.mtx.Unlock()
	return sLog.logger(STDOUT, prefix), sLog.logger(FILE, prefix)
}

func (sl *simpleLog) initialize(buffer int) {
	// setup log handler
	// The log handler map stores log handler with different properties, e.g. target and/or message prefixes.
	sl.logHandle = make(map[string]*log.Logger)

	// setup channels
	sl.data = make(chan logMessage, buffer)
	sl.config = make(chan cfgMessage)
	sl.stopLogService = make(chan signal)

	// setup service state
	sl.state = RUNNING
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
