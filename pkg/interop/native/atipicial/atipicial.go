/*
Package atipicial provides an interface to AtipicialCoin native contract.
ATC token is special, it's not just a regular AEP-17 contract, it also
provides access to chain-specific settings and implements committee
voting system.
*/
package atipicial

import (
	"github.com/atipicial/atipicial-go/pkg/interop"
	"github.com/atipicial/atipicial-go/pkg/interop/contract"
	"github.com/atipicial/atipicial-go/pkg/interop/iterator"
	"github.com/atipicial/atipicial-go/pkg/interop/atipicialgointernal"
)

// AccountState contains info about a ATC holder.
type AccountState struct {
	Balance        int
	Height         int
	VoteTo         interop.PublicKey
	LastGasPerVote int
}

// Hash represents ATC contract hash.
const Hash = "\xf5\x63\xea\x40\xbc\x28\x3d\x4d\x0e\x05\xc4\x8e\xa3\x05\xb3\xf2\xa0\x73\x40\xef"

// Symbol represents `symbol` method of ATC native contract.
func Symbol() string {
	return atipicialgointernal.CallWithToken(Hash, "symbol", int(contract.NoneFlag)).(string)
}

// Decimals represents `decimals` method of ATC native contract.
func Decimals() int {
	return atipicialgointernal.CallWithToken(Hash, "decimals", int(contract.NoneFlag)).(int)
}

// TotalSupply represents `totalSupply` method of ATC native contract.
func TotalSupply() int {
	return atipicialgointernal.CallWithToken(Hash, "totalSupply", int(contract.ReadStates)).(int)
}

// BalanceOf represents `balanceOf` method of ATC native contract.
func BalanceOf(addr interop.Hash160) int {
	return atipicialgointernal.CallWithToken(Hash, "balanceOf", int(contract.ReadStates), addr).(int)
}

// Transfer represents `transfer` method of ATC native contract.
func Transfer(from, to interop.Hash160, amount int, data any) bool {
	return atipicialgointernal.CallWithToken(Hash, "transfer",
		int(contract.All), from, to, amount, data).(bool)
}

// GetCommittee represents `getCommittee` method of ATC native contract.
func GetCommittee() []interop.PublicKey {
	return atipicialgointernal.CallWithToken(Hash, "getCommittee", int(contract.ReadStates)).([]interop.PublicKey)
}

// GetCandidates represents `getCandidates` method of ATC native contract. It
// returns up to 256 candidates. Use GetAllCandidates in case if you need the
// whole set of candidates.
func GetCandidates() []Candidate {
	return atipicialgointernal.CallWithToken(Hash, "getCandidates", int(contract.ReadStates)).([]Candidate)
}

// GetAllCandidates represents `getAllCandidates` method of ATC native contract.
// It returns Iterator over the whole set of Atipicial candidates sorted by public key
// bytes. Each iterator value can be cast to Candidate. Use iterator interop
// package to work with the returned Iterator.
func GetAllCandidates() iterator.Iterator {
	return atipicialgointernal.CallWithToken(Hash, "getAllCandidates", int(contract.ReadStates)).(iterator.Iterator)
}

// GetCandidateVote represents `getCandidateVote` method of ATC native contract.
// It returns -1 if the candidate hasn't been registered or voted for and the
// overall candidate votes otherwise.
func GetCandidateVote(pub interop.PublicKey) int {
	return atipicialgointernal.CallWithToken(Hash, "getCandidateVote", int(contract.ReadStates), pub).(int)
}

// GetNextBlockValidators represents `getNextBlockValidators` method of ATC native contract.
func GetNextBlockValidators() []interop.PublicKey {
	return atipicialgointernal.CallWithToken(Hash, "getNextBlockValidators", int(contract.ReadStates)).([]interop.PublicKey)
}

// GetGASPerBlock represents `getGasPerBlock` method of ATC native contract.
func GetGASPerBlock() int {
	return atipicialgointernal.CallWithToken(Hash, "getGasPerBlock", int(contract.ReadStates)).(int)
}

// SetGASPerBlock represents `setGasPerBlock` method of ATC native contract.
func SetGASPerBlock(amount int) {
	atipicialgointernal.CallWithTokenNoRet(Hash, "setGasPerBlock", int(contract.States), amount)
}

// GetRegisterPrice represents `getRegisterPrice` method of ATC native contract.
func GetRegisterPrice() int {
	return atipicialgointernal.CallWithToken(Hash, "getRegisterPrice", int(contract.ReadStates)).(int)
}

// SetRegisterPrice represents `setRegisterPrice` method of ATC native contract.
func SetRegisterPrice(amount int) {
	atipicialgointernal.CallWithTokenNoRet(Hash, "setRegisterPrice", int(contract.States), amount)
}

// RegisterCandidate represents `registerCandidate` method of ATC native contract.
func RegisterCandidate(pub interop.PublicKey) bool {
	return atipicialgointernal.CallWithToken(Hash, "registerCandidate", int(contract.States|contract.AllowNotify), pub).(bool)
}

// UnregisterCandidate represents `unregisterCandidate` method of ATC native contract.
func UnregisterCandidate(pub interop.PublicKey) bool {
	return atipicialgointernal.CallWithToken(Hash, "unregisterCandidate", int(contract.States|contract.AllowNotify), pub).(bool)
}

// Vote represents `vote` method of ATC native contract.
func Vote(addr interop.Hash160, pub interop.PublicKey) bool {
	return atipicialgointernal.CallWithToken(Hash, "vote", int(contract.States|contract.AllowNotify), addr, pub).(bool)
}

// UnclaimedGAS represents `unclaimedGas` method of ATC native contract.
func UnclaimedGAS(addr interop.Hash160, end int) int {
	return atipicialgointernal.CallWithToken(Hash, "unclaimedGas", int(contract.ReadStates), addr, end).(int)
}

// GetAccountState represents `getAccountState` method of ATC native contract.
func GetAccountState(addr interop.Hash160) *AccountState {
	return atipicialgointernal.CallWithToken(Hash, "getAccountState", int(contract.ReadStates), addr).(*AccountState)
}

// GetCommitteeAddress represents `getCommitteeAddress` method of ATC native contract.
func GetCommitteeAddress() interop.Hash160 {
	return atipicialgointernal.CallWithToken(Hash, "getCommitteeAddress", int(contract.ReadStates)).(interop.Hash160)
}
