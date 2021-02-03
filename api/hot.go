package api

import (
	"fmt"
	"net/http"

	"github.com/1TheBrightOne1/RedditAPIWrapper/oauth"
)

func Get_Hot(subreddit, before, after string, limit int) *http.Request {
	url := AttachParams(fmt.Sprintf("%s/r/%s/hot", oauth.Endpoint, subreddit), map[string]string{
		"before": before,
		"after":  after,
		"limit":  fmt.Sprint(limit),
	})

	req, _ := http.NewRequest("GET", url, nil)

	return req
}
