package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strconv"
	"sync"
	"sync/atomic"
	"text/template"
	"time"
)

var t *template.Template
var routeMatch *regexp.Regexp
var count uint64
var startTime *time.Time
var mu sync.Mutex

const (
	jsonContentType     = "application/json; charset=utf-8"
	markdownContentType = "text/markdown; charset=utf-8"
	textContentType     = "text/plain; charset=utf-8"
	htmlContentType     = "text/html; charset=utf-8"
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

// PageData data for a page's templates
type PageData struct {
	Timestamp time.Time
	LoadStart time.Time
	LoadTime  time.Duration
	PageLoads uint64
	Uptime    time.Duration
}

func newPageData() *PageData {
	pd := PageData{}
	pd.LoadStart = time.Now()

	return &pd
}

func (pd *PageData) finalize() {
	pd.LoadTime = time.Since(pd.LoadStart)
	pd.PageLoads = counterIncrement()
	pd.Uptime = time.Since(*startTime)
}

func counterIncrement() uint64 {
	return atomic.AddUint64(&count, 1)
}

// init initialize counter and parse templates.
func init() {
	tm := time.Now()
	startTime = &tm

	// We need to convert the embed FS to an io.FS in order to work with it
	fsys := fs.FS(dynamic)
	contentDynamic, _ := fs.Sub(fsys, "dynamic")

	var err error
	t, err = template.ParseFS(contentDynamic, "templates/*.html")
	if err != nil {
		log.Println("Cannot parse templates:", err)
		os.Exit(-1)
	}
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
		if t.Lookup(page) != nil {
			w.Header().Add("Content-Type", htmlContentType)
			w.WriteHeader(http.StatusOK)
			pd.finalize()
			t.ExecuteTemplate(w, page, pd)
			return
		}
	} else if r.URL.Path == "/" {
		w.Header().Add("Content-Type", htmlContentType)
		w.WriteHeader(http.StatusOK)
		pd.finalize()
		t.ExecuteTemplate(w, "index.html", pd)
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
