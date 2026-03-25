package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/ak68a/agentid/pkg/key"
	"github.com/ak68a/agentid/pkg/models"
	"github.com/ak68a/agentid/pkg/signer"
	"github.com/urfave/cli/v2"
)

func main() {
	app := &cli.App{
		Name:  "agentid",
		Usage: "AgentID Core CLI - Identity and Authorization Management",
		Commands: []*cli.Command{
			{
				Name:  "generate",
				Usage: "Generate a new agent identity",
				Action: func(c *cli.Context) error {
					return generateAgent()
				},
			},
			{
				Name:  "create-claim",
				Usage: "Create and sign a new claim",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "private-key",
						Usage:    "Private key in hex format",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "owner-did",
						Usage:    "Owner's DID (e.g., did:web:acme-corp.com)",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "action",
						Usage:    "Action type (transfer, quote, booking, etc.)",
						Required: true,
					},
					&cli.StringFlag{
						Name:     "scope",
						Usage:    "Scope/resource (ETH, USD, flights, etc.)",
						Required: true,
					},
					&cli.StringFlag{
						Name:  "max-amount",
						Usage: "Maximum amount for financial actions",
					},
					&cli.Int64Flag{
						Name:  "expires-in",
						Usage: "Expiration time in hours from now",
						Value: 24,
					},
				},
				Action: func(c *cli.Context) error {
					return createClaim(c)
				},
			},
			{
				Name:  "verify-claim",
				Usage: "Verify a claim's signature",
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "claim-file",
						Usage:    "Path to JSON file containing the claim",
						Required: true,
					},
				},
				Action: func(c *cli.Context) error {
					return verifyClaim(c)
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}

func ensureBuildDir() error {
	buildDir := "build"
	if err := os.MkdirAll(buildDir, 0755); err != nil {
		return fmt.Errorf("failed to create build directory: %w", err)
	}
	return nil
}

func generateAgent() error {
	fmt.Println("🔐 Generating new agent identity...")
	
	// Ensure build directory exists
	if err := ensureBuildDir(); err != nil {
		return err
	}
	
	// Generate a new agent key
	agentKey, err := key.GenerateAgentKey()
	if err != nil {
		return fmt.Errorf("failed to generate agent key: %w", err)
	}

	// Print the agent details
	fmt.Printf("\n✅ Agent Generated:\n")
	fmt.Printf("   DID: %s\n", agentKey.DID)
	fmt.Printf("   Address: %s\n", agentKey.Address.Hex())
	fmt.Printf("   Private Key: %s\n", agentKey.GetPrivateKeyHex())
	fmt.Printf("   Public Key: %s\n", agentKey.GetPublicKeyHex())
	
	// Save to file
	output := map[string]string{
		"did":         agentKey.DID,
		"address":     agentKey.Address.Hex(),
		"private_key": agentKey.GetPrivateKeyHex(),
		"public_key":  agentKey.GetPublicKeyHex(),
	}
	
	jsonData, err := json.MarshalIndent(output, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal agent data: %w", err)
	}
	
	filename := filepath.Join("build", fmt.Sprintf("agent_%s.json", agentKey.Address.Hex()[:8]))
	if err := os.WriteFile(filename, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to save agent data: %w", err)
	}
	
	fmt.Printf("\n📄 Agent data saved to: %s\n", filename)
	return nil
}

func createClaim(c *cli.Context) error {
	// Import the agent key
	privateKey := c.String("private-key")
	agentKey, err := key.ImportFromHex(privateKey)
	if err != nil {
		return fmt.Errorf("failed to import private key: %w", err)
	}

	// Create the claim
	ownerDID := c.String("owner-did")
	action := c.String("action")
	scope := c.String("scope")
	expiresIn := c.Int64("expires-in")
	
	// Generate a nonce
	nonceBytes := make([]byte, 16)
	if _, err := rand.Read(nonceBytes); err != nil {
		return fmt.Errorf("failed to generate nonce: %w", err)
	}
	nonce := hex.EncodeToString(nonceBytes)
	
	// Create expiration time
	expiresAt := time.Now().Add(time.Duration(expiresIn) * time.Hour).Unix()
	
	// Create the delegation claim
	claim := &models.DelegationClaim{
		DelegatorDID: agentKey.DID,
		DelegateDID:  ownerDID,
		Action:       action,
		Scope:        scope,
		IssuedAt:     time.Now().Unix(),
		ExpiresAt:    expiresAt,
		Nonce:        nonce,
		Type:         models.AgentAuthorizationCredentialType,
		Context:      models.StandardContexts,
		Issuer:       agentKey.DID,
		Subject:      ownerDID,
		Constraints:  make(map[string]interface{}),
	}
	
	// Add max amount if provided
	if maxAmount := c.String("max-amount"); maxAmount != "" {
		claim.Constraints["max_amount"] = maxAmount
	}
	
	// Sign the claim
	signer := signer.NewClaimSigner(agentKey)
	if err := signer.SignDelegationClaim(claim); err != nil {
		return fmt.Errorf("failed to sign claim: %w", err)
	}
	
	// Print claim details
	fmt.Printf("\n📜 Claim Created:\n")
	fmt.Printf("   Delegator DID: %s\n", claim.DelegatorDID)
	fmt.Printf("   Delegate DID: %s\n", claim.DelegateDID)
	fmt.Printf("   Action: %s\n", claim.Action)
	fmt.Printf("   Scope: %s\n", claim.Scope)
	if maxAmount, ok := claim.Constraints["max_amount"].(string); ok {
		fmt.Printf("   Max Amount: %s\n", maxAmount)
	}
	fmt.Printf("   Expires: %s\n", time.Unix(claim.ExpiresAt, 0).Format(time.RFC3339))
	fmt.Printf("   Nonce: %s\n", claim.Nonce)
	if claim.Proof != nil {
		fmt.Printf("   Proof: %s\n", claim.Proof.ProofValue[:16] + "...")
	}
	
	// Save to file
	jsonData, err := json.MarshalIndent(claim, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal claim: %w", err)
	}
	
	// Ensure build directory exists
	if err := ensureBuildDir(); err != nil {
		return err
	}
	
	filename := filepath.Join("build", fmt.Sprintf("claim_%s.json", nonce[:8]))
	if err := os.WriteFile(filename, jsonData, 0600); err != nil {
		return fmt.Errorf("failed to save claim: %w", err)
	}
	
	fmt.Printf("\n📄 Claim saved to: %s\n", filename)
	return nil
}

func verifyClaim(c *cli.Context) error {
	// Read claim file
	claimFile := c.String("claim-file")
	// If the file doesn't exist in the current directory, try the build directory
	if _, err := os.Stat(claimFile); os.IsNotExist(err) {
		buildPath := filepath.Join("build", claimFile)
		if _, err := os.Stat(buildPath); err == nil {
			claimFile = buildPath
		}
	}
	
	jsonData, err := os.ReadFile(claimFile)
	if err != nil {
		return fmt.Errorf("failed to read claim file: %w", err)
	}
	
	// Parse claim
	var claim models.DelegationClaim
	if err := json.Unmarshal(jsonData, &claim); err != nil {
		return fmt.Errorf("failed to parse claim: %w", err)
	}
	
	// Create signer with the delegator's key
	// Note: In a real system, you'd need to get the private key securely
	// For demo purposes, we'll just verify the signature
	signer := signer.NewClaimSigner(nil) // No key needed for verification
	
	// Verify the claim
	valid, err := signer.VerifyDelegationClaim(&claim, claim.DelegatorDID)
	if err != nil {
		return fmt.Errorf("verification failed: %w", err)
	}
	
	if valid {
		fmt.Printf("\n✅ Claim Verified Successfully:\n")
		fmt.Printf("   Delegator: %s\n", claim.DelegatorDID)
		fmt.Printf("   Delegate: %s\n", claim.DelegateDID)
		fmt.Printf("   Action: %s\n", claim.Action)
		fmt.Printf("   Scope: %s\n", claim.Scope)
		fmt.Printf("   Expires: %s\n", time.Unix(claim.ExpiresAt, 0).Format(time.RFC3339))
		if claim.IsExpired() {
			fmt.Printf("   ⚠️  Claim has expired\n")
		}
	} else {
		fmt.Printf("\n❌ Claim Verification Failed\n")
	}
	
	return nil
} 