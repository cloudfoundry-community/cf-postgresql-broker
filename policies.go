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
	"crypto/x509"

	"github.com/cloudfoundry-community/cf-postgresql-broker/util"
)

const (
	minBitLength = 2048
	algRSA       = "RSA"
	algECDSA     = "ECDSA"
)

// SetPreferredCipherSuites will receive a tls.Config pointer and set in
// CipherSuites array, in addition set PreferServerCipherSuites as true
func SetPreferredCipherSuites(config *tls.Config) {
	config.CipherSuites = []uint16{
		// Prefer suites with forward secrecy, then larger ciphers
		tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
		tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
		// Omit suites without forward secrecy if you control the client and
		// it supports the suites above
		tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
		tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
		tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
	}
	config.PreferServerCipherSuites = true
}

// MeetsPolicies takes in the key and certificate's path and evaluates all
// defined policies, returns true if all security policies are met.
func MeetsPolicies(certFile string, keyFile string) (bool, tls.Certificate, error) {
	var err error
	var p bool
	var cert tls.Certificate
	pass := true
	// Load certificate
	cert, err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return false, cert, err
	}

	p = MeetsCertBitLengthPolicy(cert)
	pass = pass && p

	p = MeetsSignatureAlgorithmPolicy(cert)
	pass = pass && p

	return pass, cert, err
}

// MeetsCertBitLengthPolicy evaluates certificate's bit length which must be
// 2048 or more. Return true if policy is met
func MeetsCertBitLengthPolicy(cert tls.Certificate) bool {
	// For security, validate if certificate's bit length is at least 2048
	len, err := util.GetCertBitLength(&cert)
	if err != nil {
		return false
	}

	if len < minBitLength {
		return false
	}
	return true
}

// MeetsSignatureAlgorithmPolicy function evaluates key and certificate files
// as string values
// This policy returns true if the following minimum requirements are met
//- Signature algorithm is SHA256, SHA384 or SHA512
func MeetsSignatureAlgorithmPolicy(cert tls.Certificate) bool {
	// For security validate if signature algorithm is sha256, sha384 or sha512
	sa, pa, err := util.GetCertSignatureAndPublicAlgorithms(cert)
	if err != nil {
		return false
	}
	// Determine signature algorithm and it's family
	var f string
	switch {
	case sa == x509.SHA256WithRSA:
		f = algRSA
	case sa == x509.SHA384WithRSA:
		f = algRSA
	case sa == x509.SHA512WithRSA:
		f = algRSA
	case sa == x509.ECDSAWithSHA256:
		f = algECDSA
	case sa == x509.ECDSAWithSHA384:
		f = algECDSA
	case sa == x509.ECDSAWithSHA512:
		f = algECDSA

	default:
		return false
	}

	// Check if public key algorithm matches signature's algorithm type
	switch {
	// both signature and public key algorithms are of the same family
	case f == algRSA && pa == x509.RSA:
	case f == algECDSA && pa == x509.ECDSA:
	// signature and public key algorithms are of different families
	default:
		return false
	}

	return true
}
