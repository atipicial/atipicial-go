package actor_test

import (
	"testing"

	"github.com/atipicial/atipicial-go/pkg/rpcclient"
	"github.com/atipicial/atipicial-go/pkg/rpcclient/actor"
)

func TestRPCActorRPCClientCompat(t *testing.T) {
	_ = actor.RPCActor(&rpcclient.WSClient{})
	_ = actor.RPCActor(&rpcclient.Client{})
}
