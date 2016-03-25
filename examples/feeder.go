package main

import (
	zmq "github.com/pebbe/zmq4"
	"bufio"
	"fmt"
	"os"
)

var syncChannel chan int
var feederState int = 1     //1 - PAUSE , 0 - RUN

// readLines reads a whole file into memory
// and returns a slice of its lines.
type BYTEARR []byte
func readLines(path string) ([]BYTEARR, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	
	var lines []BYTEARR
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, []byte(scanner.Text()))
	}
	return lines, scanner.Err()
}



func PauseResume() {

	cmdSocket, _ := zmq.NewSocket(zmq.PULL)
	defer cmdSocket.Close()
	cmdSocket.Bind("tcp://*:23257")

	for { 
		cmdSocket.RecvBytes(0)
		feederState = 0
		syncChannel <- 0
		cmdSocket.RecvBytes(0)
		feederState = 1
	}
}

func main() {

	if len(os.Args) != 2 {
		fmt.Println("usage: feeder <inputfile>")
		return
	}

	dataSocket, _ := zmq.NewSocket(zmq.PUB)
	defer dataSocket.Close()
	dataSocket.Bind("tcp://*:23256")

	syncChannel = make(chan int)

	go PauseResume()


	fmt.Println("Reading input file ...")
	lines, err := readLines(os.Args[1])
	if err != nil {
		panic("Can not read input file")
	}
	fmt.Println("Done processing input file")

	var counter int
	for {
		if feederState == 1 {
			fmt.Println("Feeder is PAUSED")
			<- syncChannel
			fmt.Println("Feeder has RESUMED")
		}
		dataSocket.SendBytes(lines[counter][5:], 0)
		counter++
		if counter >= len(lines) { 
			counter = 0 // start over from the begining
		}
	}
}
