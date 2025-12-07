# Chain Lens 🔍

**Chain Lens** is a high-performance CLI tool designed to provide visibility into EVM blockchain data. It currently supports batch querying of ERC-20 token balances using concurrent processing.

Built with **Go**, it is engineered for speed, precision, and ease of use for developers and data analysts.

## 🚀 Features

- **⚡️ High Performance:** Leverages Go's Goroutines to query hundreds of wallet addresses concurrently.
- **⚙️ Configurable:** Switch RPC endpoints (Infura, Alchemy, Ankr, etc.) and target Token contracts instantly via `config.json`.
- **🎯 High Precision:** Implements `math/big` to handle raw blockchain integers without precision loss.
- **📂 Simple Input:** Accepts a list of target addresses via a local text file.

## 🛠️ Getting Started

### 1. Prerequisites
- **Go 1.19** or higher installed.

### 2. Installation

Clone the repository and prepare dependencies:



```bash
git clone [https://github.com/zz-gopher/chain-lens.git] 
cd chain-lens
go mod tidy

