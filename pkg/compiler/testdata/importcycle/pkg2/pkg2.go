package pkg2

import (
	"github.com/atipicial/atipicial-go/pkg/compiler/testdata/importcycle/pkg3"
)

var A int

func init() {
	pkg3.A = 2
}
