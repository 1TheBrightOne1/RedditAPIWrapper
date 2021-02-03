package wsbmonitor

import (
	"sync"
	"time"

	"github.com/1TheBrightOne1/RedditAPIWrapper/models"
)

const maxWatchTime = "1d"

type WatchedPost struct {
	Link          string
	FreshComments int
	LastScraped   time.Time
}

func NewWatchedPost(listing models.Listing) WatchedPost {
	return WatchedPost{
		Link:        listing.Data.Link,
		LastScraped: time.Now(),
	}
}

type Manager struct {
	Posts []WatchedPost
	Lock  *sync.RWMutex
}

func (m *Manager) GetFreshPost() WatchedPost {
	m.Lock.RLock()
	defer m.Lock.RUnlock()

	next := m.Posts[0]
	m.Posts = m.Posts[1:]
	m.Posts = append(m.Posts, next)

	return next
}

func (m *Manager) UpdatePost(link string, freshComments int) {
	for i, post := range m.Posts {
		if post.Link == link {
			if float64(freshComments)/time.Now().Sub(post.LastScraped).Hours() < 3.0 {
				m.Lock.Lock()
				defer m.Lock.Unlock()

				m.Posts = append(m.Posts[0:i], m.Posts[i+1:]...)
			}
			break
		}
	}
}
