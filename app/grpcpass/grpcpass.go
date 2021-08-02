package grpcpass

import (
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
	"golang.org/x/net/context"
)

// XKCD a struct to contain the elements of an xkcd image to be used by the app
type XKCD struct {
	Number     int    `json:"number"`
	Date       string `json:"date"`
	Title      string `json:"title"`
	AltText    string `json:"alt"`
	Img        string `json:"img"`
	NextLoadMS int    `json:"nextloadms"` // random next load time
}

// NewXKCD get a reference to a new XKCD struct
func NewXKCD() *XKCD {
	xkcd := XKCD{}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	n := r.Intn(300)
	if n < 90 {
		n += 90
	}

	// next load in random number of milliseconds from 30 up to 120
	xkcd.NextLoadMS = int(n * int(time.Second) / 1000000)

	return &xkcd
}

// MessageNumber input for requests for a message
// type MessageNumber struct {
// 	Number int `json:"number"`
// }

// XKCDService a server
type XKCDService struct {
	UnimplementedXKCDServiceServer
}

/*

Notes:
- brew install protobuf
- https://tutorialedge.net/golang/go-grpc-beginners-tutorial/

*/

// GetXKCD get an xkcd url and description
// Presumably this will be used by the GRPC infrastructure
// func (s *XKCDService) GetXKCD(ctx context.Context, in *MessageNumber, opts ...grpc.CallOption) (*Message, error) {
func (s *XKCDService) GetXKCD(ctx context.Context, in *MessageNumber) (*Message, error) {
	var bytes []byte
	var err error

	num := int(in.GetNumber())
	if num == 0 {
		bytes, err = FetchRandomXKCD()
		if err != nil {
			return &Message{}, err
		}
	} else {
		bytes, err = FetchXKCD(num)
		if err != nil {
			return &Message{}, err
		}
	}

	xkcd, err := ParseXKCDJSON(bytes)
	if err != nil {
		return &Message{}, err
	}

	msg := &Message{
		Number: int64(xkcd.Number),
		Date:   xkcd.Date,
		Img:    xkcd.Img,
		Title:  xkcd.Title,
		Alt:    xkcd.AltText,
	}

	return msg, nil
}

/*
	The xkcd illustrator has kindly made his many years' worth of daily cartoons
	available using a simple JSON GET endpoint.

	See: https://xkcd.com/json.html
*/

// FetchRandomXKCD fetch info for a comic for a day from xkcd
func FetchRandomXKCD() ([]byte, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return FetchXKCD(r.Intn(2000) + 1)
}

// FetchXKCD fetch info for a comic for a day from xkcd
func FetchXKCD(num int) ([]byte, error) {
	if num > 2000 {
		return []byte{}, fmt.Errorf("Invalid index %d", num)
	}

	url := "http://xkcd.com/" + fmt.Sprintf("%v", num) + "/info.0.json"

	resp, err := http.Get(url)
	if err != nil {
		return []byte{}, err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

// ParseXKCDJSON rather than use a map[string]interface{} use a library that handles
// JSON Path and type conversion.
func ParseXKCDJSON(input []byte) (*XKCD, error) {
	parsed := gjson.Parse(string(input))

	d := int(parsed.Get("day").Int())
	m := int(parsed.Get("month").Int())
	y := int(parsed.Get("year").Int())

	date := time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC).Format("2006-01-02")

	xkcd := NewXKCD()
	xkcd.Date = date
	xkcd.Number = int(parsed.Get("num").Int())
	xkcd.Title = parsed.Get("safe_title").String()
	xkcd.AltText = parsed.Get("alt").String()
	xkcd.Img = parsed.Get("img").String()

	// xkcd := XKCD{
	// 	Date:    date,
	// 	Number:  int(parsed.Get("num").Int()),
	// 	Title:   parsed.Get("safe_title").String(),
	// 	AltText: parsed.Get("alt").String(),
	// 	Img:     parsed.Get("img").String(),
	// }

	// Maybe the safe_title is not there sometimes
	if xkcd.Title == "" {
		xkcd.Title = parsed.Get("title").String()
	}

	return xkcd, nil
}
