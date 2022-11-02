package main

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"

	"gopkg.in/alecthomas/kingpin.v2"
)

var (
	version string
	commit  string
	date    string
	builtBy string
)

type release struct {
	TagName string `json:"tag_name"`
}

func makeRequest(url string) (body []byte, err error) {
	req, _ := http.NewRequest("GET", url, nil)
	var client http.Client      // setup client
	resp, err := client.Do(req) // make request
	if err != nil {
		return body, err
	}

	if resp.StatusCode != http.StatusOK {
		return body, errors.New(strconv.Itoa(resp.StatusCode) + ": " + resp.Status)
	}

	body, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return body, err
	}

	return body, err
}

func main() {
	app := kingpin.New("hosts-updater", "downloads latest stevenblack/hosts release").Author("intrand")
	output := app.Flag("output", "Path to write hosts file").Envar("hosts_updater_output").Default("hosts").String()
	kingpin.MustParse(app.Parse(os.Args[1:]))

	latestReleaseUrl := "https://api.github.com/repos/StevenBlack/hosts/releases/latest" // url to latest github release

	body, err := makeRequest(latestReleaseUrl) // get the json payload from API
	if err != nil {
		log.Fatalln(err)
	}

	latestRelease := release{}                             // make empty release
	json.Unmarshal(body, &latestRelease)                   // put API release into our empty release
	log.Println("Found release: " + latestRelease.TagName) // notify for admin use

	var latestHostsUrl string = "https://raw.githubusercontent.com/StevenBlack/hosts/" + latestRelease.TagName + "/hosts" // create URL for raw file
	body, err = makeRequest(latestHostsUrl)                                                                               // download file
	if err != nil {
		log.Fatalln(err)
	}
	if len(body) < 1 { // just in case :)
		log.Fatalln("something bad happened somewhere")
	}

	body = []byte(strings.ReplaceAll(string(body), "0.0.0.0", "127.127.127.127")) // 0.0.0.0 is valid; I prefer loopback
	err = ioutil.WriteFile(*output, body, 0o644)                                  // write body to file (overwriting)
	if err != nil {
		log.Fatalln(err)
	}
}
