package main

import (
	"github.com/Queens-Hacks/propagate/sim"
	"github.com/Sirupsen/logrus"
	"golang.org/x/net/context"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	total := make(chan []byte)
	diff := make(chan []byte)

	s := sim.SimpleState(50, 50)
	data, err := sim.MarshalState(s)
	if err != nil {
		logrus.Fatal(err)
	}

	go func() {
		total <- data
	}()

	port := ":4444"

	logrus.Infof("Listening on port %s", port)
	New(ctx, total, diff, port)
}
