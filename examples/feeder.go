package main

import (
	zmq "github.com/pebbe/zmq4"
	"encoding/binary"
	"math/rand"
	"time"
)

func main() {
	dataSocket, _ := zmq.NewSocket(zmq.PUSH)
	defer dataSocket.Close()
	dataSocket.Bind("tcp://*:23256")
	rand.Seed(time.Now().Unix())
	buffer := make([]byte, 200)
	buffer[0] = 1 //simulate active data (as opposed to commands)
	for {
		binary.PutVarint(buffer[1:], rand.Int63n(1024))
		dataSocket.SendBytes(buffer, 0)
		//time.Sleep(time.Duration(5 * time.Microsecond))
	}
}
