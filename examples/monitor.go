package main

import (
	zmq "github.com/pebbe/zmq4"
	ctrl "github.com/g-zone/controller"
	"bufio"
	"fmt"
	"os"
)

func main() {

	reader := bufio.NewReader(os.Stdin)
	command, _ := zmq.NewSocket(zmq.PAIR)
	defer command.Close()
	command.Bind("inproc://commands")

	go ctrl.HeartBeatMonitor(ctrl.DEFAULT_CONTROLLER_HOST, ctrl.DEFAULT_COMMAND_PORT, ctrl.DEFAULT_PING_PORT, ctrl.DEFAULT_PING_TIMEOUT)

	for {
		fmt.Print("Enter command > ")
		text, _ := reader.ReadString('\n')
		if text == "\n" {
			continue
		}
		if text == "quit\n" {
			break 
		}
		command.Send(text,0)
	}
}
