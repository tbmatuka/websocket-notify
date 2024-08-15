package websocket_notify

import (
	"net/http"

	"github.com/gorilla/websocket"
)

type SocketConnection struct {
	Id      uint
	Managed bool
	Conn    *websocket.Conn
	Channel chan []byte
	Tags    map[string]bool
}

type WebsocketMessage struct {
	Tags        []string `json:"tags"`
	Signature   string   `json:"signature"`
	Unsubscribe bool     `json:"unsubscribe"`
}

var upgrader = websocket.Upgrader{} //nolint:gochecknoglobals

func handleWebsocketRequest(responseWriter http.ResponseWriter, request *http.Request) {
	conn, err := upgrader.Upgrade(responseWriter, request, nil)
	if err != nil {
		return
	}

	socketConnection := SocketConnection{
		Managed: false,
		Conn:    conn,
		Channel: make(chan []byte),
		Tags:    make(map[string]bool),
	}

	go socketConnection.ListenForEvents()

	manager := getSubscriptionManager()
	for {
		msg := WebsocketMessage{}
		err = conn.ReadJSON(&msg)
		if err != nil {
			socketConnection.Close()

			return
		}

		if msg.Unsubscribe {
			manager.Unsubscribe(&socketConnection, msg.Tags)
		} else {
			err := manager.Subscribe(&socketConnection, msg.Tags, msg.Signature)
			if err != nil {
				socketConnection.Close()

				return
			}
		}
	}
}

func (connection *SocketConnection) ListenForEvents() {
	for {
		msg := <-connection.Channel

		if len(msg) == 0 {
			return
		}

		err := connection.Conn.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			connection.Close()

			return
		}
	}
}

func (connection *SocketConnection) Close() {
	if connection.Managed {
		manager := getSubscriptionManager()
		manager.CloseConnection(connection)
	}

	connection.Channel <- []byte(``)

	_ = connection.Conn.Close()
}
