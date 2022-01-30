package main

import (
	"sync"
)

var waitGroup = &sync.WaitGroup{}

func goWithWait(f func()) {
	go func() {
		waitGroup.Add(1)
		f()
		waitGroup.Done()
	}()
}

func main() {
	waitGroup.Add(1)
	go func() {
		InteractiveInput()
		waitGroup.Done()
	}()
	goWithWait(func() { Elevator(1, ElevatorParamA, FloorAvailableA) })
	// goWithWait(func() { Elevator(2, ElevatorParamB, FloorAvailableB) })
	// goWithWait(func() { Elevator(3, ElevatorParamC, FloorAvailableC) })
	waitGroup.Wait()
}
