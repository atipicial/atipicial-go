package invalid3

import "github.com/atipicial/atipicial-go/pkg/interop/runtime"

func Main() {
	runtime.Notify("SomeEvent", "p1", 5)
}
