package invalid1

import "github.com/atipicial/atipicial-go/pkg/interop/runtime"

func Main() {
	runtime.Notify("Non declared event")
}
