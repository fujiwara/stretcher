package stretcher

import (
	"encoding/base64"
	"encoding/json"
	"io"
	"log"
)

type ConsulEvent struct {
	ID      string `json:"ID"`
	Name    string `json:"Name"`
	Payload string `json:"Payload"`
	LTime   int    `json:"LTime"`
}

func (ev ConsulEvent) PayloadString() string {
	raw, err := base64.StdEncoding.DecodeString(ev.Payload)
	if err != nil {
		return ""
	}
	return string(raw)
}

type ConsulEvents []ConsulEvent

func ParseConsulEvents(in io.Reader) (*ConsulEvent, error) {
	var evs ConsulEvents
	dec := json.NewDecoder(in)
	if err := dec.Decode(&evs); err != nil {
		return nil, err
	}
	if len(evs) == 0 {
		return nil, nil
	}
	ev := &evs[len(evs)-1]
	log.Println("Consul event ID:", ev.ID)
	return ev, nil
}
