package atipicialtest

import (
	"encoding/json"
	"io"
	"os"
	"testing"

	"github.com/atipicial/atipicial-go/cli/smartcontract"
	"github.com/atipicial/atipicial-go/pkg/compiler"
	"github.com/atipicial/atipicial-go/pkg/config"
	"github.com/atipicial/atipicial-go/pkg/core/state"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/aef"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/stretchr/testify/require"
)

// Contract contains contract info for deployment.
type Contract struct {
	Hash      util.Uint160
	AEF       *aef.File
	Manifest  *manifest.Manifest
	DebugInfo *compiler.DebugInfo
}

// contracts caches the compiled contracts from FS across multiple tests. The key is a
// concatenation of the source file path and the config file path split by | symbol.
var contracts = make(map[string]*Contract)

// CompileSource compiles a contract from the reader and returns its AEF, manifest and hash.
// Compiled contract will have "contract.go" used for its file name and coverage
// data collection can give wrong results for it, so it's recommended to disable
// coverage ([*Executor.DisableCoverage]) when you're deploying contracts
// compiled via this function.
func CompileSource(t testing.TB, sender util.Uint160, src io.Reader, opts *compiler.Options) *Contract {
	// aef.NewFile() cares about version a lot.
	config.Version = "atipicialtest"

	ne, di, err := compiler.CompileWithOptions("contract.go", src, opts)
	require.NoError(t, err)

	m, err := compiler.CreateManifest(di, opts)
	require.NoError(t, err)

	return &Contract{
		Hash:      state.CreateContractHash(sender, ne.Checksum, m.Name),
		AEF:       ne,
		Manifest:  m,
		DebugInfo: di,
	}
}

// CompileFile compiles a contract from the file and returns its AEF, manifest and hash.
// It uses contracts cache (keyed by source and config path) for AEF/manifest/debug
// info (hash is sender-dependent, so always recalculated according to parameter).
func CompileFile(t testing.TB, sender util.Uint160, srcPath string, configPath string) *Contract {
	cacheKey := srcPath + "|" + configPath
	if c, ok := contracts[cacheKey]; ok {
		var cc = *c
		cc.Hash = state.CreateContractHash(sender, c.AEF.Checksum, c.Manifest.Name)
		return &cc
	}

	// aef.NewFile() cares about version a lot.
	config.Version = "atipicialtest"

	conf, err := smartcontract.ParseContractConfig(configPath)
	require.NoError(t, err)

	o := &compiler.Options{}
	o.Name = conf.Name
	o.ContractEvents = conf.Events
	o.DeclaredNamedTypes = conf.NamedTypes
	o.ContractSupportedStandards = conf.SupportedStandards
	o.Permissions = make([]manifest.Permission, len(conf.Permissions))
	for i := range conf.Permissions {
		o.Permissions[i] = manifest.Permission(conf.Permissions[i])
	}
	o.SafeMethods = conf.SafeMethods
	o.Overloads = conf.Overloads
	o.SourceURL = conf.SourceURL
	ne, di, err := compiler.CompileWithOptions(srcPath, nil, o)
	require.NoError(t, err)
	m, err := compiler.CreateManifest(di, o)
	require.NoError(t, err)

	c := &Contract{
		Hash:      state.CreateContractHash(sender, ne.Checksum, m.Name),
		AEF:       ne,
		Manifest:  m,
		DebugInfo: di,
	}
	contracts[cacheKey] = c
	return c
}

// ReadAEF loads a contract from the specified AEF and manifest files.
func ReadAEF(t testing.TB, sender util.Uint160, aefPath, manifestPath string) *Contract {
	cacheKey := sender.StringLE() + "|" + aefPath + "|" + manifestPath
	if c, ok := contracts[cacheKey]; ok {
		return c
	}

	aefBytes, err := os.ReadFile(aefPath)
	require.NoError(t, err)

	ne, err := aef.FileFromBytes(aefBytes)
	require.NoError(t, err)

	manifestBytes, err := os.ReadFile(manifestPath)
	require.NoError(t, err)

	m := new(manifest.Manifest)
	err = json.Unmarshal(manifestBytes, m)
	require.NoError(t, err)

	hash := state.CreateContractHash(sender, ne.Checksum, m.Name)
	err = m.IsValid(hash, true)
	require.NoError(t, err)

	c := &Contract{
		Hash:     hash,
		AEF:      &ne,
		Manifest: m,
	}

	contracts[cacheKey] = c
	return c
}
