package service

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sort"
	"strings"

	"github.com/1TheBrightOne1/RedditAPIWrapper/wsbmonitor"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
)

func MakeHTTPHandler(s Service) http.Handler {
	r := mux.NewRouter()
	e := MakeEndpoints(s)

	r.Methods("GET").Path("/stocks/{id}").Handler(httptransport.NewServer(
		e.GetArticlesByStock,
		decodeStockSymbolRequest,
		encodeGetArticlesResponse,
	))

	r.Methods("PUT").Path("/ignoredstocks/{id}").Handler(httptransport.NewServer(
		e.AddToIgnoredStocks,
		decodeStockSymbolRequest,
		encodePutIgnoredStockResponse,
	))

	return r
}

func encodePutIgnoredStockResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	w.WriteHeader(200)

	return nil
}

func decodeStockSymbolRequest(ctx context.Context, r *http.Request) (request interface{}, err error) {
	vars := mux.Vars(r)
	id, ok := vars["id"]

	if !ok {
		return nil, errors.New("stock missing")
	}
	return StockSymbolRequest{Stock: id}, nil
}

func encodeGetArticlesResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	articles := response.([]*Article)
	if len(articles) == 0 {
		return nil
	}

	primaryStock := articles[0].PrimaryStock

	sort.SliceStable(articles, func(i, j int) bool {
		return articles[i].Upvotes[primaryStock] > articles[j].Upvotes[primaryStock]
	})

	w.Header().Add("Content-Type", "text/html")
	var b strings.Builder

	b.WriteString("<div>")

	for _, article := range articles {
		b.WriteString("<div>")
		b.WriteString(fmt.Sprintf("<div><div><a target=_blank href='https://reddit.com/%s'>%s</a></div><ul>%s</ul><ul>%s</ul><div>%s</div></div>", article.Link, article.Link, stocksToHTML(article.Upvotes), commentsToHTML(article.Comments), article.LastScraped.String()))
		b.WriteString("</div>")
	}
	b.WriteString("</div>")

	w.Write([]byte(b.String()))
	return nil
}

func commentsToHTML(comments []string) string {
	var b strings.Builder
	stocks := make(map[string][]string)

	b.WriteString("<div>")
	for _, comment := range comments {
		addStock := func(stock string) {
			stocks[stock] = append(stocks[stock], comment)
		}

		wsbmonitor.ExtractTickers(comment, addStock)
	}

	writeComments := func(comments []string) string {
		var b strings.Builder
		for _, comment := range comments {
			b.WriteString(fmt.Sprintf("<li>%s</li>", comment))
		}
		return b.String()
	}

	for stock, comments := range stocks {
		b.WriteString(fmt.Sprintf("<div>%s</div><ul>%s</ul></div>", stock, writeComments(comments)))
	}
	b.WriteString("</div>")
	return b.String()
}

func stocksToHTML(stocks map[string]int) string {
	var b strings.Builder

	for key, val := range stocks {
		b.WriteString(fmt.Sprintf("<li>%s\t\t\t%d</li>", key, val))
	}

	return b.String()
}
