package entity

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
