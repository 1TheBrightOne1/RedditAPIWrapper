package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/1TheBrightOne1/RedditAPIWrapper/oauth"
)

func Search_Reddit_Names(query string, exact bool, include_over_18 bool, include_unadvertisable bool) *http.Request {
	body := strings.NewReader(
		fmt.Sprintf(
			"exact=%t&include_over_18=%t&include_unadvertisable=%t&query=%s",
			exact,
			include_over_18,
			include_unadvertisable,
			query,
		),
	)
	req, _ := http.NewRequest("POST", oauth.Endpoint+"/api/search_reddit_names", body)

	req.Header.Add("Content-Type", "application/x-www-urlformencoded")
	return req
}
