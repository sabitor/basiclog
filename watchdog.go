package simplelog

import (
	"time"
)

// general
const (
	heartBeatInterval = time.Second
	latency_factor    = 2
)

// watchdog is a structure to watch and handle the communication with the assigned log service.
type watchdog struct {
	serviceRunning         chan signal    // the channel for sending a serviceRunning signal to the watchdog
	serviceRunningResponse chan bool      // the channel for sending a serviceRunningResponse message to the caller
	heartBeatMonitor       chan time.Time // the channel is used to monitor and evaluate the heartbeats sent by the log service
}

// watchdog instance
var w = &watchdog{}

// init starts the watchdog.
// The watchdog monitors the log service and based on the monitoring results
// it can answer questions regarding the availability of this service.
func init() {
	w.serviceRunning = make(chan signal, 1)
	w.serviceRunningResponse = make(chan bool, 1)
	w.heartBeatMonitor = make(chan time.Time)

	watchdogRunning := make(chan bool)
	go w.run(watchdogRunning)
	if !<-watchdogRunning {
		panic("watchdog is not running")
	}
}

// run detects if the log service is running.
func (w *watchdog) run(watchdogRunning chan bool) {
	var t time.Time = time.Now()
	var timeDiff_ns int64
	var max_service_response_delay int64 = latency_factor * heartBeatInterval.Nanoseconds()
	defer close(watchdogRunning)

	for {
		select {
		case watchdogRunning <- true:
		case t = <-w.heartBeatMonitor:
			timeDiff_ns = time.Until(t).Nanoseconds() * (-1)
		case <-w.serviceRunning:
			if timeDiff_ns == 0 || timeDiff_ns > max_service_response_delay {
				// if the log service has not responded within a defined interval the watchdog assumes it isn't running
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

// getHeartBeatMonitor returns the heartBeatMonitor channel
func (w *watchdog) getHeartBeatMonitor() chan time.Time {
	return w.heartBeatMonitor
}
