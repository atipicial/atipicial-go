package standard

import (
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
)

// Aep22 is a AEP-22 Standard describing smart contract update functionality.
var Aep22 = &Standard{
	Manifest: manifest.Manifest{
		ABI: manifest.ABI{
			Methods: []manifest.Method{
				{
					Name: "update",
					Parameters: []manifest.Parameter{
						{Name: "aefFile", Type: smartcontract.ByteArrayType},
						{Name: "manifest", Type: smartcontract.ByteArrayType},
						{Name: "data", Type: smartcontract.AnyType},
					},
					ReturnType: smartcontract.VoidType,
				},
			},
		},
	},
}
