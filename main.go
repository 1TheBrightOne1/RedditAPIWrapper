package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/1TheBrightOne1/RedditAPIWrapper/api"
	"github.com/1TheBrightOne1/RedditAPIWrapper/models"
	"github.com/1TheBrightOne1/RedditAPIWrapper/oauth"
)

func main() {
	creds := oauth.GetCredentials()
	resp, err := creds.SendRequest(api.Get_Hot("wallstreetbets", "", "", 0))

	if err != nil {
		log.Fatal(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	file, _ := os.Create("Response.json")

	fmt.Fprintf(file, string(body))

	file, _ = os.Open("Response.json")
	b, _ := ioutil.ReadAll(file)
	posts := models.NewPosts(b)
	fmt.Println(posts)
}
