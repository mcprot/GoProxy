package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

type ConfigJson struct {
	Routes map[string]string `json:"routes"`
	Test   string            `json:"test"`
}

func loadConfig(filename string) (*ConfigJson, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}

	config := ConfigJson{}
	data, err := ioutil.ReadAll(f)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(data, &config)
	return &config, nil

}

func (c *ConfigJson) getDestinationAddress(addr string) string {
	dest, ok := c.Routes[addr]
	if ok {
		return dest
	}

	defaultDest, ok := c.Routes["default"]
	if ok {
		return defaultDest
	}
	return ""
}
