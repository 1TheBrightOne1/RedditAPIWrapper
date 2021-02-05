package wsbmonitor

import (
	"sync"
	"time"

	"github.com/1TheBrightOne1/RedditAPIWrapper/metrics"
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

func newWatchedItem(id, article string, stocks []string, upvotes int) watchedItem {
	w := watchedItem{
		id:           id,
		article:      article,
		stocks:       make(map[string]int),
		firstScraped: time.Now(),
		lastScraped:  time.Now(),
		upvotes:      upvotes,
	}

	if upvotes < 0 {
		upvotes = 0
	}

	for _, stock := range stocks {
		w.stocks[stock] = 1

		metrics.Counter.WithLabelValues(stock).Add(float64(upvotes))
	}

	return w
}

func (w *watchedItem) update(newUpvotes int) bool {
	if newUpvotes-w.upvotes > 0 {
		for stock, _ := range w.stocks {
			metrics.Counter.WithLabelValues(stock).Add(float64(newUpvotes - w.upvotes))
		}
	}
	w.upvotes = newUpvotes

	ratio := float64(newUpvotes) / time.Now().Sub(w.lastScraped).Hours()
	metrics.UpvotesPerHour.Observe(ratio)
	if ratio < 3.0 {
		return false
	}

	//TODO: Add logic for max expiration
	return true
}

type watchList struct {
	posts    []watchedItem
	comments map[string][]string //key=articleID, value is a slice of strings with other watched comments so they can be batch fetched
	lock     *sync.RWMutex
}

func newWatchList() *watchList {
	return &watchList{
		posts:    make([]watchedItem, 10),
		lock:     &sync.RWMutex{},
		comments: make(map[string][]string),
	}
}

func (m *watchList) addToWatchList(id, permalink string, stocks []string, upvotes int) {
	if id[:2] == "t1" {
		permalink := permalink[:len(permalink)-len(id)]
		commentSlice := m.comments[permalink]
		m.comments[permalink] = append(commentSlice, id)
	}
	m.posts = append(m.posts, newWatchedItem(id, permalink, stocks, upvotes))
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

func (m *watchList) updatePost(id string, upvotes int) {
	m.lock.RLock()
	for i, post := range m.posts {
		if post.id == id {
			if !post.update(upvotes) {
				m.lock.RUnlock()
				m.lock.Lock()
				defer m.lock.Unlock()
				m.posts = append(m.posts[0:i], m.posts[i+1:]...)
				m.pruneWatchList(post)
				return
			}
			break
		}
	}
	m.lock.RUnlock()
}

func (m *watchList) pruneWatchList(post watchedItem) {
	if post.id[0:2] == "t3" {
		comments := m.comments[post.article]
		m.comments[post.article] = []string{}

		for _, comment := range comments {
			for i, post := range m.posts {
				if comment == post.id {
					m.posts = append(m.posts[0:i], m.posts[i+1:]...)
				}
			}
		}
	} else if post.id[0:2] == "t1" {
		for i, comment := range m.comments[post.article] {
			if post.id == comment {
				m.comments[post.article] = append(m.comments[post.article][0:i], m.comments[post.article][i+1:]...)
			}
		}
	}
}
