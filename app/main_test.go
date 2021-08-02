package main

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	"github.com/imarsman/nanovms/app/grpcpass"
	"google.golang.org/grpc"
)

// var ch chan (int)

func NewGRPCServer(t *testing.T) (*grpc.Server, func()) {
	var grpcServer *grpc.Server

	lis, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	grpcServer = grpc.NewServer()
	grpcpass.RegisterXKCDServiceServer(grpcServer, &grpcpass.XKCDService{})
	t.Log("Serving GRPC on port", lis.Addr())

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	time.Sleep(5 * time.Second)

	return grpcServer, func() {
		grpcServer.GracefulStop()
	}
}

func init() {

}

func TestCallGRPC(t *testing.T) {
	_, cleanup := NewGRPCServer(t)
	defer cleanup()

	var opts []grpc.DialOption
	opts = append(opts, grpc.WithInsecure())
	opts = append(opts, grpc.WithBlock())

	serverAddr := "localhost:9000"
	conn, err := grpc.Dial(serverAddr, opts...)
	if err != nil {
		log.Fatalf("fail to dial: %v", err)
	}

	client := grpcpass.NewXKCDServiceClient(conn)
	// time.Sleep(15 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	number := grpcpass.MessageNumber{}
	number.Number = 1001

	t.Log("client", client)
	// callOpts := grpc.Cal
	callOption := grpc.MaxCallRecvMsgSize(0)
	message, err := client.GetXKCD(ctx, &number, callOption)
	if err != nil {
		log.Fatalf("%v.GetXKCD(_) = _, %v: ", client, err)
	}

	t.Logf("%+v", message)
}
