package main

import (
	zmq "github.com/pebbe/zmq4"
	"encoding/binary"
	"fmt"
)

//
// This very simple class is an example of a functor
// that can be passed into the reactor. It has to implement
// the methods defined in MonitoredSocket interface (see reactor.go)
//

type ComputeAverage struct {

	//Public members
        Socket *zmq.Socket  //This is usually provided at construction time

	//Private members
	CmdOutSocket *zmq.Socket
	sum   int64
	count int64
}

func (this *ComputeAverage) GetDataInSocket() *zmq.Socket {
	return this.Socket
}

func (this *ComputeAverage) HandleData(data []byte) {
	if n, nBytes := binary.Varint(data[1:9]) ; nBytes>0 {
		this.sum += n
		this.count++
		/*
		if this.count == 1e6 {
			fmt.Println("Sending exit msg")
			this.CmdOutSocket.Send("\x00CMD * 01 ",0)
		}
		*/
	}
}

func (this *ComputeAverage) HandleCommand(cmd, params []byte) {
	fmt.Print("Got command " + string(cmd) + " : " + string(params) + "\n")
}

func (this *ComputeAverage) Value() float64 {
	return float64(this.sum)/float64(this.count)
}

