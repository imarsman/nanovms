package msg

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"regexp"
	"text/template"

	"github.com/nats-io/nats-server/v2/server"
	stand "github.com/nats-io/nats-streaming-server/server"
	"github.com/nats-io/nats.go"
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

var nc *nats.Conn
var natsServer *server.Server

// http://api.plos.org/solr/examples/
// http://api.plos.org/search?q=title:covid
// - &start=[]

// ResultSet a list of results
type ResultSet struct {
	NumFound int      `json:"numFound"`
	Start    int      `json:"start"`
	Docs     []Result `json:"docs"`
}

// Result a query result
type Result struct {
	ID         int    `json:"id"`
	Abstract   string `json:"abstract"`
	Journal    string `json:"journal"`
	SearchTerm string `json:"searchTerm"`
	Message    string `json:"message"`
	Error      bool   `json:"error"`
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

// https://golangrepo.com/repo/nats-io-nats-go-messaging
func init() {
	// We need to convert the embed FS to an io.FS in order to work with it
	fsys := fs.FS(dynamic)
	contentDynamic, _ := fs.Sub(fsys, "dynamic")

	// Load templates by pattern into a structure for later use
	var err error
	templates, err = template.ParseFS(contentDynamic, "templates/*.html")
	if err != nil {
		log.Println("Cannot parse templates:", err)
		os.Exit(-1)
	}
	// Set up our route matching pattern
	routeMatch, err = regexp.Compile(`^\/(\w+)`)
	if err != nil {
		log.Println("Problems with regular expression:", err)
		os.Exit(-1)
	}

	nc, err = nats.Connect(nats.DefaultURL)
	if err != nil {

	}

	// Simple Async Subscriber
	nc.Subscribe("search", func(m *nats.Msg) {
		sample := fetchSearch(string(m.Data))
		m.Respond([]byte(sample))
	})

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

func fetchSearch(search string) string {
	return "hello"
}
