/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package manager

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"runtime"

	"github.com/ernestio/ernest-cli/helper"
	"github.com/ernestio/ernest-cli/model"
	"github.com/ernestio/ernest-cli/view"
	"github.com/fatih/color"
)

// ListServices ...
func (m *Manager) ListServices(token string) (services []model.Service, err error) {
	body, resp, err := m.doRequest("/api/services/", "GET", []byte(""), token, "")
	if err != nil {
		if resp == nil {
			return nil, CONNECTIONREFUSED
		}
		return nil, err
	}
	err = json.Unmarshal([]byte(body), &services)
	if err != nil {
		return nil, err
	}
	return services, err
}

// ServiceStatus ...
func (m *Manager) ServiceStatus(token string, serviceName string) (service model.Service, err error) {
	body, resp, err := m.doRequest("/api/services/"+serviceName, "GET", []byte(""), token, "")
	if err != nil {
		if resp == nil {
			return service, CONNECTIONREFUSED
		}
		if resp.StatusCode == 403 {
			return service, errors.New("You don't have permissions to perform this action")
		}
		if resp.StatusCode == 404 {
			return service, errors.New("Specified service name does not exist")
		}
		return service, err
	}
	if body == "null" {
		return service, errors.New("Unexpected endpoint response : " + string(body))
	}
	err = json.Unmarshal([]byte(body), &service)
	if err != nil {
		return service, err
	}
	return service, nil
}

// ResetService ...
func (m *Manager) ResetService(name string, token string) error {
	s, err := m.ServiceStatus(token, name)
	if err != nil {
		return err
	}
	if s.Status != "in_progress" {
		return errors.New("The service '" + name + "' cannot be reset as its status is '" + s.Status + "'")
	}
	_, resp, err := m.doRequest("/api/services/"+name+"/reset/", "POST", nil, token, "application/yaml")
	if err != nil {
		if resp == nil {
			return CONNECTIONREFUSED
		}
	}
	return err
}

// RevertService reverts a service to a previous known state using a build ID
func (m *Manager) RevertService(name string, buildID string, token string, dry bool) (string, error) {
	// get requested manifest
	s, err := m.ServiceBuildStatus(token, name, buildID)
	if err != nil {
		return "", err
	}
	payload := []byte(s.Definition)

	// apply requested manifest
	var d model.Definition

	err = d.Load(payload)
	if err != nil {
		return "", errors.New("Could not process definition yaml")
	}

	payload, err = d.Save()
	if err != nil {
		return "", errors.New("Could not finalize definition yaml")
	}

	if dry {
		return m.dryApply(token, payload)
	}

	color.Green("Reverting service...")

	streamID := m.GetUUID(token, payload)
	if streamID == "" {
		color.Red("Please log in")
		return "", nil
	}

	go helper.Monitorize(m.URL, "/events", token, streamID)

	if body, resp, err := m.doRequest("/api/services/", "POST", payload, token, "application/yaml"); err != nil {
		if resp == nil {
			return "", CONNECTIONREFUSED
		}
		var internalError struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal([]byte(body), &internalError); err != nil {
			return "", errors.New(body)
		}
		return "", errors.New(internalError.Message)
	}

	runtime.Goexit()

	return streamID, nil
}

// Destroy : Destroys an existing service
func (m *Manager) Destroy(token string, name string, monit bool) error {
	s, err := m.ServiceStatus(token, name)
	if err != nil {
		return err
	}
	if s.Status == "in_progress" {
		return errors.New("The service " + name + " cannot be destroyed as it is currently '" + s.Status + "'")
	}

	body, resp, err := m.doRequest("/api/services/"+name, "DELETE", nil, token, "application/yaml")
	if err != nil {
		if resp == nil {
			return CONNECTIONREFUSED
		}
		if resp.StatusCode == 404 {
			return errors.New("Specified service name does not exist")
		}
		return err
	}

	var res map[string]interface{}
	err = json.Unmarshal([]byte(body), &res)
	if err != nil {
		return err
	}

	if monit == true {
		if str, ok := res["stream_id"].(string); ok {
			helper.Monitorize(m.URL, "/events", token, str)
			runtime.Goexit()
		}
	}

	return nil
}

// ForceDestroy : Destroys an existing service by forcing it
func (m *Manager) ForceDestroy(token, name string) error {
	_, resp, err := m.doRequest("/api/services/"+name+"/force/", "DELETE", nil, token, "application/yaml")
	if err != nil {
		if resp == nil {
			return CONNECTIONREFUSED
		}
		if resp.StatusCode == 404 {
			return errors.New("Specified service name does not exist")
		}
		return err
	}

	return nil
}

// Apply : Applies a yaml to create / update a new service
func (m *Manager) Apply(token string, path string, monit, dry bool) (string, error) {
	var d model.Definition

	payload, err := ioutil.ReadFile(path)
	if err != nil {
		return "", errors.New("You should specify a valid template path or store an ernest.yml on the current folder")
	}

	err = d.Load(payload)
	if err != nil {
		return "", errors.New("Could not process definition yaml")
	}

	// Load any imported files
	if err := d.LoadFileImports(); err != nil {
		return "", err
	}

	payload, err = d.Save()

	if err != nil {
		return "", errors.New("Could not finalize definition yaml")
	}

	if dry {
		return m.dryApply(token, payload)
	}

	color.Green("Environment creation requested")
	fmt.Println("Ernest will show you all output from your requested service creation")
	fmt.Println("You can cancel at any moment with Ctrl+C, even the service is still being created, you won't have any output")

	streamID := m.GetUUID(token, payload)
	if streamID == "" {
		color.Red("Please log in")
		return "", nil
	}

	if monit == true {
		go helper.Monitorize(m.URL, "/events", token, streamID)
	} else {
		fmt.Println("Additionally you can trace your service on ernest monitor tool with id: " + streamID)
	}

	if body, resp, err := m.doRequest("/api/services/", "POST", payload, token, "application/yaml"); err != nil {
		if resp == nil {
			return "", CONNECTIONREFUSED
		}
		var internalError struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal([]byte(body), &internalError); err != nil {
			return "", errors.New(body)
		}
		return "", errors.New(internalError.Message)
	}

	if monit == true {
		runtime.Goexit()
	}
	return streamID, nil
}

// Import : Imports an existing service
func (m *Manager) Import(token string, name string, datacenter string, filters []string) (streamID string, err error) {
	var s struct {
		Name          string   `json:"name"`
		Datacenter    string   `json:"datacenter"`
		ImportFilters []string `json:"import_filters,omitempty"`
	}
	s.Name = name
	s.Datacenter = datacenter
	s.ImportFilters = filters
	payload, err := json.Marshal(s)
	if err != nil {
		return "", errors.New("Invalid name or datacenter")
	}

	color.Green("Environment import requested")
	fmt.Println("Ernest will show you all output from your requested service creation")
	fmt.Println("You can cancel at any moment with Ctrl+C, even the service is still being imported, you won't have any output")

	streamID = m.GetUUID(token, payload)
	if streamID == "" {
		color.Red("Please log in")
		return "", nil
	}

	go helper.Monitorize(m.URL, "/events", token, streamID)

	if body, resp, err := m.doRequest("/api/services/import/", "POST", payload, token, "application/yaml"); err != nil {
		if resp == nil {
			return "", CONNECTIONREFUSED
		}
		return "", errors.New(body)
	}
	runtime.Goexit()

	return streamID, nil
}

// ServiceSync : will call remote /api/services/<myservice/sync
func (m *Manager) ServiceSync(token string, name string) (string, error) {
	body, resp, err := m.doRequest("/api/services/"+name+"/sync/", "POST", []byte(""), token, "application/yaml")
	if err != nil {
		if resp == nil {
			return "", CONNECTIONREFUSED
		}
		return "", errors.New(body)
	}
	runtime.Goexit()

	return body, nil
}

func (m *Manager) dryApply(token string, payload []byte) (string, error) {
	var body string
	body, resp, err := m.doRequest("/api/services/?dry=true", "POST", payload, token, "application/yaml")
	if err != nil {
		if resp == nil {
			return "", CONNECTIONREFUSED
		}
		var internalError struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal([]byte(body), &internalError); err != nil {
			return "", errors.New(body)
		}
		return "", errors.New(internalError.Message)
	}
	view.ServiceDry(body)
	return "", nil
}
