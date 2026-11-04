package transport

import (
	"errors"
	"net"
	"testing"
	"time"
)

var managerTestDefaultMac = [6]byte{}

func TestNewClient(t *testing.T) {
	if _, err := NewClient(net.Interface{Name: "test0"}, 1234, nil); err == nil {
		t.Fatal("NewClient() with an empty key succeeded, want an error")
	}

	iface := net.Interface{Name: "test0"}
	client, err := NewClient(iface, 1234, []byte("test-key"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client.Port != 1234 {
		t.Errorf("Port = %d, want 1234", client.Port)
	}
	if client.Interface.Name != iface.Name {
		t.Errorf("Interface.Name = %q, want %q", client.Interface.Name, iface.Name)
	}
	if client.ResponseHandlers == nil {
		t.Fatal("ResponseHandlers is nil")
	}
	if state, ok := client.States.Get(managerTestDefaultMac); !ok || state == nil || state.Encryption == nil {
		t.Fatal("default state is missing or has no encryption key")
	}
}

func TestClientSendNotConnected(t *testing.T) {
	client := &Client{}
	err := client.Send(Message{}, &net.UDPAddr{}, nil, time.Second)
	if !errors.Is(err, ErrNotConnected) {
		t.Fatalf("Send() error = %v, want ErrNotConnected", err)
	}
	var notConnected NotConnectedError
	if !errors.As(err, &notConnected) {
		t.Errorf("Send() error type = %T, want NotConnectedError", err)
	}
}

func TestClientSendDelegates(t *testing.T) {
	client := &Client{}
	wantErr := errors.New("send failed")
	message := Message{RequestId: 42}
	address := &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 4321}
	timeout := 3 * time.Second
	handlerCalled := false
	handler := func(*Message, *net.UDPAddr) bool {
		handlerCalled = true
		return true
	}
	calls := 0

	client.sender = func(gotMessage Message, gotAddress *net.UDPAddr, gotHandler ResponseHandler, gotTimeout time.Duration) error {
		calls++
		if gotMessage.RequestId != message.RequestId {
			t.Errorf("sender received RequestId = %d, want %d", gotMessage.RequestId, message.RequestId)
		}
		if gotAddress != address {
			t.Errorf("sender received address %p, want %p", gotAddress, address)
		}
		if gotHandler == nil {
			t.Error("sender received a nil response handler")
		} else if !gotHandler(&Message{}, address) {
			t.Error("sender received a response handler with an unexpected result")
		}
		if gotTimeout != timeout {
			t.Errorf("sender received timeout = %s, want %s", gotTimeout, timeout)
		}
		return wantErr
	}

	if err := client.Send(message, address, handler, timeout); !errors.Is(err, wantErr) {
		t.Errorf("Send() error = %v, want %v", err, wantErr)
	}
	if calls != 1 {
		t.Errorf("sender called %d times, want 1", calls)
	}
	if !handlerCalled {
		t.Error("sender did not receive and invoke the supplied response handler")
	}
}

func encodeTestMessage(t *testing.T, keyState *State, message Message) []byte {
	t.Helper()
	body := message.Body
	header := &wireHeader{
		Version:   1,
		Method:    byte(message.Method),
		SwitchMac: [6]byte(message.SwitchMac),
		ClientMac: [6]byte(message.ClientMac),
		RequestId: message.RequestId,
		Status:    uint32(message.State),
		Length:    uint16(32 + len(body)),
		Fragment:  message.Fragment,
		Flags:     0,
		TokenId:   keyState.TokenId,
		CRC:       0,
	}

	header.Version = 1
	if packet, err := keyState.encode(*header, body); err != nil {
		t.Fatalf("encode() error = %v", err)
	} else {
		return packet
	}
	return nil
}

func TestClientResolveTargetWithDefaultKey(t *testing.T) {
	client, err := NewClient(net.Interface{}, 0, []byte("default-key"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	defaultState, _ := client.States.Get(managerTestDefaultMac)
	message := Message{
		Method:    MethodDiscover,
		SwitchMac: MACAddress{1, 2, 3, 4, 5, 6},
		ClientMac: MACAddress{6, 5, 4, 3, 2, 1},
		RequestId: 17,
		Body:      []byte("value"),
	}

	target, got, err := client.resolveTarget(encodeTestMessage(t, defaultState, message))
	if err != nil {
		t.Fatalf("resolveTarget() error = %v", err)
	}
	if target == nil {
		t.Fatal("resolveTarget() target is nil")
	}
	if target.Encryption != defaultState.Encryption {
		t.Error("resolved target does not have the default encryption key")
	}
	if got.RequestId != message.RequestId || got.Method != message.Method {
		t.Errorf("resolved message = (%v, %v), want (%v, %v)", got.RequestId, got.Method, message.RequestId, message.Method)
	}
	if len(got.Body) != 5 || string(got.Body) != string(message.Body) {
		t.Errorf("resolved body = %#v, want %#v", got.Body, message.Body)
	}
	if registered, ok := client.States.Get([6]byte(message.SwitchMac)); !ok || registered != target {
		t.Error("resolved target was not registered in the state table")
	}
}

func TestClientResolveTargetWithPeerKey(t *testing.T) {
	client, err := NewClient(net.Interface{}, 0, []byte("default-key"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	peerMac := [6]byte{9, 8, 7, 6, 5, 4}
	peerKey, err := NewStateTable([]byte("peer-key"))
	if err != nil {
		t.Fatalf("NewStateTable() error = %v", err)
	}
	peerState, _ := peerKey.Get(managerTestDefaultMac)
	client.States.Set(peerMac, peerState)

	message := Message{
		Method:    MethodPostResponse,
		SwitchMac: MACAddress{4, 3, 2, 1, 0, 9},
		ClientMac: MACAddress{1, 3, 5, 7, 9, 2},
		RequestId: 29,
	}
	packet := encodeTestMessage(t, peerState, message)

	target, got, err := client.resolveTarget(packet)
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
	if registered, ok := client.States.Get([6]byte(message.SwitchMac)); !ok || registered != target {
		t.Error("discovered peer target was not registered in the state table")
	}
}

func TestClientConnectSendAndHandle(t *testing.T) {
	client, err := NewClient(net.Interface{}, 0, []byte("client-send-key"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	conn := newTestConn(t)

	client.onConnect(conn)
	message := Message{
		Method:    MethodGetRequest,
		SwitchMac: MACAddress{1, 2, 3, 4, 5, 6},
		ClientMac: MACAddress{6, 5, 4, 3, 2, 1},
		RequestId: 104,
	}
	handlerCalled := false
	handler := func(got *Message, addr *net.UDPAddr) bool {
		handlerCalled = got != nil && got.RequestId == message.RequestId && addr != nil
		return true
	}
	if err := client.Send(message, testPeerAddr, handler, time.Hour); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	defaultState, _ := client.States.Get([6]byte{})
	assertReceivedMessage(t, conn, defaultState, message)

	if registered, ok := client.ResponseHandlers.Get(message.RequestId); !ok || registered == nil {
		t.Fatal("Send() did not register the response handler")
	}
	client.handle(&message, testPeerAddr)
	if !handlerCalled {
		t.Error("handle() did not call the registered response handler")
	}
	if _, ok := client.ResponseHandlers.Get(message.RequestId); ok {
		t.Error("handle() did not remove the response handler after it requested drop")
	}

	client.onClose()
	if err := client.Send(message, testPeerAddr, nil, time.Second); !errors.Is(err, ErrNotConnected) {
		t.Errorf("Send() after onClose error = %v, want ErrNotConnected", err)
	}
}

func TestClientSendTimeoutCallsHandler(t *testing.T) {
	client, err := NewClient(net.Interface{}, 0, []byte("client-timeout-key"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	conn := newTestConn(t)
	client.onConnect(conn)

	timedOut := make(chan bool, 1)
	message := Message{
		Method:    MethodDiscover,
		SwitchMac: MACAddress{1, 2, 3, 4, 5, 6},
		ClientMac: MACAddress{6, 5, 4, 3, 2, 1},
		RequestId: 105,
	}
	if err := client.Send(message, testPeerAddr, func(got *Message, addr *net.UDPAddr) bool {
		timedOut <- got == nil && addr == nil
		return true
	}, 10*time.Millisecond); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	select {
	case got := <-timedOut:
		if !got {
			t.Error("timeout handler received a message or address, want nil values")
		}
	case <-time.After(time.Second):
		t.Fatal("response timeout did not invoke the handler")
	}
}

func TestClientHandleKeepsHandlerWhenDropIsFalse(t *testing.T) {
	client, err := NewClient(net.Interface{}, 0, []byte("client-handler-key"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	message := Message{RequestId: 106}
	called := false
	client.ResponseHandlers.Set(message.RequestId, func(got *Message, addr *net.UDPAddr) bool {
		called = got == &message && addr != nil
		return false
	})

	client.handle(&message, &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 4321})
	if !called {
		t.Error("handle() did not call the registered handler")
	}
	if _, ok := client.ResponseHandlers.Get(message.RequestId); !ok {
		t.Error("handle() removed the handler when it returned false")
	}

	client.handle(&Message{RequestId: 999}, &net.UDPAddr{})
}

func TestClientSendFailureRemovesHandler(t *testing.T) {
	client, err := NewClient(net.Interface{}, 0, []byte("client-send-failure-key"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	conn := newTestConn(t)
	client.onConnect(conn)
	conn.Close()

	called := make(chan struct{}, 1)
	message := Message{
		SwitchMac: MACAddress{1, 2, 3, 4, 5, 6},
		ClientMac: MACAddress{6, 5, 4, 3, 2, 1},
		RequestId: 107,
	}
	err = client.Send(message, testPeerAddr, func(*Message, *net.UDPAddr) bool {
		called <- struct{}{}
		return true
	}, 20*time.Millisecond)
	if err == nil {
		t.Fatal("Send() succeeded on a closed connection, want an error")
	}
	if _, ok := client.ResponseHandlers.Get(message.RequestId); ok {
		t.Error("Send() left the response handler registered after a failed send")
	}
	select {
	case <-called:
		t.Error("response handler was called after a failed send")
	case <-time.After(100 * time.Millisecond):
	}
}

func TestClientConnectCloseSendIsRaceFree(t *testing.T) {
	client, err := NewClient(net.Interface{}, 0, []byte("client-race-key"))
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}
	conn := newTestConn(t)
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < 100; i++ {
			client.onConnect(conn)
			client.onClose()
		}
	}()
	for i := 0; i < 100; i++ {
		_ = client.Send(Message{}, testPeerAddr, nil, time.Millisecond)
	}
	<-done
}
