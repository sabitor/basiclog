package simplelog

import (
	"os"
	"strconv"
)

// control instance
var c = new(control)

// control is a structure used to control the log service and log service workflows.
type control struct {
	checkServiceState         chan int    // the channel for receiving a state check request from the caller
	checkServiceStateResponse chan bool   // the channel for sending a boolean response to the caller
	setServiceState           chan int    // the channel for receiving a state change request from the caller
	execServiceAction         chan int    // the channel for receiving a service action request from the caller
	execServiceActionResponse chan signal // the channel for sending a signal response to the caller to continue the workflow
	stopService               chan signal // the channel that signals any listening goroutine to stop
}

// init starts the control.
// The control monitors the log service and triggers actions to be started by the log service.
func init() {
	c.checkServiceState = make(chan int)
	c.checkServiceStateResponse = make(chan bool)
	c.setServiceState = make(chan int)
	c.execServiceAction = make(chan int)
	c.execServiceActionResponse = make(chan signal)

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
		case action := <-c.execServiceAction:
			switch action {
			case start:
				// init log service resources
				buf, _ := strconv.Atoi(convertToString(s.attribute[logbuffer]))
				s.logData = make(chan logMessage, buf)
				s.serviceConfig = make(chan configMessage)
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
					c.execServiceActionResponse <- signal{}
				}()
			case stop:
				archive, _ := strconv.ParseBool(convertToString(s.attribute[logarchive]))
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
					logFileName := s.fileDesc.Name()
					s.closeLogFile()
					if archive {
						s.archiveLogFile(logFileName)
					}
					c.execServiceActionResponse <- signal{}
				}()
			case initlog:
				append, _ := strconv.ParseBool(convertToString(s.attribute[appendlog]))
				logName := convertToString(s.attribute[logfilename])
				if !append {
					// don't append - remove old log
					var err error
					if _, err = os.Stat(logName); err == nil {
						if err = os.Remove(logName); err != nil {
							panic(err)
						}
					}
				}
				s.serviceConfig <- configMessage{initlog, logName}
			case switchlog:
				newLogName := convertToString(s.attribute[logfilename])
				if _, err := os.Stat(newLogName); err == nil {
					panic(m006)
				}
				s.serviceConfig <- configMessage{switchlog, newLogName}
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
