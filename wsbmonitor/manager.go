package wsbmonitor

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/1TheBrightOne1/RedditAPIWrapper/metrics"
	"github.com/1TheBrightOne1/RedditAPIWrapper/models"
)

const maxWatchTime = "2d"

type watchedItem struct {
	id           string
	article      string
	stocks       map[string]int
	upvotes      int
	lastScraped  time.Time
	firstScraped time.Time
}

func newWatchedItem(id, article string, stocks map[string]int) watchedItem {
	w := watchedItem{
		id:           id,
		article:      article,
		stocks:       stocks,
		firstScraped: time.Now(),
		lastScraped:  time.Now(),
	}

	for stock, score := range stocks {
		if score > 0 {
			metrics.Counter.WithLabelValues(stock).Add(float64(score))
		}
	}

	return w
}

func (w *watchedItem) update(updatedStocks map[string]int) bool {
	totalIncrease := 0
	for stock, score := range updatedStocks {
		if scoreIncrease := score - w.stocks[stock]; scoreIncrease > 0 {
			metrics.Counter.WithLabelValues(stock).Add(float64(scoreIncrease))
			totalIncrease += scoreIncrease
		}
	}

	ratio := float64(totalIncrease) / time.Now().Sub(w.lastScraped).Hours()
	metrics.UpvotesPerHour.Observe(ratio)
	if ratio < 30.0 {
		return false
	}

	w.lastScraped = time.Now()
	w.stocks = updatedStocks
	return true
}

type watchList struct {
	posts []watchedItem
	lock  *sync.RWMutex
}

func newWatchList() *watchList {
	return &watchList{
		posts: make([]watchedItem, 10),
		lock:  &sync.RWMutex{},
	}
}

func (m *watchList) GetArticlesByStock(stock string) {
	f, _ := os.Create(fmt.Sprintf("%s.requested", stock))
	defer f.Close()
	for _, watched := range m.posts {
		for s := range watched.stocks {
			if stock == s {
				fmt.Fprintf(f, "%s\n%v\n", watched.article, watched.stocks)
			}
		}
	}
}

func (m *watchList) addToWatchList(id, permalink string, stocks map[string]int) {
	m.lock.Lock()
	defer m.lock.Unlock()
	m.posts = append(m.posts, newWatchedItem(id, permalink, stocks))
}

func (m *watchList) getFreshPost() watchedItem {
	m.lock.Lock()
	defer m.lock.Unlock()

	for i := 0; i < len(m.posts); i++ {
		next := m.posts[0]
		m.posts = m.posts[1:]
		m.posts = append(m.posts, next)

		if next.lastScraped.Sub(time.Now()).Minutes() > 5 {
			return next
		}
	}

	return watchedItem{}
}

func (m *watchList) updatePost(listing models.Listing, stocks map[string]int) {
	m.lock.RLock()
	for i, post := range m.posts {
		if post.id == listing.Data.Name {
			if !post.update(stocks) {
				m.lock.RUnlock()
				m.lock.Lock()
				defer m.lock.Unlock()
				m.posts = append(m.posts[0:i], m.posts[i+1:]...)
				return
			}
			return
		}
	}

	m.lock.RUnlock()
	m.addToWatchList(listing.Data.Name, listing.Data.Link, stocks)
}
