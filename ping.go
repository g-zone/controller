package controller

import (
	zmq "github.com/pebbe/zmq4"
	"fmt"
	"time"
)


func Ping(pingHost string, pingPort int32, serverId string, interval int32)  {

	if interval < 1 {
		interval = 5
	}

	//Prepare the publisher
	publisher, _ := zmq.NewSocket(zmq.PUB)
	publisher.SetImmediate(true)
	publisher.SetConflate(true)
	defer publisher.Close()
	connectString := fmt.Sprintf("tcp://*:%d", /* pingHost, */ pingPort) 
	fmt.Printf("Publishing PING messages on port %d\n", pingPort)
	publisher.Bind(connectString) // XXX

	msg := fmt.Sprintf("PING %s", serverId)

	// loop forever
	for {
		nowTime := time.Now()
		now := nowTime.Unix()
		rawMsg := fmt.Sprintf("%s %d", msg, now)
		publisher.Send(rawMsg, 0)
		time.Sleep(time.Duration(interval) * time.Second)
	}
}
