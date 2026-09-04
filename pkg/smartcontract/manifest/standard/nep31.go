package standard

import (
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
)

// Aep31 is a AEP-31 Standard describing smart contract destroy functionality.
var Aep31 = &Standard{
	Manifest: manifest.Manifest{
		ABI: manifest.ABI{
			Methods: []manifest.Method{
				{
					Name:       "destroy",
					Parameters: []manifest.Parameter{},
					ReturnType: smartcontract.VoidType,
				},
			},
		},
	},
}
