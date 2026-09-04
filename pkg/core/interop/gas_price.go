package interop

import (
	"github.com/atipicial/atipicial-go/pkg/core/fee"
	"github.com/atipicial/atipicial-go/pkg/vm/opcode"
)

// GetPrice returns a price for executing op with the provided parameter in
// picoGAS units.
func (ic *Context) GetPrice(op opcode.Opcode, parameter []byte) int64 {
	return fee.Opcode(ic.baseExecFee, op)
}
