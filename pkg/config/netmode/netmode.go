package netmode

import "strconv"

const (
	// MainNet contains magic code used in the Atipicial main official network.
	MainNet Magic = 0x334f454e // ATIPICIAL
	// TestNet contains magic code used in the Atipicial testing network.
	TestNet Magic = 0x3554334e // N3T5
	// PrivNet contains magic code usually used for Atipicial private networks.
	PrivNet Magic = 56753 // docker privnet
	// UnitTestNet is a stub magic code used for testing purposes.
	UnitTestNet Magic = 42
	//MainNetAtipicialFs contains magic code used in the AtipicialFs main network.
	MainNetAtipicialFs Magic = 0x572dfa5 // AtipicialFs mainnet
	//TestNetAtipicialFs contains magic code used in the AtipicialFs test network.
	TestNetAtipicialFs Magic = 0x2bdb2b5f // AtipicialFs testnet
)

// Magic describes the network the blockchain will operate on.
type Magic uint32

// String implements the stringer interface.
func (n Magic) String() string {
	switch n {
	case PrivNet:
		return "privnet"
	case TestNet:
		return "testnet"
	case MainNet:
		return "mainnet"
	case UnitTestNet:
		return "unit_testnet"
	case MainNetAtipicialFs:
		return "mainnet.atipicialfs"
	case TestNetAtipicialFs:
		return "testnet.atipicialfs"
	default:
		return "net 0x" + strconv.FormatUint(uint64(n), 16)
	}
}
