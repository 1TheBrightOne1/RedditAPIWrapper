package wsbmonitor

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/1TheBrightOne1/RedditAPIWrapper/api"
	"github.com/1TheBrightOne1/RedditAPIWrapper/config"
	"github.com/1TheBrightOne1/RedditAPIWrapper/models"
	"github.com/1TheBrightOne1/RedditAPIWrapper/oauth"
)

var (
	charsOnly     = regexp.MustCompile(`\w+`)
	watchedStocks map[string]string
)

type Scraper struct {
	creds     *oauth.Credentials
	after     string
	watchList *watchList
}

func NewScraper() *Scraper {
	return &Scraper{
		creds:     oauth.GetCredentials(),
		after:     "",
		watchList: newWatchList(),
	}
}

func (s *Scraper) Scrape() {
	if len(s.watchList.Posts) == 0 {
		s.getHotArticles()
	}

	s.after = ""
	for {
		s.getNewArticles()

		go s.getUpdatedListings()
		dur, _ := time.ParseDuration("60s")
		time.Sleep(dur)
	}
}

func (s *Scraper) GetArticlesByStock(stock string) []WatchedItem {
	return s.watchList.GetArticlesByStock(stock)
}

func (s *Scraper) getHotArticles() {
	fmt.Println("Getting Hot Articles")
	resp, err := s.creds.SendRequest(api.Get_Hot("wallstreetbets", "", "", 100))

	if err != nil {
		fmt.Println(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	listings := models.NewListing(body)

	models.WalkListing(listings, s.addListingsToWatchList())
}

func (s *Scraper) getNewArticles() {
	fmt.Println("Getting New Articles")
	resp, err := s.creds.SendRequest(api.Get_New("wallstreetbets", "", s.after, 100))

	if err != nil {
		log.Println(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	listings := models.NewListing(body)

	s.after = listings[0].Data.After

	models.WalkListing(listings, s.addListingsToWatchList())
}

func (s *Scraper) getUpdatedListings() {
	dur, _ := time.ParseDuration("60s")
	stopAfter := time.Now().Add(dur)
	for {
		if stopAfter.Sub(time.Now()).Seconds() < 0 {
			return
		}

		post := s.watchList.getFreshPost()
		if post.Id == "" {
			return
		}

		resp, err := s.creds.SendRequest(api.Get_Comments(post.Article))

		if err != nil {
			return
		}

		body, err := ioutil.ReadAll(resp.Body)

		if err != nil {
			fmt.Println(err.Error())
			return
		}

		listing := models.NewListing(body)
		s.updateArticleScore(listing)
	}
}

func (s *Scraper) getCommentsForArticle(link string) {
	fmt.Printf("Getting comments for article %s\n", link)
	resp, err := s.creds.SendRequest(api.Get_Comments(link))

	if err != nil {
		fmt.Println(err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		fmt.Println(err.Error())
		return
	}

	listings := models.NewListing(body)

	s.updateArticleScore(listings)
}

func (s *Scraper) addListingsToWatchList() func(models.Listing) {
	return func(listing models.Listing) {
		if listing.Kind == "t3" {
			fmt.Printf("Adding %s to watch list\n", listing.Data.Name)

			s.getCommentsForArticle(listing.Data.Link)
		}
	}
}

//updateArticleScore takes a root listing and aggregates all stock mentions and upvotes
func (s *Scraper) updateArticleScore(listings []models.Listing) {
	var f *os.File

	var root models.Listing
	score := make(map[string]int)
	aggregator := func(listing models.Listing) {
		if listing.Kind == "t3" {
			root = listing
			f, _ = os.Create(fmt.Sprintf("%s/%s.comments", config.GlobalConfig.HomePath, listing.Data.Name))

			stocks := extractTickers(listing.Data.Title)
			for stock := range stocks {
				score[stock] += listing.Data.Ups
			}
		} else if listing.Kind == "t1" {
			stocks := extractTickers(listing.Data.Body)
			for stock := range stocks {
				score[stock] += listing.Data.Ups
			}

			if len(stocks) > 0 && f != nil {
				fmt.Fprintf(f, "%s\n", listing.Data.Body)
			}
		}
	}

	models.WalkListing(listings, aggregator)
	if f != nil {
		f.Close()
	}

	s.watchList.updatePost(root, score)
}

func ExtractTickers(text string, handler func(string)) {
	words := strings.Split(text, " ")

	for _, word := range words {
		if len(word) > 4 && word[0:4] == "http" {
			continue
		}
		matches := charsOnly.FindAllString(word, -1)

		for _, match := range matches {
			cleaned := strings.ToLower(match)
			if stock, ok := watchedStocks[cleaned]; ok {
				handler(stock)
			}
		}
	}
}

func extractTickers(text string) map[string]int {
	stocks := make(map[string]int)

	writeToStocks := func(stock string) {
		stocks[stock] = 1
	}

	ExtractTickers(text, writeToStocks)

	return stocks
}

func AddToIgnoredStocks(stock string) {
	delete(watchedStocks, stock)

	f, _ := os.OpenFile(config.GlobalConfig.HomePath+"/ignoredStocks.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()

	f.WriteString(stock + "\n")
}

func InitTickers() {
	ignoredStocks := make(map[string]string)
	file, err := os.Open(config.GlobalConfig.HomePath + "/ignoredStocks.txt")
	if err == nil {
		reader := bufio.NewReader(file)
		for {
			word, err := reader.ReadString('\n')
			word = strings.TrimSpace(word)
			if err != nil {
				break
			}
			ignoredStocks[word] = word
		}
	}

	watchedStocks = make(map[string]string)
	file, err = os.Open(config.GlobalConfig.HomePath + "/tickers.txt")
	if err != nil {
		log.Fatal(err)
	}

	reader := bufio.NewReader(file)
	for {
		ticker, err := reader.ReadString('\n')
		line := strings.Split(strings.TrimSpace(ticker), "~")
		if err != nil {
			break
		}

		_, ok := ignoredStocks[line[0]]
		if !ok {
			watchedStocks[line[0]] = line[1]
		} else {
			fmt.Println(line[0])
		}
	}
}
