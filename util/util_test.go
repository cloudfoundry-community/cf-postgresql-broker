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
	"crypto/rand"
	"encoding/base64"
	"testing"
)

const (
	bSize int = 10 // byte array size for random generation
)

func TestEncodeSha512(t *testing.T) {
	// generate random string
	b := make([]byte, bSize)
	// read bSize cryptographically secure pseudorandom numbers from
	// rand.Reader and writes to b
	_, err := rand.Read(b)
	if err != nil {
		t.Error("Random number not properly generated")
	}
	s := base64.StdEncoding.EncodeToString(b)
	c, _ := EncodeSha512(s)
	b, err = base64.StdEncoding.DecodeString(s)
	if err != nil {
		t.Fail()
	}
	if bytes.Equal(b, c) == true {
		t.Error("Original string and encoded string are equal")
	}

}

func TestCheckSha512(t *testing.T) {
	// generate random string
	b := make([]byte, bSize)
	// read bSize cryptographically secure pseudorandom numbers from
	// rand.Reader and writes to b
	_, err := rand.Read(b)
	if err != nil {
		t.Error("Random number not properly generated")
	}
	s := base64.StdEncoding.EncodeToString(b)

	cs, _ := EncodeSha512(s)
	if CheckSha512(s, cs) == false {
		t.Error("Checksum for the same origin value doesn't match")
	}
}
