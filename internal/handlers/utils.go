package handlers

import (
	"net/http"
	"os"
	"time"

	"github.com/CatSprite-dev/fireball/internal/pkg"
)

func setSessionCookie(w http.ResponseWriter, sessionID string, setToDelete bool) {
	expiration := time.Now().Add(24 * time.Hour)
	maxAge := 0
	if setToDelete {
		maxAge = -1
		expiration = time.Unix(0, 0)
	}
	cookie := http.Cookie{
		Name:     "session_token",
		Value:    sessionID,
		Expires:  expiration,
		HttpOnly: true,
		Secure:   os.Getenv("ENV") == "production",
		SameSite: http.SameSiteStrictMode,
		MaxAge:   maxAge,
	}
	http.SetCookie(w, &cookie)
}

func getSessionFromCookie(r *http.Request) (string, error) {
	cookie, err := r.Cookie("session_token")
	if err != nil {
		return "", err
	}
	return cookie.Value, nil
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
