package observer

import (
	"context"
	"errors"
	"net"

	"github.com/randr79/go-easysmarter/pkg/tlv"
	"github.com/randr79/go-easysmarter/pkg/transport"
)

type MessageTarget uint

const (
	MessageTargetToSwitch  = 0
	MessageTargetToManager = 1
)

type ObserverHandler func(header transport.Header, tlvs []tlv.TLV, address *net.UDPAddr, source MessageTarget)

type statetable interface {
	Get(key transport.MACAddress) (*transport.State, bool)
	Keys() []transport.MACAddress
}
type Observer struct {
	states      statetable
	partials    map[uint16]*tlv.FragmentList
	switchLink  *transport.Observer
	managerLink *transport.Observer
	handler     ObserverHandler
}

func (o *Observer) Observe(ctx context.Context, handler ObserverHandler) error {
	var serr error
	var merr error
	o.handler = handler
	go func() { serr = o.switchLink.Listen(ctx) }()
	go func() { merr = o.managerLink.Listen(ctx) }()
	<-ctx.Done()
	o.handler = nil
	return errors.Join(serr, merr)
}
func (o *Observer) messageForSwitch(msg *transport.Message, address *net.UDPAddr) {
	if msg == nil {
		return
	} else if o.handler != nil {
		tlvs, _ := tlv.Parse(msg.Body)
		o.handler(msg.Header, tlvs, address, MessageTargetToSwitch)
	}
}
func (o *Observer) messageForManager(msg *transport.Message, address *net.UDPAddr) {
	if msg == nil {
		return
	} else if o.handler != nil {
		tlvs, _ := tlv.Parse(msg.Body)
		o.handler(msg.Header, tlvs, address, MessageTargetToManager)
	}
}

func NewObserver(iface net.Interface, defaultCryptKey []byte) (*Observer, error) {
	obs := Observer{
		partials: make(map[uint16]*tlv.FragmentList),
	}
	links := []transport.SharedStateLink{
		{
			Interface: iface,
			Port:      transport.PortManager,
			Handler:   obs.messageForManager,
		}, {
			Interface: iface,
			Port:      transport.PortSwitch,
			Handler:   obs.messageForSwitch,
		},
	}
	if observers, err := transport.NewSharedStateObservers(links, defaultCryptKey); err != nil {
		return nil, err
	} else {
		obs.states = observers[0].States
		obs.managerLink = observers[0]
		obs.switchLink = observers[1]
	}
	return &obs, nil
}
