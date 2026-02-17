package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"slices"
	"strings"
	"time"
)

const MoexSecuritiesURL = "https://iss.moex.com/iss/engines/stock/markets/shares/securities.json"

func main() {
	cli := http.Client{
		Timeout: 10 * time.Second,
	}

	body := strings.NewReader("")
	req, err := http.NewRequest("GET", MoexSecuritiesURL, body)
	if err != nil {
		log.Fatalf("cannot form request")
	}

	res, err := cli.Do(req)
	if err != nil {
		log.Fatalf("request failed")
	}

	defer res.Body.Close()
	data, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatalf("reading failed")
	}

	securities := MoexSecurities{}
	err = json.Unmarshal(data, &securities)
	if err != nil {
		log.Fatalf("unmarshalling failed")
	}

	columns := securities.Securities.Columns
	idxSECID := slices.Index(columns, "SECID")
	idxSHORTNAME := slices.Index(columns, "SHORTNAME")
	idxSECNAME := slices.Index(columns, "SECNAME")
	idxISIN := slices.Index(columns, "ISIN")
	idxLATNAME := slices.Index(columns, "LATNAME")
	idxCURRENCYID := slices.Index(columns, "CURRENCYID")

	securitiesList := []MoexSecurityEntry{}
	for _, item := range securities.Securities.Data {
		securitiesList = append(securitiesList, MoexSecurityEntry{
			secID:      item[idxSECID].(string),
			shortName:  item[idxSHORTNAME].(string),
			secName:    item[idxSECNAME].(string),
			ISIN:       item[idxISIN].(string),
			latName:    item[idxLATNAME].(string),
			currencyID: item[idxCURRENCYID].(string),
		})
	}

	r := securitiesList[0]
	fmt.Printf("%s %s %s %s %s %s ", r.secID, r.shortName, r.secName, r.ISIN, r.latName, r.currencyID)
}
