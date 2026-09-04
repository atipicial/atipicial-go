package standard

import (
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
)

// Aep30 is a AEP-30 Standard describing smart contract verify method functionality.
var Aep30 = &Standard{
	Manifest: manifest.Manifest{
		ABI: manifest.ABI{
			Methods: []manifest.Method{
				{
					// nil parameters implies no parameters check, AEP-30 allows an
					// undefined number of parameters for verify.
					Name:       "verify",
					Safe:       true,
					ReturnType: smartcontract.BoolType,
				},
			},
		},
	},
}
