package simplelog

import (
	"time"
)

// general
const (
	heartBeatInterval = 1000 * time.Millisecond
	latency_factor    = 2
)

// watchdog is a structure to watch and handle the communication with the assigned log service.
type watchdog struct {
	serviceRunning         chan signal // the channel for sending a serviceRunning signal to the watchdog
	serviceRunningResponse chan bool   // the channel for sending a serviceRunningResponse message to the caller
}

// global watchdog instance
var w = &watchdog{}

// init starts the watchdog.
// The watchdog monitors the log service and based on the monitoring results
// it can answer questions regarding the availability of this service.
func init() {
	w.serviceRunning = make(chan signal, 1)
	w.serviceRunningResponse = make(chan bool, 1)

	watchdogRunning := make(chan bool)
	go w.run(watchdogRunning)
	if !<-watchdogRunning {
		panic("watchdog is not running")
	}
}

// watchdog detects if the log service is running.
func (w *watchdog) run(watchdogRunning chan bool) {
	var t time.Time = time.Now()
	var timeDiff_ms int64
	var max_service_response_delay int64 = latency_factor * heartBeatInterval.Milliseconds()
	defer close(watchdogRunning)

	for {
		select {
		case watchdogRunning <- true:
		case t = <-s.getServiceHeartBeat():
		case <-w.serviceRunning:
			timeDiff_ms = time.Until(t).Milliseconds() * (-1)
			if timeDiff_ms == 0 || timeDiff_ms > max_service_response_delay {
				// if the log service has not responded within a defined interval it is assumed the service isn't running
				w.serviceRunningResponse <- false
			} else {
				w.serviceRunningResponse <- true
			}
		}
	}
}

// checkService checks if the service is running (true) or if it is not running (false).
func (w *watchdog) checkService() bool {
	w.serviceRunning <- signal{}
	if <-w.serviceRunningResponse {
		return true
	} else {
		return false
	}
}
