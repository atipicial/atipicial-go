package testchain

import "github.com/atipicial/atipicial-go/pkg/config/netmode"

// Network returns testchain network's magic number.
func Network() netmode.Magic {
	return netmode.UnitTestNet
}
