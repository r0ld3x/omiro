package redis

import (
	"encoding/json"
	"fmt"
	"time"
)

func StoreChatMessage(roomID, senderID, message string) error {
	key := fmt.Sprintf("chat:%s", roomID)
	msgData := map[string]any{
		"sender_id": senderID,
		"message":   message,
		"timestamp": time.Now().Unix(),
	}
	jsonData, err := json.Marshal(msgData)
	if err != nil {
		return err
	}
	return Client.LPush(Ctx, key, jsonData).Err()
}

func GetChatHistory(roomID string, count int64) ([]string, error) {
	key := fmt.Sprintf("chat:%s", roomID)
	return Client.LRange(Ctx, key, 0, count-1).Result()
}
