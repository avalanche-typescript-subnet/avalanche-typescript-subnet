package wasmsdk

import (
	"github.com/ava-labs/hypersdk/codec"
)

type Transaction struct {
	Base *Base `json:"base"`

	Actions []CompactAction `json:"actions"`
	// Auth    Auth            `json:"auth"`

	// bytes     []byte
	// size      int
	// id        ids.ID
	// stateKeys state.Keys
}

func (t *Transaction) Digest() ([]byte, error) {
	p := codec.NewWriter(0, 1024)
	t.Base.Marshal(p)

	p.PackByte(uint8(len(t.Actions)))
	for _, action := range t.Actions {
		p.PackByte(action.GetTypeID())
		action.Marshal(p)
	}
	return p.Bytes(), p.Err()
}
