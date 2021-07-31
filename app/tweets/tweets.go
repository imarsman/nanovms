package tweets

import (
	"context"
	_ "embed" // for Twitter bearer token
	"encoding/json"
	"fmt"
	"log"
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
	ID       string
	NextLoad time.Time
}

type authorize struct {
	Token string
}

// See https://github.com/g8rswimmer/go-twitter/tree/master/v2

// Sample:
// curl https://api.twitter.com/2/tweets/search/recent?query=from%3Atwitterdev%20new%20-is%3Aretweet&max_results=10 -H "Authorization: Bearer AAAAAAAAAAAAAAAAAAAAAF7xSAEAAAAAVQigSbjlsluePHIMttuwgzQiqWs%3D2ZrUT9ew0mE4tHzHR11muoo98A3GEIBaYiWHgBiARfL30SZKgT"

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

// GetBySearch get tweets by search term
func GetBySearch() error {

	opts := twitter.TweetRecentSearchOpts{
		Expansions:  []twitter.Expansion{twitter.Expansion(twitter.TweetFieldText)},
		TweetFields: []twitter.TweetField{twitter.TweetFieldID},
		MaxResults:  5,
	}

	tweetDictionary, err := client.TweetRecentSearch(context.Background(), "#kitten or #kittens", opts)
	if err != nil {
		return fmt.Errorf("tweet lookup error: %v", err)
	}

	enc, err := json.MarshalIndent(tweetDictionary, "", "    ")
	if err != nil {
		log.Panic(err)
	}
	fmt.Println(string(enc))

	return nil
}
