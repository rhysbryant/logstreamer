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
	"bytes"
	"sync"
	"time"
)

const (
	readerShutdownDelay = 5
)

type MemoryBuffer struct {
	buf         bytes.Buffer
	lock        sync.Mutex
	finalized   bool
	writerCount int
	readerCount int
	expiry      *time.Time
}

func (b *MemoryBuffer) Read(data []byte) (int, error) {
	b.lock.Lock()
	len, err := b.buf.Read(data)
	b.lock.Unlock()
	return len, err
}

func (b *MemoryBuffer) Write(data []byte) (int, error) {
	b.lock.Lock()
	len, err := b.buf.Write(data)
	b.lock.Unlock()
	return len, err
}

func (b *MemoryBuffer) WriterStart() {
	b.finalized = false
	b.expiry = nil
	b.writerCount++
}

func (b *MemoryBuffer) ReaderStart() {
	b.readerCount++
}

func (b *MemoryBuffer) ReaderFinalize() {
	b.readerCount--
}

func (b *MemoryBuffer) WriterFinalize() {
	b.writerCount--
	if b.writerCount == 0 {
		if b.readerCount > 0 {
			//readers are still connected delay finalization as the writer may be just reconnecting
			t := time.Now().Add(readerShutdownDelay * time.Second)
			b.expiry = &t
		}
		b.finalized = true
	}
}

func (b *MemoryBuffer) Finalized() bool {
	return b.finalized && b.expiry == nil || (b.expiry != nil && b.expiry.Before(time.Now()))
}

func (b *MemoryBuffer) Len() int {
	return b.buf.Len()
}
