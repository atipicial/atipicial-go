package standard

import (
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
)

// Aep11Base is a Standard containing common AEP-11 methods.
var Aep11Base = &Standard{
	Base: DecimalTokenBase,
	Manifest: manifest.Manifest{
		ABI: manifest.ABI{
			Methods: []manifest.Method{
				{
					Name: "balanceOf",
					Parameters: []manifest.Parameter{
						{Name: "owner", Type: smartcontract.Hash160Type},
					},
					ReturnType: smartcontract.IntegerType,
					Safe:       true,
				},
				{
					Name: "tokensOf",
					Parameters: []manifest.Parameter{
						{Name: "owner", Type: smartcontract.Hash160Type},
					},
					ReturnType: smartcontract.InteropInterfaceType,
					Safe:       true,
				},
				{
					Name: "transfer",
					Parameters: []manifest.Parameter{
						{Name: "to", Type: smartcontract.Hash160Type},
						{Name: "tokenId", Type: smartcontract.ByteArrayType},
						{Name: "data", Type: smartcontract.AnyType},
					},
					ReturnType: smartcontract.BoolType,
				},
			},
			Events: []manifest.Event{
				{
					Name: "Transfer",
					Parameters: []manifest.Parameter{
						{Name: "from", Type: smartcontract.Hash160Type},
						{Name: "to", Type: smartcontract.Hash160Type},
						{Name: "amount", Type: smartcontract.IntegerType},
						{Name: "tokenId", Type: smartcontract.ByteArrayType},
					},
				},
			},
		},
	},
	Optional: []manifest.Method{
		{
			Name: "properties",
			Parameters: []manifest.Parameter{
				{Name: "tokenId", Type: smartcontract.ByteArrayType},
			},
			ReturnType: smartcontract.MapType,
			Safe:       true,
		},
		{
			Name:       "tokens",
			Parameters: []manifest.Parameter{}, // nil value implies no parameters check.
			ReturnType: smartcontract.InteropInterfaceType,
			Safe:       true,
		},
	},
}

// Aep11NonDivisible is a AEP-11 non-divisible Standard.
var Aep11NonDivisible = &Standard{
	Base: Aep11Base,
	Manifest: manifest.Manifest{
		ABI: manifest.ABI{
			Methods: []manifest.Method{
				{
					Name: "ownerOf",
					Parameters: []manifest.Parameter{
						{Name: "tokenId", Type: smartcontract.ByteArrayType},
					},
					ReturnType: smartcontract.Hash160Type,
					Safe:       true,
				},
			},
		},
	},
}

// Aep11Divisible is a AEP-11 divisible Standard.
var Aep11Divisible = &Standard{
	Base: Aep11Base,
	Manifest: manifest.Manifest{
		ABI: manifest.ABI{
			Methods: []manifest.Method{
				{
					Name: "balanceOf",
					Parameters: []manifest.Parameter{
						{Name: "owner", Type: smartcontract.Hash160Type},
						{Name: "tokenId", Type: smartcontract.ByteArrayType},
					},
					ReturnType: smartcontract.IntegerType,
					Safe:       true,
				},
				{
					Name: "ownerOf",
					Parameters: []manifest.Parameter{
						{Name: "tokenId", Type: smartcontract.ByteArrayType},
					},
					ReturnType: smartcontract.InteropInterfaceType, // iterator
					Safe:       true,
				},
				{
					Name: "transfer",
					Parameters: []manifest.Parameter{
						{Name: "from", Type: smartcontract.Hash160Type},
						{Name: "to", Type: smartcontract.Hash160Type},
						{Name: "amount", Type: smartcontract.IntegerType},
						{Name: "tokenId", Type: smartcontract.ByteArrayType},
						{Name: "data", Type: smartcontract.AnyType},
					},
					ReturnType: smartcontract.BoolType,
				},
			},
		},
	},
}
