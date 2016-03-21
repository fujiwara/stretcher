package stretcher

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
)

type ConsulEvent struct {
	ID      string `json:"ID"`
	Name    string `json:"Name"`
	Payload []byte `json:"Payload"`
	LTime   int    `json:"LTime"`
}

type ConsulEvents []ConsulEvent

func ParseConsulEvents(in io.Reader) (*ConsulEvent, error) {
	var evs ConsulEvents
	dec := json.NewDecoder(in)
	if err := dec.Decode(&evs); err != nil {
		return nil, err
	}
	if len(evs) == 0 {
		return nil, fmt.Errorf("No Consul events found")
	}
	ev := &evs[len(evs)-1]
	log.Println("Consul event ID:", ev.ID)
	return ev, nil
}
