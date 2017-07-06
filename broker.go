/*
// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
*/

package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"

	"github.com/cloudfoundry-community/cf-postgresql-broker/api"
	"github.com/gorilla/mux"
)

const (
	keyFlag  = "key"
	certFlag = "cert"
	address  = ":8080"
)

func main() {
	// Define and parse flags
	var k = flag.String(keyFlag, "", "usage -key=filename")
	var c = flag.String(certFlag, "", "usage -cert=filename")
	flag.Parse()
	// Retrieve TLS certFile and keyFile from flag pointers
	keyFile := *k
	certFile := *c
	kptr := flag.Lookup(keyFlag)
	cptr := flag.Lookup(certFlag)

	// Verify there are no positional arguments, which are not expected by the
	// software
	if len(flag.Args()) != 0 {
		flag.Usage()
		return
	}

	// Verify flags are being used through the command-line, if not, end
	// program
	switch {
	case kptr == nil:
		log.Fatal("key flag is required.")
	case cptr == nil:
		log.Fatal("cert flag is required.")
	}

	// Verify zero values (flags are not empty)
	if keyFile == "" || certFile == "" {
		log.Println("Invalid usage")
		log.Println(kptr.Usage)
		log.Println(cptr.Usage)
	}

	r := mux.NewRouter()
	handler := api.DbHandler{Name: "sqlite3", Path: "./foo.db"}
	handler.Setup() // setup database

	r.HandleFunc("/v2/catalog", api.Catalog).
		Methods("GET")

	r.HandleFunc("/v2/service_instances/{id}", handler.Provision).
		Methods("PUT")

	r.HandleFunc("/v2/service_instances/{id}", handler.Deprovision).
		Methods("DELETE")

	http.Handle("/", r)

	// Verify if key and certificate meet minimum security policies, terminate
	// program on failure
	pass, cert, err := MeetsPolicies(certFile, keyFile)
	switch {
	case err != nil:
		log.Fatal(err)
	case pass == false:
		log.Fatal("Minimum security policies not met, terminating")
	}

	// Set tls configurations
	tlsConfig := tls.Config{
		Certificates: []tls.Certificate{cert},
	}
	SetPreferredCipherSuites(&tlsConfig)
	server := http.Server{
		Addr:      address,
		TLSConfig: &tlsConfig,
	}
	err = server.ListenAndServeTLS("", "")
	if err != nil {
		log.Fatal(err.Error())
	}

}
