package simplelog

// control instance
var c = new(control)

// control is a structure to control and handle the communication with the log service.
type control struct {
	checkServiceState         chan int    // the channel for receiving a state check request from the caller
	checkServiceStateResponse chan bool   // the channel for sending a boolean response to the caller
	serviceState              chan int    // the channel for receiving a state change request from the caller
	resetControl              chan signal // the channel for receiving a control reset request from the caller
}

// init starts the control.
// The control monitors the log service and can answer questions regarding the state of the service.
func init() {
	c.checkServiceState = make(chan int, 1)
	c.checkServiceStateResponse = make(chan bool, 1)
	c.serviceState = make(chan int)
	c.resetControl = make(chan signal)

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
	var newState int
	totalState := newState

	for {
		select {
		case controlRunning <- true:
		case <-c.resetControl:
			newState = 0
			totalState = newState
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

// getCheckServiceStateChan returns the checkServiceState channel
func (c *control) checkServiceStateChan() chan int {
	return c.checkServiceState
}

// getCheckServiceStateResponseChan returns the checkServiceStateResponse channel
func (c *control) checkServiceStateResponseChan() chan bool {
	return c.checkServiceStateResponse
}

// setServiceStateChan returns the serviceState channel
func (c *control) setServiceStateChan() chan int {
	return c.serviceState
}
