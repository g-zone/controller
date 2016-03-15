package controller

import (
	"fmt"
	zmq "github.com/pebbe/zmq4"
)

const (
	EXIT   = "01"
	PAUSE  = "02"
	RESUME = "03"
	CNTMSG = "04"
	PARAMS = "05"
)

type MonitoredSocket interface {
	GetDataInSocket() *zmq.Socket
	HandleData(data []byte)
	HandleCommand(cmd, params []byte)
}

type Reactor struct {
	ServerID string

	CommandInSocket *zmq.Socket

	CommandOutSocket  *zmq.Socket
	ControllerCmdPort int32
	ControllerHost    string

	PingPort    int32
	PingTimeout int32

	MonitoredSockets []MonitoredSocket

	privateCommandInSocket, privateCommandOutSocket bool
	TotalMsgCounter, TotalInputDataSize             uint64
}

func (reactor *Reactor) initDefault() *Reactor {
	if reactor.ServerID == "" {
		fmt.Println("ERROR: missing ServerID")
		return nil
	}

	for i := 0; i < len(reactor.MonitoredSockets); i++ {
		if reactor.MonitoredSockets[i].GetDataInSocket() == nil {
			fmt.Printf("ERROR: uninitialized DataInSocket member #%d\n", i)
			return nil
		}
	}

	if reactor.CommandInSocket == nil {
		if reactor.ControllerHost == "" {
			reactor.ControllerHost = DEFAULT_CONTROLLER_HOST
		}
		if reactor.ControllerCmdPort <= 0 {
			reactor.ControllerCmdPort = DEFAULT_COMMAND_PORT
		}
		reactor.CommandInSocket, _ = zmq.NewSocket(zmq.SUB)
		connectString := fmt.Sprintf("tcp://%s:%d", reactor.ControllerHost, reactor.ControllerCmdPort)
		fmt.Printf("Subscribing for commands from %s on 'CMD %s ' and 'CMD * ' topics\n", connectString, reactor.ServerID)
		reactor.CommandInSocket.Connect(connectString)
		reactor.CommandInSocket.SetSubscribe("CMD " + reactor.ServerID + " ")
		reactor.CommandInSocket.SetSubscribe("CMD * ")
		reactor.privateCommandInSocket = true
	}

	if reactor.CommandOutSocket == nil {
		if reactor.ControllerHost == "" {
			reactor.ControllerHost = DEFAULT_CONTROLLER_HOST
		}
		if reactor.PingPort <= 0 {
			reactor.PingPort = DEFAULT_PING_PORT
		}
		reactor.CommandOutSocket, _ = zmq.NewSocket(zmq.PUB)
		reactor.CommandOutSocket.SetConflate(true)
		reactor.CommandOutSocket.Connect(fmt.Sprintf("tcp://%s:%d", reactor.ControllerHost, reactor.PingPort)) //XXX
		reactor.privateCommandOutSocket = true
	}

	if reactor.PingPort <= 0 {
		reactor.PingPort = DEFAULT_PING_PORT
	}

	if reactor.PingTimeout <= 0 {
		reactor.PingTimeout = DEFAULT_PING_TIMEOUT
	}

	return reactor
}

func (reactor *Reactor) CleanUp() {
	if reactor.privateCommandInSocket {
		reactor.CommandInSocket.Close()
	}
	if reactor.privateCommandOutSocket {
		reactor.CommandOutSocket.Close()
	}
}

func (reactor *Reactor) Run() {

	if reactor.initDefault() == nil {
		return
	}

	go Ping(reactor.ControllerHost, reactor.PingPort, reactor.ServerID, reactor.PingTimeout)

	// Initialize poll set
	poller := zmq.NewPoller()
	poller.Add(reactor.CommandInSocket, zmq.POLLIN)
	for i := 0; i < len(reactor.MonitoredSockets); i++ {
		poller.Add(reactor.MonitoredSockets[i].GetDataInSocket(), zmq.POLLIN)
	}

	bPaused := false
	bExitMainLoop := false
	cmdPrefixLength := len(reactor.ServerID) + 5
	msgOut := ""
	padding := int(0)

	for {
		//Wait (=block) until one of the sockets in the Poller is available for reading
		sockets, err := poller.PollAll(-1)

		if err != nil {
			continue //Interrupted; resume waiting
		}

		//
		// First check if a command just came in
		//

		if sockets[0].Events&zmq.POLLIN != 0 { //A new command is available on CommandInSocket

			cmd, _ := reactor.CommandInSocket.RecvBytes(0)

			if cmd[4] == '*' {
				padding = 6
			} else {
				padding = cmdPrefixLength
			}

			switch string(cmd[padding : padding+2]) {
			case EXIT:
				bExitMainLoop = true
				msgOut = fmt.Sprintf("CMDR %s EXIT", reactor.ServerID)
			case PAUSE:
				bPaused = true
				msgOut = fmt.Sprintf("CMDR %s PAUSED", reactor.ServerID)
			case RESUME:
				bPaused = false
				msgOut = fmt.Sprintf("CMDR %s RESUMED", reactor.ServerID)
			case CNTMSG:
				msgOut = fmt.Sprintf("CMDR %s PROCESSED %d msgs totaling %d bytes", reactor.ServerID, reactor.TotalMsgCounter, reactor.TotalInputDataSize)
			case PARAMS:
				msgOut = fmt.Sprintf("CMDR %s PARAMS", reactor.ServerID)
			}
			for i := 0; i < len(reactor.MonitoredSockets); i++ {
				reactor.MonitoredSockets[i].HandleCommand(cmd[padding:padding+2], cmd[padding+2:])
			}
			reactor.CommandOutSocket.Send(msgOut, 0)
		}

		if bExitMainLoop == true {
			break
		}

		//
		// Second, check if data is available on any of the data in monitored sockets
		//

		for i := 0; i < len(reactor.MonitoredSockets); i++ {
			if sockets[1+i].Events&zmq.POLLIN != 0 { //New data package is available on DataInSocket
				data, _ := reactor.MonitoredSockets[i].GetDataInSocket().RecvBytes(0)
				if bPaused == false {
					reactor.MonitoredSockets[i].HandleData(data)
					reactor.TotalInputDataSize += uint64(len(data))
					reactor.TotalMsgCounter++
				}
			}
		}
	}
}
