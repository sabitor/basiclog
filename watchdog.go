package simplelog

import (
	"time"
)

// watchdog instance
var w = new(watchdog)

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

// run represents the watchdog service.
// This utility function runs in a dedicated goroutine and is started when the init function is implicitly called.
// It handles requests by listening on the following channels:
//   - heartBeatMonitor
//   - serviceRunning
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

// getServiceRunning returns the serviceRunning channel
func (w *watchdog) getServiceRunning() chan signal {
	return w.serviceRunning
}

// getServiceRunningResponse returns the serviceRunningResponse channel
func (w *watchdog) getServiceRunningResponse() chan bool {
	return w.serviceRunningResponse
}

// getHeartBeatMonitor returns the heartBeatMonitor channel
func (w *watchdog) getHeartBeatMonitor() chan time.Time {
	return w.heartBeatMonitor
}
