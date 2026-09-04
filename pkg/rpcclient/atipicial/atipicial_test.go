package atipicial

import (
	"errors"
	"math/big"
	"testing"

	"github.com/google/uuid"
	"github.com/atipicial/atipicial-go/pkg/core/state"
	"github.com/atipicial/atipicial-go/pkg/core/transaction"
	"github.com/atipicial/atipicial-go/pkg/crypto/keys"
	"github.com/atipicial/atipicial-go/pkg/atipicialrpc/result"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/atipicial/atipicial-go/pkg/vm/stackitem"
	"github.com/stretchr/testify/require"
)

type testAct struct {
	err error
	ser error
	res *result.Invoke
	rre *result.Invoke
	rer error
	tx  *transaction.Transaction
	txh util.Uint256
	vub uint32
	inv *result.Invoke
}

func (t *testAct) Call(contract util.Uint160, operation string, params ...any) (*result.Invoke, error) {
	return t.res, t.err
}
func (t *testAct) MakeRun(script []byte) (*transaction.Transaction, error) {
	return t.tx, t.err
}
func (t *testAct) MakeUnsignedRun(script []byte, attrs []transaction.Attribute) (*transaction.Transaction, error) {
	return t.tx, t.err
}
func (t *testAct) SendRun(script []byte) (util.Uint256, uint32, error) {
	return t.txh, t.vub, t.err
}
func (t *testAct) MakeCall(contract util.Uint160, method string, params ...any) (*transaction.Transaction, error) {
	return t.tx, t.err
}
func (t *testAct) MakeUnsignedCall(contract util.Uint160, method string, attrs []transaction.Attribute, params ...any) (*transaction.Transaction, error) {
	return t.tx, t.err
}
func (t *testAct) SendCall(contract util.Uint160, method string, params ...any) (util.Uint256, uint32, error) {
	return t.txh, t.vub, t.err
}
func (t *testAct) Run(script []byte) (*result.Invoke, error) {
	return t.rre, t.rer
}
func (t *testAct) MakeUnsignedUncheckedRun(script []byte, sysFee int64, attrs []transaction.Attribute) (*transaction.Transaction, error) {
	return t.tx, t.err
}
func (t *testAct) Sign(tx *transaction.Transaction) error {
	return t.ser
}
func (t *testAct) SignAndSend(tx *transaction.Transaction) (util.Uint256, uint32, error) {
	return t.txh, t.vub, t.err
}
func (t *testAct) CallAndExpandIterator(contract util.Uint160, method string, maxItems int, params ...any) (*result.Invoke, error) {
	return t.inv, t.err
}
func (t *testAct) TerminateSession(sessionID uuid.UUID) error {
	return t.err
}
func (t *testAct) TraverseIterator(sessionID uuid.UUID, iterator *result.Iterator, num int) ([]stackitem.Item, error) {
	return t.res.Stack, t.err
}

func TestGetAccountState(t *testing.T) {
	ta := &testAct{}
	atipicial := NewReader(ta)

	ta.err = errors.New("")
	_, err := atipicial.GetAccountState(util.Uint160{})
	require.Error(t, err)

	ta.err = nil
	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.Make(42),
		},
	}
	_, err = atipicial.GetAccountState(util.Uint160{})
	require.Error(t, err)

	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.Null{},
		},
	}
	st, err := atipicial.GetAccountState(util.Uint160{})
	require.NoError(t, err)
	require.Nil(t, st)

	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.Make([]stackitem.Item{
				stackitem.Make(100500),
				stackitem.Make(42),
				stackitem.Null{},
				stackitem.Make(0),
			}),
		},
	}
	st, err = atipicial.GetAccountState(util.Uint160{})
	require.NoError(t, err)
	require.Equal(t, &state.ATCBalance{
		AEP17Balance: state.AEP17Balance{
			Balance: *big.NewInt(100500),
		},
		BalanceHeight: 42,
	}, st)
}

func TestGetAllCandidates(t *testing.T) {
	ta := &testAct{}
	atipicial := NewReader(ta)

	ta.err = errors.New("")
	_, err := atipicial.GetAllCandidates()
	require.Error(t, err)

	ta.err = nil
	iid := uuid.New()
	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.NewInterop(result.Iterator{
				ID: &iid,
			}),
		},
	}
	_, err = atipicial.GetAllCandidates()
	require.Error(t, err)

	// Session-based iterator.
	sid := uuid.New()
	ta.res = &result.Invoke{
		Session: sid,
		State:   "HALT",
		Stack: []stackitem.Item{
			stackitem.NewInterop(result.Iterator{
				ID: &iid,
			}),
		},
	}
	iter, err := atipicial.GetAllCandidates()
	require.NoError(t, err)

	k, err := keys.NewPrivateKey()
	require.NoError(t, err)
	ta.res = &result.Invoke{
		Stack: []stackitem.Item{
			stackitem.Make([]stackitem.Item{
				stackitem.Make(k.PublicKey().Bytes()),
				stackitem.Make(100500),
			}),
		},
	}
	vals, err := iter.Next(10)
	require.NoError(t, err)
	require.Equal(t, 1, len(vals))
	require.Equal(t, result.Validator{
		PublicKey: *k.PublicKey(),
		Votes:     100500,
	}, vals[0])

	ta.err = errors.New("")
	_, err = iter.Next(1)
	require.Error(t, err)

	err = iter.Terminate()
	require.Error(t, err)

	// Value-based iterator.
	ta.err = nil
	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.NewInterop(result.Iterator{
				Values: []stackitem.Item{
					stackitem.Make(k.PublicKey().Bytes()),
					stackitem.Make(100500),
				},
			}),
		},
	}
	iter, err = atipicial.GetAllCandidates()
	require.NoError(t, err)

	ta.err = errors.New("")
	err = iter.Terminate()
	require.NoError(t, err)
}

func TestGetCandidates(t *testing.T) {
	ta := &testAct{}
	atipicial := NewReader(ta)

	ta.err = errors.New("")
	_, err := atipicial.GetCandidates()
	require.Error(t, err)

	ta.err = nil
	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.Make([]stackitem.Item{}),
		},
	}
	cands, err := atipicial.GetCandidates()
	require.NoError(t, err)
	require.Equal(t, 0, len(cands))

	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{stackitem.Make(42)},
	}
	_, err = atipicial.GetCandidates()
	require.Error(t, err)

	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.Make([]stackitem.Item{
				stackitem.Make(42),
			}),
		},
	}
	_, err = atipicial.GetCandidates()
	require.Error(t, err)

	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.Make([]stackitem.Item{
				stackitem.Make([]stackitem.Item{}),
			}),
		},
	}
	_, err = atipicial.GetCandidates()
	require.Error(t, err)

	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.Make([]stackitem.Item{
				stackitem.Make([]stackitem.Item{
					stackitem.Null{},
					stackitem.Null{},
				}),
			}),
		},
	}
	_, err = atipicial.GetCandidates()
	require.Error(t, err)

	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.Make([]stackitem.Item{
				stackitem.Make([]stackitem.Item{
					stackitem.Make("some"),
					stackitem.Null{},
				}),
			}),
		},
	}
	_, err = atipicial.GetCandidates()
	require.Error(t, err)

	k, err := keys.NewPrivateKey()
	require.NoError(t, err)
	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.Make([]stackitem.Item{
				stackitem.Make([]stackitem.Item{
					stackitem.Make(k.PublicKey().Bytes()),
					stackitem.Null{},
				}),
			}),
		},
	}
	_, err = atipicial.GetCandidates()
	require.Error(t, err)

	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.Make([]stackitem.Item{
				stackitem.Make([]stackitem.Item{
					stackitem.Make(k.PublicKey().Bytes()),
					stackitem.Make("canbeabigint"),
				}),
			}),
		},
	}
	_, err = atipicial.GetCandidates()
	require.Error(t, err)
}

func TestGetKeys(t *testing.T) {
	ta := &testAct{}
	atipicial := NewReader(ta)

	k, err := keys.NewPrivateKey()
	require.NoError(t, err)

	for _, m := range []func() (keys.PublicKeys, error){atipicial.GetCommittee, atipicial.GetNextBlockValidators} {
		ta.err = errors.New("")
		_, err := m()
		require.Error(t, err)

		ta.err = nil
		ta.res = &result.Invoke{
			State: "HALT",
			Stack: []stackitem.Item{
				stackitem.Make([]stackitem.Item{stackitem.Make(k.PublicKey().Bytes())}),
			},
		}
		ks, err := m()
		require.NoError(t, err)
		require.NotNil(t, ks)
		require.Equal(t, 1, len(ks))
		require.Equal(t, k.PublicKey(), ks[0])
	}
}

func TestGetInts(t *testing.T) {
	ta := &testAct{}
	atipicial := NewReader(ta)

	meth := []func() (int64, error){
		atipicial.GetGasPerBlock,
		atipicial.GetRegisterPrice,
	}

	ta.err = errors.New("")
	for _, m := range meth {
		_, err := m()
		require.Error(t, err)
	}

	ta.err = nil
	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.Make(42),
		},
	}
	for _, m := range meth {
		val, err := m()
		require.NoError(t, err)
		require.Equal(t, int64(42), val)
	}
}

func TestUnclaimedGas(t *testing.T) {
	ta := &testAct{}
	atipicial := NewReader(ta)

	ta.err = errors.New("")
	_, err := atipicial.UnclaimedGas(util.Uint160{}, 100500)
	require.Error(t, err)

	ta.err = nil
	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.Make([]stackitem.Item{}),
		},
	}
	_, err = atipicial.UnclaimedGas(util.Uint160{}, 100500)
	require.Error(t, err)

	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.Make(42),
		},
	}
	val, err := atipicial.UnclaimedGas(util.Uint160{}, 100500)
	require.NoError(t, err)
	require.Equal(t, big.NewInt(42), val)
}

func TestIntSetters(t *testing.T) {
	ta := new(testAct)
	atipicial := New(ta)

	meth := []func(int64) (util.Uint256, uint32, error){
		atipicial.SetGasPerBlock,
		atipicial.SetRegisterPrice,
	}

	ta.err = errors.New("")
	for _, m := range meth {
		_, _, err := m(42)
		require.Error(t, err)
	}

	ta.err = nil
	ta.txh = util.Uint256{1, 2, 3}
	ta.vub = 42
	for _, m := range meth {
		h, vub, err := m(100)
		require.NoError(t, err)
		require.Equal(t, ta.txh, h)
		require.Equal(t, ta.vub, vub)
	}
}

func TestIntTransactions(t *testing.T) {
	ta := new(testAct)
	atipicial := New(ta)

	for _, fun := range []func(int64) (*transaction.Transaction, error){
		atipicial.SetGasPerBlockTransaction,
		atipicial.SetGasPerBlockUnsigned,
		atipicial.SetRegisterPriceTransaction,
		atipicial.SetRegisterPriceUnsigned,
	} {
		ta.err = errors.New("")
		_, err := fun(1)
		require.Error(t, err)

		ta.err = nil
		ta.tx = &transaction.Transaction{Nonce: 100500, ValidUntilBlock: 42}
		tx, err := fun(1)
		require.NoError(t, err)
		require.Equal(t, ta.tx, tx)
	}
}

func TestVote(t *testing.T) {
	ta := new(testAct)
	atipicial := New(ta)

	k, err := keys.NewPrivateKey()
	require.NoError(t, err)

	ta.err = errors.New("")
	_, _, err = atipicial.Vote(util.Uint160{}, nil)
	require.Error(t, err)
	_, _, err = atipicial.Vote(util.Uint160{}, k.PublicKey())
	require.Error(t, err)
	_, err = atipicial.VoteTransaction(util.Uint160{}, nil)
	require.Error(t, err)
	_, err = atipicial.VoteTransaction(util.Uint160{}, k.PublicKey())
	require.Error(t, err)
	_, err = atipicial.VoteUnsigned(util.Uint160{}, nil)
	require.Error(t, err)
	_, err = atipicial.VoteUnsigned(util.Uint160{}, k.PublicKey())
	require.Error(t, err)

	ta.err = nil
	ta.txh = util.Uint256{1, 2, 3}
	ta.vub = 42

	h, vub, err := atipicial.Vote(util.Uint160{}, nil)
	require.NoError(t, err)
	require.Equal(t, ta.txh, h)
	require.Equal(t, ta.vub, vub)
	h, vub, err = atipicial.Vote(util.Uint160{}, k.PublicKey())
	require.NoError(t, err)
	require.Equal(t, ta.txh, h)
	require.Equal(t, ta.vub, vub)

	ta.tx = &transaction.Transaction{Nonce: 100500, ValidUntilBlock: 42}
	tx, err := atipicial.VoteTransaction(util.Uint160{}, nil)
	require.NoError(t, err)
	require.Equal(t, ta.tx, tx)
	tx, err = atipicial.VoteUnsigned(util.Uint160{}, k.PublicKey())
	require.NoError(t, err)
	require.Equal(t, ta.tx, tx)
}

func TestRegisterCandidate(t *testing.T) {
	ta := new(testAct)
	atipicial := New(ta)

	k, err := keys.NewPrivateKey()
	require.NoError(t, err)
	pk := k.PublicKey()

	ta.rer = errors.New("")
	_, _, err = atipicial.RegisterCandidate(pk)
	require.Error(t, err)
	_, err = atipicial.RegisterCandidateTransaction(pk)
	require.Error(t, err)
	_, err = atipicial.RegisterCandidateUnsigned(pk)
	require.Error(t, err)

	ta.rer = nil
	ta.txh = util.Uint256{1, 2, 3}
	ta.vub = 42
	ta.rre = &result.Invoke{
		GasConsumed: 100500,
	}
	ta.res = &result.Invoke{
		State: "HALT",
		Stack: []stackitem.Item{
			stackitem.Make(42),
		},
	}

	h, vub, err := atipicial.RegisterCandidate(pk)
	require.NoError(t, err)
	require.Equal(t, ta.txh, h)
	require.Equal(t, ta.vub, vub)

	ta.tx = &transaction.Transaction{Nonce: 100500, ValidUntilBlock: 42}
	tx, err := atipicial.RegisterCandidateTransaction(pk)
	require.NoError(t, err)
	require.Equal(t, ta.tx, tx)
	tx, err = atipicial.RegisterCandidateUnsigned(pk)
	require.NoError(t, err)
	require.Equal(t, ta.tx, tx)

	ta.ser = errors.New("")
	_, err = atipicial.RegisterCandidateTransaction(pk)
	require.Error(t, err)

	ta.err = errors.New("")
	_, err = atipicial.RegisterCandidateUnsigned(pk)
	require.Error(t, err)
}

func TestUnregisterCandidate(t *testing.T) {
	ta := new(testAct)
	atipicial := New(ta)

	k, err := keys.NewPrivateKey()
	require.NoError(t, err)
	pk := k.PublicKey()

	ta.err = errors.New("")
	_, _, err = atipicial.UnregisterCandidate(pk)
	require.Error(t, err)
	_, err = atipicial.UnregisterCandidateTransaction(pk)
	require.Error(t, err)
	_, err = atipicial.UnregisterCandidateUnsigned(pk)
	require.Error(t, err)

	ta.err = nil
	ta.txh = util.Uint256{1, 2, 3}
	ta.vub = 42

	h, vub, err := atipicial.UnregisterCandidate(pk)
	require.NoError(t, err)
	require.Equal(t, ta.txh, h)
	require.Equal(t, ta.vub, vub)

	ta.tx = &transaction.Transaction{Nonce: 100500, ValidUntilBlock: 42}
	tx, err := atipicial.UnregisterCandidateTransaction(pk)
	require.NoError(t, err)
	require.Equal(t, ta.tx, tx)
	tx, err = atipicial.UnregisterCandidateUnsigned(pk)
	require.NoError(t, err)
	require.Equal(t, ta.tx, tx)
}
