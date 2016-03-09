package stretcher

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
)

type ConsulEvent struct {
	ID      string  `json:"ID"`
	Name    string  `json:"Name"`
	Payload Payload `json:"Payload"`
	LTime   int     `json:"LTime"`
}

type ConsulEvents []ConsulEvent

type Payload struct {
	string
}

func (p Payload) String() string {
	return p.string
}

func (p *Payload) UnmarshalText(src []byte) error {
	b := make([]byte, len(src))
	n, err := base64.StdEncoding.Decode(b, src)
	if err != nil {
		return err
	}
	p.string = string(b[0:n])
	return nil
}

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
