package tweets

import (
	"testing"

	"github.com/matryer/is"
)

func TestSearch(t *testing.T) {
	is := is.New(t)

	for i := 0; i < 20; i++ {
		results, err := GetTweetData()
		is.NoErr(err)
		t.Log(results)
	}
}
