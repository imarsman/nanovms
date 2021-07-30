package main

import (
	"embed"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/gorilla/mux"
)

//go:embed static/*
var static embed.FS

//go:embed transactions.json
var transactionJSON string

//go:embed .context
var context string

const (
	jsonContentType     = "application/json; charset=utf-8"
	markdownContentType = "text/markdown; charset=utf-8"
	textContentType     = "text/plain; charset=utf-8"
)

// TransactionList a list of transactions. Allows for JSON list to be read
type TransactionList struct {
	Transactions []Transaction `json:"transactions"`
}

// Transaction a transaction with attributes
type Transaction struct {
	ID                  int    `json:"id"`
	Amount              int    `json:"amount"`
	MessageType         string `json:"conversation_type"`
	CreatedAt           string `json:"created_at"`
	TransactionID       int    `json:"transaction_id"`
	TransactionCategory string `json:"transaction_category"`
	PostedTimeStamp     string `json:"posted_timestamp"`
	TransactionType     string `json:"transaction_type"`
	SendingAccount      int    `json:"sending_account"`
	ReceivingAccount    int    `json:"receiving_account"`
	TransactionNote     string `json:"transaction_note"`
}

// obscureTransactionID obsure PAN attribute
func obscureTransactionID(transactionlist TransactionList) (TransactionList, error) {
	newTrans := TransactionList{}
	for i := 0; i < len(transactionlist.Transactions); i++ {
		transaction := transactionlist.Transactions[i]
		s := fmt.Sprint(transaction.TransactionID)
		var lastDigits int = 0
		if len(s) > 0 {
			if len(s) >= 4 {
				s = s[len(s)-4:]
			}
		}
		lastDigits, err := strconv.Atoi(s)
		if err != nil {
			return TransactionList{}, err
		}
		transaction.TransactionID = lastDigits
		newTrans.Transactions = append(newTrans.Transactions, transaction)
	}

	return newTrans, nil
}

// readTransactions read sample transactions set from JSON file
func readTransactions() (TransactionList, error) {

	if transactionJSON == "" {
		return TransactionList{}, errors.New("Could not load transactions")
	}
	var transactionList TransactionList

	json.Unmarshal([]byte(transactionJSON), &transactionList)

	return transactionList, nil
}

// sortDescendingPostTimestamp sort transaction slice descending by post
// timestamp. A production function would likely not be hard coded in this way
// unless there was a rule requiring this specific sort.
func sortDescendingPostTimestamp(transactions *TransactionList) *TransactionList {
	sort.SliceStable(transactions.Transactions, func(i, j int) bool {
		return transactions.Transactions[i].PostedTimeStamp > transactions.Transactions[j].PostedTimeStamp
	})

	return transactions
}

// toJSON get JSON for Transactions struct
func toJSON(transactions TransactionList) (string, error) {

	// t := obscurePan(transactions)
	// Indent for clarity here but would consider not for machine->machine communication
	bytes, err := json.MarshalIndent(&transactions, "", "  ")
	if err != nil {
		fmt.Println("error")
		return "", err
	}

	return string(bytes), nil
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

// Main method for app. A simple router and a simple handler.
func main() {
	infiniteWait := make(chan string)

	cloud := strings.TrimSpace(context) == "cloud"
	fmt.Println("context", context, "cloud", cloud)

	router := mux.NewRouter().StrictSlash(true)

	// Sample JSON returning function
	router.HandleFunc("/transactions", GetTransactionsHandler).Methods(http.MethodGet).Name("Sample transactions")

	// We need to convert the embed FS to an io.FS in order to work with it
	fsys := fs.FS(static)
	contentStatic, _ := fs.Sub(fsys, "static")

	// Handle static content
	// Note that we use http.FS to access our io.FS instead of trying to treat
	// it like a local directory. If you run the build in place it will work but
	// if you move the binary the files will not be available as http.Dir looks
	// for a locally available fileystem, not an embed one.

	// Normally with a system filesystem we'd use
	// ... http.FileServer(http.Dir("static")))).Name("Documentation")
	router.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.FS(contentStatic)))).Name("Documentation")

	// For now just use an unprivileged port. Running locally as non-root would
	// fail but running in the cloud should be fine, but that would take more
	// effort than is currently warrrented. May revisit.
	if cloud {
		go func() {
			fmt.Println("Running in cloud mode with nanovms unikernel. Serving transactions on port", "8000")
			http.ListenAndServe(":8000", router)
		}()
	} else {
		go func() {
			fmt.Println("Running locally in OS. Serving transactions on port", "8000")
			http.ListenAndServe(":8000", router)
		}()
	}

	<-infiniteWait
}
