package oauth

import (
	"encoding/json"
	"fmt"
	"log"
	"time"
)

type Token struct {
	Token     string    `json:"access_token"`
	Expires   int       `json:"expires_in"`
	Refresh   string    `json:"refresh_token"`
	Scope     string    `json:"scope"`
	ExpiresOn time.Time `json:"expires_on"`
}

func NewToken(respBody []byte) *Token {
	token := &Token{}

	fmt.Println(string(respBody))
	err := json.Unmarshal(respBody, token)
	if err != nil {
		log.Fatal(err)
	}

	dur, _ := time.ParseDuration(fmt.Sprintf("%ds", token.Expires))
	token.ExpiresOn = time.Now().Add(dur)

	fmt.Println(token)

	return token
}
