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

// Define object types for broker API

// CatalogObject - /v2/catalog
type CatalogObject struct {
	Services []Service `json:"services"`
}

// Service implements the structure defined on CF's Service Broker API,
// returned in the /v2/catalog endpoint
type Service struct {
	Name           string          `json:"name"`
	ID             string          `json:"id"`
	Description    string          `json:"description"`
	Tags           []string        `json:"tags"`
	Requires       []string        `json:"requires"`
	Bindable       bool            `json:"bindable"`
	Metadata       []interface{}   `json:"metadata"`
	DClient        DashboardClient `json:"dashboard_client"`
	PlanUpdateable bool            `json:"plan_updateable"`
	Plans          []Plan          `json:"plans"`
}

// DashboardClient implements object as defined on CF's Service Broker api
type DashboardClient struct {
	ID          string `json:"id"`
	Secret      string `json:"secret"`
	RedirectURI string `json:"redirect_uri"`
}

// Plan implements object as defined on CF's Service Broker api
type Plan struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Metadata    []interface{} `json:"metadata"`
	Free        bool          `json:"free"`
	Bindable    bool          `json:"bindable"`
}

// ProvisionRequest is the Body struct expected from requests
// PUT/v2/service_instances/:instance_id
// expected body
type ProvisionRequest struct {
	ServiceID         string      `json:"service_id"`
	PlanID            string      `json:"plan_id"`
	Parameters        interface{} `json:"parameters"`
	AcceptsIncomplete bool        `json:"accepts_incomplete"`
	OrganizationGUID  string      `json:"organization_guid"`
	SpaceGUID         string      `json:"space_guid"`
}

// ProvisionResponse as specified in CF's Service Broker api responds a valid
// json object, at least "{}
type ProvisionResponse struct {
	DashboardURL string   `json:"dashboard_url"`
	Database     DataBase `json:"database"`
}

// DeprovisionRequest type specification
// service_id*        string
// plan_id*           string
// accepts_incomplete boolean
// required fields marked with [*]
type DeprovisionRequest struct {
	ServiceID         string `json:"service_id"`
	PlanID            string `json:"plan_id"`
	AcceptsIncomplete bool   `json:"accepts_incomplete"`
}

// DeprovisionResponse type specification
// vars [id]
// Request body
// service_id* - string
// plan_id* - string
// accepts_incomplete - bolean
// Response might be just {} for mvp
type DeprovisionResponse struct {
	ID     string `json:"id"`
	Status string `json:"status"`
}

// ServiceInstance type holds required data in order to provision service with
// docker engine
type ServiceInstance struct {
	ID      string
	Port    int
	Info    string
	Service string
}

// Inspect type is used to consult running service instance on docker engine,
// allows to retrieve details from docker's cli output
type Inspect struct {
	List []Container
}

// DataBase type is used to define a database service to be provisioned by the
// service broker implementation
type DataBase struct {
	Name     string `json:"name"`
	Status   string `json:"status"`
	Provider string `json:"provider"`
}

// NetworkSettings type allows consultation on network data from a running
// service on the docker engine(used in conjunction with Inspect struct)
type NetworkSettings struct {
	IPAddress string
}

// Container type used in conjunction with Inspect and NetworkSettings type to
// retrieve data from docker cli
type Container struct {
	NetworkSettings NetworkSettings
}

// Empty type used for marshalling empty jsons on byte slices to return in a
// response body
type Empty struct{}
