package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"time"
)

const (
	MetabaseSessionIdHeader = "X-Metabase-Session"
)

type Client struct {
	BaseURL    string
	HttpClient *http.Client

	Auth      AuthDetails
	SessionId string
}

type AuthDetails struct {
	Username string `json:"username"`
	Password string `json:"password"`
}
type AuthResponse struct {
	SessionId string `json:"id"`
}
type SuccessResponse struct {
	Success bool `json:"success"`
}

func NewClient(host string, username string, password string) (*Client, error) {
	if host == "" {
		return nil, fmt.Errorf("must provide a valid host URL")
	}
	if username == "" {
		return nil, fmt.Errorf("must provide a valid username")
	}
	if password == "" {
		return nil, fmt.Errorf("must provide a valid password")
	}

	c := &Client{
		BaseURL: host,
		HttpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		Auth: AuthDetails{
			Username: username,
			Password: password,
		},
	}

	err := c.signIn()
	if err != nil {
		return nil, err
	}

	return c, nil
}

func (c *Client) doRequest(req *http.Request, response interface{}) error {
	req.Header.Set("Content-Type", "application/json")

	if c.SessionId != "" {
		req.Header.Set(MetabaseSessionIdHeader, c.SessionId)
	}

	res, err := c.HttpClient.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode < 200 || res.StatusCode >= 300 {
		return errors.New(string(body))
	}

	err = json.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) doGet(path string, response interface{}) error {
	req, err := http.NewRequest("GET", c.makeUrl(path), nil)
	if err != nil {
		return err
	}

	err = c.doRequest(req, &response)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) doPost(path string, request interface{}, response interface{}) error {
	reqBody, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.makeUrl(path), bytes.NewBuffer(reqBody))
	if err != nil {
		return err
	}

	err = c.doRequest(req, &response)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) doPut(path string, request interface{}, response interface{}) error {
	var bodyBuffer bytes.Buffer
	if request != nil {
		reqBody, err := json.Marshal(request)
		if err != nil {
			return err
		}
		bodyBuffer = *bytes.NewBuffer(reqBody)
	}

	req, err := http.NewRequest("PUT", c.makeUrl(path), &bodyBuffer)
	if err != nil {
		return err
	}

	err = c.doRequest(req, &response)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) doDelete(path string, response interface{}) error {
	req, err := http.NewRequest("DELETE", c.makeUrl(path), nil)
	if err != nil {
		return err
	}

	err = c.doRequest(req, &response)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) makeUrl(url string) string {
	return fmt.Sprintf("%s/api%s", c.BaseURL, url)
}

func (c *Client) signIn() error {
	// Reset the session ID, so we log in without the header
	c.SessionId = ""

	var authResponse AuthResponse
	err := c.doPost("/session", c.Auth, &authResponse)
	if err != nil {
		return err
	}

	c.SessionId = authResponse.SessionId
	return nil
}
