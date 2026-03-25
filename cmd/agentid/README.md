# AgentID CLI Tool

The AgentID CLI tool provides a command-line interface for managing agent identities and their authorization claims. It allows you to generate agent identities, create signed claims, and verify those claims.

## Installation

1. Clone the repository:
```bash
git clone https://github.com/ak68a/agentid.git
cd agentid
```

2. Build the CLI tool:
```bash
go build -o agentid cmd/agentid/main.go
```

The `agentid` binary will be created in your current directory.

## Usage

The CLI tool provides three main commands:

### 1. Generate Agent Identity

Generate a new agent identity with a unique DID and keypair:

```bash
./agentid generate
```

This will:
- Generate a new Ethereum keypair
- Create a DID for the agent
- Save the agent details to a JSON file (e.g., `build/agent_0x1234abcd.json`)

Example output:
```
🔐 Generating new agent identity...

✅ Agent Generated:
   DID: did:ackid:0x1234...
   Address: 0x1234...
   Private Key: abcd...
   Public Key: 04ef...

📄 Agent data saved to: build/agent_0x1234abcd.json
```

### 2. Create Claim

Create and sign a new authorization claim:

```bash
./agentid create-claim --private-key "your_private_key" --owner-did "did:web:example.com" --action "transfer" --scope "ETH" --max-amount "1000000000000000000" --expires-in 24
```

Parameters:
- `--private-key`: Agent's private key in hex format (required)
- `--owner-did`: Owner's DID (required)
- `--action`: Action type (required, e.g., "transfer", "quote", "booking")
- `--scope`: Scope/resource (required, e.g., "ETH", "USD", "flights")
- `--max-amount`: Maximum amount for financial actions (optional)
- `--expires-in`: Expiration time in hours (optional, default: 24)

Example output:
```
📜 Claim Created:
   Delegator DID: did:ackid:0x1234...
   Delegate DID: did:web:example.com
   Action: transfer
   Scope: ETH
   Max Amount: 1000000000000000000
   Expires: 2024-03-21T15:30:00Z
   Nonce: abcd1234...
   Proof: 5b5a5ff8...

📄 Claim saved to: build/claim_abcd1234.json
```

### 3. Verify Claim

Verify the signature and validity of a claim:

```bash
./agentid verify-claim --claim-file claim_abcd1234.json
```

Example output:
```
✅ Claim Verified Successfully:
   Delegator: did:ackid:0x1234...
   Delegate: did:web:example.com
   Action: transfer
   Scope: ETH
   Expires: 2024-03-21T15:30:00Z
```

## Complete Example Workflow

1. Generate a new agent:
```bash
./agentid generate
```

2. Create a claim using the generated agent's private key:
```bash
./agentid create-claim --private-key "$(jq -r .private_key build/agent_0x1234abcd.json)" --owner-did "did:web:acme-corp.com" --action "transfer" --scope "ETH" --max-amount "1000000000000000000"
```

3. Verify the claim:
```bash
./agentid verify-claim --claim-file build/claim_abcd1234.json
```

## Claim Types and Scopes

### Available Actions
- `transfer`: Transfer funds or assets
- `quote`: Get price quotes
- `booking`: Book services or resources

### Available Scopes
- `ETH`: Ethereum cryptocurrency
- `USD`: US Dollar
- `flights`: Flight bookings
- `hotels`: Hotel bookings

## Security Notes

1. Always keep your private keys secure and never share them
2. The agent files and claim files contain sensitive information and are saved with restricted permissions (0600)
3. In production environments, consider using a secure key management system instead of storing private keys in files

## Troubleshooting

1. If verification fails with "delegation claim has no proof", ensure you're using the `create-claim` command to properly sign the claim
2. If you get "invalid private key" errors, ensure the private key is in the correct hex format
3. For expired claims, the verification will still succeed but will show a warning about expiration

## License

This tool is part of the AgentID Core project and is licensed under the same terms as the main project. 