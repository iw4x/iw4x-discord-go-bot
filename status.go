package main

import (
	"net/http"
	"encoding/json"
	"io/ioutil"
	"log"
	"strconv"
)

func fetch_players() (string) {
	type Server struct {
		Client int `json:"clients"` // each `servers` entry contains a `clients` variable, pull that
	}

	type Response struct {
		Servers []Server `json:"servers"` // we're looking through entries in `servers`
	}

	r, err := http.Get("https://master." + base_url + "v1/servers/iw4x?protocol=152")
	if err != nil {
		log.Print(err)
	}
	defer r.Body.Close()

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Print(err)
	}

	var response Response
	json.Unmarshal(body, &response)

	var result int = 0
	for _, p := range response.Servers {
		result += p.Client // for every entry, sum with current value of result
	}

	// this needs to be a string when used for status, convert from int
	result_output := strconv.Itoa(result)

	return result_output
}
