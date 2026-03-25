# agentid-core Roadmap

Go + Solidity reference implementation of ACK-ID with unique delegation chains and smart contracts.

**Repo**: `ak68a/agentid-core`
**Location**: `/Users/vayu/Documents/Dev/projects/agentcommercekit/agentid-core`

---

## Phase 1: Critical Fixes — DONE

| # | Task | Effort | Status |
|---|------|--------|--------|
| 1.1 | Remove 7 debug `fmt.Printf()` calls in signer.go leaking crypto data | Small | DONE ✓ |
| 1.2 | Add nil guard on `SignDelegationClaim()` — returns error instead of panic in verification-only mode | Small | DONE ✓ |
| 1.3 | Delete `contracts_temp/` placeholder (Counter.sol, not needed) | Trivial | DONE ✓ |

All three in commit `ce79115` on branch `fix/phase-1-critical`.

---

## Phase 2: Test Foundation ← NEXT

| # | Task | Effort | Status |
|---|------|--------|--------|
| 2.1 | Models package tests — AgentClaim, OwnershipClaim, DelegationClaim, RevocationClaim | Medium | DONE ✓ |
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

## Phase 4: Integration & Advanced

| # | Task | Effort | Status |
|---|------|--------|--------|
| 4.1 | Storage interface — `AgentStore` abstraction (memory, file, postgres). Unblocks `GetChain()` | Large | TODO |
| 4.2 | Implement parent delegation fetching in `GetChain()` (currently returns "not implemented") | Medium | TODO |
| 4.3 | Go-Solidity bridge tests — verify Go-signed delegations work with Solidity `ecrecover` | Large | TODO |
| 4.4 | Solidity invariant tests — Foundry fuzz tests for registry/delegation state consistency | Medium | TODO |
| 4.5 | Go verifier for on-chain state — call Solidity contracts from Go to verify delegations | Large | TODO |

---

## Phase 5: Polish

| # | Task | Effort | Status |
|---|------|--------|--------|
| 5.1 | CONTRIBUTING.md + CODE_OF_CONDUCT.md (referenced in README but don't exist) | Small | TODO |

---

## Key Context

- **Unique value**: Delegation chains with depth limits + Solidity contracts — ACK TypeScript repo has neither
- **Complementary to ACK**: Go + on-chain vs TypeScript + off-chain
- **Zero test coverage currently** — Phase 2 is high priority after fixing the bugs
