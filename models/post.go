package models

import (
	"encoding/json"
	"log"
)

type Posts struct {
	data struct {
		children []struct {
			title string `json:"title"`
			ups   int    `json:"ups"`
			downs int    `json:"downs"`
		}
	}
}

func NewPosts(bytes []byte) *Posts {
	posts := &Posts{}
	err := json.Unmarshal(bytes, posts)
	if err != nil {
		log.Fatal(err)
	}
	return posts
}
