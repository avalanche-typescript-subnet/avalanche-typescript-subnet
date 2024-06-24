package wasmsdk

import (
	"github.com/ava-labs/hypersdk/codec"
)

type Base struct {
	Timestamp int64    `json:"timestamp"`
	ChainID   [32]byte `json:"chainId"`
	MaxFee    uint64   `json:"maxFee"`
}

func (b *Base) MarshalPacker(p *codec.Packer) {
	p.PackInt64(b.Timestamp)
	p.PackID(b.ChainID)
	p.PackUint64(b.MaxFee)
}

func Digest(base *Base) ([]byte, error) {
	p := codec.NewWriter(0, 1024)
	base.MarshalPacker(p)
	p.PackByte(0)

	return p.Bytes(), p.Err()
}
