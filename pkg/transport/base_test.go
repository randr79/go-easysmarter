package transport

import (
	"context"
	"errors"
	"net"
	"testing"
	"time"
)

type testTransportWorkflow struct {
	connected chan struct{}
	closed    chan struct{}
}

func (s *testTransportWorkflow) onConnect(UDPConn) {
	s.connected <- struct{}{}
}

func (s *testTransportWorkflow) onClose() {
	s.closed <- struct{}{}
}

func (s *testTransportWorkflow) handle(*Message, *net.UDPAddr) {}

func TestTransportBaseListenLifecycle(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	workflow := &testTransportWorkflow{
		connected: make(chan struct{}, 1),
		closed:    make(chan struct{}, 1),
	}
	factory := newMockConnectionFactory()
	listenErr := make(chan error, 1)
	go func() {
		listenErr <- (transportBase{ConnFactory: factory.create}).listen(ctx, workflow)
	}()

	select {
	case <-workflow.connected:
	case <-time.After(2 * time.Second):
		t.Fatal("listen() did not connect")
	}

	cancel()
	select {
	case <-workflow.closed:
	case <-time.After(2 * time.Second):
		t.Fatal("listen() did not invoke onClose after cancellation")
	}
	select {
	case err := <-listenErr:
		if !errors.Is(err, net.ErrClosed) {
			t.Errorf("listen() error = %v, want %v", err, net.ErrClosed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("listen() did not return after cancellation")
	}
}

func TestResolveTargetReusesExistingPeerState(t *testing.T) {
	states, err := NewStateTable([]byte("resolve-default-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	defaultState, _ := states.Get([6]byte{})
	message := Message{
		Header: Header{
			Method:    MethodGetResponse,
			SwitchMac: MACAddress{1, 2, 3, 4, 5, 6},
			ClientMac: MACAddress{6, 5, 4, 3, 2, 1},
			RequestId: 321,
			Fragment:  4,
			State:     5,
		},
		Body: []byte("payload"),
	}
	peer := &State{Encryption: defaultState.Encryption, TokenId: 99}
	states.Set([6]byte(message.SwitchMac), peer)

	target, got, err := (transportBase{States: states}).resolveTarget(encodeTestMessage(t, defaultState, message))
	if err != nil {
		t.Fatalf("resolveTarget() error = %v", err)
	}
	if target != peer {
		t.Fatalf("resolveTarget() target = %p, want existing state %p", target, peer)
	}
	if got == nil {
		t.Fatal("resolveTarget() message = nil")
	}
	if got.RequestId != message.RequestId || got.Method != message.Method || got.Fragment != message.Fragment ||
		got.State != message.State || got.SwitchMac != message.SwitchMac ||
		got.ClientMac != message.ClientMac || string(got.Body) != string(message.Body) {
		t.Errorf("resolveTarget() message = %#v, want %#v", got, message)
	}
}

func TestFindEncryptedPeerSkipsUnencryptedStates(t *testing.T) {
	states, err := NewStateTable([]byte("default-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	states.Set([6]byte{1, 1, 1, 1, 1, 1}, &State{})
	otherStates, err := NewStateTable([]byte("unknown-peer-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	otherState, _ := otherStates.Get([6]byte{})
	message := Message{
		Header: Header{
			Method:    MethodDiscover,
			SwitchMac: MACAddress{9, 8, 7, 6, 5, 4},
			ClientMac: MACAddress{4, 5, 6, 7, 8, 9},
		},
	}

	target, header, body := (transportBase{States: states}).findEncryptedPeer(encodeTestMessage(t, otherState, message))
	if target != nil || header != nil || body != nil {
		t.Errorf("findEncryptedPeer() = (%p, %#v, %x), want all nil", target, header, body)
	}
}

func TestListenPacketsSkipsUnreadablePackets(t *testing.T) {
	states, err := NewStateTable([]byte("listen-default-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	defaultState, _ := states.Get([6]byte{})
	listener := newMockUDPConnection()
	t.Cleanup(func() { listener.Close() })
	source := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 4000}

	message := Message{
		Header: Header{
			Method:    MethodDiscover,
			SwitchMac: MACAddress{1, 2, 3, 4, 5, 6},
			ClientMac: MACAddress{6, 5, 4, 3, 2, 1},
			RequestId: 654,
		},
	}
	unrecognizedKey, err := NewStateTable([]byte("unrecognized-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	unrecognizedState, _ := unrecognizedKey.Get([6]byte{})
	unreadable := encodeTestMessage(t, unrecognizedState, message)

	valid := encodeTestMessage(t, defaultState, message)

	handlerCalled := make(chan *Message, 1)
	listenErr := make(chan error, 1)
	go func() {
		listenErr <- (transportBase{States: states}).listenPackets(listener, func(msg *Message, _ *net.UDPAddr) {
			handlerCalled <- msg
		})
	}()

	for _, packet := range [][]byte{unreadable} {
		listener.inject(packet, source)
	}
	listener.inject(valid, source)

	select {
	case got := <-handlerCalled:
		if got == nil || got.RequestId != message.RequestId {
			t.Errorf("handler got %#v, want RequestId %d", got, message.RequestId)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("listener did not continue after an unreadable packet")
	}

	if err := listener.Close(); err != nil {
		t.Fatalf("Close listener error = %v", err)
	}
	select {
	case err := <-listenErr:
		if !errors.Is(err, net.ErrClosed) {
			t.Errorf("listenPackets() error = %v, want %v", err, net.ErrClosed)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("listenPackets() did not return after socket close")
	}
}

func TestTransportBaseListenWithoutFactory(t *testing.T) {
	err := (transportBase{}).listen(context.Background(), &testTransportWorkflow{})
	if !errors.Is(err, ErrNoConnFactory) {
		t.Errorf("listen() error = %v, want %v", err, ErrNoConnFactory)
	}
}

func TestTransportBaseListenFactoryError(t *testing.T) {
	factory := newMockConnectionFactory()
	factory.err = errors.New("boom")
	workflow := &testTransportWorkflow{
		connected: make(chan struct{}, 1),
		closed:    make(chan struct{}, 1),
	}
	err := (transportBase{ConnFactory: factory.create}).listen(context.Background(), workflow)
	if !errors.Is(err, factory.err) {
		t.Errorf("listen() error = %v, want %v", err, factory.err)
	}
	select {
	case <-workflow.connected:
		t.Error("onConnect called despite factory error")
	default:
	}
}

func TestTransportBaseSendUsesDefaultKeyForNewPeer(t *testing.T) {
	states, err := NewStateTable([]byte("default-send-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	defaultState, _ := states.Get([6]byte{})
	conn := newTestConn(t)

	message := Message{
		Header: Header{
			Method:    MethodPostRequest,
			SwitchMac: MACAddress{1, 2, 3, 4, 5, 6},
			ClientMac: MACAddress{6, 5, 4, 3, 2, 1},
			RequestId: 101,
			Fragment:  3,
			State:     8,
		},
		Body: []byte("outbound"),
	}
	base := transportBase{States: states}
	if err := base.send(conn, message, testPeerAddr); err != nil {
		t.Fatalf("send() error = %v", err)
	}
	assertReceivedMessage(t, conn, defaultState, message)

	peer, ok := states.Get([6]byte(message.SwitchMac))
	if !ok || peer == nil || peer.Encryption != defaultState.Encryption {
		t.Fatal("send() did not register the peer with the default encryption key")
	}
}

func TestTransportBaseSendPreservesPeerState(t *testing.T) {
	states, err := NewStateTable([]byte("default-send-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	peerStates, err := NewStateTable([]byte("peer-send-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	peerState, _ := peerStates.Get([6]byte{})
	peerState.TokenId = 23
	message := Message{
		Header: Header{
			Method:    MethodGetRequest,
			SwitchMac: MACAddress{7, 6, 5, 4, 3, 2},
			ClientMac: MACAddress{2, 3, 4, 5, 6, 7},
			RequestId: 102,
		},
	}
	states.Set([6]byte(message.SwitchMac), peerState)

	conn := newTestConn(t)

	if err := (transportBase{States: states}).send(conn, message, testPeerAddr); err != nil {
		t.Fatalf("send() error = %v", err)
	}
	assertReceivedMessage(t, conn, peerState, message)
}

func TestTransportBaseSendRejectsInvalidMessage(t *testing.T) {
	validMac := MACAddress{1, 2, 3, 4, 5, 6}
	tests := []struct {
		name     string
		message  Message
		wantMacs int
		wantSize bool
	}{
		{name: "oversized body", message: Message{Header: Header{SwitchMac: validMac, ClientMac: validMac}, Body: make([]byte, maxContentLength+1)}, wantSize: true},
		{name: "everything invalid", message: Message{Body: make([]byte, maxContentLength+1)}, wantMacs: 2, wantSize: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			states, err := NewStateTable([]byte("default-send-key"))
			if err != nil {
				t.Fatalf("NewStateTable() error = %v", err)
			}
			conn := newTestConn(t)

			err = (transportBase{States: states}).send(conn, tt.message, testPeerAddr)
			if err == nil {
				t.Fatal("send() succeeded for invalid message, want an error")
			}
			if len(conn.Written()) != 0 {
				t.Error("send() wrote a packet for an invalid message")
			}

			var sizeErr ContentLengthExceededError
			if got := errors.As(err, &sizeErr); got != tt.wantSize {
				t.Errorf("send() ContentLengthExceededError present = %v, want %v", got, tt.wantSize)
			}
		})
	}
}

func TestTransportBaseSendReturnsWriteError(t *testing.T) {
	states, err := NewStateTable([]byte("write-error-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	conn := newMockUDPConnection()
	if err := conn.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	message := Message{
		Header: Header{
			Method:    MethodDiscover,
			SwitchMac: MACAddress{1, 2, 3, 4, 5, 6},
			ClientMac: MACAddress{6, 5, 4, 3, 2, 1},
		},
	}
	err = (transportBase{States: states}).send(conn, message, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 9})
	if err == nil {
		t.Fatal("send() succeeded on a closed socket, want an error")
	}
	if errors.Is(err, net.ErrClosed) {
		return
	}
	var opErr *net.OpError
	if !errors.As(err, &opErr) {
		t.Errorf("send() error = %T, want a network write error", err)
	}
}
