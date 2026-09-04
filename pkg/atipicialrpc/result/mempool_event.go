package result

import (
	"github.com/atipicial/atipicial-go/pkg/core/mempoolevent"
	"github.com/atipicial/atipicial-go/pkg/core/transaction"
)

// MempoolEvent represents a transaction event either added to
// or removed from the mempool.
type MempoolEvent struct {
	Type mempoolevent.Type        `json:"type"`
	Tx   *transaction.Transaction `json:"transaction"`
}
