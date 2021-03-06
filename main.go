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
	maxRetries = 36
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
			log.Printf("TTY Stream Writer, reconnecting to  %s retry count %d\n", url, retryCount)
		}
		//we need to hide io.Closer implementation or Post will close Stdin on error
		_, err := http.DefaultClient.Post(url, "text/plain", &ReaderOnly{os.Stdin})
		if err != nil {
			log.Println(err)
			continue
		}
		return
	}
}

func startHTTPClientReader(url string) {
	for retryCount := 0; retryCount <= maxRetries; retryCount++ {
		if retryCount > 0 {
			log.Printf("TTY Stream Reader, reconnecting to %s retry count %d\n", url, retryCount)
		}
		resp, err := http.DefaultClient.Get(url)
		if resp != nil && resp.StatusCode != http.StatusOK {
			if _, err := io.Copy(os.Stderr, resp.Body); err != nil {
				log.Println(err)
				return
			}
		}

		if err != nil {
			log.Println(err)
			return
		} else if resp.StatusCode == http.StatusGatewayTimeout || resp.StatusCode == http.StatusBadGateway {
			continue //treat GatewayTimeout as a retryable error
		} else if resp.StatusCode == http.StatusNotFound {
			return //channel does not exist don't retry
		} else if resp.StatusCode == http.StatusConflict {
			log.Println("got error too many connections, when trying to reconnect")
			return
		}

		if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
			log.Println(err)
			continue
		}
		return
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
		log.Println("Starting TTY stream writer client version", version)
		startHTTPClientWriter(os.Args[2])
	case "read":
		log.SetOutput(os.Stderr)
		log.Println("Starting TTY stream reader client version", version)
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
