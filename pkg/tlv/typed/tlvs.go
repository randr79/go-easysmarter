package typed

import (
	"encoding/json"
	"errors"

	"github.com/randr79/go-easysmarter/pkg/tlv"
)

type TLV interface {
	Type() Type
	Encode() ([]byte, error)
	Decode([]byte) error
}

type TLVs []TLV

func (t TLVs) Encoded() ([]tlv.TLV, error) {
	var errs []error
	res := make([]tlv.TLV, len(t))
	for i, j := range t {

		if v, err := j.Encode(); err != nil {
			errs = append(errs, err)
		} else {
			res[i] = tlv.TLV{
				Type:  uint16(j.Type()),
				Value: v,
			}
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	} else {
		return res, nil
	}

}

func NewTLVs(tlvs []tlv.TLV) (TLVs, error) {
	var errs []error
	res := make([]TLV, len(tlvs))

	for i, j := range tlvs {
		if t, err := TLVByString(Type(j.Type).String()); err != nil {
			errs = append(errs, err)
		} else if err := t.Decode(j.Value); err != nil {
			errs = append(errs, err)
		} else {
			res[i] = t
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	} else {
		return res, nil
	}
}

type namedTLV struct {
	Type  string
	Value json.RawMessage
}

func (t TLVs) MarshalJSON() ([]byte, error) {
	var errs []error
	tmp := make([]namedTLV, len(t))
	for i, j := range t {
		if v, err := json.Marshal(j); err != nil {
			errs = append(errs, err)
		} else {
			tmp[i] = namedTLV{
				Type:  j.Type().String(),
				Value: v,
			}
		}
	}
	if len(errs) > 0 {
		return nil, errors.Join(errs...)
	}

	return json.Marshal(tmp)
}
func (t *TLVs) UnMarshalJSON(data []byte) error {
	items := make([]namedTLV, 0)
	if err := json.Unmarshal(data, &items); err != nil {
		return err
	}
	tmp := make([]TLV, len(items))
	var errs []error

	for k, v := range items {
		if instance, err := TLVByString(v.Type); err != nil {
			errs = append(errs, err)
		} else if err := json.Unmarshal(v.Value, instance); err != nil {
			errs = append(errs, err)
		} else {
			tmp[k] = instance
		}
	}
	if len(errs) > 0 {
		return errors.Join(errs...)
	}

	*t = TLVs(tmp)
	return nil

}
