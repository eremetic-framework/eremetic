package client

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rockerbox/eremetic/api"
	"github.com/rockerbox/eremetic/version"
)

func TestClient_AddTask(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

	}))
	defer ts.Close()

	var httpClient http.Client

	c, err := New(ts.URL, &httpClient)
	if err != nil {
		t.Fatal(err)
	}

	var req api.RequestV1

	if err := c.AddTask(req); err != nil {
		t.Fatal(err)
	}
}

func TestClient_Tasks(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`[{
			"id": "eremetic-id-12345"
		}]`))
	}))
	defer ts.Close()

	var httpClient http.Client

	c, err := New(ts.URL, &httpClient)
	if err != nil {
		t.Fatal(err)
	}

	tasks, err := c.Tasks()
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 1 {
		t.Fail()
	}
}

func TestClient_KillTask(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	var httpClient http.Client

	c, err := New(ts.URL, &httpClient)
	if err != nil {
		t.Fatal(err)
	}

	if err := c.Kill("1234"); err != nil {
		t.Fatal(err)
	}
}

func TestClient_ReadTask(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"mem":22.0, "image": "busybox", "command": "echo hello", "cpu":0.5}`))
	}))
	defer ts.Close()

	var httpClient http.Client
	c, err := New(ts.URL, &httpClient)
	if err != nil {
		t.Fatal(err)
	}
	task, err := c.Task("1234")
	if err != nil {
		t.Fatal(err)
	}

	if task.TaskMem != 22 || task.Image != "busybox" || task.Command != "echo hello" || task.TaskCPUs != 0.5 {
		t.Fatal(errors.New("Unexpected task"))
	}
}
func TestClient_Version(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(version.Version))
	}))
	defer ts.Close()

	var httpClient http.Client
	c, err := New(ts.URL, &httpClient)
	if err != nil {
		t.Fatal(err)
	}
	version, err := c.Version()
	if err != nil {
		t.Fatal(err)
	}

	if len(version) == 0 {
		t.Fatal("Missing version")
	}
}

func TestClient_Sandbox(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("remember remember the 5th of november\nthe gunpowder treason and plot.\nI see no reason the gunpowder treason should ever be forgot.\n"))
	}))
	defer ts.Close()

	var httpClient http.Client
	c, err := New(ts.URL, &httpClient)
	if err != nil {
		t.Fatal(err)
	}
	poem, err := c.Sandbox("1234", "poem")
	if err != nil {
		t.Fatal(err)
	}

	if len(poem) == 0 {
		t.Fatal("Failed to get file")
	}
}
