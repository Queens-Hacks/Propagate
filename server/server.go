package main

import (
	"io"
	"net/http"

	"github.com/Sirupsen/logrus"

	"code.google.com/p/go.net/websocket"
	"github.com/Queens-Hacks/Propagate/sim"
	"golang.org/x/net/context"
)

type webSocketDone struct {
	ws   *websocket.Conn
	done chan struct{}
	ctx  context.Context
	id   int
}

func New(ctx context.Context, total, diff chan []byte, actions chan sim.Action, port string) {
	var worldData []byte
	nextId := 0

	newConns := make(chan webSocketDone)
	conns := map[int]chan []byte{}

	handleWebSocket := func(ws *websocket.Conn) {
		done := make(chan struct{})
		ctx, cancel := context.WithCancel(context.Background())
		newConns <- webSocketDone{ws, done, ctx, 0}

		for {
			var a sim.Action
			err := websocket.JSON.Receive(ws, &a)
			if err == io.EOF {
				cancel()
				break
			} else if err != nil {
				logrus.Error(err)
				break
			}

			logrus.Infof("Found a kind: %s", a.Kind)
			actions <- a
		}
		<-done
		logrus.Infof("Closing websocket: %v", ws)
	}

	handleLocalWebSocket := func(ws *websocket.Conn) {
		s := sim.NewState(500, 125)
		actions := make(chan sim.Action)

		species := s.AddSpecies(275, "", "Me")
		s.AddSpore(sim.Location{0, 100}, species)

		ss := s.StartSimulate(actions)
		defer s.StopSimulate()

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			// First send the starting state
			ms := <-ss
			err := websocket.Message.Send(ws, ms.State)
			if err != nil {
				logrus.Error(err)
			}

			// Then send diffs
			for {
				select {
				case ms := <-ss:
					err := websocket.Message.Send(ws, ms.Diff)
					if err != nil {
						logrus.Error(err)
					}
					continue
				case <-ctx.Done():
				}
				break
			}
		}()

		for {
			var a sim.Action
			err := websocket.JSON.Receive(ws, &a)
			if err == io.EOF {
				cancel()
				break
			} else if err != nil {
				logrus.Error(err)
				break
			}

			logrus.Infof("Found a kind: %s", a.Kind)
			actions <- a
		}
	}

	go func() {
		http.Handle("/global", websocket.Handler(handleWebSocket))
		http.Handle("/local", websocket.Handler(handleLocalWebSocket))
		// XXX SUPER SKETCH
		http.Handle("/", http.FileServer(http.Dir("../client")))
		err := http.ListenAndServe(port, nil)
		if err != nil {
			logrus.Error(err)
		}
	}()

	for {
		select {
		case data := <-total:
			worldData = data

		case data := <-diff:
			// logrus.Infof("sending diff to %d clients with total %d", len(conns), nextId)
			for key, c := range conns {
				logrus.Infof("sending DIFF to %d", key)
				go func(b chan []byte) { b <- data }(c)
			}

		case wd := <-newConns:
			sendWorld(wd, worldData)
			c := make(chan []byte, 100)
			conns[nextId] = c
			id := nextId
			wd.id = id
			nextId++
			go func() {
				sendDiffs(ctx, wd, c)
				close(wd.done)
				delete(conns, id)
			}()
		}
	}
}
func sendWorld(wd webSocketDone, data []byte) {
	err := websocket.Message.Send(wd.ws, data)
	if err != nil {
		logrus.Error(err)
	}
	// logrus.Infof("Sent world to conn %d", wd.id)
}

func sendDiffs(ctx context.Context, wd webSocketDone, diff chan []byte) {
	for {
		select {
		case data := <-diff:
			// logrus.Infof("conn %d is waiting for diff", wd.id)
			err := websocket.Message.Send(wd.ws, data)
			if err != nil {
				logrus.Error(err)
			}
			// logrus.Infof("Sent diff to conn %d", wd.id)

		case <-ctx.Done():
			return

		case <-wd.ctx.Done():
			return
		}
	}
}
