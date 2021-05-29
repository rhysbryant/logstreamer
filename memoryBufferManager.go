package main

import (
	"errors"
	"sync"
	"time"
)

var (
	ErrChannelDoesNotExist = errors.New("channel does not exist")
)

type FinalizedFunc func(channelName string)

type MemoryBufferManager struct {
	lock          sync.Mutex
	channels      map[string]*MemoryBuffer
	finalizedFunc FinalizedFunc
}

func NewBufferManager(f FinalizedFunc) *MemoryBufferManager {
	m := MemoryBufferManager{}
	m.channels = map[string]*MemoryBuffer{}
	m.finalizedFunc = f
	go m.cleanupThread()
	return &m
}

func (mbm *MemoryBufferManager) cleanupThread() {
	timerCh := time.Tick(time.Minute)

	for {
		<-timerCh
		mbm.lock.Lock()
		channels := mbm.channels
		for key, channel := range channels {
			if channel.Finalized() {
				delete(mbm.channels, key)
				mbm.finalizedFunc(key)
			}
		}
		mbm.lock.Unlock()
	}

}

func (mbm *MemoryBufferManager) CheckoutForRead(channelName string) (*MemoryBuffer, error) {
	mbm.lock.Lock()
	defer mbm.lock.Unlock()

	if ch, ok := mbm.channels[channelName]; ok {
		return ch, nil
	}
	return nil, ErrChannelDoesNotExist

}

func (mbm *MemoryBufferManager) CheckoutForWrite(channelName string) *MemoryBuffer {
	mbm.lock.Lock()
	defer mbm.lock.Unlock()
	if ch, ok := mbm.channels[channelName]; ok {
		return ch
	}
	mb := MemoryBuffer{}
	mbm.channels[channelName] = &mb
	return &mb
}
