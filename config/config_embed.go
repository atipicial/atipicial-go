// Package config contains embedded YAML configuration files for different network modes
// of the Atipicial  blockchain and for AtipicialFs mainnet and testnet networks.
package config

import (
	_ "embed"
)

// MainNet is the Atipicial  mainnet configuration.
//
//go:embed protocol.mainnet.yml
var MainNet []byte

// TestNet is the Atipicial  testnet configuration.
//
//go:embed protocol.testnet.yml
var TestNet []byte

// PrivNet is the private network configuration.
//
//go:embed protocol.privnet.yml
var PrivNet []byte

// MainNetAtipicialFs is the mainnet AtipicialFs configuration.
//
//go:embed protocol.mainnet.atipicialfs.yml
var MainNetAtipicialFs []byte

// TestNetAtipicialFs is the testnet AtipicialFs configuration.
//
//go:embed protocol.testnet.atipicialfs.yml
var TestNetAtipicialFs []byte

// UnitTestNet is the unit test network configuration.
//
//go:embed protocol.unit_testnet.yml
var UnitTestNet []byte
