package pkg

import (
	"TokenHoldersAnalyse/internal/entity"
	"TokenHoldersAnalyse/internal/usecase"
	"bytes"
	"context"
	"encoding/json"
	"github.com/redis/go-redis/v9"
	"io"
	"net/http"
)

func FetchTokenNameAndDecimals(tokenHash string, rdb *redis.Client) (int, string, string) {
	hashedData, err := usecase.GetTokenDataFromCache(tokenHash, rdb)
	if err == nil {
		return hashedData.Decimals, hashedData.TokenInfo.Name, hashedData.TokenInfo.Symbol
	}

	resp, err := http.Get("https://api.solana.fm/v1/tokens/" + tokenHash)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {

		}
	}(resp.Body)

	var responseBody = bytes.Buffer{}
	_, err = responseBody.ReadFrom(resp.Body)
	decodedResp := responseBody.String()

	var data entity.TokenInfo
	err = json.Unmarshal([]byte(decodedResp), &data)
	if err != nil {
		return 0, "Unknown", "Unknown"
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	err = rdb.HSet(context.Background(), "tokenData", tokenHash, jsonData).Err()
	if err != nil {
		panic(err)
	}

	return data.Decimals, data.TokenInfo.Name, data.TokenInfo.Symbol
}
