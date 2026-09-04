package native_test

import (
	"encoding/json"
	"testing"

	"github.com/atipicial/atipicial-go/internal/basicchain"
	"github.com/atipicial/atipicial-go/pkg/config"
	"github.com/atipicial/atipicial-go/pkg/core/native/nativenames"
	"github.com/atipicial/atipicial-go/pkg/core/state"
	"github.com/atipicial/atipicial-go/pkg/atipicialtest"
	"github.com/atipicial/atipicial-go/pkg/atipicialtest/chain"
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/aef"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/atipicial/atipicial-go/pkg/vm/opcode"
	"github.com/atipicial/atipicial-go/pkg/vm/stackitem"
	"github.com/stretchr/testify/require"
)

func TestManagement_GetAEP17Contracts(t *testing.T) {
	t.Run("empty chain", func(t *testing.T) {
		bc, validators, committee := chain.NewMulti(t)
		e := atipicialtest.NewExecutor(t, bc, validators, committee)

		require.ElementsMatch(t, []util.Uint160{e.NativeHash(t, nativenames.Atipicial),
			e.NativeHash(t, nativenames.Gas)}, bc.GetAEP17Contracts())
	})

	t.Run("basic chain", func(t *testing.T) {
		bc, validators, committee := chain.NewMulti(t)
		e := atipicialtest.NewExecutor(t, bc, validators, committee)
		basicchain.Init(t, "../../../", e)

		require.ElementsMatch(t, []util.Uint160{e.NativeHash(t, nativenames.Atipicial),
			e.NativeHash(t, nativenames.Gas), e.ContractHash(t, 1)}, bc.GetAEP17Contracts())
	})
}

func TestManagement_DeployUpdate_HFBasilisk(t *testing.T) {
	bc, acc := chain.NewSingleWithCustomConfig(t, func(c *config.Blockchain) {
		c.Hardforks = map[string]uint32{
			config.HFBasilisk.String(): 2,
		}
	})
	e := atipicialtest.NewExecutor(t, bc, acc, acc)

	ne, err := aef.NewFile([]byte{byte(opcode.JMP), 0x05})
	require.NoError(t, err)

	m := &manifest.Manifest{
		Name:     "ctr",
		Features: json.RawMessage("{}"),
		Groups:   []manifest.Group{},
		Trusts:   manifest.WildPermissionDescs{Wildcard: true},
		ABI: manifest.ABI{
			Methods: []manifest.Method{
				{
					Name:   "main",
					Offset: 0,
				},
			},
		},
	}
	ctr := &atipicialtest.Contract{

		Hash:     state.CreateContractHash(e.Validator.ScriptHash(), ne.Checksum, m.Name),
		AEF:      ne,
		Manifest: m,
	}

	// Block 1: no script check on deploy.
	e.DeployContract(t, ctr, nil)
	e.AddNewBlock(t)

	// Block 3: script check on deploy.
	ctr.Manifest.Name = "other name"
	e.DeployContractCheckFAULT(t, ctr, nil, "invalid contract script: invalid offset 5 ip at 0")
}

func TestManagement_CallInTheSameBlock(t *testing.T) {
	bc, acc := chain.NewSingle(t)

	e := atipicialtest.NewExecutor(t, bc, acc, acc)

	ne, err := aef.NewFile([]byte{byte(opcode.PUSH1)})
	require.NoError(t, err)

	m := &manifest.Manifest{
		Name:     "ctr",
		Features: json.RawMessage("{}"),
		Groups:   []manifest.Group{},
		Trusts:   manifest.WildPermissionDescs{Wildcard: true},
		ABI: manifest.ABI{
			Methods: []manifest.Method{
				{
					Name:   "main",
					Offset: 0,
				},
			},
		},
	}

	t.Run("same tx", func(t *testing.T) {
		h := state.CreateContractHash(e.Validator.ScriptHash(), ne.Checksum, m.Name)

		neb, err := ne.Bytes()
		require.NoError(t, err)

		rawManifest, err := json.Marshal(m)
		require.NoError(t, err)

		b := smartcontract.NewBuilder()
		b.InvokeWithAssert(bc.ManagementContractHash(), "deploy", neb, rawManifest) // We need to drop the resulting contract and ASSERT does that.
		b.InvokeWithAssert(bc.ManagementContractHash(), "hasMethod", h, "main", 0)
		b.InvokeMethod(h, "main")

		script, err := b.Script()
		require.NoError(t, err)
		txHash := e.InvokeScript(t, script, []atipicialtest.Signer{e.Validator})
		e.CheckHalt(t, txHash, stackitem.Make(1))
	})
	t.Run("next tx", func(t *testing.T) {
		m.Name = "another contract"
		h := state.CreateContractHash(e.Validator.ScriptHash(), ne.Checksum, m.Name)

		txDeploy := e.NewDeployTx(t, &atipicialtest.Contract{Hash: h, AEF: ne, Manifest: m}, nil)
		txHasMethod := e.NewTx(t, []atipicialtest.Signer{e.Validator}, bc.ManagementContractHash(), "hasMethod", h, "main", 0)

		txCall := e.NewUnsignedTx(t, h, "main") // Test invocation doesn't give true GAS cost before deployment.
		txCall = e.SignTx(t, txCall, 1_0000_0000, e.Validator)

		e.AddNewBlock(t, txDeploy, txHasMethod, txCall)
		e.CheckHalt(t, txHasMethod.Hash(), stackitem.Make(true))
		e.CheckHalt(t, txCall.Hash(), stackitem.Make(1))
	})
}
