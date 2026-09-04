package rpcevent

import (
	"github.com/atipicial/atipicial-go/pkg/core/block"
	"github.com/atipicial/atipicial-go/pkg/core/state"
	"github.com/atipicial/atipicial-go/pkg/core/transaction"
	"github.com/atipicial/atipicial-go/pkg/atipicialrpc"
	"github.com/atipicial/atipicial-go/pkg/atipicialrpc/result"
	"github.com/atipicial/atipicial-go/pkg/vm/stackitem"
)

type (
	// Comparator is an interface required from notification event filter to be able to
	// filter notifications.
	Comparator interface {
		EventID() atipicialrpc.EventID
		Filter() atipicialrpc.SubscriptionFilter
	}
	// Container is an interface required from notification event to be able to
	// pass filter.
	Container interface {
		EventID() atipicialrpc.EventID
		EventPayload() any
	}
)

// Matches filters our given Container against Comparator filter.
func Matches(f Comparator, r Container) bool {
	expectedEvent := f.EventID()
	filter := f.Filter()
	if r.EventID() != expectedEvent {
		return false
	}
	if filter == nil {
		return true
	}
	switch f.EventID() {
	case atipicialrpc.BlockEventID, atipicialrpc.HeaderOfAddedBlockEventID:
		filt := filter.(atipicialrpc.BlockFilter)
		var b *block.Header
		if f.EventID() == atipicialrpc.HeaderOfAddedBlockEventID {
			b = r.EventPayload().(*block.Header)
		} else {
			b = &r.EventPayload().(*block.Block).Header
		}
		primaryOk := filt.Primary == nil || *filt.Primary == b.PrimaryIndex
		sinceOk := filt.Since == nil || *filt.Since <= b.Index
		tillOk := filt.Till == nil || b.Index <= *filt.Till
		return primaryOk && sinceOk && tillOk
	case atipicialrpc.TransactionEventID:
		filt := filter.(atipicialrpc.TxFilter)
		tx := r.EventPayload().(*transaction.Transaction)
		senderOK := filt.Sender == nil || tx.Sender().Equals(*filt.Sender)
		signerOK := true
		if filt.Signer != nil {
			signerOK = false
			for i := range tx.Signers {
				if tx.Signers[i].Account.Equals(*filt.Signer) {
					signerOK = true
					break
				}
			}
		}
		return senderOK && signerOK
	case atipicialrpc.NotificationEventID:
		filt := filter.(atipicialrpc.NotificationFilter)
		notification := r.EventPayload().(*state.ContainedNotificationEvent)
		hashOk := filt.Contract == nil || notification.ScriptHash.Equals(*filt.Contract)
		nameOk := filt.Name == nil || notification.Name == *filt.Name
		parametersOk := true
		if len(filt.Parameters) > 0 {
			stackItems := notification.Item.Value().([]stackitem.Item)
			parameters, err := filt.ParametersAsStackItems()
			if err != nil {
				return false
			}
			if len(parameters) > len(stackItems) {
				return false
			}
			for i, p := range parameters {
				if p.Type() == stackitem.AnyT && p.Value() == nil {
					continue
				}
				if !p.Equals(stackItems[i]) {
					parametersOk = false
					break
				}
			}
		}
		return hashOk && nameOk && parametersOk
	case atipicialrpc.ExecutionEventID:
		filt := filter.(atipicialrpc.ExecutionFilter)
		applog := r.EventPayload().(*state.AppExecResult)
		stateOK := filt.State == nil || applog.VMState.String() == *filt.State
		containerOK := filt.Container == nil || applog.Container.Equals(*filt.Container)
		return stateOK && containerOK
	case atipicialrpc.NotaryRequestEventID:
		filt := filter.(atipicialrpc.NotaryRequestFilter)
		req := r.EventPayload().(*result.NotaryRequestEvent)
		typeOk := filt.Type == nil || req.Type == *filt.Type
		senderOk := filt.Sender == nil || req.NotaryRequest.FallbackTransaction.Signers[1].Account == *filt.Sender
		signerOK := true
		if filt.Signer != nil {
			signerOK = false
			for _, signer := range req.NotaryRequest.MainTransaction.Signers {
				if signer.Account.Equals(*filt.Signer) {
					signerOK = true
					break
				}
			}
		}
		return senderOk && signerOK && typeOk
	case atipicialrpc.MempoolEventID:
		filt := filter.(atipicialrpc.MempoolEventFilter)
		memEvent := r.EventPayload().(*result.MempoolEvent)
		if filt.Type != nil && memEvent.Type != *filt.Type {
			return false
		}
		if filt.Sender != nil && !memEvent.Tx.Sender().Equals(*filt.Sender) {
			return false
		}
		if filt.Signer != nil {
			for _, signer := range memEvent.Tx.Signers {
				if signer.Account.Equals(*filt.Signer) {
					return true
				}
			}
			return false
		}
		return true
	default:
		return false
	}
}
