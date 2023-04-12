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
		t.Error("Expected state", running, ", but got ", s)
	} else {
		sLog.stopLogService <- signal{}
	}
}

func Test_StopService(t *testing.T) {
	StartService(1)
	StopService()

	if s := sLog.serviceState(); s != stopped {
		t.Error("Expected state", stopped, ", but got ", s)
		sLog.stopLogService <- signal{}
	}
}

func Test_InitLogFile(t *testing.T) {
	logFile := "test1.log"
	filePerms := "-rw-r--r--"
	fileSize := 0

	if _, err := os.Stat(logFile); err == nil {
		os.Remove(logFile)
	}

	StartService(1)
	InitLogFile(logFile)
	StopService()

	data, err := os.Stat(logFile)
	if err != nil {
		t.Error("Expected to find file", logFile, "- but got:", err)
	} else if data.Mode().String() != filePerms {
		t.Error("Expected file permissions", filePerms, "but found:", data.Mode().String())
	} else if data.Size() != 0 {
		t.Error("Expected file size", fileSize, "but found:", data.Size())
	} else {
		os.Remove(logFile)
	}
}

func Test_ChangeLogFile(t *testing.T) {
	logFile1 := "test1.log"
	logFile2 := "test2.log"
	filePerms := "-rw-r--r--"
	fileSize := 0

	if _, err := os.Stat(logFile1); err == nil {
		os.Remove(logFile1)
	}
	if _, err := os.Stat(logFile2); err == nil {
		os.Remove(logFile2)
	}

	StartService(1)
	InitLogFile(logFile1)
	ChangeLogFile(logFile2)
	StopService()

	data, err := os.Stat(logFile1)
	if err != nil {
		t.Error("Expected to find file", logFile1, "- but got:", err)
	} else if data.Mode().String() != filePerms {
		t.Error("Expected file permissions", filePerms, "but found:", data.Mode().String())
	} else if data.Size() != 0 {
		t.Error("Expected file size", fileSize, "but found:", data.Size())
	} else {
		os.Remove(logFile1)
	}

	data, err = os.Stat(logFile2)
	if err != nil {
		t.Error("Expected to find file", logFile2, "- but got:", err)
	} else if data.Mode().String() != filePerms {
		t.Error("Expected file permissions", filePerms, "but found:", data.Mode().String())
	} else if data.Size() != 0 {
		t.Error("Expected file size", fileSize, "but found:", data.Size())
	} else {
		os.Remove(logFile2)
	}
}

func Test_WriteToStdout(t *testing.T) {
	stdOut := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w

	StartService(1)
	WriteToStdout("The answer to all questions is", 42)
	StopService()

	_ = w.Close()

	result, _ := io.ReadAll(r)
	output := string(result)

	os.Stdout = stdOut

	if !strings.Contains(output, "The answer to all questions is "+fmt.Sprint(42)) {
		t.Error("Expected to find:", "The answer to all questions is "+fmt.Sprint(42), "- but found:", output)
	}
}

func Test_WriteToFile(t *testing.T) {
	logFile := "test1.log"

	if _, err := os.Stat(logFile); err == nil {
		os.Remove(logFile)
	}

	StartService(1)
	InitLogFile(logFile)
	WriteToFile("The answer to all questions is", 42)
	StopService()

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Error("Expected to find file", logFile, "- but got:", err)
	} else if !strings.Contains(string(data), "The answer to all questions is "+fmt.Sprint(42)) {
		t.Error("Expected log record contains:", "The answer to all questions is "+fmt.Sprint(42), "- but it doesn't:", string(data))
	} else {
		os.Remove(logFile)
	}
}

func Test_WriteToMulti(t *testing.T) {
	stdOut := os.Stdout
	logFile := "test2.log"

	r, w, _ := os.Pipe()
	os.Stdout = w

	if _, err := os.Stat(logFile); err == nil {
		os.Remove(logFile)
	}

	StartService(1)
	InitLogFile(logFile)
	WriteToMulti("The answer to all questions is", 42)
	StopService()

	_ = w.Close()

	result, _ := io.ReadAll(r)
	output := string(result)

	os.Stdout = stdOut

	// check output sent to stdout
	if !strings.Contains(output, "The answer to all questions is "+fmt.Sprint(42)) {
		t.Error("Expected to find:", "The answer to all questions is "+fmt.Sprint(42), "- but found:", output)
	}

	// check output sent to file
	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Error("Expected to find file", logFile, "- but got:", err)
	} else if !strings.Contains(string(data), "The answer to all questions is "+fmt.Sprint(42)) {
		t.Error("Expected log record contains:", "The answer to all questions is "+fmt.Sprint(42), "- but it doesn't:", string(data))
	} else {
		os.Remove(logFile)
	}
}
