package controller

import (
	zmq "github.com/pebbe/zmq4"
	"fmt"
	"time"
)


func Ping(controllerHost string, controllerPingPort int32, serverId string, interval int32)  {

	if interval < 1 {
		interval = 5
	}

	//Prepare the publisher
	publisher, _ := zmq.NewSocket(zmq.PUB)
	publisher.SetImmediate(true)
	publisher.SetConflate(true)
	defer publisher.Close()
	fmt.Printf("Pinging monitor app on tcp://%s:%d\n", controllerHost, controllerPingPort)
	publisher.Connect(fmt.Sprintf("tcp://%s:%d", controllerHost, controllerPingPort))

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
