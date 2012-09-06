// FIXME: Write a full test that starts a backend and seamless, and then
// switches
package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

var port int = 6777

func init() {
	backends = []string{"localhost:8888"}
	go startHttpServer(port)
	time.Sleep(1 * time.Second)
}

func callAPI(suffix string) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/%s", port, suffix)
	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error connecting to /current: %v\n", err)
	}
	defer resp.Body.Close()

	reply, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading reply: %v\n", err)
	}

	return string(reply), nil
}

func getTest(suffix string, t *testing.T) {
	reply, err := callAPI(suffix)
	if err != nil {
		t.Fatalf("%s", err)
	}

	if reply != fmt.Sprintf("%s\n", backends[0]) {
		t.Fatalf("bad reply: %s\n", string(reply))
	}
}

func TestHttpOldAPI(t *testing.T) {
	getTest("current", t)
}

func TestHTTPGet(t *testing.T) {
	getTest("get", t)
}

func TestAdd(t *testing.T) {
	backends = []string{"localhost:8888"}
	backend := "localhost:8887"

	reply, err := callAPI(fmt.Sprintf("add?backend=%s", backend))
	if err != nil {
		t.Fatalf("%s", err)
	}

	if len(backends) != 2 {
		t.Fatalf("bad number of backends (%d)\nreply: %s", len(backends), reply)
	}

	if backends[1] != backend {
		t.Fatalf("bad backend - %s", backends[0])
	}

	if reply != fmt.Sprintf("%s,%s\n", backends[0], backends[1]) {
		t.Fatalf("bad reply - %s\n", reply)
	}
}
