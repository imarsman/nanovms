package app

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/matryer/is"
)

// newTransactionsRequestServer a simple server to handle the /transactions endpoint
func newTransactionsRequestServer() http.Handler {
	r := mux.NewRouter()
	r.HandleFunc("/transactions", GetTransactions).Methods("GET")
	return r
}

// callGetTransactions call the /transactions endpoint
func callGetTransactions(t *testing.T) error {
	r := strings.NewReader("")

	// Define the request then the recorder for the request
	req, _ := http.NewRequest("GET", "/transactions", r)
	res := httptest.NewRecorder()
	// Get an instance of server with set endpint for /transactions
	newTransactionsRequestServer().ServeHTTP(res, req)

	// Check the status code is what we expect.
	if status := res.Code; status != http.StatusOK {
		return fmt.Errorf("Got unexpected response code %d", res.Code)
	}

	// Check the response body is what we expect.
	if len(res.Body.String()) == 0 {
		return fmt.Errorf("Got unexpected returned body %v", res.Body.String())
	}

	// Print what was recieved
	t.Logf("Get Transactions response: %s", res.Body.String())

	return nil
}

// Test of call handlers. There is only one here but some sort of method would
// be useful if there were more to run them all.
func TestCallHandlers(t *testing.T) {
	is := is.New(t)

	is.True(1 == 1)

	t.Log("Calling GetTransactions at /transactions")
	err := callGetTransactions(t)
	is.NoErr(err)
}
