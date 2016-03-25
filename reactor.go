package controller

import (
	zmq "github.com/pebbe/zmq4"
	"fmt"
)

const ( 
	EXIT   = "01"
	PAUSE  = "02"
        RESUME = "03"
        CNTMSG = "04"
        PARAMS = "05" )



type AppCallBack interface {
	HandleData(data []byte)
	HandleCommand(cmd, params []byte)
}

type Reactor struct {
	ServerID          string
	LocalHost         string

	DataInSocket      *zmq.Socket

	CommandInSocket   *zmq.Socket
	LocalCmdInPort     int32

	CommandOutSocket  *zmq.Socket
	LocalCmdOutPort    int32
	

	PingPort          int32
	PingTimeout       int32

	CallBackApp       AppCallBack

	privateCommandInSocket, privateCommandOutSocket bool
	TotalMsgCounter, TotalInputDataSize  uint64
}

func (reactor *Reactor) initDefault() *Reactor {
	if reactor.ServerID == "" {
		fmt.Println("ERROR: missing ServerID")
		return nil
	}

	if reactor.CallBackApp != nil {
		if reactor.DataInSocket == nil {
			fmt.Println("ERROR: uninitialized DataInSocket member")
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
		reactor.CommandInSocket.SetSubscribe("\x00CMD " + reactor.ServerID + " ")
		reactor.CommandInSocket.SetSubscribe("\x00CMD * ")
		reactor.privateCommandInSocket = true
	}

	if reactor.CommandOutSocket == nil {
		if reactor.LocalCmdOutPort <= 0 {
			reactor.LocalCmdOutPort = DEFAULT_COMMAND_OUT_PORT
		}
		reactor.CommandOutSocket, _ = zmq.NewSocket(zmq.PUB)
		reactor.CommandOutSocket.SetConflate(true)
		connectString := fmt.Sprintf("tcp://%s:%d", reactor.LocalHost, reactor.LocalCmdOutPort) 
		fmt.Printf("Publishing REPLY CMD  messages on port %d\n", reactor.LocalCmdOutPort)
		reactor.CommandOutSocket.Bind(connectString)
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


func (reactor *Reactor) receiveCommands(dataSocketType zmq.Type) {
	
	socket, _ := zmq.NewSocket(dataSocketType)
	defer socket.Close()
	socket.Bind("inproc://commands")
	for {
		cmd, _ := reactor.CommandInSocket.RecvBytes(0)
		socket.SendBytes(cmd, 0)
	}
}


func (reactor *Reactor) Run() {

	if reactor.initDefault() == nil {
		return
	}

	go Ping( reactor.LocalHost, reactor.PingPort, reactor.ServerID, reactor.PingTimeout )

	var dataSocketType zmq.Type

	dataSocketType, _ = reactor.DataInSocket.GetType()
	switch dataSocketType {
	case zmq.SUB  : 
		dataSocketType = zmq.PUB
		fmt.Println("Using PUB socket type for sending internal commands")
		reactor.DataInSocket.SetSubscribe("\x00CMD ")
	case zmq.PULL : 
		fmt.Println("Using PUSH socket type for sending internal commands")
		dataSocketType = zmq.PUSH
	default: 
		panic("Unsupported input data socket type")
	}

	go reactor.receiveCommands(dataSocketType)

	bPaused := false
	bExitMainLoop := false
	cmdPrefixLength := len(reactor.ServerID) + 6
	msgOut := ""
	padding := int(0)
	reactor.DataInSocket.Connect("inproc://commands")

	for {
		//Read in messages/commands
		msg, _ := reactor.DataInSocket.RecvBytes(0)

		if len(msg) == 0 {
			continue
		}

		if msg[0] == 0 {
			//
			// COMMANDS
			//
			if msg[5] == '*' {
				padding = 7
			} else {
				padding = cmdPrefixLength
			}
			switch string(msg[padding : padding+2]) {
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
				msgOut = fmt.Sprintf("CMDR %s PROCESSED %d msgs totaling %d bytes" , reactor.ServerID, reactor.TotalMsgCounter, reactor.TotalInputDataSize)
			case PARAMS:
				msgOut = fmt.Sprintf("CMDR %s PARAMS",  reactor.ServerID) 
			}
			(reactor.CallBackApp).HandleCommand(msg[padding : padding+2], msg[padding+2:])
			reactor.CommandOutSocket.Send(msgOut, 0)
		} else {
			//
			// REGULAR MESSAGES
			//
			if bPaused == false {
				(reactor.CallBackApp).HandleData( msg )
				reactor.TotalInputDataSize += uint64(len(msg))
				reactor.TotalMsgCounter++
			}
		}

		if bExitMainLoop == true {
			break
		}
	}
}
