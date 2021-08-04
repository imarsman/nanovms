package grpcpass

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/imarsman/nanovms/app/creds"
	"github.com/tidwall/gjson"
	"golang.org/x/net/context"
	"google.golang.org/grpc"
)

const ( // various content types
	jsonContentType = "application/json; charset=utf-8"
	textContentType = "text/plain; charset=utf-8"
	htmlContentType = "text/html; charset=utf-8"
)

var grpcServer *grpc.Server

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

// GRPCServer get GRPC server
func GRPCServer() *grpc.Server {
	return grpcServer
}

func init() {
	// https://grpc.io/docs/languages/go/basics/
	// https://github.com/grpc/grpc-go/tree/master/examples
	// var opts []grpc.ServerOption
	grpcServer = grpc.NewServer(grpc.Creds(*creds.TransportCredentials()))
	// grpcServer = grpc.NewServer()

	RegisterXKCDServiceServer(grpcServer, &XKCDService{})
	fmt.Printf("grpc server: %+v\n", grpcServer.GetServiceInfo())
}

// XkcdHandler handler for XKCD data
func XkcdHandler(w http.ResponseWriter, r *http.Request) {
	// serverAddr := "localhost:9000"
	serverAddr := "[::1]:5222"

	var opts []grpc.DialOption

	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	// Connect with credentials
	// Currently trying only to use TLS to allow GCP to permit the connection
	conn, err := grpc.Dial(serverAddr, grpc.WithTransportCredentials(*creds.ClientCredentials()))
	if err != nil {
		log.Fatalf("did not connect: %v", err)
	}
	client := NewXKCDServiceClient(conn)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	number := MessageNumber{}
	number.Number = 0

	callOption := grpc.MaxCallRecvMsgSize(5000)
	message, err := client.GetXKCD(ctx, &number, callOption)
	if err != nil {
		log.Fatalf("%v.GetXKCD(_) = _, %v: ", client, err)
	}

	xkcd := NewXKCD()
	xkcd.Number = int(message.GetNumber())
	xkcd.Img = message.GetImg()
	xkcd.Date = message.Date
	xkcd.Title = message.GetTitle()
	xkcd.AltText = message.Alt

	json, err := json.MarshalIndent(&xkcd, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", jsonContentType)
	w.Write(json)
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
