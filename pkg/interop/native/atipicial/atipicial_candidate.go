package atipicial

import "github.com/atipicial/atipicial-go/pkg/interop"

// Candidate represents a single native Atipicial candidate.
type Candidate struct {
	Key   interop.PublicKey
	Votes int
}
