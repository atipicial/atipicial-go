package wallet

import (
	"encoding/json"
	"testing"

	"github.com/atipicial/atipicial-go/pkg/smartcontract/manifest"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/stretchr/testify/require"
)

func TestToken_MarshalJSON(t *testing.T) {
	// From the https://atipicial-python.readthedocs.io/en/latest/prompt.html#import-aep5-compliant-token
	h, err := util.Uint160DecodeStringLE("f8d448b227991cf07cb96a6f9c0322437f1599b9")
	require.NoError(t, err)

	tok := NewToken(h, "AEP-17 standard token", "AEPT", 8, manifest.AEP17StandardName)
	require.Equal(t, "AEP-17 standard token", tok.Name)
	require.Equal(t, "AEPT", tok.Symbol)
	require.EqualValues(t, 8, tok.Decimals)
	require.Equal(t, h, tok.Hash)
	require.Equal(t, "NcqKahsZ93ZyYS5bep8G2TY1zRB7tfUPdK", tok.Address())

	data, err := json.Marshal(tok)
	require.NoError(t, err)

	actual := new(Token)
	require.NoError(t, json.Unmarshal(data, actual))
	require.Equal(t, tok, actual)
}
