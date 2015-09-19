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

	s := sim.SimpleState(100, 50)
	species := s.AddSpecies("ASD", "while 1 do grow(\"up\") end", "Me")
	s.AddPlant(sim.Location{50, 25}, species)

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
