package simplelog

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func Test_StartService(t *testing.T) {
	StartService(1)

	if s := sLog.serviceState(); s != running {
		t.Error("Expected", running, ", got ", s)
	} else {
		sLog.stopLogService <- signal{}
	}
}

func Test_StopService(t *testing.T) {
	StartService(1)
	StopService()

	if s := sLog.serviceState(); s != stopped {
		t.Error("Expected", stopped, ", got ", s)
		sLog.stopLogService <- signal{}
	}
}

func Test_WriteToStdout(t *testing.T) {
	stdOut := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w

	StartService(1)
	WriteToStdout("Write something to STDOUT.")
	StopService()

	_ = w.Close()

	result, _ := io.ReadAll(r)
	output := string(result)

	os.Stdout = stdOut

	if !strings.Contains(output, "Write something to STDOUT") {
		t.Error("Expected to find:", "Write something to STDOUT", "- but found:", output)
	}
}

func Test_WriteToFile(t *testing.T) {
	logFile := "test.log"
	logStr := "The answer to all questions is"
	logInt := 42

	if _, err := os.Stat(logFile); err == nil {
		os.Remove(logFile)
	}

	StartService(1)
	InitLogFile(logFile)
	WriteToFile(logStr, logInt)
	StopService()

	logRecord := logStr + " " + fmt.Sprint(logInt)
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Error("Expected to find the log file", logFile, "- but got the error:", err)
	} else if !strings.Contains(string(data), logRecord) {
		t.Error("Expected log record to contain:", logRecord, "- but it doesn't:", string(data))
	} else {
		os.Remove(logFile)
	}
}

func Test_WriteToMulti(t *testing.T) {
	// TODO: implement
}
