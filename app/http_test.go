package main

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/matryer/is"
)

// newRootServer a simple server for documents at /
func newTemplateServer() *mux.Router {
	r := mux.NewRouter()

	// fsys := fs.FS(static)
	// contentStatic, _ := fs.Sub(fsys, "static")
	// r.PathPrefix("/css").Handler(http.StripPrefix("/css", http.FileServer(http.FS(contentStatic)))).Name("Documentation")
	r.HandleFunc("/", parsePageHandler).Methods("GET")

	return r
}

// newCSSServer a simple server for documents at /
func newCSSServer() *mux.Router {
	r := mux.NewRouter()

	fsys := fs.FS(static)
	contentCSS, _ := fs.Sub(fsys, "static/css")

	// Handle static content
	// Note that we use http.FS to access our io.FS instead of trying to treat
	// it like a local directory. If you run the build in place it will work but
	// if you move the binary the files will not be available as http.Dir looks
	// for a locally available fileystem, not an embed one.

	// Normally with a system filesystem we'd use
	// ... http.FileServer(http.Dir("static")))).Name("Documentation")
	r.PathPrefix("/css").Handler(http.StripPrefix("/css", http.FileServer(http.FS(contentCSS)))).Name("CSS Files")

	return r
}

// newTransactionsRequestServer a simple server to handle the /transactions endpoint
func newTransactionsRequestServer() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/transactions/", getTransactionsHandler).Methods("GET")

	return r
}

// callServer call a server given a router with routes set up
func callServer(t *testing.T, router *mux.Router, method, path string, expected int) error {
	r := strings.NewReader("")

	// Define the request then the recorder for the request
	req, _ := http.NewRequest(method, path, r)
	res := httptest.NewRecorder()

	// Get an instance of server with set endpoint
	router.ServeHTTP(res, req)

	// Check the status code is what we expect.
	if status := res.Code; status != expected {
		return fmt.Errorf("Got unexpected response code %d", res.Code)
	}

	// Check the response body is what we expect.
	if res.Body.String() == "" {
		return fmt.Errorf("Got unexpected returned body %v", res.Body.String())
	}

	// Print what was recieved
	t.Logf("Path %s response\n %s", path, res.Body.String())

	return nil
}

// Test of call handlers.
func TestCallHandlers(t *testing.T) {
	is := is.New(t)

	t.Log("Calling GetTransactions at /transactions")
	err := callServer(t, newTransactionsRequestServer(), http.MethodGet, "/transactions/", http.StatusOK)
	is.NoErr(err)

	t.Log("Calling css docs at /css")
	err = callServer(t, newCSSServer(), http.MethodGet, "/css/simple.min.css", http.StatusOK)
	is.NoErr(err)

	t.Log("Calling template docs at /")
	err = callServer(t, newTemplateServer(), http.MethodGet, "/", http.StatusOK)
	is.NoErr(err)
}
