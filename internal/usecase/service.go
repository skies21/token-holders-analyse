package usecase

import (
	"TokenHoldersAnalyse/internal/entity"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
)

func GetTokenDataFromCache(tokenHash string, rdb *redis.Client) (entity.TokenInfo, error) {
	val, err := rdb.HGet(context.Background(), "tokenData", tokenHash).Result()
	if err != nil {
		return entity.TokenInfo{}, err
	}

	var data entity.TokenInfo
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		return entity.TokenInfo{}, err
	}

	return data, nil
}
