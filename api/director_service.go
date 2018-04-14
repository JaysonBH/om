package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	yamlConverter "github.com/ghodss/yaml"
	yaml "gopkg.in/yaml.v2"
)

//go:generate counterfeiter -o ./fakes/logger.go --fake-name Logger . logger

type logger interface {
	Println(v ...interface{})
}

type DirectorService struct {
	client httpClient
	stderr logger
}

type AvailabilityZoneInput struct {
	AvailabilityZones json.RawMessage `json:"availability_zones"`
}

type AZ struct {
	GUID   string                 `yaml:"guid,omitempty"`
	Name   string                 `yaml:"name"`
	Fields map[string]interface{} `yaml:",inline"`
}

type AvailabilityZones struct {
	AvailabilityZones []*AZ `yaml:"availability_zones"`
}

type NetworkAndAZConfiguration struct {
	NetworkAZ json.RawMessage `json:"network_and_az,omitempty"`
}

type DirectorProperties struct {
	IAASConfiguration     json.RawMessage `json:"iaas_configuration,omitempty"`
	DirectorConfiguration json.RawMessage `json:"director_configuration,omitempty"`
	SecurityConfiguration json.RawMessage `json:"security_configuration,omitempty"`
	SyslogConfiguration   json.RawMessage `json:"syslog_configuration,omitempty"`
}

func NewDirectorService(client httpClient, stderr logger) DirectorService {
	return DirectorService{
		client: client,
		stderr: stderr,
	}
}

func (d DirectorService) SetAZConfiguration(input AvailabilityZoneInput) error {
	azs := AvailabilityZones{}
	err := yaml.Unmarshal(input.AvailabilityZones, &azs.AvailabilityZones)
	if err != nil {
		return fmt.Errorf("provided AZ config is not well-formed JSON: %s", err)
	}

	for i, az := range azs.AvailabilityZones {
		if az.Name == "" {
			return fmt.Errorf("provided AZ config [%d] does not specify the AZ 'name'", i)
		}
	}

	azs, err = d.addGUIDToExistingAZs(azs)
	if err != nil {
		return err
	}

	decoratedConfig, err := yaml.Marshal(azs)
	if err != nil {
		return fmt.Errorf("problem marshalling request: %s", err) // un-tested
	}

	jsonData, err := yamlConverter.YAMLToJSON(decoratedConfig)
	if err != nil {
		return fmt.Errorf("problem converting request to JSON: %s", err) // un-tested
	}

	_, err = d.sendAPIRequest("PUT", "/api/v0/staged/director/availability_zones", jsonData)
	return err
}

func (d DirectorService) SetNetworksConfiguration(input json.RawMessage) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return fmt.Errorf("could not marshal json: %s", err)
	}

	_, err = d.sendAPIRequest("PUT", "/api/v0/staged/director/networks", jsonData)
	return err
}

func (d DirectorService) SetNetworkAndAZ(input NetworkAndAZConfiguration) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return fmt.Errorf("could not marshal json: %s", err)
	}

	_, err = d.sendAPIRequest("PUT", "/api/v0/staged/director/network_and_az", jsonData)
	return err
}

func (d DirectorService) SetProperties(input DirectorProperties) error {
	jsonData, err := json.Marshal(&input)
	if err != nil {
		return fmt.Errorf("could not marshal json: %s", err)
	}

	_, err = d.sendAPIRequest("PUT", "/api/v0/staged/director/properties", jsonData)
	return err
}

func (d DirectorService) addGUIDToExistingAZs(azs AvailabilityZones) (AvailabilityZones, error) {
	existingAzsResponse, err := d.sendAPIRequest("GET", "/api/v0/staged/director/availability_zones", nil)
	if err != nil {
		if existingAzsResponse.StatusCode != http.StatusNotFound {
			return AvailabilityZones{}, fmt.Errorf("unable to fetch existing AZ configuration: %s", err)
		}
	}

	if existingAzsResponse.StatusCode == http.StatusNotFound {
		d.stderr.Println("unable to retrieve existing AZ configuration, attempting to configure anyway")
		return azs, nil
	}

	existingAzsJSON, err := ioutil.ReadAll(existingAzsResponse.Body)
	if err != nil {
		return AvailabilityZones{}, fmt.Errorf("unable to read existing AZ configuration: %s", err) // un-tested
	}

	var existingAZs AvailabilityZones
	err = yaml.Unmarshal(existingAzsJSON, &existingAZs)
	if err != nil {
		return AvailabilityZones{}, fmt.Errorf("problem retrieving existing AZs: response is not well-formed: %s", err)
	}

	for _, az := range azs.AvailabilityZones {
		for _, existingAZ := range existingAZs.AvailabilityZones {
			if az.Name == existingAZ.Name {
				az.GUID = existingAZ.GUID
				break
			}
		}
	}
	return azs, nil
}

func (d DirectorService) sendAPIRequest(verb, endpoint string, jsonData []byte) (*http.Response, error) {
	req, err := http.NewRequest(verb, endpoint, bytes.NewReader(jsonData))
	if err != nil {
		return nil, fmt.Errorf("could not create api request %s %s: %s", verb, endpoint, err.Error())
	}

	req.Header.Add("Content-Type", "application/json")

	resp, err := d.client.Do(req)
	if err != nil {
		return resp, fmt.Errorf("could not send api request to %s %s: %s", verb, endpoint, err.Error())
	}

	err = ValidateStatusOK(resp)
	if err != nil {
		return resp, err
	}

	return resp, nil
}
