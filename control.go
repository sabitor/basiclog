package simplelog

import (
	"strconv"
)

// control instance
var c = new(control)

// control is a structure used to control the log service and log service workflows.
type control struct {
	checkServiceState         chan int    // the channel for receiving a state check request from the caller
	checkServiceStateResponse chan bool   // the channel for sending a boolean response to the caller
	setServiceState           chan int    // the channel for receiving a state change request from the caller
	execServiceAction         chan int    // TBD
	execServiceActionResponse chan signal // TBD
}

// init starts the control.
// The control monitors the log service and can answer questions regarding the state of the service.
func init() {
	c.checkServiceState = make(chan int)
	c.checkServiceStateResponse = make(chan bool)
	c.setServiceState = make(chan int)
	c.execServiceAction = make(chan int)
	c.execServiceActionResponse = make(chan signal)

	controlRunning := make(chan bool)
	go c.run(controlRunning)
	if !<-controlRunning {
		panic("control is not running")
	}
}

// run represents the control service.
// This utility function runs in a dedicated goroutine and is started when the init function is implicitly called.
// It handles requests by listening on the following channels:
//   - execServiceAction
//   - setServiceState
//   - checkServiceState
func (c *control) run(controlRunning chan bool) {
	var newState, totalState int

	for {
		select {
		case controlRunning <- true:
		case action := <-c.execServiceAction:
			switch action {
			case start:
				// allocate log service channels
				buf, _ := strconv.Atoi(convertToString(s.attribute[logbuffer]))
				s.data = make(chan logMessage, buf)
				s.config = make(chan configMessage)
				s.stop = make(chan signal)
				s.confirmed = make(chan signal)

				// reset state attribute (after the log service has restarted)
				if totalState == stopped {
					totalState = 0
				}

				// start log service
				go s.run()
				// reply to the caller when the service has started
				go func() {
					for {
						// wait until the service is running
						if c.checkState(running) {
							break
						}
					}
					c.execServiceActionResponse <- signal{}
				}()
			case stop:
				// stop log service
				s.stop <- signal{}
				// reply to the caller when the service has stopped
				go func() {
					for {
						// wait until the service is stopped
						if c.checkState(stopped) {
							break
						}
					}
					c.execServiceActionResponse <- signal{}

					// cleanup log service resources
					s.fileDesc.Close()
				}()
			}
		case newState = <-c.setServiceState:
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
func (c *control) service(action int) signal {
	c.execServiceAction <- action
	return <-c.execServiceActionResponse
}

// checkState checks if the service has set the specified state.
func (c *control) checkState(state int) bool {
	c.checkServiceState <- state
	return <-c.checkServiceStateResponse
}

// setState sets the state of the log service.
func (c *control) setState(state int) {
	c.setServiceState <- state
}
