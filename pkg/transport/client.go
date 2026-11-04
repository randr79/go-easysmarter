package transport

import (
	"context"
	"log"
	"net"
	"sync"
	"time"
)

type responseHandlers = syncMap[uint16, ResponseHandler]

func NewResponseHandlers() *responseHandlers {
	return &syncMap[uint16, ResponseHandler]{
		items: make(map[uint16]ResponseHandler),
	}
}

type Client struct {
	transportBase
	ResponseHandlers *responseHandlers
	mu               sync.RWMutex
	sender           func(msg Message, address *net.UDPAddr, responsehandler ResponseHandler, responseTimeout time.Duration) error
}

type ResponseHandler func(msg *Message, address *net.UDPAddr) (drop bool)

func (s *Client) Send(msg Message, address *net.UDPAddr, responsehandler ResponseHandler, responseTimeout time.Duration) error {
	if sender := s.getSender(); sender == nil {
		return ErrNotConnected
	} else {
		return sender(msg, address, responsehandler, responseTimeout)
	}
}

// blocks untill closed, close by canceling ctx
func (s *Client) Listen(ctx context.Context) error {
	return s.transportBase.listen(ctx, s)
}
func (s *Client) getSender() func(msg Message, address *net.UDPAddr, responsehandler ResponseHandler, responseTimeout time.Duration) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sender
}

func (s *Client) setSender(sender func(msg Message, address *net.UDPAddr, responsehandler ResponseHandler, responseTimeout time.Duration) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sender = sender
}

func (s *Client) onClose() {
	s.setSender(nil)
}
func (s *Client) handle(msg *Message, addr *net.UDPAddr) {
	if !s.ResponseHandlers.Expire(msg.RequestId, func(handler ResponseHandler) bool {
		return handler(msg, addr)
	}) {
		log.Printf("msg received from %v while handler for %d is gone", addr, msg.RequestId)
	}
}

func (s *Client) onConnect(conn UDPConn) {
	s.setSender(func(msg Message, addr *net.UDPAddr, resp ResponseHandler, timeout time.Duration) error {
		s.ResponseHandlers.Set(msg.RequestId, resp)
		timer := time.AfterFunc(timeout, func() {
			//timeout expired, if we still have it we call it with nil,nil to indicate done
			if handler, ok := s.ResponseHandlers.Pop(msg.RequestId); ok {
				handler(nil, nil)
			}
		})
		if err := s.send(conn, msg, addr); err != nil {
			timer.Stop()
			s.ResponseHandlers.Del(msg.RequestId)
			return err
		}
		return nil
	})
}

func NewClient(iface net.Interface, port uint16, key []byte) (*Client, error) {
	if st, err := NewStateTable(key); err != nil {
		return nil, err
	} else {
		return &Client{
			transportBase: transportBase{
				Port:        port,
				Interface:   iface,
				States:      st,
				ConnFactory: NewUDPConn,
			},
			ResponseHandlers: NewResponseHandlers(),
		}, nil
	}

}
