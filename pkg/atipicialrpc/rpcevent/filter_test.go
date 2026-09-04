package rpcevent

import (
	"testing"

	"github.com/atipicial/atipicial-go/pkg/core/block"
	"github.com/atipicial/atipicial-go/pkg/core/mempoolevent"
	"github.com/atipicial/atipicial-go/pkg/core/state"
	"github.com/atipicial/atipicial-go/pkg/core/transaction"
	"github.com/atipicial/atipicial-go/pkg/atipicialrpc"
	"github.com/atipicial/atipicial-go/pkg/atipicialrpc/result"
	"github.com/atipicial/atipicial-go/pkg/network/payload"
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/atipicial/atipicial-go/pkg/vm/stackitem"
	"github.com/atipicial/atipicial-go/pkg/vm/vmstate"
	"github.com/stretchr/testify/require"
)

type (
	testComparator struct {
		id     atipicialrpc.EventID
		filter atipicialrpc.SubscriptionFilter
	}
	testContainer struct {
		id  atipicialrpc.EventID
		pld any
	}
)

func (c testComparator) EventID() atipicialrpc.EventID {
	return c.id
}
func (c testComparator) Filter() atipicialrpc.SubscriptionFilter {
	return c.filter
}
func (c testContainer) EventID() atipicialrpc.EventID {
	return c.id
}
func (c testContainer) EventPayload() any {
	return c.pld
}

func TestMatches(t *testing.T) {
	primary := byte(1)
	badPrimary := byte(2)
	index := uint32(5)
	badHigherIndex := uint32(6)
	badLowerIndex := index - 1
	sender := util.Uint160{1, 2, 3}
	signer := util.Uint160{4, 5, 6}
	contract := util.Uint160{7, 8, 9}
	badUint160 := util.Uint160{9, 9, 9}
	cnt := util.Uint256{1, 2, 3}
	badUint256 := util.Uint256{9, 9, 9}
	name := "ntf name"
	goodType := mempoolevent.TransactionAdded
	badType := mempoolevent.TransactionRemoved
	parameters, err := smartcontract.NewParametersFromValues(1, "2", []byte{3})
	require.NoError(t, err)
	badParameters, err := smartcontract.NewParametersFromValues([]byte{3}, "2", []byte{1})
	require.NoError(t, err)
	bContainer := testContainer{
		id: atipicialrpc.BlockEventID,
		pld: &block.Block{
			Header: block.Header{PrimaryIndex: byte(primary), Index: index},
		},
	}
	headerContainer := testContainer{
		id:  atipicialrpc.HeaderOfAddedBlockEventID,
		pld: &block.Header{PrimaryIndex: byte(primary), Index: index},
	}
	st := vmstate.Halt
	txContainer := testContainer{
		id:  atipicialrpc.TransactionEventID,
		pld: &transaction.Transaction{Signers: []transaction.Signer{{Account: sender}, {Account: signer}}},
	}
	ntfContainer := testContainer{
		id:  atipicialrpc.NotificationEventID,
		pld: &state.ContainedNotificationEvent{NotificationEvent: state.NotificationEvent{ScriptHash: contract, Name: name}},
	}
	ntfContainerParameters := testContainer{
		id: atipicialrpc.NotificationEventID,
		pld: &state.ContainedNotificationEvent{
			NotificationEvent: state.NotificationEvent{
				ScriptHash: contract,
				Name:       name,
				Item:       stackitem.NewArray(prmsToStack(t, parameters)),
			},
		},
	}
	exContainer := testContainer{
		id:  atipicialrpc.ExecutionEventID,
		pld: &state.AppExecResult{Container: cnt, Execution: state.Execution{VMState: st}},
	}
	ntrContainer := testContainer{
		id: atipicialrpc.NotaryRequestEventID,
		pld: &result.NotaryRequestEvent{
			Type: goodType,
			NotaryRequest: &payload.P2PNotaryRequest{
				MainTransaction:     &transaction.Transaction{Signers: []transaction.Signer{{Account: signer}}},
				FallbackTransaction: &transaction.Transaction{Signers: []transaction.Signer{{Account: util.Uint160{}}, {Account: sender}}},
			},
		},
	}
	mempoolContainer := testContainer{
		id: atipicialrpc.MempoolEventID,
		pld: &result.MempoolEvent{
			Type: goodType,
			Tx:   &transaction.Transaction{Signers: []transaction.Signer{{Account: sender}, {Account: signer}}},
		},
	}
	missedContainer := testContainer{
		id: atipicialrpc.MissedEventID,
	}
	var testCases = []struct {
		name       string
		comparator testComparator
		container  testContainer
		expected   bool
	}{
		{
			name:       "ID mismatch",
			comparator: testComparator{id: atipicialrpc.TransactionEventID},
			container:  bContainer,
			expected:   false,
		},
		{
			name:       "missed event",
			comparator: testComparator{id: atipicialrpc.BlockEventID},
			container:  missedContainer,
			expected:   false,
		},
		{
			name:       "block, no filter",
			comparator: testComparator{id: atipicialrpc.BlockEventID},
			container:  bContainer,
			expected:   true,
		},
		{
			name: "block, primary mismatch",
			comparator: testComparator{
				id:     atipicialrpc.BlockEventID,
				filter: atipicialrpc.BlockFilter{Primary: &badPrimary},
			},
			container: bContainer,
			expected:  false,
		},
		{
			name: "block, since mismatch",
			comparator: testComparator{
				id:     atipicialrpc.BlockEventID,
				filter: atipicialrpc.BlockFilter{Since: &badHigherIndex},
			},
			container: bContainer,
			expected:  false,
		},
		{
			name: "block, till mismatch",
			comparator: testComparator{
				id:     atipicialrpc.BlockEventID,
				filter: atipicialrpc.BlockFilter{Till: &badLowerIndex},
			},
			container: bContainer,
			expected:  false,
		},
		{
			name: "block, filter match",
			comparator: testComparator{
				id:     atipicialrpc.BlockEventID,
				filter: atipicialrpc.BlockFilter{Primary: &primary, Since: &index, Till: &index},
			},
			container: bContainer,
			expected:  true,
		},
		{
			name:       "header, no filter",
			comparator: testComparator{id: atipicialrpc.HeaderOfAddedBlockEventID},
			container:  headerContainer,
			expected:   true,
		},
		{
			name: "header, primary mismatch",
			comparator: testComparator{
				id:     atipicialrpc.HeaderOfAddedBlockEventID,
				filter: atipicialrpc.BlockFilter{Primary: &badPrimary},
			},
			container: headerContainer,
			expected:  false,
		},
		{
			name: "header, since mismatch",
			comparator: testComparator{
				id:     atipicialrpc.HeaderOfAddedBlockEventID,
				filter: atipicialrpc.BlockFilter{Since: &badHigherIndex},
			},
			container: headerContainer,
			expected:  false,
		},
		{
			name: "header, till mismatch",
			comparator: testComparator{
				id:     atipicialrpc.HeaderOfAddedBlockEventID,
				filter: atipicialrpc.BlockFilter{Till: &badLowerIndex},
			},
			container: headerContainer,
			expected:  false,
		},
		{
			name: "header, filter match",
			comparator: testComparator{
				id:     atipicialrpc.HeaderOfAddedBlockEventID,
				filter: atipicialrpc.BlockFilter{Primary: &primary, Since: &index, Till: &index},
			},
			container: headerContainer,
			expected:  true,
		},
		{
			name:       "transaction, no filter",
			comparator: testComparator{id: atipicialrpc.TransactionEventID},
			container:  txContainer,
			expected:   true,
		},
		{
			name: "transaction, sender mismatch",
			comparator: testComparator{
				id:     atipicialrpc.TransactionEventID,
				filter: atipicialrpc.TxFilter{Sender: &badUint160},
			},
			container: txContainer,
			expected:  false,
		},
		{
			name: "transaction, signer mismatch",
			comparator: testComparator{
				id:     atipicialrpc.TransactionEventID,
				filter: atipicialrpc.TxFilter{Signer: &badUint160},
			},
			container: txContainer,
			expected:  false,
		},
		{
			name: "transaction, filter match",
			comparator: testComparator{
				id:     atipicialrpc.TransactionEventID,
				filter: atipicialrpc.TxFilter{Sender: &sender, Signer: &signer},
			},
			container: txContainer,
			expected:  true,
		},
		{
			name:       "notification, no filter",
			comparator: testComparator{id: atipicialrpc.NotificationEventID},
			container:  ntfContainer,
			expected:   true,
		},
		{
			name: "notification, contract mismatch",
			comparator: testComparator{
				id:     atipicialrpc.NotificationEventID,
				filter: atipicialrpc.NotificationFilter{Contract: &badUint160},
			},
			container: ntfContainer,
			expected:  false,
		},
		{
			name: "notification, name mismatch",
			comparator: testComparator{
				id:     atipicialrpc.NotificationEventID,
				filter: atipicialrpc.NotificationFilter{Name: new("bad name")},
			},
			container: ntfContainer,
			expected:  false,
		},
		{
			name: "notification, filter match",
			comparator: testComparator{
				id:     atipicialrpc.NotificationEventID,
				filter: atipicialrpc.NotificationFilter{Name: &name, Contract: &contract},
			},
			container: ntfContainer,
			expected:  true,
		},
		{
			name: "notification, parameters match",
			comparator: testComparator{
				id:     atipicialrpc.NotificationEventID,
				filter: atipicialrpc.NotificationFilter{Name: &name, Contract: &contract, Parameters: parameters},
			},
			container: ntfContainerParameters,
			expected:  true,
		},
		{
			name: "notification, parameters mismatch",
			comparator: testComparator{
				id:     atipicialrpc.NotificationEventID,
				filter: atipicialrpc.NotificationFilter{Name: &name, Contract: &contract, Parameters: badParameters},
			},
			container: ntfContainerParameters,
			expected:  false,
		},
		{
			name:       "execution, no filter",
			comparator: testComparator{id: atipicialrpc.ExecutionEventID},
			container:  exContainer,
			expected:   true,
		},
		{
			name: "execution, state mismatch",
			comparator: testComparator{
				id:     atipicialrpc.ExecutionEventID,
				filter: atipicialrpc.ExecutionFilter{State: new("FAULT")},
			},
			container: exContainer,
			expected:  false,
		},
		{
			name: "execution, container mismatch",
			comparator: testComparator{
				id:     atipicialrpc.ExecutionEventID,
				filter: atipicialrpc.ExecutionFilter{Container: &badUint256},
			},
			container: exContainer,
			expected:  false,
		},
		{
			name: "execution, filter mismatch",
			comparator: testComparator{
				id:     atipicialrpc.ExecutionEventID,
				filter: atipicialrpc.ExecutionFilter{State: new(st.String()), Container: &cnt},
			},
			container: exContainer,
			expected:  true,
		},
		{
			name:       "notary request, no filter",
			comparator: testComparator{id: atipicialrpc.NotaryRequestEventID},
			container:  ntrContainer,
			expected:   true,
		},
		{
			name: "notary request, sender mismatch",
			comparator: testComparator{
				id:     atipicialrpc.NotaryRequestEventID,
				filter: atipicialrpc.NotaryRequestFilter{Sender: &badUint160},
			},
			container: ntrContainer,
			expected:  false,
		},
		{
			name: "notary request, signer mismatch",
			comparator: testComparator{
				id:     atipicialrpc.NotaryRequestEventID,
				filter: atipicialrpc.NotaryRequestFilter{Signer: &badUint160},
			},
			container: ntrContainer,
			expected:  false,
		},
		{
			name: "notary request, type mismatch",
			comparator: testComparator{
				id:     atipicialrpc.NotaryRequestEventID,
				filter: atipicialrpc.NotaryRequestFilter{Type: &badType},
			},
			container: ntrContainer,
			expected:  false,
		},
		{
			name: "notary request, filter match",
			comparator: testComparator{
				id:     atipicialrpc.NotaryRequestEventID,
				filter: atipicialrpc.NotaryRequestFilter{Sender: &sender, Signer: &signer, Type: &goodType},
			},
			container: ntrContainer,
			expected:  true,
		},
		{
			name:       "mempool event, no filter",
			comparator: testComparator{id: atipicialrpc.MempoolEventID},
			container:  mempoolContainer,
			expected:   true,
		},
		{
			name: "mempool event, sender mismatch",
			comparator: testComparator{
				id:     atipicialrpc.MempoolEventID,
				filter: atipicialrpc.MempoolEventFilter{Sender: &badUint160},
			},
			container: mempoolContainer,
			expected:  false,
		},
		{
			name: "mempool event, signer mismatch",
			comparator: testComparator{
				id:     atipicialrpc.MempoolEventID,
				filter: atipicialrpc.MempoolEventFilter{Signer: &badUint160},
			},
			container: mempoolContainer,
			expected:  false,
		},
		{
			name: "mempool event, type mismatch",
			comparator: testComparator{
				id:     atipicialrpc.MempoolEventID,
				filter: atipicialrpc.MempoolEventFilter{Type: &badType},
			},
			container: mempoolContainer,
			expected:  false,
		},
		{
			name: "mempool event, filter match",
			comparator: testComparator{
				id:     atipicialrpc.MempoolEventID,
				filter: atipicialrpc.MempoolEventFilter{Sender: &sender, Signer: &signer, Type: &goodType},
			},
			container: mempoolContainer,
			expected:  true,
		},
	}
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.expected, Matches(tc.comparator, tc.container))
		})
	}
}

func prmsToStack(t *testing.T, pp []smartcontract.Parameter) []stackitem.Item {
	res := make([]stackitem.Item, 0, len(pp))
	for _, p := range pp {
		s, err := p.ToStackItem()
		require.NoError(t, err)
		res = append(res, s)
	}
	return res
}
