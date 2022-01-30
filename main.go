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
	StartElevator(1, ElevatorParamA, FloorAvailableA)
	StartElevator(2, ElevatorParamB, FloorAvailableB)
	StartElevator(3, ElevatorParamC, FloorAvailableC)
	waitGroup.Wait()
}
