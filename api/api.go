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
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// Catalog is executed when /v2/catalog is called via HTTP GET method
// It returns catalog of available services
func Catalog(w http.ResponseWriter, r *http.Request) {
	plans := []Plan{
		{
			ID:          "83c8811b-f3db-17ef-6eb3-bbe944b47262",
			Name:        "5mb",
			Description: "5 mb of psql database",
			Metadata:    nil,
			Free:        true,
			Bindable:    true,
		},
		{
			ID:          "39292da3-de98-2891-11f0-c36a3264dbb5",
			Name:        "50mb",
			Description: "50 mb of psql database",
			Free:        true,
			Bindable:    true,
		},
	}
	dashb := DashboardClient{
		ID:          "test",
		Secret:      "test",
		RedirectURI: "http://localhost:9000"}

	data := Service{
		Name:           "posgreSQL",
		ID:             "1",
		Description:    "A postgresql DB service",
		Tags:           []string{},
		Requires:       []string{},
		Bindable:       true,
		Metadata:       nil,
		DClient:        dashb,
		PlanUpdateable: true,
		Plans:          plans,
	}
	catalog := CatalogObject{
		[]Service{data},
	}

	js, err := json.Marshal(catalog)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(js)
	if err != nil {
		log.Print(err)
	}
}

// Provision /create a DB
// vars [id]
//Expected Body:
// *organization_guid - string
// *plan_id - string
// *service_id - string
// *space_guid - string
// parameters -json obj
// accepts_incomplete - boolean
func (h *DbHandler) Provision(w http.ResponseWriter, r *http.Request) {
	var status int
	var body []byte
	var err error
	var werr error

	//Print write errors if any
	defer func(e *error) {
		if e != nil {
			log.Print(e)
		}
	}(&werr)
	//close body if exists
	defer func() {
		if r.Body != nil {
			e := r.Body.Close()
			if e != nil {
				log.Print(e.Error())
			}
		}
	}()

	defer func(s *int, b *[]byte) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(*s)
		_, err = w.Write(*b)
		if err != nil {
			log.Print(err)
		}
	}(&status, &body)

	vars := mux.Vars(r)
	id := vars["id"]

	// Input validation: check if id is a valid UUID string
	if IsValidUUID(id) == false {
		status = http.StatusBadRequest
		writeEmptyJSON(&body)
		return
	}

	// Input validation: provided body on request must be compliant with the
	// definition of the ProvisionRequest type
	decoder := json.NewDecoder(r.Body)
	var provisionRequest ProvisionRequest
	err = decoder.Decode(&provisionRequest)
	if err != nil {
		status = http.StatusBadRequest
		writeEmptyJSON(&body)
		return
	}

	//Input validation, verify that provisionRequest has valid required fields
	if IsValidProvisionRequest(provisionRequest) == false {
		status = http.StatusBadRequest
		writeEmptyJSON(&body)
		return
	}

	si, err := h.Add(id, provisionRequest)
	if err != nil {
		status = http.StatusInternalServerError
		writeEmptyJSON(&body)
		return
	}

	//define new DB
	db := new(DataBase)
	db.Name = si.ID
	db.Status = si.Info
	db.Provider = si.Service

	resp := new(ProvisionResponse)
	resp.DashboardURL = "" + si.ID + ";" + strconv.Itoa(si.Port) + ";" + si.Info
	resp.Database = *db
	responseBody, err := json.Marshal(resp)
	if err != nil {
		status = http.StatusBadRequest
		writeEmptyJSON(&body)
		return
	}

	status = http.StatusCreated // set status created for new instance
	body = responseBody
}

//Deprovision deletes a DB service instance
// expected status codes are 200, 202, 410 and 422 according to Service Broker API
// specification
func (h *DbHandler) Deprovision(w http.ResponseWriter, r *http.Request) {
	var status int
	var body []byte
	var err error

	defer func(s *int, b *[]byte) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(*s)
		_, err = w.Write(*b)
		if err != nil {
			log.Print(err)
		}
	}(&status, &body)

	vars := mux.Vars(r)
	id := vars["id"]

	// Input validation: check if id is a valid UUID string
	if IsValidUUID(id) == false {
		//invalid request implies status code 422
		status = http.StatusBadRequest
		writeEmptyJSON(&body)
		return
	}

	// Verify if request Query params are valid by decoding it into DeprovisionRequest
	// type
	// 1. Did request sent a query params?
	params := r.URL.Query()
	if params == nil {
		status = http.StatusBadRequest
		writeEmptyJSON(&body)
		return
	}
	// assign query values to deprovisionRequest fields
	var deprovisionRequest DeprovisionRequest
	deprovisionRequest.ServiceID = params.Get("service_id")
	deprovisionRequest.PlanID = params.Get("plan_id")
	switch params.Get("accepts_incomplete") {
	case "true":
		deprovisionRequest.AcceptsIncomplete = true
	default:
		// Default behavior assings 'bool' type  zero value
		deprovisionRequest.AcceptsIncomplete = false
	}

	// Verify if DeprovisionRequest values are valid and expected
	if valid := IsValidDeprovisionRequest(deprovisionRequest); valid == false {
		status = http.StatusBadRequest
		writeEmptyJSON(&body)
		return
	}

	rowsAffected, err := h.Remove(id)
	log.Print("Delete rows affected:", rowsAffected)
	if err != nil {
		// Errors on DB are unexpected and imply internal Broker errors
		status = http.StatusInternalServerError
		writeEmptyJSON(&body)
		return
	}
	if rowsAffected == 0 {
		status = http.StatusGone
		writeEmptyJSON(&body)
		return
	}

	status = http.StatusOK
	body, _ = json.Marshal(DeprovisionResponse{
		ID:     id,
		Status: "destroyed",
	})
}

func writeEmptyJSON(body *[]byte) {
	*body, _ = json.Marshal(Empty{})
}
