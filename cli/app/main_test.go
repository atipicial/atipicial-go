package app_test

import (
	"testing"

	"github.com/atipicial/atipicial-go/internal/testcli"
	"github.com/atipicial/atipicial-go/internal/versionutil"
	"github.com/atipicial/atipicial-go/pkg/config"
)

func TestCLIVersion(t *testing.T) {
	config.Version = versionutil.TestVersion // Zero-length version string disables '--version' completely.
	e := testcli.NewExecutor(t, false)
	e.Run(t, "atipicial-go", "--version")
	e.CheckNextLine(t, "^AtipicialGo")
	e.CheckNextLine(t, "^Version:")
	e.CheckNextLine(t, "^GoVersion:")
	e.CheckEOF(t)
}
