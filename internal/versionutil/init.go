package versionutil

import "github.com/atipicial/atipicial-go/pkg/config"

// TestVersion is a AtipicialGo version that should be used to keep all
// compiled AEFs the same from run to run for tests.
const TestVersion = "0.90.0-test"

// init sets config.Version to a dummy TestVersion value to keep contract AEFs
// consistent between test runs for those packages who import it. For test usage
// only!
func init() {
	config.Version = TestVersion
}
