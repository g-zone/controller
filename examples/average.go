package main

//
// This very simple class is an example of a functor
// that can be passed into the reactor. It has to implement
// the methods defined in AppCallBack interface (see reactor.go)
//

import (
	zmq "github.com/pebbe/zmq4"  // Required by SingalReactorToExit(); REMOVE IN PROD
	ob "github.com/g-zone/tengine"
	"strings"
	"strconv"
	"fmt"
)

type SampleAlgo struct {
	count int64

	OrderBook *ob.OrderBook //This is provided/initialised by the caller
	HiPx, LowPx int
	vwap, typicalPx float32
	currQty, currTotQty, prevTotQty int
}

func (this *SampleAlgo) HandleCommand(cmd, params []byte) {
	fmt.Print("Got command " + string(cmd) + " : " + string(params) + "\n")
}

func (this *SampleAlgo) HandleData(data []byte) {

	this.count++
	// Exit after we have processed enough messages
	if this.count == 1e6 {  
		this.OrderBook.Dump()
		fmt.Println("Sending exit msg")
		SingalReactorToExit()
	}

	//return


	//build order book, using level updates of QB/QS
	//keep running calc of this.vwap when either Turnover happens or Market order comes in
	//check teh VL and the Turnover to see if a trade was done.
	//if the calced this.vwap hits the triggers compared to best bid/best ask on ob, make a trade.
	//internal message
	//ticker    Level update
	//Order came in
	//B/Sell
	//price/qty
	//test := "PX:1037,TV:143000,TO:147548000|MOB:0,SIDE:BUY,PX:1037,QTY:0|MOB:1,SIDE:BUY,PX:1036,QTY:4000|"
	s := strings.Split(string(data), "|") //split by messages
	for i := range s {
		v := strings.Split(s[i], ":")[0]
		//fmt.Println(v)
		switch v {
		case "PX":
			t := strings.Split(s[i], ",")
			px, tv := t[0], t[1]
			price, _ := strconv.Atoi(strings.Split(px, ":")[1])
			if tv != ""{
				this.currTotQty, _ = strconv.Atoi(strings.Split(tv, ":")[1])
			} else{
			continue}
			//fmt.Println( px + " " + tv)
			this.currQty = this.currTotQty - this.prevTotQty
			this.prevTotQty = this.currTotQty
			if this.HiPx == -1 {
				this.HiPx, this.LowPx = price, price
			}
			if (price > this.HiPx) && (price != 0) {
				this.HiPx = price
			}
			if (price < this.LowPx) && (price != 0) {
				this.LowPx = price
			}
			this.typicalPx += float32((this.HiPx + this.LowPx + price) / 3 * this.currQty)
			this.vwap = this.typicalPx / float32(this.currTotQty)
			
		case "MOB":
			t := strings.Split(s[i], ",")
			
			oSide := strings.Split(t[1], ":")[1]
			oPx, _ := strconv.Atoi(strings.Split(t[2], ":")[1])
			oQty, _ := strconv.Atoi(strings.Split(t[3], ":")[1])
			oMob, _ := strconv.Atoi(strings.Split(t[0],":")[1])
			if oSide == "S" {
			
				this.OrderBook.UpdateLevels(ob.SELL, int32(oPx), int32(oQty), int32(oMob))
			} else {
				this.OrderBook.UpdateLevels(ob.BUY, int32(oPx), int32(oQty), int32(oMob))
			}

			//build the orderbook with level updates
			//		MinAsk  int32
			//MaxAsk	int32  //depth
			//MaxBid  int32
			//MinBid  int32  //Depth
			//this.OrderBook.Dump()
		}

		//if this.vwap > ob.lowoffer put a buy order in
		//if this.vwap < ob.bestbid  put a sell order in
	}
}


/* +--------------------------------------+
   | REMOVE FROM PROD VERSION OF THIS APP |
   +--------------------------------------+ */
var cmdSocket *zmq.Socket = initCmdSocket()

func SingalReactorToExit() {
	cmdSocket.Send("\x00CMD * 01 ",0)
}

func initCmdSocket() *zmq.Socket {
	cmdSocket, _ := zmq.NewSocket(zmq.PUB)
	cmdSocket.Connect("tcp://localhost:50000")
	return cmdSocket
}
