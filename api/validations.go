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
	"regexp"
)

// IsValidUUID verifies that the provided string is a valid UUID value.
// According to the Cloud Foundry environment variables, generated ID's are
// referred as GUID which is a synonym for this standard.
// returns false for all empty and unmatched string values
func IsValidUUID(s string) bool {
	// Regex to validate if a string is a valid UUID
	p := "^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$"
	match, err := regexp.MatchString(p, s) // verify pattern with provided string
	if err != nil {
		return false
	}
	return match
}

// IsValidProvisionRequest verifies that request has a valid Body by checking
// that obligatory fields are included in the request and are valid UUID values
func IsValidProvisionRequest(r ProvisionRequest) bool {
	/* Obligatory fields
	ServiceID         string      `json:"service_id"`
	PlanID            string      `json:"plan_id"`
	OrganizationGUID  string      `json:"organization_guid"` // not used by broker but is always sent by CF
	SpaceGUID         string      `json:"space_guid"`        // not used by broker but is always sent by CF
	*/
	switch {
	case IsValidUUID(r.ServiceID) == false:
	case IsValidUUID(r.PlanID) == false:
	case IsValidUUID(r.OrganizationGUID) == false:
	case IsValidUUID(r.SpaceGUID) == false:
	default:
		return true
	}
	return false
}

// IsValidDeprovisionRequest verifies that the provided request is valid for
// the DELETE http method, this is for input validation
func IsValidDeprovisionRequest(r DeprovisionRequest) bool {
	valid := false
	switch {
	case r == DeprovisionRequest{}: // request must not be empty or have zero value
	case IsValidUUID(r.ServiceID) == false: // ServiceId must be a valid UUID
	case IsValidUUID(r.PlanID) == false: // PlanId must also be a valid UUID
	default:
		valid = true // Request meets expectations
	}
	return valid
}
