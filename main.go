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

var kong string

func addACL(api Api) error {
	data := url.Values{}
	data.Add("name", "acl")
	data.Add("config.whitelist", api.Groups)

	kong_url := fmt.Sprintf("%v/apis/%v/plugins", kong, api.Name)

	log.Println("Adding acl plugin to api:", api.Name)
	return sendRequest("POST", kong_url, data)
}

func addKeyAuth(api Api) error {

	data := url.Values{}
	data.Add("name", "key-auth")
	data.Add("config.key_names", "X-apikey")

	kong_url := fmt.Sprintf("%v/apis/%v/plugins", kong, api.Name)

	log.Println("Adding key-auth plugin to api:", api.Name)
	return sendRequest("POST", kong_url, data)
}

func deleteApi(api Api) error {
	kong_url := fmt.Sprintf("%v/apis/%v", kong, api.Name)

	log.Println("Deleting api:", api.Name)
	return sendRequest("DELETE", kong_url, nil)
}

func registerApi(api Api) error {
	data := url.Values{}
	data.Add("upstream_url", api.URL)
	data.Add("name", api.Name)
	data.Add("request_path", api.Path)
	data.Add("strip_request_path", "true")

	kong_url := fmt.Sprintf("%v/apis", kong)

	log.Println("Registering API")
	return sendRequest("POST", kong_url, data)
}

func sendRequest(method string, kong_url string, data url.Values) error {
	req, err := http.NewRequest(method, kong_url, strings.NewReader(data.Encode()))
	if err != nil {
		log.Println(err)
		return err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := http.DefaultClient.Do(req)
	defer resp.Body.Close()
	if err != nil {
		log.Println(err)
		return err
	}

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

	deleteApi(api) // Clear up an existing endpoint with the same name
	registerApi(api)
	addKeyAuth(api)
	addACL(api)

	if err != nil {
		rw.Write([]byte(fmt.Sprintf("Failed to register the API with error %v", err)))
	} else {
		log.Println(api.Name, "added")
		rw.Write([]byte(fmt.Sprintf("Registered API %v", api.Name)))
	}
}

func main() {
	kong = os.Getenv("KONG_URL")
	_, err := url.ParseRequestURI(kong)
	if err != nil {
		panic("KONG_URL not set or is invalid!")
	}

	http.HandleFunc("/add", receiveApiDetails)
	http.HandleFunc("/remove", deleteOldApis)
	log.Fatal(http.ListenAndServe(":8080", nil))
}
