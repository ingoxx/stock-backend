package redis

import (
	"bytes"
	"encoding/json"
	"github.com/go-redis/redis"
	"github.com/ingoxx/stock-backend/internal/domain"
	cusErr "github.com/ingoxx/stock-backend/internal/error"
	"sync"
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
