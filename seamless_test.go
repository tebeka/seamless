package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"testing"
	"time"
)

func TestHttp(t *testing.T) {
	backend = "hello"
	port := 6777
	go startHttpServer(port)

	time.Sleep(1 * time.Second)

	resp, err := http.Get(fmt.Sprintf("http://localhost:%d/current", port))
	if err != nil {
		t.Fatalf("error connecting to /current: %v\n", err)
	}
	defer resp.Body.Close()

	reply, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatalf("error reading reply: %v\n", err)
	}

	if string(reply) != fmt.Sprintf("%s\n", backend) {
		t.Fatalf("bad reply: %s\n", string(reply))
	}
}
