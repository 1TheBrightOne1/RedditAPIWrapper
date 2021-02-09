package oauth

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/1TheBrightOne1/RedditAPIWrapper/config"
)

func GetCredentials() *Credentials {
	if _, err := os.Stat(config.GlobalConfig.HomePath + "/.credentials"); err != nil {
		getAppCredentials()
	}

	creds := loadCredentialsFromFile()

	if creds.Token == nil {
		creds.startAuthorizationGrant()
	}

	creds.Wait = make(chan int)

	return creds
}

func getAppCredentials() {

	getInput := func(prompt string) string {
		reader := bufio.NewReader(os.Stdin)
		fmt.Println(prompt)

		value, _ := reader.ReadString('\n')
		return strings.TrimSpace(value)
	}

	clientID := getInput("Enter client ID")
	clientSecret := getInput("Enter client secret")
	userAgent := getInput("Enter user agent name")
	redirectURL := getInput("Enter redirect URL")

	creds := &Credentials{
		UserAgent:    userAgent,
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
	}

	file, err := os.Create(config.GlobalConfig.HomePath + "/.credentials")
	defer file.Close()
	if err != nil {
		log.Fatal(err)
	}

	if err := creds.writeToFile(file); err != nil {
		log.Fatal(err)
	}
}

func loadCredentialsFromFile() *Credentials {
	file, err := os.Open(config.GlobalConfig.HomePath + "/.credentials")
	if err != nil {
		log.Fatal(err)
	}

	body, _ := ioutil.ReadAll(file)
	creds := &Credentials{}
	err = json.Unmarshal(body, creds)

	if err != nil {
		log.Fatal(err)
	}

	if creds == nil {
		return nil
	}

	return creds
}
