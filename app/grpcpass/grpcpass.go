package grpcpass

import (
	"log"

	"golang.org/x/net/context"
)

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
	return &Message{Desc: "Hello From the Server!"}, nil
}
