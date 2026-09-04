/*
Package aep22 contains RPC wrappers to work with AEP-22 contracts.

Contract provides state-changing method of upgradeable contracts.
*/
package aep22

import (
	"github.com/atipicial/atipicial-go/pkg/core/transaction"
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/util"
)

// Actor is used by Contract wrapper to create and send transactions.
type Actor interface {
	MakeRun(script []byte) (*transaction.Transaction, error)
	MakeUnsignedRun(script []byte, attrs []transaction.Attribute) (*transaction.Transaction, error)
	SendRun(script []byte) (util.Uint256, uint32, error)
}

// Contract is an RPC wrapper that implements the full AEP-22 interface.
type Contract struct {
	hash  util.Uint160
	actor Actor
}

// NewContract creates an instance of Contract for contract with the given hash
// using the given Actor.
func NewContract(actor Actor, hash util.Uint160) *Contract {
	return &Contract{
		hash:  hash,
		actor: actor,
	}
}

// Update creates and sends a transaction that performs an `update` method
// call using the given parameters. This transaction is signed and immediately
// sent to the network. The returned values are its hash, ValidUntilBlock value
// and an error, if any.
func (t *Contract) Update(aef []byte, manifest []byte, data any) (util.Uint256, uint32, error) {
	script, err := t.updateScript(aef, manifest, data)
	if err != nil {
		return util.Uint256{}, 0, err
	}
	return t.actor.SendRun(script)
}

// UpdateTransaction creates and signs a transaction invoking the `update` method
// with the provided AEF bytes, manifest and arbitrary data,
// without sending it. The returned values are the signed Transaction and an error
// if any.
func (t *Contract) UpdateTransaction(aef []byte, manifest []byte, data any) (*transaction.Transaction, error) {
	script, err := t.updateScript(aef, manifest, data)
	if err != nil {
		return nil, err
	}
	return t.actor.MakeRun(script)
}

// UpdateUnsigned creates an unsigned transaction invoking the `update` method
// with the provided AEF bytes, manifest and arbitrary data payload,
// without sending it. The returned values are the unsigned Transaction and
// an error if any.
func (t *Contract) UpdateUnsigned(aef []byte, manifest []byte, data any) (*transaction.Transaction, error) {
	script, err := t.updateScript(aef, manifest, data)
	if err != nil {
		return nil, err
	}
	return t.actor.MakeUnsignedRun(script, nil)
}

// updateScript builds a VM script invoking the `update` method of the contract
// using the given parameters. The returned values are the raw script
// and an error, if any.
func (t *Contract) updateScript(aef []byte, manifest []byte, data any) ([]byte, error) {
	b := smartcontract.NewBuilder()
	b.InvokeMethod(t.hash, "update", aef, manifest, data)
	return b.Script()
}
