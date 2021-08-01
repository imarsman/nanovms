package grpcpass

import (
	"log"

	"github.com/tidwall/gjson"
	"golang.org/x/net/context"
)

type XKCD struct {
	Day   int    `json:"day"`
	Month int    `json:"month"`
	Year  int    `json:"yeare"`
	Num   int    `json:""num"`
	Title string `json:"title"`
	Desc  string `json:"desc"`
	Img   string `json:"img"`
}

// Server a server
type Server struct {
}

func parseJSON(input []byte) (*XKCD, error) {
	xkcd := XKCD{}

	json := string(input)

	res := gjson.Get(json, "day")
	xkcd.Day = int(res.Int())

	res = gjson.Get(json, "month")
	xkcd.Month = int(res.Int())

	res = gjson.Get(json, "year")
	xkcd.Year = int(res.Int())

	res = gjson.Get(json, "num")
	xkcd.Num = int(res.Int())

	res = gjson.Get(json, "title")
	xkcd.Title = res.String()

	res = gjson.Get(json, "img")
	xkcd.Img = res.String()

	return &xkcd, nil
}

/*

Notes:
- brew install protobuf
- https://tutorialedge.net/golang/go-grpc-beginners-tutorial/

*/

// GetXKCD get an xkcd url and description
func (s *Server) GetXKCD(ctx context.Context, in *Message) (*Message, error) {
	log.Printf("Receive message body from client: %s", in.Desc)
	return &Message{
		Img:  "https://imgs.xkcd.com/comics/mine_captcha.png",
		Desc: "Mine Captcha",
		Alt:  "This data is actually going into improving our self-driving car project, so hurry up--it's almost at the minefield.",
	}, nil
}

func FetchXKCD() []byte {

}
