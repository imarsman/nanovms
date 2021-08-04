package msg

import (
	"embed"
	"io/fs"
	"log"
	"os"
	"regexp"
	"text/template"

	"github.com/nats-io/nats.go"
)

//go:embed dynamic/*
var dynamic embed.FS

var templates *template.Template // templates for dynamic pages
var routeMatch *regexp.Regexp    // template route regex

var nc *nats.Conn

// Query a query
type Query struct {
	SearchTerm string
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
}

func fetchSearch(search string) string {
	return "hello"
}
