package signer

import (
	"testing"
	"time"

	"github.com/ak68a/agentid-core/pkg/key"
	"github.com/ak68a/agentid-core/pkg/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupTestKeys creates a set of test keys for delegation testing
func setupTestKeys(t *testing.T) (*key.AgentKey, *key.AgentKey, *key.AgentKey) {
	// Create three keys: root delegator, intermediate delegate, and final delegate
	rootKey, err := key.GenerateAgentKey()
	require.NoError(t, err, "Failed to generate root key")

	intermediateKey, err := key.GenerateAgentKey()
	require.NoError(t, err, "Failed to generate intermediate key")

	finalKey, err := key.GenerateAgentKey()
	require.NoError(t, err, "Failed to generate final key")

	return rootKey, intermediateKey, finalKey
}

// createTestDelegationClaim creates a test delegation claim
func createTestDelegationClaim(delegatorDID, delegateDID string, action, scope string) *models.DelegationClaim {
	now := time.Now().Unix()
	return &models.DelegationClaim{
		DelegatorDID: delegatorDID,
		DelegateDID:  delegateDID,
		Action:       action,
		Scope:        scope,
		IssuedAt:     now,
		ExpiresAt:    now + 3600, // 1 hour from now
		Nonce:        "test_nonce",
		MaxDepth:     2,
		CurrentDepth: 0,
		Type:         []string{"VerifiableCredential", "DelegationCredential"},
		Context:      []string{"https://www.w3.org/2018/credentials/v1"},
		Issuer:       delegatorDID,
		Subject:      delegateDID,
	}
}

// Deep copy helper for DelegationClaim
func deepCopyDelegationClaim(claim *models.DelegationClaim) *models.DelegationClaim {
	if claim == nil {
		return nil
	}
	copyClaim := *claim
	if claim.Proof != nil {
		proofCopy := *claim.Proof
		copyClaim.Proof = &proofCopy
	}
	if claim.ParentDelegation != nil {
		parentCopy := *claim.ParentDelegation
		copyClaim.ParentDelegation = &parentCopy
	}
	if claim.Type != nil {
		typeCopy := make([]string, len(claim.Type))
		copy(typeCopy, claim.Type)
		copyClaim.Type = typeCopy
	}
	if claim.Context != nil {
		contextCopy := make([]string, len(claim.Context))
		copy(contextCopy, claim.Context)
		copyClaim.Context = contextCopy
	}
	return &copyClaim
}

// Deep copy helper for DelegationChain
func deepCopyDelegationChain(chain *models.DelegationChain) *models.DelegationChain {
	if chain == nil {
		return nil
	}
	copyChain := &models.DelegationChain{
		Delegations: make([]*models.DelegationClaim, len(chain.Delegations)),
	}
	for i, claim := range chain.Delegations {
		copyChain.Delegations[i] = deepCopyDelegationClaim(claim)
	}
	return copyChain
}

func TestSignDelegationClaim(t *testing.T) {
	// Setup
	rootKey, _, _ := setupTestKeys(t)
	signer := NewClaimSigner(rootKey)

	// Create a test claim
	claim := createTestDelegationClaim(
		rootKey.DID,
		"did:ackid:0x742d35Cc6634C0532925a3b844Bc454e4438f44e",
		"transfer",
		"ETH",
	)

	// Test signing
	err := signer.SignDelegationClaim(claim)
	require.NoError(t, err, "Failed to sign claim")

	// Verify the proof was added
	assert.NotNil(t, claim.Proof, "Proof should be added to claim")
	assert.Equal(t, string(models.EcdsaSecp256k1Signature2019), claim.Proof.Type)
	assert.Equal(t, string(models.AssertionMethod), claim.Proof.ProofPurpose)
	assert.NotEmpty(t, claim.Proof.ProofValue, "Proof value should not be empty")
	assert.Equal(t, rootKey.DID+"#key-1", claim.Proof.VerificationMethod)

	// Verify the signature
	valid, err := signer.VerifyDelegationClaim(claim, rootKey.DID)
	require.NoError(t, err, "Failed to verify claim")
	assert.True(t, valid, "Signature should be valid")
}

func TestVerifyDelegationClaim(t *testing.T) {
	// Setup
	rootKey, delegateKey, _ := setupTestKeys(t)
	signer := NewClaimSigner(rootKey)

	// Create and sign a claim
	claim := createTestDelegationClaim(rootKey.DID, delegateKey.DID, "transfer", "ETH")
	err := signer.SignDelegationClaim(claim)
	require.NoError(t, err, "Failed to sign claim")

	// Test cases
	tests := []struct {
		name              string
		claim             *models.DelegationClaim
		expectedDelegator string
		wantValid         bool
		wantErr          bool
	}{
		{
			name:              "Valid signature",
			claim:             claim,
			expectedDelegator: rootKey.DID,
			wantValid:         true,
			wantErr:          false,
		},
		{
			name:              "Wrong delegator",
			claim:             claim,
			expectedDelegator: delegateKey.DID,
			wantValid:         false,
			wantErr:          true,
		},
		{
			name:              "No proof",
			claim:             createTestDelegationClaim(rootKey.DID, delegateKey.DID, "transfer", "ETH"),
			expectedDelegator: rootKey.DID,
			wantValid:         false,
			wantErr:          true,
		},
		{
			name:              "Modified claim",
			claim: func() *models.DelegationClaim {
				modified := *claim
				modified.Action = "read" // Modify the action
				return &modified
			}(),
			expectedDelegator: rootKey.DID,
			wantValid:         false,
			wantErr:          false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := signer.VerifyDelegationClaim(tt.claim, tt.expectedDelegator)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, tt.wantValid, valid)
		})
	}
}

func TestVerifyDelegationChain(t *testing.T) {
	// Setup
	rootKey, intermediateKey, finalKey := setupTestKeys(t)
	rootSigner := NewClaimSigner(rootKey)
	intermediateSigner := NewClaimSigner(intermediateKey)

	// Create a chain of delegations
	rootDelegation := createTestDelegationClaim(rootKey.DID, intermediateKey.DID, "transfer", "ETH")
	err := rootSigner.SignDelegationClaim(rootDelegation)
	require.NoError(t, err, "Failed to sign root delegation")

	intermediateDelegation := createTestDelegationClaim(intermediateKey.DID, finalKey.DID, "transfer", "ETH")
	intermediateDelegation.ParentDelegation = &rootDelegation.Nonce
	intermediateDelegation.CurrentDepth = 1
	err = intermediateSigner.SignDelegationClaim(intermediateDelegation)
	require.NoError(t, err, "Failed to sign intermediate delegation")

	baseChain := &models.DelegationChain{
		Delegations: []*models.DelegationClaim{rootDelegation, intermediateDelegation},
	}

	// Test cases
	tests := []struct {
		name      string
		chain     *models.DelegationChain
		wantValid bool
		wantErr   bool
	}{
		{
			name:      "Valid chain",
			chain:     deepCopyDelegationChain(baseChain),
			wantValid: true,
			wantErr:   false,
		},
		{
			name:      "Empty chain",
			chain:     &models.DelegationChain{Delegations: []*models.DelegationClaim{}},
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "Modified intermediate delegation",
			chain: func() *models.DelegationChain {
				modified := deepCopyDelegationChain(baseChain)
				modified.Delegations[1].Action = "read"
				return modified
			}(),
			wantValid: false,
			wantErr:   true,
		},
		{
			name: "Broken chain (wrong delegator)",
			chain: func() *models.DelegationChain {
				broken := deepCopyDelegationChain(baseChain)
				broken.Delegations[1].DelegatorDID = finalKey.DID
				return broken
			}(),
			wantValid: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			valid, err := rootSigner.VerifyDelegationChain(tt.chain)
			
			if tt.wantErr {
				assert.Error(t, err, "Expected an error for invalid chain")
				assert.False(t, valid, "Expected verification to be invalid when error occurs")
			} else {
				assert.NoError(t, err, "Unexpected error for valid chain")
				assert.True(t, valid, "Expected verification to be valid for valid chain")
			}
		})
	}
}

func TestSignDelegationClaim_NilKey(t *testing.T) {
	// A verification-only signer (nil key) must return an error, not panic
	signer := NewClaimSigner(nil)
	claim := createTestDelegationClaim(
		"did:ackid:0xDelegator",
		"did:ackid:0xDelegate",
		"transfer", "ETH",
	)

	err := signer.SignDelegationClaim(claim)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "verification-only mode")
}

func TestHashDelegationClaim(t *testing.T) {
	// Setup
	rootKey, delegateKey, _ := setupTestKeys(t)
	signer := NewClaimSigner(rootKey)

	// Create a test claim
	claim := createTestDelegationClaim(rootKey.DID, delegateKey.DID, "transfer", "ETH")

	// Test hashing
	hash1, err := signer.hashDelegationClaim(claim)
	require.NoError(t, err, "Failed to hash claim")
	assert.NotEmpty(t, hash1, "Hash should not be empty")

	// Test determinism
	hash2, err := signer.hashDelegationClaim(claim)
	require.NoError(t, err, "Failed to hash claim second time")
	assert.Equal(t, hash1, hash2, "Hashes should be deterministic")

	// Test that proof doesn't affect hash
	claim.Proof = &models.CredentialProof{
		Type:               string(models.EcdsaSecp256k1Signature2019),
		Created:            time.Now().Format(time.RFC3339),
		VerificationMethod: "test",
		ProofPurpose:       string(models.AssertionMethod),
		ProofValue:         "test",
	}
	hash3, err := signer.hashDelegationClaim(claim)
	require.NoError(t, err, "Failed to hash claim with proof")
	assert.Equal(t, hash1, hash3, "Hash should not change with proof")
} 