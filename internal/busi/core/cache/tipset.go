package cache

import (
	"errors"
	"sync"

	"github.com/filecoin-project/lotus/chain/types"
)

var (
	ErrCacheEmpty = errors.New("cache empty")
)

// TipSetCache is a cache of recent tipsets that can keep track of reversions.
// Inspired by tipSetCache in Lotus chain/events package.
type TipSetCache struct {
	buffer []*types.TipSet
	mu     sync.Mutex
}

func NewTipSetCache() *TipSetCache {
	return &TipSetCache{
		buffer: make([]*types.TipSet, 0),
	}
}

func (c *TipSetCache) PopAll() ([]*types.TipSet, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.buffer) == 0 {
		return nil, ErrCacheEmpty
	}

	p := c.buffer
	c.buffer = make([]*types.TipSet, 0)

	return p, nil
}

func (c *TipSetCache) Add(ts *types.TipSet) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if len(c.buffer) == 0 {
		c.buffer = append(c.buffer, ts)
		// log.Infof("Notify emitted tipset: %v", event.Val.Height())
	} else {
		if ts.Height() != c.buffer[len(c.buffer)-1].Height() { // remove-duplicates-from-sorted-array
			c.buffer = append(c.buffer, ts)
			// log.Infof("Notify emitted tipset: %v", event.Val.Height())
		}
	}
}
