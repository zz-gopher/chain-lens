# Chain Lens 🔍

**Chain Lens** is a high-performance CLI tool designed to provide visibility into EVM blockchain data. It currently supports batch querying of **ERC-20, ERC-721, and native token balances** using concurrent processing.

Built with **Go**, it is engineered for speed, precision, and ease of use for developers and data analysts.

## 🚀 Features

- **⚡ High Performance:** Leverages Go's Goroutines to query hundreds of wallet addresses concurrently.
- **⚙ Configurable:** Switch RPC endpoints (Infura, Alchemy, Ankr, etc.) and target token contracts instantly via `config.json`.
- **🎯 High Precision:** Implements `math/big` to handle raw blockchain integers without precision loss.
- **📂 Simple Input:** Accepts a list of target addresses via a local text file.
- **💎 Multi-Token Support:** Supports native coins, ERC20 tokens, and ERC721 (NFT) balances.

## 🛠️ Getting Started

### 1. Prerequisites
- **Go 1.19** or higher installed.

### 2. Installation

Clone the repository and install dependencies:

```bash
git clone https://github.com/zz-gopher/chain-lens.git
cd chain-lens
go mod tidy
```

### 3. Configuration
Edit config.json to set your RPC URL, target token address, and optionally token type:
```txt
{
  "rpc_url": "https://rpc.soneium.org",
  "token_address": "0x7BF02b42b9d4cCD85b497C9F53e6b7474f9c2546",
  "token_type": "erc20" // optional: "erc20", "erc721", or "native"
}
```
- token_type is optional; if omitted, the tool will automatically detect the token type.

### 4. Prepare Wallet List

Create a text file (e.g., wallets.txt) with one wallet address per line:
```txt
0x1234...abcd
0x5678...efgh
...
```
### 5. Run the Tool
```bash
go run main.go -config=config.json -wallets=wallets.txt
```
- The CLI will print balances for each address and token.

- Supports ERC20, ERC721, and native token balances in one run.

### 6. Example Output
```yaml
status_messages:
  - "📂 Successfully loaded 3 wallet addresses"
  - "Connected to EVM"
  - "🔹 Auto-detect ERC20 token"
  - "✅ [2] Address: 0xde0B... | Balance: 1061.4302 USDT"
  - "✅ [1] Address: 0xd8dA... | Balance: 262.6767 USDT"
  - "✅ [3] Address: 0xBE0e... | Balance: 0.0000 USDT"
  - "🎉 All tasks completed! Total elapsed time: 1.766289s"

```
### 7. Notes

Ensure the RPC endpoint supports the network you are querying.

For ERC721 (NFTs), the balance represents the number of tokens owned.