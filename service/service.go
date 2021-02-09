package service

import (
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/1TheBrightOne1/RedditAPIWrapper/config"
	"github.com/1TheBrightOne1/RedditAPIWrapper/wsbmonitor"
)

type Service interface {
	GetArticles(context.Context, string) []*Article
	AddToIgnoredStocks(context.Context, string) bool
}

type Article struct {
	Link         string
	Comments     []string
	LastScraped  time.Time
	Upvotes      map[string]int
	PrimaryStock string
}

type WSBService struct {
	Scraper *wsbmonitor.Scraper
}

func (w *WSBService) GetArticles(ctx context.Context, tickerSymbol string) []*Article {
	articles := make([]*Article, 0)
	posts := w.Scraper.GetArticlesByStock(tickerSymbol)

	for _, post := range posts {
		article := &Article{
			Link:         post.Article,
			LastScraped:  post.LastScraped,
			Upvotes:      post.Stocks,
			PrimaryStock: tickerSymbol,
		}

		f, err := os.Open(fmt.Sprintf("%s/%s.comments", config.GlobalConfig.HomePath, post.Id))
		defer f.Close()
		if err == nil {
			bytes, _ := ioutil.ReadAll(f)

			article.Comments = strings.Split(string(bytes), "\n")
		}

		articles = append(articles, article)
	}

	return articles
}

func (w *WSBService) AddToIgnoredStocks(ctx context.Context, tickerSymbol string) bool {
	wsbmonitor.AddToIgnoredStocks(tickerSymbol)
	return true
}
