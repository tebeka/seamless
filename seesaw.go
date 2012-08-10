/* Proxy TCP between two servers. Allow you to deploy new code to one server
   then switch traffic to it without downtime.

   Switching server is done with HTTP interface with the following API:
   /toggle - will toggle between servers
   /current - will return (in plain text) current server
   /servers - will return active, backup servers

Original code by Roger Peppe at http://bit.ly/Oc1YtF
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

var remotes = []string{"", ""}
var current int

func main() {
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "usage: seesaw LISTEN FIRST SECOND\n")
		flag.PrintDefaults()
	}
	port := flag.Int("httpPort", 6777, "http interface port")
	flag.Parse()
	if flag.NArg() != 3 {
		flag.Usage()
		os.Exit(1)
	}
	localAddr := flag.Arg(0)
	remotes[0] = flag.Arg(1)
	remotes[1] = flag.Arg(2)

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
		go forward(conn, remotes[current])
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
	http.HandleFunc("/toggle", toggleHandler)
	http.HandleFunc("/current", currentHandler)
	http.HandleFunc("/servers", serversHandler)
	http.ListenAndServe(fmt.Sprintf(":%d", port), nil)
}

func toggleHandler(w http.ResponseWriter, req *http.Request) {
	current = 1 - current
	currentHandler(w, req)
}

func currentHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%s\n", remotes[current])
}

func serversHandler(w http.ResponseWriter, req *http.Request) {
	fmt.Fprintf(w, "%s\n", remotes[current])
	fmt.Fprintf(w, "%s\n", remotes[1-current])
}
