package main

import (
	"encoding/json"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"time"
)

type proxiesApi struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Data    []struct {
		Targets  []string `json:"targets"`
		ID       string   `json:"_id"`
		User     string   `json:"user"`
		Options  map[string]interface{}
		Hostname string    `json:"hostname"`
		Expiry   time.Time `json:"expiry"`
		Plan     string    `json:"plan"`
	} `json:"data"`
}

type Proxies map[string]Server

type Target struct {
	IpAddress   string
	Port        string
	Connections int
	Online      bool
}

type Server struct {
	Hostname string
	Targets  []Target
	Options  map[string]interface{}
}

func Update(proxies *Proxies) {
	var proxiesTemp proxiesApi
	getProxies(&proxiesTemp)

	for _, p := range proxiesTemp.Data {
		(*proxies)[p.Hostname] = Server{Hostname: p.Hostname, Targets: createTargets(p.Targets)}
	}
}

func createTargets(t []string) []Target {
	var targets []Target
	for _, target := range t {
		tStr := strings.Split(target, ":")
		tLen := len(tStr)

		tPort := "25565"
		if tLen > 1 {
			tPort = tStr[1]
		}

		// TODO make it figure out if it is online lolz

		targets = append(targets, Target{
			IpAddress:   tStr[0],
			Port:        tPort,
			Connections: 0,
			Online:      true,
		})
	}

	return targets
}

func getProxies(proxies *proxiesApi) {
	p, err := apiPull("proxies")
	if err != nil {
		log.Fatal(err)
	}
	json.Unmarshal(p, &proxies)
}

func apiPull(path string) ([]byte, error) {
	var apiURL = strings.ReplaceAll(os.Getenv("API_URL"), "%key%", os.Getenv("API_KEY"))
	resp, err := http.Get(strings.ReplaceAll(apiURL, "%path%", path))
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	return body, err
}
