package pkg

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func FetchTokenHolders(tokenHash string) []string {
	var page = 1
	var addresses []string
	for {
		params := map[string]interface{}{
			"page":  page,
			"limit": 1000,
			"mint":  tokenHash,
		}

		content := map[string]interface{}{
			"jsonrpc": "2.0",
			"method":  "getTokenAccounts",
			"id":      "text",
			"params":  params,
		}

		dataInBytes, err := json.Marshal(content)
		if err != nil {
			panic(err)
		}

		resp, err := http.Post("https://mainnet.helius-rpc.com/?api-key="+os.Getenv("HELIUS_API"), "application/json", bytes.NewBuffer(dataInBytes))
		if err != nil {
			panic(err)
		}

		var responseBody bytes.Buffer
		_, err = responseBody.ReadFrom(resp.Body)
		if err != nil {
			err = resp.Body.Close()
			if err != nil {
				return nil
			}
			return nil
		}
		err = resp.Body.Close()
		if err != nil {
			panic(err)
		}

		decodedResponse := responseBody.String()

		var data map[string]interface{}
		err = json.Unmarshal([]byte(decodedResponse), &data)
		if err != nil {
			fmt.Println("Ошибка при разборе JSON:", err)
			break
		}

		if data["result"] == nil {
			break
		}
		if tokenAccounts, ok := data["result"].(map[string]interface{})["token_accounts"].([]interface{}); ok && len(tokenAccounts) == 0 {
			break
		} else {
			for _, account := range tokenAccounts {
				if acc, ok := account.(map[string]interface{}); ok {
					if address, ok := acc["address"].(string); ok {
						addresses = append(addresses, address)
					}
				}
			}
		}
		page++
	}
	return addresses
}
