package redis

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/config"
	"github.com/ingoxx/stock-backend/internal/domain"
	cusErr "github.com/ingoxx/stock-backend/internal/error"
)

type VerifyRepo struct {
	mu     sync.RWMutex
	client *redis.Client
}

func NewVerifyRepo(client *redis.Client) domain.VerifyRepository {
	return &VerifyRepo{client: client}
}

func (vr *VerifyRepo) GetAuthData(vd string) error {
	result, err := vr.client.HGet("auth", "users").Result()
	if err != nil {
		return err
	}

	var data []string
	bufferString := bytes.NewBufferString(result)
	if err = json.Unmarshal(bufferString.Bytes(), &data); err != nil {
		return err
	}

	var isFind bool
	for _, v := range data {
		if v == vd {
			isFind = true
			break
		}
	}

	if !isFind {
		return cusErr.AuthError
	}

	return nil
}

func (vr *VerifyRepo) GetJwtToken(user, jt string) error {
	result, err := vr.client.HGet(config.Jak, user).Result()
	if err != nil {
		return err
	}

	if result != jt {
		return fmt.Errorf("无效Token")
	}

	return nil
}

func (vr *VerifyRepo) DelJwtToken(user string) error {
	if err := vr.client.HDel(config.Jak, user).Err(); err != nil {
		return err
	}

	return nil
}
