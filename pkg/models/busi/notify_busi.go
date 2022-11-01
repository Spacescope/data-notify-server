package busi

import "github.com/filecoin-project/lotus/chain/types"

type Message struct {
	Version int
	Tipset  types.TipSet
}
