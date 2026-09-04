package core_test

import (
	"fmt"
	"math/big"
	"path/filepath"
	"testing"

	"github.com/atipicial/atipicial-go/internal/random"
	"github.com/atipicial/atipicial-go/pkg/config/netmode"
	"github.com/atipicial/atipicial-go/pkg/core/native/nativenames"
	"github.com/atipicial/atipicial-go/pkg/core/state"
	"github.com/atipicial/atipicial-go/pkg/core/storage"
	"github.com/atipicial/atipicial-go/pkg/core/storage/dbconfig"
	"github.com/atipicial/atipicial-go/pkg/core/transaction"
	"github.com/atipicial/atipicial-go/pkg/crypto/keys"
	"github.com/atipicial/atipicial-go/pkg/atipicialtest"
	"github.com/atipicial/atipicial-go/pkg/atipicialtest/chain"
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/wallet"
	"github.com/stretchr/testify/require"
)

func BenchmarkBlockchain_VerifyWitness(t *testing.B) {
	bc, acc := chain.NewSingle(t)
	e := atipicialtest.NewExecutor(t, bc, acc, acc)
	tx := e.NewTx(t, []atipicialtest.Signer{acc}, e.NativeHash(t, nativenames.Gas), "transfer", acc.ScriptHash(), acc.Script(), 1, nil)

	for t.Loop() {
		_, err := bc.VerifyWitness(tx.Signers[0].Account, tx, &tx.Scripts[0], 100000000)
		require.NoError(t, err)
	}
}

func BenchmarkBlockchain_ForEachAEP17Transfer(t *testing.B) {
	var stores = map[string]func(testing.TB) storage.Store{
		"MemPS": func(t testing.TB) storage.Store {
			return storage.NewMemoryStore()
		},
		"BoltPS":  newBoltStoreForTesting,
		"LevelPS": newLevelDBForTesting,
	}
	startFrom := []int{1, 100, 1000}
	blocksToTake := []int{100, 1000}
	for psName, newPS := range stores {
		for _, startFromBlock := range startFrom {
			for _, nBlocksToTake := range blocksToTake {
				t.Run(fmt.Sprintf("%s_StartFromBlockN-%d_Take%dBlocks", psName, startFromBlock, nBlocksToTake), func(t *testing.B) {
					ps := newPS(t)
					t.Cleanup(func() { ps.Close() })
					benchmarkForEachAEP17Transfer(t, ps, startFromBlock, nBlocksToTake)
				})
			}
		}
	}
}

func BenchmarkATC_GetGASPerVote(t *testing.B) {
	var stores = map[string]func(testing.TB) storage.Store{
		"MemPS": func(t testing.TB) storage.Store {
			return storage.NewMemoryStore()
		},
		"BoltPS":  newBoltStoreForTesting,
		"LevelPS": newLevelDBForTesting,
	}
	for psName, newPS := range stores {
		for nRewardRecords := 10; nRewardRecords <= 1000; nRewardRecords *= 10 {
			for rewardDistance := 1; rewardDistance <= 1000; rewardDistance *= 10 {
				t.Run(fmt.Sprintf("%s_%dRewardRecords_%dRewardDistance", psName, nRewardRecords, rewardDistance), func(t *testing.B) {
					ps := newPS(t)
					t.Cleanup(func() { ps.Close() })
					benchmarkGasPerVote(t, ps, nRewardRecords, rewardDistance)
				})
			}
		}
	}
}

func benchmarkForEachAEP17Transfer(t *testing.B, ps storage.Store, startFromBlock, nBlocksToTake int) {
	var (
		chainHeight       = 2_100                            // constant chain height to be able to compare paging results
		transfersPerBlock = state.TokenTransferBatchSize/4 + // 4 blocks per batch
			state.TokenTransferBatchSize/32 // shift
	)

	bc, validators, committee := chain.NewMultiWithCustomConfigAndStore(t, nil, ps, true)

	e := atipicialtest.NewExecutor(t, bc, validators, committee)
	gasHash := e.NativeHash(t, nativenames.Gas)

	acc := random.Uint160()
	from := e.Validator.ScriptHash()

	for range chainHeight {
		b := smartcontract.NewBuilder()
		for range transfersPerBlock {
			b.InvokeWithAssert(gasHash, "transfer", from, acc, 1, nil)
		}
		script, err := b.Script()
		require.NoError(t, err)
		tx := transaction.New(script, int64(1100_0000*transfersPerBlock))
		tx.NetworkFee = 1_0000_000
		tx.ValidUntilBlock = bc.BlockHeight() + 1
		tx.Nonce = atipicialtest.Nonce()
		tx.Signers = []transaction.Signer{{Account: from, Scopes: transaction.CalledByEntry}}
		require.NoError(t, validators.SignTx(netmode.UnitTestNet, tx))
		e.AddNewBlock(t, tx)
		e.CheckHalt(t, tx.Hash())
	}

	newestB, err := bc.GetBlock(bc.GetHeaderHash(bc.BlockHeight() - uint32(startFromBlock) + 1))
	require.NoError(t, err)
	newestTimestamp := newestB.Timestamp
	oldestB, err := bc.GetBlock(bc.GetHeaderHash(newestB.Index - uint32(nBlocksToTake)))
	require.NoError(t, err)
	oldestTimestamp := oldestB.Timestamp

	t.ReportAllocs()
	t.StartTimer()
	for t.Loop() {
		require.NoError(t, bc.ForEachAEP17Transfer(acc, newestTimestamp, func(t *state.AEP17Transfer) (bool, error) {
			if t.Timestamp < oldestTimestamp {
				// iterating from newest to oldest, already have reached the needed height
				return false, nil
			}
			return true, nil
		}))
	}
	t.StopTimer()
}

func newLevelDBForTesting(t testing.TB) storage.Store {
	dbPath := t.TempDir()
	dbOptions := dbconfig.LevelDBOptions{
		DataDirectoryPath: dbPath,
	}
	newLevelStore, err := storage.NewLevelDBStore(dbOptions)
	require.Nil(t, err, "NewLevelDBStore error")
	return newLevelStore
}

func newBoltStoreForTesting(t testing.TB) storage.Store {
	d := t.TempDir()
	dbPath := filepath.Join(d, "test_bolt_db")
	boltDBStore, err := storage.NewBoltDBStore(dbconfig.BoltDBOptions{FilePath: dbPath})
	require.NoError(t, err)
	return boltDBStore
}

func benchmarkGasPerVote(t *testing.B, ps storage.Store, nRewardRecords int, rewardDistance int) {
	bc, validators, committee := chain.NewMultiWithCustomConfigAndStore(t, nil, ps, true)
	cfg := bc.GetConfig()

	e := atipicialtest.NewExecutor(t, bc, validators, committee)
	atipicialHash := e.NativeHash(t, nativenames.Atipicial)
	gasHash := e.NativeHash(t, nativenames.Gas)
	atipicialSuperInvoker := e.NewInvoker(atipicialHash, validators, committee)
	atipicialValidatorsInvoker := e.ValidatorInvoker(atipicialHash)
	gasValidatorsInvoker := e.ValidatorInvoker(gasHash)

	// Vote for new committee.
	sz := len(cfg.StandbyCommittee)
	voters := make([]*wallet.Account, sz)
	candidates := make(keys.PublicKeys, sz)
	txs := make([]*transaction.Transaction, 0, len(voters)*3)
	for i := range sz {
		priv, err := keys.NewPrivateKey()
		require.NoError(t, err)
		candidates[i] = priv.PublicKey()
		voters[i], err = wallet.NewAccount()
		require.NoError(t, err)
		registerTx := atipicialSuperInvoker.PrepareInvoke(t, "registerCandidate", candidates[i].Bytes())
		txs = append(txs, registerTx)

		to := voters[i].Contract.ScriptHash()
		transferAtipicialTx := atipicialValidatorsInvoker.PrepareInvoke(t, "transfer", e.Validator.ScriptHash(), to, big.NewInt(int64(sz-i)*1000000).Int64(), nil)
		txs = append(txs, transferAtipicialTx)

		transferGasTx := gasValidatorsInvoker.PrepareInvoke(t, "transfer", e.Validator.ScriptHash(), to, int64(1_000_000_000), nil)
		txs = append(txs, transferGasTx)
	}
	e.AddNewBlock(t, txs...)
	for _, tx := range txs {
		e.CheckHalt(t, tx.Hash())
	}
	voteTxs := make([]*transaction.Transaction, 0, sz)
	for i := range sz {
		priv := voters[i].PrivateKey()
		h := priv.GetScriptHash()
		voteTx := e.NewTx(t, []atipicialtest.Signer{atipicialtest.NewSingleSigner(voters[i])}, atipicialHash, "vote", h, candidates[i].Bytes())
		voteTxs = append(voteTxs, voteTx)
	}
	e.AddNewBlock(t, voteTxs...)
	for _, tx := range voteTxs {
		e.CheckHalt(t, tx.Hash())
	}

	// Collect set of nRewardRecords reward records for each voter.
	e.GenerateNewBlocks(t, len(cfg.StandbyCommittee))

	// Transfer some more ATC to first voter to update his balance height.
	to := voters[0].Contract.ScriptHash()
	atipicialValidatorsInvoker.Invoke(t, true, "transfer", e.Validator.ScriptHash(), to, int64(1), nil)

	// Advance chain one more time to avoid same start/end rewarding bounds.
	e.GenerateNewBlocks(t, rewardDistance)
	end := bc.BlockHeight()

	t.ReportAllocs()
	t.StartTimer()
	for t.Loop() {
		_, err := bc.CalculateClaimable(to, end)
		require.NoError(t, err)
	}
	t.StopTimer()
}
