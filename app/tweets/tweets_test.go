package tweets

import (
	"testing"

	"github.com/matryer/is"
)

func TestSearch(t *testing.T) {
	is := is.New(t)

	results, err := GetBySearch()
	is.NoErr(err)
	t.Log(string(results))
}
