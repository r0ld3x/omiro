package main

import (
	"encoding/json"
	"log"

	"github.com/gorilla/websocket"
)

func handleWebRTCOffer(c *Client, data json.RawMessage) {
	if c.Partner == nil {
		return
	}

	var payload struct {
		SDP string `json:"sdp"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		log.Println("invalid offer:", err)
		return
	}

	partner := safeGetClient(c.Partner.ID)
	if partner == nil {
		return
	}

	outgoing := map[string]any{
		"op": "webrtc_offer",
		"data": map[string]any{
			"from": c.ID,
			"sdp":  payload.SDP,
		},
	}

	bytes, _ := json.Marshal(outgoing)
	partner.Send <- SendMessageType{Type: websocket.TextMessage, Message: bytes}
}

func handleWebRTCAnswer(c *Client, data json.RawMessage) {
	if c.Partner == nil {
		return
	}

	var payload struct {
		SDP string `json:"sdp"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		log.Println("invalid answer:", err)
		return
	}

	partner := safeGetClient(c.Partner.ID)
	if partner == nil {
		return
	}

	outgoing := map[string]any{
		"op": "webrtc_answer",
		"data": map[string]any{
			"from": c.ID,
			"sdp":  payload.SDP,
		},
	}

	bytes, _ := json.Marshal(outgoing)
	partner.Send <- SendMessageType{Type: websocket.TextMessage, Message: bytes}
}

func handleICECandidate(c *Client, data json.RawMessage) {
	if c.Partner == nil {
		return
	}

	var payload struct {
		Candidate map[string]any `json:"candidate"`
	}

	if err := json.Unmarshal(data, &payload); err != nil {
		log.Println("invalid ice candidate:", err)
		return
	}

	partner := safeGetClient(c.Partner.ID)
	if partner == nil {
		return
	}

	outgoing := map[string]any{
		"op": "ice_candidate",
		"data": map[string]any{
			"from":      c.ID,
			"candidate": payload.Candidate,
		},
	}

	bytes, _ := json.Marshal(outgoing)
	partner.Send <- SendMessageType{Type: websocket.TextMessage, Message: bytes}
}

func safeGetClient(id string) *Client {
	clientsMu.RLock()
	c := clients[id]
	clientsMu.RUnlock()
	return c
}
