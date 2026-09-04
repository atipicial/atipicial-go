package standard

import (
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
)

// MethodRoyaltyInfo is the name of the method that returns royalty information.
const MethodRoyaltyInfo = "royaltyInfo"

// Aep24 is a AEP-24 Standard for NFT royalties.
var Aep24 = &Standard{
	Manifest: manifest.Manifest{
		ABI: manifest.ABI{
			Methods: []manifest.Method{
				{
					Name: MethodRoyaltyInfo,
					Parameters: []manifest.Parameter{
						{Name: "tokenId", Type: smartcontract.ByteArrayType},
						{Name: "royaltyToken", Type: smartcontract.Hash160Type},
						{Name: "salePrice", Type: smartcontract.IntegerType},
					},
					ReturnType: smartcontract.ArrayType,
					Safe:       true,
				},
			},
		},
	},
	Required: []string{manifest.AEP11StandardName},
}

// Aep24Payable contains an event that MUST be triggered after marketplaces
// transferring royalties to the royalty recipient if royaltyInfo method is implemented.
var Aep24Payable = &Standard{
	Manifest: manifest.Manifest{
		ABI: manifest.ABI{
			Events: []manifest.Event{
				{
					Name: "RoyaltiesTransferred",
					Parameters: []manifest.Parameter{
						{Name: "royaltyToken", Type: smartcontract.Hash160Type},
						{Name: "royaltyRecipient", Type: smartcontract.Hash160Type},
						{Name: "buyer", Type: smartcontract.Hash160Type},
						{Name: "tokenId", Type: smartcontract.ByteArrayType},
						{Name: "amount", Type: smartcontract.IntegerType},
					},
				},
			},
		},
	},
}
