package main

import (
	"io"
	"net/http"

	"github.com/Sirupsen/logrus"

	"code.google.com/p/go.net/websocket"
	"golang.org/x/net/context"
)

var newConns = make(chan webSocketDone)

func New(ctx context.Context, total, diff chan []byte, port string) {

	var worldData []byte
	conns := make([]chan []byte, 0)

	go handleConnections(port)

	for {
		select {
		case data := <-total:
			worldData = data

		case data := <-diff:
			logrus.Info("sending diff to all clients")
			for _, c := range conns {
				c <- data
			}

		case wd := <-newConns:
			go sendWorld(wd, worldData)
			c := make(chan []byte)
			conns = append(conns, c)
			logrus.Info("starting diff loop")
			go func() { sendDiffs(ctx, wd, c); close(wd.done) }()
		}
	}
}

type webSocketDone struct {
	ws   *websocket.Conn
	done chan struct{}
	ctx  context.Context
}

func handleWebSocket(ws *websocket.Conn) {
	logrus.Infof("Accepted conn: %v", ws)
	done := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	newConns <- webSocketDone{ws, done, ctx}
	var b []byte
	err := websocket.Message.Receive(ws, b)
	if err == io.EOF {
		cancel()
	} else if err != nil {
		logrus.Error(err)
	}
	<-done
	logrus.Infof("Closing websocket: %v", ws)
}

func handleConnections(port string) {
	http.Handle("/", websocket.Handler(handleWebSocket))
	err := http.ListenAndServe(port, nil)
	if err != nil {
		logrus.Error(err)
	}
}

func sendWorld(wd webSocketDone, data []byte) {
	err := websocket.Message.Send(wd.ws, data)
	if err != nil {
		logrus.Error(err)
	}
	logrus.Infof("Sent world to conn: %v", wd.ws)
}

func sendDiffs(ctx context.Context, wd webSocketDone, diff chan []byte) {
	for {
		select {
		case data := <-diff:
			logrus.Info("client being sent diff")
			err := websocket.Message.Send(wd.ws, data)
			if err != nil {
				logrus.Error(err)
			}
			logrus.Infof("Sent diff to conn: %v", wd.ws)

		case <-ctx.Done():
			return

		case <-wd.ctx.Done():
			return
		}
	}
}
