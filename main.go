package main

import (
	"bufio"
	"net/http"
	"os"
	"strings"

	"github.com/1TheBrightOne1/RedditAPIWrapper/wsbmonitor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	go startPrometheus()

	scraper := wsbmonitor.NewScraper()
	go scraper.Scrape()

	cmd := bufio.NewReader(os.Stdin)
	for {
		text, _ := cmd.ReadString('\n')
		text = strings.TrimSpace(text)
		values := strings.Split(text, " ")

		if len(values) > 1 && values[0] == "remove" {
			wsbmonitor.AddToIgnoredStocks(values[1])
		} else if len(values) > 1 && values[0] == "get" {
			scraper.GetArticlesByStock(values[1])
		}
	}
}

func startPrometheus() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
