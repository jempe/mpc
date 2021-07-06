package mpcremote

import (
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

type remote struct {
	forward chan message
	join    chan *client
	leave   chan *client
	clients map[*client]bool
}

func NewRemote() *remote {
	return &remote{
		forward: make(chan message),
		join:    make(chan *client),
		leave:   make(chan *client),
		clients: make(map[*client]bool),
	}
}

func (r *remote) Run() {
	for {
		select {
		case client := <-r.join:
			// joining
			r.clients[client] = true
			log.Println("New remote joined")
		case client := <-r.leave:
			// leaving
			delete(r.clients, client)
			close(client.send)
			log.Println("Remote Control left")
		case message := <-r.forward:
			msg := message.msg
			log.Println("Message received: ", string(msg), "Player ID:", message.senderId)
			// forward message to all clients
			for client := range r.clients {
				client.send <- msg
				log.Println(" -- sent to Player")
			}
		}
	}
}

const (
	socketBufferSize  = 1024
	messageBufferSize = 256
)

var upgrader = &websocket.Upgrader{ReadBufferSize: socketBufferSize, WriteBufferSize: socketBufferSize}

func (r *remote) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	socket, err := upgrader.Upgrade(w, req, nil)
	if err != nil {
		log.Fatal("ServeHTTP:", err)
		return
	}

	for cl, status := range r.clients {
		log.Println(cl.playerId, status)
	}

	client := &client{
		socket:   socket,
		send:     make(chan []byte, messageBufferSize),
		remote:   r,
		playerId: req.URL.Query().Get("player_id"),
		remoteId: req.URL.Query().Get("id"),
	}
	r.join <- client
	defer func() { r.leave <- client }()
	go client.write()
	client.read()
}

type client struct {
	socket   *websocket.Conn
	send     chan []byte
	remote   *remote
	isPlayer bool
	remoteId string
	playerId string
}

type message struct {
	msg      []byte
	senderId string
}

func (c *client) read() {
	defer c.socket.Close()
	for {
		_, msg, err := c.socket.ReadMessage()
		if err != nil {
			return
		}
		c.remote.forward <- message{msg: msg, senderId: c.playerId}
	}
}

func (c *client) write() {
	defer c.socket.Close()
	for msg := range c.send {
		err := c.socket.WriteMessage(websocket.TextMessage, msg)
		if err != nil {
			return
		}
	}
}
