package main

import (
	"math/rand"

	"github.com/Queens-Hacks/Propagate/sim"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

var maxMemory string = `
while 1 do
	grow("up")
	grow("up")
	split("up","up")
	grow("left")
	grow("left")
	split("left","left")
	grow("right")
	grow("right")
	split("right","right")
end
`

var crystal string = `
while 1 do
  grow("up")
  split("up", "right")
  split("left", "left")
  split("right", "right")
end
`

var twistyRight string = `
while 1 do
  grow("right")
  grow("up")
  grow("up")
end
`

var twistyLeft string = `
while 1 do
  grow("left")
  grow("up")
  grow("up")
end
`

var twistyUp string = `
while 1 do
  grow("left")
  grow("up")
  grow("up")
  grow("left")
  grow("up")
  grow("up")
  grow("right")
  grow("up")
  grow("up")
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
    grow("up")
end

`

var fearnLeft string = `
while 1 do
  if meta() == "" then
    grow("up")
    grow("up")
	split("left", "left")
  else
  	grow("left")
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

	species := s.AddSpecies(256, maxMemory, "Me")
	for i := 0; i < 10; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), 75}, species)
	}

	species = s.AddSpecies(63, crystal, "Me")
	for i := 0; i < 10; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), 75}, species)
	}

	species = s.AddSpecies(128, upUp, "Me")
	for i := 0; i < 10; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), 75}, species)
	}

	species = s.AddSpecies(44, fearnRight, "Me")
	for i := 0; i < 5; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), 75}, species)
	}

	species = s.AddSpecies(14, fearnLeft, "Me")
	for i := 0; i < 5; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), 75}, species)
	}

	species = s.AddSpecies(275, twistyUp, "Me")
	for i := 0; i < 10; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), 75}, species)
	}

	species = s.AddSpecies(233, twistyLeft, "Me")
	for i := 0; i < 10; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), 75}, species)
	}

	species = s.AddSpecies(333, twistyRight, "Me")
	for i := 0; i < 10; i++ {
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
