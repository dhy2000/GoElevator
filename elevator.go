package main

import (
	"container/list"
	"fmt"
	"sync"
	"time"
)

const (
	MinFloor = 1
	MaxFloor = 20
)

var ElevatorTypes = map[string]bool{"A": true, "B": true, "C": true}

type ElevatorParam struct {
	elevatorType   string
	openDelay      int
	closeDelay     int
	moveDelay      int
	passengerLimit int
}

var elevatorLock = sync.Mutex{}

var elevatorSignals = make([]chan bool, 0)

// enum direction
const (
	UP   = 1
	DOWN = -1
)

/* parameters for each elevator type */
var (
	ElevatorParamA = &ElevatorParam{
		elevatorType:   "A",
		openDelay:      200,
		closeDelay:     200,
		moveDelay:      600,
		passengerLimit: 8,
	}
	ElevatorParamB = &ElevatorParam{
		elevatorType:   "B",
		openDelay:      200,
		closeDelay:     200,
		moveDelay:      400,
		passengerLimit: 6,
	}
	ElevatorParamC = &ElevatorParam{
		elevatorType:   "C",
		openDelay:      200,
		closeDelay:     200,
		moveDelay:      200,
		passengerLimit: 4,
	}
)

var (
	FloorAvailableA = func(n int) bool { return n >= MinFloor && n <= MaxFloor }
	FloorAvailableB = func(n int) bool { return n >= MinFloor && n <= MaxFloor && n%2 == 0 }
	FloorAvailableC = func(n int) bool { return (n >= 1 && n <= 3) || (n >= 18 && n <= 20) }
)

func DirectionSame(request *PassengerRequest, direction int) bool {
	if request.to > request.from && direction == UP {
		return true
	}
	if request.to < request.from && direction == DOWN {
		return true
	}
	return false
}

func Elevator(id int, param *ElevatorParam, floorAvailable func(int) bool, signalChan <-chan bool) {
	pos := 1
	direction := UP
	open := false
	passenger := list.List{}

	var lastOpenMilli int64 = 0
	ensureOpen := func(o bool) {
		if open == o {
			return
		}
		if o == true {
			lastOpenMilli = time.Now().UnixMilli()
			TimedPrintln(fmt.Sprintf("OPEN-%v-%v", pos, id))
		} else {
			nowMilli := time.Now().UnixMilli()
			if remain := int64(param.openDelay+param.closeDelay) - (nowMilli - lastOpenMilli); remain > 0 {
				time.Sleep(time.Duration(remain) * time.Millisecond)
			}
			TimedPrintln(fmt.Sprintf("CLOSE-%v-%v", pos, id))
		}
		open = o
	}

	// one step
	for {
		if floorAvailable(pos) {
			/* 1. Release passenger off */
			if passenger.Len() > 0 {
				offElements := list.List{}
				for e := passenger.Front(); e != nil; e = e.Next() {
					if r, ok := e.Value.(*LinkedRequest); ok && r.request.to == pos {
						offElements.PushBack(e)
					}
				}
				for e := offElements.Front(); e != nil; e = e.Next() {
					elem, _ := e.Value.(*list.Element)
					r, _ := elem.Value.(*LinkedRequest)
					passenger.Remove(elem)
					ensureOpen(true)
					TimedPrintln(fmt.Sprintf("OUT-%v-%v-%v", r.request.id, pos, id))
					if r.next != nil {
						PutRequest(r.next)
					} else {
						finishSignals <- SIGNAL
					}
				}
			}
			/* 2. Pick passenger up */
			for passenger.Len() < param.passengerLimit {
				r := FetchOneRequest(pos, floorAvailable, direction)
				if r == nil {
					break
				}
				ensureOpen(true)
				TimedPrintln(fmt.Sprintf("IN-%v-%v-%v", r.request.id, pos, id))
				passenger.PushBack(r)
			}
			ensureOpen(false)
		}
		/* 3. Move, Turn or Suspend */
		turn := func() bool {
			return passenger.Len() == 0 && !HasRequest(pos, floorAvailable, direction)
		}
		// Turn: elevator is empty, no same-direction requests
		if turn() {
			direction = -direction
		} else {
			// Move
			if next := pos + direction; next >= MinFloor && next <= MaxFloor {
				pos += direction
				time.Sleep(time.Duration(param.moveDelay) * time.Millisecond)
				TimedPrintln(fmt.Sprintf("ARRIVE-%v-%v", pos, id))
			}
			continue
		}
		// Suspend: turn and turn back
		if turn() {
			signal := <-signalChan
			if !signal {
				return // this elevator should terminate
			}
		}
	}
}

func StartElevator(id int, param *ElevatorParam, floorAvailable func(int) bool) {
	signal := make(chan bool, 1000)
	elevatorLock.Lock()
	defer elevatorLock.Unlock()
	elevatorSignals = append(elevatorSignals, signal)
	goWithWait(func() { Elevator(id, param, floorAvailable, signal) })
}

func NotifyAllElevator(signal bool) {
	elevatorLock.Lock()
	defer elevatorLock.Unlock()
	for _, signalChan := range elevatorSignals {
		signalChan <- signal
	}
}
