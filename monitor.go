package controller

import (
	zmq "github.com/pebbe/zmq4"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type PingInfo struct {
	lastUpdate  int64
	reported    bool  // true client is has timedout
}

type CmdHistory struct {
	sent, received int64
	cmdline string
	serverId string
	completed bool
}


var COMMANDS = map[string]int32 {
	"EXIT"   : 1,
	"PAUSE"  : 2,
	"RESUME" : 3,
        "CNTMSG" : 4,
	"PARAMS" : 5,
}

func HeartBeatMonitor(hostname string, cmdPort, pingPort, timeout int64) {

	if timeout < 1 {
		timeout = 5
	}

	monitoredClients := make(map[string]PingInfo)
	allHistory := make([]CmdHistory,0)

	heartBeatSocket, _ := zmq.NewSocket(zmq.SUB)
	defer heartBeatSocket.Close()
	connectString := fmt.Sprintf("tcp://*:%d", /* hostname, */ pingPort)
	fmt.Printf("Listening for hearbeats on port %s\n", connectString)
	heartBeatSocket.Bind(connectString) // XXX

	//  Subscribe to "PING " & "CMDR " messages
	heartBeatSocket.SetSubscribe("PING ")
	heartBeatSocket.SetSubscribe("CMDR ")
	
	commandInSocket, _ := zmq.NewSocket(zmq.PAIR)
	defer commandInSocket.Close()
	commandInSocket.Connect("inproc://commands")

	commandOutSocket, _ := zmq.NewSocket(zmq.PUB)
	defer commandOutSocket.Close()
	connectString = fmt.Sprintf("tcp://*:%d", /* hostname, */ cmdPort)
	fmt.Printf("Broadcasting commands on port %s\n", connectString)
	commandOutSocket.Bind(connectString)


	//  Initialize poll set
	poller := zmq.NewPoller()
	poller.Add(heartBeatSocket, zmq.POLLIN)
	poller.Add(commandInSocket, zmq.POLLIN)	

	//loop forever
	for {
		sockets, err := poller.PollAll(time.Duration(timeout) * time.Second)

		if err != nil {
			continue //  Interrupted
		}

		nowTime := time.Now()
		now := nowTime.Unix()

		if sockets[0].Events&zmq.POLLIN != 0  { //A new heart beat msg is available
			msg, _ := heartBeatSocket.Recv(0)
			msgs := strings.Fields(msg)
			serverId := msgs[1]
			
			if msg[0:5] == "PING " {
				lastUpdate, _ := strconv.ParseInt(msgs[2], 10, 64)
				
				if elem, found :=  monitoredClients[serverId] ; found == true {
					if lastUpdate < elem.lastUpdate {
						fmt.Printf("Client %s sent out of order PING msg (%d < %d)\n", serverId, lastUpdate, elem.lastUpdate)
					} else {
						elem.lastUpdate = lastUpdate
						monitoredClients[serverId] = elem
					}
				} else {
					fmt.Printf("Got new client: %s (@ %d)\n" , serverId, lastUpdate)
					monitoredClients[serverId] = PingInfo{lastUpdate, false}
				}
			} else { //"CMDR "
				fmt.Printf("\nReply from client %s: %s\n" , serverId, msg[5+len(serverId)+1:])
			}
		}

		for serverId, pingInfo := range monitoredClients {
			if now >= pingInfo.lastUpdate {
				if  now - pingInfo.lastUpdate > timeout {
					if pingInfo.reported == false {
						fmt.Printf("Client %s timed out (%d sec)\n", serverId, now - pingInfo.lastUpdate)
						pingInfo.reported = true
						monitoredClients[serverId] = pingInfo
					}
				} else {
					if pingInfo.reported == true {
						fmt.Printf("Client %s is back online\n", serverId)
						pingInfo.reported = false
						monitoredClients[serverId] = pingInfo
					}
				}
			} else {
				if pingInfo.lastUpdate - now > timeout {
					fmt.Printf("Clock skew for client %s : %d seconds\n", serverId, pingInfo.lastUpdate - now)
				}
			}
		}

		if sockets[1].Events&zmq.POLLIN != 0 { //A new user command just came in
			cmd, _ := commandInSocket.Recv(0)
			cmd = strings.TrimRight(cmd, "\n")
			cmds := strings.Fields(cmd)

			if len(cmds) == 1  {
				fmt.Println("\nusage: <host> <command> [ <params> ]\n\nrecognized commands:\n\texit\n\tpause\n\tresume\n\tcntmsg\n\tparams\n")
				continue
			}

			serverId := cmds[0]
			cmdId, found := COMMANDS[strings.ToUpper(cmds[1])]
			if found == false {
				fmt.Printf("Command (%s) NOT RECOGNIZED\n", cmds[1])
				continue
			}
			
			if serverId != "*" {
				if elem, found :=  monitoredClients[serverId] ; found == true {
					if elem.reported == false {
						thisCommand := CmdHistory{now, 0, cmd, serverId, false}
						allHistory = append(allHistory, thisCommand)
					} else {
						fmt.Printf("%s is down momentarilly; try again later\n", serverId)
						continue
					}
				} else {
					fmt.Printf("Server '%s' has never registered with this controller\n", serverId)
					continue
				}
			}
			rawCmd := fmt.Sprintf("CMD %s %02d %s", serverId, cmdId, strings.Join(cmds[2:], " "))
			commandOutSocket.Send(rawCmd, 0)
		}
	}
}
