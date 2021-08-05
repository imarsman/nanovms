package msg

import (
	"bytes"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"regexp"
	"strings"
	"text/template"
	"time"

	"github.com/nats-io/nats-server/v2/server"
	stand "github.com/nats-io/nats-streaming-server/server"
	"github.com/nats-io/nats.go"
)

const ( // various content types
	jsonContentType = "application/json; charset=utf-8"
	textContentType = "text/plain; charset=utf-8"
	htmlContentType = "text/html; charset=utf-8"
)

//go:embed dynamic/*
var dynamic embed.FS

// go get github.com/nats-io/nkeys/nk
// https://github.com/nats-io/nkeys/blob/master/nk/README.md

//go:embed secrets/nkeyuser.seed
var nkeyUserSeed string

//go:embed secrets/nkeyuser.pub
var nkeyUserPub string

// //go:embed dynamic/*
// var dynamic embed.FS

var templates *template.Template // templates for dynamic pages
var routeMatch *regexp.Regexp    // template route regex

// var natsConn *nats.Conn
var natsServer *server.Server

// http://api.plos.org/solr/examples/
// http://api.plos.org/search?q=title:covid
// - &start=[]

type Response struct {
	ResultSet ResultSet `json:"response"`
}

// ResultSet a list of results
type ResultSet struct {
	SearchTerm string   `json:"searchTerm"`
	NumFound   int      `json:"numFound"`
	Start      int      `json:"start"`
	Docs       []Result `json:"docs"`
}

// Result a query result
type Result struct {
	ID              string   `json:"id"`
	Title           string   `json:"title"`
	Abstract        []string `json:"abstract_primary_display"`
	Journal         string   `json:"journal"`
	Author          []string `json:"author"`
	SearchTerm      string   `json:"searchTerm"`
	Message         string   `json:"message"`
	PublicationDate string   `json:"publication_date"`
	Error           bool     `json:"error"`
}

// Payload a payload to send back to browser
type Payload struct {
	payload string
}

// NewPayload get a new result instance
func NewPayload() *Payload {
	p := Payload{}

	return &p
}

// Query a query
type Query struct {
	SearchTerm string
	Start      int
}

// NewQuery make a new query
func NewQuery(searchTerm string, start int) *Query {
	q := Query{}
	q.SearchTerm = searchTerm
	q.Start = start

	return &q
}

// NATServer get reference to NATS server
func NATServer() *server.Server {
	return natsServer
}

var funcMap = template.FuncMap{
	"StringsJoin": strings.Join, "StringsTrim": strings.TrimSpace,
}

// https://golangrepo.com/repo/nats-io-nats-go-messaging
func init() {
	// We need to convert the embed FS to an io.FS in order to work with it
	fsys := fs.FS(dynamic)
	contentDynamic, _ := fs.Sub(fsys, "dynamic")

	// Load templates by pattern into a structure for later use
	// Add in a funcion map
	var err error
	templates, err = template.New("templates").Funcs(funcMap).ParseFS(contentDynamic, "templates/*.html")
	if err != nil {
		log.Println("Cannot parse templates:", err)
		os.Exit(-1)
	}
	// templates.Funcs(template.FuncMap{"StringsJoin": strings.Join})

	// templates = templates.Funcs(template.FuncMap{"StringsJoin": strings.Join})

	// templates.Funcs(template.FuncMap{"StringsJoin": strings.Join})
	// Set up our route matching pattern
	routeMatch, err = regexp.Compile(`^\/(\w+)`)
	if err != nil {
		log.Println("Problems with regular expression:", err)
		os.Exit(-1)
	}

	// https://sourcegraph.com/github.com/nats-io/nats-server@6da5d2f4907a03c8ba26fc8b6ca2aed903ac80f8/-/blob/main.go
	// Now we want to setup the monitoring port for NATS Streaming.
	// We still need NATS Options to do so, so create NATS Options
	// using the NewNATSOptions() from the streaming server package.
	snopts := stand.NewNATSOptions()
	snopts.Port = nats.DefaultPort
	snopts.HTTPPort = 8223

	// Now run the server with the streaming and streaming/nats options.
	natsServer, err = server.NewServer(snopts)
	if err != nil {
		panic(err)
	}
}

func QueryNATS(search string) ([]byte, error) {
	natsConn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatal(err)
	}

	// Subscribe
	sub, err := natsConn.SubscribeSync("search")
	if err != nil {
		log.Fatal(err)
	}

	rs, err := fetchSearch(search)
	if err != nil {
		return nil, err
	}

	json, err := json.MarshalIndent(rs, "", "  ")
	if err != nil {
		return nil, err
	}

	natsConn.Publish("search", json)

	// Wait for a message
	msg, err := sub.NextMsg(5 * time.Second)
	if err != nil {
		log.Fatal(err)
	}

	return msg.Data, nil
}

func fetchSearch(search string) (*ResultSet, error) {
	results, err := queryAPI(search)
	if err != nil {
		return &ResultSet{}, err
	}

	// resultSet, err := ToResultSet(results)
	// if err != nil {
	// 	return &ResultSet{}, err
	// }

	response := Response{}

	err = json.Unmarshal(results, &response)
	if err != nil {
		return nil, err
	}

	response.ResultSet.SearchTerm = search

	return &response.ResultSet, nil
}

func queryAPI(search string) ([]byte, error) {
	url := "http://api.plos.org/search?q=title:" + fmt.Sprintf("%v", search) + "&fl=id,title,abstract_primary_display,journal,publication_date,author"

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

// // ToResultSet get result set from payload
// func ToResultSet(payload []byte) (ResultSet, error) {
// 	response := Response{}

// 	err := json.Unmarshal(payload, &response)
// 	if err != nil {
// 		return ResultSet{}, err
// 	}

// 	rs := response.ResultSet
// 	// fmt.Printf("RESULTSET %+v\n", response.ResultSet)
// 	return rs, nil
// }

// ToHTML process template to HTML
func ToHTML(rs *ResultSet, isErr bool) (string, error) {
	buf := new(bytes.Buffer)

	// fmt.Printf("RESULTSET! %+v\n", rs)
	page := "search.html"
	if isErr {
		page = "error.html"
	}
	if templates.Lookup(page) != nil {
		templates.ExecuteTemplate(buf, page, rs)
	} else {
		return "", errors.New("Could not find template")
	}

	return buf.String(), nil
}
