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
		Targets  []string          `json:"targets"`
		ID       string            `json:"_id"`
		User     string            `json:"user"`
		Options  map[string]string `json:"options"`
		Hostname string            `json:"hostname"`
		Expiry   time.Time         `json:"expiry"`
		Plan     string            `json:"plan"`
	} `json:"data"`
}

type plansApi struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Data    []struct {
		ID          string `json:"_id"`
		Connections int    `json:"connections"`
		Custom      bool   `json:"custom"`
	} `json:"data"`
}

type analyticsApi struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Data    []struct {
		Connections map[string]int `json:"connections"`
		ID          string         `json:"_id"`
		ProxyID     string         `json:"proxy_id"`
	} `json:"data"`
}

type serversApi struct {
	Message string `json:"message"`
	Status  int    `json:"status"`
	Data    []struct {
		ID        string `json:"_id"`
		IPAddress string `json:"ip_address"`
		APIKey    string `json:"api_key"`
	} `json:"data"`
}

type Proxies map[string]Server

type Target struct {
	IPAddress   string
	Port        string
	Connections int
	Online      bool
}

type Server struct {
	Hostname string
	Targets  []Target
	Options  map[string]interface{}
	Plan     string
}

func Update(proxies *Proxies, signer Signer) {
	var proxiesTemp proxiesApi
	getProxies(&proxiesTemp)

	for _, p := range proxiesTemp.Data {
		(*proxies)[p.Hostname] = Server{Hostname: p.Hostname, Targets: createTargets(p.Targets, signer)}
	}
}

func createTargets(t []string, signer Signer) []Target {
	var targets []Target
	for _, target := range t {
		tStr := strings.Split(target, ":")
		tLen := len(tStr)

		tPort := "25565"
		if tLen > 1 {
			tPort = tStr[1]
		}

		online := true

		hostStr := tStr[0] + ":" + tPort

		// TODO make it figure out if it is online lolz
		if !QueryStatus(hostStr, 150*time.Millisecond, signer) {
			online = false
		}

		targets = append(targets, Target{
			IPAddress:   tStr[0],
			Port:        tPort,
			Connections: 0,
			Online:      online,
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
