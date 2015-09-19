package main

import (
	"net"
	"net/http"

	"github.com/Sirupsen/logrus"

	"code.google.com/p/go.net/websocket"
	"golang.org/x/net/context"
)

var newConns = make(chan *websocket.Conn)

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

		case ws := <-newConns:
			go sendWorld(ws, worldData)
			// c := make(chan []byte)
			// conns = append(conns, c)
			// go sendDiffs(ctx, conn, c)
		}
	}
}

func handleWebSocket(conn *websocket.Conn) {
	logrus.Infof("Accepted conn: %v", conn)
	newConns <- conn
}

func handleConnections(port string) {
	http.Handle("/", websocket.Handler(handleWebSocket))
	err := http.ListenAndServe(port, nil)
	if err != nil {
		logrus.Error(err)
	}
}

func sendWorld(ws *websocket.Conn, data []byte) {
	err := websocket.Message.Send(ws, data)
	logrus.Infof("Checking for error", ws)
	if err != nil {
		logrus.Error(err)
	}
	logrus.Infof("Sent world to conn: %v", ws)
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
