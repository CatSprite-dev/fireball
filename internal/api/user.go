package api

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type UserInfo struct {
	QualifiedForWorkWith []string `json:"qualifiedForWorkWith"`
	RiskLevelCode        string   `json:"riskLevelCode"`
	QualStatus           bool     `json:"qualStatus"`
	PremStatus           bool     `json:"premStatus"`
	Tariff               string   `json:"tariff"`
	UserID               string   `json:"userId"`
}

func (c *Client) GetUserInfo(url string) (UserInfo, error) {
	userUrl := url + ".UsersService/GetInfo"
	_ = godotenv.Load()
	token := os.Getenv("token")

	value, ok := c.cache.Get(url)
	if ok {
		var user UserInfo
		err := json.Unmarshal(value, &user)
		if err != nil {
			return UserInfo{}, fmt.Errorf("unmarshal error: %s", err)
		}
		return user, nil
	}

	data, err := c.DoRequest(userUrl, token)
	if err != nil {
		return UserInfo{}, fmt.Errorf("do request error: %s", err)
	}

	var user UserInfo
	err = json.Unmarshal(data, &user)
	if err != nil {
		return UserInfo{}, fmt.Errorf("unmarshal error: %s", err)
	}
	return user, nil
}
