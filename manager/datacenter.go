/* This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/. */

package manager

import (
	"encoding/json"
	"errors"
	"strconv"

	"github.com/ernestio/ernest-cli/model"
)

// CreateVcloudDatacenter : Creates a VCloud datacenter
func (m *Manager) CreateVcloudDatacenter(token, name, rtype, user, password, url, network, vseURL string) (string, error) {
	payload := []byte(`{"name": "` + name + `", "type":"` + rtype + `", "region": "", "username":"` + user + `", "password":"` + password + `", "external_network":"` + network + `", "vcloud_url":"` + url + `", "vse_url":"` + vseURL + `"}`)
	body, res, err := m.doRequest("/api/datacenters/", "POST", payload, token, "")
	if err != nil {
		if res.StatusCode == 409 {
			return "Datacenter '" + name + "' already exists, please specify a different name", err
		}
		return body, err
	}
	return body, err
}

// CreateAWSDatacenter : Creates an AWS datacenter
func (m *Manager) CreateAWSDatacenter(token, name, rtype, region, awsAccessKeyID, awsSecretAccessKey string) (string, error) {
	payload := []byte(`{"name": "` + name + `", "type":"` + rtype + `", "region":"` + region + `", "username":"` + name + `", "aws_access_key_id":"` + awsAccessKeyID + `", "aws_secret_access_key":"` + awsSecretAccessKey + `"}`)
	body, res, err := m.doRequest("/api/datacenters/", "POST", payload, token, "")
	if err != nil {
		if res.StatusCode == 409 {
			return "Datacenter '" + name + "' already exists, please specify a different name", err
		}
		return body, err
	}
	return body, err
}

// CreateAzureDatacenter : Creates an Azure datacenter
func (m *Manager) CreateAzureDatacenter(token, name, rtype, subscription, client, secret, tenant, environment string) (string, error) {
	payload := []byte(`{"name": "` + name + `", "type":"` + rtype + `", "azure_subscription_id": "` + subscription + `", "azure_client_id":"` + client + `", "azure_client_secret":"` + secret + `", "azure_tenant_id":"` + tenant + `", "azure_environment":"` + environment + `"}`)
	body, res, err := m.doRequest("/api/datacenters/", "POST", payload, token, "")
	if err != nil {
		if res.StatusCode == 409 {
			return "Datacenter '" + name + "' already exists, please specify a different name", err
		}
		return body, err
	}
	return body, err
}

// ListDatacenters : Lists all datacenters on your account
func (m *Manager) ListDatacenters(token string) (datacenters []model.Datacenter, err error) {
	body, _, err := m.doRequest("/api/datacenters/", "GET", []byte(""), token, "")
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal([]byte(body), &datacenters)
	if err != nil {
		return nil, err
	}
	return datacenters, err
}

// DeleteDatacenter : Deletes an existing datacenter by its name
func (m *Manager) DeleteDatacenter(token string, name string) (err error) {
	g, err := m.getDatacenterByName(token, name)
	if err != nil {
		return errors.New("Datacenter '" + name + "' does not exist, please specify a different datacenter name")
	}
	id := strconv.Itoa(g.ID)

	body, res, err := m.doRequest("/api/datacenters/"+id, "DELETE", []byte(""), token, "")
	if err != nil {
		if res.StatusCode == 400 {
			return errors.New(body)
		}

		return err
	}
	return nil
}

// UpdateVCloudDatacenter : updates vcloud datacenter details
func (m *Manager) UpdateVCloudDatacenter(token, name, user, password string) (err error) {
	g, err := m.getDatacenterByName(token, name)
	if err != nil {
		return errors.New("Datacenter '" + name + "' does not exist, please specify a different datacenter name")
	}
	id := strconv.Itoa(g.ID)

	payload := []byte(`{"username":"` + user + `", "password":"` + password + `"}`)
	body, res, err := m.doRequest("/api/datacenters/"+id, "PUT", payload, token, "")
	if err != nil {
		if res.StatusCode == 400 {
			return errors.New(body)
		}

		return err
	}

	return nil
}

// UpdateAWSDatacenter : updates awsdatacenter details
func (m *Manager) UpdateAWSDatacenter(token, name, awsAccessKeyID, awsSecretAccessKey string) (err error) {
	g, err := m.getDatacenterByName(token, name)
	if err != nil {
		return errors.New("Datacenter '" + name + "' does not exist, please specify a different datacenter name")
	}
	id := strconv.Itoa(g.ID)

	payload := []byte(`{"aws_access_key_id":"` + awsAccessKeyID + `", "aws_secret_access_key":"` + awsSecretAccessKey + `"}`)
	body, res, err := m.doRequest("/api/datacenters/"+id, "PUT", payload, token, "")
	if err != nil {
		if res.StatusCode == 400 {
			return errors.New(body)
		}

		return err
	}

	return nil
}
