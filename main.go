package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

type Api struct {
	URL    string `json:"url"`
	Name   string `json:"name"`
	Path   string `json:"path"`
	Groups string `json:"groups"`
}

var kong = os.Getenv("KONG_URL")

func addACL(api Api) error {
	log.Println("Adding acl plugin to api:", api.Name)

	data := url.Values{}
	data.Add("name", "acl")
	data.Add("config.whitelist", api.Groups)

	kong_url := fmt.Sprintf("%v/apis/%v/plugins", kong, api.Name)

	req, err := http.NewRequest("POST", kong_url, strings.NewReader(data.Encode()))
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

func addKeyAuth(api Api) error {
	log.Println("Adding key-auth plugin to api:", api.Name)

	data := url.Values{}
	data.Add("name", "key-auth")
	data.Add("config.key_names", "X-apikey")

	kong_url := fmt.Sprintf("%v/apis/%v/plugins", kong, api.Name)

	req, err := http.NewRequest("POST", kong_url, strings.NewReader(data.Encode()))
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	return nil
}

func deleteApi(api Api) error {
	log.Println("Deleting api:", api.Name)

	kong_url := fmt.Sprintf("%v/apis/%v", kong, api.Name)

	req, err := http.NewRequest("DELETE", kong_url, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

func registerApi(api Api) error {
	data := url.Values{}
	data.Add("upstream_url", api.URL)
	data.Add("name", api.Name)
	data.Add("request_path", api.Path)
	data.Add("strip_request_path", "true")

	kong_url := fmt.Sprintf("%v/apis", kong)

	req, err := http.NewRequest("POST", kong_url, strings.NewReader(data.Encode()))
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return err
	}
	defer resp.Body.Close()

	return nil
}

func deleteOldApis(rw http.ResponseWriter, req *http.Request) {
	log.Println("This doesn't do anything, for now")
}

func receiveApiDetails(rw http.ResponseWriter, req *http.Request) {
	decoder := json.NewDecoder(req.Body)
	var api Api

	err := decoder.Decode(&api)
	if err != nil {
		log.Println(err)
		rw.Write([]byte("Could not decode request"))
	}

	defer req.Body.Close()
	api.Name = fmt.Sprintf("autokong-%v", api.Name)

	log.Println(api.Name)
	log.Println(api.URL)
	log.Println(api.Path)
	log.Println(api.Groups)

	err = deleteApi(api) // Clear up an existing endpoint with the same name
	log.Println(err)
	err = registerApi(api)
	err = addKeyAuth(api)
	err = addACL(api)

	if err != nil {
		rw.Write([]byte(fmt.Sprintf("Failed to register the API with error %v", err)))
	} else {
		log.Println(api.Name, "added")
		rw.Write([]byte(fmt.Sprintf("Registered API %v", api.Name)))
	}
}

func main() {
	http.HandleFunc("/add", receiveApiDetails)
	http.HandleFunc("/remove", deleteOldApis)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
