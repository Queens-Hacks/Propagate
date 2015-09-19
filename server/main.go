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
	data, err := sim.MarshalGameState(s)
	if err != nil {
		logrus.Fatal(err)
	}

	go func() {
		total <- data
	}()

	port := ":4444"

	logrus.Infof("Listening on port %s", port)
	go New(ctx, total, diff, port)

	for {
		updateState(&s, sim.DirtTile)
		updateState(&s, sim.AirTile)
	}
}

func updateState(s *sim.State, t sim.TileType) {
	for x := 40; x < 60; x++ {
		for y := 20; y < 25; y++ {
			s.SetTile(sim.Location{x, y}, sim.Tile{T: t})
		}
	}
}
