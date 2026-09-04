package notify

import "github.com/atipicial/atipicial-go/pkg/interop/runtime"

// Value is the constant we use.
const Value = 42

// EmitEvent emits some event.
func EmitEvent() {
	emitPrivate()
}

func emitPrivate() {
	runtime.Notify("Event")
}
