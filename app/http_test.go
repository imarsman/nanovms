package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/matryer/is"
)

// newRootServer a simple server for documents at /
func newRootServer() *mux.Router {
	r := mux.NewRouter()
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./static")))).Name("Documentation")

	return r
}

// newTransactionsRequestServer a simple server to handle the /transactions endpoint
func newTransactionsRequestServer() *mux.Router {
	r := mux.NewRouter()
	r.HandleFunc("/transactions", GetTransactionsHandler).Methods("GET")

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
	err := callServer(t, newTransactionsRequestServer(), http.MethodGet, "/transactions", http.StatusOK)
	is.NoErr(err)

	t.Log("Calling static docs at /")
	err = callServer(t, newRootServer(), http.MethodGet, "/", http.StatusOK)
	is.NoErr(err)
}
