package redis

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

/********************************
 * CLIENT META
 ********************************/

type ClientMeta struct {
	ID        string `json:"id"`
	IP        string `json:"ip"`
	ServerID  string `json:"server_id"`
	PartnerID string `json:"partner_id,omitempty"`
	InQueue   bool   `json:"in_queue"`
}

func RegisterClient(clientID, ip, serverID string) error {
	data := ClientMeta{
		ID:       clientID,
		IP:       ip,
		ServerID: serverID,
		InQueue:  false,
	}

	b, _ := json.Marshal(data)

	key := "client:" + clientID
	return Client.Set(Ctx, key, b, 2*time.Hour).Err()
}

func GetClient(clientID string) (*ClientMeta, error) {
	key := "client:" + clientID

	raw, err := Client.Get(Ctx, key).Result()
	if err != nil {
		return nil, err
	}

	var meta ClientMeta
	if err := json.Unmarshal([]byte(raw), &meta); err != nil {
		return nil, err
	}

	return &meta, nil
}

/********************************
 * SERVER REGISTRATION
 ********************************/

func RegisterServer(serverID string) {
	key := "server:" + serverID

	err := Client.Set(Ctx, key, time.Now().Unix(), 60*time.Second).Err()
	if err != nil {
		log.Fatal("failed to register server:", err)
	}

	go func() {
		ticker := time.NewTicker(55 * time.Second)
		for range ticker.C {
			Client.Set(Ctx, key, time.Now().Unix(), 60*time.Second)
		}
	}()

	log.Println("Server registered:", serverID)
}

/********************************
 * SEND TO CLIENT (PUBSUB ROUTING)
 ********************************/

func SendToClient(clientID string, payload any) error {
	meta, err := GetClient(clientID)
	if err != nil {
		return err
	}

	if meta.ServerID == "" {
		return fmt.Errorf("client %s has no server", clientID)
	}

	channel := "signal:" + meta.ServerID

	msg := map[string]any{
		"user_id": clientID,
		"payload": payload,
	}

	b, _ := json.Marshal(msg)

	return Client.Publish(Ctx, channel, b).Err()
}

/********************************
 * SIGNAL SUBSCRIBER (WS SERVER)
 ********************************/

// This must be called inside EACH WS server, passing its serverID AND a handler
func StartSignalSubscriber(serverID string, handler func(userID string, payload json.RawMessage)) {
	ch := Client.Subscribe(Ctx, "signal:"+serverID).Channel()

	go func() {
		for msg := range ch {
			var incoming struct {
				UserID  string          `json:"user_id"`
				Payload json.RawMessage `json:"payload"`
			}

			if err := json.Unmarshal([]byte(msg.Payload), &incoming); err != nil {
				log.Println("signal parse error:", err)
				continue
			}

			handler(incoming.UserID, incoming.Payload)
		}
	}()
}

/********************************
 * MATCHMAKER
 ********************************/

func StartMatchmaker() {
	log.Println("🎯 Matchmaker started - waiting for users in queue...")

	for {
		log.Println("📋 Waiting for first user from queue...")
		res, err := Client.BLPop(Ctx, 0*time.Second, "matchmaking:queue").Result()
		if err != nil {
			log.Printf("❌ Error popping first user from queue: %v\n", err)
			continue
		}
		if len(res) < 2 {
			log.Printf("⚠️ Invalid queue response (length %d)\n", len(res))
			continue
		}
		u1 := res[1]
		log.Printf("✅ Got first user from queue: %s\n", u1)

		log.Println("📋 Waiting for second user from queue...")
		res2, err := Client.BLPop(Ctx, 0*time.Second, "matchmaking:queue").Result()
		if err != nil {
			log.Printf("❌ Error popping second user from queue: %v (putting %s back)\n", err, u1)
			// Put user1 back
			Client.LPush(Ctx, "matchmaking:queue", u1)
			continue
		}
		if len(res2) < 2 {
			log.Printf("⚠️ Invalid queue response for second user (length %d), putting %s back\n", len(res2), u1)
			Client.LPush(Ctx, "matchmaking:queue", u1)
			continue
		}
		u2 := res2[1]
		log.Printf("✅ Got second user from queue: %s\n", u2)

		log.Printf("🔗 Creating pair: %s <-> %s\n", u1, u2)
		go createPair(u1, u2)
	}
}

func createPair(u1, u2 string) {
	log.Printf("🎯 createPair: Starting to create pair for %s and %s\n", u1, u2)

	c1, err1 := GetClient(u1)
	c2, err2 := GetClient(u2)

	if err1 != nil {
		log.Printf("❌ createPair: Error getting client %s: %v\n", u1, err1)
		return
	}
	if err2 != nil {
		log.Printf("❌ createPair: Error getting client %s: %v\n", u2, err2)
		return
	}
	if c1 == nil {
		log.Printf("❌ createPair: Client %s is nil\n", u1)
		return
	}
	if c2 == nil {
		log.Printf("❌ createPair: Client %s is nil\n", u2)
		return
	}

	log.Printf("✅ createPair: Both clients found, assigning partners\n")

	c1.PartnerID = u2
	c2.PartnerID = u1

	b1, err := json.Marshal(c1)
	if err != nil {
		log.Printf("❌ createPair: Error marshaling client %s: %v\n", u1, err)
		return
	}
	b2, err := json.Marshal(c2)
	if err != nil {
		log.Printf("❌ createPair: Error marshaling client %s: %v\n", u2, err)
		return
	}

	err = Client.Set(Ctx, "client:"+u1, b1, 2*time.Hour).Err()
	if err != nil {
		log.Printf("❌ createPair: Error saving client %s to Redis: %v\n", u1, err)
		return
	}

	err = Client.Set(Ctx, "client:"+u2, b2, 2*time.Hour).Err()
	if err != nil {
		log.Printf("❌ createPair: Error saving client %s to Redis: %v\n", u2, err)
		return
	}

	log.Printf("✅ createPair: Saved partner assignments to Redis\n")

	// notify both
	log.Printf("📤 createPair: Notifying %s about match with %s\n", u1, u2)
	SendToClient(u1, map[string]any{
		"type":       "match_found",
		"partner_id": u2,
	})

	log.Printf("📤 createPair: Notifying %s about match with %s\n", u2, u1)
	SendToClient(u2, map[string]any{
		"type":       "match_found",
		"partner_id": u1,
	})

	log.Printf("🎉 createPair: Successfully created pair %s <-> %s\n", u1, u2)
}

/********************************
 * RATE LIMIT
 ********************************/

func CheckRateLimit(ip string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("ratelimit:%s", ip)

	count, err := Client.Incr(Ctx, key).Result()
	if err != nil {
		return false, err
	}

	if count == 1 {
		Client.Expire(Ctx, key, window)
	}

	return count <= int64(limit), nil
}
