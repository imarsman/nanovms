package grpcpass

import (
	"testing"

	"github.com/matryer/is"
)

// TestCall test call for numbered cartoon
func TestCall(t *testing.T) {
	is := is.New(t)

	bytes, err := FetchXKCD(1001)
	is.NoErr(err)
	is.True(len(bytes) > 0)

	t.Log(string(bytes))

	xkcd, err := ParseXKCDJSON(bytes)
	is.NoErr(err)

	t.Logf("%+v", xkcd)
}

// TestCallRandom test call for random cartoon
func TestCallRandom(t *testing.T) {
	is := is.New(t)

	bytes, err := FetchRandomXKCD()
	is.NoErr(err)
	is.True(len(bytes) > 0)

	t.Log(string(bytes))

	xkcd, err := ParseXKCDJSON(bytes)
	is.NoErr(err)

	t.Logf("%+v", xkcd)
}
