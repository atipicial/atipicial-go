package standard

import (
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
)

// Aep29 is a AEP-29 Standard describing smart contract _deploy method functionality.
var Aep29 = &Standard{
	Manifest: manifest.Manifest{
		ABI: manifest.ABI{
			Methods: []manifest.Method{
				{
					Name: "_deploy",
					Parameters: []manifest.Parameter{
						{Name: "data", Type: smartcontract.AnyType},
						{Name: "update", Type: smartcontract.BoolType},
					},
					ReturnType: smartcontract.VoidType,
				},
			},
		},
	},
}
