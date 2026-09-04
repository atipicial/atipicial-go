package address

import (
	"errors"

	"github.com/atipicial/atipicial-go/pkg/encoding/base58"
	"github.com/atipicial/atipicial-go/pkg/util"
)

const (
	// ATC2Prefix is the first byte of an address for ATC2.
	ATC2Prefix byte = 0x17
	// ATIPICIALPrefix is the first byte of an address for ATIPICIAL.
	ATIPICIALPrefix byte = 0x35
)

// Prefix is the byte used to prepend to addresses when encoding them, it can
// be changed and defaults to 53 (0x35), the standard ATC prefix.
var Prefix = ATIPICIALPrefix

// Uint160ToString returns the "ATC address" from the given Uint160.
func Uint160ToString(u util.Uint160) string {
	// Don't forget to prepend the Address version 0x17 (23) A
	b := append([]byte{Prefix}, u.BytesBE()...)
	return base58.CheckEncode(b)
}

// StringToUint160 attempts to decode the given ATC address string
// into a Uint160.
func StringToUint160(s string) (u util.Uint160, err error) {
	b, err := base58.CheckDecode(s)
	if err != nil {
		return u, err
	}
	if b[0] != Prefix {
		return u, errors.New("wrong address prefix")
	}
	return util.Uint160DecodeBytesBE(b[1:21])
}
