package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func FetchHistoryCoinPrice(symbol string, timestamp string) float64 {
	params := map[string]interface{}{
		"fsym":  symbol,
		"tsym":  "USDT",
		"toTs":  timestamp,
		"limit": 1,
	}

	dataInBytes, err := json.Marshal(params)
	if err != nil {
		panic(err)
	}

	resp, err := http.Post("https://min-api.cryptocompare.com/data/v2/histoday", "application/json", bytes.NewBuffer(dataInBytes))
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

	var data map[string]interface{}
	err = json.Unmarshal([]byte(decodedResp), &data)
	if err != nil {
		fmt.Println("fetchHistoryCoinPrice(): ", err)
	}

	dataData, ok := data["Data"].(map[string]interface{})
	if !ok || dataData == nil {
		return 0.0
	}

	highPrices, ok := dataData["Data"].([]interface{})
	if !ok || highPrices == nil {
		return 0.0
	}

	totalHighPrice := 0.0
	for _, entry := range highPrices {
		entryData := entry.(map[string]interface{})
		highPrice := entryData["high"].(float64)
		totalHighPrice += highPrice
	}

	avgPrice := totalHighPrice / float64(len(highPrices))
	return avgPrice
}
