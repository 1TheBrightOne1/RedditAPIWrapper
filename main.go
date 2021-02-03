package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/1TheBrightOne1/RedditAPIWrapper/api"
	"github.com/1TheBrightOne1/RedditAPIWrapper/models"
	"github.com/1TheBrightOne1/RedditAPIWrapper/oauth"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var KnownStocks map[string]int

var counter = promauto.NewCounterVec(prometheus.CounterOpts{Name: "posts"}, []string{"ticker"})

var watchList []string

func main() {
	go startPrometheus()

	KnownStocks = LoadTickers()

	creds := oauth.GetCredentials()
	after := ""
	for {
		//GetNewListings. Add to list that gets monitored for how many new comments are coming in.
		resp, err := creds.SendRequest(api.Get_New("wallstreetbets", "", after, 0))

		if err != nil {
			log.Fatal(err)
		}

		body, _ := ioutil.ReadAll(resp.Body)

		file, _ := os.Create("dirty.json")
		fmt.Fprintf(file, string(body))
		listings := models.NewListing(body)

		fmt.Println(listings)
		//after = listings.Data.ChildrenWrapper.Listings[len(listings.Data.ChildrenWrapper.Listings)-1].Data.Name
		//PullTickers(posts)

		dur, _ := time.ParseDuration("60s")
		time.Sleep(dur)
	}
}

func startPrometheus() {
	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":2112", nil)
}

func ExtractTickers(text string) []string {
	var tickers []string
	words := strings.Split(text, " ")

	for _, word := range words {
		if _, ok := KnownStocks[strings.ToLower(strings.TrimSpace(word))]; ok {
			fmt.Printf("Matched for %s\n", word)
			tickers = append(tickers, word)
		}
	}

	return tickers
}

func LoadTickers() map[string]int {
	out := make(map[string]int)
	file, err := os.Open("tickers.txt")
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(file)
	for {
		ticker, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		out[strings.TrimSpace(ticker)] = 1
	}
	return out
}

func WalkListing(in interface{}, handle func(models.Listing)) {
	if listing, ok := in.(models.Listing); ok {
		if listing.Kind == "Listing" {
			WalkListing(listing.Data.Children.Listings, handle)
		} else {
			handle(listing)
		}

		if len(listing.Data.Children.Listings) > 0 {
			for _, child := range listing.Data.Children.Listings {
				WalkListing(child, handle)
			}
		}

	} else if listingSlice, ok := in.([]models.Listing); ok {
		for _, listing := range listingSlice {
			WalkListing(listing, handle)
		}
	}
}

func parseListingsForTickers(listing models.Listing) {
	if listing.Kind == "t1" {
		tickers := ExtractTickers(listing.Data.Body)
		for _, ticker := range tickers {
			counter.WithLabelValues(ticker).Add(float64(listing.Data.Ups))
		}
		//TODO process replies
	} else if listing.Kind == "t3" {
		tickers := ExtractTickers(listing.Data.Title)
		for _, ticker := range tickers {
			counter.WithLabelValues(ticker).Add(float64(listing.Data.Ups))
		}
	} else if listing.Kind == "more" {
		fmt.Println(listing.Data.Children.MoreComments)
	}
}

func getArticles(listing models.Listing) {
	if listing.Kind == "t3" {
		watchList = append(watchList, listing.Data.Name)
	}
}
