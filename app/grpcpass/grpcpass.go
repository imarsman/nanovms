package grpcpass

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/tidwall/gjson"
	"golang.org/x/net/context"
)

type XKCD struct {
	Day     int    `json:"day"`
	Month   int    `json:"month"`
	Year    int    `json:"year"`
	Number  int    `json:"number"`
	Title   string `json:"title"`
	AltText string `json:"alt"`
	Img     string `json:"img"`
}

// Server a server
type Server struct {
}

/*

Notes:
- brew install protobuf
- https://tutorialedge.net/golang/go-grpc-beginners-tutorial/

*/

// GetXKCD get an xkcd url and description
func (s *Server) GetXKCD(ctx context.Context, in *Message) (*Message, error) {
	log.Printf("Receive message body from client: %d", in.GetNumber())

	num := int(in.GetNumber())
	bytes, err := fetchXKCD(num)
	if err != nil {
		return &Message{}, err
	}

	xkcd, err := parseJSON(bytes)
	if err != nil {
		return &Message{}, err
	}

	msg := &Message{
		Number: int64(xkcd.Number),
		Img:    xkcd.Img,
		Title:  xkcd.Title,
		Alt:    xkcd.AltText,
	}

	return msg, nil
}

// fetchXKCD fetch info for a comic for a day from xkcd
func fetchRandomXKCD() ([]byte, error) {
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	return fetchXKCD(r.Intn(2000) + 1)
}

// fetchXKCD fetch info for a comic for a day from xkcd
func fetchXKCD(num int) ([]byte, error) {
	if num > 2000 {
		return []byte{}, fmt.Errorf("Invalid index %d", num)
	}

	url := "http://xkcd.com/" + fmt.Sprintf("%v", num) + "/info.0.json"
	fmt.Println(url)

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

func parseJSON(input []byte) (*XKCD, error) {
	xkcd := XKCD{}
	json := string(input)

	var res gjson.Result

	res = gjson.Get(json, "day")
	xkcd.Day = int(res.Int())

	res = gjson.Get(json, "month")
	xkcd.Month = int(res.Int())

	res = gjson.Get(json, "year")
	xkcd.Year = int(res.Int())

	res = gjson.Get(json, "num")
	xkcd.Number = int(res.Int())

	res = gjson.Get(json, "safe_title")
	xkcd.Title = res.String()

	res = gjson.Get(json, "alt")
	xkcd.AltText = res.String()

	res = gjson.Get(json, "img")
	xkcd.Img = res.String()

	return &xkcd, nil
}
