package main

import (
	zmq "github.com/pebbe/zmq4"
	ctrl "github.com/g-zone/controller"
	"time"
	"fmt"
	"os"
)

var counter int = 0

func main() {
	if len(os.Args) !=2 {
		fmt.Printf("usage: %s <serverName>\n", os.Args[0])
		return
	}
	serverID := os.Args[1]

	//Setup data IN socket
	dataInSocket, _ := zmq.NewSocket(zmq.PAIR)
	defer dataInSocket.Close()
	dataInSocket.Bind("inproc://data")
	
	//Start the fake input data feeder
	go FakeInputDataFeeder()

	start := time.Now()
	
	//Setup the reactor
	reactor := &ctrl.Reactor{ 
		ServerID : serverID , 
		MonitoredSockets : []ctrl.MonitoredSocket{ {dataInSocket,  processData} }, //optional
		CommandCallBack : processCommands,                                         //optional
	}
	defer reactor.CleanUp()

	//Start the reactor
	reactor.Run() 

	ellapsed := time.Now().Sub(start)
	fmt.Printf("Ellapsed %.1f seconds. Processed %.2f K messages.\nSpeed %.2fK msg/msec ( %.2f usec/msg )\n", 
		   ellapsed.Seconds(), float64(counter)/1024, 
		   float64(counter)/(ellapsed.Seconds()*1e3), (ellapsed.Seconds()*1e6)/float64(counter) )
}

func processCommands(cmd, params []byte) {
	fmt.Print("Got command " + string(cmd) + " : " + string(params) + "\n")
}

func processData(msgIn []byte) {
	//Add here processing code of each message
	counter++
}

func FakeInputDataFeeder() {
	dataSocket, _ := zmq.NewSocket(zmq.PAIR)
	defer dataSocket.Close()
	dataSocket.Connect("inproc://data")
	for {
		dataSocket.Send("blah", 0)
		//time.Sleep(time.Duration(1 * time.Microsecond))
	}
}
