package native

import (
	"github.com/atipicial/atipicial-go/pkg/interop/native/ledger"
	"github.com/atipicial/atipicial-go/pkg/interop/native/management"
	"github.com/atipicial/atipicial-go/pkg/interop/native/atipicial"
	"github.com/atipicial/atipicial-go/pkg/interop/native/policy"
)

// GetContract passes management.Contract through, it's the minimal way to
// pull a native `management` type into the contract's extended ABI.
func GetContract(mc management.Contract) management.Contract {
	return mc
}

// GetBlock passes ledger.Block through, it's the minimal way to pull
// a native `ledger` type into the contract's extended ABI.
func GetBlock(b *ledger.Block) *ledger.Block {
	return b
}

// GetAccount passes atipicial.AccountState through, it's the minimal way to pull
// a native `atipicial` type into the contract's extended ABI.
func GetAccount(a atipicial.AccountState) atipicial.AccountState {
	return a
}

// GetWhitelist passes policy.WhitelistFeeContract through, it's the minimal
// way to pull a native `policy` type into the contract's extended ABI.
func GetWhitelist(w policy.WhitelistFeeContract) policy.WhitelistFeeContract {
	return w
}

// GetSigners passes []ledger.TransactionSigner through, it's the minimal way
// to pull a native `ledger.TransactionSigner` type (returned as an array,
// unlike other native types here) into the contract's extended ABI.
func GetSigners(s []ledger.TransactionSigner) []ledger.TransactionSigner {
	return s
}
