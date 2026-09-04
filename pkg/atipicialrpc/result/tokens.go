package result

import (
	"github.com/atipicial/atipicial-go/pkg/util"
)

// AEP11Balances is a result for the getaep11balances RPC call.
type AEP11Balances struct {
	Balances []AEP11AssetBalance `json:"balance"`
	Address  string              `json:"address"`
}

// AEP11AssetBalance is a structure holding balance of a AEP-11 asset.
type AEP11AssetBalance struct {
	Asset    util.Uint160        `json:"assethash"`
	Decimals int                 `json:"decimals,string"`
	Name     string              `json:"name"`
	Symbol   string              `json:"symbol"`
	Tokens   []AEP11TokenBalance `json:"tokens"`
}

// AEP11TokenBalance represents balance of a single NFT.
type AEP11TokenBalance struct {
	ID          string `json:"tokenid"`
	Amount      string `json:"amount"`
	LastUpdated uint32 `json:"lastupdatedblock"`
}

// AEP17Balances is a result for the getaep17balances RPC call.
type AEP17Balances struct {
	Balances []AEP17Balance `json:"balance"`
	Address  string         `json:"address"`
}

// AEP17Balance represents balance for the single token contract.
type AEP17Balance struct {
	Asset       util.Uint160 `json:"assethash"`
	Amount      string       `json:"amount"`
	Decimals    int          `json:"decimals,string"`
	LastUpdated uint32       `json:"lastupdatedblock"`
	Name        string       `json:"name"`
	Symbol      string       `json:"symbol"`
}

// AEP11Transfers is a result for the getaep11transfers RPC.
type AEP11Transfers struct {
	Sent     []AEP11Transfer `json:"sent"`
	Received []AEP11Transfer `json:"received"`
	Address  string          `json:"address"`
}

// AEP11Transfer represents single AEP-11 transfer event.
type AEP11Transfer struct {
	Timestamp   uint64       `json:"timestamp"`
	Asset       util.Uint160 `json:"assethash"`
	Address     string       `json:"transferaddress,omitzero"`
	ID          string       `json:"tokenid"`
	Amount      string       `json:"amount"`
	Index       uint32       `json:"blockindex"`
	NotifyIndex uint32       `json:"transfernotifyindex"`
	TxHash      util.Uint256 `json:"txhash"`
}

// AEP17Transfers is a result for the getaep17transfers RPC.
type AEP17Transfers struct {
	Sent     []AEP17Transfer `json:"sent"`
	Received []AEP17Transfer `json:"received"`
	Address  string          `json:"address"`
}

// AEP17Transfer represents single AEP17 transfer event.
type AEP17Transfer struct {
	Timestamp   uint64       `json:"timestamp"`
	Asset       util.Uint160 `json:"assethash"`
	Address     string       `json:"transferaddress,omitzero"`
	Amount      string       `json:"amount"`
	Index       uint32       `json:"blockindex"`
	NotifyIndex uint32       `json:"transfernotifyindex"`
	TxHash      util.Uint256 `json:"txhash"`
}

// KnownAEP11Properties contains a list of well-known AEP-11 token property names.
var KnownAEP11Properties = map[string]bool{
	"description": true,
	"image":       true,
	"name":        true,
	"tokenURI":    true,
}
