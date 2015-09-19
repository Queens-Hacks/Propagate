package main

import (
	"github.com/Queens-Hacks/Propagate/sim"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	total := make(chan []byte)
	diff := make(chan []byte)

	s := sim.SimpleState(1000, 500)
	for i := 0; i < 3; i++ {
		species := s.AddSpecies("ASD", "while 1 do grow(\"up\") end", "Me")
		s.AddPlant(sim.Location{250 + i*250, 250}, species)
	}

	port := ":4444"

	logrus.Infof("Listening on port %s", port)
	go New(ctx, total, diff, port)

	ss := s.StartSimulate()
	for {
		ms := <-ss
		total <- ms.State
		diff <- ms.Diff
	}
}
