package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
)

type Listing struct {
	Kind string `json:"kind"`
	Data struct {
		Children ChildrenWrapper `json:"children"`
		Title    string          `json:"title"`
		Ups      int             `json:"ups"`
		Downs    int             `json:"downs"`
		Name     string          `json:"name"`
		Body     string          `json:"body"`
		Replies  RepliesWrapper  `json:"replies"`
		Link     string          `json:"permalink"`
	} `json:"data"`
}

type RepliesWrapper struct {
	Replies []Listing
}

func (r *RepliesWrapper) UnmarshalJSON(bytes []byte) error {
	var replies []Listing
	err := json.Unmarshal(bytes, &replies)
	if err == nil {
		r.Replies = replies
	}
	return nil
}

type ChildrenWrapper struct {
	Listings     []Listing
	MoreComments []string
}

func (c *ChildrenWrapper) UnmarshalJSON(bytes []byte) error {
	var more []string
	var listings []Listing
	if err := json.Unmarshal(bytes, &more); err == nil {
		c.MoreComments = more
	} else if err := json.Unmarshal(bytes, &listings); err == nil {
		c.Listings = listings
	} else {
		return errors.New("unable to unmarshal ChildrenWrapper " + string(bytes))
	}

	return nil
}

func NewListing(bytes []byte) []Listing {
	html := regexp.MustCompile(`"selftext_html":.*?, "`)
	bytes = html.ReplaceAll(bytes, []byte(`"`))

	missing := regexp.MustCompile(`\(MISSING\)`)
	bytes = missing.ReplaceAll(bytes, []byte(""))

	file, _ := os.Create("cleaned.json")
	fmt.Fprintf(file, string(bytes))

	var listingSlice []Listing
	err := json.Unmarshal(bytes, &listingSlice)
	if err == nil {
		return listingSlice
	}

	var listing Listing
	err = json.Unmarshal(bytes, &listing)
	if err != nil {
		log.Fatal(err)
	}

	return []Listing{listing}
}
