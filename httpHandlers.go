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
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
)

const (
	readLength           = 50 //we keep this small to prevent locking for too long
	errMsgTooManyReaders = "Only one reader connection supported per channel"
)

var bufferMap = map[string]*MemoryBuffer{}

func getChannelName(r *http.Request) string {
	v := mux.Vars(r)
	if channel, ok := v["channel"]; ok {
		return channel
	}
	return ""
}

func getTag(r *http.Request) string {
	return r.URL.RawQuery
}

func logWriteRequest(w http.ResponseWriter, r *http.Request) {
	channelName := getChannelName(r)
	if channelName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var channel *MemoryBuffer
	var ok bool
	if channel, ok = bufferMap[channelName]; !ok {
		channel = &MemoryBuffer{}
		bufferMap[channelName] = channel
	}

	tag := getTag(r)
	if tag != "" {
		fmt.Fprintf(channel, "====%s====\n", tag)
	}

	channel.WriterStart()
	for {
		_, err := io.CopyN(channel, r.Body, readLength)
		if err != nil {
			break
		}

	}
	if tag != "" {
		fmt.Fprintf(channel, "====%s====\n", tag)
	}
	channel.WriterFinalize()
}

func logReadRequest(w http.ResponseWriter, r *http.Request) {

	channelName := getChannelName(r)
	if channelName == "" {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	var channel *MemoryBuffer
	var ok bool
	if channel, ok = bufferMap[channelName]; !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	w.Header().Add("Content-Type", "text/plan")

	if channel.readerCount > 0 {
		w.WriteHeader(http.StatusConflict)
		w.Write([]byte(errMsgTooManyReaders))

		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		return
	}
	w.WriteHeader(http.StatusOK)
	channel.ReaderStart()

	for {
		len, err := io.CopyN(w, channel, int64(channel.Len()))
		if err != nil {
			if err != io.EOF {
				break

			}
		}
		if len > 0 {
			flusher.Flush()
		} else if channel.finalized && channel.Len() == 0 {
			delete(bufferMap, channelName)
			return
		}
	}

	channel.ReaderFinalize()
}

func logDlRequest(w http.ResponseWriter, r *http.Request) {
	f, err := os.Open(os.Args[0])
	if err != nil {
		return
	}
	defer f.Close()
	s, err := os.Stat(os.Args[0])
	if err != nil {
		return
	}
	w.Header().Add("Content-Length", strconv.Itoa(int(s.Size())))
	w.Header().Add("Content-Type", "application/binary")
	io.Copy(w, f)
}

func startHTTPServer(listener string) error {
	router := mux.NewRouter().StrictSlash(false)
	router.HandleFunc("/log/{channel:[\\w]+}", logWriteRequest).Methods("POST")
	router.HandleFunc("/log/{channel:[\\w]+}", logReadRequest).Methods("GET")
	router.HandleFunc("/dl/client", logDlRequest).Methods("GET")
	return http.ListenAndServe(listener, router)
}
