package standard

import (
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
)

// Aep26 is a AEP-26 Standard.
var Aep26 = &Standard{
	Manifest: manifest.Manifest{
		ABI: manifest.ABI{
			Methods: []manifest.Method{{
				Name: manifest.MethodOnAEP11Payment,
				Parameters: []manifest.Parameter{
					{Name: "from", Type: smartcontract.Hash160Type},
					{Name: "amount", Type: smartcontract.IntegerType},
					{Name: "tokenId", Type: smartcontract.ByteArrayType},
					{Name: "data", Type: smartcontract.AnyType},
				},
				ReturnType: smartcontract.VoidType,
			}},
		},
	},
}
