package main

import (
	zmq "github.com/pebbe/zmq4"
	ctrl "github.com/g-zone/controller"
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
	dataInSocket, _ := zmq.NewSocket(zmq.PULL)
	defer dataInSocket.Close()
	dataInSocket.Connect("tcp://localhost:23256")
	
	//Setup a functor that we'll pass it into the reactor
	cmdSocket, _ := zmq.NewSocket(zmq.PUB)
	cmdSocket.Connect("tcp://localhost:50000")
	average := &ComputeAverage{ Socket:dataInSocket, CmdOutSocket:cmdSocket }

	start := time.Now()
	
	//Setup the reactor
	reactor := &ctrl.Reactor{ 
		ServerID : serverID , 
		MonitoredSockets : average , //optional
	}
	defer reactor.CleanUp()

	//Start the reactor
	reactor.Run() 

	ellapsed := time.Now().Sub(start)
	fmt.Printf("Ellapsed %.1f seconds. Processed %.2f K messages.\nSpeed %.2fK msg/msec ( %.2f usec/msg )\n", 
		   ellapsed.Seconds(), float64(reactor.TotalMsgCounter)/1e3, 
		   float64(reactor.TotalMsgCounter)/(ellapsed.Seconds()*1e3), (ellapsed.Seconds()*1e6)/float64(reactor.TotalMsgCounter) )
	fmt.Printf("Average = %.2f\n", average.Value())
}

