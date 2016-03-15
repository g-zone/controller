package main

import (
	"encoding/binary"
	"fmt"
	ctrl "github.com/g-zone/controller"
	zmq "github.com/pebbe/zmq4"
	"math/rand"
	"os"
	"time"
)

func main() {
	if len(os.Args) != 2 {
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

	//Setup a functor that we'll pass it into the reactor
	average := &ComputeAverage{Socket: dataInSocket}

	start := time.Now()

	//Setup the reactor
	reactor := &ctrl.Reactor{
		ServerID:         serverID,
		MonitoredSockets: []ctrl.MonitoredSocket{average}, //optional
	}
	defer reactor.CleanUp()

	//Start the reactor
	reactor.Run()

	ellapsed := time.Now().Sub(start)
	fmt.Printf("Ellapsed %.1f seconds. Processed %.2f K messages.\nSpeed %.2fK msg/msec ( %.2f usec/msg )\n",
		ellapsed.Seconds(), float64(reactor.TotalMsgCounter)/1024,
		float64(reactor.TotalMsgCounter)/(ellapsed.Seconds()*1e3), (ellapsed.Seconds()*1e6)/float64(reactor.TotalMsgCounter))

	fmt.Printf("Average = %.2f\n", average.Value())
}

func FakeInputDataFeeder() {
	dataSocket, _ := zmq.NewSocket(zmq.PAIR)
	defer dataSocket.Close()
	dataSocket.Connect("inproc://data")
	rand.Seed(23)
	buffer := make([]byte, 64)
	for {
		i := binary.PutVarint(buffer, rand.Int63n(1024))
		dataSocket.SendBytes(buffer[0:i], 0)
		time.Sleep(time.Duration(10 * time.Millisecond))
	}
}
