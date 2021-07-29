package main

import (
	"net/http"

	"github.com/gorilla/mux"
)

const (
	jsonContentType     = "application/json; charset=utf-8"
	markdownContentType = "text/markdown; charset=utf-8"
	textContentType     = "text/plain; charset=utf-8"
)

func init() {
	router = mux.NewRouter().StrictSlash(true)

	// Handle static content
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./static")))).Name("Documentation")

	// Sample JSON returning function
	router.HandleFunc("/transactions", GetTransactionsHandler).Methods("GET").Name("Sample transactions")
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
