package handlers

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync/atomic"
	"text/template"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"github.com/nats-io/nats.go"
	cache "github.com/patrickmn/go-cache"

	// "github.com/imarsman/nanovms/app"

	"github.com/imarsman/nanovms/app/grpcpass"
	"github.com/imarsman/nanovms/app/msg"
	"github.com/imarsman/nanovms/app/tweets"
)

// //go:embed secrets/servercert.pem
// var servercert []byte

// //go:embed secrets/serverkey.pem
// var serverkey []byte

// var transportCredentials credentials.TransportCredentials
// var clientCredentials credentials.TransportCredentials

//go:embed dynamic/*
var dynamic embed.FS

//go:embed static/*
var static embed.FS

//go:embed transactions.json
var transactionJSON string

//go:embed static/assets/IanResume_go.pdf
var resume []byte

var templates *template.Template // templates for dynamic pages
var routeMatch *regexp.Regexp    // template route regex
var count uint64                 // page hit counter
var startTime *time.Time         // start time of server running
var csrfCache *cache.Cache

const ( // various content types
	jsonContentType = "application/json; charset=utf-8"
	textContentType = "text/plain; charset=utf-8"
	htmlContentType = "text/html; charset=utf-8"
)

// PageData data for a page's templates
// Capitalized because it is used in templates and needs to be public
type PageData struct {
	Timestamp     time.Time
	LoadStart     time.Time
	LoadTime      time.Duration
	PageLoads     uint64
	Uptime        time.Duration
	CsrfToken     string
	IPAddress     string
	ServerAddress string
}

var router *mux.Router

// func TransportCredentials() credentials.TransportCredentials {
// 	return transportCredentials
// }

// GetRouter get reference to HTTP router
func GetRouter(inCloud bool) *mux.Router {
	router = mux.NewRouter().StrictSlash(true)

	// Sample JSON returning function
	router.HandleFunc("/transactions", GetTransactionsHandler).Methods(http.MethodGet).Name("Sample transactions")

	// We need to convert the embed FS to an io.FS in order to work with it
	fsys := fs.FS(static)

	// Handle static content
	// Note that we use http.FS to access our io.FS instead of trying to treat
	// it like a local directory. If you run the build in place it will work but
	// if you move the binary the files will not be available as http.Dir looks
	// for a locally available fileystem, not an embed one.

	// Normally with a system filesystem we'd use
	// ... http.FileServer(http.Dir("static")))).Name("Documentation")

	// Set file serving for css files
	contentCSS, _ := fs.Sub(fsys, "static/css")
	router.PathPrefix("/css").Handler(http.StripPrefix("/css", http.FileServer(http.FS(contentCSS)))).Name("CSS Files")

	// Set file serving for JS files
	contentJS, _ := fs.Sub(fsys, "static/js")
	router.PathPrefix("/js").Handler(http.StripPrefix("/js", http.FileServer(http.FS(contentJS)))).Name("JS Files")

	// For page tweets
	router.PathPrefix("/gettweet").HandlerFunc(TwitterHandler).Methods(http.MethodGet).Name("Get tweets")

	// NATS demo
	router.PathPrefix("/msg").HandlerFunc(NatsHandler).Methods(http.MethodGet).Name("Get NATS request")

	if inCloud {
		// For GRPC test using XKCD fetches
		router.PathPrefix("/getimage").HandlerFunc(XkcdNoGRPCHandler).Methods(http.MethodGet).Name("Get visa Non GRPC")
		// router.PathPrefix("/getimage").HandlerFunc(xkcdHandler).Methods(http.MethodGet).Name("Get
		// via GRPC")

	} else {
		// GRPC server not currently working
		// router.PathPrefix("/getimage").HandlerFunc(grpcpass.XkcdHandler).Methods(http.MethodGet).Name("Get via GRPC")
		router.PathPrefix("/getimage").HandlerFunc(XkcdNoGRPCHandler).Methods(http.MethodGet).Name("Get visa Non GRPC")
	}
	router.PathPrefix("/").HandlerFunc(TemplatePageHandler).Methods(http.MethodGet).Name("Dynamic pages")

	return router
}

func init() {
	// // Set up certificate that client and server can use
	// cert, err := tls.X509KeyPair(servercert, serverkey)
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// // Make the CertPool.
	// pool := x509.NewCertPool()
	// pool.AppendCertsFromPEM(servercert)

	// clientCredentials = credentials.NewClientTLSFromCert(pool, "grpc.com")

	// // Create the TLS credentials for GRPC server
	// transportCredentials = credentials.NewTLS(&tls.Config{
	// 	ClientAuth: tls.NoClientCert,
	// 	// Don't ask for a client certificate for now
	// 	// tls.RequireAndVerifyClientCert,
	// 	Certificates:       []tls.Certificate{cert},
	// 	ClientCAs:          pool,
	// 	InsecureSkipVerify: false,
	// })
}

// // ClientCredentials credentials for connecting to GRPC
// func ClientCredentials() *credentials.TransportCredentials {
// 	return &clientCredentials
// }

// uniqueToken get a random string that can be used as a CSRF header and later to
// fetch the server-stored JSR token string.
func uniqueToken() (token string, err error) {
	id, err := uuid.NewRandom()
	if err != nil {
		return "", err
	}
	token = id.String()

	return token, nil
}

// newPageData create a pointer to a new PageData struct instance
func newPageData() *PageData {
	pd := PageData{}
	pd.LoadStart = time.Now()

	return &pd
}

func getServerAddress(r *http.Request) string {
	ctx := r.Context()

	srvAddr := ctx.Value(http.LocalAddrContextKey).(net.Addr)

	return srvAddr.String()
}

func (pd *PageData) setServerAddress(address string) {
	pd.ServerAddress = address
}

func (pd *PageData) setToken(token string) {
	if token != "" {
		pd.CsrfToken = token
	}
}

// finalize finish off page info that is time specific
func (pd *PageData) finalize() {
	if pd.CsrfToken == "" {
		token, err := uniqueToken()
		if err != nil {
			token = ""
		} else {
			pd.setToken(token)
			err := csrfCache.Add(token, "", cache.DefaultExpiration)
			if err != nil {
				token = ""
			}
		}
		pd.CsrfToken = token
	}

	pd.LoadTime = time.Since(pd.LoadStart)
	pd.PageLoads = counterIncrement()
	pd.Uptime = time.Since(*startTime).Round(time.Second)
}

func findTokenFromRequest(r *http.Request) string {
	// This is not meant to be definitive but rather to avoid doing work for
	// free. The csrf token will be renewed frequently and will expire quickly.
	token := r.URL.Query().Get("csrf")
	if token == "" {
		return ""
	}
	_, ok := csrfCache.Get(token)
	if ok == false {
		return ""
	}
	// Renew cache
	csrfCache.Set(token, "", cache.DefaultExpiration)

	return token
}

// counterIncrement a simple increment of page hit count
func counterIncrement() uint64 {
	return atomic.AddUint64(&count, 1)
}

// init initialize counter and parse templates.
func init() {
	csrfCache = cache.New(5*time.Minute, 2*time.Minute)

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

// ResumeHandler serve up resume
func ResumeHandler(w http.ResponseWriter, r *http.Request) {
	//Set header
	w.Header().Set("Content-type", "application/pdf")

	//Stream to response
	if _, err := w.Write(resume); err != nil {
		w.WriteHeader(500)
	}
}

// NatsHandler NATS request handler
func NatsHandler(w http.ResponseWriter, r *http.Request) {
	// Connect to a server
	nc, _ := nats.Connect(nats.DefaultURL)

	// Make new search query parameters
	request := msg.NewQuery("plos", 0)

	// Create the request payload
	payload, err := json.MarshalIndent(request, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Requests
	msg, err := nc.Request("plos", []byte(payload), 10*time.Millisecond)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(msg.Data)
}

// XkcdNoGRPCHandler handler for XKCD with no GRPC
func XkcdNoGRPCHandler(w http.ResponseWriter, r *http.Request) {
	bytes, err := grpcpass.FetchRandomXKCD()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	xkcd, err := grpcpass.ParseXKCDJSON(bytes)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	json, err := json.MarshalIndent(&xkcd, "", "  ")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", jsonContentType)
	w.Write(json)

}

// TwitterHandler get an id for a tweet
func TwitterHandler(w http.ResponseWriter, r *http.Request) {
	findTokenFromRequest(r)

	td, err := tweets.GetTweetData()
	if err != nil { // simulate error getting data
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// fmt.Println("tweet data", td)
	payload, err := json.MarshalIndent(td, "", "  ")
	if err != nil { // simulate error getting data
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("Content-Type", jsonContentType)
	w.Write(payload)
}

// TemplatePageHandler use template collection to produce output
func TemplatePageHandler(w http.ResponseWriter, r *http.Request) {
	pd := newPageData()

	token := findTokenFromRequest(r)
	address := getServerAddress(r)
	pd.setServerAddress(address)

	matches := routeMatch.FindStringSubmatch(r.URL.Path)
	if len(matches) >= 1 {
		page := matches[1] + ".html"
		if templates.Lookup(page) != nil {
			w.Header().Add("Content-Type", htmlContentType)
			w.WriteHeader(http.StatusOK)
			pd.setToken(token)
			pd.finalize()
			templates.ExecuteTemplate(w, page, pd)
			return
		}
	} else if r.URL.Path == "/" {
		w.Header().Add("Content-Type", htmlContentType)
		w.WriteHeader(http.StatusOK)
		pd.setToken(token)
		pd.finalize()
		templates.ExecuteTemplate(w, "index.html", pd)
		return
	}

	w.WriteHeader(http.StatusNotFound)
	w.Header().Add("Content-Type", textContentType)
	w.Write([]byte("NOT FOUND"))
}

// GetTransactionsHandler get list of transactions
func GetTransactionsHandler(w http.ResponseWriter, r *http.Request) {

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
