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

// //go:embed serverkey.pem servercert.pem
// var certs embed.FS

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

// func generateCert(ca bool, parent *x509.Certificate) (*pem.Block, *pem.Block) {
// 	// Generate a key.
// 	key, err := ecdsa.GenerateKey(elliptic.P224(), rand.Reader)
// 	if err != nil {
// 		log.Fatalf("failed to generate private key: %s", err)
// 	}
// 	// Fill out the template.
// 	template := x509.Certificate{
// 		SerialNumber:          new(big.Int).SetInt64(0),
// 		Subject:               pkix.Name{Organization: []string{host}},
// 		NotBefore:             time.Now(),
// 		NotAfter:              time.Date(2049, 12, 31, 23, 59, 59, 0, time.UTC),
// 		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
// 		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
// 		BasicConstraintsValid: true,
// 		IPAddresses:           []net.IP{net.ParseIP(host)},
// 	}
// 	if ca {
// 		template.IsCA = true
// 		template.KeyUsage |= x509.KeyUsageCertSign
// 	}
// 	if parent == nil {
// 		parent = &template
// 	}
// 	// Generate the certificate.
// 	cert, err := x509.CreateCertificate(rand.Reader, &template, parent, &key.PublicKey, key)
// 	if err != nil {
// 		log.Fatalf("Failed to create certificate: %s", err)
// 	}
// 	// Marshal the key.
// 	b, err := x509.MarshalECPrivateKey(key)
// 	if err != nil {
// 		log.Fatalf("Failed to marshal ecdsa: %s", err)
// 	}
// 	return &pem.Block{Type: "CERTIFICATE", Bytes: cert},
// 		&pem.Block{Type: "ECDSA PRIVATE KEY", Bytes: b}
// }

// ClientCredentials credentials for connecting to GRPC
func ClientCredentials() *credentials.TransportCredentials {
	return &clientCredentials
}

func init() {
	// Set up certificate that client and server can use

	//https://play.golang.org/p/Tk9CR4BUyU
	// https://blog.gopheracademy.com/advent-2019/go-grps-and-tls/
	// https://play.golang.org/p/NyImQd5Xym

	// fmt.Println(string(servercert))
	// fmt.Println(string(serverkey))

	// rootCertPem, _ := generateCert(true, nil)
	// rootCert, err := x509.ParseCertificate(rootCertPem.Bytes)
	// if err != nil {
	// 	log.Fatalf("failt to make parent: %s", err)
	// }

	// Read in the cert file
	// certs, err := ioutil.ReadFile("")
	// if err != nil {
	// 	log.Fatalf("Failed to append %q to RootCAs: %v", "", err)
	// }

	// // cert, err := tls.X509KeyPair(servercert, serverkey)
	cert, err := tls.X509KeyPair(servercert, serverkey)
	if err != nil {
		log.Fatal(err)
	}

	// Create client certificate.
	// clientCertPem, clientKeyPem := generateCert(false, rootCert)
	// c, err := tls.X509KeyPair(pem.EncodeToMemory(clientCertPem),
	// 	pem.EncodeToMemory(clientKeyPem))
	// if err != nil {
	// 	log.Fatalf("making client TLS cert: %s", err)
	// }

	// Make the CertPool.
	pool := x509.NewCertPool()
	pool.AppendCertsFromPEM(servercert)

	// pool.AddCert(cert.Leaf)

	// fmt.Println(cert)

	// certPool := x509.NewCertPool()
	// certPool.AddCert(cert.Leaf)

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

	//  ..AppendCertsFromPEM(ca); !ok {
	// 	return errors.New("failed to append client certs")
	// }
	// contentCSS, _ := fs.Sub(certfs, "static/css")
	// // contentCSS

	// pool, _ := x509.SystemCertPool()

	// var tc *credentials.TransportCredentials
	// cred := credentials.NewClientTLSFromFile()

	// serverCreds = credentials.new
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

	lis, err := net.Listen("tcp", ":9080")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	// https://grpc.io/docs/languages/go/basics/
	// https://github.com/grpc/grpc-go/tree/master/examples
	// var opts []grpc.ServerOption
	grpcServer := grpc.NewServer(grpc.Creds(transportCredentials))

	// grpcServer.RegisterService(grpcpass.XKCDService_ServiceDesc, grpcServer)
	grpcpass.RegisterXKCDServiceServer(grpcServer, &grpcpass.XKCDService{})
	fmt.Printf("grpc server service info: %+v\n", grpcServer.GetServiceInfo())
	// .RegisterService(grpcpass.XKCDService_ServiceDesc, grpcServer)
	// grpcServer.RegisterService(grpcpass.GetXKCD, grpcpass.XKCDService)

	// For now just use an unprivileged port. Running locally as non-root would
	// fail but running in the cloud should be fine, but that would take more
	// effort than is currently warrrented. May revisit.
	if cloud {
		go func() {
			fmt.Println("Running in cloud mode with nanovms unikernel. Serving transactions on port", "8000")
			// For GRPC test using XKCD fetches
			router.PathPrefix("/getimage").HandlerFunc(xkcdNoGRPCHandler).Methods(http.MethodGet).Name("Get via GRPC")
			// Default
			router.PathPrefix("/").HandlerFunc(templatePageHandler).Methods(http.MethodGet).Name("Dynamic pages")
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
			// For GRPC test using XKCD fetches
			// router.PathPrefix("/getimage").HandlerFunc(xkcdNoGRPCHandler).Methods(http.MethodGet).Name("Get via GRPC")
			router.PathPrefix("/getimage").HandlerFunc(xkcdHandler).Methods(http.MethodGet).Name("Get via GRPC")
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
