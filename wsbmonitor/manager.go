package wsbmonitor

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"sync"
	"time"

	"github.com/1TheBrightOne1/RedditAPIWrapper/metrics"
	"github.com/1TheBrightOne1/RedditAPIWrapper/models"
)

const (
	maxWatchTime  = "2d"
	watchListPath = "/var/stonks/watchList.json"
)

type watchedItem struct {
	Id           string         `json:"id"`
	Article      string         `json:"article"`
	Stocks       map[string]int `json:"stocks"`
	LastScraped  time.Time      `json:"lastScraped"`
	FirstScraped time.Time      `json:"firstScraped"`
}

func newWatchedItem(id, article string, stocks map[string]int) watchedItem {
	w := watchedItem{
		Id:           id,
		Article:      article,
		Stocks:       stocks,
		FirstScraped: time.Now(),
		LastScraped:  time.Now(),
	}

	for stock, score := range stocks {
		if score > 0 {
			metrics.Counter.WithLabelValues(stock).Add(float64(score))

			fmt.Printf("Increasing score for %s by %d\n", stock, score)
		}
	}

	return w
}

func (w *watchedItem) update(updatedStocks map[string]int) bool {
	totalIncrease := 0
	for stock, score := range updatedStocks {
		if scoreIncrease := score - w.Stocks[stock]; scoreIncrease > 0 {
			metrics.Counter.WithLabelValues(stock).Add(float64(scoreIncrease))
			totalIncrease += scoreIncrease

			fmt.Printf("Increasing score for %s by %d\n", stock, scoreIncrease)
		}
	}

	ratio := float64(totalIncrease) / time.Now().Sub(w.LastScraped).Hours()
	metrics.UpvotesPerHour.Observe(ratio)
	if ratio < 10.0 {
		return false
	}

	w.LastScraped = time.Now()
	w.Stocks = updatedStocks
	return true
}

type watchList struct {
	Posts []watchedItem `json:"posts"`
	lock  *sync.RWMutex `json:"-"`
}

func newWatchList() *watchList {
	f, err := os.Open(watchListPath)
	if err != nil {
		fmt.Println("Can't load watch list from file")
		fmt.Println(err)

		return &watchList{
			Posts: make([]watchedItem, 10),
			lock:  &sync.RWMutex{},
		}
	}

	bytes, _ := ioutil.ReadAll(f)

	w := &watchList{}

	err = json.Unmarshal(bytes, w)
	if err != nil {
		fmt.Println("Can't load watch list from file")
		fmt.Println(err)

		return &watchList{
			Posts: make([]watchedItem, 10),
			lock:  &sync.RWMutex{},
		}
	}

	w.lock = &sync.RWMutex{}

	return w
}

func (m *watchList) GetArticlesByStock(stock string) {
	f, _ := os.Create(fmt.Sprintf("/var/stonks/%s.requested", stock))
	defer f.Close()
	for _, watched := range m.Posts {
		for s := range watched.Stocks {
			if stock == s {
				fmt.Fprintf(f, "%s\n%v\n", watched.Article, watched.Stocks)
			}
		}
	}
}

func (m *watchList) addToWatchList(id, permalink string, stocks map[string]int) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.Posts = append(m.Posts, newWatchedItem(id, permalink, stocks))
	m.persistToFile()
}

func (m *watchList) getFreshPost() watchedItem {
	m.lock.Lock()
	defer m.lock.Unlock()

	for i := 0; i < len(m.Posts); i++ {
		next := m.Posts[0]
		m.Posts = m.Posts[1:]
		m.Posts = append(m.Posts, next)

		if next.LastScraped.Sub(time.Now()).Hours() > 2 {
			return next
		}
	}

	return watchedItem{}
}

func (m *watchList) updatePost(listing models.Listing, stocks map[string]int) {
	m.lock.RLock()
	for i, post := range m.Posts {
		if post.Id == listing.Data.Name {
			if !post.update(stocks) {
				m.lock.RUnlock()
				m.lock.Lock()
				defer m.lock.Unlock()
				m.Posts = append(m.Posts[0:i], m.Posts[i+1:]...)
				return
			}
			m.persistToFile()
			m.lock.RUnlock()
			return
		}
	}

	m.lock.RUnlock()
	m.addToWatchList(listing.Data.Name, listing.Data.Link, stocks)
}

func (m *watchList) persistToFile() {
	f, err := os.Create(watchListPath)
	defer f.Close()
	if err != nil {
		fmt.Println("Cannot persist watch list to file")
		fmt.Println(err)
	}

	bytes, err := json.Marshal(m)

	if err != nil {
		fmt.Println("Cannot persist watch list to file")
		fmt.Println(err)
	}

	fmt.Fprintf(f, "%s", string(bytes))
}
