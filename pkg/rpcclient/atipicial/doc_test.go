package atipicial_test

import (
	"cmp"
	"context"
	"math/big"
	"slices"

	"github.com/atipicial/atipicial-go/pkg/encoding/address"
	"github.com/atipicial/atipicial-go/pkg/atipicialrpc/result"
	"github.com/atipicial/atipicial-go/pkg/rpcclient"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/actor"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/invoker"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/atipicial"
	"github.com/atipicial/atipicial-go/pkg/wallet"
)

func ExampleContractReader() {
	// No error checking done at all, intentionally.
	c, _ := rpcclient.New(context.Background(), "url", rpcclient.Options{})

	// Safe methods are reachable with just an invoker, no need for an account there.
	inv := invoker.New(c, nil)

	// Create a reader interface.
	atipicialToken := atipicial.NewReader(inv)

	// Account hash we're interested in.
	accHash, _ := address.StringToUint160("NdypBhqkz2CMMnwxBgvoC9X2XjKF5axgKo")

	// Get the account balance.
	balance, _ := atipicialToken.BalanceOf(accHash)
	_ = balance

	// Get the extended ATC-specific balance data.
	bAtipicial, _ := atipicialToken.GetAccountState(accHash)

	// Account can have no associated vote.
	if bAtipicial.VoteTo == nil {
		return
	}
	// Committee keys.
	comm, _ := atipicialToken.GetCommittee()

	// Check if the vote is made for a committee member.
	var votedForCommitteeMember bool
	if slices.ContainsFunc(comm, bAtipicial.VoteTo.Equal) {
		votedForCommitteeMember = true
	}
	_ = votedForCommitteeMember
}

func ExampleContract() {
	// No error checking done at all, intentionally.
	w, _ := wallet.NewWalletFromFile("somewhere")
	defer w.Close()

	c, _ := rpcclient.New(context.Background(), "url", rpcclient.Options{})

	// Create a simple CalledByEntry-scoped actor (assuming there is an account
	// inside the wallet).
	a, _ := actor.NewSimple(c, w.Accounts[0])

	// Create a complete contract representation.
	atipicialToken := atipicial.New(a)

	tgtAcc, _ := address.StringToUint160("NdypBhqkz2CMMnwxBgvoC9X2XjKF5axgKo")

	// Send a transaction that transfers one token to another account.
	txid, vub, _ := atipicialToken.Transfer(a.Sender(), tgtAcc, big.NewInt(1), nil)
	_ = txid
	_ = vub

	// Get a list of candidates (it's limited, but should be sufficient in most cases).
	cands, _ := atipicialToken.GetCandidates()

	// Sort by votes.
	slices.SortFunc(cands, func(a, b result.Validator) int { return cmp.Compare(a.Votes, b.Votes) })

	// Get the extended ATC-specific balance data.
	bAtipicial, _ := atipicialToken.GetAccountState(a.Sender())

	// If not yet voted, or voted for suboptimal candidate (we want the one with the least votes),
	// send a new voting transaction
	if bAtipicial.VoteTo == nil || !bAtipicial.VoteTo.Equal(&cands[0].PublicKey) {
		txid, vub, _ = atipicialToken.Vote(a.Sender(), &cands[0].PublicKey)
		_ = txid
		_ = vub
	}
}
