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
	for {
		u1, _ := Client.RPop(Ctx, "matchmaking:queue").Result()
		if u1 == "" {
			time.Sleep(100 * time.Millisecond)
			continue
		}

		u2, _ := Client.RPop(Ctx, "matchmaking:queue").Result()
		if u2 == "" {
			Client.LPush(Ctx, "matchmaking:queue", u1)
			time.Sleep(100 * time.Millisecond)
			continue
		}

		go createPair(u1, u2)
	}
}

func createPair(u1, u2 string) {
	c1, err1 := GetClient(u1)
	c2, err2 := GetClient(u2)

	if err1 != nil || err2 != nil || c1 == nil || c2 == nil {
		log.Println("pair error: client not found")
		return
	}

	// assign
	c1.PartnerID = u2
	c2.PartnerID = u1

	// persist
	b1, _ := json.Marshal(c1)
	b2, _ := json.Marshal(c2)

	Client.Set(Ctx, "client:"+u1, b1, 2*time.Hour)
	Client.Set(Ctx, "client:"+u2, b2, 2*time.Hour)

	// notify both
	SendToClient(u1, map[string]any{
		"type":       "match_found",
		"partner_id": u2,
	})

	SendToClient(u2, map[string]any{
		"type":       "match_found",
		"partner_id": u1,
	})
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
