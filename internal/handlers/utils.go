package handlers

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/CatSprite-dev/fireball/internal/pkg"
)

func getTokenFromHeader(headers http.Header) (string, error) {
	token := headers.Get("T-Token")
	if token == "" {
		return "", errors.New("token is not provided")
	}
	if len(token) < 10 || !strings.HasPrefix(token, "t.") {
		return "", errors.New("invalid token")
	}
	return token, nil
}

func PeriodToParams(period string) (time.Time, time.Time, pkg.CandleInterval) {
	now := time.Now().UTC()
	switch period {
	case "7d":
		return now.AddDate(0, 0, -7), now, pkg.CandleInterval4Hour
	case "1M":
		return now.AddDate(0, -1, 0), now, pkg.CandleIntervalDay
	case "3M":
		return now.AddDate(0, -3, 0), now, pkg.CandleIntervalDay
	case "6M":
		return now.AddDate(0, -6, 0), now, pkg.CandleIntervalDay
	case "YTD":
		return time.Date(now.Year(), 1, 1, 0, 0, 0, 0, now.Location()), now, pkg.CandleIntervalDay
	case "1Y":
		return now.AddDate(-1, 0, 0), now, pkg.CandleIntervalDay
	case "5Y":
		return now.AddDate(-5, 0, 0), now, pkg.CandleIntervalWeek
	case "ALL":
		return now.AddDate(-10, 0, 0), now, pkg.CandleIntervalMonth
	default:
		return now.AddDate(0, -6, 0), now, pkg.CandleIntervalDay
	}
}
