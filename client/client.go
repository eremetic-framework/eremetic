package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/klarna/eremetic"
)

type Client struct {
	httpClient *http.Client
	endpoint   string
}

func New(endpoint string, client *http.Client) (*Client, error) {
	return &Client{
		httpClient: client,
		endpoint:   endpoint,
	}, nil
}

func (c *Client) AddTask(r eremetic.Request) error {
	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(r)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.endpoint+"/task", &buf)
	if err != nil {
		return err
	}

	_, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) Task(id string) (*eremetic.Task, error) {
	req, err := http.NewRequest("GET", c.endpoint+"/task/"+id, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	var task eremetic.Task

	err = json.NewDecoder(resp.Body).Decode(&task)
	if err != nil {
		return nil, err
	}

	return &task, nil
}

func (c *Client) Tasks() ([]eremetic.Task, error) {
	req, err := http.NewRequest("GET", c.endpoint+"/task", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	var tasks []eremetic.Task

	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		return nil, err
	}

	return tasks, nil
}

func (c *Client) Sandbox(taskID, file string) ([]byte, error) {
	u := fmt.Sprintf("%s/task/%s/%s", c.endpoint, taskID, file)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (c *Client) Version() (string, error) {
	u := fmt.Sprintf("%s/version", c.endpoint)
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return "", err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	return string(b), nil
}
