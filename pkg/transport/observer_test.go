package transport

import (
	"errors"
	"net"
	"testing"
	"time"
)

func TestNewObserver(t *testing.T) {
	handler := func(*Message, *net.UDPAddr) {}
	if _, err := NewObserver(net.Interface{Name: "test0"}, 1234, nil, handler); err == nil {
		t.Fatal("NewObserver() with an empty key succeeded, want an error")
	}

	iface := net.Interface{Name: "test0"}
	observer, err := NewObserver(iface, 1234, []byte("test-key"), handler)
	if err != nil {
		t.Fatalf("NewObserver() error = %v", err)
	}
	if observer.Port != 1234 {
		t.Errorf("Port = %d, want 1234", observer.Port)
	}
	if observer.Interface.Name != iface.Name {
		t.Errorf("Interface.Name = %q, want %q", observer.Interface.Name, iface.Name)
	}
	if observer.States == nil {
		t.Fatal("States is nil")
	}
	if state, ok := observer.States.Get([6]byte{}); !ok || state == nil || state.Encryption == nil {
		t.Fatal("default state is missing or has no encryption key")
	}
}

func TestObserverSendNotConnected(t *testing.T) {
	observer := &Observer{}
	err := observer.Send(Message{}, &net.UDPAddr{})
	if !errors.Is(err, ErrNotConnected) {
		t.Fatalf("Send() error = %v, want ErrNotConnected", err)
	}
	var notConnected NotConnectedError
	if !errors.As(err, &notConnected) {
		t.Errorf("Send() error type = %T, want NotConnectedError", err)
	}
}

func TestObserverSendDelegates(t *testing.T) {
	observer := &Observer{}
	wantErr := errors.New("send failed")
	message := Message{RequestId: 42}
	address := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 4321}
	calls := 0

	observer.sender = func(gotMessage Message, gotAddress *net.UDPAddr) error {
		calls++
		if gotMessage.RequestId != message.RequestId {
			t.Errorf("sender received RequestId = %d, want %d", gotMessage.RequestId, message.RequestId)
		}
		if gotAddress != address {
			t.Errorf("sender received address %p, want %p", gotAddress, address)
		}
		return wantErr
	}

	if err := observer.Send(message, address); !errors.Is(err, wantErr) {
		t.Errorf("Send() error = %v, want %v", err, wantErr)
	}
	if calls != 1 {
		t.Errorf("sender called %d times, want 1", calls)
	}
}

func TestObserverResolveTargetWithDefaultKey(t *testing.T) {
	observer, err := NewObserver(net.Interface{}, 0, []byte("default-key"), nil)
	if err != nil {
		t.Fatalf("NewObserver() error = %v", err)
	}
	defaultState, _ := observer.States.Get([6]byte{})
	message := Message{
		Method:    MethodDiscover,
		SwitchMac: MACAddress{1, 2, 3, 4, 5, 6},
		ClientMac: MACAddress{6, 5, 4, 3, 2, 1},
		RequestId: 17,
		Body:      []byte("value"),
	}

	target, got, err := observer.resolveTarget(encodeTestMessage(t, defaultState, message))
	if err != nil {
		t.Fatalf("resolveTarget() error = %v", err)
	}
	if target == nil {
		t.Fatal("resolveTarget() target is nil")
	}
	if target.Encryption != defaultState.Encryption {
		t.Error("resolved target does not have the default encryption key")
	}
	if got == nil || got.RequestId != message.RequestId || got.Method != message.Method {
		t.Fatalf("resolved message = %#v, want RequestId %d and method %d", got, message.RequestId, message.Method)
	}
	if len(got.Body) != 5 || string(got.Body) != string(message.Body) {
		t.Errorf("resolved body = %#v, want %#v", got.Body, message.Body)
	}
	if registered, ok := observer.States.Get([6]byte(message.SwitchMac)); !ok || registered != target {
		t.Error("resolved target was not registered in the state table")
	}
}

func TestObserverResolveTargetWithPeerKey(t *testing.T) {
	observer, err := NewObserver(net.Interface{}, 0, []byte("default-key"), nil)
	if err != nil {
		t.Fatalf("NewObserver() error = %v", err)
	}
	peerStates, err := NewStateTable([]byte("peer-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	peerState, _ := peerStates.Get([6]byte{})
	observer.States.Set([6]byte{9, 8, 7, 6, 5, 4}, peerState)

	message := Message{
		Method:    MethodPostResponse,
		SwitchMac: MACAddress{4, 3, 2, 1, 0, 9},
		ClientMac: MACAddress{1, 3, 5, 7, 9, 2},
		RequestId: 29,
	}
	target, got, err := observer.resolveTarget(encodeTestMessage(t, peerState, message))
	if err != nil {
		t.Fatalf("resolveTarget() error = %v", err)
	}
	if target == nil || target == peerState {
		t.Fatalf("resolveTarget() target = %p, want a new state cloned from the peer key", target)
	}
	if target.Encryption != peerState.Encryption {
		t.Error("resolved target does not share the peer encryption key")
	}
	if got == nil || got.RequestId != message.RequestId || got.Method != message.Method {
		t.Errorf("resolved message = %#v, want RequestId %d and method %d", got, message.RequestId, message.Method)
	}
	if registered, ok := observer.States.Get([6]byte(message.SwitchMac)); !ok || registered != target {
		t.Error("discovered peer target was not registered in the state table")
	}
}

func TestObserverResolveTargetWithoutMatchingKey(t *testing.T) {
	observer, err := NewObserver(net.Interface{}, 0, []byte("default-key"), nil)
	if err != nil {
		t.Fatalf("NewObserver() error = %v", err)
	}
	peerStates, err := NewStateTable([]byte("different-peer-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	peerState, _ := peerStates.Get([6]byte{})
	message := Message{
		Method:    MethodDiscover,
		SwitchMac: MACAddress{1, 2, 3, 4, 5, 6},
		ClientMac: MACAddress{6, 5, 4, 3, 2, 1},
	}

	target, got, err := observer.resolveTarget(encodeTestMessage(t, peerState, message))
	if err != nil {
		t.Fatalf("resolveTarget() error = %v", err)
	}
	if target != nil || got != nil {
		t.Errorf("resolveTarget() = (%p, %#v), want (nil, nil)", target, got)
	}
}

func TestObserverListenDispatchesResolvedMessage(t *testing.T) {
	type received struct {
		message *Message
		address *net.UDPAddr
	}
	receivedMessages := make(chan received, 1)
	observer, err := NewObserver(net.Interface{}, 0, []byte("default-key"), func(message *Message, address *net.UDPAddr) {
		receivedMessages <- received{message: message, address: address}
	})
	if err != nil {
		t.Fatalf("NewObserver() error = %v", err)
	}
	defaultState, _ := observer.States.Get([6]byte{})
	message := Message{
		Method:    MethodDiscover,
		SwitchMac: MACAddress{1, 2, 3, 4, 5, 6},
		ClientMac: MACAddress{6, 5, 4, 3, 2, 1},
		RequestId: 71,
		Body:      []byte("captured"),
	}

	listener := newMockUDPConnection()
	defer listener.Close()

	listenErr := make(chan error, 1)
	go func() {
		listenErr <- observer.transportBase.listenPackets(listener, func(message *Message, address *net.UDPAddr) {
			observer.Handler(message, address)
		})
	}()

	listener.inject(encodeTestMessage(t, defaultState, message), &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 4000})

	select {
	case got := <-receivedMessages:
		if got.message == nil || got.message.RequestId != message.RequestId || got.message.Method != message.Method {
			t.Errorf("handler received message %#v, want RequestId %d and method %d", got.message, message.RequestId, message.Method)
		}
		if got.address == nil || !got.address.IP.Equal(net.IPv4(127, 0, 0, 1)) {
			t.Errorf("handler received address %#v, want loopback", got.address)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("response handler was not called")
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
		t.Fatal("listenPackets() did not return after closing the socket")
	}
}

func TestObserverConnectSendAndClose(t *testing.T) {
	observer, err := NewObserver(net.Interface{}, 0, []byte("observer-send-key"), nil)
	if err != nil {
		t.Fatalf("NewObserver() error = %v", err)
	}
	conn := newTestConn(t)

	observer.onConnect(conn)
	message := Message{
		Method:    MethodDiscover,
		SwitchMac: MACAddress{1, 2, 3, 4, 5, 6},
		ClientMac: MACAddress{6, 5, 4, 3, 2, 1},
		RequestId: 103,
	}
	if err := observer.Send(message, testPeerAddr); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	defaultState, _ := observer.States.Get([6]byte{})
	assertReceivedMessage(t, conn, defaultState, message)

	handled := make(chan bool, 1)
	observer.Handler = func(got *Message, addr *net.UDPAddr) {
		handled <- got != nil && got.RequestId == message.RequestId && addr != nil
	}
	observer.handle(&message, testPeerAddr)
	if ok := <-handled; !ok {
		t.Error("handle() did not forward the message and address to Observer.Handler")
	}

	observer.onClose()
	if err := observer.Send(message, testPeerAddr); !errors.Is(err, ErrNotConnected) {
		t.Errorf("Send() after onClose error = %v, want ErrNotConnected", err)
	}
}

func TestObserverHandleWithNilHandler(t *testing.T) {
	observer, err := NewObserver(net.Interface{}, 0, []byte("nil-handler-key"), nil)
	if err != nil {
		t.Fatalf("NewObserver() error = %v", err)
	}
	observer.handle(&Message{}, testPeerAddr)
}

func TestObserverConnectCloseSendIsRaceFree(t *testing.T) {
	observer, err := NewObserver(net.Interface{}, 0, []byte("observer-race-key"), nil)
	if err != nil {
		t.Fatalf("NewObserver() error = %v", err)
	}
	conn := newTestConn(t)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			observer.onConnect(conn)
			observer.onClose()
		}
	}()
	for i := 0; i < 100; i++ {
		_ = observer.Send(Message{}, testPeerAddr)
	}
	<-done
}
