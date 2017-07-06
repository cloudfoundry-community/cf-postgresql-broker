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

package util

import (
	"bytes"
	"crypto/rsa"
	"crypto/sha512"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
)

/* util.go
A set of utility functions used within this project
*/

//EncodeSha512 uses crypto/sha512 package which implements hash algorithms
//as defined in FIPS 180-4
// Take string value and compute sha512 checksum
// returns hash byte array and  string from processing standard base64 encoding to checksum's array
// of bytes
func EncodeSha512(value string) ([]byte, string) {
	h := sha512.New()
	_, err := h.Write([]byte(value))

	if err != nil {
		panic(err)
	}

	// appends the current hash to b and returns the resulting slice ([]byte)
	c := h.Sum(nil)
	s := base64.StdEncoding.EncodeToString(c)
	return c, s
}

//CheckSha512 validates that a given value's hashed checksum is equal to a
//provided checksum
func CheckSha512(value string, checksum []byte) bool {
	c, _ := EncodeSha512(value)
	return bytes.Equal(c, checksum)
}

// GetCertSignatureAndPublicAlgorithms accepts a certificate's path(string) and
// returns the algorithm used to sign provided certificate.
func GetCertSignatureAndPublicAlgorithms(cert tls.Certificate) (x509.SignatureAlgorithm, x509.PublicKeyAlgorithm, error) {
	// crypto/tls module appends pem.Block.Bytes([]byte) value into cert.Certificate[0]
	// when constructing tls.Certificate type
	c, err := x509.ParseCertificate(cert.Certificate[0])
	//x509.ParseCertificate(block.Bytes)
	if err != nil {
		return 0, 0, err
	}
	return c.SignatureAlgorithm, c.PublicKeyAlgorithm, nil
}

// GetCertBitLength returns the bit length of the tld certificate, currently
// supports only certificates signed with RSA
func GetCertBitLength(cert *tls.Certificate) (int, error) {

	var len int
	switch privKey := cert.PrivateKey.(type) {
	case *rsa.PrivateKey:
		len = privKey.D.BitLen()
	default:
		return 0, errors.New("Unsupported private key")
	}
	return len, nil
}
