package main

import (
	"math/rand"

	"github.com/Queens-Hacks/Propagate/sim"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	total := make(chan []byte)
	diff := make(chan []byte)

	s := sim.NewState(500, 125)
	species := s.AddSpecies(128, `
while 1 do
  if meta() == "" then
    grow("up")
    grow("up")
    split("right", "right")
  else
    debug(meta())
    grow("right")
  end
end
`, "Me")
	for i := 0; i < 1; i++ {
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
