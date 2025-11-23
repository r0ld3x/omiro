package main

import (
	"log"
	"time"

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
	ticker := time.NewTicker(30 * time.Second) // Ping every 30 seconds
	defer func() {
		ticker.Stop()
		clientsMu.Lock()
		delete(clients, c.ID)
		clientsMu.Unlock()
		c.Conn.Close()
		log.Println("client disconnected (writePump):", c.ID)
	}()

	for {
		select {
		case msg, ok := <-c.Send:
			if !ok {
				// Channel closed
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			err := c.Conn.WriteMessage(msg.Type, msg.Message)
			if err != nil {
				log.Printf("error writing message to %s: %s\n", c.ID, err)
				return
			}

		case <-ticker.C:
			// Send ping to keep connection alive
			log.Printf("📡 Sending ping to client %s\n", c.ID)
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("❌ Ping failed for %s: %s\n", c.ID, err)
				return
			}
		}
	}
}
