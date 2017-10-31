/*
 * Copyright (c) 2001-2017 TIBCO Software Inc.
 * All Rights Reserved. Confidential & Proprietary.
 * For more information, please contact:
 * TIBCO Software Inc., Palo Alto, California, USA
 *
 * $Id: eftl.go 92045 2017-03-04 00:29:04Z $
 */

package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"io/ioutil"
	"log"
	"time"

	"tibco.com/eftl"
)

/*
  This is an example of a basic eFTL client which publishes
  messages to the specified destination.
*/

func main() {

	urlPtr := flag.String("url", "ws://localhost:9191/channel", "server URL")
	clientIDPtr := flag.String("clientid", "", "unique client identifier")
	usernamePtr := flag.String("username", "", "username for authentication")
	passwordPtr := flag.String("password", "", "password for authentication")
	destinationPtr := flag.String("destination", "sample", "destination on which to publish messages")
	caPtr := flag.String("ca", "", "server CA certificate PEM file")

	flag.Parse()

	var tlsConfig *tls.Config

	if *caPtr != "" {
		// TLS configuration uses CA certificate from a PEM file to
		// authenticate the server certificate when using wss:// for
		// a secure connection
		caCert, err := ioutil.ReadFile(*caPtr)
		if err != nil {
			log.Printf("unable to load CA file: %s", err)
			return
		}
		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		tlsConfig = &tls.Config{
			RootCAs: caCertPool,
		}
	} else {
		// TLS configuration accepts all server certificates
		// when using wss:// for a secure connection
		tlsConfig = &tls.Config{
			InsecureSkipVerify: true,
		}
	}

	// channel for receiving connection errors
	errChan := make(chan error, 1)

	// set connection options
	opts := &eftl.Options{
		ClientID:  *clientIDPtr,
		Username:  *usernamePtr,
		Password:  *passwordPtr,
		TLSConfig: tlsConfig,
	}

	// connect to the server
	conn, err := eftl.Connect(*urlPtr, opts, errChan)
	if err != nil {
		log.Printf("connect failed: %s", err)
		return
	}

	// close the connection when done
	defer conn.Disconnect()

	// channel for receiving publish completions
	compChan := make(chan *eftl.Completion, 1000)

	// publish messages once per second
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case now := <-ticker.C:
			// publish the message
			conn.PublishAsync(eftl.Message{
				"_dest": *destinationPtr,
				"now":   now,
			}, compChan)
		case comp := <-compChan:
			if comp.Error != nil {
				log.Printf("publish operation failed: %s", comp.Error)
				return
			}
			log.Printf("published message: %s", comp.Message)
		case err := <-errChan:
			log.Printf("connection error: %s", err)
			return
		}
	}
}
