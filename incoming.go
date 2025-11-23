package main

import (
	"encoding/json"
	"log"
	"omiro/redis"
)

func (c *Client) readPump() {
	defer func() {
		handleLeaveQueue(c)
		notifyPartnerLeft(c.ID)

		// Remove from local memory
		clientsMu.Lock()
		delete(clients, c.ID)
		clientsMu.Unlock()

		c.Conn.Close()
		log.Println("client disconnected:", c.ID)
	}()

	for {
		_, msg, err := c.Conn.ReadMessage()
		if err != nil {
			return
		}

		// Parse incoming payload (per-message, not globally)
		var incoming struct {
			Op   string          `json:"op"`
			Data json.RawMessage `json:"data"`
		}

		if err := json.Unmarshal(msg, &incoming); err != nil {
			log.Println("json unmarshal error:", err)
			continue
		}

		routeMessage(c, incoming.Op, incoming.Data)
	}
}

func routeMessage(c *Client, op string, data json.RawMessage) {
	switch op {

	case "join_queue":
		handleJoinQueue(c)

	case "chat":
		handleChat(c, data)

	case "disconnect":
		handleClientDisconnect(c)

	case "next":
		handleNextPartner(c)

	case "webrtc_offer":
		handleWebRTCOffer(c, data)

	case "webrtc_answer":
		handleWebRTCAnswer(c, data)

	case "ice_candidate":
		handleICECandidate(c, data)

	default:
		log.Println("unknown op:", op)
	}
}

func handleClientDisconnect(c *Client) {
	log.Println("client disconnected via request:", c.ID)

	// Make Redis notify partner
	notifyPartnerLeft(c.ID)

	// Remove from queue
	handleLeaveQueue(c)

	// Remove from memory
	clientsMu.Lock()
	delete(clients, c.ID)
	clientsMu.Unlock()

	c.Conn.Close()
}

func handleNextPartner(c *Client) {
	log.Println("client looking for next partner:", c.ID)

	clientsMu.Lock()
	partner := c.Partner
	if partner != nil {
		c.Partner = nil
		partner.Partner = nil
	}
	clientsMu.Unlock()

	// Now notify the partner (outside of lock to avoid deadlock)
	if partner != nil {
		msg := []byte(`{"op":"partner_disconnected"}`)
		select {
		case partner.Send <- SendMessageType{
			Message: msg,
			Type:    1, // websocket.TextMessage
		}:
			log.Printf("notified partner %s that %s is looking for next\n", partner.ID, c.ID)
		default:
			log.Println("failed to notify partner")
		}
	}

	log.Printf("client %s ready for next match\n", c.ID)
}

func notifyPartnerLeft(userID string) {
	redis.SendToClient(userID, map[string]any{
		"op": "partner_disconnected",
	})
}
