package royalty

import (
	"github.com/atipicial/atipicial-go/pkg/interop"
	"github.com/atipicial/atipicial-go/pkg/interop/native/std"
	"github.com/atipicial/atipicial-go/pkg/interop/runtime"
)

// RoyaltiesTransferred notifies about royalty payment. This method is called by marketplace
// contract when royalties are transferred.
func RoyaltiesTransferred(royaltyToken, royaltyRecipient, buyer interop.Hash160, tokenId []byte, amount int) {
	runtime.Notify("RoyaltiesTransferred", royaltyToken, royaltyRecipient, buyer, std.Deserialize(tokenId).([]byte), amount)
}
