package chain

import (
	"fmt"

	"github.com/ava-labs/avalanchego/ids"
	"github.com/ava-labs/hypersdk/codec"
	"github.com/ava-labs/hypersdk/consts"
	"github.com/ava-labs/hypersdk/utils"
)

type Chunk struct {
	Expiry int64 `json:"expiry"`

	Txs []*Transaction `json:"txs"`

	size int
	id   ids.ID
}

func (c *Chunk) ID() (ids.ID, error) {
	if c.id != ids.Empty {
		return c.id, nil
	}

	bytes, err := c.Marshal()
	if err != nil {
		return ids.ID{}, err
	}
	c.id = utils.ToID(bytes)
	return c.id, nil
}

func (c *Chunk) Size() int {
	return c.size
}

func (c *Chunk) Marshal() ([]byte, error) {
	size := consts.Uint64Len + consts.IntLen + codec.CummSize(c.Txs)
	p := codec.NewWriter(size, consts.NetworkSizeLimit)

	p.PackInt64(c.Expiry)
	p.PackInt(len(c.Txs))
	for _, tx := range c.Txs {
		if err := tx.Marshal(p); err != nil {
			return nil, err
		}
	}
	bytes := p.Bytes()
	if err := p.Err(); err != nil {
		return nil, err
	}
	c.size = len(bytes)
	return bytes, nil
}

func UnmarshalChunk(raw []byte, parser Parser) (*Chunk, error) {
	var (
		actionRegistry, authRegistry = parser.Registry()
		p                            = codec.NewReader(raw, consts.NetworkSizeLimit)
		c                            Chunk
	)
	c.id = utils.ToID(raw)
	c.size = len(raw)
	c.Expiry = p.UnpackInt64(false)

	// Parse transactions
	txCount := p.UnpackInt(false) // can produce empty blocks
	c.Txs = []*Transaction{}      // don't preallocate all to avoid DoS
	for i := 0; i < txCount; i++ {
		tx, err := UnmarshalTx(p, actionRegistry, authRegistry)
		if err != nil {
			return nil, err
		}
		c.Txs = append(c.Txs, tx)
	}

	// Ensure no leftover bytes
	if !p.Empty() {
		return nil, fmt.Errorf("%w: remaining=%d", ErrInvalidObject, len(raw)-p.Offset())
	}
	return &c, p.Err()
}
