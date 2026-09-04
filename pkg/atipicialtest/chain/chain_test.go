package chain

import (
	"testing"

	"github.com/atipicial/atipicial-go/pkg/atipicialtest"
	"github.com/stretchr/testify/require"
)

// TestNewMulti checks that the transaction and the block are signed correctly for multi-node setup.
func TestNewMulti(t *testing.T) {
	bc, vAcc, cAcc := NewMulti(t)
	e := atipicialtest.NewExecutor(t, bc, vAcc, cAcc)

	require.NotEqual(t, vAcc.ScriptHash(), cAcc.ScriptHash())

	const amount = int64(10_0000_0000)

	c := e.CommitteeInvoker(bc.UtilityTokenHash()).WithSigners(vAcc)
	c.Invoke(t, true, "transfer", e.Validator.ScriptHash(), e.Committee.ScriptHash(), amount, nil)
}
