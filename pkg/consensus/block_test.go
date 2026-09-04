package consensus

import (
	"testing"

	"github.com/github.com/atipicial/dbft"
	"github.com/atipicial/atipicial-go/pkg/core/transaction"
	"github.com/atipicial/atipicial-go/pkg/crypto/keys"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/atipicial/atipicial-go/pkg/vm/opcode"
	"github.com/stretchr/testify/require"
)

func TestAtipicialBlock_Sign(t *testing.T) {
	b := new(atipicialBlock)
	priv, _ := keys.NewPrivateKey()

	require.NoError(t, b.Sign(priv))
	require.NoError(t, b.Verify(priv.PublicKey(), b.Signature()))
}

func TestAtipicialBlock_Setters(t *testing.T) {
	b := new(atipicialBlock)

	b.Block.Index = 12
	require.EqualValues(t, 12, b.Index())

	b.Block.Timestamp = 777
	// 777ms -> 777000000ns
	require.EqualValues(t, 777000000, b.Timestamp())

	b.Block.MerkleRoot = util.Uint256{1, 2, 3, 4}
	require.Equal(t, util.Uint256{1, 2, 3, 4}, b.MerkleRoot())

	b.Block.PrevHash = util.Uint256{9, 8, 7}
	require.Equal(t, util.Uint256{9, 8, 7}, b.PrevHash())

	txx := []dbft.Transaction[util.Uint256]{transaction.New([]byte{byte(opcode.PUSH1)}, 1)}
	b.SetTransactions(txx)
	require.Equal(t, txx, b.Transactions())
}
