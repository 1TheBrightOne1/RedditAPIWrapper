package oauth

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

const Endpoint = "https://oauth.reddit.com"

type Credentials struct {
	ClientID     string `json:"clientID"`
	ClientSecret string `json:"clientSecret"`
	UserAgent    string `json:"userAgent"`
	RedirectURL  string `json:"redirectURL"`
	Token        *Token `json:"token"`
	Used         int64
	Remaining    float64
	ResetTime    time.Time
	Lock         *sync.RWMutex
}

func (creds *Credentials) SendRequest(req *http.Request) (*http.Response, error) {
	creds.Lock.RLock()
	defer creds.Lock.RUnlock()

	if time.Now().Before(creds.ResetTime) && creds.Remaining <= 0 {
		dur := time.Now().Sub(creds.ResetTime)
		time.Sleep(dur)
	}

	bearer := "Bearer " + creds.Token.Token

	req.Header.Add("Authorization", bearer)
	req.Header.Add("User-Agent", creds.UserAgent)

	client := &http.Client{}
	resp, err := client.Do(req)

	creds.Used, _ = strconv.ParseInt(resp.Header.Get("X-Ratelimit-Used"), 10, 64)
	creds.Remaining, _ = strconv.ParseFloat(resp.Header.Get("X-Ratelimit-Remaining"), 64)
	resetTime := resp.Header.Get("X-Ratelimit-Reset")

	fmt.Println(creds.Used, creds.Remaining, resetTime)
	return resp, err
}

func (creds *Credentials) startAuthorizationGrant() {
	url := fmt.Sprintf(
		"https://www.reddit.com/api/v1/authorize?client_id=%s&response_type=code&state=notneeded&redirect_uri=%s&duration=permanent&scope=identity edit flair history modconfig modflair modlog modposts modwiki mysubreddits privatemessages read report save submit subscribe vote wikiedit wikiread",
		creds.ClientID,
		creds.RedirectURL,
	)

	fmt.Printf("Please authorize app at %s\n", url)
	fmt.Println("Enter URL after being redirected")

	reader := bufio.NewReader(os.Stdin)
	redirect, _ := reader.ReadString('\n')
	redirect = strings.TrimSpace(redirect)

	regex := regexp.MustCompile(`.*?code=(?P<Code>.+?)$`)
	match := regex.FindStringSubmatch(redirect)

	result := make(map[string]string)
	for i, name := range regex.SubexpNames() {
		if i != 0 && name != "" {
			result[name] = match[i]
		}
	}

	creds.getToken(strings.TrimSpace(result["Code"]))
}

func (creds *Credentials) getToken(code string) {
	reader := strings.NewReader(fmt.Sprintf("grant_type=authorization_code&code=%s&redirect_uri=%s", code, creds.RedirectURL))
	client := &http.Client{}
	req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", reader)
	req.SetBasicAuth(creds.ClientID, creds.ClientSecret)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("User-Agent", creds.UserAgent)
	resp, err := client.Do(req)

	if err != nil {
		log.Fatal(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)
	creds.Token = NewToken(body)

	file, _ := os.Create(CredentialsFilePath)
	creds.writeToFile(file)
	go creds.manageRefresh()
}

func (creds *Credentials) manageRefresh() {
	for {
		dur := creds.Token.ExpiresOn.Sub(time.Now())
		if dur.Seconds() > 0 {
			time.Sleep(dur)
		}

		creds.Lock.Lock()
		defer creds.Lock.Unlock()
		reader := strings.NewReader(fmt.Sprintf("grant_type=refresh_token&refresh_token=%s", creds.Token.Refresh))
		client := &http.Client{}
		req, _ := http.NewRequest("POST", "https://www.reddit.com/api/v1/access_token", reader)

		req.SetBasicAuth(creds.ClientID, creds.ClientSecret)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("User-Agent", creds.UserAgent)

		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}

		body, _ := ioutil.ReadAll(resp.Body)
		newToken := NewToken(body)
		newToken.Refresh = creds.Token.Refresh
		creds.Token = newToken

		file, _ := os.Create(CredentialsFilePath)
		creds.writeToFile(file)
		return
	}
}

func (creds *Credentials) writeToFile(file *os.File) error {
	writer := bufio.NewWriter(file)

	bytes, err := json.Marshal(creds)
	if err != nil {
		log.Fatal(err)
	}

	writer.Write(bytes)
	writer.Flush()

	return err
}

func (creds *Credentials) isExpired() bool {
	return time.Now().After(creds.Token.ExpiresOn)
}
