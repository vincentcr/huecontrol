package services

import (
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"regexp"

	"github.com/PuerkitoBio/goquery"
)

const BASE_URL = "https://www.meethue.com"

type HueTokenCredentials struct {
	DeviceID string
	Username string
	Password string
}

func GetToken(creds HueTokenCredentials) (string, error) {
	return GetTokenAtBaseURL(BASE_URL, creds)
}

func GetTokenAtBaseURL(baseURL string, creds HueTokenCredentials) (string, error) {
	hueClient, err := newHueClient(baseURL)
	if err != nil {
		return "", err
	}

	if err := startTokenSession(hueClient, creds.DeviceID); err != nil {
		return "", err
	}

	tokenURL, err := grantAccess(hueClient, creds.Username, creds.Password)
	if err != nil {
		return "", err
	}

	return fetchToken(hueClient, tokenURL)
}

func SendMessage(token string) {
}

type HueLight struct {
	Name    string
	UID     string `json:"uniqueid"`
	ID      string
	Type    string
	ModelID string `json:"modelid"`
	State   struct {
		On         bool
		Brightness int `json:"bri"`
		Alert      string
		Reachable  bool
	}
}

func GetStatus(token string) {

}

// func getBridgeID() (string, error) {
// 	url := BASE_URL + "/api/nupnp"
// 	resp, err := httpClient.Get(url)
// 	if err != nil {
// 		return "", fmt.Errorf("getBridgeID: request to %v failed: %v", url, err)
// 	}
// 	defer resp.Body.Close()
// 	body, err := ioutil.ReadAll(resp.Body)
// 	if err != nil {
// 		return "", fmt.Errorf("getBridgeID: reading response body failed", err)
// 	}
//
// 	var bridges []struct{ ID string }
// 	err = json.Unmarshal(body, &bridges)
// 	if err != nil {
// 		return "", fmt.Errorf("getBridgeID: parsing response body failed", err)
// 	}
//
// 	return bridges[0].ID, nil
// }

func startTokenSession(client *hueClient, bridgeID string) error {
	resp, err := client.get("/en-us/api/gettoken?devicename=iPhone+5&appid=hueapp&deviceid=" + bridgeID)
	if err != nil {
		return fmt.Errorf("starTokenSession failed: %v", err)
	}
	defer resp.Body.Close()

	return nil
}

func grantAccess(client *hueClient, email, password string) (string, error) {
	form := url.Values{"email": {email}, "password": {password}}
	resp, err := client.postForm("/en-us/api/getaccesstokengivepermission", form)
	if err != nil {
		return "", fmt.Errorf("grantAccess: error posting token access form: %v", err)
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

func fetchToken(hueClient *hueClient, tokenURL string) (string, error) {
	resp, err := hueClient.get(tokenURL)
	if err != nil {
		return "", fmt.Errorf("fetchToken: error getting url %v: %v", tokenURL, err)
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

type hueClient struct {
	baseURL string
	http    http.Client
}

func newHueClient(baseURL string) (*hueClient, error) {
	c := hueClient{}
	c.http = http.Client{}
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("newHueClient: failed to create cookie jar: %v", err)
	}
	c.http.Jar = jar
	return &c, nil
}

func (c *hueClient) get(path string) (*http.Response, error) {
	return c.http.Get(fmt.Sprintf("%s/%s", c.baseURL, path))
}

func (c *hueClient) post(path string, bodyType string, body io.Reader) (*http.Response, error) {
	return c.http.Post(fmt.Sprintf("%s/%s", c.baseURL, path), bodyType, body)
}

func (c *hueClient) postForm(path string, form url.Values) (*http.Response, error) {
	return c.http.PostForm(fmt.Sprintf("%s/%s", c.baseURL, path), form)
}
