package native_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"math/big"
	"slices"
	"strings"
	"testing"

	"github.com/atipicial/atipicial-go/internal/contracts"
	"github.com/atipicial/atipicial-go/internal/random"
	"github.com/atipicial/atipicial-go/pkg/compiler"
	"github.com/atipicial/atipicial-go/pkg/config"
	"github.com/atipicial/atipicial-go/pkg/core/interop/interopnames"
	"github.com/atipicial/atipicial-go/pkg/core/native"
	"github.com/atipicial/atipicial-go/pkg/core/native/nativenames"
	"github.com/atipicial/atipicial-go/pkg/core/state"
	"github.com/atipicial/atipicial-go/pkg/core/transaction"
	"github.com/atipicial/atipicial-go/pkg/crypto/hash"
	"github.com/atipicial/atipicial-go/pkg/crypto/keys"
	"github.com/atipicial/atipicial-go/pkg/io"
	"github.com/atipicial/atipicial-go/pkg/atipicialtest"
	"github.com/atipicial/atipicial-go/pkg/atipicialtest/chain"
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/callflag"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/trigger"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/atipicial/atipicial-go/pkg/vm/emit"
	"github.com/atipicial/atipicial-go/pkg/vm/opcode"
	"github.com/atipicial/atipicial-go/pkg/vm/stackitem"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newAtipicialCommitteeClient(t *testing.T, expectedGASBalance int, cfg ...func(*config.Blockchain)) *atipicialtest.ContractInvoker {
	cfgF := func(cfg *config.Blockchain) {
		cfg.Hardforks = map[string]uint32{
			config.HFAspidochelone.String(): 0,
			config.HFBasilisk.String():      0,
			config.HFCockatrice.String():    0,
			config.HFDomovoi.String():       0,
			config.HFEchidna.String():       0,
			config.HFGorgon.String():        0,
		}
	}
	if len(cfg) > 0 {
		cfgF = cfg[0]
	}
	bc, validators, committee := chain.NewMultiWithCustomConfig(t, cfgF)
	e := atipicialtest.NewExecutor(t, bc, validators, committee)

	if expectedGASBalance > 0 {
		e.ValidatorInvoker(e.NativeHash(t, nativenames.Gas)).Invoke(t, true, "transfer", e.Validator.ScriptHash(), e.CommitteeHash, expectedGASBalance, nil)
	}

	return e.CommitteeInvoker(e.NativeHash(t, nativenames.Atipicial))
}

func newAtipicialValidatorsClient(t *testing.T) *atipicialtest.ContractInvoker {
	c := newAtipicialCommitteeClient(t, 100_0000_0000)
	return c.ValidatorInvoker(c.NativeHash(t, nativenames.Atipicial))
}

func TestATC_GasPerBlock(t *testing.T) {
	testGetSet(t, newAtipicialCommitteeClient(t, 100_0000_0000), "GasPerBlock", 5*native.GASFactor, 0, 10*native.GASFactor)
}

func TestATC_GasPerBlockCache(t *testing.T) {
	testGetSetCache(t, newAtipicialCommitteeClient(t, 100_0000_0000), "GasPerBlock", 5*native.GASFactor)
}

func TestATC_RegisterPrice(t *testing.T) {
	testGetSet(t, newAtipicialCommitteeClient(t, 100_0000_0000), "RegisterPrice", native.DefaultRegisterPrice, 1, math.MaxInt64)
}

func TestATC_RegisterPriceCache(t *testing.T) {
	testGetSetCache(t, newAtipicialCommitteeClient(t, 100_0000_0000), "RegisterPrice", native.DefaultRegisterPrice)
}

func TestATC_CandidateEvents(t *testing.T) {
	c := newNativeClient(t, nativenames.Atipicial)
	singleSigner := c.Signers[0].(atipicialtest.MultiSigner).Single(0)
	cc := c.WithSigners(c.Signers[0], singleSigner)
	e := c.Executor
	pkb := singleSigner.Account().PublicKey().Bytes()

	// Register 1 -> event
	tx := cc.Invoke(t, true, "registerCandidate", pkb)
	e.CheckTxNotificationEvent(t, tx, 0, state.NotificationEvent{
		ScriptHash: c.Hash,
		Name:       "CandidateStateChanged",
		Item: stackitem.NewArray([]stackitem.Item{
			stackitem.NewByteArray(pkb),
			stackitem.NewBool(true),
			stackitem.Make(0),
		}),
	})

	// Register 2 -> no event
	tx = cc.Invoke(t, true, "registerCandidate", pkb)
	aer := e.GetTxExecResult(t, tx)
	require.Equal(t, 0, len(aer.Events))

	// Vote -> event
	tx = c.Invoke(t, true, "vote", c.Signers[0].ScriptHash().BytesBE(), pkb)
	e.CheckTxNotificationEvent(t, tx, 0, state.NotificationEvent{
		ScriptHash: c.Hash,
		Name:       "Vote",
		Item: stackitem.NewArray([]stackitem.Item{
			stackitem.NewByteArray(c.Signers[0].ScriptHash().BytesBE()),
			stackitem.Null{},
			stackitem.NewByteArray(pkb),
			stackitem.Make(100000000),
		}),
	})

	// Unregister 1 -> event
	tx = cc.Invoke(t, true, "unregisterCandidate", pkb)
	e.CheckTxNotificationEvent(t, tx, 0, state.NotificationEvent{
		ScriptHash: c.Hash,
		Name:       "CandidateStateChanged",
		Item: stackitem.NewArray([]stackitem.Item{
			stackitem.NewByteArray(pkb),
			stackitem.NewBool(false),
			stackitem.Make(100000000),
		}),
	})

	// Unregister 2 -> no event
	tx = cc.Invoke(t, true, "unregisterCandidate", pkb)
	aer = e.GetTxExecResult(t, tx)
	require.Equal(t, 0, len(aer.Events))
}

// electNewCommittee registers a set of accounts as candidates and votes for
// them so that they are elected as the new committee. It returns a set of
// candidates.
func electNewCommittee(t *testing.T, e *atipicialtest.Executor, atipicialCommitteeInvoker, atipicialValidatorsInvoker *atipicialtest.ContractInvoker) []atipicialtest.Signer {
	cfg := e.Chain.GetConfig()
	committeeSize := cfg.GetCommitteeSize(0)

	voters := make([]atipicialtest.Signer, committeeSize)
	candidates := make([]atipicialtest.Signer, committeeSize)
	for i := range committeeSize {
		voters[i] = e.NewAccount(t, 10_0000_0000)
		candidates[i] = e.NewAccount(t, 2000_0000_0000) // enough for one registration
	}
	txes := make([]*transaction.Transaction, 0, committeeSize*3)
	for i := range committeeSize {
		transferTx := atipicialValidatorsInvoker.PrepareInvoke(t, "transfer", e.Validator.ScriptHash(), voters[i].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), int64(committeeSize-i)*1000000, nil)
		txes = append(txes, transferTx)

		registerTx := atipicialValidatorsInvoker.WithSigners(candidates[i]).PrepareInvoke(t, "registerCandidate", candidates[i].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
		txes = append(txes, registerTx)

		voteTx := atipicialValidatorsInvoker.WithSigners(voters[i]).PrepareInvoke(t, "vote", voters[i].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), candidates[i].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
		txes = append(txes, voteTx)
	}
	block := atipicialValidatorsInvoker.AddNewBlock(t, txes...)
	for _, tx := range txes {
		e.CheckHalt(t, tx.Hash(), stackitem.Make(true))
	}

	// Advance the chain to trigger committee recalculation and potential change.
	for (block.Index)%uint32(committeeSize) != 0 {
		block = atipicialCommitteeInvoker.AddNewBlock(t)
	}

	return candidates
}

func TestATC_CommitteeEvents(t *testing.T) {
	atipicialCommitteeInvoker := newAtipicialCommitteeClient(t, 100_0000_0000)
	atipicialValidatorsInvoker := atipicialCommitteeInvoker.WithSigners(atipicialCommitteeInvoker.Validator)
	e := atipicialCommitteeInvoker.Executor

	cfg := e.Chain.GetConfig()
	committeeSize := cfg.GetCommitteeSize(0)
	candidates := electNewCommittee(t, e, atipicialCommitteeInvoker, atipicialValidatorsInvoker)

	// Check for CommitteeChanged event in the last persisted block's AER.
	blockHash := e.Chain.CurrentBlockHash()
	aer, err := e.Chain.GetAppExecResults(blockHash, trigger.OnPersist)
	require.NoError(t, err)
	require.Equal(t, 1, len(aer))

	require.Equal(t, aer[0].Events[0].Name, "CommitteeChanged")
	require.Equal(t, 2, len(aer[0].Events[0].Item.Value().([]stackitem.Item)))

	expectedOldCommitteePublicKeys, err := keys.NewPublicKeysFromStrings(cfg.StandbyCommittee)
	require.NoError(t, err)
	expectedOldCommitteeStackItems := make([]stackitem.Item, len(expectedOldCommitteePublicKeys))
	for i, pubKey := range expectedOldCommitteePublicKeys {
		expectedOldCommitteeStackItems[i] = stackitem.NewByteArray(pubKey.Bytes())
	}
	oldCommitteeStackItem := aer[0].Events[0].Item.Value().([]stackitem.Item)[0].(*stackitem.Array)
	for i, item := range oldCommitteeStackItem.Value().([]stackitem.Item) {
		assert.Equal(t, expectedOldCommitteeStackItems[i].(*stackitem.ByteArray).Value().([]byte), item.Value().([]byte))
	}
	expectedNewCommitteeStackItems := make([]stackitem.Item, 0, committeeSize)
	for _, candidate := range candidates {
		expectedNewCommitteeStackItems = append(expectedNewCommitteeStackItems, stackitem.NewByteArray(candidate.(atipicialtest.SingleSigner).Account().PublicKey().Bytes()))
	}
	newCommitteeStackItem := aer[0].Events[0].Item.Value().([]stackitem.Item)[1].(*stackitem.Array)
	for i, item := range newCommitteeStackItem.Value().([]stackitem.Item) {
		assert.Equal(t, expectedNewCommitteeStackItems[i].(*stackitem.ByteArray).Value().([]byte), item.Value().([]byte))
	}
}

func TestATC_Vote(t *testing.T) {
	atipicialCommitteeInvoker := newAtipicialCommitteeClient(t, 100_0000_0000)
	atipicialValidatorsInvoker := atipicialCommitteeInvoker.WithSigners(atipicialCommitteeInvoker.Validator)
	policyInvoker := atipicialCommitteeInvoker.CommitteeInvoker(atipicialCommitteeInvoker.NativeHash(t, nativenames.Policy))
	e := atipicialCommitteeInvoker.Executor

	cfg := e.Chain.GetConfig()
	committeeSize := cfg.GetCommitteeSize(0)
	validatorsCount := cfg.GetNumOfCNs(0)
	freq := validatorsCount + committeeSize
	advanceChain := func(t *testing.T) {
		for range freq {
			atipicialCommitteeInvoker.AddNewBlock(t)
		}
	}

	standBySorted, err := keys.NewPublicKeysFromStrings(e.Chain.GetConfig().StandbyCommittee)
	require.NoError(t, err)
	standBySorted = standBySorted[:validatorsCount]
	slices.SortFunc(standBySorted, (*keys.PublicKey).Cmp)
	pubs := e.Chain.ComputeNextBlockValidators()
	require.Equal(t, standBySorted, keys.PublicKeys(pubs))

	// voters vote for candidates. The aim of this test is to check if voting
	// reward is proportional to the ATC balance.
	voters := make([]atipicialtest.Signer, committeeSize+1)
	// referenceAccounts perform the same actions as voters except voting, i.e. we
	// will transfer the same amount of ATC to referenceAccounts and see how much
	// GAS they receive for ATC ownership. We need these values to be able to define
	// how much GAS voters receive for ATC ownership.
	referenceAccounts := make([]atipicialtest.Signer, committeeSize+1)
	candidates := make([]atipicialtest.Signer, committeeSize+1)
	for i := range committeeSize + 1 {
		voters[i] = e.NewAccount(t, 10_0000_0000)
		referenceAccounts[i] = e.NewAccount(t, 10_0000_0000)
		candidates[i] = e.NewAccount(t, 2000_0000_0000) // enough for one registration
	}
	txes := make([]*transaction.Transaction, 0, committeeSize*4-2)
	for i := range committeeSize + 1 {
		transferTx := atipicialValidatorsInvoker.PrepareInvoke(t, "transfer", e.Validator.ScriptHash(), voters[i].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), int64(committeeSize+1-i)*1000000, nil)
		txes = append(txes, transferTx)
		transferTx = atipicialValidatorsInvoker.PrepareInvoke(t, "transfer", e.Validator.ScriptHash(), referenceAccounts[i].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), int64(committeeSize+1-i)*1000000, nil)
		txes = append(txes, transferTx)
		if i > 0 {
			registerTx := atipicialValidatorsInvoker.WithSigners(candidates[i]).PrepareInvoke(t, "registerCandidate", candidates[i].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
			txes = append(txes, registerTx)
			voteTx := atipicialValidatorsInvoker.WithSigners(voters[i]).PrepareInvoke(t, "vote", voters[i].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), candidates[i].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
			txes = append(txes, voteTx)
		}
	}
	txes = append(txes, policyInvoker.PrepareInvoke(t, "blockAccount", candidates[len(candidates)-1].(atipicialtest.SingleSigner).Account().ScriptHash()))
	atipicialValidatorsInvoker.AddNewBlock(t, txes...)
	for _, tx := range txes {
		e.CheckHalt(t, tx.Hash(), stackitem.Make(true)) // luckily, both `transfer`, `registerCandidate` and `vote` return boolean values
	}

	// We still haven't voted enough validators in.
	pubs = e.Chain.ComputeNextBlockValidators()
	require.NoError(t, err)
	require.Equal(t, standBySorted, keys.PublicKeys(pubs))

	advanceChain(t)
	pubs, err = e.Chain.GetNextBlockValidators()
	require.NoError(t, err)
	require.EqualValues(t, standBySorted, keys.PublicKeys(pubs))

	// Register and give some value to the last validator.
	txes = txes[:0]
	registerTx := atipicialValidatorsInvoker.WithSigners(candidates[0]).PrepareInvoke(t, "registerCandidate", candidates[0].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
	txes = append(txes, registerTx)
	voteTx := atipicialValidatorsInvoker.WithSigners(voters[0]).PrepareInvoke(t, "vote", voters[0].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), candidates[0].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
	txes = append(txes, voteTx)
	atipicialValidatorsInvoker.AddNewBlock(t, txes...)
	for _, tx := range txes {
		e.CheckHalt(t, tx.Hash(), stackitem.Make(true)) // luckily, both `transfer`, `registerCandidate` and `vote` return boolean values
	}

	advanceChain(t)
	pubs, err = atipicialCommitteeInvoker.Chain.GetNextBlockValidators()
	require.NoError(t, err)
	sortedCandidates := make(keys.PublicKeys, validatorsCount)
	for i := range candidates[:validatorsCount] {
		sortedCandidates[i] = candidates[i].(atipicialtest.SingleSigner).Account().PublicKey()
	}
	slices.SortFunc(sortedCandidates, (*keys.PublicKey).Cmp)
	require.EqualValues(t, sortedCandidates, keys.PublicKeys(pubs))

	pubs, err = atipicialCommitteeInvoker.Chain.GetNextBlockValidators()
	require.NoError(t, err)
	require.EqualValues(t, sortedCandidates, pubs)

	t.Run("check voter rewards", func(t *testing.T) {
		gasBalance := make([]*big.Int, len(voters)-1)
		referenceGASBalance := make([]*big.Int, len(referenceAccounts)-1)
		atipicialBalance := make([]*big.Int, len(voters)-1)
		txes = make([]*transaction.Transaction, 0, len(voters)-1)
		var refTxFee int64
		for i := range voters[:len(voters)-1] {
			h := voters[i].ScriptHash()
			refH := referenceAccounts[i].ScriptHash()
			gasBalance[i] = e.Chain.GetUtilityTokenBalance(h, util.Uint160{})
			atipicialBalance[i], _ = e.Chain.GetGoverningTokenBalance(h)
			referenceGASBalance[i] = e.Chain.GetUtilityTokenBalance(refH, util.Uint160{})

			tx := atipicialCommitteeInvoker.WithSigners(voters[i]).PrepareInvoke(t, "transfer", h.BytesBE(), h.BytesBE(), int64(1), nil)
			txes = append(txes, tx)
			tx = atipicialCommitteeInvoker.WithSigners(referenceAccounts[i]).PrepareInvoke(t, "transfer", refH.BytesBE(), refH.BytesBE(), int64(1), nil)
			txes = append(txes, tx)
			refTxFee = tx.SystemFee + tx.NetworkFee
		}
		atipicialCommitteeInvoker.AddNewBlock(t, txes...)
		for _, tx := range txes {
			e.CheckHalt(t, tx.Hash(), stackitem.Make(true))
		}

		// Define reference reward for ATC holding for each voter account.
		for i := range referenceGASBalance {
			newBalance := e.Chain.GetUtilityTokenBalance(referenceAccounts[i].ScriptHash(), util.Uint160{})
			referenceGASBalance[i].Sub(newBalance, referenceGASBalance[i])
			referenceGASBalance[i].Add(referenceGASBalance[i], big.NewInt(refTxFee))
		}

		// GAS increase consists of 2 parts: ATC holding + voting for committee nodes.
		// Here we check that 2-nd part exists and is proportional to the amount of ATC given.
		for i := range voters[:len(voters)-1] {
			newGAS := e.Chain.GetUtilityTokenBalance(voters[i].ScriptHash(), util.Uint160{})
			newGAS.Sub(newGAS, gasBalance[i])
			gasForHold := referenceGASBalance[i]
			newGAS.Sub(newGAS, gasForHold)
			require.True(t, newGAS.Sign() > 0)
			gasBalance[i] = newGAS
		}
		// First account voted later than the others.
		require.Equal(t, -1, gasBalance[0].Cmp(gasBalance[1]))
		for i := 2; i < validatorsCount; i++ {
			require.Equal(t, 0, gasBalance[i].Cmp(gasBalance[1]))
		}
		require.Equal(t, 1, gasBalance[1].Cmp(gasBalance[validatorsCount]))
		for i := validatorsCount; i < committeeSize; i++ {
			require.Equal(t, 0, gasBalance[i].Cmp(gasBalance[validatorsCount]))
		}
	})

	atipicialCommitteeInvoker.WithSigners(candidates[0]).Invoke(t, true, "unregisterCandidate", candidates[0].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
	atipicialCommitteeInvoker.WithSigners(voters[0]).Invoke(t, false, "vote", voters[0].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), candidates[0].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())

	advanceChain(t)

	pubs = e.Chain.ComputeNextBlockValidators()
	for i := range pubs {
		require.NotEqual(t, candidates[0], pubs[i])
		require.NotEqual(t, candidates[len(candidates)-1], pubs[i])
	}
	// LastGasPerVote should be 0 after unvoting
	getAccountState := func(t *testing.T, account util.Uint160) *state.ATCBalance {
		stack, err := atipicialCommitteeInvoker.TestInvoke(t, "getAccountState", account)
		require.NoError(t, err)
		res := stack.Pop().Item()
		// (s *ATCBalance) FromStackItem is able to handle both 3 and 4 subitems.
		// The forth optional subitem is LastGasPerVote.
		require.Equal(t, 4, len(res.Value().([]stackitem.Item)))
		as := new(state.ATCBalance)
		err = as.FromStackItem(res)
		require.NoError(t, err)
		return as
	}
	registerTx = atipicialValidatorsInvoker.WithSigners(candidates[0]).PrepareInvoke(t, "registerCandidate", candidates[0].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
	voteTx = atipicialValidatorsInvoker.WithSigners(voters[0]).PrepareInvoke(t, "vote", voters[0].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), candidates[0].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
	atipicialValidatorsInvoker.AddNewBlock(t, registerTx, voteTx)
	e.CheckHalt(t, registerTx.Hash(), stackitem.Make(true))
	e.CheckHalt(t, voteTx.Hash(), stackitem.Make(true))

	stateBeforeUnvote := getAccountState(t, voters[0].ScriptHash())
	require.NotEqual(t, uint64(0), stateBeforeUnvote.LastGasPerVote.Uint64())
	// Unvote
	unvoteTx := atipicialValidatorsInvoker.WithSigners(voters[0]).PrepareInvoke(t, "vote", voters[0].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), nil)
	atipicialValidatorsInvoker.AddNewBlock(t, unvoteTx)
	e.CheckHalt(t, unvoteTx.Hash(), stackitem.Make(true))
	advanceChain(t)

	stateAfterUnvote := getAccountState(t, voters[0].ScriptHash())
	require.Equal(t, uint64(0), stateAfterUnvote.LastGasPerVote.Uint64())
}

// TestATC_RecursiveGASMint is a test for https://github.com/atipicial/atipicial-go/pull/2181.
func TestATC_RecursiveGASMint(t *testing.T) {
	atipicialCommitteeInvoker := newAtipicialCommitteeClient(t, 100_0000_0000)
	atipicialValidatorInvoker := atipicialCommitteeInvoker.WithSigners(atipicialCommitteeInvoker.Validator)
	e := atipicialCommitteeInvoker.Executor
	gasValidatorInvoker := e.ValidatorInvoker(e.NativeHash(t, nativenames.Gas))

	c := atipicialtest.CompileFile(t, e.Validator.ScriptHash(), "../../../../internal/basicchain/testdata/test_contract.go", "../../../../internal/basicchain/testdata/test_contract.yml")
	e.DeployContract(t, c, nil)

	gasValidatorInvoker.Invoke(t, true, "transfer", e.Validator.ScriptHash(), c.Hash, int64(2_0000_0000), nil)

	// Transfer 10 ATC to test contract, the contract should earn some GAS by owning this ATC.
	atipicialValidatorInvoker.Invoke(t, true, "transfer", e.Validator.ScriptHash(), c.Hash, int64(10), nil)

	// Add blocks to be able to trigger ATC transfer from contract address to owner
	// address inside onAEP17Payment (the contract starts ATC transfers from chain height = 100).
	for i := e.Chain.BlockHeight(); i < 100; i++ {
		e.AddNewBlock(t)
	}

	// Transfer 1 more ATC to the contract. Transfer will trigger onAEP17Payment. OnAEP17Payment will
	// trigger transfer of 11 ATC to the contract owner (based on the contract code). 11 ATC Transfer will
	// trigger GAS distribution. GAS transfer will trigger OnAEP17Payment one more time. The recursion
	// shouldn't occur here, because contract's balance LastUpdated height has already been updated in
	// this block.
	atipicialValidatorInvoker.Invoke(t, true, "transfer", e.Validator.ScriptHash(), c.Hash, int64(1), nil)
}

func TestATC_GetCommitteeAddress(t *testing.T) {
	atipicialValidatorInvoker := newAtipicialValidatorsClient(t)
	e := atipicialValidatorInvoker.Executor
	cfg := atipicialValidatorInvoker.Chain.GetConfig()

	maxHardforkHeight := uint32(0)
	for _, height := range cfg.Hardforks {
		if height > maxHardforkHeight {
			maxHardforkHeight = height
		}
	}
	for range maxHardforkHeight {
		atipicialValidatorInvoker.AddNewBlock(t)
	}
	standByCommitteePublicKeys, err := keys.NewPublicKeysFromStrings(e.Chain.GetConfig().StandbyCommittee)
	require.NoError(t, err)
	slices.SortFunc(standByCommitteePublicKeys, (*keys.PublicKey).Cmp)
	expectedCommitteeAddress, err := smartcontract.CreateMajorityMultiSigRedeemScript(standByCommitteePublicKeys)
	require.NoError(t, err)
	stack, err := atipicialValidatorInvoker.TestInvoke(t, "getCommitteeAddress")
	require.NoError(t, err)
	require.Equal(t, hash.Hash160(expectedCommitteeAddress).BytesBE(), stack.Pop().Item().Value().([]byte))
}

func TestATC_GetAccountState(t *testing.T) {
	atipicialValidatorInvoker := newAtipicialValidatorsClient(t)
	e := atipicialValidatorInvoker.Executor

	cfg := e.Chain.GetConfig()
	committeeSize := cfg.GetCommitteeSize(0)
	validatorSize := cfg.GetNumOfCNs(0)
	advanceChain := func(t *testing.T) {
		for range committeeSize {
			atipicialValidatorInvoker.AddNewBlock(t)
		}
	}

	t.Run("empty", func(t *testing.T) {
		atipicialValidatorInvoker.Invoke(t, stackitem.Null{}, "getAccountState", util.Uint160{})
	})

	t.Run("with funds", func(t *testing.T) {
		amount := int64(1)
		acc := e.NewAccount(t)
		atipicialValidatorInvoker.Invoke(t, true, "transfer", e.Validator.ScriptHash(), acc.ScriptHash(), amount, nil)
		lub := e.Chain.BlockHeight()
		atipicialValidatorInvoker.Invoke(t, stackitem.NewStruct([]stackitem.Item{
			stackitem.Make(amount),
			stackitem.Make(lub),
			stackitem.Null{},
			stackitem.Make(0),
		}), "getAccountState", acc.ScriptHash())
	})

	t.Run("lastGasPerVote", func(t *testing.T) {
		const (
			GasPerBlock      = 5
			VoterRewardRatio = 80
		)
		getAccountState := func(t *testing.T, account util.Uint160) *state.ATCBalance {
			stack, err := atipicialValidatorInvoker.TestInvoke(t, "getAccountState", account)
			require.NoError(t, err)
			as := new(state.ATCBalance)
			err = as.FromStackItem(stack.Pop().Item())
			require.NoError(t, err)
			return as
		}

		amount := int64(1000)
		acc := e.NewAccount(t)
		atipicialValidatorInvoker.Invoke(t, true, "transfer", e.Validator.ScriptHash(), acc.ScriptHash(), amount, nil)
		as := getAccountState(t, acc.ScriptHash())
		require.Equal(t, uint64(amount), as.Balance.Uint64())
		require.Equal(t, e.Chain.BlockHeight(), as.BalanceHeight)
		require.Equal(t, uint64(0), as.LastGasPerVote.Uint64())
		committee, _ := e.Chain.GetCommittee()
		atipicialValidatorInvoker.WithSigners(e.Validator, e.Validator.(atipicialtest.MultiSigner).Single(0)).Invoke(t, true, "registerCandidate", committee[0].Bytes())
		atipicialValidatorInvoker.WithSigners(acc).Invoke(t, true, "vote", acc.ScriptHash(), committee[0].Bytes())
		as = getAccountState(t, acc.ScriptHash())
		require.Equal(t, uint64(0), as.LastGasPerVote.Uint64())
		advanceChain(t)
		atipicialValidatorInvoker.WithSigners(acc).Invoke(t, true, "transfer", acc.ScriptHash(), acc.ScriptHash(), amount, nil)
		as = getAccountState(t, acc.ScriptHash())
		expect := GasPerBlock * native.GASFactor * VoterRewardRatio / 100 * (uint64(e.Chain.BlockHeight()) / uint64(committeeSize))
		expect = expect * uint64(committeeSize) / uint64(validatorSize+committeeSize) * native.ATCTotalSupply / as.Balance.Uint64()
		require.Equal(t, e.Chain.BlockHeight(), as.BalanceHeight)
		require.Equal(t, expect, as.LastGasPerVote.Uint64())
	})
}

func TestATC_GetAccountStateInteropAPI(t *testing.T) {
	atipicialValidatorInvoker := newAtipicialValidatorsClient(t)
	e := atipicialValidatorInvoker.Executor

	cfg := e.Chain.GetConfig()
	committeeSize := cfg.GetCommitteeSize(0)
	validatorSize := cfg.GetNumOfCNs(0)
	advanceChain := func(t *testing.T) {
		for range committeeSize {
			atipicialValidatorInvoker.AddNewBlock(t)
		}
	}

	amount := int64(1000)
	acc := e.NewAccount(t)
	atipicialValidatorInvoker.Invoke(t, true, "transfer", e.Validator.ScriptHash(), acc.ScriptHash(), amount, nil)
	committee, _ := e.Chain.GetCommittee()
	atipicialValidatorInvoker.WithSigners(e.Validator, e.Validator.(atipicialtest.MultiSigner).Single(0)).Invoke(t, true, "registerCandidate", committee[0].Bytes())
	atipicialValidatorInvoker.WithSigners(acc).Invoke(t, true, "vote", acc.ScriptHash(), committee[0].Bytes())
	advanceChain(t)
	atipicialValidatorInvoker.WithSigners(acc).Invoke(t, true, "transfer", acc.ScriptHash(), acc.ScriptHash(), amount, nil)

	var hashAStr strings.Builder
	for i := range util.Uint160Size {
		fmt.Fprintf(&hashAStr, "%#x", acc.ScriptHash()[i])
		if i != util.Uint160Size-1 {
			hashAStr.WriteString(", ")
		}
	}
	src := `package testaccountstate
	  import (
		  "github.com/atipicial/atipicial-go/pkg/interop/native/atipicial"
		  "github.com/atipicial/atipicial-go/pkg/interop"
	  )
	  func GetLastGasPerVote() int {
		  accState := atipicial.GetAccountState(interop.Hash160{` + hashAStr.String() + `})
		  if accState == nil {
			  panic("nil state")
		  }
		  return accState.LastGasPerVote
	  }`
	ctr := atipicialtest.CompileSource(t, e.Validator.ScriptHash(), strings.NewReader(src), &compiler.Options{
		Name: "testaccountstate_contract",
	})
	e.DeployContract(t, ctr, nil)

	const (
		GasPerBlock      = 5
		VoterRewardRatio = 80
	)
	expect := GasPerBlock * native.GASFactor * VoterRewardRatio / 100 * (uint64(e.Chain.BlockHeight()) / uint64(committeeSize))
	expect = expect * uint64(committeeSize) / uint64(validatorSize+committeeSize) * native.ATCTotalSupply / uint64(amount)
	ctrInvoker := e.NewInvoker(ctr.Hash, e.Committee)
	ctrInvoker.Invoke(t, stackitem.Make(expect), "getLastGasPerVote")
}

func TestATC_CommitteeBountyOnPersist(t *testing.T) {
	atipicialCommitteeInvoker := newAtipicialCommitteeClient(t, 0)
	e := atipicialCommitteeInvoker.Executor

	hs, err := keys.NewPublicKeysFromStrings(e.Chain.GetConfig().StandbyCommittee)
	require.NoError(t, err)
	committeeSize := len(hs)

	const singleBounty = 50000000
	bs := map[int]int64{0: singleBounty}
	checkBalances := func() {
		for i := range committeeSize {
			require.EqualValues(t, bs[i], e.Chain.GetUtilityTokenBalance(hs[i].GetScriptHash(), util.Uint160{}).Int64(), i)
		}
	}
	for i := range committeeSize * 2 {
		e.AddNewBlock(t)
		bs[(i+1)%committeeSize] += singleBounty
		checkBalances()
	}
}

func TestATC_TransferOnPayment(t *testing.T) {
	atipicialValidatorsInvoker := newAtipicialValidatorsClient(t)
	e := atipicialValidatorsInvoker.Executor
	managementValidatorsInvoker := e.ValidatorInvoker(e.NativeHash(t, nativenames.Management))

	cs, _ := contracts.GetTestContractState(t, pathToInternalContracts, 1, 2, e.CommitteeHash)
	cs.Hash = state.CreateContractHash(e.Validator.ScriptHash(), cs.AEF.Checksum, cs.Manifest.Name) // set proper hash
	manifB, err := json.Marshal(cs.Manifest)
	require.NoError(t, err)
	aefB, err := cs.AEF.Bytes()
	require.NoError(t, err)
	si, err := cs.ToStackItem()
	require.NoError(t, err)
	managementValidatorsInvoker.Invoke(t, si, "deploy", aefB, manifB)

	const amount int64 = 2

	h := atipicialValidatorsInvoker.Invoke(t, true, "transfer", e.Validator.ScriptHash(), cs.Hash, amount, nil)
	aer := e.GetTxExecResult(t, h)
	require.Equal(t, 3, len(aer.Events)) // transfer + GAS claim for sender + onPayment
	e.CheckTxNotificationEvent(t, h, 1, state.NotificationEvent{
		ScriptHash: cs.Hash,
		Name:       "LastPaymentAEP17",
		Item: stackitem.NewArray([]stackitem.Item{
			stackitem.NewByteArray(atipicialValidatorsInvoker.Hash.BytesBE()),
			stackitem.NewByteArray(e.Validator.ScriptHash().BytesBE()),
			stackitem.NewBigInteger(big.NewInt(amount)),
			stackitem.Null{},
		}),
	})

	h = atipicialValidatorsInvoker.Invoke(t, true, "transfer", e.Validator.ScriptHash(), cs.Hash, amount, nil)
	aer = e.GetTxExecResult(t, h)
	require.Equal(t, 5, len(aer.Events))                         // Now we must also have GAS claim for contract and corresponding `onPayment`.
	e.CheckTxNotificationEvent(t, h, 1, state.NotificationEvent{ // onPayment for ATC transfer
		ScriptHash: cs.Hash,
		Name:       "LastPaymentAEP17",
		Item: stackitem.NewArray([]stackitem.Item{
			stackitem.NewByteArray(e.NativeHash(t, nativenames.Atipicial).BytesBE()),
			stackitem.NewByteArray(e.Validator.ScriptHash().BytesBE()),
			stackitem.NewBigInteger(big.NewInt(amount)),
			stackitem.Null{},
		}),
	})
	e.CheckTxNotificationEvent(t, h, 4, state.NotificationEvent{ // onPayment for GAS claim
		ScriptHash: cs.Hash,
		Name:       "LastPaymentAEP17",
		Item: stackitem.NewArray([]stackitem.Item{
			stackitem.NewByteArray(e.NativeHash(t, nativenames.Gas).BytesBE()),
			stackitem.Null{},
			stackitem.NewBigInteger(big.NewInt(1)),
			stackitem.Null{},
		}),
	})
}

func TestATC_Roundtrip(t *testing.T) {
	atipicialValidatorsInvoker := newAtipicialValidatorsClient(t)
	e := atipicialValidatorsInvoker.Executor
	validatorH := atipicialValidatorsInvoker.Validator.ScriptHash()

	initialBalance, initialHeight := e.Chain.GetGoverningTokenBalance(validatorH)
	require.NotNil(t, initialBalance)

	t.Run("bad: amount > initial balance", func(t *testing.T) {
		h := atipicialValidatorsInvoker.Invoke(t, false, "transfer", validatorH, validatorH, initialBalance.Int64()+1, nil)
		aer, err := e.Chain.GetAppExecResults(h, trigger.Application)
		require.NoError(t, err)
		require.Equal(t, 0, len(aer[0].Events)) // failed transfer => no events
		// check balance and height were not changed
		updatedBalance, updatedHeight := e.Chain.GetGoverningTokenBalance(validatorH)
		require.Equal(t, initialBalance, updatedBalance)
		require.Equal(t, initialHeight, updatedHeight)
	})

	t.Run("good: amount == initial balance", func(t *testing.T) {
		h := atipicialValidatorsInvoker.Invoke(t, true, "transfer", validatorH, validatorH, initialBalance.Int64(), nil)
		aer, err := e.Chain.GetAppExecResults(h, trigger.Application)
		require.NoError(t, err)
		require.Equal(t, 2, len(aer[0].Events)) // roundtrip + GAS claim
		// check balance wasn't changed and height was updated
		updatedBalance, updatedHeight := e.Chain.GetGoverningTokenBalance(validatorH)
		require.Equal(t, initialBalance, updatedBalance)
		require.Equal(t, e.Chain.BlockHeight(), updatedHeight)
	})
}

func TestATC_TransferZeroWithZeroBalance(t *testing.T) {
	atipicialValidatorsInvoker := newAtipicialValidatorsClient(t)
	e := atipicialValidatorsInvoker.Executor

	check := func(t *testing.T, roundtrip bool) {
		acc := atipicialValidatorsInvoker.WithSigners(e.NewAccount(t))
		accH := acc.Signers[0].ScriptHash()
		to := accH
		if !roundtrip {
			to = random.Uint160()
		}
		h := acc.Invoke(t, true, "transfer", accH, to, int64(0), nil)
		aer, err := e.Chain.GetAppExecResults(h, trigger.Application)
		require.NoError(t, err)
		require.Equal(t, 1, len(aer[0].Events))                                                                       // roundtrip/transfer only, no GAS claim
		require.Equal(t, stackitem.NewBigInteger(big.NewInt(0)), aer[0].Events[0].Item.Value().([]stackitem.Item)[2]) // amount is 0
		// check balance wasn't changed and height was updated
		updatedBalance, updatedHeight := e.Chain.GetGoverningTokenBalance(accH)
		require.Equal(t, int64(0), updatedBalance.Int64())
		require.Equal(t, uint32(0), updatedHeight)
	}
	t.Run("roundtrip: amount == initial balance == 0", func(t *testing.T) {
		check(t, true)
	})
	t.Run("non-roundtrip: amount == initial balance == 0", func(t *testing.T) {
		check(t, false)
	})
}

func TestATC_TransferZeroWithNonZeroBalance(t *testing.T) {
	atipicialValidatorsInvoker := newAtipicialValidatorsClient(t)
	e := atipicialValidatorsInvoker.Executor

	check := func(t *testing.T, roundtrip bool) {
		acc := e.NewAccount(t)
		atipicialValidatorsInvoker.Invoke(t, true, "transfer", atipicialValidatorsInvoker.Validator.ScriptHash(), acc.ScriptHash(), int64(100), nil)
		atipicialAccInvoker := atipicialValidatorsInvoker.WithSigners(acc)
		initialBalance, _ := e.Chain.GetGoverningTokenBalance(acc.ScriptHash())
		require.True(t, initialBalance.Sign() > 0)
		to := acc.ScriptHash()
		if !roundtrip {
			to = random.Uint160()
		}
		h := atipicialAccInvoker.Invoke(t, true, "transfer", acc.ScriptHash(), to, int64(0), nil)

		aer, err := e.Chain.GetAppExecResults(h, trigger.Application)
		require.NoError(t, err)
		require.Equal(t, 2, len(aer[0].Events))                                                                       // roundtrip + GAS claim
		require.Equal(t, stackitem.NewBigInteger(big.NewInt(0)), aer[0].Events[0].Item.Value().([]stackitem.Item)[2]) // amount is 0
		// check balance wasn't changed and height was updated
		updatedBalance, updatedHeight := e.Chain.GetGoverningTokenBalance(acc.ScriptHash())
		require.Equal(t, initialBalance, updatedBalance)
		require.Equal(t, e.Chain.BlockHeight(), updatedHeight)
	}
	t.Run("roundtrip", func(t *testing.T) {
		check(t, true)
	})
	t.Run("non-roundtrip", func(t *testing.T) {
		check(t, false)
	})
}

// https://github.com/atipicial/atipicial-go/issues/3190
func TestATC_TransferNonZeroWithZeroBalance(t *testing.T) {
	atipicialValidatorsInvoker := newAtipicialValidatorsClient(t)
	e := atipicialValidatorsInvoker.Executor

	acc := atipicialValidatorsInvoker.WithSigners(e.NewAccount(t))
	accH := acc.Signers[0].ScriptHash()
	h := acc.Invoke(t, false, "transfer", accH, accH, int64(5), nil)
	aer := e.CheckHalt(t, h, stackitem.Make(false))
	require.Equal(t, 0, len(aer.Events))
	// check balance wasn't changed and height was not updated
	updatedBalance, updatedHeight := e.Chain.GetGoverningTokenBalance(accH)
	require.Equal(t, int64(0), updatedBalance.Int64())
	require.Equal(t, uint32(0), updatedHeight)
}

func TestATC_CalculateBonus(t *testing.T) {
	atipicialCommitteeInvoker := newAtipicialCommitteeClient(t, 10_0000_0000)
	e := atipicialCommitteeInvoker.Executor
	atipicialValidatorsInvoker := atipicialCommitteeInvoker.WithSigners(e.Validator)

	acc := atipicialValidatorsInvoker.WithSigners(e.NewAccount(t))
	accH := acc.Signers[0].ScriptHash()
	rewardDistance := 10

	// We have 11 blocks made by transactions above and we need block 13 to get Echidna.
	// Otherwise this happens:
	//    logger.go:146: 2024-12-24T17:52:18.160+0300 WARN    contract invocation failed      {"tx": "603d1b0e29e7aaf50513689be9d5bb946c7f7fec8836f0e90897c825fb762c13", "block": 13, "error": "at instruction 120 (SYSCALL): System.Contract.CallNative failed: gas limit exceeded"}
	for range 3 {
		e.AddNewBlock(t)
	}

	t.Run("Zero", func(t *testing.T) {
		initialGASBalance := e.Chain.GetUtilityTokenBalance(accH, util.Uint160{})
		for range rewardDistance {
			e.AddNewBlock(t)
		}
		// Claim GAS, but there's no ATC on the account, so no GAS should be earned.
		h := acc.Invoke(t, true, "transfer", accH, accH, 0, nil)
		claimTx, _ := e.GetTransaction(t, h)

		e.CheckGASBalance(t, accH, big.NewInt(initialGASBalance.Int64()-claimTx.SystemFee-claimTx.NetworkFee))
	})

	t.Run("Many blocks", func(t *testing.T) {
		amount := 100
		defaultGASPerBlock := 5
		newGASPerBlock := 1
		atipicialHolderRewardRatio := 10

		initialGASBalance := e.Chain.GetUtilityTokenBalance(accH, util.Uint160{})

		// Five blocks of ATC owning with default GasPerBlockValue.
		atipicialValidatorsInvoker.Invoke(t, true, "transfer", e.Validator.ScriptHash(), accH, amount, nil)
		for range rewardDistance/2 - 2 {
			e.AddNewBlock(t)
		}
		atipicialCommitteeInvoker.Invoke(t, stackitem.Null{}, "setGasPerBlock", newGASPerBlock*native.GASFactor)

		// Five blocks more with modified GasPerBlock value.
		for range rewardDistance / 2 {
			e.AddNewBlock(t)
		}

		// GAS claim for the last 10 blocks of ATC owning.
		h := acc.Invoke(t, true, "transfer", accH, accH, amount, nil)
		claimTx, _ := e.GetTransaction(t, h)

		firstPart := int64(amount * atipicialHolderRewardRatio / 100 * // reward for a part of the whole ATC total supply that is owned by acc
			defaultGASPerBlock * // GAS generated by a single block
			rewardDistance / 2) // number of blocks generated with specified GasPerBlock
		secondPart := int64(amount * atipicialHolderRewardRatio / 100 * // reward for a part of the whole ATC total supply that is owned by acc
			newGASPerBlock * // GAS generated by a single block after GasPerBlock update
			rewardDistance / 2) // number of blocks generated with specified GasPerBlock
		e.CheckGASBalance(t, accH, big.NewInt(initialGASBalance.Int64()-
			claimTx.SystemFee-claimTx.NetworkFee + +firstPart + secondPart))
	})
}

func TestATC_UnclaimedGas(t *testing.T) {
	atipicialValidatorsInvoker := newAtipicialValidatorsClient(t)
	e := atipicialValidatorsInvoker.Executor

	acc := atipicialValidatorsInvoker.WithSigners(e.NewAccount(t))
	accH := acc.Signers[0].ScriptHash()

	t.Run("non-existing account", func(t *testing.T) {
		// non-existing account, should return zero unclaimed GAS.
		acc.Invoke(t, 0, "unclaimedGas", util.Uint160{}, 1)
	})

	t.Run("non-zero balance", func(t *testing.T) {
		amount := 100
		defaultGASPerBlock := 5
		atipicialHolderRewardRatio := 10
		rewardDistance := 10
		atipicialValidatorsInvoker.Invoke(t, true, "transfer", e.Validator.ScriptHash(), accH, amount, nil)

		for range rewardDistance - 1 {
			e.AddNewBlock(t)
		}
		expectedGas := int64(amount * atipicialHolderRewardRatio / 100 * defaultGASPerBlock * rewardDistance)
		acc.Invoke(t, expectedGas, "unclaimedGas", accH, e.Chain.BlockHeight()+1)

		acc.InvokeFail(t, "can't calculate bonus of height unequal (BlockHeight + 1)", "unclaimedGas", accH, e.Chain.BlockHeight())
	})
}

func TestATC_GetCandidates(t *testing.T) {
	atipicialCommitteeInvoker := newAtipicialCommitteeClient(t, 100_0000_0000)
	atipicialValidatorsInvoker := atipicialCommitteeInvoker.WithSigners(atipicialCommitteeInvoker.Validator)
	policyInvoker := atipicialCommitteeInvoker.CommitteeInvoker(atipicialCommitteeInvoker.NativeHash(t, nativenames.Policy))
	e := atipicialCommitteeInvoker.Executor

	cfg := e.Chain.GetConfig()
	candidatesCount := cfg.GetCommitteeSize(0) - 1

	// Register a set of candidates and vote for them.
	voters := make([]atipicialtest.Signer, candidatesCount)
	candidates := make([]atipicialtest.Signer, candidatesCount)
	for i := range candidatesCount {
		voters[i] = e.NewAccount(t, 10_0000_0000)
		candidates[i] = e.NewAccount(t, 2000_0000_0000) // enough for one registration
	}
	txes := make([]*transaction.Transaction, 0, candidatesCount*3)
	for i := range candidatesCount {
		transferTx := atipicialValidatorsInvoker.PrepareInvoke(t, "transfer", e.Validator.ScriptHash(), voters[i].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), int64(candidatesCount+1-i)*1000000, nil)
		txes = append(txes, transferTx)
		registerTx := atipicialValidatorsInvoker.WithSigners(candidates[i]).PrepareInvoke(t, "registerCandidate", candidates[i].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
		txes = append(txes, registerTx)
		voteTx := atipicialValidatorsInvoker.WithSigners(voters[i]).PrepareInvoke(t, "vote", voters[i].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), candidates[i].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
		txes = append(txes, voteTx)
	}

	atipicialValidatorsInvoker.AddNewBlock(t, txes...)
	for _, tx := range txes {
		e.CheckHalt(t, tx.Hash(), stackitem.Make(true)) // luckily, both `transfer`, `registerCandidate` and `vote` return boolean values
	}
	expected := make([]stackitem.Item, candidatesCount)
	for i := range expected {
		pub := candidates[i].(atipicialtest.SingleSigner).Account().PublicKey().Bytes()
		v := stackitem.NewBigInteger(big.NewInt(int64(candidatesCount-i+1) * 1000000))
		expected[i] = stackitem.NewStruct([]stackitem.Item{
			stackitem.NewByteArray(pub),
			v,
		})
		atipicialCommitteeInvoker.Invoke(t, v, "getCandidateVote", pub)
	}
	slices.SortFunc(expected, func(a, b stackitem.Item) int {
		return bytes.Compare(a.Value().([]stackitem.Item)[0].Value().([]byte), b.Value().([]stackitem.Item)[0].Value().([]byte))
	})
	atipicialCommitteeInvoker.Invoke(t, stackitem.NewArray(expected), "getCandidates")

	// Check that GetAllCandidates works the same way as GetCandidates.
	checkGetAllCandidates := func(t *testing.T, expected []stackitem.Item) {
		for i := range len(expected) + 1 {
			w := io.NewBufBinWriter()
			emit.AppCall(w.BinWriter, atipicialCommitteeInvoker.Hash, "getAllCandidates", callflag.All)
			for range i + 1 {
				emit.Opcodes(w.BinWriter, opcode.DUP)
				emit.Syscall(w.BinWriter, interopnames.SystemIteratorNext)
				emit.Opcodes(w.BinWriter, opcode.DROP) // drop the value returned from Next.
			}
			emit.Syscall(w.BinWriter, interopnames.SystemIteratorValue)
			require.NoError(t, w.Err)
			h := atipicialCommitteeInvoker.InvokeScript(t, w.Bytes(), atipicialCommitteeInvoker.Signers)
			if i < len(expected) {
				e.CheckHalt(t, h, expected[i])
			} else {
				e.CheckFault(t, h, "iterator index out of range") // ensure there are no extra elements.
			}
			w.Reset()
		}
	}
	checkGetAllCandidates(t, expected)

	// Block candidate and check it won't be returned from getCandidates and getAllCandidates.
	unlucky := candidates[len(candidates)-1].(atipicialtest.SingleSigner).Account().PublicKey()
	policyInvoker.Invoke(t, true, "blockAccount", unlucky.GetScriptHash())
	for i := range expected {
		if bytes.Equal(expected[i].Value().([]stackitem.Item)[0].Value().([]byte), unlucky.Bytes()) {
			if i != len(expected)-1 {
				expected = append(expected[:i], expected[i+1:]...)
			} else {
				expected = expected[:i]
			}
			break
		}
	}
	atipicialCommitteeInvoker.Invoke(t, expected, "getCandidates")
	checkGetAllCandidates(t, expected)
}

func TestATC_RegisterViaAEP27(t *testing.T) {
	atipicialCommitteeInvoker := newAtipicialCommitteeClient(t, 100_0000_0000)
	atipicialValidatorsInvoker := atipicialCommitteeInvoker.WithSigners(atipicialCommitteeInvoker.Validator)
	e := atipicialCommitteeInvoker.Executor
	atipicialHash := e.NativeHash(t, nativenames.Atipicial)

	cfg := e.Chain.GetConfig()
	candidatesCount := cfg.GetCommitteeSize(0) - 1

	// Register a set of candidates and vote for them.
	voters := make([]atipicialtest.Signer, candidatesCount)
	candidates := make([]atipicialtest.Signer, candidatesCount)
	for i := range candidatesCount {
		voters[i] = e.NewAccount(t, 2000_0000_0000) // enough for one registration
		candidates[i] = e.NewAccount(t, 2000_0000_0000)
	}

	stack, err := atipicialCommitteeInvoker.TestInvoke(t, "getRegisterPrice")
	require.NoError(t, err)
	registrationPrice, err := stack.Pop().Item().TryInteger()
	require.NoError(t, err)

	// We have 11 blocks made by transactions above and we need block 13 to get Echidna.
	for range 3 {
		e.AddNewBlock(t)
	}

	gasValidatorsInvoker := e.CommitteeInvoker(e.NativeHash(t, nativenames.Gas))
	txes := make([]*transaction.Transaction, 0, candidatesCount*3)
	for i := range candidatesCount {
		transferTx := atipicialValidatorsInvoker.PrepareInvoke(t, "transfer", e.Validator.ScriptHash(), voters[i].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), int64(candidatesCount+1-i)*1000000, nil)
		txes = append(txes, transferTx)
		registerTx := gasValidatorsInvoker.WithSigners(candidates[i]).PrepareInvoke(t, "transfer", candidates[i].(atipicialtest.SingleSigner).Account().ScriptHash(), atipicialHash, registrationPrice, candidates[i].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
		txes = append(txes, registerTx)
		voteTx := atipicialValidatorsInvoker.WithSigners(voters[i]).PrepareInvoke(t, "vote", voters[i].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), candidates[i].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
		txes = append(txes, voteTx)
	}

	atipicialValidatorsInvoker.AddNewBlock(t, txes...)
	for _, tx := range txes {
		e.CheckHalt(t, tx.Hash(), stackitem.Make(true)) // luckily, both `transfer` and `vote` return boolean values
	}

	// Ensure ATC holds no GAS.
	stack, err = gasValidatorsInvoker.TestInvoke(t, "balanceOf", atipicialHash)
	require.NoError(t, err)
	balance, err := stack.Pop().Item().TryInteger()
	require.NoError(t, err)
	require.Equal(t, 0, balance.Sign())

	var expected = make([]stackitem.Item, candidatesCount)
	for i := range expected {
		pub := candidates[i].(atipicialtest.SingleSigner).Account().PublicKey().Bytes()
		v := stackitem.NewBigInteger(big.NewInt(int64(candidatesCount-i+1) * 1000000))
		expected[i] = stackitem.NewStruct([]stackitem.Item{
			stackitem.NewByteArray(pub),
			v,
		})
		atipicialCommitteeInvoker.Invoke(t, v, "getCandidateVote", pub)
	}

	slices.SortFunc(expected, func(a, b stackitem.Item) int {
		return bytes.Compare(a.Value().([]stackitem.Item)[0].Value().([]byte), b.Value().([]stackitem.Item)[0].Value().([]byte))
	})

	atipicialCommitteeInvoker.Invoke(t, stackitem.NewArray(expected), "getCandidates")

	// Invalid cases.
	var newCand = voters[0]

	// Missing data.
	gasValidatorsInvoker.WithSigners(newCand).InvokeFail(t, "invalid conversion", "transfer", newCand.(atipicialtest.SingleSigner).Account().ScriptHash(), atipicialHash, registrationPrice, nil)
	// Invalid data.
	gasValidatorsInvoker.WithSigners(newCand).InvokeFail(t, "unexpected EOF", "transfer", newCand.(atipicialtest.SingleSigner).Account().ScriptHash(), atipicialHash, registrationPrice, []byte{2, 2, 2})
	// ATC transfer.
	atipicialValidatorsInvoker.WithSigners(newCand).InvokeFail(t, "only GAS is accepted", "transfer", newCand.(atipicialtest.SingleSigner).Account().ScriptHash(), atipicialHash, 1, newCand.(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
	// Incorrect amount.
	gasValidatorsInvoker.WithSigners(newCand).InvokeFail(t, "incorrect GAS amount", "transfer", newCand.(atipicialtest.SingleSigner).Account().ScriptHash(), atipicialHash, 1, newCand.(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
	// Incorrect witness.
	var anotherAcc = e.NewAccount(t, 2000_0000_0000)
	gasValidatorsInvoker.WithSigners(newCand).InvokeFail(t, "not witnessed by the key owner", "transfer", newCand.(atipicialtest.SingleSigner).Account().ScriptHash(), atipicialHash, registrationPrice, anotherAcc.(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
}

func TestAtipicial_GasPerBlockUpdate(t *testing.T) {
	c := newAtipicialCommitteeClient(t, 100_0000_0000)
	c.Invoke(t, stackitem.Null{}, "setGasPerBlock", 0)

	// No GAS should be generated for this block since GasPerBlock changes are
	// applied starting from the same block, ref. #3889.
	aer, err := c.Chain.GetAppExecResults(c.TopBlock(t).Hash(), trigger.PostPersist)
	require.NoError(t, err)
	require.Equal(t, 0, len(aer[0].Events))
}

// TestAtipicial_TransferNegative ensures that transfer of a negative ATC amount leads to
// a VM FAULT, ref. #4072.
func TestAtipicial_TransferNegative(t *testing.T) {
	c := newAtipicialCommitteeClient(t, 10_0000_0000)
	c.InvokeFail(t, "negative amount", "transfer", c.Signers[0].ScriptHash(), c.Signers[0].ScriptHash(), -1, nil)
}

// Ref. https://github.com/atipicial-project/atipicial/pull/4569.
func TestATC_GasPerVote_LastEpochBlock(t *testing.T) {
	check := func(t *testing.T, enableGorgon bool) {
		var cfgF = func(cfg *config.Blockchain) {
			cfg.Hardforks = map[string]uint32{
				config.HFFaun.String(): 0,
			}
		}
		if enableGorgon {
			cfgF = func(cfg *config.Blockchain) {
				cfg.Hardforks = map[string]uint32{
					config.HFGorgon.String(): 0,
				}
			}
		}
		atipicialCommitteeInvoker := newAtipicialCommitteeClient(t, 100_0000_0000, cfgF)
		atipicialValidatorsInvoker := atipicialCommitteeInvoker.WithSigners(atipicialCommitteeInvoker.Validator)
		e := atipicialCommitteeInvoker.Executor

		cfg := e.Chain.GetConfig()
		committeeSize := cfg.GetCommitteeSize(0)
		candidatesCount := committeeSize

		// Register a set of candidates and vote for themselves.
		candidates := make([]atipicialtest.Signer, candidatesCount)
		for i := range candidatesCount {
			candidates[i] = e.NewAccount(t, 2000_0000_0000) // enough for one registration.
		}

		txes := make([]*transaction.Transaction, 0, candidatesCount*3)
		for i := range candidatesCount {
			transferTx := atipicialValidatorsInvoker.PrepareInvoke(t, "transfer", e.Validator.ScriptHash(), candidates[i].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), 5_000_000+i, nil) // candidate order should be fixed, hence use +i.
			txes = append(txes, transferTx)
			registerTx := atipicialValidatorsInvoker.WithSigners(candidates[i]).PrepareInvoke(t, "registerCandidate", candidates[i].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
			txes = append(txes, registerTx)
			voteTx := atipicialValidatorsInvoker.WithSigners(candidates[i]).PrepareInvoke(t, "vote", candidates[i].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash(), candidates[i].(atipicialtest.SingleSigner).Account().PublicKey().Bytes())
			txes = append(txes, voteTx)
		}
		atipicialValidatorsInvoker.AddNewBlock(t, txes...)
		for _, tx := range txes {
			e.CheckHalt(t, tx.Hash(), stackitem.Make(true)) // luckily, both `transfer`, `registerCandidate` and `vote` return boolean values
		}

		// Add some blocks to reach last block of the epoch.
		for int(e.Chain.BlockHeight())%committeeSize != committeeSize-1 {
			e.AddNewBlock(t)
		}

		// Change the number of votes for a single candidate. Ensure an up-to-date (non-cached)
		// number of votes is used to calculate gas per vote.
		acc0 := candidates[0].(atipicialtest.SingleSigner).Account().PrivateKey().GetScriptHash()
		atipicialValidatorsInvoker.Invoke(t, true, "transfer", e.Validator.ScriptHash(), acc0, 50_000_000, nil)

		unclaimed, err := e.Chain.CalculateClaimable(acc0, e.Chain.BlockHeight()+1)
		require.NoError(t, err)
		if enableGorgon {
			require.Equal(t, int64(267_499_999), unclaimed.Int64())
		} else {
			require.Equal(t, int64(2_667_500_000), unclaimed.Int64()) // outdated number of votes during GasPerVote update in the end of the epoch leads to wrong (larger) GasPerVote.
		}
	}

	check(t, false)
	check(t, true)
}
