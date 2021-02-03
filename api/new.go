package api

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/1TheBrightOne1/RedditAPIWrapper/oauth"
)

func Get_New(subreddit, before, after string, limit int) (*http.Request, error) {
	url := fmt.Sprintf("%s/r/%s/new?", oauth.Endpoint, subreddit)

	if before != "" && after != "" {
		return nil, errors.New("before and after both specified")
	} else if before != "" {
		url += "before=" + before + "&"
	} else if after != "" {
		url += "after=" + after + "&"
	}

	if limit > 0 {
		url += "limit=" + fmt.Sprint(limit)
	}

	req, err := http.NewRequest("GET", url, nil)

	return req, err
}
