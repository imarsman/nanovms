package grpcpass

import (
	"log"

	"golang.org/x/net/context"
)

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
	log.Printf("Receive message body from client: %s", in.Desc)
	return &Message{
		Img:  "https://imgs.xkcd.com/comics/mine_captcha.png",
		Desc: "Mine Captcha",
		Alt:  "This data is actually going into improving our self-driving car project, so hurry up--it's almost at the minefield.",
	}, nil
}
