package main

import (
	"io/fs"
	"log"
	"net/http"
	"os"
	"regexp"
	"sync/atomic"
	"text/template"
	"time"
)

var templates *template.Template // templates for dynamic pages
var routeMatch *regexp.Regexp    // template route regex
var count uint64                 // page hit counter
var startTime *time.Time         // start time of server running

const ( // various content types
	jsonContentType = "application/json; charset=utf-8"
	textContentType = "text/plain; charset=utf-8"
	htmlContentType = "text/html; charset=utf-8"
)

// PageData data for a page's templates
// Capitalized because it is used in templates and needs to be public
type PageData struct {
	Timestamp time.Time
	LoadStart time.Time
	LoadTime  time.Duration
	PageLoads uint64
	Uptime    time.Duration
}

// newPageData create a pointer to a new PageData struct instance
func newPageData() *PageData {
	pd := PageData{}
	pd.LoadStart = time.Now()

	return &pd
}

// finalize finish off page info that is time specific
func (pd *PageData) finalize() {
	pd.LoadTime = time.Since(pd.LoadStart)
	pd.PageLoads = counterIncrement()
	pd.Uptime = time.Since(*startTime).Round(time.Second)
}

// counterIncrement a simple increment of page hit count
func counterIncrement() uint64 {
	return atomic.AddUint64(&count, 1)
}

// init initialize counter and parse templates.
func init() {
	t := time.Now()
	startTime = &t

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
}

// templatePageHandler use template collection to produce output
func templatePageHandler(w http.ResponseWriter, r *http.Request) {
	pd := newPageData()

	matches := routeMatch.FindStringSubmatch(r.URL.Path)
	if len(matches) >= 1 {
		page := matches[1] + ".html"
		if templates.Lookup(page) != nil {
			w.Header().Add("Content-Type", htmlContentType)
			w.WriteHeader(http.StatusOK)
			pd.finalize()
			templates.ExecuteTemplate(w, page, pd)
			return
		}
	} else if r.URL.Path == "/" {
		w.Header().Add("Content-Type", htmlContentType)
		w.WriteHeader(http.StatusOK)
		pd.finalize()
		templates.ExecuteTemplate(w, "index.html", pd)
		return
	}
	w.WriteHeader(http.StatusNotFound)
	w.Header().Add("Content-Type", textContentType)
	w.Write([]byte("NOT FOUND"))
}

// getTransactionsHandler get list of transactions
func getTransactionsHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", jsonContentType)
	transactionList, err := readTransactions()
	if err != nil { // simulate error getting data
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	sortDescendingPostTimestamp(&transactionList)

	// obscured, err := obscured(transactions)
	transactionList, err = obscureTransactionID(transactionList) // allow for error
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json, err := toJSON(transactionList)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(json))
}
