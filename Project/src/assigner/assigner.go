// Package liftAssigner gathers the cost values of the lifts on the network,
// and assigns the best candidate to each order.

// Recive costs from elevators and compare

package assigner

import (
	def "definitions"
	"log"
	"queue"
	"time"
)

type reply struct {
	cost int
	elevator string
}
type request struct {
	floor  int
	button int
	timer  *time.Timer
}

var NumOnlineCh = make(chan int)

func CollectCosts(costReply chan def.Message, numOnlineCh chan int){
	requestMap := make( map[request][]reply)
	var timeout = make(chan *request)
	var numOnline int = 1
	for{
		select{
		case message := <-costReply:
			newRequest := request{floor: message.Floor,button: message.Button}
			newReply := reply{cost: message.Cost, elevator: message.Addr}
			log.Println(def.ColR, "New Cost incomming from: ", message.Addr, " for cost: ", message.Cost,def.ColN)

			// Compare requests on content, without the timer
			for existingRequest := range requestMap{
				if equal(existingRequest, newRequest){
					newRequest = existingRequest
				}
			}

			// Check if request is in queue
			if replyList, exist := requestMap[newRequest]; exist {
				// Check if newReply already is registered.
				found := false
				for _, reply := range replyList {
					if reply == newReply {
						found = true
					}
				}
				// Add to list if not found
				if !found {
					requestMap[newRequest] = append(requestMap[newRequest], newReply)
					newRequest.timer.Reset(def.CostReplyTimeoutDuration)
				}
			} else {
				// If order not in queue at all, init order list with it
				newRequest.timer = time.NewTimer(def.CostReplyTimeoutDuration)
				requestMap[newRequest] = []reply{newReply}
				go costTimer(&newRequest, timeout)
			}
			chooseBestElevator(requestMap,numOnline,false)
		case numOnlineUpdate := <- numOnlineCh:
			numOnline = numOnlineUpdate
				log.Println(def.ColR,"Number online in assignement: ",numOnline,def.ColN)
		case <- timeout:
			log.Println(def.ColR,"Not all costs received in time!",def.ColN)
			chooseBestElevator(requestMap,numOnline,true)
		}
	}
}

func chooseBestElevator(requestMap map[request][]reply, numOnline int, isTimeout bool){
	var bestElevator string

	// Go through list of requests and find the best elevator in each replyList
	for request,replyList := range requestMap{
		log.Println(def.ColR,"Num online: ",numOnline,def.ColN)
		log.Println(def.ColR,"Num reply: ",len(replyList),def.ColN)
		if len(replyList) == numOnline || isTimeout{
			log.Println(def.ColB,"All costs are collected, or reply timed out: ", isTimeout)
			lowestCost := 10000
			for _,reply := range replyList{
				if reply.cost < lowestCost{
					lowestCost = reply.cost
					bestElevator = reply.elevator
				}else if reply.cost == lowestCost{
					if reply.elevator < bestElevator{
						bestElevator = reply.elevator
					}
				}
			}
			log.Println(def.ColB,"Will now add request to Floor: ",request.floor,", Button: ",request.button,", to Elevator: ",bestElevator, def.ColN )
			queue.AddRequest(request.floor, request.button, bestElevator)
			request.timer.Stop()
			delete(requestMap, request)
		}
	}
}

func equal(r1,r2 request)bool{
	return r1.floor == r2.floor && r1.button == r2.button
}

func costTimer(newRequest *request, timeout chan<- *request) {
	<-newRequest.timer.C
	timeout <- newRequest
}
