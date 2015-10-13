package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/howeyc/gopass"
)

const BASE_URL = "https://www.meethue.com"

var httpClient http.Client

func init() {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err.Error())
	}
	httpClient = http.Client{Jar: jar}

}

func main() {
	username, password := getCredentials()
	token, err := getToken(username, password)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	} else {
		fmt.Println("Token:", token)
	}
}

func getCredentials() (string, string) {
	var password string
	var username string
	flag.StringVar(&password, "password", "", "password")
	flag.StringVar(&username, "username", "", "username")
	flag.Parse()

	reader := bufio.NewReader(os.Stdin)
	if username == "" {
		fmt.Print("Username: ")
		line, _ := reader.ReadString('\n')
		username = strings.TrimSpace(line)
	}
	if password == "" {
		fmt.Print("Password: ")
		password = string(gopass.GetPasswdMasked())
	}

	return username, password
}

func getToken(username, password string) (string, error) {
	bridgeID, err := getBridgeID()
	if err != nil {
		return "", err
	}
	if err := startTokenSession(bridgeID); err != nil {
		return "", err
	}

	tokenURL, err := grantAccess(username, password)
	if err != nil {
		return "", err
	}

	return fetchToken(tokenURL)
}

func getBridgeID() (string, error) {
	url := BASE_URL + "/api/nupnp"
	resp, err := httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("getBridgeID: request to %v failed: %v", url, err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("getBridgeID: reading response body failed", err)
	}

	var bridges []struct{ ID string }
	err = json.Unmarshal(body, &bridges)
	if err != nil {
		return "", fmt.Errorf("getBridgeID: parsing response body failed", err)
	}

	return bridges[0].ID, nil
}

func startTokenSession(bridgeID string) error {
	url := BASE_URL + "/en-us/api/gettoken?devicename=iPhone+5&appid=hueapp&deviceid=" + bridgeID
	resp, err := httpClient.Get(url)
	if err != nil {
		return fmt.Errorf("starTokenSession failed: %v", err)
	}
	defer resp.Body.Close()

	return nil
}

func grantAccess(email, password string) (string, error) {
	grantAccessURL := BASE_URL + "/en-us/api/getaccesstokengivepermission"
	form := url.Values{"email": {email}, "password": {password}}
	resp, err := httpClient.PostForm(grantAccessURL, form)
	if err != nil {
		return "", fmt.Errorf("grantAccess: error posting form to %v", grantAccessURL, err)
	}
	defer resp.Body.Close()

	gq, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("grantAccess: Unable to parse response", err)
	}

	tokenURL, found := gq.Find("[data-role='yes']").Attr("href")
	if !found {
		return "", fmt.Errorf("Invalid credentials")
	} else {
		return tokenURL, nil
	}

}

func fetchToken(tokenURL string) (string, error) {
	fullTokenURL := BASE_URL + tokenURL
	resp, err := httpClient.Get(fullTokenURL)
	if err != nil {
		return "", fmt.Errorf("fetchToken: error getting url %v: %v", fullTokenURL, err)
	}
	defer resp.Body.Close()

	gq, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		return "", fmt.Errorf("fetchToken: Unable to parse response", err)
	}
	successURL, found := gq.Find("a.button-primary").Attr("href")
	if !found {
		return "", fmt.Errorf("Unable to find token link in body")
	}

	tokenExtractRe := regexp.MustCompile("^phhueapp://sdk/login/(.+)$")
	matches := tokenExtractRe.FindStringSubmatch(successURL)
	if len(matches) == 0 {
		return "", fmt.Errorf("Unable to parse token from link")
	}

	return matches[1], nil
}
