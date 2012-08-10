/* Proxy TCP between two servers. Allow you to deploy new code to one server then switch traffic to it without downtime.

Original code by Roger Peppe at http://bit.ly/Oc1YtF
*/
package main

import (
	"fmt"
	"io"
	"net"
	"os"
)

func main() {
	if len(os.Args) != 3 {
		fatal("usage: netfwd local remote")
	}
	localAddr := os.Args[1]
	remoteAddr := os.Args[2]
	local, err := net.Listen("tcp", localAddr)
	if local == nil {
		fatal("cannot listen: %v", err)
	}
	for {
		conn, err := local.Accept()
		if conn == nil {
			fatal("accept failed: %v", err)
		}
		go forward(conn, remoteAddr)
	}
}

func forward(local net.Conn, remoteAddr string) {
	remote, err := net.Dial("tcp", "", remoteAddr)
	if remote == nil {
		fmt.Fprintf(os.Stderr, "remote dial failed: %v\n", err)
		return
	}
	go io.Copy(local, remote)
	go io.Copy(remote, local)
}

func fatal(s string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, "netfwd: %s\n", fmt.Sprintf(s, a))
	os.Exit(2)
}
