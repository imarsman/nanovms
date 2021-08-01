package main

import (
	"embed"
	_ "embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"google.golang.org/grpc"
)

//go:embed dynamic/*
var dynamic embed.FS

//go:embed static/*
var static embed.FS

//go:embed transactions.json
var transactionJSON string

//go:embed .context
var context string

// Main method for app. A simple router and static, struct/json producing
// template, Golang template pages, and a Twitter API handler.
func main() {
	infiniteWait := make(chan string)

	cloud := strings.TrimSpace(context) == "cloud"

	router := mux.NewRouter().StrictSlash(true)

	// Sample JSON returning function
	router.HandleFunc("/transactions", getTransactionsHandler).Methods(http.MethodGet).Name("Sample transactions")

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
	router.PathPrefix("/gettweet").HandlerFunc(twitterHandler).Methods(http.MethodGet).Name("Get tweets")

	// Default
	router.PathPrefix("/").HandlerFunc(templatePageHandler).Methods(http.MethodGet).Name("Dynamic pages")

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()

	// For now just use an unprivileged port. Running locally as non-root would
	// fail but running in the cloud should be fine, but that would take more
	// effort than is currently warrrented. May revisit.
	if cloud {
		go func() {
			fmt.Println("Running in cloud mode with nanovms unikernel. Serving transactions on port", "8000")
			http.ListenAndServe(":8000", router)
		}()
		go func() {
			fmt.Printf("Starting GRPC server on port %v\n", lis.Addr().String())
			if err := grpcServer.Serve(lis); err != nil {
				log.Fatalf("failed to serve: %s", err)
			}
		}()
	} else {
		go func() {
			fmt.Println("Running locally in OS. Serving transactions on port", "8000")
			http.ListenAndServe(":8000", router)
		}()
		go func() {
			fmt.Printf("Starting GRPC server on port %v\n", lis.Addr().String())
			if err := grpcServer.Serve(lis); err != nil {
				log.Fatalf("failed to serve: %s", err)
			}
		}()
	}

	<-infiniteWait
}
