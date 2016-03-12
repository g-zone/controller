package main

import (
	zmq "github.com/pebbe/zmq4"
	ctrl "github.com/g-zone/controller"
	"strconv"
	"strings"
	"bufio"
	"fmt"
	"os"
)

func atoi(s string) int64 {
	if v, e := strconv.Atoi(s) ; e == nil {
		return 0
	} else {
		return int64(v)
	}
}
		
	

func main() {

	reader := bufio.NewReader(os.Stdin)
	command, _ := zmq.NewSocket(zmq.PAIR)
	defer command.Close()
	command.Bind("inproc://commands")

	hostMap := make(map[string]ctrl.Host)

	for i:=1 ; i<len(os.Args) ; i++ {
		token := strings.Split(os.Args[i], ":")
		if len(token) != 3 {
			fmt.Printf("Ignoring entry %d : '%s'", i, os.Args[i])
			continue
		}
		host, found := hostMap[token[0]]
		if found == false {
			host = *new(ctrl.Host)
		}
		switch strings.ToUpper(token[1]) {
		case "HOST"      : host.Host       = token[2]
		case "PINGPORT"  : host.PingPort   = atoi(token[2])
		case "CMDOUTPORT": host.CmdOutPort = atoi(token[2])
		case "CMDINPORT" : host.CmdInPort  = atoi(token[2])
		default: fmt.Printf("Urecognized (ignored) option: '%s'\n", token[1])
		}
		hostMap[token[0]] = host
	}

	// Convert map to slice
	hosts := make([]ctrl.Host, 0, len(hostMap))
	for _ , host := range hostMap {
		hosts = append(hosts, host)
	}

	go ctrl.HeartBeatMonitor(hosts, ctrl.DEFAULT_PING_TIMEOUT)

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
