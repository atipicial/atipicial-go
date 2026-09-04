package aep17_test

import (
	"context"
	"math/big"

	"github.com/atipicial/atipicial-go/pkg/encoding/address"
	"github.com/atipicial/atipicial-go/pkg/rpcclient"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/actor"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/invoker"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/aep17"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/atipicial/atipicial-go/pkg/wallet"
)

func ExampleTokenReader() {
	// No error checking done at all, intentionally.
	c, _ := rpcclient.New(context.Background(), "url", rpcclient.Options{})

	// Safe methods are reachable with just an invoker, no need for an account there.
	inv := invoker.New(c, nil)

	// AEP-17 contract hash.
	aep17Hash := util.Uint160{9, 8, 7}

	// And a reader interface.
	n17 := aep17.NewReader(inv, aep17Hash)

	// Get the metadata. Even though these methods are implemented in aeptoken package,
	// they're available for AEP-17 wrappers.
	symbol, _ := n17.Symbol()
	supply, _ := n17.TotalSupply()
	_ = symbol
	_ = supply

	// Account hash we're interested in.
	accHash, _ := address.StringToUint160("NdypBhqkz2CMMnwxBgvoC9X2XjKF5axgKo")

	// Get account balance.
	balance, _ := n17.BalanceOf(accHash)
	_ = balance
}

func ExampleToken() {
	// No error checking done at all, intentionally.
	w, _ := wallet.NewWalletFromFile("somewhere")
	defer w.Close()

	c, _ := rpcclient.New(context.Background(), "url", rpcclient.Options{})

	// Create a simple CalledByEntry-scoped actor (assuming there is an account
	// inside the wallet).
	a, _ := actor.NewSimple(c, w.Accounts[0])

	// AEP-17 contract hash.
	aep17Hash := util.Uint160{9, 8, 7}

	// Create a complete AEP-17 contract representation.
	n17 := aep17.New(a, aep17Hash)

	tgtAcc, _ := address.StringToUint160("NdypBhqkz2CMMnwxBgvoC9X2XjKF5axgKo")

	// Send a transaction that transfers one token to another account.
	txid, vub, _ := n17.Transfer(a.Sender(), tgtAcc, big.NewInt(1), nil)
	_ = txid
	_ = vub
}
