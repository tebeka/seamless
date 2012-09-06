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

func TestHTTPAdd(t *testing.T) {
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

func TestHTTPRemove(t *testing.T) {
	backend1, backend2 := "localhost:8888", "localhost:8887"
	backends = []string{backend1, backend2}
	reply, err := callAPI(fmt.Sprintf("remove?backend=%s", backend1))
	if err != nil {
		t.Fatalf("%s", err)
	}

	if len(backends) != 1 {
		t.Fatalf("bad number of backends (%d)\nreply: %s", len(backends), reply)
	}

	if backends[0] != backend2 {
		t.Fatalf("bad backend left - %s", backends[0])
	}
}

func Test_isValidBackend(t *testing.T) {
	names := map[bool]string{
		true:  "valid",
		false: "invalid",
	}

	cases := []struct {
		value string
		valid bool
	}{
		{"localhost:7", true},
		{"foo.com:8080", true},
		{"", false},
		{"foo.com", false},
		{"localhost", false},
		{"foo.com:", false},
	}

	for _, c := range cases {
		if isValidBackend(c.value) != c.valid {
			t.Fatalf("`%s` should be %s", c.value, names[c.valid])
		}
	}
}

func Test_nextBackend(t *testing.T) {
	backend1, backend2 := "localhost:8888", "localhost:8887"
	backends = []string{backend1, backend2}
	currentBackend = 0

	for i, expected := range []string{backend1, backend2, backend1} {
		next, _ := nextBackend()
		if next != expected {
			t.Fatalf("backend should be %s at %d (was %s)", expected, i, next)
		}
	}

	backends = []string{backend1, backend2}
	_, err := nextBackend()
	if err != nil {
		t.Fatalf("managed to get backend from empty list")
	}
}

func arreq(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if b[i] != v {
			return false
		}
	}

	return true
}

func Test_parseBackends(t *testing.T) {
	cases := []struct {
		backends string
		expected []string
		ok       bool
	}{
		{"localhost:8080", []string{"localhost:8080"}, true},
		{"localhost:8080,localhost:8887", []string{"localhost:8080", "localhost:8887"}, true},
		{"", []string{}, false},
		{"foo", []string{}, false},
		{"localhost:8080,localhost", []string{}, false},
	}

	for _, c := range cases {
		value, err := parseBackends(c.backends)
		ok := err == nil

		if ok != c.ok {
			t.Fatalf("bad error for %v", c.backends)
		}

		if !arreq(value, c.expected) {
			t.Fatalf("go %v for %s (expected %v)", value, c.backends, c.expected)
		}


	}
}
