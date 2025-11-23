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

	// If client has a partner, notify them and clear relationship
	if c.Partner != nil {
		partner := c.Partner
		c.Partner = nil
		partner.Partner = nil

		// Notify partner about disconnection
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

	// Client stays connected, just cleared partner relationship
	// Frontend will automatically call join_queue after this
	log.Printf("client %s ready for next match\n", c.ID)
}

func forwardWebRTC(c *Client, typ string, payload json.RawMessage) {
	// Payload MUST contain "to": "<partnerID>"
	var pkt struct {
		To string `json:"to"`
	}

	json.Unmarshal(payload, &pkt)
	if pkt.To == "" {
		log.Println("missing 'to' field for WebRTC op")
		return
	}

	redis.SendToClient(pkt.To, map[string]any{
		"op":   typ,
		"data": payload,
	})
}

func notifyPartnerLeft(userID string) {
	redis.SendToClient(userID, map[string]any{
		"op": "partner_disconnected",
	})
}
