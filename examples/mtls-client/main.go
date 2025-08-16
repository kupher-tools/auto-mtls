package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"
)

func main() {
	// Read server hostname from env
	serverHost := os.Getenv("MTLS_SERVER_HOST")
	if serverHost == "" {
		log.Fatal("MTLS_SERVER_HOST environment variable not set")
	}
	url := fmt.Sprintf("https://%s:8443", serverHost)

	// Load client certificate
	clientCert, err := tls.LoadX509KeyPair("/etc/tls/tls.crt", "/etc/tls/tls.key")
	if err != nil {
		log.Fatalf("failed to load client cert: %v", err)
	}

	// Load CA certificate to verify server
	caCert, err := os.ReadFile("/etc/ca/ca.crt")
	if err != nil {
		log.Fatalf("failed to read CA cert: %v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// TLS configuration with verbose handshake logging
	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{clientCert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
		VerifyPeerCertificate: func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
			fmt.Println("----- TLS Handshake Info -----")
			for i, chain := range verifiedChains {
				fmt.Printf("Chain %d:\n", i)
				for j, cert := range chain {
					fmt.Printf("  Cert %d: CN=%s, Issuer=%s\n", j, cert.Subject.CommonName, cert.Issuer.CommonName)
				}
			}
			fmt.Println("-------------------------------")
			return nil
		},
	}

	client := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: tlsConfig,
		},
		Timeout: 5 * time.Second,
	}

	// Loop to call server every 2 seconds
	for {
		resp, err := client.Get(url)
		if err != nil {
			log.Printf("request failed: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Printf("reading response failed: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Printf("Server response: %s\n\n", string(body))
		time.Sleep(2 * time.Second)
	}
}
