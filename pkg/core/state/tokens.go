package state

import (
	"bytes"
	"math/big"
	"slices"

	"github.com/atipicial/atipicial-go/pkg/config/limits"
	"github.com/atipicial/atipicial-go/pkg/encoding/bigint"
	"github.com/atipicial/atipicial-go/pkg/io"
	"github.com/atipicial/atipicial-go/pkg/util"
)

// TokenTransferBatchSize is the maximum number of entries for TokenTransferLog.
const TokenTransferBatchSize = 128

// TokenTransferLog is a serialized log of token transfers.
type TokenTransferLog struct {
	Raw []byte
	buf *bytes.Buffer
	iow *io.BinWriter
}

// AEP17Transfer represents a single AEP-17 Transfer event.
type AEP17Transfer struct {
	// Asset is a AEP-17 contract ID.
	Asset int32
	// Counterparty is the address of the sender/receiver (the other side of the transfer).
	Counterparty util.Uint160
	// Amount is the amount of tokens transferred.
	// It is negative when tokens are sent and positive if they are received.
	Amount *big.Int
	// Block is a number of block when the event occurred.
	Block uint32
	// Timestamp is the timestamp of the block where transfer occurred.
	Timestamp uint64
	// Tx is a hash the transaction.
	Tx util.Uint256
}

// AEP11Transfer represents a single AEP-11 Transfer event.
type AEP11Transfer struct {
	AEP17Transfer

	// ID is a AEP-11 token ID.
	ID []byte
}

// TokenTransferInfo stores a map of the contract IDs to the balance's last updated
// block trackers along with the information about AEP-17 and AEP-11 transfer batch.
type TokenTransferInfo struct {
	LastUpdated map[int32]uint32
	// NextAEP11Batch stores the index of the next AEP-11 transfer batch.
	NextAEP11Batch uint32
	// NextAEP17Batch stores the index of the next AEP-17 transfer batch.
	NextAEP17Batch uint32
	// NextAEP11NewestTimestamp stores the block timestamp of the first AEP-11 transfer in raw.
	NextAEP11NewestTimestamp uint64
	// NextAEP17NewestTimestamp stores the block timestamp of the first AEP-17 transfer in raw.
	NextAEP17NewestTimestamp uint64
	// NewAEP11Batch is true if batch with the `NextAEP11Batch` index should be created.
	NewAEP11Batch bool
	// NewAEP17Batch is true if batch with the `NextAEP17Batch` index should be created.
	NewAEP17Batch bool
}

// NewTokenTransferInfo returns new TokenTransferInfo.
func NewTokenTransferInfo() *TokenTransferInfo {
	return &TokenTransferInfo{
		NewAEP11Batch: true,
		NewAEP17Batch: true,
		LastUpdated:   make(map[int32]uint32),
	}
}

// DecodeBinary implements the io.Serializable interface.
func (bs *TokenTransferInfo) DecodeBinary(r *io.BinReader) {
	bs.NextAEP11Batch = r.ReadU32LE()
	bs.NextAEP17Batch = r.ReadU32LE()
	bs.NextAEP11NewestTimestamp = r.ReadU64LE()
	bs.NextAEP17NewestTimestamp = r.ReadU64LE()
	bs.NewAEP11Batch = r.ReadBool()
	bs.NewAEP17Batch = r.ReadBool()
	lenBalances := r.ReadVarUint()
	m := make(map[int32]uint32, lenBalances)
	for range lenBalances {
		key := int32(r.ReadU32LE())
		m[key] = r.ReadU32LE()
	}
	bs.LastUpdated = m
}

// EncodeBinary implements the io.Serializable interface.
func (bs *TokenTransferInfo) EncodeBinary(w *io.BinWriter) {
	w.WriteU32LE(bs.NextAEP11Batch)
	w.WriteU32LE(bs.NextAEP17Batch)
	w.WriteU64LE(bs.NextAEP11NewestTimestamp)
	w.WriteU64LE(bs.NextAEP17NewestTimestamp)
	w.WriteBool(bs.NewAEP11Batch)
	w.WriteBool(bs.NewAEP17Batch)
	w.WriteVarUint(uint64(len(bs.LastUpdated)))
	for k, v := range bs.LastUpdated {
		w.WriteU32LE(uint32(k))
		w.WriteU32LE(v)
	}
}

// Append appends a single transfer to a log.
func (lg *TokenTransferLog) Append(tr io.Serializable) error {
	// The first entry, set up counter.
	if len(lg.Raw) == 0 {
		lg.Raw = append(lg.Raw, 0)
	}

	if lg.buf == nil {
		lg.buf = bytes.NewBuffer(lg.Raw)
	}
	if lg.iow == nil {
		lg.iow = io.NewBinWriterFromIO(lg.buf)
	}

	tr.EncodeBinary(lg.iow)
	if lg.iow.Err != nil {
		return lg.iow.Err
	}
	lg.Raw = lg.buf.Bytes()
	lg.Raw[0]++
	return nil
}

// Reset resets the state of the log, clearing all entries, but keeping existing
// buffer for future writes.
func (lg *TokenTransferLog) Reset() {
	lg.Raw = lg.Raw[:0]
	lg.buf = nil
	lg.iow = nil
}

// ForEachAEP11 iterates over a transfer log returning on the first error.
func (lg *TokenTransferLog) ForEachAEP11(f func(*AEP11Transfer) (bool, error)) (bool, error) {
	if lg == nil || len(lg.Raw) == 0 {
		return true, nil
	}
	transfers := make([]AEP11Transfer, lg.Size())
	r := io.NewBinReaderFromBuf(lg.Raw[1:])
	for i := range transfers {
		transfers[i].DecodeBinary(r)
	}
	if r.Err != nil {
		return false, r.Err
	}
	for i := range slices.Backward(transfers) {
		cont, err := f(&transfers[i])
		if err != nil || !cont {
			return false, err
		}
	}
	return true, nil
}

// ForEachAEP17 iterates over a transfer log returning on the first error.
func (lg *TokenTransferLog) ForEachAEP17(f func(*AEP17Transfer) (bool, error)) (bool, error) {
	if lg == nil || len(lg.Raw) == 0 {
		return true, nil
	}
	transfers := make([]AEP17Transfer, lg.Size())
	r := io.NewBinReaderFromBuf(lg.Raw[1:])
	for i := range transfers {
		transfers[i].DecodeBinary(r)
	}
	if r.Err != nil {
		return false, r.Err
	}
	for i := range slices.Backward(transfers) {
		cont, err := f(&transfers[i])
		if err != nil || !cont {
			return false, err
		}
	}
	return true, nil
}

// Size returns the amount of the transfer written in the log.
func (lg *TokenTransferLog) Size() int {
	if len(lg.Raw) == 0 {
		return 0
	}
	return int(lg.Raw[0])
}

// EncodeBinary implements the io.Serializable interface.
func (t *AEP17Transfer) EncodeBinary(w *io.BinWriter) {
	var buf [bigint.MaxBytesLen]byte

	w.WriteU32LE(uint32(t.Asset))
	w.WriteBytes(t.Tx[:])
	w.WriteBytes(t.Counterparty[:])
	w.WriteU32LE(t.Block)
	w.WriteU64LE(t.Timestamp)
	amount := bigint.ToPreallocatedBytes(t.Amount, buf[:])
	w.WriteVarBytes(amount)
}

// DecodeBinary implements the io.Serializable interface.
func (t *AEP17Transfer) DecodeBinary(r *io.BinReader) {
	t.Asset = int32(r.ReadU32LE())
	r.ReadBytes(t.Tx[:])
	r.ReadBytes(t.Counterparty[:])
	t.Block = r.ReadU32LE()
	t.Timestamp = r.ReadU64LE()
	amount := r.ReadVarBytes(bigint.MaxBytesLen)
	t.Amount = bigint.FromBytes(amount)
}

// EncodeBinary implements the io.Serializable interface.
func (t *AEP11Transfer) EncodeBinary(w *io.BinWriter) {
	t.AEP17Transfer.EncodeBinary(w)
	w.WriteVarBytes(t.ID)
}

// DecodeBinary implements the io.Serializable interface.
func (t *AEP11Transfer) DecodeBinary(r *io.BinReader) {
	t.AEP17Transfer.DecodeBinary(r)
	t.ID = r.ReadVarBytes(limits.MaxStorageKeyLen)
}
