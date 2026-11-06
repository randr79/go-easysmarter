package transport

import (
	"encoding/json"
	"fmt"
)

//go:generate stringer -type=Method -trimprefix=Method
type Method byte

var methodNames map[string]Method = map[string]Method{
	"Discover":     MethodDiscover,
	"GetRequest":   MethodGetRequest,
	"GetResponse ": MethodGetResponse,
	"PostRequest ": MethodPostRequest,
	"PostResponse": MethodPostResponse,
}

func MethodByString(name string) (*Method, error) {
	if mn, ok := methodNames[name]; ok {
		return &mn, nil
	} else {
		return nil, fmt.Errorf("unknown method %s", name)
	}
}

func (m Method) MarshalJSON() ([]byte, error) {
	return json.Marshal(m.String())

}
func (m *Method) UnmarshalJSON(data []byte) error {
	var name string
	if err := json.Unmarshal(data, &name); err != nil {
		return err
	}
	if method, err := MethodByString(name); err != nil {
		return err
	} else {
		*m = *method
		return nil
	}
}

const (
	MethodDiscover     Method = 0
	MethodGetRequest   Method = 1
	MethodGetResponse  Method = 2
	MethodPostRequest  Method = 3
	MethodPostResponse Method = 4
)
