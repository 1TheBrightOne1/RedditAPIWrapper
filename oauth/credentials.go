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
	"time"

	"github.com/1TheBrightOne1/RedditAPIWrapper/config"
)

const Endpoint = "https://oauth.reddit.com"

type Credentials struct {
	ClientID     string    `json:"clientID"`
	ClientSecret string    `json:"clientSecret"`
	UserAgent    string    `json:"userAgent"`
	RedirectURL  string    `json:"redirectURL"`
	Token        *Token    `json:"token"`
	Used         int64     `json:"-"`
	Wait         chan int  `json:"-"`
	ResetTime    time.Time `json:"-"`
}

func (creds *Credentials) SendRequest(req *http.Request) (*http.Response, error) {
	fmt.Println(req.URL)

	creds.Used++

	if creds.Used == 1 {
		fmt.Println("Starting rate timer")

		dur, _ := time.ParseDuration("60s")
		creds.ResetTime = time.Now().Add(dur)
	}

	if creds.LimitHit() {
		fmt.Println("Waiting for rate limit reset")
		time.Sleep(time.Now().Sub(creds.ResetTime))
		creds.Used = 0
		fmt.Println("Resuming")
	}

	bearer := "Bearer " + creds.Token.Token

	req.Header.Add("Authorization", bearer)
	req.Header.Add("User-Agent", creds.UserAgent)

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		for {
			fmt.Printf("Error sending request. Waiting and retrying. %s\n", err.Error())
			time.Sleep(time.Second * 30)

			fmt.Println("Retrying")
			resp, err = client.Do(req)
			if err != nil {
				break
			}
		}

	}

	if resp.StatusCode != 200 {
		fmt.Printf("Received status code %d. Waiting and retrying\n", resp.StatusCode)

		creds.refreshToken()

		resp, err = client.Do(req)

		if err != nil {
			return nil, err
		}
	}

	rateUsed, _ := strconv.ParseInt(resp.Header.Get("X-Ratelimit-Used"), 10, 64)
	remaining, _ := strconv.ParseFloat(resp.Header.Get("X-Ratelimit-Remaining"), 64)
	resetTime, _ := strconv.ParseInt(resp.Header.Get("X-Ratelimit-Reset"), 10, 64)

	if remaining <= 0 {
		dur, _ := time.ParseDuration(fmt.Sprintf("%ds", resetTime))
		time.Sleep(dur)
	}

	fmt.Println(rateUsed, remaining, resetTime)
	return resp, err
}

func (cred *Credentials) LimitHit() bool {
	return cred.Used >= 60
}

func (creds *Credentials) manageRate() {
	dur, _ := time.ParseDuration("60s")
	time.Sleep(dur)

	fmt.Println("Resetting rate")

	creds.Used = 0
	creds.Wait <- 1
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

	file, _ := os.Create(config.GlobalConfig.HomePath + "/.credentials")
	creds.writeToFile(file)
}

func (creds *Credentials) refreshToken() {
	fmt.Println("Refreshing token")
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

	if newToken != nil {
		newToken.Refresh = creds.Token.Refresh
		creds.Token = newToken

		file, _ := os.Create(config.GlobalConfig.HomePath + "/.credentials")
		creds.writeToFile(file)
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
	file.Close()

	return err
}

func (creds *Credentials) isExpired() bool {
	return time.Now().After(creds.Token.ExpiresOn)
}
