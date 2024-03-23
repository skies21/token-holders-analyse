package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"io"
	"log"
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
	InstructionIndex       int         `json:"instructionIndex"`
	InnerInstructionIndex  int         `json:"innerInstructionIndex"`
	Action                 string      `json:"action"`
	Amount                 int         `json:"amount"`
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

func init() {
	// loads values from .env into the system
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

func main() {
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
		println("its i", i)
		for _, account := range accounts[i*rateLimit : (i+1)*rateLimit] {
			wg.Add(1)
			go func(address string) {
				defer wg.Done()
				accountTransfers := fetchAccountTransfers(address, tokenHash)
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
	for key, value := range transfers {
		fmt.Println(key, value)
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

func fetchAccountTransfers(address string, tokenHash string) map[string]interface{} {
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
		fmt.Println("Ошибка при разборе JSON:", err)
		return nil
	}

	transfersData := make(map[string]interface{})
	for _, transfer := range response.Results {
		for _, transferData := range transfer.Data {
			if transferData.Action == "transfer" {
				transfersData = map[string]interface{}{
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
			}
		}
	}

	return transfersData
}
