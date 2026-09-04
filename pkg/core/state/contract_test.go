package state

import (
	"math"
	"math/big"
	"testing"

	"github.com/atipicial/atipicial-go/internal/testserdes"
	"github.com/atipicial/atipicial-go/pkg/crypto/hash"
	"github.com/atipicial/atipicial-go/pkg/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/aef"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/atipicial/atipicial-go/pkg/vm/stackitem"
	"github.com/stretchr/testify/require"
)

func TestContractStateToFromSI(t *testing.T) {
	script := []byte("testscript")

	h := hash.Hash160(script)
	m := manifest.NewManifest("Test")
	m.ABI.Methods = []manifest.Method{{
		Name: "main",
		Parameters: []manifest.Parameter{
			{
				Name: "amount",
				Type: smartcontract.IntegerType,
			},
			{
				Name: "hash",
				Type: smartcontract.Hash160Type,
			},
		},
		ReturnType: smartcontract.BoolType,
	}}
	contract := &Contract{
		UpdateCounter: 42,
		ContractBase: ContractBase{
			ID:   123,
			Hash: h,
			AEF: aef.File{
				Header: aef.Header{
					Magic:    aef.Magic,
					Compiler: "atipicial-go.test-test",
				},
				Tokens:   []aef.MethodToken{},
				Script:   script,
				Checksum: 0,
			},
			Manifest: *m,
		},
	}
	contract.AEF.Checksum = contract.AEF.CalculateChecksum()

	t.Run("Convertible", func(t *testing.T) {
		contractDecoded := new(Contract)
		testserdes.ToFromStackItem(t, contract, contractDecoded)

		t.Run("preserve wildcard trusts", func(t *testing.T) {
			contract.Manifest.Trusts.Value = nil
			contract.Manifest.Trusts.Wildcard = true
			require.True(t, contract.Manifest.Trusts.IsWildcard())
			actual := new(Contract)
			item, err := contract.ToStackItem()
			require.NoError(t, err)
			require.NoError(t, actual.FromStackItem(item))
			require.True(t, actual.Manifest.Trusts.IsWildcard())
		})
	})
	t.Run("JSON", func(t *testing.T) {
		contractDecoded := new(Contract)
		testserdes.MarshalUnmarshalJSON(t, contract, contractDecoded)
	})
}

func TestCreateContractHash(t *testing.T) {
	var aeff = aef.File{
		Header: aef.Header{
			Compiler: "test",
			Magic:    aef.Magic,
		},
		Tokens: []aef.MethodToken{},
		Script: []byte{1, 2, 3},
	}
	var sender util.Uint160
	var err error

	aeff.Checksum = aeff.CalculateChecksum()
	require.Equal(t, "9b9628e4f1611af90e761eea8cc21372380c74b6", CreateContractHash(sender, aeff.Checksum, "").StringLE())
	sender, err = util.Uint160DecodeStringLE("a400ff00ff00ff00ff00ff00ff00ff00ff00ff01")
	require.NoError(t, err)
	require.Equal(t, "66eec404d86b918d084e62a29ac9990e3b6f4286", CreateContractHash(sender, aeff.Checksum, "").StringLE())
}

func TestContractFromStackItem(t *testing.T) {
	var (
		id           = stackitem.Make(42)
		counter      = stackitem.Make(11)
		chash        = stackitem.Make(util.Uint160{1, 2, 3}.BytesBE())
		script       = []byte{0, 9, 8}
		aefFile, _   = aef.NewFile(script)
		rawAef, _    = aefFile.Bytes()
		aefItem      = stackitem.NewByteArray(rawAef)
		manifest     = manifest.DefaultManifest("stack item")
		manifItem, _ = manifest.ToStackItem()

		badCases = []struct {
			name string
			item stackitem.Item
		}{
			{"not an array", stackitem.Make(1)},
			{"wrong array", stackitem.Make([]stackitem.Item{})},
			{"id is not a number", stackitem.Make([]stackitem.Item{manifItem, counter, chash, aefItem, manifItem})},
			{"id is out of range", stackitem.Make([]stackitem.Item{stackitem.Make(math.MaxUint32), counter, chash, aefItem, manifItem})},
			{"counter is not a number", stackitem.Make([]stackitem.Item{id, manifItem, chash, aefItem, manifItem})},
			{"counter is out of range", stackitem.Make([]stackitem.Item{id, stackitem.Make(100500), chash, aefItem, manifItem})},
			{"hash is not a byte string", stackitem.Make([]stackitem.Item{id, counter, stackitem.NewArray(nil), aefItem, manifItem})},
			{"hash is not a hash", stackitem.Make([]stackitem.Item{id, counter, stackitem.Make([]byte{1, 2, 3}), aefItem, manifItem})},
			{"aef is not a byte string", stackitem.Make([]stackitem.Item{id, counter, chash, stackitem.NewArray(nil), manifItem})},
			{"manifest is not an array", stackitem.Make([]stackitem.Item{id, counter, chash, aefItem, stackitem.NewByteArray(nil)})},
			{"manifest is not correct", stackitem.Make([]stackitem.Item{id, counter, chash, aefItem, stackitem.NewArray([]stackitem.Item{stackitem.Make(100500)})})},
		}
	)
	for _, cs := range badCases {
		t.Run(cs.name, func(t *testing.T) {
			var c = new(Contract)
			err := c.FromStackItem(cs.item)
			require.Error(t, err)
		})
	}
	var c = new(Contract)
	err := c.FromStackItem(stackitem.Make([]stackitem.Item{id, counter, chash, aefItem, manifItem}))
	require.NoError(t, err)
}

func TestContract_ToSCParameter(t *testing.T) {
	script := []byte("testscript")
	h := hash.Hash160(script)
	m := manifest.NewManifest("Test")
	m.ABI.Methods = []manifest.Method{{
		Name: "main",
		Parameters: []manifest.Parameter{
			{Name: "amount", Type: smartcontract.IntegerType},
			{Name: "hash", Type: smartcontract.Hash160Type},
		},
		ReturnType: smartcontract.BoolType,
	}}
	contract := &Contract{
		UpdateCounter: 42,
		ContractBase: ContractBase{
			ID:   123,
			Hash: h,
			AEF: aef.File{
				Header: aef.Header{
					Magic:    aef.Magic,
					Compiler: "atipicial-go.test-test",
				},
				Tokens:   []aef.MethodToken{},
				Script:   script,
				Checksum: 0,
			},
			Manifest: *m,
		},
	}
	contract.AEF.Checksum = contract.AEF.CalculateChecksum()

	prm, err := contract.ToSCParameter()
	require.NoError(t, err)
	require.Equal(t, smartcontract.ArrayType, prm.Type)
	arr, ok := prm.Value.([]smartcontract.Parameter)
	require.True(t, ok)
	require.Len(t, arr, 5)

	require.Equal(t, smartcontract.Parameter{Type: smartcontract.IntegerType, Value: big.NewInt(int64(contract.ID))}, arr[0])
	require.Equal(t, smartcontract.Parameter{Type: smartcontract.IntegerType, Value: big.NewInt(int64(contract.UpdateCounter))}, arr[1])
	require.Equal(t, smartcontract.Parameter{Type: smartcontract.Hash160Type, Value: contract.Hash}, arr[2])
	rawAef, err := contract.AEF.Bytes()
	require.NoError(t, err)
	require.Equal(t, smartcontract.Parameter{Type: smartcontract.ByteArrayType, Value: rawAef}, arr[3])
	mPrm, err := contract.Manifest.ToSCParameter()
	require.NoError(t, err)
	require.Equal(t, mPrm, arr[4])

	t.Run("nil", func(t *testing.T) {
		var nilContract *Contract
		prm, err := nilContract.ToSCParameter()
		require.NoError(t, err)
		require.Equal(t, smartcontract.Parameter{Type: smartcontract.AnyType}, prm)
	})

	// Unlike Block/Transaction/ATCBalance/WhitelistFeeContract, a round trip
	// via Parameter.ToStackItem() isn't possible here: Manifest.ToSCParameter
	// always includes a MapType parameter for the (always empty) Features
	// field, and smartcontract.ExpandParameterToEmitable doesn't support
	// MapType (it's not used by any production code path, only by tests).
	// The FromStackItem/ToStackItem round trip is already covered by
	// TestContractStateToFromSI.
}
