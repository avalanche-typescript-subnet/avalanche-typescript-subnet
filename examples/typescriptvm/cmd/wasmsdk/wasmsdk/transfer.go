package wasmsdk

import (
	"github.com/ava-labs/hypersdk/codec"
)

type CompactAction interface {
	GetTypeID() uint8
	Marshal(p *codec.Packer)
}

var _ CompactAction = (*Transfer)(nil)

type Transfer struct {
	To    codec.Address `json:"to"`
	Value uint64        `json:"value"`
}

func (*Transfer) GetTypeID() uint8 {
	return 0
}

func (t *Transfer) Marshal(p *codec.Packer) {
	p.PackAddress(t.To)
	p.PackUint64(t.Value)
}
