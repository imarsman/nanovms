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
	"github.com/imarsman/nanovms/app/grpcpass"
	"github.com/imarsman/nanovms/app/handlers"
	"google.golang.org/grpc"

	"github.com/nats-io/nats-server/v2/server"
	stand "github.com/nats-io/nats-streaming-server/server"
	"github.com/nats-io/nats.go"
)

//go:embed static/*
var static embed.FS

//go:embed .context
var runContext string

func init() {
}

// Main method for app. A simple router and static, struct/json producing
// template, Golang template pages, and a Twitter API handler.
func main() {
	infiniteWait := make(chan string)

	cloud := strings.TrimSpace(runContext) == "cloud"

	router := mux.NewRouter().StrictSlash(true)

	// Sample JSON returning function
	router.HandleFunc("/transactions", handlers.GetTransactionsHandler).Methods(http.MethodGet).Name("Sample transactions")

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
	router.PathPrefix("/gettweet").HandlerFunc(handlers.TwitterHandler).Methods(http.MethodGet).Name("Get tweets")

	// NATS demo
	router.PathPrefix("/msg").HandlerFunc(handlers.NatsHandler).Methods(http.MethodGet).Name("Get NATS request")

	// router.PathPrefix("/getimage").HandlerFunc(xkcdNoGRPCHandler).Methods(http.MethodGet).Name("Get via GRPC")

	lis, err := net.Listen("tcp", ":5222")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// https://grpc.io/docs/languages/go/basics/
	// https://github.com/grpc/grpc-go/tree/master/examples
	// var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(grpc.Creds(handlers.TransportCredentials()))

	// Register service. Code for XKCDService implemented by code generation.
	grpcpass.RegisterXKCDServiceServer(grpcServer, &grpcpass.XKCDService{})
	fmt.Printf("grpc server: %+v\n", grpcServer.GetServiceInfo())

	// https://sourcegraph.com/github.com/nats-io/nats-server@6da5d2f4907a03c8ba26fc8b6ca2aed903ac80f8/-/blob/main.go
	// Now we want to setup the monitoring port for NATS Streaming.
	// We still need NATS Options to do so, so create NATS Options
	// using the NewNATSOptions() from the streaming server package.
	snopts := stand.NewNATSOptions()
	snopts.Port = nats.DefaultPort
	snopts.HTTPPort = 8223

	// Now run the server with the streaming and streaming/nats options.
	natsServer, err := server.NewServer(snopts)
	if err != nil {
		panic(err)
	}

	// For now just use an unprivileged port. Running locally as non-root would
	// fail but running in the cloud should be fine, but that would take more
	// effort than is currently warrrented. May revisit.
	if cloud {
		go func() {
			fmt.Println("Running in cloud mode with nanovms unikernel. Serving transactions on port", "8000")
			// For GRPC test using XKCD fetches
			router.PathPrefix("/getimage").HandlerFunc(handlers.XkcdNoGRPCHandler).Methods(http.MethodGet).Name("Get visa Non GRPC")
			// router.PathPrefix("/getimage").HandlerFunc(xkcdHandler).Methods(http.MethodGet).Name("Get
			// via GRPC")

			router.PathPrefix("/").HandlerFunc(handlers.TemplatePageHandler).Methods(http.MethodGet).Name("Dynamic pages")

			http.ListenAndServe(":8000", router)

		}()
		go func() {
			fmt.Printf("Starting GRPC server on port %v\n", lis.Addr().String())
			if err := grpcServer.Serve(lis); err != nil {
				fmt.Printf("failed to serve: %s", err)
				// log.Fatalf("failed to serve: %s", err)
			}
		}()
		go func() {
			fmt.Println("Starting NATS server")
			// Start things up. Block here until done.
			if err := server.Run(natsServer); err != nil {
				server.PrintAndDie(err.Error())
			}
			natsServer.WaitForShutdown()
		}()
	} else {
		go func() {
			fmt.Println("Running locally in OS. Serving transactions on port", "8000")
			// For GRPC test using XKCD fetches
			// router.PathPrefix("/getimage").HandlerFunc(xkcdNoGRPCHandler).Methods(http.MethodGet).Name("Get via GRPC")
			router.PathPrefix("/getimage").HandlerFunc(handlers.XkcdHandler).Methods(http.MethodGet).Name("Get via GRPC")
			// Default
			router.PathPrefix("/").HandlerFunc(handlers.TemplatePageHandler).Methods(http.MethodGet).Name("Dynamic pages")

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
		go func() {
			fmt.Println("Starting NATS server")

			// Start things up. Block here until done.
			if err := server.Run(natsServer); err != nil {
				server.PrintAndDie(err.Error())
			}
			natsServer.WaitForShutdown()
		}()
	}

	<-infiniteWait
}
