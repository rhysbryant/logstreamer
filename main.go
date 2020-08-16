package main

/**
	This file is part of logstreamer.
	logstreamer - printer status page and protocol relay for daVinci jr 3d printers
    logstreamer is free software: you can redistribute it and/or modify
    it under the terms of the GNU General Public License as published by
    the Free Software Foundation, either version 3 of the License, or
    (at your option) any later version.
    logstreamer is distributed in the hope that it will be useful,
    but WITHOUT ANY WARRANTY; without even the implied warranty of
    MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
    GNU General Public License for more details.
    You should have received a copy of the GNU General Public License
    along with logstreamer.  If not, see <http://www.gnu.org/licenses/>.
**/
import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

const (
	maxRetries = 12
)

var (
	version = "dev-build"
)

func printArgError() {
	fmt.Println("expected " + os.Args[0] + " {server :{port}|read {url}|write {url}}")
}

func startHTTPClientWriter(url string) {
	for retryCount := 0; retryCount <= maxRetries; retryCount++ {
		if retryCount > 0 {
			log.Printf("sending Request to %s retry count %d\n", url, retryCount)
		}
		_, err := http.DefaultClient.Post(url, "text/plain", os.Stdin)
		if err != nil {
			log.Println(err)
			continue
		}
		return
	}
}

func startHTTPClientReader(url string) {
	for retryCount := 0; retryCount <= maxRetries; retryCount++ {

		resp, err := http.DefaultClient.Get(url)
		if err != nil {
			log.Fatal(err)
		} else if resp.StatusCode == http.StatusNotFound {
			return
		} else if resp.StatusCode == http.StatusConflict {
			log.Printf(resp.Status)
			return
		}

		if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
			log.Fatal(err)
		}
		if retryCount > 0 {
			log.Printf("sending Request to %s retry count %d\n", url, retryCount)
		}
	}
}

func main() {

	if len(os.Args) < 3 {
		log.Println("version", version)
		printArgError()
		return
	}

	switch os.Args[1] {
	case "write":
		log.SetOutput(os.Stderr)
		startHTTPClientWriter(os.Args[2])
	case "read":
		log.SetOutput(os.Stderr)
		startHTTPClientReader(os.Args[2])
	case "server":
		log.Println("Starting server version", version)
		err := startHTTPServer(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
	default:
		printArgError()

	}
}
