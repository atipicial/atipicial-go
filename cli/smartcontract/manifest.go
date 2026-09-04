package smartcontract

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/atipicial/atipicial-go/cli/cmdargs"
	"github.com/atipicial/atipicial-go/cli/flags"
	"github.com/atipicial/atipicial-go/cli/options"
	"github.com/atipicial/atipicial-go/pkg/core/state"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
	"github.com/atipicial/atipicial-go/pkg/smartcontract/aef"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/urfave/cli/v2"
)

func manifestAddGroup(ctx *cli.Context) error {
	if err := cmdargs.EnsureNone(ctx); err != nil {
		return err
	}
	sender := ctx.Generic("sender").(*flags.Address)
	nf, _, err := readAEFFile(ctx.String("aef"))
	if err != nil {
		return cli.Exit(fmt.Errorf("can't read AEF file: %w", err), 1)
	}

	mPath := ctx.String("manifest")
	m, _, err := readManifest(mPath, util.Uint160{})
	if err != nil {
		return cli.Exit(fmt.Errorf("can't read contract manifest: %w", err), 1)
	}

	h := state.CreateContractHash(sender.Uint160(), nf.Checksum, m.Name)

	gAcc, w, err := options.GetAccFromContext(ctx)
	if err != nil {
		return cli.Exit(fmt.Errorf("can't get account to sign group with: %w", err), 1)
	}
	defer w.Close()

	var found bool

	sig := gAcc.PrivateKey().Sign(h.BytesBE())
	pub := gAcc.PublicKey()
	for i := range m.Groups {
		if m.Groups[i].PublicKey.Equal(pub) {
			m.Groups[i].Signature = sig
			found = true
			break
		}
	}
	if !found {
		m.Groups = append(m.Groups, manifest.Group{
			PublicKey: pub,
			Signature: sig,
		})
	}

	rawM, err := json.Marshal(m)
	if err != nil {
		return cli.Exit(fmt.Errorf("can't marshal manifest: %w", err), 1)
	}

	err = os.WriteFile(mPath, rawM, os.ModePerm)
	if err != nil {
		return cli.Exit(fmt.Errorf("can't write manifest file: %w", err), 1)
	}
	return nil
}

func readAEFFile(filename string) (*aef.File, []byte, error) {
	f, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	aefFile, err := aef.FileFromBytes(f)
	if err != nil {
		return nil, nil, fmt.Errorf("can't parse AEF file: %w", err)
	}

	return &aefFile, f, nil
}

// readManifest unmarshalls manifest got from the provided filename and checks
// it for validness against the provided contract hash. If empty hash is specified
// then no hash-related manifest groups check is performed.
func readManifest(filename string, hash util.Uint160) (*manifest.Manifest, []byte, error) {
	manifestBytes, err := os.ReadFile(filename)
	if err != nil {
		return nil, nil, err
	}

	m := new(manifest.Manifest)
	err = json.Unmarshal(manifestBytes, m)
	if err != nil {
		return nil, nil, err
	}
	if err := m.IsValid(hash, true); err != nil {
		return nil, nil, fmt.Errorf("manifest is invalid: %w", err)
	}
	return m, manifestBytes, nil
}
