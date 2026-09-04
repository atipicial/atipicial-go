package ledger

import "github.com/atipicial/atipicial-go/pkg/interop"

// Transaction represents a ATC transaction. It's similar to Transaction class
// in Atipicial .net framework.
type Transaction struct {
	// Hash represents the hash (256 bit BE value in a 32 byte slice) of the
	// given transaction (which also is its ID).
	Hash interop.Hash256
	// Version represents the transaction version.
	Version int
	// Nonce is a random number to avoid hash collision.
	Nonce int
	// Sender represents the sender (160 bit BE value in a 20 byte slice) of the
	// given Transaction.
	Sender interop.Hash160
	// SysFee represents the fee to be burned.
	SysFee int
	// NetFee represents the fee to be distributed to consensus nodes.
	NetFee int
	// ValidUntilBlock is the maximum blockchain height exceeding which
	// a transaction should fail verification.
	ValidUntilBlock int
	// Script represents a code to run in AtipicialVM for this transaction.
	Script []byte
}
