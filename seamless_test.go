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

var apiPort int = 6777
var numBackends int = 3
var proxyPort = 6888

type testHandler int

func (h testHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "%d", h)
}

func startBackend(i int) {
	handler := testHandler(i)
	port := 6700 + i
	server := http.Server{Handler: handler, Addr: fmt.Sprintf(":%d", port)}
	go server.ListenAndServe()
}

func init() {
	for i := 0; i < numBackends; i++ {
		startBackend(i)
	}

	out := make(chan error)
	go seamless(fmt.Sprintf(":%d", proxyPort), apiPort, backends, out)

	time.Sleep(1 * time.Second)
}

func call(url string) (string, error) {
	// We really don't want keep alive or caching :)
	client := &http.Client{Transport: &http.Transport{DisableKeepAlives: true}}
	resp, err := client.Get(url)
	if err != nil {
		return "", fmt.Errorf("can't GET %s: %v\n", url, err)
	}
	defer resp.Body.Close()

	reply, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading reply: %v\n", err)
	}

	return string(reply), nil
}

func callAPI(suffix string) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/%s", apiPort, suffix)
	return call(url)
}

func TestHTTPGet(t *testing.T) {
	setBackends([]string{"localhost:8080"})
	reply, err := callAPI("get")
	if err != nil {
		t.Fatalf("%s", err)
	}

	if reply != fmt.Sprintf("%s\n", backends[0]) {
		t.Fatalf("bad reply: %s\n", string(reply))
	}
}

func TestHTTPAdd(t *testing.T) {
	setBackends([]string{"localhost:8888"})
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
	setBackends([]string{backend1, backend2})
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
	setBackends([]string{backend1, backend2})

	for i, expected := range []string{backend2, backend1, backend2} {
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

func backendAddr(i int) string {
	return fmt.Sprintf("localhost:%d", 6700+i)
}

func callProxy() (string, error) {
	url := fmt.Sprintf("http://localhost:%d", proxyPort)
	return call(url)
}

func TestProxy(t *testing.T) {
	setBackends([]string{backendAddr(0), backendAddr(1)})

	for i := 0; i < 7; i++ {
		reply, err := callProxy()
		if err != nil {
			t.Fatalf("can't call proxy - %v", err)
		}
		expected := fmt.Sprintf("%d", (i+1)%len(backends))
		if reply != expected {
			t.Fatalf("bad backend for i=%d: got %s instead of %s", i, reply, expected)
		}
	}
}

func TestProxyRemove(t *testing.T) {
	setBackends([]string{backendAddr(0), backendAddr(1)})
	suffix := fmt.Sprintf("remove?backend=%s", backendAddr(0))
	if _, err := callAPI(suffix); err != nil {
		t.Fatalf("can't remove %s - %s", backendAddr(0), err)
	}

	for i := 0; i < 7; i++ {
		reply, err := callProxy()
		if err != nil {
			t.Fatalf("can't call proxy - %v", err)
		}
		if reply != "1" {
			t.Fatalf("bad reply %s (expected 1)", reply)
		}
	}
}

func TestProxyAdd(t *testing.T) {
	setBackends([]string{backendAddr(0), backendAddr(1)})

	suffix := fmt.Sprintf("add?backend=%s", backendAddr(2))
	if _, err := callAPI(suffix); err != nil {
		t.Fatalf("can't remove %s - %s", backendAddr(0), err)
	}

	for i := 0; i < 7; i++ {
		reply, err := callProxy()
		if err != nil {
			t.Fatalf("can't call proxy - %v", err)
		}
		expected := fmt.Sprintf("%d", (i+1)%len(backends))
		if reply != expected {
			t.Fatalf("bad reply %s (expected %s)", reply, expected)
		}
	}
}
