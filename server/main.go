package main

import (
	"math/rand"

	"github.com/Queens-Hacks/Propagate/sim"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

var scripty string = `
function randDir()
	return "up"
  -- d = math.random(3)
  -- if d == 0 then return "right" end
  -- if d == 1 then return "up" end
  -- if d == 2 then return "left" end
end
 
while 1 do
  grow(randDir())
  grow(randDir())
  split(randDir(), "right")
end
`

var other_scripty string = `
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

	s := sim.NewState(500, 125)
	species := s.AddSpecies(128, other_scripty, "Me")
	for i := 0; i < 100; i++ {
		s.AddSpore(sim.Location{rand.Intn(500), 75}, species)
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
