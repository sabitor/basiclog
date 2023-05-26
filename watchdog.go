package simplelog

import (
	"time"
)

// general
const (
	heartBeatInterval = 1000 * time.Millisecond
	latency_factor    = 2
)

// init starts a watchdog.
// The watchdog monitors the log service and based on the monitoring results
// it can answer questions regarding the availability of this service.
func init() {
	s.serviceRunning = make(chan signal, 1)
	s.serviceRunningResponse = make(chan bool, 1)
	s.serviceHeartBeat = make(chan time.Time)

	watchdogRunning := make(chan bool)
	go watchdog(watchdogRunning)
	if !<-watchdogRunning {
		panic("watchdog is not running")
	}
}

// watchdog detects if the log service is running.
func watchdog(watchdogRunning chan bool) {
	var t time.Time = time.Now()
	var timeDiff_ms int64
	var max_service_response_delay int64 = latency_factor * heartBeatInterval.Milliseconds()
	defer close(watchdogRunning)

	for {
		select {
		case watchdogRunning <- true:
		case t = <-s.serviceHeartBeat:
		case <-s.serviceRunning:
			timeDiff_ms = time.Until(t).Milliseconds() * (-1)
			if timeDiff_ms == 0 || timeDiff_ms > max_service_response_delay {
				// if the log service has not responded within a defined interval it is assumed the service isn't running
				s.serviceRunningResponse <- false
			} else {
				s.serviceRunningResponse <- true
			}
		}
	}
}
