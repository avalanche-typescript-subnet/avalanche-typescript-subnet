package wasmsdk

import (
	"github.com/ava-labs/hypersdk/codec"
)

type Transfer struct {
	// To is the recipient of the [Value].
	To codec.Address `json:"to"`

	// Amount are transferred to [To].
	Value uint64 `json:"value"`
}

func (*Transfer) GetTypeID() uint8 {
	return 0
}

func (t *Transfer) Marshal(p *codec.Packer) {
	p.PackAddress(t.To)
	p.PackUint64(t.Value)
}
