package rpcsrv

import (
	"github.com/atipicial/atipicial-go/pkg/atipicialrpc/result"
)

// tokenTransfers is a generic type used to represent AEP-11 and AEP-17 transfers.
type tokenTransfers struct {
	Sent     []any  `json:"sent"`
	Received []any  `json:"received"`
	Address  string `json:"address"`
}

// aep17TransferToAEP11 adds an ID to the provided AEP-17 transfer and returns a new
// AEP-11 structure.
func aep17TransferToAEP11(t17 *result.AEP17Transfer, id string) result.AEP11Transfer {
	return result.AEP11Transfer{
		Timestamp:   t17.Timestamp,
		Asset:       t17.Asset,
		Address:     t17.Address,
		ID:          id,
		Amount:      t17.Amount,
		Index:       t17.Index,
		NotifyIndex: t17.NotifyIndex,
		TxHash:      t17.TxHash,
	}
}
