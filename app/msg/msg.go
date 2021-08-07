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
	"net/url"
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

// Response corresponds to JSON fetched
type Response struct {
	ResultSet *ResultSet `json:"response"`
}

// ResultSet a list of results
type ResultSet struct {
	SearchTerm   string    `json:"searchTerm"`
	NumFound     int       `json:"numFound"`
	Start        int       `json:"start"`
	Next         int       `json:"next"`
	Docs         []*Result `json:"docs"`
	Error        bool      `json:"error"`
	ErrorMessage string    `json:"errormsg"`
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
	Next       int
}

// NewQuery make a new query
func NewQuery(searchTerm string, next int) *Query {
	q := Query{}
	q.SearchTerm = searchTerm
	q.Next = next

	return &q
}

// NATServer get reference to NATS server
func NATServer() *server.Server {
	return natsServer
}

var funcMap = template.FuncMap{
	"StringsJoin": strings.Join,
	"StringsTrim": strings.TrimSpace,
	"Add": func(a, b int) int {
		return a + b
	},
	"Sub": func(a, b int) int {
		return a - b
	},
	"Unescape": func(a string) string {
		a, err := url.QueryUnescape(a)
		if err != nil {
			return ""
		}
		return a
	},
}

// Set up templates and the server
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
	// https://golangrepo.com/repo/nats-io-nats-go-messaging

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

// getError simple error output
func getError(search string, errMsg string) []byte {
	rs := ResultSet{}
	rs.SearchTerm = search
	rs.Error = true
	rs.ErrorMessage = errMsg

	output, err := json.MarshalIndent(&rs, "", "  ")
	if err != nil {
		fmt.Println("Got error creating error", err)
	}

	return output
}

// Get a local connection or one to a demo for nats.io
func getConnection(isInCloud bool) (*nats.Conn, error) {
	if isInCloud {
		nc, err := nats.Connect("nats://demo.nats.io:4222", nats.Timeout(10*time.Second))
		if err != nil {
			return nil, err
		}
		return nc, nil
	}
	// "nats://0.0.0.0:4222"
	nc, err := nats.Connect("nats://0.0.0.0:4222", nats.Timeout(10*time.Second))
	if err != nil {
		return nil, err
	}
	return nc, nil
}

// QueryNATS query a nats server
func QueryNATS(search string, next int, isInCloud bool) ([]byte, error) {
	// Get a connection
	nc, err := getConnection(isInCloud)

	// Get escaped query
	search = url.QueryEscape(search)

	// To make things easier set up a JSON encoded connection
	ec, err := nats.NewEncodedConn(nc, nats.JSON_ENCODER)
	if err != nil {
		log.Fatal(err)
	}
	defer ec.Close()

	// Create a unique subject name for replies.
	uniqueReplyTo := nats.NewInbox()

	// Subscribe to a unique subject
	sub, err := nc.SubscribeSync(uniqueReplyTo)
	if err != nil {
		return getError(search, err.Error()), nil
	}

	// Get the data to publish
	rs, err := fetchSearch(search, next)
	if err != nil {
		return getError(search, err.Error()), nil
	}
	rs.Next = rs.Start + len(rs.Docs)
	for _, r := range rs.Docs {
		t, err := time.Parse("2006-01-02T15:04:05Z", r.PublicationDate)
		if err != nil {

		}
		r.PublicationDate = t.Format("2006-01-02")
	}

	if len(rs.Docs) == 0 {
		return getError(search, "Nothing found for search \""+search+"\""), nil
	}

	// Publish the resulting object, which will be turned into JSON
	ec.Publish(uniqueReplyTo, rs)

	// Wait for a message - blocking
	msg, err := sub.NextMsg(5 * time.Second)
	if err != nil {
		return getError(search, err.Error()), nil
	}

	return msg.Data, nil
}

func fetchSearch(search string, next int) (*ResultSet, error) {
	results, err := queryAPI(search, next)
	if err != nil {
		return &ResultSet{}, err
	}

	response := Response{}

	err = json.Unmarshal(results, &response)
	if err != nil {
		return nil, err
	}

	response.ResultSet.SearchTerm = search
	response.ResultSet.Next = response.ResultSet.Start + len(response.ResultSet.Docs)

	return response.ResultSet, nil
}

// queryAPI query the PLOS JSON api
func queryAPI(search string, start int) ([]byte, error) {
	search = strings.TrimSpace(search)
	search, _ = url.QueryUnescape(search)
	var u string
	u = "http://api.plos.org/search?"

	// https://www.crossref.org/blog/dois-and-matching-regular-expressions/
	// ^10.\d{4,9}/[-._;()/:A-Z0-9]+$
	isLinkSearch, _ := regexp.MatchString(`^10\.\d{4,9}/[-\._;\(\)\/\:a-zA-Z0-9]+$`, search)
	// isLinkSearch, _ := regexp.MatchString(`^\d+\.\d+\/journal\.[^\.]+\.\d+$`, search)
	// fmt.Println("search", search, "matched", isLinkSearch)

	if isLinkSearch {
		u = u + "q=id:\"" + fmt.Sprintf("%v", search) + "\"&fl=id,title,abstract_primary_display,journal,publication_date,author&start=" + fmt.Sprintf("%d", start)
	} else {
		search = url.QueryEscape(search)
		u = u + "q=title:" + fmt.Sprintf("%v", search) + "&fl=id,title,abstract_primary_display,journal,publication_date,author&start=" + fmt.Sprintf("%d", start)
	}

	resp, err := http.Get(u)
	if err != nil {
		return []byte{}, err
	}

	bytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []byte{}, err
	}

	return bytes, nil
}

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
