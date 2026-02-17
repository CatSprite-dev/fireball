package main

type MoexSecurityEntry struct {
	secID      string
	shortName  string
	secName    string
	ISIN       string
	latName    string
	currencyID string
}

type MoexSecurities struct {
	Securities struct {
		Metadata struct {
			SECID struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"SECID"`
			BOARDID struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"BOARDID"`
			SHORTNAME struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"SHORTNAME"`
			PREVPRICE struct {
				Type string `json:"type"`
			} `json:"PREVPRICE"`
			LOTSIZE struct {
				Type string `json:"type"`
			} `json:"LOTSIZE"`
			FACEVALUE struct {
				Type string `json:"type"`
			} `json:"FACEVALUE"`
			STATUS struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"STATUS"`
			BOARDNAME struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"BOARDNAME"`
			DECIMALS struct {
				Type string `json:"type"`
			} `json:"DECIMALS"`
			SECNAME struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"SECNAME"`
			REMARKS struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"REMARKS"`
			MARKETCODE struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"MARKETCODE"`
			INSTRID struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"INSTRID"`
			SECTORID struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"SECTORID"`
			MINSTEP struct {
				Type string `json:"type"`
			} `json:"MINSTEP"`
			PREVWAPRICE struct {
				Type string `json:"type"`
			} `json:"PREVWAPRICE"`
			FACEUNIT struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"FACEUNIT"`
			PREVDATE struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"PREVDATE"`
			ISSUESIZE struct {
				Type string `json:"type"`
			} `json:"ISSUESIZE"`
			ISIN struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"ISIN"`
			LATNAME struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"LATNAME"`
			REGNUMBER struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"REGNUMBER"`
			PREVLEGALCLOSEPRICE struct {
				Type string `json:"type"`
			} `json:"PREVLEGALCLOSEPRICE"`
			CURRENCYID struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"CURRENCYID"`
			SECTYPE struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"SECTYPE"`
			LISTLEVEL struct {
				Type string `json:"type"`
			} `json:"LISTLEVEL"`
			SETTLEDATE struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"SETTLEDATE"`
		} `json:"metadata"`
		Columns []string        `json:"columns"`
		Data    [][]interface{} `json:"data"`
	} `json:"securities"`
	Marketdata struct {
		Metadata struct {
			SECID struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"SECID"`
			BOARDID struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"BOARDID"`
			BID struct {
				Type string `json:"type"`
			} `json:"BID"`
			BIDDEPTH struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"BIDDEPTH"`
			OFFER struct {
				Type string `json:"type"`
			} `json:"OFFER"`
			OFFERDEPTH struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"OFFERDEPTH"`
			SPREAD struct {
				Type string `json:"type"`
			} `json:"SPREAD"`
			BIDDEPTHT struct {
				Type string `json:"type"`
			} `json:"BIDDEPTHT"`
			OFFERDEPTHT struct {
				Type string `json:"type"`
			} `json:"OFFERDEPTHT"`
			OPEN struct {
				Type string `json:"type"`
			} `json:"OPEN"`
			LOW struct {
				Type string `json:"type"`
			} `json:"LOW"`
			HIGH struct {
				Type string `json:"type"`
			} `json:"HIGH"`
			LAST struct {
				Type string `json:"type"`
			} `json:"LAST"`
			LASTCHANGE struct {
				Type string `json:"type"`
			} `json:"LASTCHANGE"`
			LASTCHANGEPRCNT struct {
				Type string `json:"type"`
			} `json:"LASTCHANGEPRCNT"`
			QTY struct {
				Type string `json:"type"`
			} `json:"QTY"`
			VALUE struct {
				Type string `json:"type"`
			} `json:"VALUE"`
			VALUEUSD struct {
				Type string `json:"type"`
			} `json:"VALUE_USD"`
			WAPRICE struct {
				Type string `json:"type"`
			} `json:"WAPRICE"`
			LASTCNGTOLASTWAPRICE struct {
				Type string `json:"type"`
			} `json:"LASTCNGTOLASTWAPRICE"`
			WAPTOPREVWAPRICEPRCNT struct {
				Type string `json:"type"`
			} `json:"WAPTOPREVWAPRICEPRCNT"`
			WAPTOPREVWAPRICE struct {
				Type string `json:"type"`
			} `json:"WAPTOPREVWAPRICE"`
			CLOSEPRICE struct {
				Type string `json:"type"`
			} `json:"CLOSEPRICE"`
			MARKETPRICETODAY struct {
				Type string `json:"type"`
			} `json:"MARKETPRICETODAY"`
			MARKETPRICE struct {
				Type string `json:"type"`
			} `json:"MARKETPRICE"`
			LASTTOPREVPRICE struct {
				Type string `json:"type"`
			} `json:"LASTTOPREVPRICE"`
			NUMTRADES struct {
				Type string `json:"type"`
			} `json:"NUMTRADES"`
			VOLTODAY struct {
				Type string `json:"type"`
			} `json:"VOLTODAY"`
			VALTODAY struct {
				Type string `json:"type"`
			} `json:"VALTODAY"`
			VALTODAYUSD struct {
				Type string `json:"type"`
			} `json:"VALTODAY_USD"`
			ETFSETTLEPRICE struct {
				Type string `json:"type"`
			} `json:"ETFSETTLEPRICE"`
			TRADINGSTATUS struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"TRADINGSTATUS"`
			UPDATETIME struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"UPDATETIME"`
			LASTBID struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"LASTBID"`
			LASTOFFER struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"LASTOFFER"`
			LCLOSEPRICE struct {
				Type string `json:"type"`
			} `json:"LCLOSEPRICE"`
			LCURRENTPRICE struct {
				Type string `json:"type"`
			} `json:"LCURRENTPRICE"`
			MARKETPRICE2 struct {
				Type string `json:"type"`
			} `json:"MARKETPRICE2"`
			NUMBIDS struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"NUMBIDS"`
			NUMOFFERS struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"NUMOFFERS"`
			CHANGE struct {
				Type string `json:"type"`
			} `json:"CHANGE"`
			TIME struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"TIME"`
			HIGHBID struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"HIGHBID"`
			LOWOFFER struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"LOWOFFER"`
			PRICEMINUSPREVWAPRICE struct {
				Type string `json:"type"`
			} `json:"PRICEMINUSPREVWAPRICE"`
			OPENPERIODPRICE struct {
				Type string `json:"type"`
			} `json:"OPENPERIODPRICE"`
			SEQNUM struct {
				Type string `json:"type"`
			} `json:"SEQNUM"`
			SYSTIME struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"SYSTIME"`
			CLOSINGAUCTIONPRICE struct {
				Type string `json:"type"`
			} `json:"CLOSINGAUCTIONPRICE"`
			CLOSINGAUCTIONVOLUME struct {
				Type string `json:"type"`
			} `json:"CLOSINGAUCTIONVOLUME"`
			ISSUECAPITALIZATION struct {
				Type string `json:"type"`
			} `json:"ISSUECAPITALIZATION"`
			ISSUECAPITALIZATIONUPDATETIME struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"ISSUECAPITALIZATION_UPDATETIME"`
			ETFSETTLECURRENCY struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"ETFSETTLECURRENCY"`
			VALTODAYRUR struct {
				Type string `json:"type"`
			} `json:"VALTODAY_RUR"`
			TRADINGSESSION struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"TRADINGSESSION"`
			TRENDISSUECAPITALIZATION struct {
				Type string `json:"type"`
			} `json:"TRENDISSUECAPITALIZATION"`
		} `json:"metadata"`
		Columns []string        `json:"columns"`
		Data    [][]interface{} `json:"data"`
	} `json:"marketdata"`
	Dataversion struct {
		Metadata struct {
			DataVersion struct {
				Type string `json:"type"`
			} `json:"data_version"`
			Seqnum struct {
				Type string `json:"type"`
			} `json:"seqnum"`
			TradeDate struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"trade_date"`
			TradeSessionDate struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"trade_session_date"`
		} `json:"metadata"`
		Columns []string        `json:"columns"`
		Data    [][]interface{} `json:"data"`
	} `json:"dataversion"`
	MarketdataYields struct {
		Metadata struct {
			Boardid struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"boardid"`
			Secid struct {
				Type    string `json:"type"`
				Bytes   int    `json:"bytes"`
				MaxSize int    `json:"max_size"`
			} `json:"secid"`
		} `json:"metadata"`
		Columns []string      `json:"columns"`
		Data    []interface{} `json:"data"`
	} `json:"marketdata_yields"`
}
