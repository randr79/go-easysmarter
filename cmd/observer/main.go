package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/randr79/go-easysmarter/pkg/observer"
	"github.com/randr79/go-easysmarter/pkg/tlv"
	"github.com/randr79/go-easysmarter/pkg/tlv/typed"
	"github.com/randr79/go-easysmarter/pkg/transport"
)

func main() {
	defer func() {
		if e := recover(); e == nil {
			os.Exit(0)
		} else if err, ok := e.(error); ok {
			os.Stderr.Write([]byte(err.Error()))
		} else if err, ok := e.(string); ok {
			os.Stderr.Write([]byte(err))
		} else {
			fmt.Fprintf(os.Stderr, "%#v", e)
		}

		os.Exit(1)
	}()
	if err := run(); err != nil {
		os.Exit(2)
	}
}
func run() error {
	filename := "/etc/unrefi/easysmarter.json"
	keys := make(map[string][]byte)
	if c, err := os.ReadFile(filename); err != nil {
		return err
	} else if err := json.Unmarshal(c, &keys); err != nil {
		return err
	}
	var obs *observer.Observer

	if iface, err := net.InterfaceByName("enX0"); err != nil {
		panic(err)
	} else if obs, err = observer.NewObserver(*iface, []byte(keys["Rc4DefaultKey"])); err != nil {
		panic(err)
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "    ")
	obs.Observe(context.Background(), func(head transport.Header, tlvs []tlv.TLV, address *net.UDPAddr, source observer.MessageTarget) {
		enc.Encode(head)
		t, err := typed.NewTLVs(tlvs)
		if err != nil {
			print(err.Error())
		}
		enc.Encode(t)
	})
	return nil
}
