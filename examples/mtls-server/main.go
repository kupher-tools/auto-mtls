package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	// Load server cert and key
	serverCert, err := tls.LoadX509KeyPair("/etc/tls/tls.crt", "/etc/tls/tls.key")
	if err != nil {
		log.Fatalf("failed to load server key pair: %v", err)
	}

	// Load CA cert to verify clients
	caCert, err := os.ReadFile("/etc/ca/ca.crt")
	if err != nil {
		log.Fatalf("failed to read CA cert: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// TLS configuration with client cert verification
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{serverCert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert, // enforce mTLS
	}

	server := &http.Server{
		Addr:      ":8443",
		TLSConfig: tlsConfig,
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			fmt.Println("Client Request from:", r.TLS.PeerCertificates[0].Subject.CommonName)
			fmt.Fprintf(w, "Hello, %s!", r.TLS.PeerCertificates[0].Subject.CommonName)
		}),
	}

	log.Println("Starting mTLS server on :8443")
	if err := server.ListenAndServeTLS("", ""); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
