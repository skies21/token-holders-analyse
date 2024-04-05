package entity

type TradeInfo struct {
	TokenName string  `json:"tokenName"`
	Amount    float64 `json:"amount"`
	Timestamp string  `json:"timestamp"`
}
