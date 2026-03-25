package signer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/ak68a/agentid-core/pkg/key"
	"github.com/ak68a/agentid-core/pkg/models"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// ClaimSigner handles signing and verification of claims
type ClaimSigner struct {
	agentKey *key.AgentKey
}

// NewClaimSigner creates a new ClaimSigner with the given agent key
func NewClaimSigner(agentKey *key.AgentKey) *ClaimSigner {
	return &ClaimSigner{
		agentKey: agentKey,
	}
}

// verifySignatureAgainstAddress verifies a signature against an Ethereum address
func verifySignatureAgainstAddress(hash []byte, signature []byte, address string) (bool, error) {
	// Ensure signature is 65 bytes (including recovery ID)
	if len(signature) != 65 {
		return false, fmt.Errorf("invalid signature length: expected 65 bytes, got %d", len(signature))
	}

	// Recover the public key from the signature (using full 65-byte signature)
	pubKey, err := crypto.SigToPub(hash, signature)
	if err != nil {
		return false, fmt.Errorf("failed to recover public key: %w", err)
	}

	// Get the address from the public key
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)

	// Compare with expected address
	return recoveredAddr == common.HexToAddress(address), nil
}

// SignDelegationClaim signs a DelegationClaim and adds the cryptographic proof
func (cs *ClaimSigner) SignDelegationClaim(claim *models.DelegationClaim) error {
	if cs.agentKey == nil {
		return fmt.Errorf("cannot sign: signer was created without a private key (verification-only mode)")
	}

	// Create canonical hash of the claim (without proof)
	hash, err := cs.hashDelegationClaim(claim)
	if err != nil {
		return fmt.Errorf("failed to hash delegation claim: %w", err)
	}

	// Sign the hash
	signature, err := cs.agentKey.Sign(hash)
	if err != nil {
		return fmt.Errorf("failed to sign delegation claim: %w", err)
	}

	// Add proof to the claim
	claim.Proof = &models.CredentialProof{
		Type:               string(models.EcdsaSecp256k1Signature2019),
		Created:            time.Now().Format(time.RFC3339),
		VerificationMethod: fmt.Sprintf("%s#key-1", cs.agentKey.DID),
		ProofPurpose:       string(models.AssertionMethod),
		ProofValue:         hex.EncodeToString(signature),
		Domain: models.EIP712Domain{
			Name:    "AgentID",
			Version: "1",
			ChainID: 1, // Mainnet
		},
	}

	return nil
}

// VerifyDelegationClaim verifies the signature on a DelegationClaim
func (cs *ClaimSigner) VerifyDelegationClaim(claim *models.DelegationClaim, expectedDelegatorDID string) (bool, error) {
	if claim.Proof == nil {
		return false, fmt.Errorf("delegation claim has no proof")
	}

	// Verify the claim is from the expected delegator
	if claim.DelegatorDID != expectedDelegatorDID {
		return false, fmt.Errorf("claim delegator DID mismatch: expected %s, got %s", 
			expectedDelegatorDID, claim.DelegatorDID)
	}

	// Extract address from delegator DID
	address, err := key.ExtractAddressFromDID(claim.DelegatorDID)
	if err != nil {
		return false, fmt.Errorf("failed to extract address from delegator DID: %w", err)
	}

	// Create hash of claim (without proof)
	tempClaim := *claim
	tempClaim.Proof = nil
	hash, err := hashDelegationClaimStruct(&tempClaim)
	if err != nil {
		return false, fmt.Errorf("failed to hash delegation claim: %w", err)
	}

	// Decode signature
	signature, err := hex.DecodeString(claim.Proof.ProofValue)
	if err != nil {
		return false, fmt.Errorf("failed to decode signature: %w", err)
	}

	// Verify signature matches the expected address
	return verifySignatureAgainstAddress(hash, signature, address.Hex())
}

// VerifyDelegationChain verifies all signatures in a delegation chain
func (cs *ClaimSigner) VerifyDelegationChain(chain *models.DelegationChain) (bool, error) {
	if len(chain.Delegations) == 0 {
		return false, fmt.Errorf("empty delegation chain")
	}

	// Verify each delegation in the chain
	for i, delegation := range chain.Delegations {
		// For first delegation, verify against root delegator
		if i == 0 {
			valid, err := cs.VerifyDelegationClaim(delegation, delegation.DelegatorDID)
			if err != nil {
				return false, fmt.Errorf("failed to verify root delegation: %w", err)
			}
			if !valid {
				return false, fmt.Errorf("failed to verify root delegation: invalid signature")
			}
			continue
		}

		// For subsequent delegations, verify against previous delegate
		prevDelegation := chain.Delegations[i-1]
		valid, err := cs.VerifyDelegationClaim(delegation, prevDelegation.DelegateDID)
		if err != nil {
			return false, fmt.Errorf("failed to verify delegation %d: %w", i, err)
		}
		if !valid {
			return false, fmt.Errorf("failed to verify delegation %d: invalid signature", i)
		}
	}

	return true, nil
}

// hashDelegationClaim creates a canonical hash of a DelegationClaim
func (cs *ClaimSigner) hashDelegationClaim(claim *models.DelegationClaim) ([]byte, error) {
	// Create a copy without the proof for hashing
	tempClaim := *claim
	tempClaim.Proof = nil
	
	return hashDelegationClaimStruct(&tempClaim)
}

// hashDelegationClaimStruct creates a deterministic hash of a DelegationClaim struct
func hashDelegationClaimStruct(claim *models.DelegationClaim) ([]byte, error) {
	// Create deterministic JSON (sorted keys)
	jsonBytes, err := json.Marshal(claim)
	if err != nil {
		return nil, err
	}
	
	// Hash with Keccak256 (Ethereum standard)
	hash := crypto.Keccak256(jsonBytes)
	return hash, nil
} 