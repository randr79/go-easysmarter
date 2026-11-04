package transport

import (
	"context"
	"errors"
	"log"
	"net"
	"os"

	transportOs "github.com/randr79/go-easysmarter/pkg/transport/os"
)

type transportWorkflow interface {
	onConnect(UDPConn)
	onClose()
	handle(*Message, *net.UDPAddr)
}

type transportBase struct {
	Port        uint16
	Interface   net.Interface
	States      *stateTable
	ConnFactory UDPConnFactory
}

func (s transportBase) listen(
	ctx context.Context,
	workflow transportWorkflow,
) error {
	if s.ConnFactory == nil {
		return ErrNoConnFactory
	}
	udpconn, err := s.ConnFactory(s.Port, s.Interface)
	if err != nil {
		return err
	}

	cctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-cctx.Done()
		workflow.onClose()
		_ = udpconn.Close()
	}()

	workflow.onConnect(udpconn)
	return s.listenPackets(udpconn, workflow.handle)
}

func (s transportBase) listenPackets(conn UDPConn, handler func(*Message, *net.UDPAddr)) error {
	msgBuffer := make([]byte, 65535)

	for {
		dataCount, addr, err := conn.ReadFromUDP(msgBuffer)
		if err == nil {
			// packet received
		} else if errors.Is(err, os.ErrDeadlineExceeded) {
			log.Print(err)
			continue
		} else if errors.Is(err, net.ErrClosed) {
			return err
		} else if transportOs.ConnectionBroken(err) {
			return err
		} else {
			log.Print(err)
			continue
		}

		if target, msg, err := s.resolveTarget(msgBuffer[:dataCount]); err != nil {
			log.Print(err)
		} else if target == nil {
			log.Printf("unreadable msg from %v (no such encryption key)", addr)
		} else {
			go handler(msg, addr)
		}
	}
}

func (s transportBase) findEncryptedPeer(data []byte) (*State, *wireHeader, []byte) {
	for mac, target := range s.States.Items() {
		if target.Encryption == nil {
			continue
		} else if head, content, _ := target.decode(data); head == nil {
			continue
		} else if head.SwitchMac == mac {
			return target, head, content
		} else {
			newState := &State{Encryption: target.Encryption}
			s.States.Set(head.SwitchMac, newState)
			return newState, head, content
		}
	}

	return nil, nil, nil
}

func (s transportBase) resolveTarget(data []byte) (*State, *Message, error) {
	defaultState, _ := s.States.Get([6]byte{})
	header, content, err := defaultState.decode(data)
	if err != nil {
		return nil, nil, err
	}

	var target *State
	if header == nil {
		target, header, content = s.findEncryptedPeer(data)
	} else if existing, exists := s.States.Get(header.SwitchMac); exists {
		target = existing
	} else {
		target = &State{Encryption: defaultState.Encryption}
		s.States.Set(header.SwitchMac, target)
	}

	if target == nil {
		return nil, nil, nil
	}
	msg := &Message{
		Method:    Method(header.Method),
		SwitchMac: header.SwitchMac,
		ClientMac: header.ClientMac,
		RequestId: header.RequestId,
		Fragment:  header.Fragment,
		Body:      content,
		State:     StatusCode(header.Status),
	}

	return target, msg, nil
}

func (s transportBase) send(conn UDPConn, msg Message, addr *net.UDPAddr) error {
	var errs []error
	var contentLength uint16
	if cl := len(msg.Body); cl > maxContentLength {
		errs = append(errs, ContentLengthExceededError(cl))
	} else {
		contentLength = uint16(cl)
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	defKey, _ := s.States.Get([6]byte{0, 0, 0, 0, 0, 0})
	target := &State{
		Encryption: defKey.Encryption,
	}
	target = s.States.GetDefault([6]byte(msg.SwitchMac), target)

	wireheader := &wireHeader{
		Version:   1,
		Method:    byte(msg.Method),
		SwitchMac: msg.SwitchMac,
		ClientMac: msg.ClientMac,
		RequestId: msg.RequestId,
		Length:    contentLength,
		Status:    uint32(msg.State),
		Fragment:  msg.Fragment,
		Flags:     0,
		TokenId:   target.TokenId,
		CRC:       0,
	}
	if payload, err := target.encode(*wireheader, msg.Body); err != nil {
		return err
	} else if _, err := conn.WriteToUDP(payload, addr); err != nil {
		return err
	} else {
		return nil
	}
}
