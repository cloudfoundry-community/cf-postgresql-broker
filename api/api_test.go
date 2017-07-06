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

package api

import (
	"bytes"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/gorilla/mux"
)

const (
	testID       = "41653aa4-3a3a-486a-4431-ef258b39f042"
	inexistentID = "00000000-0000-0000-0000-000000000000"
	invalidID    = "1234567"
)

func TestBrokerCatalog(t *testing.T) {
	// testing api catalog endpoint
	req, err := http.NewRequest("GET", "/v2/catalog", nil)
	if err != nil {
		t.Fatal(err)
	}

	// Create response recorder to satisfy http.ResponseWriter
	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Catalog)

	handler.ServeHTTP(rr, req)
	//Verify status
	if status := rr.Code; status != http.StatusOK {
		t.Error("Catalog: return wrong status function, obtained ", status)
	}
}

// TestBrokerProvision validates the procesing of the provision function served
// using HTTP POST
func TestBrokerProvision(t *testing.T) {
	// testing api provision endpoint
	// test json string for request

	// json must represent the 'ProvisionRequest' type in order to be a valid
	// request
	jsonStr := []byte(`
	{
		"service_id":"41653aa4-3a3a-486a-4431-ef258b39f042",
		"plan_id":"41653aa4-3a3a-486a-4431-ef258b39f042",
		"parameters":{},
		"accepts_incomplete":true,
		"organization_guid":"41653aa4-3a3a-486a-4431-ef258b39f042",
		"space_guid":"41653aa4-3a3a-486a-4431-ef258b39f042"
	}
	`)
	invalidJSONStr := []byte(`
	{
		"service_id":"123",
		"plan_id":"123",
		"parameters":{},
		"accepts_incomplete":true,
		"organization_guid":"123",
		"space_guid":"123"
	}
	`)

	req, err := http.NewRequest("PUT", "/v2/service_instances/"+testID, bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}

	// Create invalid request, expected results of such calls must be error
	// responses
	// invalid request 1 has an invlid ID on the endpoint URI
	ir1, err := http.NewRequest("PUT", "/v2/service_instances/"+invalidID, bytes.NewBuffer(jsonStr))
	if err != nil {
		t.Fatal(err)
	}
	// invalid request 2 has an invalid Body on it's request
	ir2, err := http.NewRequest("PUT", "/v2/service_instances/"+testID, bytes.NewBuffer(invalidJSONStr))
	if err != nil {
		t.Fatal(err)
	}

	// Create response recorder to satisfy http.ResponseWriter
	rr := httptest.NewRecorder()
	// setup handler
	dbhandler := DbHandler{Name: "sqlite3", Path: "./foo.db"}
	dbhandler.Setup() // setup database

	//Set Mux
	r := mux.NewRouter()
	r.HandleFunc("/v2/service_instances/{id}", dbhandler.Provision).Methods("PUT")
	r.ServeHTTP(rr, req)

	//Verify status
	bodyText, err := ioutil.ReadAll(rr.Body)
	if err != nil {
		panic(err)
	}
	// In provision and according to CF's API, code 201 (Status created) must
	// be returned upon success
	if status := rr.Code; status != http.StatusCreated {
		t.Error("Provision: return wrong status function, obtained ", status,
			"\nDetail: "+string(bodyText))
	}
	rr = httptest.NewRecorder()
	// Test invalid inputs on request, these should return errors
	r.ServeHTTP(rr, ir1)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Error("Invalid request expects status code 400(bad request) got ", status)
	}

	rr = httptest.NewRecorder()
	r.ServeHTTP(rr, ir2)
	if status := rr.Code; status != http.StatusBadRequest {
		t.Error("Invalid request expects status code 400 (Bad Request) got ", status)
	}

	time.Sleep(500 * time.Millisecond)
	// Validate that instance can be deprovisioned
	brokerDeprovision(t, dbhandler)
	//test unexpectedDeprovision
	testUnexpectedDeprovision(t)

}

func testUnexpectedDeprovision(t *testing.T) {
	// setup handler
	dbhandler := DbHandler{Name: "sqlite3", Path: "./foo.db"}
	dbhandler.Setup() // setup database

	queryValues := url.Values{}
	queryValues.Add("service_id", "41653aa4-3a3a-486a-4431-ef258b39f042")
	queryValues.Add("plan_id", "41653aa4-3a3a-486a-4431-ef258b39f042")

	// testing api deprovision endpoint
	req, err := http.NewRequest("DELETE", "/v2/service_instances/"+inexistentID+"?"+
		queryValues.Encode(), nil)
	if err != nil {
		t.Error(err)
	}

	// Create response recorder to satisfy http.ResponseWriter
	rr := httptest.NewRecorder()

	//Set Mux
	r := mux.NewRouter()
	r.HandleFunc("/v2/service_instances/{id}", dbhandler.Deprovision).Methods("DELETE")
	r.ServeHTTP(rr, req)

	//Verify expected status Gone(410) for inexistent services (according to Service Broker
	//API definition)
	if status := rr.Code; status != http.StatusGone {
		t.Error("Deprovision: expected code is 410 'Gone' got ", status)
	}

}

func brokerDeprovision(t *testing.T, dbhandler DbHandler) {
	// testing api deprovision endpoint
	queryValues := url.Values{}
	queryValues.Add("service_id", "41653aa4-3a3a-486a-4431-ef258b39f042")
	queryValues.Add("plan_id", "41653aa4-3a3a-486a-4431-ef258b39f042")

	req, err := http.NewRequest("DELETE", "/v2/service_instances/"+testID+"?"+
		queryValues.Encode(), nil)
	if err != nil {
		t.Error(err)
	}

	// Create response recorder to satisfy http.ResponseWriter
	rr := httptest.NewRecorder()

	//Set Mux
	r := mux.NewRouter()
	r.HandleFunc("/v2/service_instances/{id}", dbhandler.Deprovision).Methods("DELETE")
	r.ServeHTTP(rr, req)

	//Verify status
	if status := rr.Code; status != http.StatusOK {
		t.Error("Deprovision: return wrong status function, obtained ", status)
	}
}
