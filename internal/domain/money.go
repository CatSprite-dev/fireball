package domain

type MoneyValue struct {
	Currency string `json:"currency"`
	Units    string `json:"units"`
	Nano     int    `json:"nano"`
}

type Quotation struct {
	Units string `json:"units"`
	Nano  int    `json:"nano"`
}
