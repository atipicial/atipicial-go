package aep_test

import (
	"io"
	"math/big"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/atipicial/atipicial-go/internal/testcli"
	"github.com/atipicial/atipicial-go/pkg/core/native/nativehashes"
	"github.com/atipicial/atipicial-go/pkg/core/native/nativenames"
	"github.com/atipicial/atipicial-go/pkg/encoding/address"
	"github.com/atipicial/atipicial-go/pkg/encoding/fixedn"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/atipicial/atipicial-go/pkg/wallet"
	"github.com/stretchr/testify/require"
)

func TestAEP17Balance(t *testing.T) {
	e := testcli.NewExecutor(t, true)

	args := []string{
		"atipicial-go", "wallet", "aep17", "multitransfer", "--force",
		"--rpc-endpoint", "http://" + e.RPC.Addresses()[0],
		"--wallet", testcli.ValidatorWallet,
		"--from", testcli.ValidatorAddr,
		"GAS:" + testcli.TestWalletMultiAccount1 + ":1",
		"ATC:" + testcli.TestWalletMultiAccount1 + ":10",
		"GAS:" + testcli.TestWalletMultiAccount3 + ":3",
	}
	e.In.WriteString("one\r")
	e.Run(t, args...)
	e.CheckTxPersisted(t)

	var checkAcc1ATC = func(t *testing.T, e *testcli.Executor, line string) {
		if line == "" {
			line = e.GetNextLine(t)
		}
		balance, index := e.Chain.GetGoverningTokenBalance(testcli.TestWalletMultiAccount1Hash)
		e.CheckLine(t, line, "^\\s*ATC:\\s+AtipicialCoin \\("+e.Chain.GoverningTokenHash().StringLE()+"\\)")
		e.CheckNextLine(t, "^\\s*Amount\\s*:\\s*"+balance.String()+"$")
		e.CheckNextLine(t, "^\\s*Updated\\s*:\\s*"+strconv.FormatUint(uint64(index), 10))
	}
	var checkAcc1GAS = func(t *testing.T, e *testcli.Executor, line string) {
		if line == "" {
			line = e.GetNextLine(t)
		}
		e.CheckLine(t, line, "^\\s*GAS:\\s+AtipicialDollar \\("+e.Chain.UtilityTokenHash().StringLE()+"\\)")
		balance := e.Chain.GetUtilityTokenBalance(testcli.TestWalletMultiAccount1Hash, util.Uint160{})
		e.CheckNextLine(t, "^\\s*Amount\\s*:\\s*"+fixedn.Fixed8(balance.Int64()).String()+"$")
		e.CheckNextLine(t, "^\\s*Updated:")
	}
	var checkAcc1Assets = func(t *testing.T, e *testcli.Executor) {
		e.CheckNextLine(t, "^Account "+testcli.TestWalletMultiAccount1)
		// The order of assets is undefined.
		for range 2 {
			line := e.GetNextLine(t)
			if strings.Contains(line, "GAS") {
				checkAcc1GAS(t, e, line)
			} else {
				checkAcc1ATC(t, e, line)
			}
		}
	}

	var (
		cmdbase      = []string{"atipicial-go", "wallet", "aep17", "balance", "--rpc-endpoint", "http://" + e.RPC.Addresses()[0]}
		addrparams   = []string{"--address", testcli.TestWalletMultiAccount1}
		walletparams = []string{"--wallet", testcli.TestWalletMultiPath}
	)
	t.Run("Bad wallet", func(t *testing.T) {
		e.RunWithError(t, append(cmdbase, "--wallet", "/dev/null")...)
	})
	t.Run("empty wallet", func(t *testing.T) {
		tmpDir := t.TempDir()
		walletPath := filepath.Join(tmpDir, "emptywallet.json")
		require.NoError(t, os.WriteFile(walletPath, []byte("{}"), 0o644))
		e.RunWithError(t, append(cmdbase, "--wallet", walletPath)...)
	})
	t.Run("no wallet or address", func(t *testing.T) {
		e.RunWithError(t, cmdbase...)
	})
	for name, params := range map[string][]string{
		"address only":        addrparams,
		"address with wallet": slices.Concat(walletparams, addrparams),
	} {
		var cmd = append(cmdbase, params...)
		t.Run(name, func(t *testing.T) {
			t.Run("all tokens", func(t *testing.T) {
				e.Run(t, cmd...)
				checkAcc1Assets(t, e)
				e.CheckEOF(t)
			})
			t.Run("excessive parameters", func(t *testing.T) {
				e.RunWithError(t, append(cmd, "--token", "ATC", "gas")...)
			})
		})
		t.Run("ATC", func(t *testing.T) {
			checkResult := func(t *testing.T) {
				e.CheckNextLine(t, "^\\s*Account\\s+"+testcli.TestWalletMultiAccount1)
				checkAcc1ATC(t, e, "")
				e.CheckEOF(t)
			}
			t.Run("Alias", func(t *testing.T) {
				e.Run(t, append(cmd, "--token", "ATC")...)
				checkResult(t)
			})
			t.Run("Hash", func(t *testing.T) {
				e.Run(t, append(cmd, "--token", e.Chain.GoverningTokenHash().StringLE())...)
				checkResult(t)
			})
		})
		t.Run("GAS", func(t *testing.T) {
			e.Run(t, append(cmd, "--token", "GAS")...)
			e.CheckNextLine(t, "^\\s*Account\\s+"+testcli.TestWalletMultiAccount1)
			checkAcc1GAS(t, e, "")
		})
		t.Run("Bad token", func(t *testing.T) {
			e.Run(t, append(cmd, "--token", "kek")...)
			e.CheckNextLine(t, "^\\s*Account\\s+"+testcli.TestWalletMultiAccount1)
			e.CheckNextLine(t, `^\s*Can't find data for "kek" token\s*`)
			e.CheckEOF(t)
		})
	}
	t.Run("inexistent wallet account", func(t *testing.T) {
		var cmd = append(cmdbase, walletparams...)
		e.RunWithError(t, append(cmd, "--address", "AaZzAtipicialTestAddressPlaceholder0000000001")...)
	})
	t.Run("zero balance of known token", func(t *testing.T) {
		e.Run(t, append(cmdbase, []string{"--token", "ATC", "--address", testcli.TestWalletMultiAccount2}...)...)
		e.CheckNextLine(t, "^Account "+testcli.TestWalletMultiAccount2)
		e.CheckNextLine(t, "^\\s*ATC:\\s+AtipicialCoin \\("+e.Chain.GoverningTokenHash().StringLE()+"\\)")
		e.CheckNextLine(t, "^\\s*Amount\\s*:\\s*"+fixedn.Fixed8(0).String()+"$")
		e.CheckNextLine(t, "^\\s*Updated:")
		e.CheckEOF(t)
	})
	t.Run("all accounts", func(t *testing.T) {
		e.Run(t, append(cmdbase, walletparams...)...)

		checkAcc1Assets(t, e)
		e.CheckNextLine(t, "^\\s*$")

		e.CheckNextLine(t, "^Account "+testcli.TestWalletMultiAccount2)
		e.CheckNextLine(t, "^\\s*$")

		e.CheckNextLine(t, "^Account "+testcli.TestWalletMultiAccount3)
		e.CheckNextLine(t, "^\\s*GAS:\\s+AtipicialDollar \\("+e.Chain.UtilityTokenHash().StringLE()+"\\)")
		balance := e.Chain.GetUtilityTokenBalance(testcli.TestWalletMultiAccount3Hash, util.Uint160{})
		e.CheckNextLine(t, "^\\s*Amount\\s*:\\s*"+fixedn.Fixed8(balance.Int64()).String()+"$")
		e.CheckNextLine(t, "^\\s*Updated:")
		e.CheckEOF(t)
	})
}

func TestAEP17Transfer(t *testing.T) {
	w, err := wallet.NewWalletFromFile("../testdata/testwallet.json")
	require.NoError(t, err)

	e := testcli.NewExecutor(t, true)
	args := []string{
		"atipicial-go", "wallet", "aep17", "transfer",
		"--rpc-endpoint", "http://" + e.RPC.Addresses()[0],
		"--wallet", testcli.ValidatorWallet,
		"--to", w.Accounts[0].Address,
		"--token", "ATC",
		"--amount", "1",
		"--from", testcli.ValidatorAddr,
	}

	t.Run("missing receiver", func(t *testing.T) {
		as := slices.Concat(args[:8], args[10:])
		e.In.WriteString("one\r")
		e.RunWithErrorCheck(t, `Required flag "to" not set`, as...)
		e.In.Reset()
	})

	t.Run("InvalidPassword", func(t *testing.T) {
		e.In.WriteString("onetwothree\r")
		e.RunWithError(t, args...)
		e.In.Reset()
	})

	t.Run("no confirmation", func(t *testing.T) {
		e.In.WriteString("one\r")
		e.RunWithError(t, args...)
		e.In.Reset()
	})
	t.Run("cancel after prompt", func(t *testing.T) {
		e.In.WriteString("one\r")
		e.RunWithError(t, args...)
		e.In.Reset()
	})

	e.In.WriteString("one\r")
	e.In.WriteString("Y\r")
	e.Run(t, args...)
	e.CheckNextLine(t, `^Network fee:\s*(\d|\.)+`)
	e.CheckNextLine(t, `^System fee:\s*(\d|\.)+`)
	e.CheckNextLine(t, `^Total fee:\s*(\d|\.)+`)
	e.CheckTxPersisted(t)

	sh := w.Accounts[0].ScriptHash()
	b, _ := e.Chain.GetGoverningTokenBalance(sh)
	require.Equal(t, big.NewInt(1), b)

	t.Run("with force", func(t *testing.T) {
		e.In.WriteString("one\r")
		e.Run(t, append(args, "--force")...)
		e.CheckTxPersisted(t)

		b, _ := e.Chain.GetGoverningTokenBalance(sh)
		require.Equal(t, big.NewInt(2), b)
	})

	hVerify := deployVerifyContract(t, e)
	const validatorDefault = "Nhfg3TbpwogLvDGVvAvqyThbsHgoSUKwtn"

	t.Run("default address", func(t *testing.T) {
		e.In.WriteString("one\r")
		e.Run(t, "atipicial-go", "wallet", "aep17", "multitransfer",
			"--rpc-endpoint", "http://"+e.RPC.Addresses()[0],
			"--wallet", testcli.ValidatorWallet,
			"--from", testcli.ValidatorAddr,
			"--force",
			"ATC:"+validatorDefault+":42",
			"GAS:"+validatorDefault+":7")
		e.CheckTxPersisted(t)

		args := args[:len(args)-2] // cut '--from' argument
		args = append(args, "--force")
		e.In.WriteString("one\r")
		e.Run(t, args...)
		e.CheckTxPersisted(t)

		b, _ := e.Chain.GetGoverningTokenBalance(sh)
		require.Equal(t, big.NewInt(3), b)

		sh, err = address.StringToUint160(validatorDefault)
		require.NoError(t, err)
		b, _ = e.Chain.GetGoverningTokenBalance(sh)
		require.Equal(t, big.NewInt(41), b)
	})

	t.Run("with signers", func(t *testing.T) {
		e.In.WriteString("one\r")
		e.Run(t, "atipicial-go", "wallet", "aep17", "multitransfer",
			"--rpc-endpoint", "http://"+e.RPC.Addresses()[0],
			"--wallet", testcli.ValidatorWallet,
			"--from", testcli.ValidatorAddr,
			"--force",
			"ATC:"+validatorDefault+":42",
			"GAS:"+validatorDefault+":7",
			"--", testcli.ValidatorAddr+":Global")
		e.CheckTxPersisted(t)
	})

	validTil := e.Chain.BlockHeight() + 100
	cmd := []string{
		"atipicial-go", "wallet", "aep17", "transfer",
		"--rpc-endpoint", "http://" + e.RPC.Addresses()[0],
		"--wallet", testcli.ValidatorWallet,
		"--token", "GAS",
		"--amount", "1",
		"--force",
		"--from", testcli.ValidatorAddr}

	t.Run("with await", func(t *testing.T) {
		e.In.WriteString("one\r")
		e.Run(t, append(cmd, "--to", nftOwnerAddr, "--await")...)
		e.CheckAwaitableTxPersisted(t)
	})

	cmd = append(cmd, "--to", address.Uint160ToString(nativehashes.Notary),
		"[", testcli.ValidatorAddr, strconv.Itoa(int(validTil)), "]")

	t.Run("with data", func(t *testing.T) {
		e.In.WriteString("one\r")
		e.Run(t, cmd...)
		e.CheckTxPersisted(t)
	})

	t.Run("with data and signers", func(t *testing.T) {
		t.Run("invalid sender's scope", func(t *testing.T) {
			e.In.WriteString("one\r")
			e.RunWithError(t, append(cmd, "--", testcli.ValidatorAddr+":None")...)
		})
		t.Run("good", func(t *testing.T) {
			e.In.WriteString("one\r")
			e.Run(t, append(cmd, "--", testcli.ValidatorAddr+":Global")...) // CalledByEntry is enough, but it's the default value, so check something else
			e.CheckTxPersisted(t)
		})
		t.Run("several signers", func(t *testing.T) {
			e.In.WriteString("one\r")
			e.Run(t, append(cmd, "--", testcli.ValidatorAddr, hVerify.StringLE())...)
			e.CheckTxPersisted(t)
		})
	})
}

func TestAEP17MultiTransfer(t *testing.T) {
	privs, _ := testcli.GenerateKeys(t, 3)

	e := testcli.NewExecutor(t, true)
	atipicialContractHash, err := e.Chain.GetNativeContractScriptHash(nativenames.Atipicial)
	require.NoError(t, err)
	args := []string{
		"atipicial-go", "wallet", "aep17", "multitransfer",
		"--rpc-endpoint", "http://" + e.RPC.Addresses()[0],
		"--wallet", testcli.ValidatorWallet,
		"--from", testcli.ValidatorAddr,
		"--force",
		"ATC:" + privs[0].Address() + ":42",
		"GAS:" + privs[1].Address() + ":7",
		atipicialContractHash.StringLE() + ":" + privs[2].Address() + ":13",
	}
	hVerify := deployVerifyContract(t, e)

	t.Run("no cosigners", func(t *testing.T) {
		e.In.WriteString("one\r")
		e.Run(t, args...)
		e.CheckTxPersisted(t)

		b, _ := e.Chain.GetGoverningTokenBalance(privs[0].GetScriptHash())
		require.Equal(t, big.NewInt(42), b)
		b = e.Chain.GetUtilityTokenBalance(privs[1].GetScriptHash(), util.Uint160{})
		require.Equal(t, big.NewInt(int64(fixedn.Fixed8FromInt64(7))), b)
		b, _ = e.Chain.GetGoverningTokenBalance(privs[2].GetScriptHash())
		require.Equal(t, big.NewInt(13), b)
	})

	t.Run("invalid sender scope", func(t *testing.T) {
		e.In.WriteString("one\r")
		e.RunWithError(t, append(args,
			"--", testcli.ValidatorAddr+":None")...) // invalid sender scope
	})
	t.Run("Global sender scope", func(t *testing.T) {
		e.In.WriteString("one\r")
		e.Run(t, append(args,
			"--", testcli.ValidatorAddr+":Global")...)
		e.CheckTxPersisted(t)
	})
	t.Run("Several cosigners", func(t *testing.T) {
		e.In.WriteString("one\r")
		e.Run(t, append(args,
			"--", testcli.ValidatorAddr, hVerify.StringLE())...)
		e.CheckTxPersisted(t)
	})
}

func TestAEP17ImportToken(t *testing.T) {
	e := testcli.NewExecutor(t, true)
	tmpDir := t.TempDir()
	walletPath := filepath.Join(tmpDir, "walletForImport.json")

	atipicialContractHash, err := e.Chain.GetNativeContractScriptHash(nativenames.Atipicial)
	require.NoError(t, err)
	gasContractHash, err := e.Chain.GetNativeContractScriptHash(nativenames.Gas)
	require.NoError(t, err)
	nnsContractHash := deployNNSContract(t, e)
	e.Run(t, "atipicial-go", "wallet", "init", "--wallet", walletPath)

	// missing token hash
	e.RunWithErrorCheck(t, `Required flag "token" not set`, "atipicial-go", "wallet", "aep17", "import",
		"--rpc-endpoint", "http://"+e.RPC.Addresses()[0],
		"--wallet", walletPath)

	// additional parameter
	e.RunWithError(t, "atipicial-go", "wallet", "aep17", "import",
		"--rpc-endpoint", "http://"+e.RPC.Addresses()[0],
		"--wallet", walletPath,
		"--token", gasContractHash.StringLE(), "useless")
	e.Run(t, "atipicial-go", "wallet", "aep17", "import",
		"--rpc-endpoint", "http://"+e.RPC.Addresses()[0],
		"--wallet", walletPath,
		"--token", gasContractHash.StringLE())
	e.Run(t, "atipicial-go", "wallet", "aep17", "import",
		"--rpc-endpoint", "http://"+e.RPC.Addresses()[0],
		"--wallet", walletPath,
		"--token", address.Uint160ToString(atipicialContractHash)) // try address instead of sh

	// not a AEP-17 token
	e.RunWithError(t, "atipicial-go", "wallet", "aep17", "import",
		"--rpc-endpoint", "http://"+e.RPC.Addresses()[0],
		"--wallet", walletPath,
		"--token", nnsContractHash.StringLE())

	t.Run("Info", func(t *testing.T) {
		checkGASInfo := func(t *testing.T) {
			e.CheckNextLine(t, "^Name:\\s*AtipicialDollar")
			e.CheckNextLine(t, "^Symbol:\\s*GAS")
			e.CheckNextLine(t, "^Hash:\\s*"+gasContractHash.StringLE())
			e.CheckNextLine(t, "^Decimals:\\s*8")
			e.CheckNextLine(t, "^Address:\\s*"+address.Uint160ToString(gasContractHash))
			e.CheckNextLine(t, "^Standard:\\s*"+string(manifest.AEP17StandardName))
		}
		t.Run("excessive parameters", func(t *testing.T) {
			e.RunWithError(t, "atipicial-go", "wallet", "aep17", "info",
				"--wallet", walletPath, "--token", gasContractHash.StringLE(), "parameter")
		})
		t.Run("WithToken", func(t *testing.T) {
			e.Run(t, "atipicial-go", "wallet", "aep17", "info",
				"--wallet", walletPath, "--token", gasContractHash.StringLE())
			checkGASInfo(t)
		})
		t.Run("NoToken", func(t *testing.T) {
			e.Run(t, "atipicial-go", "wallet", "aep17", "info",
				"--wallet", walletPath)
			checkGASInfo(t)
			_, err := e.Out.ReadString('\n')
			require.NoError(t, err)
			e.CheckNextLine(t, "^Name:\\s*AtipicialCoin")
			e.CheckNextLine(t, "^Symbol:\\s*ATC")
			e.CheckNextLine(t, "^Hash:\\s*"+atipicialContractHash.StringLE())
			e.CheckNextLine(t, "^Decimals:\\s*0")
			e.CheckNextLine(t, "^Address:\\s*"+address.Uint160ToString(atipicialContractHash))
			e.CheckNextLine(t, "^Standard:\\s*"+string(manifest.AEP17StandardName))
		})
		t.Run("Remove", func(t *testing.T) {
			e.RunWithError(t, "atipicial-go", "wallet", "aep17", "remove",
				"--wallet", walletPath, "--token", atipicialContractHash.StringLE(), "add")
			e.In.WriteString("y\r")
			e.Run(t, "atipicial-go", "wallet", "aep17", "remove",
				"--wallet", walletPath, "--token", atipicialContractHash.StringLE())
			e.Run(t, "atipicial-go", "wallet", "aep17", "info",
				"--wallet", walletPath)
			checkGASInfo(t)
			_, err := e.Out.ReadString('\n')
			require.Equal(t, err, io.EOF)
		})
	})
}
