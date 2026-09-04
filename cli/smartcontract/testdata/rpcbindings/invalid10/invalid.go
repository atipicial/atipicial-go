package invalid10

import (
	"github.com/atipicial/atipicial-go/pkg/interop/runtime"
	"github.com/atipicial/atipicial-go/pkg/interop/storage"
)

func f() any {
	return 1
}

func Main() {
	ctx := storage.GetContext()
	// Both arguments have the same real (Go) type "any" (see the default case in
	// scAndVMInteropTypeFromExpr for storage.Context), but their SC types differ:
	// InteropInterface vs Any.
	runtime.Notify("SomeEvent", ctx)
	runtime.Notify("SomeEvent", f())
}
