package tweets

import (
	"context"
	_ "embed" // for Twitter bearer token
	"fmt"
	"math/rand"
	"net/http"
	"time"

	twitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/imarsman/nanovms/app/stack"
)

// Not necessarily ideal to embed the token as it may change but this should do
// for a test.
//go:embed bearer_token.txt
var token string

var tweetStack *stack.Stack

// TweetData summary tweet data for client
type TweetData struct {
	ID         string `json:"id"`         // tweet id for client lookup
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
func NewTweetData(id string) *TweetData {
	// rand.Seed(time.Now().UnixNano())
	td := TweetData{}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(120)
	if n < 30 {
		n += 30
	}

	td.ID = id
	// next load in random number of milliseconds from 30 up to 120
	td.NextLoadMS = int(time.Duration(n) * 1000)
	fmt.Println("NextLoadMS", td.NextLoadMS)

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

	// fmt.Println("stack size", tweetStack.Size())
	if tweetStack.Empty() == false {
		v, err := tweetStack.Front()
		if err != nil {
			return TweetDataError(), err
		}
		id := fmt.Sprintf("%v", v)
		td := NewTweetData(id)
		tweetStack.Pop()

		return td, nil
	}

	fmt.Println("empty stack, refilling")

	opts := twitter.TweetRecentSearchOpts{
		TweetFields: []twitter.TweetField{twitter.TweetFieldCreatedAt, twitter.TweetFieldLanguage},
		MaxResults:  10,
	}

	recentSearchResponse, err := client.TweetRecentSearch(context.Background(), "sea otter", opts)
	if err != nil {
		// fmt.Println(tweetDictionary.Raw)
		return TweetDataError(), fmt.Errorf("tweet lookup error: %v", err)
	}

	var tdList = make([]TweetData, 0, 5)

	// Twitter errors at less than 10 results but we only want 5
	for i, v := range recentSearchResponse.Raw.Tweets {
		if i < 10 {
			tweetStack.Push(v.ID)

			// tdList = append(tdList, *td)
		} else {
			break
		}
	}

	for _, v := range tdList {
		fmt.Println(v)
	}

	v, err := tweetStack.Front()
	if err != nil {
		return TweetDataError(), err
	}
	tweetStack.Pop()
	var td = NewTweetData(fmt.Sprintf("%v", v))

	return td, nil
}
