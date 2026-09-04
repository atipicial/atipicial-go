package interop_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/atipicial/atipicial-go/pkg/compiler"
	"github.com/atipicial/atipicial-go/pkg/core/native/nativehashes"
	"github.com/atipicial/atipicial-go/pkg/atipicialtest"
	"github.com/atipicial/atipicial-go/pkg/atipicialtest/chain"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
)

func TestContext_LoadToken(t *testing.T) {
	bc, acc := chain.NewSingle(t)
	e := atipicialtest.NewExecutor(t, bc, acc, acc)
	m := e.CommitteeInvoker(nativehashes.ContractManagement)

	check := func(t *testing.T, callStr string, errText string) {
		src := `package contract
		import (
			"github.com/atipicial/atipicial-go/pkg/interop/contract"
			"github.com/atipicial/atipicial-go/pkg/interop/atipicialgointernal"
		)
		const hash = "%s" 

		func CallMyMethod() {
			` + callStr + `
		}`
		var hStr strings.Builder
		for _, b := range nativehashes.StdLib.BytesBE() {
			fmt.Fprintf(&hStr, "\\x%02x", b)
		}
		ctr := atipicialtest.CompileSource(t, e.Validator.ScriptHash(), strings.NewReader(fmt.Sprintf(src, hStr.String())), &compiler.Options{
			Name: "CALLT contract",
			Permissions: []manifest.Permission{
				{
					Contract: manifest.PermissionDesc{
						Value: nativehashes.StdLib,
					},
				},
			},
		})
		m.DeployContract(t, ctr, nil)

		ctr1Invoker := e.NewInvoker(ctr.Hash, e.Committee)
		ctr1Invoker.InvokeFail(t, errText, "callMyMethod")
	}

	// CALLT contains some unknown method of another contract.
	t.Run("missing method", func(t *testing.T) {
		check(t, `atipicialgointernal.CallWithTokenNoRet(hash, "myMethod", int(contract.All), "arg1", "arg2")`,
			"at instruction 13 (CALLT): token method not found: myMethod/2")
	})

	// CALLT return value doesn't match the actual method's return value
	t.Run("invalid return value", func(t *testing.T) {
		check(t, `atipicialgointernal.CallWithTokenNoRet(hash, "hexDecode", int(contract.All), "0x01")`,
			"at instruction 6 (CALLT): token hexDecode/1 return value (false) doesn't match the actual method return value")
	})
}
