package main

import (
	"crypto/tls"
	"crypto/x509"
	"embed"
	_ "embed"
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	"github.com/imarsman/nanovms/app/grpcpass"
	"google.golang.org/grpc"

	// "google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials"
	// "github.com/imarsman/nanovms/app/grpcpass"
)

//go:embed dynamic/*
var dynamic embed.FS

//go:embed static/*
var static embed.FS

//go:embed serverkey.pem
var serverkey []byte

//go:embed servercert.pem
var servercert []byte

//go:embed transactions.json
var transactionJSON string

//go:embed .context
var runContext string

var transportCredentials credentials.TransportCredentials
var clientCredentials credentials.TransportCredentials

// ClientCredentials credentials for connecting to GRPC
func ClientCredentials() *credentials.TransportCredentials {
	return &clientCredentials
}

func init() {
	// Set up certificate that client and server can use
	cert, err := tls.X509KeyPair(servercert, serverkey)
	if err != nil {
		log.Fatal(err)
	}

	// Make the CertPool.
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(servercert)

	clientCredentials = credentials.NewClientTLSFromCert(pool, "grpc.com")

	// Create the TLS credentials for GRPC server
	transportCredentials = credentials.NewTLS(&tls.Config{
		ClientAuth: tls.NoClientCert,
		// Don't ask for a client certificate for now
		// tls.RequireAndVerifyClientCert,
		Certificates:       []tls.Certificate{cert},
		ClientCAs:          pool,
		InsecureSkipVerify: false,
	})
}

// Main method for app. A simple router and static, struct/json producing
// template, Golang template pages, and a Twitter API handler.
func main() {
	infiniteWait := make(chan string)

	cloud := strings.TrimSpace(runContext) == "cloud"

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

	// router.PathPrefix("/getimage").HandlerFunc(xkcdNoGRPCHandler).Methods(http.MethodGet).Name("Get via GRPC")

	lis, err := net.Listen("tcp", ":5222")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// https://grpc.io/docs/languages/go/basics/
	// https://github.com/grpc/grpc-go/tree/master/examples
	// var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(grpc.Creds(transportCredentials))

	// Register service. Code for XKCDService implemented by code generation.
	grpcpass.RegisterXKCDServiceServer(grpcServer, &grpcpass.XKCDService{})
	fmt.Printf("grpc server: %+v\n", grpcServer.GetServiceInfo())

	// For now just use an unprivileged port. Running locally as non-root would
	// fail but running in the cloud should be fine, but that would take more
	// effort than is currently warrrented. May revisit.
	if cloud {
		go func() {
			fmt.Println("Running in cloud mode with nanovms unikernel. Serving transactions on port", "8000")
			// For GRPC test using XKCD fetches
			router.PathPrefix("/getimage").HandlerFunc(xkcdNoGRPCHandler).Methods(http.MethodGet).Name("Get visa Non GRPC")
			// router.PathPrefix("/getimage").HandlerFunc(xkcdHandler).Methods(http.MethodGet).Name("Get via GRPC")
			// Default
			router.PathPrefix("/").HandlerFunc(templatePageHandler).Methods(http.MethodGet).Name("Dynamic pages")
			http.ListenAndServe(":8000", router)
		}()
		go func() {
			fmt.Printf("Starting GRPC server on port %v\n", lis.Addr().String())
			if err := grpcServer.Serve(lis); err != nil {
				fmt.Printf("failed to serve: %s", err)
				// log.Fatalf("failed to serve: %s", err)
			}
		}()
	} else {
		go func() {
			fmt.Println("Running locally in OS. Serving transactions on port", "8000")
			// For GRPC test using XKCD fetches
			router.PathPrefix("/getimage").HandlerFunc(xkcdNoGRPCHandler).Methods(http.MethodGet).Name("Get via GRPC")
			// router.PathPrefix("/getimage").HandlerFunc(xkcdHandler).Methods(http.MethodGet).Name("Get via GRPC")
			// Default
			router.PathPrefix("/").HandlerFunc(templatePageHandler).Methods(http.MethodGet).Name("Dynamic pages")

			// router.PathPrefix("/getimage").HandlerFunc(xkcdHandler).Methods(http.MethodGet).Name("Get via GRPC")
			http.ListenAndServe(":8000", router)
		}()
		go func() {
			fmt.Printf("Starting GRPC server on port %v\n", lis.Addr().String())
			if err := grpcServer.Serve(lis); err != nil {
				log.Fatalf("failed to serve: %s", err)
			}
			fmt.Println("started grpc")
		}()
	}

	<-infiniteWait
}
