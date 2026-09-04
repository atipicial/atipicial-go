package runtime_test

import (
	"math"
	"testing"

	"github.com/atipicial/atipicial-go/pkg/core/interop/interopnames"
	"github.com/atipicial/atipicial-go/pkg/core/transaction"
	"github.com/atipicial/atipicial-go/pkg/io"
	"github.com/atipicial/atipicial-go/pkg/atipicialtest"
	"github.com/atipicial/atipicial-go/pkg/atipicialtest/chain"
	"github.com/atipicial/atipicial-go/pkg/vm/emit"
)

func TestRuntime_BurnGas(t *testing.T) {
	bc, acc := chain.NewSingle(t)
	e := atipicialtest.NewExecutor(t, bc, acc, acc)

	script := io.NewBufBinWriter()
	emit.Int(script.BinWriter, math.MaxInt64)
	emit.Syscall(script.BinWriter, interopnames.SystemRuntimeBurnGas)

	tx := transaction.New(script.Bytes(), 0)
	tx.Nonce = atipicialtest.Nonce()
	tx.ValidUntilBlock = e.Chain.BlockHeight() + 1
	e.SignTx(t, tx, 1_0000_0000, acc)
	e.AddNewBlock(t, tx)
	e.CheckFault(t, tx.Hash(), "System.Runtime.BurnGas failed: GAS limit exceeded")
}
