/*
Package gas provides a convenience wrapper for GAS contract to use it via RPC.

GAS itself only has standard AEP-17 methods, so this package only contains its
hash and allows to create AEP-17 structures in an easier way. Refer to [aep17]
package for more details on AEP-17 interface.
*/
package gas

import (
	"github.com/atipicial/atipicial-go/pkg/core/native/nativehashes"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/aep17"
)

// Hash stores the hash of the native GAS contract.
var Hash = nativehashes.AtipicialDollar

// NewReader creates a AEP-17 reader for the GAS contract.
func NewReader(invoker aep17.Invoker) *aep17.TokenReader {
	return aep17.NewReader(invoker, Hash)
}

// New creates a AEP-17 contract instance for the native GAS contract.
func New(actor aep17.Actor) *aep17.Token {
	return aep17.New(actor, Hash)
}
