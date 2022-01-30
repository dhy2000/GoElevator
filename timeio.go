package main

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"
)

var startMilli = time.Now().UnixMilli()

var printLock = sync.Mutex{}

// TimedPrintln Print with timestamp prefix
func TimedPrintln(s string) {
	printLock.Lock()
	defer printLock.Unlock()
	nowMilli := time.Now().UnixMilli()
	secs := float64(nowMilli-startMilli) / 1000
	fmt.Printf("[%9.4f]%s\n", secs, s)
}

// handleInput is a goroutine
func handleInput(in <-chan string) {
	for {
		s, ok := <-in
		if !ok {
			break
		}
		// try parse PassengerRequest and AddElevatorRequest
		passengerRequest, errPassenger := parsePassengerRequest(s)
		addElevatorRequest, errAddElevator := parseAddElevatorRequest(s)
		if passengerRequest != nil {
			if errPassenger != nil {
				_, _ = os.Stderr.WriteString(fmt.Sprintf("Bad PassengerRequest %s: %s\n", s, errPassenger))
				continue
			}
			PutRequest(ProcessRequest(passengerRequest))
			continue
		}
		if addElevatorRequest != nil {
			if errAddElevator != nil {
				_, _ = os.Stderr.WriteString(fmt.Sprintf("Bad AddElevatorRequest %s: %s\n", s, errAddElevator))
				continue
			}
			// add elevator
			AddElevatorCount()
			switch addElevatorRequest.elevatorType {
			case "A":
				StartElevator(addElevatorRequest.id, ElevatorParamA, FloorAvailableA)
			case "B":
				StartElevator(addElevatorRequest.id, ElevatorParamB, FloorAvailableB)
			case "C":
				StartElevator(addElevatorRequest.id, ElevatorParamC, FloorAvailableC)
			}
			continue
		}
		_, _ = os.Stderr.WriteString(fmt.Sprintf("Bad Request: %s\n", s))
	}
	Terminate()
}

func InteractiveInput() {
	fmt.Println(">>> Start .......")
	ch := make(chan string)
	var s string
	goWithWait(func() { handleInput(ch) })
	for {
		_, err := fmt.Scanln(&s)
		if err == io.EOF {
			break
		}
		ch <- s
	}
	close(ch)
}
