package tweets

import (
	"context"
	_ "embed" // for Twitter bearer token
	"fmt"
	"math/rand"
	"net/http"
	"sync"
	"time"

	twitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/imarsman/nanovms/app/stack"
)

// Not necessarily ideal to embed the token as it may change but this should do
// for a test.
//go:embed bearer_token.txt
var token string

var tweetStack *stack.Stack
var mu *sync.Mutex

// TweetData summary tweet data for client
type TweetData struct {
	TweetID    string `json:"tweetid"`    // tweet id for client lookup
	NextLoadMS int    `json:"nextloadms"` // random next load time
	Error      string `json:"error"`
}

// TweetDataError get a tweet data instance with an error message
func TweetDataError() *TweetData {
	td := TweetData{}
	td.Error = "Problem loading tweet"

	return &td
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
func NewTweetData(tweetID string) *TweetData {
	// rand.Seed(time.Now().UnixNano())
	td := TweetData{}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(60)
	if n < 15 {
		n += 5
	}

	td.TweetID = tweetID
	// next load in random number of milliseconds from 30 up to 120
	td.NextLoadMS = int(time.Duration(n) * 1000)
	// fmt.Println("NextLoadMS", td.NextLoadMS)

	return &td
}

func jsTimeSting(t time.Time) string {
	return t.Format("2006-01-02 15:04:05.000000000 -0700 MST")
}

func (a authorize) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

var client *twitter.Client

// var r *rand.Rand

func init() {
	mu = &sync.Mutex{}

	tweetStack = stack.NewStack()
	// rand.Seed(time.Now().UnixNano())
	client = &twitter.Client{
		Authorizer: authorize{
			Token: token,
		},
		Client: http.DefaultClient,
		Host:   "https://api.twitter.com",
	}
}

// curl "https://api.twitter.com/2/tweets/search/recent?query=text=#kittens&max_results=10" -H "Authorization: Bearer "

// GetTweetData get a new TweetData item. Fetch 10 at a time and cache them,
// giving out one at a time until the cache stack is empty, then reload.
func GetTweetData() (*TweetData, error) {
	mu.Lock()
	defer mu.Unlock()

	// fmt.Println("getting tweet data")
	var td *TweetData
	if tweetStack.Empty() == false {
		v, err := tweetStack.Front()
		if err != nil {
			return TweetDataError(), err
		}
		id := fmt.Sprintf("%v", v)
		td = NewTweetData(id)
		tweetStack.Pop()

		return td, nil
	}

	opts := twitter.TweetRecentSearchOpts{
		TweetFields: []twitter.TweetField{twitter.TweetFieldCreatedAt, twitter.TweetFieldLanguage},
		MaxResults:  50,
	}

	recentSearchResponse, err := client.TweetRecentSearch(context.Background(), "\"sea otter\"", opts)
	if err != nil {
		return TweetDataError(), fmt.Errorf("tweet lookup error: %v", err)
	}

	chosen := make(map[int]int)
	count := 0
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// Twitter errors at less than 10 results but we only want 5
	for {
		count++
		offset := r.Intn(50)
		// Try again if we already have included one
		if _, ok := chosen[offset]; ok {
			continue
		} else {
			if count < 20 {
				v := recentSearchResponse.Raw.Tweets[offset]
				count++
				if count < 10 {
					tweetStack.Push(v.ID)
				} else {
					break
				}
			} else {
				break
			}
		}
	}

	v, err := tweetStack.Front()
	if err != nil {
		return TweetDataError(), err
	}
	tweetStack.Pop()
	td = NewTweetData(fmt.Sprintf("%v", v))

	return td, nil
}
