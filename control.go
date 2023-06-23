package simplelog

import (
	"strconv"
)

// control instance
var c = new(control)

// control is a structure used to control the log service and log service workflows.
type control struct {
	checkServiceState         chan int  // the channel for receiving a state check request from the caller
	checkServiceStateResponse chan bool // the channel for sending a boolean response to the caller
	serviceState              chan int  // the channel for receiving a state change request from the caller
	serviceAction             chan int  // TBD
	serviceActionResponse     chan bool // TBD
}

// init starts the control.
// The control monitors the log service and can answer questions regarding the state of the service.
func init() {
	c.checkServiceState = make(chan int, 1)
	c.checkServiceStateResponse = make(chan bool, 1)
	c.serviceState = make(chan int)
	c.serviceAction = make(chan int)
	c.serviceActionResponse = make(chan bool)

	controlRunning := make(chan bool)
	go c.run(controlRunning)
	if !<-controlRunning {
		panic("control is not running")
	}
}

// run represents the control service.
// This utility function runs in a dedicated goroutine and is started when the init function is implicitly called.
// It handles requests by listening on the following channels:
//   - resetControl
//   - serviceState
//   - checkServiceState
func (c *control) run(controlRunning chan bool) {
	var newState, totalState int

	for {
		select {
		case controlRunning <- true:
		case action := <-c.serviceAction:
			switch action {
			case start:
				// allocate service channels
				buf, _ := strconv.Atoi(convertToString(s.attribute[logbuffer]))
				s.data = make(chan logMessage, buf)
				s.config = make(chan configMessage)
				s.stop = make(chan signal)
				s.confirmed = make(chan signal)

				// reset state attribute (after the log service has restarted)
				if totalState == stopped {
					totalState = 0
				}

				// start service
				go s.run()
				// reply to the caller when the service has started
				go func() {
					for {
						// wait until the service is running
						if c.testServiceState(running) {
							break
						}
					}
					c.serviceActionResponse <- true
				}()
			case stop:
				// stop service
				s.stop <- signal{}
				// reply to the caller when the service has stopped
				go func() {
					for {
						// wait until the service is stopped
						if c.testServiceState(stopped) {
							break
						}
					}
					c.serviceActionResponse <- true

					// cleanup service resources
					s.fileDesc.Close()
				}()
			}
		case newState = <-c.serviceState:
			if newState == stopped {
				// unset all other states
				totalState = newState
			} else {
				// add new state to the total state
				totalState |= newState
			}
		case state := <-c.checkServiceState:
			if totalState&state == state {
				c.checkServiceStateResponse <- true
			} else {
				c.checkServiceStateResponse <- false
			}
		}
	}
}

// service handles actions to be processed by the log service.
func (c *control) service(action int) bool {
	c.serviceAction <- action
	return <-c.serviceActionResponse
}

// checkServiceState checks if the service has set the specified state.
func (c *control) testServiceState(state int) bool {
	c.checkServiceState <- state
	return <-c.checkServiceStateResponse
}

// setServiceState sets the state of the log service.
func (c *control) setServiceState(state int) {
	c.serviceState <- state
}
