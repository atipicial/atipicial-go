package native

import (
	"math/big"
	"testing"

	"github.com/atipicial/atipicial-go/pkg/config"
	"github.com/atipicial/atipicial-go/pkg/core/dao"
	"github.com/atipicial/atipicial-go/pkg/core/interop"
	"github.com/atipicial/atipicial-go/pkg/core/state"
	"github.com/atipicial/atipicial-go/pkg/core/storage"
	"github.com/atipicial/atipicial-go/pkg/crypto/keys"
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/aef"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/atipicial/atipicial-go/pkg/vm/opcode"
	"github.com/stretchr/testify/require"
)

type atipicial struct{}

func (n atipicial) Initialize(context *interop.Context, hardfork *config.Hardfork, md *interop.HFSpecificContractMD) error {
	panic("TODO")
}
func (n atipicial) ActiveIn() *config.Hardfork { panic("TODO") }
func (n atipicial) InitializeCache(isHardforkEnabled interop.IsHardforkEnabled, blockHeight uint32, d *dao.Simple) error {
	panic("TODO")
}
func (n atipicial) Metadata() *interop.ContractMD                                { panic("TODO") }
func (n atipicial) OnPersist(context *interop.Context) error                     { panic("TODO") }
func (n atipicial) PostPersist(context *interop.Context) error                   { panic("TODO") }
func (n atipicial) GetCommitteeAddress(d *dao.Simple) util.Uint160               { panic("TODO") }
func (n atipicial) GetNextBlockValidatorsInternal(d *dao.Simple) keys.PublicKeys { panic("TODO") }
func (n atipicial) BalanceOf(d *dao.Simple, acc util.Uint160) (*big.Int, uint32) { panic("TODO") }
func (n atipicial) CalculateBonus(ic *interop.Context, acc util.Uint160, endHeight uint32) (*big.Int, error) {
	panic("TODO")
}
func (n atipicial) GetCommitteeMembers(d *dao.Simple) keys.PublicKeys        { panic("TODO") }
func (n atipicial) ComputeNextBlockValidators(d *dao.Simple) keys.PublicKeys { panic("TODO") }
func (n atipicial) GetCandidates(d *dao.Simple) ([]state.Validator, error)   { panic("TODO") }
func (n atipicial) CheckCommittee(ic *interop.Context) bool                  { panic("TODO") }
func (n atipicial) CheckAlmostFullCommittee(ic *interop.Context) bool        { panic("TODO") }
func (n atipicial) RevokeVotesDeferrable(ic *interop.Context, h util.Uint160, continuation func()) (bool, error) {
	return true, nil
}

func TestDeployGetUpdateContract(t *testing.T) {
	mgmt := NewManagement()
	p := NewPolicy()
	p.ATC = atipicial{}
	mgmt.Policy = p

	d := dao.NewSimple(storage.NewMemoryStore(), false)
	ic := &interop.Context{DAO: d}
	err := mgmt.Initialize(ic, nil, nil)
	require.NoError(t, err)
	require.NoError(t, mgmt.Policy.Initialize(&interop.Context{DAO: d}, nil, nil))
	script := []byte{byte(opcode.RET)}
	sender := util.Uint160{1, 2, 3}
	ne, err := aef.NewFile(script)
	require.NoError(t, err)
	manif := manifest.NewManifest("Test")
	manif.ABI.Methods = append(manif.ABI.Methods, manifest.Method{
		Name:       "dummy",
		ReturnType: smartcontract.VoidType,
		Parameters: []manifest.Parameter{},
	})

	h := state.CreateContractHash(sender, ne.Checksum, manif.Name)

	contract, err := mgmt.Deploy(ic, sender, ne, manif)
	require.NoError(t, err)
	require.Equal(t, int32(1), contract.ID)
	require.Equal(t, uint16(0), contract.UpdateCounter)
	require.Equal(t, h, contract.Hash)
	require.Equal(t, ne, &contract.AEF)
	require.Equal(t, *manif, contract.Manifest)

	// Double deploy.
	_, err = mgmt.Deploy(ic, sender, ne, manif)
	require.Error(t, err)

	// Different sender.
	sender2 := util.Uint160{3, 2, 1}
	contract2, err := mgmt.Deploy(ic, sender2, ne, manif)
	require.NoError(t, err)
	require.Equal(t, int32(2), contract2.ID)
	require.Equal(t, uint16(0), contract2.UpdateCounter)
	require.Equal(t, state.CreateContractHash(sender2, ne.Checksum, manif.Name), contract2.Hash)
	require.Equal(t, ne, &contract2.AEF)
	require.Equal(t, *manif, contract2.Manifest)

	refContract, err := GetContract(d, mgmt.ID, h)
	require.NoError(t, err)
	require.Equal(t, contract, refContract)

	refContract, err = GetContractByID(d, mgmt.ID, contract.ID)
	require.NoError(t, err)
	require.Equal(t, contract, refContract)

	upContract, err := mgmt.Update(ic, h, ne, manif)
	refContract.UpdateCounter++
	require.NoError(t, err)
	require.Equal(t, refContract, upContract)
}

func TestManagement_Initialize(t *testing.T) {
	t.Run("good", func(t *testing.T) {
		d := dao.NewSimple(storage.NewMemoryStore(), false)
		mgmt := NewManagement()
		require.NoError(t, mgmt.InitializeCache(func(hf *config.Hardfork, blockHeight uint32) bool { return false }, 0, d))
	})
	t.Run("invalid contract state", func(t *testing.T) {
		d := dao.NewSimple(storage.NewMemoryStore(), false)
		mgmt := NewManagement()
		d.PutStorageItem(mgmt.Metadata().ID, []byte{PrefixContract}, state.StorageItem{0xFF})
		require.Error(t, mgmt.InitializeCache(func(hf *config.Hardfork, blockHeight uint32) bool { return false }, 0, d))
	})
}

func TestManagement_GetAEP17Contracts(t *testing.T) {
	mgmt := NewManagement()
	mgmt.Policy = NewPolicy()
	d := dao.NewSimple(storage.NewMemoryStore(), false)
	err := mgmt.Initialize(&interop.Context{DAO: d}, nil, nil)
	require.NoError(t, err)
	require.NoError(t, mgmt.Policy.Initialize(&interop.Context{DAO: d}, nil, nil))
	err = mgmt.InitializeCache(func(hf *config.Hardfork, blockHeight uint32) bool { return false }, 0, d)
	require.NoError(t, err)

	require.Empty(t, mgmt.GetAEP17Contracts(d))
	private := d.GetPrivate()
	ic := &interop.Context{DAO: private}

	// Deploy AEP-17 contract
	script := []byte{byte(opcode.RET)}
	sender := util.Uint160{1, 2, 3}
	ne, err := aef.NewFile(script)
	require.NoError(t, err)
	manif := manifest.NewManifest("Test")
	manif.ABI.Methods = append(manif.ABI.Methods, manifest.Method{
		Name:       "dummy",
		ReturnType: smartcontract.VoidType,
		Parameters: []manifest.Parameter{},
	})
	manif.SupportedStandards = []string{manifest.AEP17StandardName}
	c1, err := mgmt.Deploy(ic, sender, ne, manif)
	require.NoError(t, err)

	// c1 contract hash should be returned, as private DAO already contains changed cache.
	require.Equal(t, []util.Uint160{c1.Hash}, mgmt.GetAEP17Contracts(private))

	// Lower DAO still shouldn't contain c1, as no Persist was called.
	require.Empty(t, mgmt.GetAEP17Contracts(d))

	// Call Persist, check c1 contract hash is returned
	_, err = private.Persist()
	require.NoError(t, err)
	require.Equal(t, []util.Uint160{c1.Hash}, mgmt.GetAEP17Contracts(d))

	// Update contract
	private = d.GetPrivate()
	manif.ABI.Methods = append(manif.ABI.Methods, manifest.Method{
		Name:       "dummy2",
		ReturnType: smartcontract.VoidType,
		Parameters: []manifest.Parameter{},
	})
	c1Updated, err := mgmt.Update(&interop.Context{DAO: private}, c1.Hash, ne, manif)
	require.NoError(t, err)
	require.Equal(t, c1.Hash, c1Updated.Hash)

	// No changes expected in lower store.
	require.Equal(t, []util.Uint160{c1.Hash}, mgmt.GetAEP17Contracts(d))
	c1Lower, err := GetContract(d, mgmt.ID, c1.Hash)
	require.NoError(t, err)
	require.Equal(t, 1, len(c1Lower.Manifest.ABI.Methods))
	require.Equal(t, []util.Uint160{c1Updated.Hash}, mgmt.GetAEP17Contracts(private))
	c1Upper, err := GetContract(private, mgmt.ID, c1Updated.Hash)
	require.NoError(t, err)
	require.Equal(t, 2, len(c1Upper.Manifest.ABI.Methods))

	// Call Persist, check c1Updated state is returned from lower.
	_, err = private.Persist()
	require.NoError(t, err)
	require.Equal(t, []util.Uint160{c1.Hash}, mgmt.GetAEP17Contracts(d))
	c1Lower, err = GetContract(d, mgmt.ID, c1.Hash)
	require.NoError(t, err)
	require.Equal(t, 2, len(c1Lower.Manifest.ABI.Methods))
}
