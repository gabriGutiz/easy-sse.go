package ssebroadcaster

import (
	"log"
	"sync"
)

type ConnChan struct {
	ConnId  uint32
	Id      string
	Channel chan string
}

type ConnPool struct {
	chans   []*ConnChan
	chansMu sync.Mutex
}

func NewConnPool() ConnPool {
	return ConnPool{}
}

func (connPool *ConnPool) AddChannel(c *ConnChan) {
	connPool.chansMu.Lock()
	connPool.chans = append(connPool.chans, c)
	connPool.chansMu.Unlock()
}

func (connPool *ConnPool) RemoveChannel(connId uint32) {
	connPool.chansMu.Lock()
	defer connPool.chansMu.Unlock()

	for i, c := range connPool.chans {
		if c.ConnId == connId {
			connPool.chans[i] = connPool.chans[len(connPool.chans)-1]
			connPool.chans = connPool.chans[:len(connPool.chans)-1]
			return
		}
	}
}

func (connPool *ConnPool) Broadcast(id string, msg string) {
	connPool.chansMu.Lock()
	for _, c := range connPool.chans {
		if c.Id == id {
			select {
			case c.Channel <- msg:
				log.Printf("sent to %s: %s", id, msg)
			default:
				log.Printf("channel busy for %s, dropping message", id)
			}
		}
	}
	connPool.chansMu.Unlock()
}
