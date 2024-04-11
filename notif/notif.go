package notif

import (
	"fmt"
	"math/rand"
	"net/http"
	"time"

	"golang.org/x/net/websocket"
)

type NotificationServer struct {
	subs map[*websocket.Conn]bool
}

func (s *NotificationServer) readLoop(ws *websocket.Conn) {
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

func NewNotificationServer() *NotificationServer {
	return &NotificationServer{
		subs: make(map[*websocket.Conn]bool),
	}
}

func StartNotificationServer(server *NotificationServer, port string) {
	http.Handle("/subscribe", websocket.Handler(server.handleSubscribe))
	http.Handle("/sub", websocket.Handler(server.handleNotify))
	http.ListenAndServe(port, nil)
}

func (s *NotificationServer) handleSubscribe(ws *websocket.Conn) {
	buff := make([]byte, 1024)
	for {
		n, err := ws.Read(buff)
		if err != nil {
			fmt.Println("Err: ", err)
			continue
		}

		msg := string(buff[:n])

		if s.subs[ws] {
			msg1 := fmt.Sprintf("This has data -> %d", rand.Intn(100))
			ws.Write([]byte(msg1))
			time.Sleep(time.Second * 2)
		} else {

			s.subs[ws] = true

			if msg == "sub" {
				ws.Write([]byte("Subscribed!"))
			}
		}
	}
}

func (s *NotificationServer) handleNotify(ws *websocket.Conn) {
	for {
		msg := fmt.Sprintf("This has data -> %d", rand.Intn(100))
		ws.Write([]byte(msg))
		time.Sleep(time.Second * 2)
	}
}
