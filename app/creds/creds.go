package creds

import (
	"crypto/tls"
	"crypto/x509"
	"log"

	// We need embed here
	_ "embed"

	"google.golang.org/grpc/credentials"
)

//go:embed secrets/servercert.pem
var servercert []byte

//go:embed secrets/serverkey.pem
var serverkey []byte

var transportCredentials *credentials.TransportCredentials
var clientCredentials *credentials.TransportCredentials

// TransportCredentials credentials for HTTP transport
func TransportCredentials() *credentials.TransportCredentials {
	return transportCredentials
}

// ClientCredentials credentials for connecting to GRPC
func ClientCredentials() *credentials.TransportCredentials {
	return clientCredentials
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

	// Create the TLS credentials for GRPC server
	tc := credentials.NewTLS(&tls.Config{
		ClientAuth: tls.NoClientCert,
		// Don't ask for a client certificate for now
		// tls.RequireAndVerifyClientCert,
		Certificates:       []tls.Certificate{cert},
		ClientCAs:          pool,
		InsecureSkipVerify: true,
	})
	transportCredentials = &tc
	// clientCredentials = *(&credentials.NewClientTLSFromCert(pool, "grpc.com"))
	cc := credentials.NewClientTLSFromCert(pool, "grpc.com")
	clientCredentials = &cc
}
