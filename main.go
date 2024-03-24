package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
	"io"
	"log"
	"math"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"
)

type Transfer struct {
	TransactionHash string         `json:"transactionHash"`
	Data            []TransferData `json:"data"`
}

type TransferData struct {
	TokenHash              string      `json:"token"`
	InstructionIndex       int         `json:"instructionIndex"`
	InnerInstructionIndex  int         `json:"innerInstructionIndex"`
	Action                 string      `json:"action"`
	Amount                 float64     `json:"amount"`
	Timestamp              int         `json:"timestamp"`
	Status                 string      `json:"status"`
	Source                 string      `json:"source"`
	SourceAssociation      interface{} `json:"sourceAssociation"`
	Destination            interface{} `json:"destination"`
	DestinationAssociation interface{} `json:"destinationAssociation"`
}

type TransferResponse struct {
	Status  string     `json:"status"`
	Message string     `json:"message"`
	Results []Transfer `json:"results"`
}

type TokenInfo struct {
	Decimals  int `json:"decimals"`
	TokenInfo struct {
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
	} `json:"tokenList"`
}

var ctx = context.Background()

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
	rdb := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
	tokenHash := "EMcz7rjNJatWAPvG34iPgrwhcnfZdBWKJQFR1b6rCWT2"
	accounts := fetchTokenHolders(tokenHash)

	rateLimit, err := strconv.Atoi(os.Getenv("RATE_LIMIT"))
	if err != nil {
		panic(err)
		return
	}
	//stackLen := len(accounts) / rateLimit
	stackLen := 1

	var wg sync.WaitGroup
	var mutex sync.Mutex

	transfers := make(map[string]interface{})
	for i := 0; i < stackLen; i++ {
		for _, account := range accounts[i*rateLimit : (i+1)*rateLimit] {
			wg.Add(1)
			go func(address string) {
				defer wg.Done()
				accountTransfers := fetchAccountTransfers(address, tokenHash, rdb)
				mutex.Lock()
				transfers[address] = accountTransfers
				mutex.Unlock()
			}(account)
		}
		wg.Wait()
		if i != stackLen-1 {
			time.Sleep(time.Minute)
		}
	}
}

func fetchTokenHolders(tokenHash string) []string {
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

func fetchAccountTransfers(address string, tokenHash string, rdb *redis.Client) map[string]interface{} {
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

	var response TransferResponse
	err = json.Unmarshal([]byte(decodedReq), &response)
	if err != nil {
		fmt.Println("fetchAccountTransfers(): ", err)
		return nil
	}

	transfersData := make(map[string]interface{})
	for _, transfer := range response.Results {
		for _, transferData := range transfer.Data {
			if transferData.Action == "transfer" {
				transfersData = map[string]interface{}{
					"token":                  transferData.TokenHash,
					"instructionIndex":       transferData.InstructionIndex,
					"innerInstructionIndex":  transferData.InnerInstructionIndex,
					"action":                 transferData.Action,
					"amount":                 transferData.Amount,
					"timestamp":              transferData.Timestamp,
					"status":                 transferData.Status,
					"source":                 transferData.Source,
					"sourceAssociation":      transferData.SourceAssociation,
					"destination":            transferData.Destination,
					"destinationAssociation": transferData.DestinationAssociation,
				}
				decimals, _, symbol := fetchTokenNameAndDecimals(transferData.TokenHash, rdb)
				var totalAmount float64
				if symbol != "Unknown" {
					coinPrice := fetchHistoryCoinPrice(symbol, strconv.Itoa(transferData.Timestamp))
					totalAmount = transferData.Amount / math.Pow(10, float64(decimals)) * coinPrice
				} else {
					break
				}
				i, err := strconv.ParseInt(strconv.Itoa(transferData.Timestamp), 10, 64)
				if err != nil {
					panic(err)
				}
				date := time.Unix(i, 3).Truncate(time.Second)
				fmt.Printf("%s %.2f %s\n", symbol, totalAmount, date)
			}
		}
	}

	return transfersData
}

func getTokenDataFromCache(tokenHash string, rdb *redis.Client) (TokenInfo, error) {
	val, err := rdb.HGet(ctx, "tokenData", tokenHash).Result()
	if err != nil {
		return TokenInfo{}, err
	}

	var data TokenInfo
	err = json.Unmarshal([]byte(val), &data)
	if err != nil {
		return TokenInfo{}, err
	}

	return data, nil
}

func fetchTokenNameAndDecimals(tokenHash string, rdb *redis.Client) (int, string, string) {
	hashedData, err := getTokenDataFromCache(tokenHash, rdb)
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

	var data TokenInfo
	err = json.Unmarshal([]byte(decodedResp), &data)
	if err != nil {
		return 0, "Unknown", "Unknown"
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}

	err = rdb.HSet(ctx, "tokenData", tokenHash, jsonData).Err()
	if err != nil {
		panic(err)
	}

	return data.Decimals, data.TokenInfo.Name, data.TokenInfo.Symbol
}

func fetchHistoryCoinPrice(symbol string, timestamp string) float64 {
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
