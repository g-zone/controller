package controller

import (
	zmq "github.com/pebbe/zmq4"
	"fmt"
)

const ( EXIT   = "01"
	PAUSE  = "02"
        RESUME = "03"
        CNTMSG = "04"
        PARAMS = "05" )



type MonitoredSocket struct {
	DataInSocket  *zmq.Socket
	DataCallBack  func (data []byte)
}

type Reactor struct {
	ServerID          string
	LocalHost         string

	CommandInSocket   *zmq.Socket
	LocalCmdInPort     int32

	CommandOutSocket  *zmq.Socket
	LocalCmdOutPort    int32
	

	PingPort          int32
	PingTimeout       int32

	MonitoredSockets  []MonitoredSocket

	CommandCallBack   func(cmd, params []byte)

	privateCommandInSocket, privateCommandOutSocket bool
}

func (reactor *Reactor) initDefault() *Reactor {
	if reactor.ServerID == "" {
		fmt.Println("ERROR: missing ServerID")
		return nil
	}

	for i:=0 ; i<len(reactor.MonitoredSockets) ; i++ {
		if reactor.MonitoredSockets[i].DataInSocket == nil {
			fmt.Printf("ERROR: uninitialized DataInSocket member #%d\n", i)
			return nil
		}
	}

	if reactor.LocalHost == "" { 
		reactor.LocalHost = DEFAULT_LOCAL_HOST
	}

	if reactor.CommandInSocket == nil {
		if reactor.LocalCmdInPort <= 0 {
			reactor.LocalCmdInPort = DEFAULT_COMMAND_IN_PORT
		}
		reactor.CommandInSocket, _ = zmq.NewSocket(zmq.SUB)
		connectString := fmt.Sprintf("tcp://*:%d", /* reactor.LocalHost,*/  reactor.LocalCmdInPort)
		fmt.Printf("Listening for commands on %s \n", connectString)
		reactor.CommandInSocket.Bind(connectString) //XXX
		reactor.CommandInSocket.SetSubscribe("CMD " + reactor.ServerID + " ")
		reactor.CommandInSocket.SetSubscribe("CMD * ")
		reactor.privateCommandInSocket = true
	}

	if reactor.CommandOutSocket == nil {
		if reactor.LocalCmdOutPort <= 0 {
			reactor.LocalCmdOutPort = DEFAULT_COMMAND_OUT_PORT
		}
		reactor.CommandOutSocket, _ = zmq.NewSocket(zmq.PUB)
		reactor.CommandOutSocket.SetConflate(true)
		connectString := fmt.Sprintf("tcp://*:%d", /* reactor.LocalHost, */ reactor.LocalCmdOutPort) 
		fmt.Printf("Publishing REPLY CMD  messages on port %d\n", reactor.LocalCmdOutPort)
		reactor.CommandOutSocket.Bind(connectString) // XXX
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

	go Ping( reactor.LocalHost, reactor.PingPort, reactor.ServerID, reactor.PingTimeout )

	// Initialize poll set
	poller := zmq.NewPoller()
	poller.Add(reactor.CommandInSocket, zmq.POLLIN)
	for i:=0 ; i<len(reactor.MonitoredSockets) ; i++ {
		poller.Add(reactor.MonitoredSockets[i].DataInSocket, zmq.POLLIN)
	}

	bPaused := false
	bExitMainLoop := false
	msgCounter := uint64(0)
	totalInputDataSize := uint64(0)
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

			if cmd [4] == '*' {
				padding = 6
			} else {
				padding = cmdPrefixLength
			}

			switch string(cmd[padding : padding+2]) {
			case EXIT:
				bExitMainLoop = true
				msgOut = fmt.Sprintf("CMDR %s EXIT" , reactor.ServerID)
			case PAUSE:
				bPaused = true
				msgOut = fmt.Sprintf("CMDR %s PAUSED" , reactor.ServerID)
			case RESUME:
				bPaused = false
				msgOut = fmt.Sprintf("CMDR %s RESUMED" , reactor.ServerID)
			case CNTMSG:
				msgOut = fmt.Sprintf("CMDR %s PROCESSED %d msgs totaling %d bytes" , reactor.ServerID, msgCounter, totalInputDataSize)
			case PARAMS:
				msgOut = fmt.Sprintf("CMDR %s PARAMS",  reactor.ServerID) 
			}
			if reactor.CommandCallBack != nil {
				reactor.CommandCallBack( cmd[padding : padding+2], cmd[padding+2:] )
			}
			reactor.CommandOutSocket.Send(msgOut, 0)
		} 

		if bExitMainLoop == true {
			break
		}

		//
		// Second, check if data is available on any of the data in monitored sockets
		//

		for i:=0 ; i<len(reactor.MonitoredSockets) ; i++ {
			if sockets[1+i].Events&zmq.POLLIN != 0 { //New data package is available on DataInSocket
				data, _ := reactor.MonitoredSockets[i].DataInSocket.RecvBytes(0)
				if bPaused == false {
					if reactor.MonitoredSockets[i].DataCallBack != nil {
						reactor.MonitoredSockets[i].DataCallBack( data )
					}
					totalInputDataSize += uint64(len(data))
					msgCounter++
				}
			}
		}
	}
}
