package hue

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

const meethueURL = "https://www.meethue.com"

type BridgeInfo struct {
	ID string
	IP string `json:"internalipaddress"`
}

type Client struct {
	rootURL  string
	Hostname string
	Username string

	client *http.Client
}

func New(hostname string, username string) *Client {
	rootURL := fmt.Sprintf("http://%v/api/%v", hostname, username)
	return &Client{
		client:   &http.Client{},
		rootURL:  rootURL,
		Username: username,
		Hostname: hostname,
	}
}

func (c *Client) get(path string, resObject interface{}) error {
	return c.do("GET", path, nil, resObject)
}

func (c *Client) post(path string, reqObject interface{}, resObject interface{}) error {
	return c.do("POST", path, reqObject, resObject)
}

func (c *Client) put(path string, reqObject interface{}, resObject interface{}) error {
	return c.do("PUT", path, reqObject, resObject)
}

func (c *Client) delete(path string, reqObject interface{}, resObject interface{}) error {
	return c.do("DELETE", path, reqObject, resObject)
}

func (c *Client) do(method string, path string, reqObject interface{}, resObject interface{}) error {
	return do(c.client, method, c.url(path), reqObject, resObject)
}

func do(client *http.Client, method string, url string, reqObject interface{}, resObject interface{}) error {
	error := func(msg string, err error) error {
		return fmt.Errorf("hue.Client %v %v: %v: %v", method, url, msg, err)
	}
	var reqBody io.Reader = nil
	if reqObject != nil {
		reqBytes, err := json.Marshal(reqObject)
		if err != nil {
			return error("encoding", err)
		}
		reqBody = bytes.NewReader(reqBytes)
	}
	log.Printf("do %v %v\n", method, url)

	req, err := http.NewRequest(method, url, reqBody)
	if err != nil {
		return error("request", err)
	}
	req.Header.Add("content-type", "application/json")
	res, err := client.Do(req)
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return error("read", err)
	}

	if res.StatusCode >= 300 {
		return error("read", fmt.Errorf("unexpected status code %v. Body: %v", res.StatusCode, string(body)))
	}

	if resObject != nil {
		err = json.Unmarshal(body, resObject)
		if err != nil {
			return error("decoding", fmt.Errorf("%v: %v", string(body), err))
		}
	}

	return nil
}

func (c *Client) url(path string) string {
	return fmt.Sprintf("%v%v", c.rootURL, path)
}
