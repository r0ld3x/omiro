package main

import (
	"fmt"
	"log"
	"slices"
	"sync"

	"github.com/gorilla/websocket"
)

var queue []string
var queueMu sync.Mutex

func handleJoinQueue(c *Client) {
	queueMu.Lock()
	defer queueMu.Unlock()

	if slices.Contains(queue, c.ID) {
		log.Println("client already in queue:", c.ID)
		return
	}
	queue = append(queue, c.ID)
	log.Println("added to queue:", c.ID)
	findMatch(c)
}

func handleLeaveQueue(c *Client) {
	queueMu.Lock()
	for i, id := range queue {
		if id == c.ID {
			queue = append(queue[:i], queue[i+1:]...)
			log.Println("removed from queue:", c.ID)
			break
		}
	}
	queueMu.Unlock()
	if c.Partner != nil {
		partner := c.Partner
		c.Partner = nil
		partner.Partner = nil

		log.Printf("notifying partner %s about disconnection of %s\n", partner.ID, c.ID)
		msg := []byte(`{"op":"partner_disconnected"}`)
		select {
		case partner.Send <- SendMessageType{
			Message: msg,
			Type:    websocket.TextMessage,
		}:
			log.Printf("successfully notified partner %s\n", partner.ID)
		default:
			log.Printf("failed to notify partner %s (channel full or closed)\n", partner.ID)
		}
	}
}

func findMatch(c *Client) {
	log.Println("finding match for:", c.ID)
	log.Printf("queue length: %d\n", len(queue))
	if len(queue) < 2 {
		log.Println("not enough clients in queue to find match")
		return
	}

	id1 := queue[0]
	id2 := queue[1]
	log.Printf("found match: %s <-> %s\n", id1, id2)
	queue = queue[2:]
	log.Printf("queue length after match: %d\n", len(queue))

	clientsMu.RLock()
	c1 := clients[id1]
	c2 := clients[id2]
	clientsMu.RUnlock()

	if c1 == nil || c2 == nil {
		log.Println("client not found:", id1, id2)
		return
	}
	c1.Partner = c2
	c2.Partner = c1

	log.Println("matched:", id1, "<->", id2)

	// Client 1 (first in queue) will be the caller
	matchMsg1 := SendMessageType{
		Type: websocket.TextMessage,
		Message: fmt.Appendf(nil,
			`{"op":"match_found","partner":"%s","should_call":true}`, id2,
		),
	}
	c1.Send <- matchMsg1

	// Client 2 will be the callee (waits for offer)
	matchMsg2 := SendMessageType{
		Type: websocket.TextMessage,
		Message: fmt.Appendf(nil,
			`{"op":"match_found","partner":"%s","should_call":false}`, id1,
		),
	}
	c2.Send <- matchMsg2

	log.Printf("Client %s is CALLER, Client %s is CALLEE\n", id1, id2)
}
