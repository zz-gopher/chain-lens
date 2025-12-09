# Chain Lens 🔍

![Go Version](https://img.shields.io/badge/Go-1.19%2B-blue)
![License](https://img.shields.io/badge/license-MIT-green)
![Status](https://img.shields.io/badge/status-active-success)

**Chain Lens** is a high-performance, industrial-grade CLI tool designed for EVM blockchain data visibility. It features a **hybrid query engine** that intelligently combines **Multicall3 batching** with **concurrent fallback mechanisms** to ensure maximum speed and reliability.

Built with **Go**, it handles ERC-20, ERC-721, and native token balances with ease, making it the perfect tool for developers and data analysts who need to query thousands of addresses without worrying about RPC limits or failures.

## 🚀 Features

- **⚡ Ultra-Fast Batching (Multicall3):** Aggregates hundreds of queries into a single RPC call using **Multicall3**, reducing network overhead by up to **95%** compared to traditional loops.
- **🛡️ Resilient Fallback System:** Built for reliability. If Multicall fails (globally or partially), the engine automatically degrades to **concurrent single-node queries** to ensure no data is left behind.
- **🧠 Smart Caching:** Implements "Lazy Loading" for token metadata (decimals/symbols), eliminating redundant RPC calls and optimizing throughput.
- **⚙️ Plug-and-Play Configuration:** Instantly switch between RPC endpoints (Infura, Alchemy, Ankr, etc.) and target contracts via `config.json`.
- **🎯 High Precision:** Utilizes `math/big` to handle raw blockchain integers, ensuring zero precision loss for financial data.
- **💎 Multi-Asset Support:** Seamlessly queries **Native Coins (ETH/BNB)**, **ERC-20 Tokens**, and **ERC-721 NFTs** in a single workflow.
- **📂 Bulk Processing:** Efficiently processes large lists of wallet addresses from local text files.

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
  "token_type": "erc721" // optional: "erc20", "erc721", or "native"
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