/*
Package ledger provides an interface to LedgerContract native contract.
It allows to access ledger contents like transactions and blocks.
*/
package ledger

import (
	"github.com/atipicial/atipicial-go/pkg/interop"
	"github.com/atipicial/atipicial-go/pkg/interop/contract"
	"github.com/atipicial/atipicial-go/pkg/interop/atipicialgointernal"
)

// Hash represents Ledger contract hash.
const Hash = "\xbe\xf2\x04\x31\x40\x36\x2a\x77\xc1\x50\x99\xc7\xe6\x4c\x12\xf7\x00\xb6\x65\xda"

// VMState represents VM execution state.
type VMState uint8

// Various VM execution states.
const (
	// NoneState represents NONE VM state.
	NoneState VMState = 0
	// HaltState represents HALT VM state.
	HaltState VMState = 1
	// FaultState represents FAULT VM state.
	FaultState VMState = 2
	// BreakState represents BREAK VM state.
	BreakState VMState = 4
)

// CurrentHash represents `currentHash` method of Ledger native contract.
func CurrentHash() interop.Hash256 {
	return atipicialgointernal.CallWithToken(Hash, "currentHash", int(contract.ReadStates)).(interop.Hash256)
}

// CurrentIndex represents `currentIndex` method of Ledger native contract.
func CurrentIndex() int {
	return atipicialgointernal.CallWithToken(Hash, "currentIndex", int(contract.ReadStates)).(int)
}

// GetBlock represents `getBlock` method of Ledger native contract.
func GetBlock(indexOrHash any) *Block {
	return atipicialgointernal.CallWithToken(Hash, "getBlock", int(contract.ReadStates), indexOrHash).(*Block)
}

// GetTransaction represents `getTransaction` method of Ledger native contract.
func GetTransaction(hash interop.Hash256) *Transaction {
	return atipicialgointernal.CallWithToken(Hash, "getTransaction", int(contract.ReadStates), hash).(*Transaction)
}

// GetTransactionHeight represents `getTransactionHeight` method of Ledger native contract.
func GetTransactionHeight(hash interop.Hash256) int {
	return atipicialgointernal.CallWithToken(Hash, "getTransactionHeight", int(contract.ReadStates), hash).(int)
}

// GetTransactionFromBlock represents `getTransactionFromBlock` method of Ledger native contract.
func GetTransactionFromBlock(indexOrHash any, txIndex int) *Transaction {
	return atipicialgointernal.CallWithToken(Hash, "getTransactionFromBlock", int(contract.ReadStates),
		indexOrHash, txIndex).(*Transaction)
}

// GetTransactionSigners represents `getTransactionSigners` method of Ledger native contract.
func GetTransactionSigners(hash interop.Hash256) []TransactionSigner {
	return atipicialgointernal.CallWithToken(Hash, "getTransactionSigners", int(contract.ReadStates),
		hash).([]TransactionSigner)
}

// GetTransactionVMState represents `getTransactionVMState` method of Ledger native contract.
func GetTransactionVMState(hash interop.Hash256) VMState {
	return atipicialgointernal.CallWithToken(Hash, "getTransactionVMState", int(contract.ReadStates), hash).(VMState)
}
