# agentid

![AgentID](https://img.shields.io/badge/AgentID-blue)
![Status](https://img.shields.io/badge/Status-Active%20Development-orange)
![License](https://img.shields.io/badge/License-Apache%202.0-green)

> **Active Development** — Core features are being implemented and tested. Not ready for production use.

Go + Solidity toolkit for on-chain agent identity verification, delegation, and revocation.

## Thesis

The agent commerce stack is forming around open protocols: [MPP](https://mpp.dev/) (Stripe/Tempo) for payments, [ERC-8004](https://eips.ethereum.org/EIPS/eip-8004) for discovery and reputation, and [ACK-ID](https://www.agentcommercekit.com/ack-id/introduction) for off-chain identity. But none of them answer: **is this agent authorized to do this specific thing?**

- MPP's model is "payment is the credential" — no identity layer
- ERC-8004 provides discovery and reputation but [explicitly cannot](https://eips.ethereum.org/EIPS/eip-8004) cryptographically verify advertised capabilities
- ACK-ID handles off-chain identity (controller credentials, A2A handshakes) but has no on-chain component

AgentID fills this gap: **on-chain delegation and authorization verification for autonomous agents.** It provides cryptographic proof that an agent was authorized by a specific principal, with scoped capabilities, depth-limited delegation chains, and instant revocation — verifiable in a smart contract during payment settlement.

### Where AgentID fits

| Question | Protocol |
|----------|----------|
| How do I find this agent? | [ERC-8004](https://eips.ethereum.org/EIPS/eip-8004) — on-chain registry, agent card, reputation |
| Who is this agent? | [ACK-ID](https://www.agentcommercekit.com/ack-id/introduction) — off-chain controller credentials, A2A handshakes |
| **Is this agent authorized to do this?** | **AgentID** — on-chain delegation chains, capability scoping, revocation |
| How does this agent pay? | [MPP](https://mpp.dev/) — HTTP 402, multi-method settlement |

### Integration strategy

AgentID is designed to plug into the existing ecosystem, not compete with it:

1. **ERC-8004**: Store AgentID delegation roots as agent metadata. Act as a [ValidationRegistry](https://eips.ethereum.org/EIPS/eip-8004) validator for delegation chain verification.
2. **MPP**: Verify agent delegation authority during payment settlement — atomic identity + payment in one transaction.
3. **ACK-ID**: Compatible key infrastructure (secp256k1, DIDs). On-chain complement to ACK-ID's off-chain credentials.

## Features

- Agent keypair generation (secp256k1, ACK-compatible DIDs)
- Structured delegation claim signing (EIP-712-ready)
- Delegation chains with depth limits and constraint propagation
- Solidity contracts for on-chain delegation verification
- Revocation claims and revocation list management

## Project Structure

```
agentid/
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
- [ERC-8004](https://eips.ethereum.org/EIPS/eip-8004) for the on-chain agent registry standard
- [MPP](https://mpp.dev/) for the machine payments protocol
