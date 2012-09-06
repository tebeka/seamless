/* A TCP proxy that allow you to deploy new code then switch traffic to it
   without downtime.

   It does "round robin" between the list of current active backends.

   Switching server is done with HTTP interface with the following API:
   /set?backends=host:port,host:port - will set list of backends
   /add?backend=host:port - will add a backend
   /remove?backend=host:port - will remove a backend
   /get - will return host:port,host:port

   Work flow:
	   Start first backend at port 4444
	   Run `./seamless 8080 localhost:4444`

	   Direct all traffic to port 8080 on local machine.

	   When you need to upgrade the backend, start a new one (with new code on
	   a different port, say 4445).
	   Then
			* `curl http://localhost:6777/add?backend=localhost:4445`
			* `curl http://localhost:6777/remove?backend=localhost:4444`
	   Or
		`curl http://localhost:6777/set?backends=localhost:4445`

	   New traffic will be directed to new server(s).

Original forward code by Roger Peppe (see http://bit.ly/Oc1YtF)
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

const (
	Version = "0.2.0"
)

// List of backends
var backends []string

// Current backend
var currentBackend int

// Sync backend changes
var backendsLock sync.RWMutex

var backendRe *regexp.Regexp = regexp.MustCompile("^[^:]+:[0-9]+$")

// isValidBackend returns true if backend is in "host:port" format.
func isValidBackend(backend string) bool {
	return backendRe.MatchString(backend)
}

// nextBackend returns the next backend to use.
// (Uses backendsLock.RLock)
func nextBackend() (string, error) {
	backendsLock.RLock()
	defer backendsLock.RUnlock()

	if len(backends) == 0 {
		return "", fmt.Errorf("No backends")
	}

	backend := backends[currentBackend]
	currentBackend = (currentBackend + 1) % len(backends)
	return backend, nil
}

func parseBackends(str string) ([]string, error) {
	backends := strings.Split(str, ",")
	if len(backends) == 0 {
		return nil, fmt.Errorf("no backends")
	}

	for i, v := range backends {
		backends[i] = strings.TrimSpace(v)
		if !isValidBackend(backends[i]) {
			return nil, fmt.Errorf("'%s' is not valid network address", backends[i])
		}
	}

	return backends, nil
}

// forward proxies traffic between local socket and remote backend
func forward(local net.Conn, remoteAddr string) {
	remote, err := net.Dial("tcp", remoteAddr)
	if remote == nil {
		log.Printf("remote dial failed: %v\n", err)
		local.Close()
		return
	}
	go io.Copy(local, remote)
	go io.Copy(remote, local)
}

// die prints error message and aborts the program
func die(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(1)
}

// startHttpServer start the HTTP server interface in a given port
func startHttpServer(port int) error {
	http.HandleFunc("/set", setHandler)
	http.HandleFunc("/get", getHandler)
	http.HandleFunc("/add", addHandler)
	http.HandleFunc("/remove", removeHandler)

	// Support pre 0.2.0 API
	http.HandleFunc("/current", getHandler)
	http.HandleFunc("/switch", switchHandler)

	return http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func setBackends(param string, w http.ResponseWriter, req *http.Request) {
	newBackends, err := parseBackends(req.FormValue(param))
	if err != nil {
		msg := fmt.Sprintf("error: %s", err)
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	backendsLock.Lock()
	defer backendsLock.Unlock()
	backends = newBackends
	getHandler(w, req)
}

// switchHandler handler /switch and switches backend
func switchHandler(w http.ResponseWriter, req *http.Request) {
	setBackends("backend", w, req)
}

// setHandler handler /switch and switches backend
func setHandler(w http.ResponseWriter, req *http.Request) {
	setBackends("backends", w, req)
}

// getHandler handles /current and return the current backend
func getHandler(w http.ResponseWriter, req *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	fmt.Fprintf(w, "%s\n", strings.Join(backends, ","))
}

func addHandler(w http.ResponseWriter, req *http.Request) {
	backend := req.FormValue("backend")
	if len(backend) == 0 {
		msg := "error: missing 'backend' parameter"
		log.Println(msg)
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	backendsLock.Lock()
	defer backendsLock.Unlock()
	backends = append(backends, backend)
	getHandler(w, req)
}

// remove removes all items matching 'item' from items.
func remove(items []string, item string) []string {
	i := 0
	for i < len(items) {
		if items[i] == item {
			items = append(items[:i], items[i+1:]...)
		} else {
			i++
		}
	}

	return items
}

func removeHandler(w http.ResponseWriter, req *http.Request) {
	err := ""

	defer func() {
		if len(err) != 0 {
			log.Println(err)
			http.Error(w, err, http.StatusBadRequest)
			return
		} else {
			getHandler(w, req)
		}
	}()

	backend := req.FormValue("backend")
	if len(backend) == 0 {
		err = "error: missing 'backend' parameter"
		return
	}

	backendsLock.Lock()
	defer backendsLock.Unlock()
	newBackends := remove(backends, backend)
	if len(newBackends) == len(backends) {
		err = fmt.Sprintf("error: backend '%s' not found", backend)
		return
	}

	backends = newBackends
}

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: seamless LISTEN_PORT BACKENDS\n")
		fmt.Fprintf(os.Stderr, "command line switches:\n")
		flag.PrintDefaults()
	}
	port := flag.Int("httpPort", 6777, "http interface port")
	version := flag.Bool("version", false, "show version and exit")
	flag.Parse()

	if *version {
		fmt.Printf("seamless %s\n", Version)
		os.Exit(0)
	}

	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}
	localAddr := fmt.Sprintf(":%s", flag.Arg(0))
	var err error
	backends, err = parseBackends(flag.Arg(1))
	if err != nil {
		die(fmt.Sprintf("%s", err))
	}

	local, err := net.Listen("tcp", localAddr)
	if local == nil {
		die("cannot listen: %v", err)
	}

	go func() {
		if err := startHttpServer(*port); err != nil {
			die("cannot listen on %d: %v", *port, err)
		}
	}()

	for {
		conn, err := local.Accept()
		if conn == nil {
			die("accept failed: %v", err)
		}
		backend, err := nextBackend()
		if err != nil {
			log.Printf("error: can't get next backend %v\n", err)
			conn.Close()
		}
		go forward(conn, backend)
	}
}
