package handlers

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
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

	// Indent for clarity here but would consider not for machine->machine communication
	bytes, err := json.MarshalIndent(&transactions, "", "  ")
	if err != nil {
		fmt.Println("error")
		return "", err
	}

	return string(bytes), nil
}

// obscureTransactionID obsure PAN attribute
func obscureTransactionID(transactionlist TransactionList) (TransactionList, error) {
	newTrans := TransactionList{}
	for i := 0; i < len(transactionlist.Transactions); i++ {
		transaction := transactionlist.Transactions[i]
		s := fmt.Sprint(transaction.TransactionID)
		var lastDigits int = 0
		// If length is >= 4 use last found digits, otherwise stick with what we
		// have
		if len(s) >= 4 {
			s = s[len(s)-4:]
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
