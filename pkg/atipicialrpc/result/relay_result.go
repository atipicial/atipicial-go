package result

import "github.com/atipicial/atipicial-go/pkg/util"

// RelayResult ia a result of `sendrawtransaction` or `submitblock` RPC calls.
type RelayResult struct {
	Hash util.Uint256 `json:"hash"`
}
