package wasmsdk

import (
	"github.com/ava-labs/hypersdk/codec"
)

type Base struct {
	Timestamp int64    `json:"timestamp"`
	ChainID   [32]byte `json:"chainId"`
	MaxFee    uint64   `json:"maxFee"`
}

func (b *Base) Marshal(p *codec.Packer) {
	p.PackInt64(b.Timestamp)
	p.PackID(b.ChainID)
	p.PackUint64(b.MaxFee)
}

func Digest(base Base, actions []CompactAction) ([]byte, error) {
	p := codec.NewWriter(0, 1024)
	base.Marshal(p)

	p.PackByte(uint8(len(actions)))
	for _, action := range actions {
		p.PackByte(action.GetTypeID())
		action.Marshal(p)
	}
	return p.Bytes(), p.Err()
}
