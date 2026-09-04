package core_test

import (
	"math/big"
	"testing"

	"github.com/atipicial/atipicial-go/pkg/config"
	"github.com/atipicial/atipicial-go/pkg/core/dao"
	"github.com/atipicial/atipicial-go/pkg/core/interop"
	"github.com/atipicial/atipicial-go/pkg/core/native"
	"github.com/atipicial/atipicial-go/pkg/core/native/nativeids"
	"github.com/atipicial/atipicial-go/pkg/core/native/nativenames"
	"github.com/atipicial/atipicial-go/pkg/core/state"
	"github.com/atipicial/atipicial-go/pkg/core/transaction"
	"github.com/atipicial/atipicial-go/pkg/crypto/keys"
	"github.com/atipicial/atipicial-go/pkg/atipicialtest"
	"github.com/atipicial/atipicial-go/pkg/atipicialtest/chain"
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/callflag"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/atipicial/atipicial-go/pkg/vm/stackitem"
)

var (
	_ = (native.IGAS)(&gas{})
	_ = (native.IATC)(&atipicial{})
	_ = (native.IPolicy)(&policy{})
)

// A set of custom native contract stubs that are used to override default native
// contract implementations. These subs don't have any state and intentionally
// designed to be as simple as possible.
type (
	// gas is a custom native utility token contract implementation that doesn't track
	// any balance changes.
	gas struct {
		interop.ContractMD // obligatory field for proper Blockchain functioning.

		policy native.IPolicy // optional field that is presented here as an example of cross-native contract interaction.
	}
	// atipicial is a custom native governing token implementation that doesn't maintain any
	// balance changes and always use constant committee list.
	atipicial struct {
		interop.ContractMD

		validators keys.PublicKeys // optional field that is presented here as an example of state cache that may be required for contract functioning.
	}
	// policy is a custom native Policy contract implementation that doesn't have any
	// state and always use constant policies.
	policy struct {
		interop.ContractMD
	}
)

func newGAS() *gas {
	g := &gas{}
	defer g.BuildHFSpecificMD(nil) // obligatory call to properly initialize MD cache.

	g.ContractMD = *interop.NewContractMD( // obligatory call for proper allocate MD cache.
		nativenames.Gas,    // obligatory name for proper Blockchain functioning.
		nativeids.AtipicialDollar, // obligatory ID even if some default native contracts are missing.
		func(m *manifest.Manifest, hf config.Hardfork) {
			m.SupportedStandards = []string{manifest.AEP17StandardName} // not really for this stub, but let's pretend to show how to setup standards.
		})

	desc := native.NewDescriptor("getInt", smartcontract.IntegerType,
		manifest.NewParameter("someBoolParam", smartcontract.BoolType))
	md := native.NewMethodAndPrice(g.getInt, 1<<10, callflag.ReadStates)
	g.AddMethod(md, desc)

	eDesc := native.NewEventDescriptor("MyPrettyEvent",
		manifest.NewParameter("pubkey", smartcontract.PublicKeyType),
		manifest.NewParameter("someBool", smartcontract.BoolType),
	)
	eMD := native.NewEvent(eDesc)
	g.AddEvent(eMD)

	return g
}
func (g *gas) Initialize(*interop.Context, *config.Hardfork, *interop.HFSpecificContractMD) error {
	return nil
}
func (g *gas) ActiveIn() *config.Hardfork { return nil }
func (g *gas) InitializeCache(isHardforkEnabled interop.IsHardforkEnabled, blockHeight uint32, d *dao.Simple) error {
	return nil
}
func (g *gas) Metadata() *interop.ContractMD {
	return &g.ContractMD
}
func (g *gas) OnPersist(*interop.Context) error {
	return nil
}
func (g *gas) PostPersist(*interop.Context) error {
	return nil
}
func (g *gas) BalanceOf(d *dao.Simple, acc util.Uint160) *big.Int {
	return big.NewInt(1)
}
func (g *gas) Burn(ic *interop.Context, h util.Uint160, amount *big.Int) {
}
func (g *gas) MintDeferrable(ic *interop.Context, h util.Uint160, amount *big.Int, callOnPayment bool, continuation func()) {
}
func (g *gas) getInt(ic *interop.Context, args []stackitem.Item) stackitem.Item {
	// An example of how cross-native methods may be used:
	if g.policy.IsBlocked(ic.DAO, ic.Container.(*transaction.Transaction).Sender()) {
		panic("blocked sender")
	}

	return stackitem.NewBigInteger(big.NewInt(1))
}

func newATC(validator *keys.PublicKey) *atipicial {
	n := &atipicial{
		validators: keys.PublicKeys{validator},
	}
	defer n.BuildHFSpecificMD(nil)

	n.ContractMD = *interop.NewContractMD(nativenames.Atipicial, nativeids.AtipicialCoin)

	desc := native.NewDescriptor("getBool", smartcontract.BoolType)
	md := native.NewMethodAndPrice(n.getBool, 1<<10, callflag.ReadStates)
	n.AddMethod(md, desc)

	return n
}
func (n *atipicial) Initialize(*interop.Context, *config.Hardfork, *interop.HFSpecificContractMD) error {
	return nil
}
func (n *atipicial) ActiveIn() *config.Hardfork { return nil }
func (n *atipicial) InitializeCache(isHardforkEnabled interop.IsHardforkEnabled, blockHeight uint32, d *dao.Simple) error {
	return nil
}
func (n *atipicial) Metadata() *interop.ContractMD {
	return &n.ContractMD
}
func (n *atipicial) OnPersist(*interop.Context) error               { return nil }
func (n *atipicial) PostPersist(*interop.Context) error             { return nil }
func (n *atipicial) GetCommitteeAddress(d *dao.Simple) util.Uint160 { return util.Uint160{1, 2, 3} }
func (n *atipicial) GetNextBlockValidatorsInternal(d *dao.Simple) keys.PublicKeys {
	return n.validators
}
func (n *atipicial) BalanceOf(d *dao.Simple, acc util.Uint160) (*big.Int, uint32) {
	return big.NewInt(100500), 0
}
func (n *atipicial) CalculateBonus(ic *interop.Context, acc util.Uint160, endHeight uint32) (*big.Int, error) {
	return big.NewInt(5), nil
}
func (n *atipicial) GetCommitteeMembers(d *dao.Simple) keys.PublicKeys {
	return n.validators
}
func (n *atipicial) ComputeNextBlockValidators(d *dao.Simple) keys.PublicKeys {
	return n.validators
}
func (n *atipicial) GetCandidates(d *dao.Simple) ([]state.Validator, error) {
	return []state.Validator{
		{
			Key:   n.validators[0],
			Votes: big.NewInt(100),
		},
	}, nil
}
func (n *atipicial) CheckCommittee(ic *interop.Context) bool {
	return true
}
func (n *atipicial) CheckAlmostFullCommittee(ic *interop.Context) bool {
	return true
}
func (n *atipicial) getBool(ic *interop.Context, args []stackitem.Item) stackitem.Item {
	return stackitem.NewBool(true)
}
func (n *atipicial) RevokeVotesDeferrable(ic *interop.Context, acc util.Uint160, continuation func()) (bool, error) {
	return true, nil
}

func newPolicy() *policy {
	p := &policy{}
	defer p.BuildHFSpecificMD(nil)

	p.ContractMD = *interop.NewContractMD(nativenames.Policy, nativeids.PolicyContract)

	// Methods that are required for inter-contract interaction should follow below. It's
	// not obligatory to include all native.IPolicy methods to the contract manifest.
	desc := native.NewDescriptor("getTimePerBlock", smartcontract.IntegerType)
	md := native.NewMethodAndPrice(p.getTimePerBlock, 1<<10, callflag.ReadStates)
	p.AddMethod(md, desc)

	return p
}
func (p *policy) Initialize(*interop.Context, *config.Hardfork, *interop.HFSpecificContractMD) error {
	return nil
}
func (p *policy) ActiveIn() *config.Hardfork { return nil }
func (p *policy) InitializeCache(isHardforkEnabled interop.IsHardforkEnabled, blockHeight uint32, d *dao.Simple) error {
	return nil
}
func (p *policy) Metadata() *interop.ContractMD                                { return &p.ContractMD }
func (p *policy) OnPersist(*interop.Context) error                             { return nil }
func (p *policy) PostPersist(*interop.Context) error                           { return nil }
func (p *policy) GetStoragePriceInternal(d *dao.Simple) int64                  { return 5 }
func (p *policy) GetMaxVerificationGas(d *dao.Simple) int64                    { return 10_0000_0000 }
func (p *policy) GetExecFeeFactorInternal(d *dao.Simple) int64                 { return 1 }
func (p *policy) GetMaxTraceableBlocksInternal(d *dao.Simple) uint32           { return 100 }
func (p *policy) GetMillisecondsPerBlockInternal(d *dao.Simple) uint32         { return 1000 }
func (p *policy) GetMaxValidUntilBlockIncrementFromCache(d *dao.Simple) uint32 { return 2 }
func (p *policy) GetAttributeFeeInternal(d *dao.Simple, attrType transaction.AttrType) int64 {
	return 1
}
func (p *policy) CheckPolicy(d *dao.Simple, tx *transaction.Transaction) error { return nil }
func (p *policy) GetFeePerByteInternal(d *dao.Simple) int64                    { return 1 }
func (p *policy) BlockAccountInternalDeferrable(ic *interop.Context, hash util.Uint160, handleRes func(bool)) {
}
func (p *policy) IsBlocked(dao *dao.Simple, hash util.Uint160) bool                 { return false }
func (p *policy) WhitelistedFee(d *dao.Simple, hash util.Uint160, offset int) int64 { return -1 }
func (p *policy) CleanWhitelist(ic *interop.Context, h util.Uint160) error          { return nil }
func (p *policy) GetMaxValidUntilBlockIncrementInternal(ic *interop.Context) uint32 { return 2 }
func (p *policy) getTimePerBlock(ic *interop.Context, args []stackitem.Item) stackitem.Item {
	return stackitem.NewBigInteger(big.NewInt(1000))
}

func newCustomNatives(cfg config.ProtocolConfiguration) []interop.Contract {
	// Use default ContractManagement and Ledger implementations:
	mgmt := native.NewManagement()
	ledger := native.NewLedger()

	// Don't even create CryptoLib, StdLib, Notary, Oracle and Treasury contracts since they are useless.

	// Use custom AtipicialDollar, AtipicialCoin and Policy implementations:
	g := newGAS()
	pk, _ := keys.NewPrivateKey()
	n := newATC(pk.PublicKey())
	p := newPolicy()

	// An example of dependent native initialisation:
	g.policy = p

	// If default ContractManagement and Ledger implementations are used, then it's
	// mandatory to properly initialize their dependent natives:
	mgmt.ATC = n
	mgmt.Policy = p
	ledger.Policy = p

	// Use default RoleManagement implementation (and properly initialize links to dependent natives):
	desig := native.NewDesignate(cfg.Genesis.Roles)
	desig.ATC = n

	// If needed, some additional custom native may be added to the list of contracts. Use ID that
	// follows Notary contract ID.

	// Return the resulting list of native contracts that is different from the default one:
	return []interop.Contract{
		mgmt,
		ledger,
		n,
		g,
		p,
		desig,
	}
}

func TestBlockchain_CustomNatives(t *testing.T) {
	// Create chain with some natives.
	bc, acc := chain.NewSingleWithOptions(t, &chain.Options{
		NewNatives: newCustomNatives,
	})

	e := atipicialtest.NewExecutor(t, bc, acc, acc)
	e.AddNewBlock(t)
}
