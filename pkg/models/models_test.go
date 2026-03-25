package models

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- Helpers ---

func futureUnix(d time.Duration) int64 {
	return time.Now().Add(d).Unix()
}

func pastUnix(d time.Duration) int64 {
	return time.Now().Add(-d).Unix()
}

// --- AgentClaim ---

func TestAgentClaim_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt int64
		want      bool
	}{
		{"zero means never expires", 0, false},
		{"future expiry is not expired", futureUnix(time.Hour), false},
		{"past expiry is expired", pastUnix(time.Hour), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim := &AgentClaim{ExpiresAt: tt.expiresAt}
			assert.Equal(t, tt.want, claim.IsExpired())
		})
	}
}

func TestAgentClaim_ToCredential(t *testing.T) {
	claim := NewAgentClaim(
		"did:ackid:0xAgent", "did:ackid:0xOwner",
		ActionTransfer, ScopeETH,
		futureUnix(time.Hour), "nonce-1",
	)
	claim.MaxAmount = "1.5"

	vc := claim.ToCredential()

	assert.Equal(t, StandardContexts, vc["@context"])
	assert.Equal(t, AgentAuthorizationCredentialType, vc["type"])
	assert.Equal(t, "did:ackid:0xOwner", vc["issuer"])

	subject := vc["credentialSubject"].(map[string]interface{})
	assert.Equal(t, "did:ackid:0xAgent", subject["agentDID"])
	assert.Equal(t, "transfer", subject["action"])
	assert.Equal(t, "1.5", subject["maxAmount"])
}

// --- OwnershipClaim ---

func TestOwnershipClaim_IsExpired(t *testing.T) {
	t.Run("zero means never expires", func(t *testing.T) {
		claim := NewOwnershipClaim("did:ackid:0xA", "did:ackid:0xO", "n1")
		assert.False(t, claim.IsExpired())
	})

	t.Run("past expiry is expired", func(t *testing.T) {
		claim := &OwnershipClaim{ExpiresAt: pastUnix(time.Hour)}
		assert.True(t, claim.IsExpired())
	})
}

// --- DelegationClaim ---

func TestDelegationClaim_IsExpired(t *testing.T) {
	tests := []struct {
		name      string
		expiresAt int64
		want      bool
	}{
		{"zero means never expires", 0, false},
		{"future expiry is not expired", futureUnix(time.Hour), false},
		{"past expiry is expired", pastUnix(time.Hour), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim := &DelegationClaim{ExpiresAt: tt.expiresAt}
			assert.Equal(t, tt.want, claim.IsExpired())
		})
	}
}

func TestDelegationClaim_CanSubDelegate(t *testing.T) {
	tests := []struct {
		name         string
		currentDepth int
		maxDepth     int
		want         bool
	}{
		{"depth below max allows sub-delegation", 0, 2, true},
		{"depth at max blocks sub-delegation", 2, 2, false},
		{"depth above max blocks sub-delegation", 3, 2, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			claim := &DelegationClaim{CurrentDepth: tt.currentDepth, MaxDepth: tt.maxDepth}
			assert.Equal(t, tt.want, claim.CanSubDelegate())
		})
	}
}

func TestDelegationClaim_GetTimeConstraint(t *testing.T) {
	t.Run("returns nil when no time constraint", func(t *testing.T) {
		claim := &DelegationClaim{Constraints: map[string]interface{}{}}
		assert.Nil(t, claim.GetTimeConstraint())
	})

	t.Run("extracts time constraint from constraints map", func(t *testing.T) {
		claim := &DelegationClaim{
			Constraints: map[string]interface{}{
				"time": map[string]interface{}{
					"valid_from":  float64(1000),
					"valid_until": float64(2000),
					"timezone":    "UTC",
					"days":        []interface{}{float64(1), float64(2), float64(3)},
					"hours":       []interface{}{float64(9), float64(17)},
				},
			},
		}

		tc := claim.GetTimeConstraint()
		require.NotNil(t, tc)
		assert.Equal(t, int64(1000), tc.ValidFrom)
		assert.Equal(t, int64(2000), tc.ValidUntil)
		assert.Equal(t, "UTC", tc.TimeZone)
		assert.Equal(t, []int{1, 2, 3}, tc.Days)
		assert.Equal(t, []int{9, 17}, tc.Hours)
	})
}

func TestDelegationClaim_GetScopeConstraint(t *testing.T) {
	t.Run("returns nil when no scope constraint", func(t *testing.T) {
		claim := &DelegationClaim{Constraints: map[string]interface{}{}}
		assert.Nil(t, claim.GetScopeConstraint())
	})

	t.Run("extracts scope constraint from constraints map", func(t *testing.T) {
		claim := &DelegationClaim{
			Constraints: map[string]interface{}{
				"scope": map[string]interface{}{
					"allowed_resources": []interface{}{"ETH", "BTC"},
					"denied_resources":  []interface{}{"DOGE"},
				},
			},
		}

		sc := claim.GetScopeConstraint()
		require.NotNil(t, sc)
		assert.Equal(t, []string{"ETH", "BTC"}, sc.AllowedResources)
		assert.Equal(t, []string{"DOGE"}, sc.DeniedResources)
	})
}

// --- DelegationChain ---

func makeTwoLinkChain() *DelegationChain {
	now := time.Now().Unix()
	return &DelegationChain{
		Delegations: []*DelegationClaim{
			{
				DelegatorDID: "did:ackid:0xRoot",
				DelegateDID:  "did:ackid:0xMiddle",
				ExpiresAt:    now + 3600,
				MaxDepth:     2,
				CurrentDepth: 0,
			},
			{
				DelegatorDID: "did:ackid:0xMiddle",
				DelegateDID:  "did:ackid:0xLeaf",
				ExpiresAt:    now + 3600,
				MaxDepth:     2,
				CurrentDepth: 1,
			},
		},
	}
}

func TestDelegationChain_ValidateChain(t *testing.T) {
	t.Run("valid two-link chain", func(t *testing.T) {
		chain := makeTwoLinkChain()
		assert.True(t, chain.ValidateChain())
		assert.True(t, chain.Valid)
	})

	t.Run("empty chain is invalid", func(t *testing.T) {
		chain := &DelegationChain{Delegations: []*DelegationClaim{}}
		assert.False(t, chain.ValidateChain())
		assert.Equal(t, "empty delegation chain", chain.Reason)
	})

	t.Run("expired delegation fails validation", func(t *testing.T) {
		chain := makeTwoLinkChain()
		chain.Delegations[0].ExpiresAt = pastUnix(time.Hour)
		assert.False(t, chain.ValidateChain())
		assert.Contains(t, chain.Reason, "expired")
	})

	t.Run("depth exceeding max fails validation", func(t *testing.T) {
		chain := makeTwoLinkChain()
		chain.Delegations[1].CurrentDepth = 5
		chain.Delegations[1].MaxDepth = 2
		assert.False(t, chain.ValidateChain())
		assert.Contains(t, chain.Reason, "exceeds max depth")
	})

	t.Run("broken chain continuity fails validation", func(t *testing.T) {
		chain := makeTwoLinkChain()
		// Second delegation's delegator doesn't match first's delegate
		chain.Delegations[1].DelegatorDID = "did:ackid:0xStranger"
		assert.False(t, chain.ValidateChain())
		assert.Contains(t, chain.Reason, "broken chain")
	})
}

func TestDelegationChain_GetRootAndLeaf(t *testing.T) {
	chain := makeTwoLinkChain()

	assert.Equal(t, "did:ackid:0xRoot", chain.GetRootDelegation().DelegatorDID)
	assert.Equal(t, "did:ackid:0xLeaf", chain.GetLeafDelegation().DelegateDID)

	// Empty chain returns nil for both
	empty := &DelegationChain{}
	assert.Nil(t, empty.GetRootDelegation())
	assert.Nil(t, empty.GetLeafDelegation())
}

func TestDelegationChain_RevokeChain(t *testing.T) {
	chain := makeTwoLinkChain()
	revocations := chain.RevokeChain("compromised")

	assert.Len(t, revocations, 2)
	for _, rev := range revocations {
		assert.Equal(t, "compromised", rev.Reason)
		// Root delegator should be the revoker
		assert.Equal(t, "did:ackid:0xRoot", rev.RevokerDID)
		assert.Equal(t, RevocationCredentialType, rev.Type)
	}
	// First revocation is for the middle delegate, second for the leaf
	assert.Equal(t, "did:ackid:0xMiddle", revocations[0].RevokedAgentDID)
	assert.Equal(t, "did:ackid:0xLeaf", revocations[1].RevokedAgentDID)
}

// --- RevocationClaim ---

func TestRevocationClaim_IsEffective(t *testing.T) {
	tests := []struct {
		name        string
		revokedAt   int64
		effectiveAt int64
		want        bool
	}{
		{"effective when effectiveAt is zero and revokedAt is past", pastUnix(time.Hour), 0, true},
		{"not effective when effectiveAt is in the future", pastUnix(time.Hour), futureUnix(time.Hour), false},
		{"effective when effectiveAt is in the past", pastUnix(time.Hour), pastUnix(30 * time.Minute), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rc := &RevocationClaim{RevokedAt: tt.revokedAt, EffectiveAt: tt.effectiveAt}
			assert.Equal(t, tt.want, rc.IsEffective())
		})
	}
}

// --- RevocationList ---

func TestRevocationList_IsRevoked(t *testing.T) {
	rl := &RevocationList{
		Revocations: []*RevocationClaim{
			{RevokedCredentialID: "cred-1", RevokedAt: pastUnix(time.Hour), EffectiveAt: 0},
			{RevokedCredentialID: "cred-2", RevokedAt: pastUnix(time.Hour), EffectiveAt: futureUnix(time.Hour)},
		},
	}

	// cred-1 is effective now
	assert.NotNil(t, rl.IsRevoked("cred-1"))
	// cred-2 is revoked but not yet effective
	assert.Nil(t, rl.IsRevoked("cred-2"))
	// cred-3 was never revoked
	assert.Nil(t, rl.IsRevoked("cred-3"))
}

func TestRevocationList_IsAgentRevoked(t *testing.T) {
	rl := &RevocationList{
		Revocations: []*RevocationClaim{
			{RevokedAgentDID: "did:ackid:0xBad", RevokedAt: pastUnix(time.Hour), EffectiveAt: 0},
			{RevokedAgentDID: "did:ackid:0xBad", RevokedAt: pastUnix(30 * time.Minute), EffectiveAt: 0},
			{RevokedAgentDID: "did:ackid:0xGood", RevokedAt: pastUnix(time.Hour), EffectiveAt: futureUnix(time.Hour)},
		},
	}

	assert.Len(t, rl.IsAgentRevoked("did:ackid:0xBad"), 2)
	assert.Len(t, rl.IsAgentRevoked("did:ackid:0xGood"), 0) // not yet effective
	assert.Len(t, rl.IsAgentRevoked("did:ackid:0xClean"), 0)
}

func TestRevocationList_AddRevocation(t *testing.T) {
	rl := &RevocationList{}
	rl.AddRevocation(&RevocationClaim{RevokedCredentialID: "cred-1"})

	assert.Len(t, rl.Revocations, 1)
	assert.NotZero(t, rl.LastUpdated)
}

func TestRevocationList_GetRevocationsSince(t *testing.T) {
	now := time.Now().Unix()
	rl := &RevocationList{
		Revocations: []*RevocationClaim{
			{RevokedCredentialID: "old", RevokedAt: now - 7200},
			{RevokedCredentialID: "recent", RevokedAt: now - 1800},
			{RevokedCredentialID: "newest", RevokedAt: now - 60},
		},
	}

	recent := rl.GetRevocationsSince(now - 3600)
	assert.Len(t, recent, 2)
	assert.Equal(t, "recent", recent[0].RevokedCredentialID)
	assert.Equal(t, "newest", recent[1].RevokedCredentialID)
}

// --- ChainError ---

func TestChainError(t *testing.T) {
	err := &ChainError{Code: "INVALID_CHAIN", Message: "broken link"}
	assert.Equal(t, "INVALID_CHAIN: broken link", err.Error())
}

// --- Helper constructors ---

func TestNewAgentClaim(t *testing.T) {
	claim := NewAgentClaim(
		"did:ackid:0xAgent", "did:ackid:0xOwner",
		ActionTransfer, ScopeETH,
		futureUnix(time.Hour), "nonce-1",
	)

	assert.Equal(t, "did:ackid:0xAgent", claim.AgentDID)
	assert.Equal(t, "did:ackid:0xOwner", claim.OwnerDID)
	assert.Equal(t, ActionTransfer, claim.Action)
	assert.Equal(t, ScopeETH, claim.Scope)
	assert.Equal(t, AgentAuthorizationCredentialType, claim.Type)
	assert.Equal(t, StandardContexts, claim.Context)
	// Issuer should be the owner, subject should be the agent
	assert.Equal(t, "did:ackid:0xOwner", claim.Issuer)
	assert.Equal(t, "did:ackid:0xAgent", claim.Subject)
	assert.NotZero(t, claim.IssuedAt)
}

func TestNewTransferClaim(t *testing.T) {
	claim := NewTransferClaim(
		"did:ackid:0xA", "did:ackid:0xO",
		ScopeETH, "10.0",
		futureUnix(time.Hour), "n1",
	)
	assert.Equal(t, ActionTransfer, claim.Action)
	assert.Equal(t, "10.0", claim.MaxAmount)
}

func TestNewOwnershipClaim(t *testing.T) {
	claim := NewOwnershipClaim("did:ackid:0xA", "did:ackid:0xO", "n1")

	assert.Equal(t, int64(0), claim.ExpiresAt) // never expires by default
	assert.Equal(t, AgentOwnershipCredentialType, claim.Type)
	assert.Equal(t, "did:ackid:0xO", claim.Issuer)
	assert.Equal(t, "did:ackid:0xA", claim.Subject)
}

func TestNewCredentialProof(t *testing.T) {
	proof := NewCredentialProof(EcdsaSecp256k1Signature2019, AssertionMethod)

	assert.Equal(t, string(EcdsaSecp256k1Signature2019), proof.Type)
	assert.Equal(t, string(AssertionMethod), proof.ProofPurpose)
	assert.NotEmpty(t, proof.Created)
	assert.Equal(t, "AgentID", proof.Domain.Name)
	assert.Equal(t, int64(1), proof.Domain.ChainID)
}
