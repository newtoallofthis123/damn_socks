package main

import (
	"fmt"

	"github.com/newtoallofthis123/damn_socks/notif"
	"golang.org/x/net/websocket"
)

type Server struct {
	conns map[*websocket.Conn]bool
}

func (s *Server) handleServer(ws *websocket.Conn) {
	fmt.Printf("Recieved incoming connection from %s\n", ws.RemoteAddr())

	s.conns[ws] = true

	s.readLoop(ws)
}

func (s *Server) readLoop(ws *websocket.Conn) {
	buff := make([]byte, 1024)
	for {
		n, err := ws.Read(buff)
		if err != nil {
			fmt.Println("Err: ", err)
			continue
		}

		msg := buff[:n]
		fmt.Println(string(msg))

		ws.Write([]byte("Hello from a ws server"))
	}
}

func (s *Server) broadcast(msg []byte) {
	for ws, stat := range s.conns {
		if stat {
			go func(c *websocket.Conn) {
				if _, err := c.Write(msg); err != nil {
					fmt.Println("Write Error: ", err)
				}
			}(ws)
		}
	}
}

func NewServer() *Server {
	return &Server{
		conns: make(map[*websocket.Conn]bool),
	}
}

func main() {
	fmt.Println("Listening on port :1234")

	server := notif.NewNotificationServer()
	notif.StartNotificationServer(server, ":1234")
}
