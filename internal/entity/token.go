package entity

type TokenInfo struct {
	Decimals  int `json:"decimals"`
	TokenInfo struct {
		Name   string `json:"name"`
		Symbol string `json:"symbol"`
	} `json:"tokenList"`
}
