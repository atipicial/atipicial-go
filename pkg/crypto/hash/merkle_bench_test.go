package hash_test

import (
	"testing"

	"github.com/atipicial/atipicial-go/internal/random"
	"github.com/atipicial/atipicial-go/pkg/crypto/hash"
	"github.com/atipicial/atipicial-go/pkg/util"
	"github.com/stretchr/testify/require"
)

func BenchmarkMerkle(t *testing.B) {
	var hashes = make([]util.Uint256, 100000)
	for i := range hashes {
		hashes[i] = random.Uint256()
	}

	t.Run("NewMerkleTree", func(t *testing.B) {
		for t.Loop() {
			tr, err := hash.NewMerkleTree(hashes)
			require.NoError(t, err)
			_ = tr.Root()
		}
	})
	t.Run("CalcMerkleRoot", func(t *testing.B) {
		for t.Loop() {
			_ = hash.CalcMerkleRoot(hashes)
		}
	})
}
