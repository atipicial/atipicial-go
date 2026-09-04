package crypto

import (
	"github.com/atipicial/atipicial-go/pkg/core/interop"
	"github.com/atipicial/atipicial-go/pkg/core/interop/interopnames"
)

var (
	atipicialCryptoCheckMultisigID = interopnames.ToID([]byte(interopnames.SystemCryptoCheckMultisig))
	atipicialCryptoCheckSigID      = interopnames.ToID([]byte(interopnames.SystemCryptoCheckSig))
)

// Interops represents sorted crypto-related interop functions.
var Interops = []interop.Function{
	{ID: atipicialCryptoCheckMultisigID, Func: ECDSASecp256r1CheckMultisig},
	{ID: atipicialCryptoCheckSigID, Func: ECDSASecp256r1CheckSig},
}

func init() {
	interop.Sort(Interops)
}
