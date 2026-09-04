package testdata

import (
	"github.com/atipicial/atipicial-go/pkg/interop"
	"github.com/atipicial/atipicial-go/pkg/interop/runtime"
)

func Verify() bool {
	return true
}

func OnAEP17Payment(from interop.Hash160, amount int, data any) {
}

// OnAEP11Payment notifies about AEP-11 payment. You don't call this method directly,
// instead it's called by AEP-11 contract when you transfer funds from your address
// to the address of this NFT contract.
func OnAEP11Payment(from interop.Hash160, amount int, tokenId []byte, data any) {
	runtime.Notify("OnAEP11Payment", from, amount, tokenId, data)
}
