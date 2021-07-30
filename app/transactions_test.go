package main

import (
	"encoding/json"
	"io/fs"
	"testing"

	"github.com/matryer/is"
)

// TestReadJSON test of JSON sample file reading with some output
func TestReadJSON(t *testing.T) {
	is := is.New(t)

	is.True(1 == 1)

	transactions, err := readTransactions()
	is.NoErr(err)

	t.Log(len(transactions.Transactions))

	for i := 0; i < len(transactions.Transactions); i++ {
		transaction := transactions.Transactions[i]
		t.Log(transaction.TransactionID)
	}
}

// TestSort test sort of transactions by post timestamp
func TestSort(t *testing.T) {
	is := is.New(t)
	transactions, err := readTransactions()
	is.NoErr(err)

	sortDescendingPostTimestamp(&transactions)

	for i := 0; i < len(transactions.Transactions); i++ {
		transaction := transactions.Transactions[i]
		t.Log(transaction.ID, transaction.TransactionID, transaction.PostedTimeStamp)
	}

}

func TestToJSON(t *testing.T) {
	is := is.New(t)
	transactions, err := readTransactions()
	is.NoErr(err)

	sortDescendingPostTimestamp(&transactions)

	t.Log(toJSON(transactions))
}

func TestObscurePAN(t *testing.T) {
	is := is.New(t)

	transactions, err := readTransactions()
	is.NoErr(err)

	transactions, err = obscureTransactionID(transactions)
	is.NoErr(err)

	json, err := json.MarshalIndent(&transactions, "", "  ")
	is.NoErr(err)
	t.Logf("%v", string(json))

	t.Log("Transactions in set")
	for i, v := range transactions.Transactions {
		t.Logf("%d %+v", i, v)
	}
}

// TestEmbeddedFS test for files in embedded FS
func TestEmbeddedFS(t *testing.T) {
	is := is.New(t)

	// In order to use embedded FS we need to convert it to an io.FS
	// the io.FS can then be used to access files in the embedded FS.
	fsys := fs.FS(static)
	contentStatic, err := fs.Sub(fsys, "static")
	is.NoErr(err)

	fs.WalkDir(contentStatic, ".",
		func(path string, d fs.DirEntry, err error) error {
			is.NoErr(err)
			if err != nil {
				return err
			}
			if d.IsDir() == false {
				t.Log("file at path", path, d.Name())
			}
			return nil
		})
}
