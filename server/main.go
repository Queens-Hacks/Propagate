package main

import (
	"math/rand"

	"github.com/Queens-Hacks/Propagate/sim"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

var twisty string = `
while 1 do
  grow("right")
  grow("up")
  grow("up")
end
`

var upUp string = `
while 1 do
  grow("up")
end
`

var fearnRight string = `
while 1 do
  if meta() == "" then
    grow("up")
    grow("up")
	split("right", "right")
  else
  	grow("right")
  end
end

`

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	total := make(chan []byte)
	diff := make(chan []byte)
	actions := make(chan sim.Action)

	s := sim.NewState(500, 125)

	species := s.AddSpecies(128, upUp, "Me")
	for i := 0; i < 20; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), 75}, species)
	}

	species = s.AddSpecies(44, fearnRight, "Me")
	for i := 0; i < 20; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), 75}, species)
	}

	species = s.AddSpecies(275, twisty, "Me")
	for i := 0; i < 20; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), 75}, species)
	}

	port := ":4444"

	logrus.Infof("Listening on port %s", port)
	go New(ctx, total, diff, actions, port)

	ss := s.StartSimulate(actions)
	for {
		ms := <-ss
		total <- ms.State
		diff <- ms.Diff
	}
}
