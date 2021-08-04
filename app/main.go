package main

import (
	"embed"
	_ "embed"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

	"github.com/imarsman/nanovms/app/grpcpass"
	"github.com/imarsman/nanovms/app/handlers"
	"github.com/imarsman/nanovms/app/msg"

	"github.com/nats-io/nats-server/v2/server"
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

	isCloud := strings.TrimSpace(runContext) == "cloud"

	router := handlers.GetRouter(isCloud)

	grpcServer := grpcpass.GRPCServer()

	natsServer := msg.NATServer()

	// For now just use an unprivileged port. Running locally as non-root would
	// fail but running in the cloud should be fine, but that would take more
	// effort than is currently warrrented. May revisit.
	if isCloud {
		fmt.Println("Running in cloud mode with nanovms unikernel. Serving transactions on port", "8000")
	} else {
		fmt.Println("Running locally in OS. Serving transactions on port", "8000")
	}
	go func() {
		fmt.Printf("Starting HTTP server on port %v\n", "8000")
		if err := http.ListenAndServe(":8000", router); err != nil {
			fmt.Printf("failed to serve: %s", err)
		}

		lis, err := net.Listen("tcp", ":5222")
		if err != nil {
			log.Fatalf("failed to listen: %v", err)
		}

		fmt.Printf("Starting GRPC server on port %v\n", lis.Addr().String())
		if err := grpcServer.Serve(lis); err != nil {
			fmt.Printf("failed to serve: %s", err)
			// log.Fatalf("failed to serve: %s", err)
		}

		fmt.Printf("Starting NATS server on %v\n", nats.DefaultPort)
		// Start things up. Block here until done.
		if err := server.Run(natsServer); err != nil {
			server.PrintAndDie(err.Error())
		}
		natsServer.WaitForShutdown()
	}()

	<-infiniteWait
}
