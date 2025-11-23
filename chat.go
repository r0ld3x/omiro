package main

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

func handleChat(c *Client, data json.RawMessage) {
	var payload struct {
		Message string `json:"message"`
	}

	// Parse message first to validate it
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Println("chat payload invalid:", err)
		return
	}

	// Safely read partner with lock
	clientsMu.RLock()
	partner := c.Partner
	clientsMu.RUnlock()

	if partner == nil {
		log.Printf("No partner for %s - message not sent: %s\n", c.ID, payload.Message)
		// Optionally send error back to client
		c.Send <- SendMessageType{
			Message: []byte(`{"op":"error","message":"No partner connected. Please find a match first."}`),
			Type:    websocket.TextMessage,
		}
		return
	}

	log.Printf("[%s] says to [%s]: %s\n", c.ID, partner.ID, payload.Message)

	// Send message to partner with non-blocking send
	select {
	case partner.Send <- SendMessageType{
		Message: fmt.Appendf(nil,
			`{"op":"chat","message":"%s"}`, payload.Message,
		),
		Type: websocket.TextMessage,
	}:
		// Message sent successfully
	default:
		log.Printf("Failed to send message to partner %s (buffer full)\n", partner.ID)
	}
}
