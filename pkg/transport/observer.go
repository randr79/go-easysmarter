package transport

import (
	"context"
	"net"
	"sync"
)

type Observer struct {
	transportBase
	Handler ObserverHandler
	mu      sync.RWMutex
	sender  func(msg Message, address *net.UDPAddr) error
}

type ObserverHandler func(msg *Message, address *net.UDPAddr)

func (s *Observer) Send(msg Message, address *net.UDPAddr) error {
	if sender := s.getSender(); sender == nil {
		return ErrNotConnected
	} else {
		return sender(msg, address)
	}
}

// blocks untill closed, close by canceling ctx
func (s *Observer) Listen(ctx context.Context) error {
	return s.transportBase.listen(ctx, s)
}
func (s *Observer) getSender() func(msg Message, address *net.UDPAddr) error {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sender
}

func (s *Observer) setSender(sender func(msg Message, address *net.UDPAddr) error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sender = sender
}

func (s *Observer) onClose() {
	s.setSender(nil)
}

func (s *Observer) handle(msg *Message, addr *net.UDPAddr) {
	if s.Handler != nil {
		s.Handler(msg, addr)
	}
}

func (s *Observer) onConnect(conn UDPConn) {
	s.setSender(func(msg Message, addr *net.UDPAddr) error {
		return s.send(conn, msg, addr)
	})
}

func NewObserver(iface net.Interface, port uint16, key []byte, handler ObserverHandler) (*Observer, error) {
	if st, err := NewStateTable(key); err != nil {
		return nil, err
	} else {
		return &Observer{
			transportBase: transportBase{
				Port:        port,
				Interface:   iface,
				States:      st,
				ConnFactory: NewUDPConn,
			},
			Handler: handler,
		}, nil
	}

}

type SharedStateLink struct {
	Interface net.Interface
	Port      uint16
	Handler   ObserverHandler
}

func NewSharedStateObservers(links []SharedStateLink, key []byte) ([]*Observer, error) {
	if st, err := NewStateTable(key); err != nil {
		return nil, err
	} else {
		res := make([]*Observer, len(links))
		for i, link := range links {
			res[i] = &Observer{
				transportBase: transportBase{
					Port:        link.Port,
					Interface:   link.Interface,
					States:      st,
					ConnFactory: NewUDPConn,
				},
				Handler: link.Handler,
			}
		}
		return res, nil
	}

}
