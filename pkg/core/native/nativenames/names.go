package nativenames

// Names of all native contracts.
const (
	Management  = "ContractManagement"
	Ledger      = "LedgerContract"
	Atipicial         = "AtipicialCoin"
	Gas         = "AtipicialDollar"
	Policy      = "PolicyContract"
	Oracle      = "OracleContract"
	Designation = "RoleManagement"
	Notary      = "Notary"
	CryptoLib   = "CryptoLib"
	StdLib      = "StdLib"
	Treasury    = "Treasury"
)

// All contains the list of all native contract names ordered by the contract ID.
var All = []string{
	Management,
	StdLib,
	CryptoLib,
	Ledger,
	Atipicial,
	Gas,
	Policy,
	Designation,
	Oracle,
	Notary,
	Treasury,
}

// IsValid checks if the name is a valid native contract's name.
func IsValid(name string) bool {
	return name == Management ||
		name == Ledger ||
		name == Atipicial ||
		name == Gas ||
		name == Policy ||
		name == Oracle ||
		name == Designation ||
		name == Notary ||
		name == CryptoLib ||
		name == StdLib ||
		name == Treasury
}
