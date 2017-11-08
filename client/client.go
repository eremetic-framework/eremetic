package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/cybricio/eremetic"
	"github.com/cybricio/eremetic/api"
)

// Client is used for communicating with an Eremetic server.
type Client struct {
	httpClient *http.Client
	endpoint   string
}

// New returns a new instance of a Client.
func New(endpoint string, client *http.Client) (*Client, error) {
	return &Client{
		httpClient: client,
		endpoint:   endpoint,
	}, nil
}

// AddTask sends a request for a new task to be scheduled.
func (c *Client) AddTask(r api.RequestV1) error {
	var buf bytes.Buffer

	err := json.NewEncoder(&buf).Encode(r)
	if err != nil {
		return err
	}

	req, err := http.NewRequest("POST", c.endpoint+"/api/v1/task", &buf)
	if err != nil {
		return err
	}

	_, err = c.httpClient.Do(req)
	if err != nil {
		return err
	}

	return nil
}

// Task returns a task with a given ID.
func (c *Client) Task(id string) (*eremetic.Task, error) {
	req, err := http.NewRequest("GET", c.endpoint+"/api/v1/task/"+id, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	var task api.TaskV1

	err = json.NewDecoder(resp.Body).Decode(&task)
	if err != nil {
		return nil, err
	}

	t := api.TaskFromV1(&task)

	return &t, nil
}

// Tasks returns all current tasks.
func (c *Client) Tasks() ([]eremetic.Task, error) {
	req, err := http.NewRequest("GET", c.endpoint+"/api/v1/task", nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	var tasks []api.TaskV1

	err = json.NewDecoder(resp.Body).Decode(&tasks)
	if err != nil {
		return nil, err
	}

	taskSlice := []eremetic.Task{}
	for _, t := range tasks {
		taskSlice = append(taskSlice, api.TaskFromV1(&t))
	}

	return taskSlice, nil
}

// Sandbox returns a sandbox resource for a given task.
func (c *Client) Sandbox(taskID, file string) ([]byte, error) {
	u := fmt.Sprintf("%s/api/v1/task/%s/%s", c.endpoint, taskID, file)
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

// Version returns the version of the Eremetic server.
func (c *Client) Version() (string, error) {
	u := fmt.Sprintf("%s/api/v1/version", c.endpoint)
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

// Kill a running task
func (c *Client) Kill(taskID string) error {
	u := fmt.Sprintf("%s/api/v1/task/%s/kill", c.endpoint, taskID)
	req, err := http.NewRequest("POST", u, nil)
	if err != nil {
		return err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusAccepted {
		return fmt.Errorf("Unexpected status code `%s`", resp.Status)
	}

	return nil
}
