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
	inCloud = strings.TrimSpace(runContext) == "cloud"
	handlers.SetInCloud(inCloud)
}

var inCloud bool

// InCloud check if running in cloud
func InCloud() bool {
	return inCloud
}

// Main method for app. A simple router and static, struct/json producing
// template, Golang template pages, and a Twitter API handler.
func main() {
	infiniteWait := make(chan string)

	// HTTP
	if inCloud {
		fmt.Println("Running in cloud mode with nanovms unikernel. Serving html on port", "8000")
	} else {
		fmt.Println("Running locally in OS. Serving html on port", "8000")
	}
	go func() {
		// Get the router with a flag for whether or not app is runnin in cloud
		httpRouter := handlers.GetRouter(inCloud)

		fmt.Printf("Starting HTTP server on port %v\n", "8000")
		if err := http.ListenAndServe(":8000", httpRouter); err != nil {
			fmt.Printf("failed to serve: %s", err)
		}
	}()

	// GRPC
	go func() {
		// Problems running in cloud for now
		if inCloud == false {
			lis, err := net.Listen("tcp", ":5222")
			if err != nil {
				log.Fatalf("failed to listen: %v", err)
			}
			grpcServer := grpcpass.GRPCServer()
			fmt.Printf("Starting GRPC server on port %v\n", lis.Addr().String())
			if err := grpcServer.Serve(lis); err != nil {
				fmt.Printf("failed to serve: %s", err)
				// log.Fatalf("failed to serve: %s", err)
			}

			fmt.Println("grpc", grpcServer.GetServiceInfo())
		}
	}()

	// NAT
	go func() {
		// Problems running in cloud for now
		if inCloud == false {
			// Skipping for now
			ns := msg.NATServer()

			fmt.Printf("Starting NAT server on %v\n", nats.DefaultPort)
			// Start things up. Block here until done.
			if err := server.Run(ns); err != nil {
				server.PrintAndDie(err.Error())
			}

			ns.WaitForShutdown()
		}
	}()

	<-infiniteWait
}
