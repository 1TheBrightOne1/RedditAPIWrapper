package api

import (
	"fmt"
	"net/http"

	"github.com/1TheBrightOne1/RedditAPIWrapper/oauth"
)

func Get_New(subreddit, before, after string, limit int) *http.Request {
	url := AttachParams(fmt.Sprintf("%s/r/%s/new", oauth.Endpoint, subreddit), map[string]string{
		"before": before,
		"after":  after,
		"limit":  fmt.Sprint(limit),
	})

	req, _ := http.NewRequest("GET", url, nil)

	return req
}
