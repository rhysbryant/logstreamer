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

func printArgError() {
	fmt.Println("expected " + os.Args[0] + " {server :{port}|read {url}|write {url}}")
}

func main() {
	log.SetOutput(os.Stderr)

	if len(os.Args) < 3 {
		printArgError()
		return
	}
	switch os.Args[1] {
	case "write":
		for retryCount := 0; retryCount <= maxRetries; retryCount++ {
			if retryCount > 0 {
				log.Printf("sending Request to %s retry count %d\n", os.Args[2], retryCount)
			}
			_, err := http.DefaultClient.Post(os.Args[2], "text/plain", os.Stdin)
			if err != nil {
				log.Println(err)
				continue
			}
			return
		}
	case "read":
		for retryCount := 0; retryCount <= maxRetries; retryCount++ {

			resp, err := http.DefaultClient.Get(os.Args[2])
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
				log.Printf("sending Request to %s retry count %d\n", os.Args[2], retryCount)
			}
		}
	case "server":
		err := startHTTPServer(os.Args[2])
		if err != nil {
			log.Fatal(err)
		}
	default:
		printArgError()

	}
}
