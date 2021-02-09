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

	/*cmd := bufio.NewReader(os.Stdin)
	for {
		text, _ := cmd.ReadString('\n')
		text = strings.TrimSpace(text)
		values := strings.Split(text, " ")

		if len(values) > 1 && values[0] == "remove" {
			wsbmonitor.AddToIgnoredStocks(values[1])
		} else if len(values) > 1 && values[0] == "get" {
			scraper.GetArticlesByStock(values[1])
		}
	}*/
}

func startPrometheus() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}
