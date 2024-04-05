package pkg

import (
	"TokenHoldersAnalyse/internal/entity"
	"TokenHoldersAnalyse/internal/redisClient"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strconv"
	"time"
)

func FetchAccountTransfers(address string, tokenHash string) []entity.TradeInfo {
	url := "https://api.solana.fm/v0/accounts/" + address + "/transfers?mint=" + tokenHash + "&page=1"
	req, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			panic(err)
		}
	}(req.Body)

	var responseBody bytes.Buffer
	_, err = responseBody.ReadFrom(req.Body)
	if err != nil {
		panic(err)
	}

	decodedReq := responseBody.String()

	var response entity.TransferResponse
	err = json.Unmarshal([]byte(decodedReq), &response)
	if err != nil {
		fmt.Println("fetchAccountTransfers(): ", err)
		return nil
	}

	var tradesInfo []entity.TradeInfo

	for _, transfer := range response.Results {
		for _, transferData := range transfer.Data {
			if transferData.Action == "transfer" {
				if transferData.TokenHash == "" {
					break
				}
				decimals, name, symbol := FetchTokenNameAndDecimals(transferData.TokenHash, redisClient.Rdb)
				var totalAmount float64
				if symbol != "Unknown" {
					coinPrice := FetchHistoryCoinPrice(symbol, strconv.Itoa(transferData.Timestamp))
					totalAmount = transferData.Amount / math.Pow(10, float64(decimals)) * coinPrice
				} else {
					break
				}
				i, err := strconv.ParseInt(strconv.Itoa(transferData.Timestamp), 10, 64)
				if err != nil {
					panic(err)
				}
				date := time.Unix(i, 3).Truncate(time.Second)
				dateString := date.Format("2006-01-02 15:04:05")

				tradesInfo = append(tradesInfo, entity.TradeInfo{
					TokenName: name,
					Amount:    totalAmount,
					Timestamp: dateString,
				})
			}
		}
	}

	return tradesInfo
}
