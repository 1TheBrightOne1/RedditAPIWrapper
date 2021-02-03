package api

import (
	"fmt"
	"net/http"

	"github.com/1TheBrightOne1/RedditAPIWrapper/oauth"
)

func Get_Hot(subreddit, before, after string, limit int) *http.Request {
	url := fmt.Sprintf("%s/r/%s/hot", oauth.Endpoint, subreddit)

	if before != "" && after != "" {
		return nil
	} else if before != "" {
		url += "?before=" + before
	} else if after != "" {
		url += "after=" + after
	}

	if limit > 0 {
		panic("Needs param formatting")
		url += "limit=" + fmt.Sprint(limit)
	}

	req, _ := http.NewRequest("GET", url, nil)

	fmt.Println(url)

	return req
}
