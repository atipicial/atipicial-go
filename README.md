<div align="center">

# 🐹 ATIPICIAL GO

### A complete Atipicial Chain node and toolkit in Go

**Full node · CLI wallet · smart-contract compiler · RPC server — the Go implementation of the Atipicial Chain**

</div>

<p align="center">
  <img alt="Founder" src="https://img.shields.io/badge/%F0%9F%91%91_Founder-xmoohad-ff006e?style=for-the-badge">
  <img alt="Chain" src="https://img.shields.io/badge/Chain-Atipicial_L1-9d4edd?style=for-the-badge">
  <img alt="ATC" src="https://img.shields.io/badge/%F0%9F%AA%99_ATC-Atipicial_Coin-ffd60a?style=for-the-badge">
  <img alt="ATD" src="https://img.shields.io/badge/%F0%9F%92%B5_ATD-AtipicialDollar-06d6a0?style=for-the-badge">
  <img alt="License" src="https://img.shields.io/badge/License-MIT-3a86ff?style=for-the-badge">
</p>

---

## 🎯 What This Is

`atipicial-go` is a from-scratch Go implementation of the Atipicial Chain:
a fully-featured node that syncs MainNet, serves the JSON-RPC API, runs dBFT
consensus, executes AtipicialVM contracts, and ships with a complete CLI for
wallets, smart contracts, and network operations.

## ✨ Capabilities

| Area | Support |
|---|---|
| **Node** | Full MainNet/TestNet sync, P2P network, block & state storage |
| **Consensus** | dBFT 2.0 — single-block finality, view changes |
| **RPC Server** | Full JSON-RPC: blockchain, state, invocation, wallet methods |
| **CLI wallet** | Create wallets (AEP-6), transfer ATC/ATD, claim, vote |
| **Contract compiler** | Compile Go contracts to AEF, deploy & invoke |
| **AtipicialVM** | Full opcode + interop surface |
| **Oracle** | HTTPS + AtipicialFs request fulfilment |
| **Notary** | Notary transaction support |

## 🚀 Quick Start

```bash
# build
make build

# run a mainnet node
./bin/atipicial-go node --mainnet

# or with docker
docker run -d --name atipicial-go -p 10332:10332 -p 10333:10333 atipicial/atipicial-go

# create a wallet (addresses start with 'A')
./bin/atipicial-go wallet create

# check chain status
curl -X POST https://seed1.atipicial.com:10332 \
  -d '{"jsonrpc":"2.0","method":"getblockcount","id":1}'
```

## 🏗️ Repository Layout

```text
atipicial-go/
├── cli/              # the command-line interface (wallet, contract, node)
├── config/           # mainnet/testnet protocol configs
├── docs/             # node, consensus, RPC documentation
├── examples/         # contract & integration examples
├── pkg/              # core packages: core, vm, consensus, crypto, rpc, network
└── scripts/          # release & CI helpers
```

## ⛓️ Built for the Atipicial Chain

| | |
|---|---|
| **ATC** | Atipicial Coin — governance & staking · 1,000,000,000 total |
| **ATD** | AtipicialDollar — settlement & fees · 500,000,000 genesis |
| **Addresses** | Begin with capital **`A`** (version byte `0x09`) |
| **Standards** | AEP-17 (fungible) · AEP-11 (NFT) · AEP-6 (wallets) · AEP-2 (keys) |
| **Genesis** | 2026-07-20 00:00:00 UTC |
| **Mainnet RPC** | `https://seed1.atipicial.com:10332` |
| **Seeds (P2P)** | `seed1-5.atipicial.com:10333` |
| **Consensus** | dBFT 2.0 — single-block finality |

---

<div align="center">

## 👑 FOUNDER

### **xmoohad**

**Founder · Architect · Blockchain Scientist · Computer Programmer**

*Atipicial Chain is a sovereign Layer-1 blockchain for smart contracts,
digital assets, and decentralized applications.*

</div>

---

*© Atipicial Chain · Founded by xmoohad · MIT License*
