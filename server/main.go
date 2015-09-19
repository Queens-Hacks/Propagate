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

	s := sim.NewState(500, 250)
	species := s.AddSpecies(128, "while 1 do grow(\"up\") end", "Me")
	for i := 0; i < 3; i++ {
		s.AddPlant(sim.Location{200 + i*50, 250}, species)
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
