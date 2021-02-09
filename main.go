package main

import (
	"net/http"

	"github.com/1TheBrightOne1/RedditAPIWrapper/config"
	"github.com/1TheBrightOne1/RedditAPIWrapper/service"
	"github.com/1TheBrightOne1/RedditAPIWrapper/wsbmonitor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	config.InitConfig()
	wsbmonitor.InitTickers()

	go startPrometheus()

	scraper := wsbmonitor.NewScraper()
	go scraper.Scrape()

	s := &service.WSBService{
		Scraper: scraper,
	}

	h := service.MakeHTTPHandler(s)

	http.ListenAndServe("0.0.0.0:9191", h)
}

func startPrometheus() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
