package main

import (
	"container/list"
	"sync"
)

type LinkedRequest struct {
	request *PassengerRequest // current stage
	next    *LinkedRequest    // next stage
}

var waitList [1 + MaxFloor]list.List

var lock = sync.RWMutex{}

var (
	elevatorCount = 3 // synchronized
	comingCount   = 0
)

// elevatorSignals
const (
	SIGNAL    = true
	TERMINATE = false
)

var finishSignals = make(chan bool, 1000)

// ProcessRequest Deal with transition and put into wait list
func ProcessRequest(request *PassengerRequest) *LinkedRequest {
	lock.Lock()
	defer lock.Unlock()
	comingCount++
	// TODO: split requests
	linkedRequest := &LinkedRequest{
		request: request,
		next:    nil,
	}
	return linkedRequest
}

func PutRequest(request *LinkedRequest) {
	lock.Lock()
	defer lock.Unlock()
	waitList[request.request.from].PushBack(request)
	NotifyAllElevator(SIGNAL)
}

// FetchOneRequest trying to fetch one request at floor, if there is no request available at this floor, return nil
// floor: the number of the floor the elevator is at
// isAvailableFloor: identifies type of the elevator
func FetchOneRequest(floor int, floorAvailable func(int) bool, direction int) *LinkedRequest {
	lock.Lock()
	defer lock.Unlock()
	q := &waitList[floor]
	for e := q.Front(); e != nil; e = e.Next() {
		if request, ok := e.Value.(*LinkedRequest); ok {
			if floorAvailable(request.request.to) && DirectionSame(request.request, direction) {
				// this elevator can carry him
				return q.Remove(e).(*LinkedRequest)
			}
		}
	}
	return nil
}

// HasRequest check whether there is requests to be done at the direction
func HasRequest(floor int, floorAvailable func(int) bool, direction int) bool {
	lock.RLock()
	defer lock.RUnlock()
	for pos := floor + direction; pos >= MinFloor && pos <= MaxFloor; pos += direction {
		q := &waitList[pos]
		for e := q.Front(); e != nil; e = e.Next() {
			if request, ok := e.Value.(*LinkedRequest); ok {
				if floorAvailable(request.request.to) {
					return true
				}
			}
		}
	}
	return false
}

func AddElevatorCount() {
	lock.Lock()
	defer lock.Unlock()
	elevatorCount++
}

// Terminate this is a goroutine
func Terminate() {
	for comingCount > 0 {
		<-finishSignals
		comingCount--
	}
	NotifyAllElevator(TERMINATE)
}
