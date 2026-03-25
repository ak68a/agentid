# agentid Roadmap

On-chain identity verification and delegation for autonomous agents. Plugs into ERC-8004 (discovery), MPP (payments), and ACK-ID (off-chain identity).

**Repo**: `ak68a/agentid`
**Location**: `/Users/vayu/Documents/Dev/projects/agentcommercekit/agentid`

---

## Phase 1: Critical Fixes — DONE

| # | Task | Effort | Status |
|---|------|--------|--------|
| 1.1 | Remove 7 debug `fmt.Printf()` calls in signer.go leaking crypto data | Small | DONE |
| 1.2 | Add nil guard on `SignDelegationClaim()` — returns error instead of panic in verification-only mode | Small | DONE |
| 1.3 | Delete `contracts_temp/` placeholder (Counter.sol, not needed) | Trivial | DONE |

---

## Phase 2: Test Foundation ← NEXT

| # | Task | Effort | Status |
|---|------|--------|--------|
| 2.1 | Models package tests — AgentClaim, OwnershipClaim, DelegationClaim, RevocationClaim | Medium | DONE |
| 2.2 | CLI integration tests — generate, create-claim, verify-claim workflows | Medium | TODO |
| 2.3 | GitHub Actions CI — `go test` + `forge test` on PR | Small | TODO |

---

## Phase 3: Solidity Feature Parity

| # | Task | Effort | Status |
|---|------|--------|--------|
| 3.1 | Delegation depth limiting in Solidity (Go has it, Solidity allows unlimited nesting) | Small | TODO |
| 3.2 | Revocation Registry contract (modeled in Go, missing from Solidity) | Large | TODO |
| 3.3 | Time-based constraints in Solidity (Go has day-of-week/hour, Solidity only validFrom/validUntil) | Medium | TODO |

---

## Phase 4: ERC-8004 Integration

| # | Task | Effort | Status |
|---|------|--------|--------|
| 4.1 | Store AgentID DIDs as ERC-8004 metadata — write Go helper to call `setMetadata(agentId, "agentid:did", ...)` | Medium | TODO |
| 4.2 | AgentID ValidationRegistry validator — Solidity contract that verifies delegation chains and posts validation responses to ERC-8004 | Large | TODO |
| 4.3 | Delegation root storage — store delegation chain roots as ERC-8004 metadata for on-chain discoverability | Medium | TODO |
| 4.4 | Go SDK for reading agent identity from ERC-8004 registry + verifying delegation chains | Large | TODO |

---

## Phase 5: Core Infrastructure

| # | Task | Effort | Status |
|---|------|--------|--------|
| 5.1 | Storage interface — `AgentStore` abstraction (memory, file, postgres). Unblocks `GetChain()` | Large | TODO |
| 5.2 | Implement parent delegation fetching in `GetChain()` (currently returns "not implemented") | Medium | TODO |
| 5.3 | Go-Solidity bridge tests — verify Go-signed delegations work with Solidity `ecrecover` | Large | TODO |
| 5.4 | Solidity invariant tests — Foundry fuzz tests for registry/delegation state consistency | Medium | TODO |

---

## Key Context

- **Positioning**: Authorization layer that plugs into ERC-8004 (discovery) + MPP (payments) + ACK-ID (off-chain identity)
- **ERC-8004 gaps AgentID fills**: No capability scoping, no cryptographic proof of authorization, no granular revocation, no delegation chains
- **Competitors**: Chitin (soul identity + W3C DIDs), MolTrust (VCs on Base) — both early, neither has delegation chains
- **Not competing with**: ERC-8004 (discovery), MPP (payments), ACK-ID (off-chain identity) — complementary to all three
