package rpc

import (
	"net/http"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{}

type wsServer struct{}

func (s *wsServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Warnf("failed to update connection to websockets: %s", err)
		return
	}

	defer conn.Close()

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Warnf("failed to read websockets message: %s", err)
			break
		}

		log.Debugf("received message over websockets: %s", message)
		err = conn.WriteMessage(messageType, message)
		if err != nil {
			log.Warnf("failed to send websockets message: %s", err)
			break
		}
	}
}
