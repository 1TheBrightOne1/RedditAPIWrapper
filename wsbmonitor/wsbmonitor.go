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
	"github.com/1TheBrightOne1/RedditAPIWrapper/models"
	"github.com/1TheBrightOne1/RedditAPIWrapper/oauth"
)

var (
	charsOnly   = regexp.MustCompile(`\w+`)
	knownStocks map[string]string
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
	s.getHotArticles()

	for {
		s.getNewArticles()

		go s.getUpdatedListings()
		dur, _ := time.ParseDuration("60s")
		time.Sleep(dur)
	}
}

func (s *Scraper) getHotArticles() {
	fmt.Println("Getting Hot Articles")
	resp, err := s.creds.SendRequest(api.Get_Hot("wallstreetbets", "", "", 0))

	if err != nil {
		log.Fatal(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	listings := models.NewListing(body)

	models.WalkListing(listings, s.addListingsToWatchList())
}

func (s *Scraper) getNewArticles() {
	fmt.Println("Getting New Articles")
	resp, err := s.creds.SendRequest(api.Get_New("wallstreetbets", "", s.after, 0))

	if err != nil {
		log.Fatal(err)
	}

	body, _ := ioutil.ReadAll(resp.Body)

	listings := models.NewListing(body)

	models.WalkListing(listings, s.addListingsToWatchList())
}

func (s *Scraper) getCommentsForArticle(link string) {
	fmt.Printf("Getting comments for article %s\n", link)
	resp, err := s.creds.SendRequest(api.Get_Comments(link))

	if err != nil {
		fmt.Println(err.Error())
	}

	body, _ := ioutil.ReadAll(resp.Body)

	listings := models.NewListing(body)

	models.WalkListing(listings, s.addCommentsToWatchList())
}

func (s *Scraper) getUpdatedListings() {
	for {
		post := s.watchList.getFreshPost()
		if post.id == "" {
			return
		}

		if post.id[0:2] == "t3" {
			resp, err := s.creds.SendRequest(api.Get_Comments(post.article))

			if err != nil {
				return
			}

			body, _ := ioutil.ReadAll(resp.Body)

			listing := models.NewListing(body)
			models.WalkListing(listing, s.updateArticle())
		} else if post.id[0:2] == "t1" {
			commentIds := s.watchList.comments[post.article]
			resp, err := s.creds.SendRequest(api.Get_MoreChildren(post.article, commentIds))

			if err != nil {
				return
			}

			body, _ := ioutil.ReadAll(resp.Body)

			listing := models.NewListing(body)
			models.WalkListing(listing, s.updateComment())
		}
	}
}

func (s *Scraper) addListingsToWatchList() func(models.Listing) {
	return func(listing models.Listing) {
		if listing.Kind == "t3" {
			fmt.Printf("Adding %s to watch list\n", listing.Data.Name)
			s.watchList.addToWatchList(listing.Data.Name, listing.Data.Link, extractTickers(listing.Data.Title), listing.Data.Ups)

			s.getCommentsForArticle(listing.Data.Link)
		}
	}
}

func (s *Scraper) addCommentsToWatchList() func(models.Listing) {
	return func(listing models.Listing) {
		if listing.Kind == "t1" {
			tickers := extractTickers(listing.Data.Body)
			if len(tickers) > 0 {
				fmt.Printf("Adding %s to watch list\n", listing.Data.Name)
				s.watchList.addToWatchList(listing.Data.Name, listing.Data.Link, tickers, listing.Data.Ups)
			}
		}
	}
}

func (s *Scraper) updateArticle() func(models.Listing) {
	return func(listing models.Listing) {
		if listing.Kind == "t3" {
			fmt.Printf("Updating article %s\n", listing.Data.Name)
			s.watchList.updatePost(listing.Data.Name, listing.Data.Ups)
		}
	}
}

func (s *Scraper) updateComment() func(models.Listing) {
	return func(listing models.Listing) {
		if listing.Kind == "t1" {
			fmt.Printf("Updating comment %s\n", listing.Data.Name)
			s.watchList.updatePost(listing.Data.Name, listing.Data.Ups)
		}
	}
}

func extractTickers(text string) []string {
	var tickers []string
	words := strings.Split(text, " ")

	for _, word := range words {
		matches := charsOnly.FindAllString(word, 10)

		for _, match := range matches {
			cleaned := strings.ToLower(match)
			if stock, ok := knownStocks[cleaned]; ok {
				tickers = append(tickers, stock)
			}
		}
	}

	if len(tickers) > 0 {
		fmt.Printf("Tickers found %v\n", tickers)
	}

	return tickers
}

func init() {
	common := make(map[string]string)
	file, err := os.Open("common.txt")
	if err == nil {
		reader := bufio.NewReader(file)
		for {
			word, err := reader.ReadString('\n')
			word = strings.TrimSpace(word)
			if err != nil {
				break
			}
			common[word] = word
		}
	}

	knownStocks = make(map[string]string)
	file, err = os.Open("tickers.txt")
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

		_, ok := common[line[0]]
		if !ok {
			knownStocks[line[0]] = line[1]
		} else {
			fmt.Println(line[0])
		}
	}
}
