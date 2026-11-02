package tinvest

type MoneyValue struct {
	Currency string `json:"currency"`
	Units    string `json:"units"`
	Nano     int32  `json:"nano"`
}

type Quotation struct {
	Units string `json:"units"`
	Nano  int32  `json:"nano"`
}
