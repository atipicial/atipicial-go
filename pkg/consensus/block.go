package consensus

import (
	"errors"

	"github.com/github.com/atipicial/dbft"
	"github.com/atipicial/atipicial-go/pkg/config/netmode"
	coreb "github.com/atipicial/atipicial-go/pkg/core/block"
	"github.com/atipicial/atipicial-go/pkg/core/transaction"
	"github.com/atipicial/atipicial-go/pkg/crypto/keys"
	"github.com/atipicial/atipicial-go/pkg/util"
)

// atipicialBlock is a wrapper of a core.Block which implements
// methods necessary for dBFT library.
type atipicialBlock struct {
	coreb.Block

	network   netmode.Magic
	signature []byte
}

var _ dbft.Block[util.Uint256] = (*atipicialBlock)(nil)

// Sign implements the block.Block interface.
func (n *atipicialBlock) Sign(key dbft.PrivateKey) error {
	k := key.(*keys.PrivateKey)
	sig := k.SignHashable(uint32(n.network), &n.Block)
	n.signature = sig
	return nil
}

// Verify implements the block.Block interface.
func (n *atipicialBlock) Verify(key dbft.PublicKey, sign []byte) error {
	k := key.(*keys.PublicKey)
	if k.VerifyHashable(sign, uint32(n.network), &n.Block) {
		return nil
	}
	return errors.New("verification failed")
}

// Transactions implements the block.Block interface.
func (n *atipicialBlock) Transactions() []dbft.Transaction[util.Uint256] {
	txes := make([]dbft.Transaction[util.Uint256], len(n.Block.Transactions))
	for i, tx := range n.Block.Transactions {
		txes[i] = tx
	}

	return txes
}

// SetTransactions implements the block.Block interface.
func (n *atipicialBlock) SetTransactions(txes []dbft.Transaction[util.Uint256]) {
	n.Block.Transactions = make([]*transaction.Transaction, len(txes))
	for i, tx := range txes {
		n.Block.Transactions[i] = tx.(*transaction.Transaction)
	}
}

// PrevHash implements the block.Block interface.
func (n *atipicialBlock) PrevHash() util.Uint256 { return n.Block.PrevHash }

// MerkleRoot implements the block.Block interface.
func (n *atipicialBlock) MerkleRoot() util.Uint256 { return n.Block.MerkleRoot }

// Timestamp implements the block.Block interface.
func (n *atipicialBlock) Timestamp() uint64 { return n.Block.Timestamp * nsInMs }

// Index implements the block.Block interface.
func (n *atipicialBlock) Index() uint32 { return n.Block.Index }

// Signature implements the block.Block interface.
func (n *atipicialBlock) Signature() []byte { return n.signature }
