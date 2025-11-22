package redis

import (
	"fmt"
	"time"
)

func BanIP(ip string, duration time.Duration, reason string) error {
	key := fmt.Sprintf("ban:%s", ip)
	return Client.Set(Ctx, key, reason, duration).Err()
}

func IsIPBanned(ip string) (bool, string, error) {
	key := fmt.Sprintf("ban:%s", ip)
	reason, err := Client.Get(Ctx, key).Result()
	if err != nil {
		if err.Error() == "redis: nil" {
			return false, "", nil
		}
		return false, "", err
	}
	return true, reason, nil
}

func UnbanIP(ip string) error {
	key := fmt.Sprintf("ban:%s", ip)
	return Client.Del(Ctx, key).Err()
}
