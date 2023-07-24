package simplelog

import (
	"strconv"
)

// control instance
var c = new(control)

// control represents an control object used to operate the log service and log service workflows.
type control struct {
	startServiceTask          chan configMessage // receiver of a service task request from the caller
	startServiceTaskResponse  chan signal        // sender of a signal response to the caller to continue the workflow
	checkServiceState         chan int           // receiver of state check request from the caller
	checkServiceStateResponse chan bool          // sender of a boolean response to the caller
	setServiceState           chan int           // receiver of a state change request from the caller
	stopService               chan signal        // to broadcast any listening goroutine to stop
}

// init starts the control.
// The control monitors the log service and triggers actions to be started by the log service.
func init() {
	c.startServiceTask = make(chan configMessage)
	c.startServiceTaskResponse = make(chan signal)
	c.checkServiceState = make(chan int)
	c.checkServiceStateResponse = make(chan bool)
	c.setServiceState = make(chan int)

	controlRunning := make(chan bool)
	go c.run(controlRunning)
	if !<-controlRunning {
		panic(m000)
	}
}

// run represents the control service.
// This utility function runs in a dedicated goroutine and is started when the init function is implicitly called.
// It handles requests by listening on the following channels:
//   - execServiceAction
//   - setServiceState
//   - checkServiceState
func (c *control) run(controlRunning chan bool) {
	var singleState, totalState int

	// service loop
	for {
		select {
		case controlRunning <- true:
		case cfgMsg := <-c.startServiceTask:
			switch cfgMsg.task {
			case start:
				// init log service resources
				buf, _ := strconv.Atoi(cfgMsg.data[logbuffer])
				s.logData = make(chan logMessage, buf)
				s.configService = make(chan configMessage)
				s.configServiceResponse = make(chan error)
				c.stopService = make(chan signal)

				// reset state attribute (after the log service has restarted)
				if totalState == stopped {
					totalState = 0
				}

				// start log service
				go s.run()
				// reply to the caller when the service has started
				// Hint: The go routine is necessary to prevent a deadlock; control must still be able to handle setServiceState messages
				go func() {
					for {
						// wait until the service is running
						if c.checkState(running) {
							break
						}
					}
					s.isUp = true
					c.startServiceTaskResponse <- signal{}
				}()
			case stop:
				archive, _ := strconv.ParseBool(cfgMsg.data[logarchive])
				// stop log service
				close(c.stopService) // closing the stopService channel sends a signal to all go routines which are listening to that channel
				// reply to the caller when the service has stopped
				// Hint: The go routine is necessary to prevent a deadlock; control must still be able to handle setServiceState messages
				go func() {
					for {
						// wait until the service is stopped
						if c.checkState(stopped) {
							break
						}
					}
					s.releaseFileLogger(archive)
					c.startServiceTaskResponse <- signal{}
				}()
			}
		case singleState = <-c.setServiceState:
			if singleState == stopped {
				// unset all other states
				totalState = singleState
			} else {
				// add new state to the total state
				totalState |= singleState
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

// checkState checks if the service has set the specified state.
func (c *control) checkState(state int) bool {
	c.checkServiceState <- state
	return <-c.checkServiceStateResponse
}

// setState sets the state of the log service.
func (c *control) setState(state int) {
	c.setServiceState <- state
}
