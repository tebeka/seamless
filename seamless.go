/* A TCP proxy that allow you to deploy new code then switch traffic to it
   without downtime.

   Switching server is done with HTTP interface with the following API:
   /switch?backend=address - will switch traffic to new backend
   /current - will return (in plain text) current server

   Work flow:
	   Start first backend at port 4444
	   Run `./seamless 8080 localhost:4444`

	   Direct all traffic to port 8080 on local machine.

	   When you need to upgrade the backend, start a new one (with new code on
	   a different port, say 4445).
	   The `curl http://localhost:6777/switch?backend=localhost:4445. 
	   New traffic will be directed to new server.

Original forward code by Roger Peppe (see http://bit.ly/Oc1YtF)
*/
package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
)

var backend string

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: seamless LISTEN_ADDR BACKEND\n")
		fmt.Fprintf(os.Stderr, "command line switches:\n")
		flag.PrintDefaults()
	}
	port := flag.Int("httpPort", 6777, "http interface port")
	flag.Parse()
	if flag.NArg() != 2 {
		flag.Usage()
		os.Exit(1)
	}
	localAddr := flag.Arg(0)
	backend = flag.Arg(1)

	local, err := net.Listen("tcp", localAddr)
	if local == nil {
		die(fmt.Sprintf("cannot listen: %v", err))
	}

	go runHttpServer(*port)

	for {
		conn, err := local.Accept()
		if conn == nil {
			die(fmt.Sprintf("accept failed: %v", err))
		}
		go forward(conn, backend)
	}
}

func forward(local net.Conn, remoteAddr string) {
	remote, err := net.Dial("tcp", remoteAddr)
	if remote == nil {
		fmt.Fprintf(os.Stderr, "remote dial failed: %v\n", err)
		return
	}
	go io.Copy(local, remote)
	go io.Copy(remote, local)
}

func die(msg string) {
	fmt.Fprintf(os.Stderr, "error: %s\n", msg)
	os.Exit(1)
}

func runHttpServer(port int) {
	http.HandleFunc("/switch", switchHandler)
	http.HandleFunc("/current", currentHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func switchHandler(w http.ResponseWriter, req *http.Request) {
	newBackend := req.FormValue("backend")
	if len(newBackend) == 0 {
		http.Error(w, "missing 'backend' parameter", http.StatusBadRequest)
		return
	}
	backend = newBackend
	currentHandler(w, req)
}

func currentHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%s\n", backend)
}
