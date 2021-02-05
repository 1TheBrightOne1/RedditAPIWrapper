package main

import (
	"net/http"

	"github.com/1TheBrightOne1/RedditAPIWrapper/wsbmonitor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	go startPrometheus()

	scraper := wsbmonitor.NewScraper()
	scraper.Scrape()
}

func startPrometheus() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
