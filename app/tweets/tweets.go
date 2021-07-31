package tweets

import (
	"context"
	_ "embed" // for Twitter bearer token
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"time"

	twitter "github.com/g8rswimmer/go-twitter/v2"
)

// Not necessarily ideal to embed the token as it may change but this should do
// for a test.
//go:embed bearer_token.txt
var token string

// TweetData summary tweet data for client
type TweetData struct {
	ID         string `json:"id"`         // tweet id for client lookup
	NextLoadMS int    `json:"nextloadms"` // random next load time
}

type authorize struct {
	Token string
}

// See https://github.com/g8rswimmer/go-twitter/tree/master/v2

// Sample:
// curl
// https://api.twitter.com/2/tweets/search/recent?query=from%3Atwitterdev%20new%20-is%3Aretweet&max_results=10
// -H

// NewTweetData get a prepared new tweet data
func NewTweetData(id string) *TweetData {
	td := TweetData{}

	n := rand.Intn(121)
	if n < 30 {
		n += 30
	}

	td.ID = id
	// next load in random number of milliseconds from 30 up to 120
	td.NextLoadMS = int(time.Duration(n) / time.Millisecond)

	return &td
}

func jsTimeSting(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000000000 -0700 MST")
}

func (a authorize) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

var client *twitter.Client

func init() {
	client = &twitter.Client{
		Authorizer: authorize{
			Token: token,
		},
		Client: http.DefaultClient,
		Host:   "https://api.twitter.com",
	}
}

// curl "https://api.twitter.com/2/tweets/search/recent?query=text=#kittens&max_results=10" -H "Authorization: Bearer "

// GetBySearch get tweets by search term
func GetBySearch() ([]byte, error) {

	opts := twitter.TweetRecentSearchOpts{
		// Expansions:  []twitter.Expansion{twitter.Expansion(twitter.TweetFieldText)},
		// TweetFields: []twitter.TweetField{twitter.TweetFieldID},
		TweetFields: []twitter.TweetField{twitter.TweetFieldCreatedAt, twitter.TweetFieldLanguage},
		MaxResults:  10,
	}

	recentSearchResponse, err := client.TweetRecentSearch(context.Background(), "#kitten", opts)
	if err != nil {
		// fmt.Println(tweetDictionary.Raw)
		return nil, fmt.Errorf("tweet lookup error: %v", err)
	}

	enc, err := json.MarshalIndent(recentSearchResponse, "", "    ")
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(string(enc))

	return enc, nil
}
