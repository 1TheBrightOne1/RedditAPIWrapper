package models

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
)

type Posts struct {
	Data struct {
		Children []struct {
			Data struct {
				Title string `json:"title"`
				Ups   int    `json:"ups"`
				Downs int    `json:"downs"`
				Name  string `json:"name"`
			} `json:"data"`
		} `json:"children"`
	} `json:"data"`
}

func NewPosts(bytes []byte) *Posts {
	regex := regexp.MustCompile(`"selftext_html":.*?, "`)
	bytes = regex.ReplaceAll(bytes, []byte(`"`))

	file, _ := os.Create("cleaned.json")
	fmt.Fprintf(file, string(bytes))

	posts := &Posts{}
	err := json.Unmarshal(bytes, posts)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("Found %d posts\n", len(posts.Data.Children))
	return posts
}
