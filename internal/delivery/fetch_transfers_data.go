package delivery

import (
	"TokenHoldersAnalyse/internal/entity"
	"TokenHoldersAnalyse/internal/pkg"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"os"
	"strconv"
	"sync"
	"time"
)

func FetchTransfersData(c *fiber.Ctx) error {
	tokenHash := c.Params("tokenHash")
	if tokenHash == "" {
		return c.SendString("Missing tokenHash")
	}

	accounts := pkg.FetchTokenHolders(tokenHash)
	if len(accounts) == 0 {
		return c.SendString("No accounts found")
	}

	rateLimit, err := strconv.Atoi(os.Getenv("RATE_LIMIT"))
	if err != nil {
		panic(err)
		return err
	}
	//stackLen := len(accounts) / rateLimit
	stackLen := 1

	var wg sync.WaitGroup
	var mutex sync.Mutex

	transfers := make(map[string][]entity.TradeInfo)
	for i := 0; i < stackLen; i++ {
		for _, account := range accounts[i*rateLimit : (i+1)*rateLimit] {
			wg.Add(1)
			go func(address string) {
				defer wg.Done()
				accountTransfers := pkg.FetchAccountTransfers(address, tokenHash)
				if accountTransfers != nil {
					mutex.Lock()
					transfers[address] = accountTransfers
					mutex.Unlock()
				}
			}(account)
		}
		wg.Wait()
		if i != stackLen-1 {
			time.Sleep(time.Minute)
		}
	}
	TransfersData, err := json.Marshal(transfers)
	if err != nil {
		panic(err)
	}
	jsonTransfersData := string(TransfersData)
	return c.SendString(jsonTransfersData)
}
