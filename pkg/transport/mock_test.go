package transport

import (
	"net"
	"sync"
	"testing"
)

type mockPacket struct {
	data []byte
	addr *net.UDPAddr
}

// mockUDPConnection is an in-memory UDPConn: packets queued with inject are
// returned by ReadFromUDP and packets passed to WriteToUDP are recorded.
type mockUDPConnection struct {
	incoming chan mockPacket
	closed   chan struct{}
	once     sync.Once

	mu      sync.Mutex
	written []mockPacket
}

func newMockUDPConnection() *mockUDPConnection {
	return &mockUDPConnection{
		incoming: make(chan mockPacket, 16),
		closed:   make(chan struct{}),
	}
}

func (m *mockUDPConnection) inject(data []byte, from *net.UDPAddr) {
	m.incoming <- mockPacket{data: data, addr: from}
}

func (m *mockUDPConnection) Written() []mockPacket {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]mockPacket(nil), m.written...)
}

func (m *mockUDPConnection) ReadFromUDP(b []byte) (int, *net.UDPAddr, error) {
	select {
	case <-m.closed:
		return 0, nil, net.ErrClosed
	case p := <-m.incoming:
		return copy(b, p.data), p.addr, nil
	}
}

func (m *mockUDPConnection) WriteToUDP(b []byte, addr *net.UDPAddr) (int, error) {
	select {
	case <-m.closed:
		return 0, net.ErrClosed
	default:
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	m.written = append(m.written, mockPacket{data: append([]byte(nil), b...), addr: addr})
	return len(b), nil
}

func (m *mockUDPConnection) Close() error {
	m.once.Do(func() { close(m.closed) })
	return nil
}

// mockConnectionFactory hands out a single prepared mockUDPConnection.
type mockConnectionFactory struct {
	conn *mockUDPConnection
	err  error
}

func newMockConnectionFactory() *mockConnectionFactory {
	return &mockConnectionFactory{conn: newMockUDPConnection()}
}

func (f *mockConnectionFactory) create(uint16, net.Interface) (UDPConn, error) {
	if f.err != nil {
		return nil, f.err
	}
	return f.conn, nil
}

var testPeerAddr = &net.UDPAddr{IP: net.IPv4(127, 0, 0, 1), Port: 4000}

func newTestConn(t *testing.T) *mockUDPConnection {
	t.Helper()
	conn := newMockUDPConnection()
	t.Cleanup(func() { conn.Close() })
	return conn
}

func assertReceivedMessage(t *testing.T, conn *mockUDPConnection, keyState *State, want Message) {
	t.Helper()
	written := conn.Written()
	if len(written) == 0 {
		t.Fatal("no packet was written to the connection")
	}
	buffer := written[len(written)-1].data
	n := len(buffer)
	header, body, err := keyState.decode(buffer[:n])
	if err != nil {
		t.Fatalf("decode() sent packet error = %v", err)
	}
	if header == nil {
		t.Fatalf("sent packet has an invalid header: decrypted bytes %x", keyState.Encryption.XORKey(buffer[:n]))
	}

	if header.RequestId != want.RequestId || Method(header.Method) != want.Method || header.Fragment != want.Fragment ||
		header.SwitchMac != want.SwitchMac || header.ClientMac != want.ClientMac {
		t.Errorf("sent message = %#v, want %#v", header, want)
	}
	if header.Status != uint32(want.State) {
		t.Errorf("sent status = %d, want %d", header.Status, want.State)
	}
	if string(body) != string(want.Body) {
		t.Errorf("sent body = %q, want %q", body, want.Body)
	}
}
