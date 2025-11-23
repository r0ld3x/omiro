package main

import (
	"log"

	"github.com/gorilla/websocket"
)

type Client struct {
	ID      string
	Conn    *websocket.Conn
	Send    chan SendMessageType
	Partner *Client
}

type SendMessageType struct {
	Message []byte
	Type    int
}

func (c *Client) writePump() {
	defer func() {
		clientsMu.Lock()
		delete(clients, c.ID)
		clientsMu.Unlock()
		c.Conn.Close()
		log.Println("client disconnected:", c.ID)
	}()

	for msg := range c.Send {
		err := c.Conn.WriteMessage(msg.Type, msg.Message)
		if err != nil {
			log.Printf("error writing message to %s: %s\n", c.ID, err)
			return
		}
	}
}
