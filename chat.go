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
	clientsMu.RLock()
	partner := c.Partner
	clientsMu.RUnlock()
	if partner == nil {
		log.Println("partner not found:", c.ID)
		return
	}
	if err := json.Unmarshal(data, &payload); err != nil {
		log.Println("chat payload invalid:", err)
		return
	}

	log.Printf("[%s] says: %s\n", c.ID, payload.Message)
	partner.Send <- SendMessageType{
		Message: fmt.Appendf(nil,
			`{"op":"chat","message":"%s"}`, payload.Message,
		),
		Type: websocket.TextMessage,
	}

}
