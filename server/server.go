package main

import (
	"net"

	"github.com/Sirupsen/logrus"

	"golang.org/x/net/context"
)

func New(ctx context.Context, total, diff chan []byte, port string) {

	var worldData []byte
	conns := make([]chan []byte, 0)
	newConns := make(chan net.Conn)

	go handleConnections(ctx, newConns, port)

	for {
		select {
		case data := <-total:
			worldData = data

		case data := <-diff:
			for _, c := range conns {
				c <- data
			}

		case conn := <-newConns:
			go sendWorld(conn, worldData)
			c := make(chan []byte)
			conns = append(conns, c)
			go sendDiffs(ctx, conn, c)
		}
	}
}

func handleConnections(ctx context.Context, newConns chan<- net.Conn, port string) {
	ln, err := net.Listen("tcp", port)
	if err != nil {
		logrus.Fatal(err)
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
			logrus.Error(err)
		}

		logrus.Infof("Accepted conn: %v", conn)

		go func() {
			newConns <- conn
		}()
	}
}

func sendWorld(conn net.Conn, data []byte) {
	_, err := conn.Write(data)
	if err != nil {
		logrus.Error(err)
	}
	logrus.Infof("Sent world to conn: %v", conn)
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
