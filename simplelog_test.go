package simplelog

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func Test_service_startup(t *testing.T) {
	Startup(1)

	if a := w.checkService(); a != true {
		t.Error("Expected state true but got", a)
	} else {
		s.stop <- signal{}
		<-s.confirmed
	}
}

func Test_service_shutdown(t *testing.T) {
	Startup(1)
	Shutdown()

	if a := w.checkService(); a == true {
		t.Error("Expected state false but got", a)
		s.stop <- signal{}
		<-s.confirmed
	}
}

func Test_service_initLogFile(t *testing.T) {
	logFile := "test1.log"
	filePerms := "-rw-r--r--"
	fileSize := 0

	if _, err := os.Stat(logFile); err == nil {
		os.Remove(logFile)
	}

	Startup(1)
	InitLogFile(logFile)
	Shutdown()

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

func Test_service_changeLogFile(t *testing.T) {
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

	Startup(1)
	InitLogFile(logFile1)
	ChangeLogName(logFile2)
	Shutdown()

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

func Test_service_writeToStdout(t *testing.T) {
	stdOut := os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w

	Startup(1)
	WriteToStdout("The answer to all questions is", 42)
	Shutdown()

	_ = w.Close()

	result, _ := io.ReadAll(r)
	output := string(result)

	os.Stdout = stdOut

	if !strings.Contains(output, "The answer to all questions is "+fmt.Sprint(42)) {
		t.Error("Expected to find:", "The answer to all questions is "+fmt.Sprint(42), "- but found:", output)
	}
}

func Test_service_writeToFile(t *testing.T) {
	logFile := "test1.log"

	if _, err := os.Stat(logFile); err == nil {
		os.Remove(logFile)
	}

	Startup(1)
	InitLogFile(logFile)
	WriteToFile("The answer to all questions is", 42)
	Shutdown()

	data, err := os.ReadFile(logFile)
	if err != nil {
		t.Error("Expected to find file", logFile, "- but got:", err)
	} else if !strings.Contains(string(data), "The answer to all questions is "+fmt.Sprint(42)) {
		t.Error("Expected log record contains:", "The answer to all questions is "+fmt.Sprint(42), "- but it doesn't:", string(data))
	} else {
		os.Remove(logFile)
	}
}

func Test_service_writeToMulti(t *testing.T) {
	stdOut := os.Stdout
	logFile := "test2.log"

	r, w, _ := os.Pipe()
	os.Stdout = w

	if _, err := os.Stat(logFile); err == nil {
		os.Remove(logFile)
	}

	Startup(1)
	InitLogFile(logFile)
	WriteToMulti("The answer to all questions is", 42)
	Shutdown()

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
		t.Error("Expected log record:", "The answer to all questions is "+fmt.Sprint(42), "- but got:", string(data))
	} else {
		os.Remove(logFile)
	}
}
