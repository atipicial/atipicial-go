package native

import (
	"errors"
	"math/big"

	"github.com/atipicial/atipicial-go/pkg/config"
	"github.com/atipicial/atipicial-go/pkg/core/dao"
	"github.com/atipicial/atipicial-go/pkg/core/interop"
	"github.com/atipicial/atipicial-go/pkg/core/native/nativeids"
	"github.com/atipicial/atipicial-go/pkg/core/native/nativenames"
	"github.com/atipicial/atipicial-go/pkg/core/state"
	"github.com/atipicial/atipicial-go/pkg/core/transaction"
	"github.com/atipicial/atipicial-go/pkg/crypto/hash"
	"github.com/atipicial/atipicial-go/pkg/crypto/keys"
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/util"
)

// GAS represents GAS native contract.
type GAS struct {
	aep17TokenNative
	ATC    IATC
	Policy IPolicy

	initialSupply int64
}

// GASFactor is a divisor for finding GAS integral value.
const GASFactor = ATCTotalSupply

// newGAS returns GAS native contract.
func newGAS(init int64) *GAS {
	g := &GAS{
		initialSupply: init,
	}
	defer g.BuildHFSpecificMD(g.ActiveIn())

	aep17 := newAEP17Native(nativenames.Gas, nativeids.AtipicialDollar, nil)
	aep17.symbol = "GAS"
	aep17.decimals = 8
	aep17.factor = GASFactor
	aep17.incBalance = g.increaseBalance
	aep17.balFromBytes = g.balanceFromBytes

	g.aep17TokenNative = *aep17

	return g
}

func (g *GAS) increaseBalance(_ *interop.Context, _ util.Uint160, si *state.StorageItem, amount *big.Int, checkBal *big.Int) (*gasDistribution, error) {
	acc, err := state.AEP17BalanceFromBytes(*si)
	if err != nil {
		return nil, err
	}
	if sign := amount.Sign(); sign == 0 {
		// Requested self-transfer amount can be higher than actual balance.
		if checkBal != nil && acc.Balance.Cmp(checkBal) < 0 {
			err = errors.New("insufficient funds")
		}
		return nil, err
	} else if sign == -1 && acc.Balance.CmpAbs(amount) == -1 {
		return nil, errors.New("insufficient funds")
	}
	acc.Balance.Add(&acc.Balance, amount)
	if acc.Balance.Sign() != 0 {
		*si = acc.Bytes(nil)
	} else {
		*si = nil
	}
	return nil, nil
}

func (g *GAS) balanceFromBytes(si *state.StorageItem) (*big.Int, error) {
	acc, err := state.AEP17BalanceFromBytes(*si)
	if err != nil {
		return nil, err
	}
	return &acc.Balance, err
}

// Initialize initializes a GAS contract.
func (g *GAS) Initialize(ic *interop.Context, hf *config.Hardfork, newMD *interop.HFSpecificContractMD) error {
	if hf != g.ActiveIn() {
		return nil
	}

	if err := g.aep17TokenNative.Initialize(ic); err != nil {
		return err
	}
	_, totalSupply := g.getTotalSupply(ic.DAO)
	if totalSupply.Sign() != 0 {
		return errors.New("already initialized")
	}
	h, err := getStandbyValidatorsHash(ic)
	if err != nil {
		return err
	}
	g.MintDeferrable(ic, h, big.NewInt(g.initialSupply), false, func() { /*no continuation*/ })
	return nil
}

// InitializeCache implements the Contract interface.
func (g *GAS) InitializeCache(_ interop.IsHardforkEnabled, blockHeight uint32, d *dao.Simple) error {
	return nil
}

// OnPersist implements the Contract interface.
func (g *GAS) OnPersist(ic *interop.Context) error {
	if len(ic.Block.Transactions) == 0 {
		return nil
	}
	for _, tx := range ic.Block.Transactions {
		absAmount := big.NewInt(tx.SystemFee + tx.NetworkFee)
		g.Burn(ic, tx.Sender(), absAmount)
	}
	validators := g.ATC.GetNextBlockValidatorsInternal(ic.DAO)
	primary := validators[ic.Block.PrimaryIndex].GetScriptHash()
	var netFee int64
	for _, tx := range ic.Block.Transactions {
		netFee += tx.NetworkFee
		// Reward for NotaryAssisted attribute will be minted to designated notary nodes
		// by Notary contract.
		attrs := tx.GetAttributes(transaction.NotaryAssistedT)
		if len(attrs) != 0 {
			na := attrs[0].Value.(*transaction.NotaryAssisted)
			netFee -= (int64(na.NKeys) + 1) * g.Policy.GetAttributeFeeInternal(ic.DAO, transaction.NotaryAssistedT)
		}
	}
	g.MintDeferrable(ic, primary, big.NewInt(int64(netFee)), false, func() { /*no continuation*/ })
	return nil
}

// PostPersist implements the Contract interface.
func (g *GAS) PostPersist(ic *interop.Context) error {
	return nil
}

// ActiveIn implements the Contract interface.
func (g *GAS) ActiveIn() *config.Hardfork {
	return nil
}

// BalanceOf returns native GAS token balance for the acc.
func (g *GAS) BalanceOf(d *dao.Simple, acc util.Uint160) *big.Int {
	return g.balanceOfInternal(d, acc)
}

func getStandbyValidatorsHash(ic *interop.Context) (util.Uint160, error) {
	cfg := ic.Chain.GetConfig()
	committee, err := keys.NewPublicKeysFromStrings(cfg.StandbyCommittee)
	if err != nil {
		return util.Uint160{}, err
	}
	s, err := smartcontract.CreateDefaultMultiSigRedeemScript(committee[:cfg.GetNumOfCNs(0)])
	if err != nil {
		return util.Uint160{}, err
	}
	return hash.Hash160(s), nil
}
