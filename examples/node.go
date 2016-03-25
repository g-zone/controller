package main

import (
	zmq "github.com/pebbe/zmq4"
	ctrl "github.com/g-zone/controller"
	ob "github.com/g-zone/tengine"
	"time"
	"fmt"
	"os"
)

func main() {
	if len(os.Args) !=2 {
		fmt.Printf("usage: %s <serverName>\n", os.Args[0])
		return
	}
	serverID := os.Args[1]


	//Setup data IN socket
	dataInSocket, _ := zmq.NewSocket(zmq.SUB)
	defer dataInSocket.Close()
	dataInSocket.SetSubscribe("")
	dataInSocket.Connect("tcp://localhost:23256")
	
	algorithm := &SampleAlgo{ 
		OrderBook    : ob.NewOrderBook(9364),
		HiPx         : -1,
		LowPx        : -1,
	}

	//Setup the reactor
	reactor := &ctrl.Reactor{ 
		ServerID     : serverID,
		DataInSocket : dataInSocket,
		CallBackApp  : algorithm, 
	}
	defer reactor.CleanUp()

	// Setup a command socket to kick/stop data feeder 
	// THIS IS NOT REQUIRED IN PROD VERSION OF THIS APP
	feederCmdSocket, _ := zmq.NewSocket(zmq.PUSH)
	defer feederCmdSocket.Close()
	feederCmdSocket.Connect("tcp://localhost:23257")
	feederCmdSocket.SendBytes([]byte(""), 0) //Signal the feeder to start 

	start := time.Now()

	//Start the reactor
	reactor.Run() 

	ellapsed := time.Now().Sub(start)

	//THIS IS NOT REQUIRED IN PROD VERSION OF THIS APP
	feederCmdSocket.SendBytes([]byte(""), 0) //Signal the feeder to pause

	fmt.Printf("Elapsed %.1f seconds. Processed %.2f K messages.\n" +
		   "Size of processed data %.2f KB\n" +
		   "Speed: %.2fK msg/msec ( %.2f usec/msg )\n", 
		   ellapsed.Seconds(), float64(reactor.TotalMsgCounter)/1e3, 
		   float64(reactor.TotalInputDataSize)/1024,
		   float64(reactor.TotalMsgCounter)/(ellapsed.Seconds()*1e3), (ellapsed.Seconds()*1e6)/float64(reactor.TotalMsgCounter) )
}

