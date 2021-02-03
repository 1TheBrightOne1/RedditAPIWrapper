package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/1TheBrightOne1/RedditAPIWrapper/oauth"
)

func Get_Comments(link string) *http.Request {
	url := fmt.Sprintf("%s%s", oauth.Endpoint, link)

	fmt.Println(url)
	req, _ := http.NewRequest("GET", url, nil)

	return req
}

func Get_MoreChildren(article string, childIds []string) *http.Request {
	url := AttachParams(fmt.Sprintf("%s/api/morechildren", oauth.Endpoint), map[string]string{
		"children": strings.Join(childIds, ","),
		"link_id":  article,
	})

	fmt.Println(url)
	req, _ := http.NewRequest("GET", url, nil)

	return req
}
