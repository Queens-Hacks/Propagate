package main

import (
	"net"
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
			for _, c := range conns {
				c <- data
			}

		case wd := <-newConns:
			go sendWorld(wd, worldData)
			// c := make(chan []byte)
			// conns = append(conns, c)
			// go sendDiffs(ctx, conn, c)
		}
	}
}

type webSocketDone struct {
	ws *websocket.Conn
	c  chan struct{}
}

func handleWebSocket(ws *websocket.Conn) {
	logrus.Infof("Accepted conn: %v", ws)
	done := make(chan struct{})
	newConns <- webSocketDone{ws, done}
	<-done
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
	logrus.Infof("Checking for error", wd.ws)
	if err != nil {
		logrus.Error(err)
	}
	logrus.Infof("Sent world to conn: %v", wd.ws)
	close(wd.c)
}

func sendDiffs(ctx context.Context, conn net.Conn, diff chan []byte) {
	for {
		select {
		case data := <-diff:
			_, err := conn.Write(data)
			if err != nil {
				logrus.Error(err)
			}
			logrus.Infof("Sent diff to conn: %v", conn)
		case <-ctx.Done():
			return
		}
	}
}
