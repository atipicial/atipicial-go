package aep24_test

import (
	"context"
	"math/big"

	"github.com/atipicial/atipicial-go/pkg/rpcclient"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/invoker"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/aep24"
	"github.com/atipicial/atipicial-go/pkg/util"
)

func ExampleRoyaltyReader() {
	// No error checking done at all, intentionally.
	c, _ := rpcclient.New(context.Background(), "url", rpcclient.Options{})

	// Safe methods are reachable with just an invoker, no need for an account there.
	inv := invoker.New(c, nil)

	// AEP-24 contract hash.
	aep24Hash := util.Uint160{9, 8, 7}

	// And a reader interface.
	n24 := aep24.NewRoyaltyReader(inv, aep24Hash)

	// Get the royalty information for a token.
	tokenID := []byte("someTokenID")
	royaltyToken := util.Uint160{1, 2, 3}
	salePrice := big.NewInt(1000)
	royaltyInfo, _ := n24.RoyaltyInfo(tokenID, royaltyToken, salePrice)
	_ = royaltyInfo
}
