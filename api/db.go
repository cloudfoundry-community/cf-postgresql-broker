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
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os/exec"
	"strconv"

	_ "github.com/mattn/go-sqlite3" // Blank import according to go-sqlite3's instructions
)

const (
	table           = "service_instance"
	defaultUser     = "postgres"
	defaultPassword = "mysecretpassword"
)

//DbHandler type holds DB connection and basic data
type DbHandler struct {
	Name string
	Path string
	Open bool
	db   *sql.DB
}

// Remove service registry from database
func (h *DbHandler) Remove(instance string) (int, error) {
	var err error

	h.db, _ = sql.Open(h.Name, h.Path)
	defer func() {
		err = h.db.Close()
		if err != nil {
			log.Print(err.Error())
		}
	}()

	// delete docker container
	cmd := exec.Command("docker", "stop", instance)
	err = cmd.Run()
	cmd = exec.Command("docker", "rm", instance)
	err = cmd.Run()

	// delete from database
	deleteQuery := "DELETE FROM " + table + " WHERE " + table + ".id = '" + instance + "'"
	res, err := h.db.Exec(deleteQuery)
	rows, _ := res.RowsAffected()
	if err != nil {
		return 0, err
	}
	return int(rows), nil
}

// Add service registry into database
func (h *DbHandler) Add(instance string, pr ProvisionRequest) (ServiceInstance, error) {
	var err error
	var rows *sql.Rows

	h.db, _ = sql.Open(h.Name, h.Path)
	defer func() {
		e := h.db.Close()
		if e != nil {
			log.Print(e.Error())
		}
	}()

	// Get the Server port to use
	rows, err = h.db.Query("SELECT MAX(port) FROM " + table + " WHERE port IS NOT NULL;")

	if err != nil {
		return ServiceInstance{}, err
	}
	rows.Next()

	// Will scan port into empty interface, this allows for nil checks in case
	// table is empty
	var portScan interface{}
	var port int
	err = rows.Scan(&portScan)
	if err != nil {
		return ServiceInstance{}, err
	}
	if portScan == nil {
		port = 0
	} else {
		port = int(portScan.(int64))
	}
	if port == 0 {
		port = 5432
	} else {
		port = port + 1
	}

	err = rows.Close()
	if err != nil {
		log.Print(err.Error())
	}
	insertQuery := "INSERT INTO " + table + "(id, service, port, info) " +
		"VALUES('" + instance + "','" + "psql'," + strconv.Itoa(port) + ",'example');"

	_, err = h.db.Exec(insertQuery)
	if err != nil {
		return ServiceInstance{}, err
	}

	si := ServiceInstance{}
	si.ID = instance
	si.Info = "default"
	si.Port = port
	si.Service = "PostgreSQL"

	cmd := exec.Command("docker", "run", "--name", si.ID,
		"-e", "POSTGRES_PASSWORD="+defaultPassword,
		"-e", "POSTGRES_USER="+defaultUser,
		"-P", // assigns free port automatically
		"-d", "postgres")
	errCmd := cmd.Run()
	if errCmd != nil {
		log.Fatal(errCmd)
	}
	// Docker inspect
	cmd = exec.Command("docker", "inspect", si.ID)
	stdout, _ := cmd.StdoutPipe()
	//reader := bufio.NewScanner(stdout)

	errCmd = cmd.Start()
	if errCmd != nil {
		log.Fatal(errCmd)
	}

	fmt.Println("docker executed:")
	//for reader.Scan() {
	//	fmt.Println(reader.Text())
	//}
	decoder := json.NewDecoder(stdout)
	var inspect []interface{}
	err = decoder.Decode(&inspect)
	if err != nil {
		return ServiceInstance{}, err
	}
	if ew := cmd.Wait(); ew != nil {
		log.Fatal(ew)
	}
	// iptables
	object := inspect[0]
	network := object.(map[string]interface{})["NetworkSettings"]
	ip := network.(map[string]interface{})["IPAddress"]
	fmt.Println(ip.(string))

	cmd = exec.Command("iptables", "-t", "nat", "-A", "DOCKER", "-p", "tcp", "--dport", strconv.Itoa(si.Port), "-j",
		"DNAT", "--to-destination", ip.(string)+":5432")
	_ = cmd.Run()
	return si, nil
}

//Setup sqlite database
func (h *DbHandler) Setup() {
	h.db, _ = sql.Open(h.Name, h.Path)
	d := h.db
	var err error
	createTableQuery := "CREATE TABLE IF NOT EXISTS " + table +
		"(id TEXT, " +
		"service TEXT, " +
		"port INTEGER, " +
		"info TEXT);"
	_, err = d.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}
	err = d.Close()
	if err != nil {
		//Unavailability to setup sqlite Db suggest failure
		panic(err)
	}
}
