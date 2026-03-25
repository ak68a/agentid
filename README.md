# agentid

![AgentID](https://img.shields.io/badge/AgentID-blue)
![Status](https://img.shields.io/badge/Status-Active%20Development-orange)
![License](https://img.shields.io/badge/License-Apache%202.0-green)

> **Active Development** — Core features are being implemented and tested. Not ready for production use.

Go + Solidity toolkit for verifiable agent identity, built on the [ACK-ID specification](https://www.agentcommercekit.com/ack-id/introduction).

## Thesis

Machine-to-machine payment protocols like [MPP](https://mpp.dev/) (Stripe/Tempo) and [ACK-Pay](https://www.agentcommercekit.com/ack-pay/introduction) solve *how* agents pay each other, but not *who* is paying or *who authorized it*. MPP's model is "payment is the credential" — there's no identity layer. ACK-ID handles identity off-chain via controller credentials and A2A handshakes, but has no on-chain component.

agentid-core fills this gap: **on-chain identity verification for autonomous agents**. It provides the building blocks to verify an agent's identity, check its delegation authority, and enforce revocation — all within a smart contract during payment settlement. No trusted third party, no off-chain API call, no replay window.

This matters when:
- Agents transact with agents they've never seen before (no prior handshake)
- Delegation chains need to be publicly auditable (who authorized this agent to spend?)
- Revocation must be instant and global, not dependent on checking a list
- Identity verification and payment should be atomic (one transaction, all-or-nothing)

## Features

- Agent keypair generation (secp256k1, ACK-compatible DIDs)
- Structured delegation claim signing (EIP-712-ready)
- Delegation chains with depth limits and constraint propagation
- Solidity contracts for on-chain identity registry and delegation verification
- Revocation claims and revocation list management

## Project Structure

```
agentid-core/
├── pkg/
│   ├── key/            # Agent keypair generation and management
│   ├── models/         # Claims, delegations, revocations, constraints
│   └── signer/         # Claim signing and verification
├── cmd/agentid/        # CLI: generate agents, create/verify claims
├── contracts/
│   └── src/            # AgentRegistry.sol, AgentDelegation.sol
└── docs/
```

## Quick Start

```bash
go get github.com/ak68a/agentid
```

```go
import (
    "github.com/ak68a/agentid/pkg/key"
    "github.com/ak68a/agentid/pkg/models"
    "github.com/ak68a/agentid/pkg/signer"
)

// Generate an agent identity
agentKey, _ := key.GenerateAgentKey()
// agentKey.DID = "did:ackid:0x..."

// Create a delegation claim
claim := &models.DelegationClaim{
    DelegatorDID: ownerKey.DID,
    DelegateDID:  agentKey.DID,
    Action:       "transfer",
    Scope:        "ETH",
    MaxDepth:     2,
}

// Sign it
s := signer.NewClaimSigner(ownerKey)
s.SignDelegationClaim(claim)

// Verify it (no private key needed)
verifier := signer.NewClaimSigner(nil)
valid, _ := verifier.VerifyDelegationClaim(claim, ownerKey.DID)
```

## Development

### Requirements

- Go 1.24.3+
- Foundry (for Solidity)
  ```bash
  curl -L https://foundry.paradigm.xyz | bash && foundryup
  ```

### Testing

```bash
go test ./...           # Go tests
cd contracts && forge test  # Solidity tests
```

## License

Apache License 2.0 — see [LICENSE](./LICENSE).

## Acknowledgments

- [Agent Commerce Kit](https://www.agentcommercekit.com) for the ACK-ID specification
- [Ethereum Foundation](https://ethereum.org) for EIP-712
