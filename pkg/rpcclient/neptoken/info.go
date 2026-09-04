package aeptoken

import (
	"fmt"

	"github.com/atipicial/atipicial-go/pkg/core/state"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/invoker"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/atipicial/atipicial-go/pkg/wallet"
)

// InfoClient is a set of RPC methods required to get all of the AEP-11/AEP-17
// token data.
type InfoClient interface {
	invoker.RPCInvoke

	GetContractStateByHash(hash util.Uint160) (*state.Contract, error)
}

// Info allows to get basic token info using RPC client.
func Info(c InfoClient, hash util.Uint160) (*wallet.Token, error) {
	cs, err := c.GetContractStateByHash(hash)
	if err != nil {
		return nil, err
	}
	var standard string
	for _, st := range cs.Manifest.SupportedStandards {
		if st == manifest.AEP17StandardName || st == manifest.AEP11StandardName {
			standard = st
			break
		}
	}
	if standard == "" {
		return nil, fmt.Errorf("contract %s is not AEP-11/AEP17", hash.StringLE())
	}
	b := New(invoker.New(c, nil), hash)
	symbol, err := b.Symbol()
	if err != nil {
		return nil, err
	}
	decimals, err := b.Decimals()
	if err != nil {
		return nil, err
	}
	return wallet.NewToken(hash, cs.Manifest.Name, symbol, int64(decimals), standard), nil
}
