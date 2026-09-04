package native

import (
	"github.com/atipicial/atipicial-go/pkg/config"
	"github.com/atipicial/atipicial-go/pkg/core/dao"
	"github.com/atipicial/atipicial-go/pkg/core/interop"
	"github.com/atipicial/atipicial-go/pkg/core/native/nativeids"
	"github.com/atipicial/atipicial-go/pkg/core/native/nativenames"
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/callflag"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
	"github.com/atipicial/atipicial-go/pkg/vm/stackitem"
)

// Treasury represents Treasury native contract.
type Treasury struct {
	interop.ContractMD
	ATC IATC
}

var _ interop.Contract = (*Treasury)(nil)

func newTreasury() *Treasury {
	t := &Treasury{
		ContractMD: *interop.NewContractMD(nativenames.Treasury, nativeids.Treasury, func(m *manifest.Manifest, hardfork config.Hardfork) {
			m.SupportedStandards = []string{manifest.AEP26StandardName, manifest.AEP27StandardName, manifest.AEP30StandardName}
		}),
	}
	defer t.BuildHFSpecificMD(t.ActiveIn())

	desc := NewDescriptor("verify", smartcontract.BoolType)
	md := NewMethodAndPrice(t.verify, 1<<5, callflag.ReadStates)
	t.AddMethod(md, desc)

	desc = NewDescriptor("onAEP11Payment", smartcontract.VoidType,
		manifest.NewParameter("from", smartcontract.Hash160Type),
		manifest.NewParameter("amount", smartcontract.IntegerType),
		manifest.NewParameter("tokenId", smartcontract.ByteArrayType),
		manifest.NewParameter("data", smartcontract.AnyType))
	md = NewMethodAndPrice(t.onAEP11Payment, 1<<5, callflag.NoneFlag)
	t.AddMethod(md, desc)

	desc = NewDescriptor("onAEP17Payment", smartcontract.VoidType,
		manifest.NewParameter("from", smartcontract.Hash160Type),
		manifest.NewParameter("amount", smartcontract.IntegerType),
		manifest.NewParameter("data", smartcontract.AnyType))
	md = NewMethodAndPrice(t.onAEP17Payment, 1<<5, callflag.NoneFlag)
	t.AddMethod(md, desc)

	return t
}

// OnPersist implements the [interop.Contract] interface.
func (t *Treasury) OnPersist(ic *interop.Context) error {
	return nil
}

// PostPersist implements the [interop.Contract] interface.
func (t *Treasury) PostPersist(ic *interop.Context) error {
	return nil
}

// Metadata implements the [interop.Contract] interface.
func (t *Treasury) Metadata() *interop.ContractMD {
	return &t.ContractMD
}

// Initialize implements the [interop.Contract] interface.
func (t *Treasury) Initialize(ic *interop.Context, hf *config.Hardfork, newMD *interop.HFSpecificContractMD) error {
	return nil
}

// InitializeCache implements the [interop.Contract] interface.
func (t *Treasury) InitializeCache(_ interop.IsHardforkEnabled, blockHeight uint32, d *dao.Simple) error {
	return nil
}

// ActiveIn implements the [interop.Contract] interface.
func (t *Treasury) ActiveIn() *config.Hardfork {
	return new(config.HFFaun)
}

func (t *Treasury) verify(ic *interop.Context, _ []stackitem.Item) stackitem.Item {
	return stackitem.NewBool(t.ATC.CheckCommittee(ic))
}

func (t *Treasury) onAEP11Payment(ic *interop.Context, args []stackitem.Item) stackitem.Item {
	// Arguments must be accessed and parsed to match the reference behaviour, but no real action is performed.
	from := toUint160(args[0])
	amount := toBigInt(args[1])
	tokenId := toBytes(args[2])
	data := args[3]

	var _, _, _, _ = from, amount, tokenId, data
	return stackitem.Null{}
}

func (t *Treasury) onAEP17Payment(ic *interop.Context, args []stackitem.Item) stackitem.Item {
	// Arguments must be accessed and parsed to match the reference behaviour, but no real action is performed.
	from := toUint160(args[0])
	amount := toBigInt(args[1])
	data := args[2]

	var _, _, _ = from, amount, data
	return stackitem.Null{}
}
