package main

import (
	"fmt"
	"time"

	"github.com/1TheBrightOne1/RedditAPIWrapper/oauth"
)

func main() {
	fmt.Println(oauth.CredentialsFilePath)

	dur, _ := time.ParseDuration(("20s"))
	time.Sleep(dur)
}
