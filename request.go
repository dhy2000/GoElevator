package main

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
)

// PassengerRequest [id]-FROM-[from]-TO-[to]
type PassengerRequest struct {
	id   int
	from int
	to   int
}

// AddElevatorRequest ADD-[id]-[elevatorType]
type AddElevatorRequest struct {
	id           int
	elevatorType string
}

var (
	passengerPattern   = regexp.MustCompile(`^(\d+)-FROM-(\d+)-TO-(\d+)$`)
	addElevatorPattern = regexp.MustCompile(`^ADD-(\d+)-(\w+)$`)
)

/* record passenger/elevator ids to detect duplication */
var passengers = make(map[int]bool)
var elevators = map[int]string{1: "A", 2: "B", 3: "C"} // elevator 1, 2, 3 is initialized

func parsePassengerRequest(input string) (*PassengerRequest, error) {
	match := passengerPattern.FindStringSubmatch(input)
	if len(match) != 4 {
		return nil, errors.New("wrong PassengerRequest format")
	}
	id, errId := strconv.Atoi(match[1])
	from, errFrom := strconv.Atoi(match[2])
	to, errTo := strconv.Atoi(match[3])
	if errId != nil || errFrom != nil || errTo != nil {
		return nil, errors.New("failed parseInt on PassengerRequest")
	}
	request := &PassengerRequest{id: id, from: from, to: to}
	if _, has := passengers[id]; has {
		return request, fmt.Errorf("duplicated passenger: %v", id)
	}
	if from < MinFloor || from > MaxFloor {
		return request, fmt.Errorf("illegal from floor: %v", from)
	}
	if to < MinFloor || to > MaxFloor {
		return request, fmt.Errorf("illegal to floor: %v", to)
	}
	passengers[id] = true
	return request, nil
}

func parseAddElevatorRequest(input string) (*AddElevatorRequest, error) {
	match := addElevatorPattern.FindStringSubmatch(input)
	if len(match) != 3 {
		return nil, errors.New("wrong AddElevatorRequest format")
	}
	id, errId := strconv.Atoi(match[1])
	elevatorType := match[2]
	if errId != nil {
		return nil, errors.New("failed parseInt on AddElevatorRequest")
	}
	if _, has := ElevatorTypes[elevatorType]; !has {
		return nil, fmt.Errorf("illegal elevator type %v", elevatorType)
	}
	if _, has := elevators[id]; has {
		return nil, fmt.Errorf("elevator %v already exists", id)
	}
	elevators[id] = elevatorType
	return &AddElevatorRequest{id: id, elevatorType: elevatorType}, nil
}
